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

package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"

	"github.com/cisco-open/libnasp/pkg/istio"
	"github.com/cisco-open/libnasp/pkg/network"
	"github.com/cisco-open/libnasp/pkg/util"
)

var mode string
var requestURL string
var requestCount int
var usePushGateway bool
var pushGatewayAddress string
var sleepBeforeClientExit time.Duration
var dumpClientResponse bool
var heimdallURL string
var sendSubsequentHTTPRequest bool

func init() {
	flag.StringVar(&mode, "mode", "server", "server/client mode")
	flag.StringVar(&requestURL, "request-url", "http://echo.testing:80", "http request url")
	flag.IntVar(&requestCount, "request-count", 1, "outgoing http request count")
	flag.BoolVar(&usePushGateway, "use-push-gateway", false, "use push gateway for metrics")
	flag.StringVar(&pushGatewayAddress, "push-gateway-address", "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091", "push gateway address")
	flag.DurationVar(&sleepBeforeClientExit, "client-sleep", 0, "sleep duration before client exit")
	flag.BoolVar(&dumpClientResponse, "dump-client-response", false, "dump http client response")
	flag.StringVar(&heimdallURL, "heimdall-url", "https://localhost:16443/config", "Heimdall URL")
	flag.BoolVar(&sendSubsequentHTTPRequest, "send-subsequent-http-request", true, "Whether to send subsequent HTTP request in server mode")
	klog.InitFlags(nil)
	flag.Parse()
}

func sendHTTPRequest(url string, transport http.RoundTripper, logger logr.Logger) error {
	httpClient := &http.Client{
		Transport: transport,
	}
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if dumpClientResponse {
		buff, _ := ioutil.ReadAll(response.Body)
		os.Stdout.Write(buff)
	}

	if state, ok := network.ConnectionStateFromContext(response.Request.Context()); ok {
		util.PrintConnectionState(state, logger)
	}

	return nil
}

func main() {
	logger := klog.TODO()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authToken := os.Getenv("NASP_AUTH_TOKEN")
	if authToken == "" {
		panic(errors.New("NASP_AUTH_TOKEN env var must be specified."))
	}

	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
	istioHandlerConfig.IstioCAConfigGetter = istio.IstioCAConfigGetterHeimdall(ctx, heimdallURL, authToken, "v1")

	if usePushGateway {
		istioHandlerConfig.PushgatewayConfig = &istio.PushgatewayConfig{
			Address:          pushGatewayAddress,
			UseUniqueIDLabel: true,
		}
	}

	iih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, logger)
	if err != nil {
		panic(err)
	}

	iih.Run(ctx)

	// use client side connection pooling
	t := http.DefaultTransport.(*http.Transport)
	t.MaxIdleConns = 50
	t.MaxConnsPerHost = 50
	t.MaxIdleConnsPerHost = 50

	transport, err := iih.GetHTTPTransport(t)
	if err != nil {
		panic(err)
	}

	if mode == "client" {
		var wg sync.WaitGroup
		i := 0
		clientErrors := []error{}
		for i < requestCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := sendHTTPRequest(requestURL, transport, logger); err != nil {
					clientErrors = append(clientErrors, err)
					logger.Error(err, "could not send http request")
				}
			}()
			i++
		}

		wg.Wait()

		time.Sleep(sleepBeforeClientExit)

		if len(clientErrors) > 0 {
			os.Exit(2)
		}

		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Logger())
	r.GET("/", func(c *gin.Context) {
		if sendSubsequentHTTPRequest {
			if err := sendHTTPRequest(requestURL, transport, logger); err != nil {
				logger.Error(err, "could not send http request")
			}
		}

		if state, ok := network.ConnectionStateFromContext(c.Request.Context()); ok {
			util.PrintConnectionState(state, logger)
		}

		c.Data(http.StatusOK, "text/html", []byte("Hello world!"))
	})
	err = iih.ListenAndServe(context.Background(), ":8080", r.Handler())
	if err != nil {
		panic(err)
	}
}
