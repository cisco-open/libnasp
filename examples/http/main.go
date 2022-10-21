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
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
	"github.com/cisco-open/nasp/pkg/environment"
	"github.com/cisco-open/nasp/pkg/istio"
)

func init() {
	for _, env := range []string{
		"NASP_ISTIO_CA_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_ISTIO_VERSION=1.13.5",
		"NASP_ISTIO_REVISION=cp-v113x.istio-system",
		"NASP_TYPE=sidecar",
		"NASP_POD_NAME=alpine-efefefef-f5wwf",
		"NASP_POD_NAMESPACE=default",
		"NASP_POD_OWNER=kubernetes://apis/apps/v1/namespaces/default/deployments/alpine",
		"NASP_WORKLOAD_NAME=alpine",
		"NASP_MESH_ID=mesh1",
		"NASP_CLUSTER_ID=waynz0r-0626-01",
		"NASP_NETWORK=network2",
		"NASP_APP_CONTAINERS=alpine",
		"NASP_INSTANCE_IP=10.20.4.75",
		"NASP_LABELS=security.istio.io/tlsMode:istio, pod-template-hash:efefefef, service.istio.io/canonical-revision:latest, istio.io/rev:cp-v111x.istio-system, topology.istio.io/network:network1, k8s-app:alpine, service.istio.io/canonical-name:alpine, app:alpine, version:v11",
		"NASP_SEARCH_DOMAINS=demo.svc.cluster.local,svc.cluster.local,cluster.local",
	} {
		p := strings.Split(env, "=")
		if len(p) != 2 {
			continue
		}
		os.Setenv(p[0], p[1])
	}
}

var mode string
var url string
var requestCount int
var usePushGateway bool
var pushGatewayAddress string
var sleepBeforeClientExit time.Duration
var dumpClientResponse bool

func init() {
	flag.StringVar(&mode, "mode", "server", "server/client mode")
	flag.StringVar(&url, "url", "http://echo:8080", "http client url")
	flag.IntVar(&requestCount, "request-count", 1, "outgoing http request count")
	flag.BoolVar(&usePushGateway, "use-push-gateway", false, "use push gateway for metrics")
	flag.StringVar(&pushGatewayAddress, "push-gateway-address", "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091", "push gateway address")
	flag.DurationVar(&sleepBeforeClientExit, "client-sleep", 0, "sleep duration before client exit")
	flag.BoolVar(&dumpClientResponse, "dump-client-response", false, "dump http client response")
	klog.InitFlags(nil)
	flag.Parse()
}

func sendHTTPRequest(url string, transport http.RoundTripper) error {
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

	return nil
}

func main() {
	istioHandlerConfig := &istio.IstioIntegrationHandlerConfig{
		MetricsAddress: ":15090",
		UseTLS:         true,
		IstioCAConfigGetter: func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
			return istio_ca.GetIstioCAClientConfig(e.ClusterID, e.IstioRevision)
		},
	}

	if usePushGateway {
		istioHandlerConfig.PushgatewayConfig = &istio.PushgatewayConfig{
			Address:          pushGatewayAddress,
			UseUniqueIDLabel: true,
		}
	}

	iih, err := istio.NewIstioIntegrationHandler(istioHandlerConfig, klog.TODO())
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih.Run(ctx)

	transport, err := iih.GetHTTPTransport(http.DefaultTransport)
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
				if err := sendHTTPRequest(url, transport); err != nil {
					klog.Error(err, "could not send http request")
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
		if err := sendHTTPRequest("http://echo:8080", transport); err != nil {
			klog.Error(err, "could not send http request")
		}
		c.Data(http.StatusOK, "text/html", []byte("Hello world!"))
	})
	err = iih.ListenAndServe(context.Background(), ":8080", r.Handler())
	if err != nil {
		panic(err)
	}
}
