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
	"crypto/tls"
	"crypto/x509"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	adsconfig "wwwin-github.cisco.com/eti/nasp/pkg/ads/config"

	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/klog/v2/klogr"

	"wwwin-github.cisco.com/eti/nasp/pkg/ads"

	klog "k8s.io/klog/v2"

	istio_ca "wwwin-github.cisco.com/eti/nasp/pkg/ca/istio"
)

var (
	clusterID     = "waynz0r-0626-01"
	istioRevision = "cp-v113x.istio-system"
)

func main() {
	flag.Parse()

	md, _ := structpb.NewStruct(map[string]interface{}{
		"NAMESPACE":     "demo",
		"MESH_ID":       "mesh1",
		"CLUSTER_ID":    clusterID,
		"NETWORK":       "network1",
		"ISTIO_VERSION": "1.13.5",
	})

	klog.Info("getting config for istio")
	istioCAClientConfig, err := istio_ca.GetIstioCAClientConfig(clusterID, istioRevision)
	if err != nil {
		panic(err)
	}

	klog.Info("get certificate")
	caClient := istio_ca.NewIstioCAClient(istioCAClientConfig, klog.TODO())
	cert, err := caClient.GetCertificate("spiffe://cluster.local/ns/demo/sa/default", time.Hour*24)
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", istioCAClientConfig.CAEndpoint)
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert.GetTLSCertificate()},
		ServerName:   "istiod-cp-v113x.istio-system.svc",
		RootCAs:      x509.NewCertPool(),
	}
	tlsConfig.RootCAs.AppendCertsFromPEM(istioCAClientConfig.CApem)

	clientConfig := &adsconfig.ClientConfig{
		NodeInfo: &adsconfig.NodeInfo{
			// workload id - update if id changed (e.g. workload was restarted)
			Id:          "sidecar~10.20.32.140~echo-76c8b47f5b-x2vn6.demo~demo.svc.cluster.local",
			ClusterName: "echo.demo",
			Metadata:    md,
		},
		ManagementServerAddress: addr,
		ClusterID:               clusterID,
		TLSConfig:               tlsConfig,
		SearchDomains:           []string{"demo.svc.cluster.local", "svc.cluster.local", "cluster.local"},
	}

	ctx := logr.NewContext(
		context.Background(),
		klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog)).WithName("adsclient"),
	)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdownSignals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	c := make(chan os.Signal, len(shutdownSignals))
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	client, err := ads.Connect(ctx, clientConfig)
	if err != nil {
		panic(err)
	}

	address := ":8080"
	listenerProps, err := client.GetListenerProperties(ctx, address)
	if err != nil {
		klog.Errorln("get listener properties failed", err)
	}

	// address = "echo.demo.svc.cluster.local:80"
	address = "echo:80"
	httpClientProps, err := client.GetHTTPClientPropertiesByHost(ctx, address)
	if err != nil {
		klog.Errorln("get http client properties", err)
	}

	address = "demo.demo.svc.cluster.local:8081"
	tcpClientProps, err := client.GetTCPClientPropertiesByHost(ctx, address)
	if err != nil {
		klog.Errorln("get tcp client properties", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-listenerProps:
			if ok {
				if p.Error() != nil {
					klog.Errorln("get listener properties failed", p.Error())
					return
				}
				klog.Infoln("listener properties", p.ListenerProperties(), "address", ":8080")
			}
		case p, ok := <-httpClientProps:
			if ok {
				if p.Error() != nil {
					klog.Errorln("get http client properties", p.Error())
					return
				}

				klog.Infoln("http client properties", p.ClientProperties(), "address", "echo.demo.svc.cluster.local:80")
			}
		case p, ok := <-tcpClientProps:
			if ok {
				if p.Error() != nil {
					klog.Errorln("get tcp client properties", p.Error())
					return
				}

				klog.Infoln("tcp client properties", p.ClientProperties(), "address", "demo.demo.svc.cluster.local:8081")
			}
		}
	}
}
