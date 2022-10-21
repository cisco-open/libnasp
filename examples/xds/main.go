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
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/logr"

	"wwwin-github.cisco.com/eti/nasp/pkg/ads"
	"wwwin-github.cisco.com/eti/nasp/pkg/environment"
	"wwwin-github.cisco.com/eti/nasp/pkg/istio/discovery"

	klog "k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	istio_ca "wwwin-github.cisco.com/eti/nasp/pkg/ca/istio"
)

func init() {
	for _, env := range []string{
		"NASP_ISTIO_CA_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_XDS_SERVER_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_TYPE=sidecar",
		"NASP_POD_NAME=demo",
		"NASP_POD_NAMESPACE=demo",
		"NASP_POD_OWNER=kubernetes://apis/apps/v1/namespaces/demo/deployments/demo",
		"NASP_WORKLOAD_NAME=demo",
		"NASP_ISTIO_VERSION=1.11.4",
		"NASP_ISTIO_REVISION=cp-v113x.istio-system",
		"NASP_MESH_ID=mesh1",
		"NASP_NETWORK=network2",
		"NASP_CLUSTER_ID=waynz0r-0626-01",
		"NASP_APP_CONTAINERS=alpine",
		"NASP_INSTANCE_IP=10.20.4.75",
		"NASP_LABELS=security.istio.io/tlsMode:istio, pod-template-hash:efefefef, service.istio.io/canonical-revision:latest, istio.io/rev:cp-v111x.istio-system, topology.istio.io/network:network1, k8s-app:alpine, service.istio.io/canonical-name:alpine",
	} {
		p := strings.Split(env, "=")
		if len(p) != 2 {
			continue
		}
		os.Setenv(p[0], p[1])
	}
}

func main() {
	flag.Parse()

	e, err := environment.GetIstioEnvironment("NASP_")
	if err != nil {
		panic(err)
	}

	klog.Info("getting config for istio")
	istioCAClientConfig, err := istio_ca.GetIstioCAClientConfig(e.ClusterID, e.IstioRevision)
	if err != nil {
		panic(err)
	}

	caClient := istio_ca.NewIstioCAClient(istioCAClientConfig, klog.TODO())

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

	client := discovery.NewXDSDiscoveryClient(e, caClient, klog.TODO())
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	cctx, ccancel := context.WithCancel(context.Background())

	p, err := client.GetListenerProperties(cctx, ":8080", func(p discovery.ListenerProperties) {
		klog.TODO().Info("listener prop callback", "properties", p)
	})
	if err != nil {
		panic(err)
	}
	klog.TODO().Info("listener prop", "properties", p)

	hp, err := client.GetHTTPClientPropertiesByHost(cctx, "joska:80", func(p discovery.HTTPClientProperties) {
		klog.TODO().Info("http prop callback", "properties", p)
	})
	if ads.IgnoreHostNotFound(err) != nil {
		panic(err)
	}
	klog.TODO().Info("http prop", "properties", hp)

	tp, err := client.GetTCPClientPropertiesByHost(cctx, "demo.demo.svc.cluster.local:8081", func(p discovery.ClientProperties) {
		klog.TODO().Info("tcp prop callback", "properties", p)
	})
	if err != nil {
		panic(err)
	}
	klog.TODO().Info("tcp prop", "properties", tp)

	go func() {
		time.Sleep(50 * time.Second)
		ccancel()
	}()

	<-ctx.Done()
	ccancel()
}
