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
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network"
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
	httpReq, err := http.NewRequest("GET", url, nil)
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
		fmt.Printf("%s\n", string(buff))
	}

	if conn, ok := network.WrappedConnectionFromContext(response.Request.Context()); ok {
		printConnectionInfo(conn, logger)
	}

	return nil
}

func printConnectionInfo(connection network.Connection, logger logr.Logger) {
	localAddr := connection.LocalAddr().String()
	remoteAddr := connection.RemoteAddr().String()
	var localSpiffeID, remoteSpiffeID string

	if cert := connection.GetLocalCertificate(); cert != nil {
		localSpiffeID = cert.GetFirstURI()
	}

	if cert := connection.GetPeerCertificate(); cert != nil {
		remoteSpiffeID = cert.GetFirstURI()
	}

	logger.Info("connection info", "localAddr", localAddr, "localSpiffeID", localSpiffeID, "remoteAddr", remoteAddr, "remoteSpiffeID", remoteSpiffeID, "ttfb", connection.GetTimeToFirstByte().Format(time.RFC3339Nano))
}

func main() {
	logger := klog.TODO()

	istioHandlerConfig := &istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, "test-http-16362813-F46B-41AC-B191-A390DB1F6BDF", "16362813-F46B-41AC-B191-A390DB1F6BDF", "v1"),
	}

	if usePushGateway {
		istioHandlerConfig.PushgatewayConfig = &istio.PushgatewayConfig{
			Address:          pushGatewayAddress,
			UseUniqueIDLabel: true,
		}
	}

	iih, err := istio.NewIstioIntegrationHandler(istioHandlerConfig, logger)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih.Run(ctx)

	// make idle timeout minimal to test least request increment/decrement
	t := http.DefaultTransport.(*http.Transport)
	t.IdleConnTimeout = time.Nanosecond * 1

	transport, err := iih.GetHTTPTransport(t)
	if err != nil {
		panic(err)
	}

	if mode == "client" {
		var wg sync.WaitGroup
		i := 0
		for i < requestCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := sendHTTPRequest(requestURL, transport, logger); err != nil {
					logger.Error(err, "could not send http request")
				}
			}()
			i++
		}

		time.Sleep(sleepBeforeClientExit)

		wg.Wait()

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
		c.Data(http.StatusOK, "text/html", []byte("Hello world!"))
	})
	err = iih.ListenAndServe(context.Background(), ":8080", r.Handler())
	if err != nil {
		panic(err)
	}
}
