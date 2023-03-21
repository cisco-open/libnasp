// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
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
	"net"
	"net/http"

	"google.golang.org/grpc"

	itcp "github.com/cisco-open/nasp/pkg/istio/tcp"
	"github.com/cisco-open/nasp/pkg/istio/discovery"
)

type passthroughIstioIntegrationHandler struct {
}

func (h *passthroughIstioIntegrationHandler) GetHTTPTransport(transport http.RoundTripper) (http.RoundTripper, error) {
	return transport, nil
}

func (h *passthroughIstioIntegrationHandler) GetTCPListener(l net.Listener) (net.Listener, error) {
	return l, nil
}

func (h *passthroughIstioIntegrationHandler) GetTCPDialer() (itcp.Dialer, error) {
	return &net.Dialer{}, nil
}

func (h *passthroughIstioIntegrationHandler) ListenAndServe(ctx context.Context, listenAddress string, handler http.Handler) error {
	return (&http.Server{
		Addr:    listenAddress,
		Handler: handler,
	}).ListenAndServe()
}

func (h *passthroughIstioIntegrationHandler) GetDiscoveryClient() discovery.DiscoveryClient {
	return nil
}

func (h *passthroughIstioIntegrationHandler) GetGRPCDialOptions() ([]grpc.DialOption, error) {
	return nil, nil
}

func (h *passthroughIstioIntegrationHandler) Run(context.Context) error {
	return nil
}
