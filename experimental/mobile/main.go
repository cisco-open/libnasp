// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package istio

import (
	"context"
	"io"
	"net/http"
	"strings"

	_ "golang.org/x/mobile/bind"

	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
)

var logger = klog.NewKlogr()

type HTTPTransport struct {
	iih    istio.IstioIntegrationHandler
	client *http.Client
	ctx    context.Context
	cancel context.CancelFunc
}

type HTTPHeader struct {
	header http.Header
}

func (h *HTTPHeader) Get(key string) string {
	return h.header.Get(key)
}

func (h *HTTPHeader) Set(key, value string) {
	h.header.Set(key, value)
}

type HTTPResponse struct {
	StatusCode int
	Version    string
	Headers    *HTTPHeader
	Body       []byte
}

func NewHTTPTransport(heimdallURL, authorizationToken string) (*HTTPTransport, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
	istioHandlerConfig.IstioCAConfigGetter = istio.IstioCAConfigGetterHeimdall(ctx, heimdallURL, authorizationToken, "v1")
	istioHandlerConfig.PushgatewayConfig = &istio.PushgatewayConfig{
		Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
		UseUniqueIDLabel: true,
	}

	iih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, logger)
	if err != nil {
		return nil, err
	}

	transport, err := iih.GetHTTPTransport(http.DefaultTransport)
	if err != nil {
		return nil, err
	}

	if err := iih.Run(ctx); err != nil {
		cancel()
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
	}

	return &HTTPTransport{iih: iih, client: client, ctx: ctx, cancel: cancel}, nil
}

func (t *HTTPTransport) Request(method, url, body string) (*HTTPResponse, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	request, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	response, err := t.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: response.StatusCode,
		Version:    response.Proto,
		Headers:    &HTTPHeader{response.Header},
		Body:       bodyBytes,
	}, nil
}

func (t *HTTPTransport) Await() {
	<-t.ctx.Done()
}

func (t *HTTPTransport) Close() {
	t.cancel()
}
