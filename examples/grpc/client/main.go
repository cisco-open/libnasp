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
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"

	"wwwin-github.cisco.com/eti/nasp/examples/grpc/pb"
	"wwwin-github.cisco.com/eti/nasp/pkg/istio"
)

func init() {
	for _, env := range []string{
		"NASP_ISTIO_CA_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_ISTIO_VERSION=1.13.5",
		"NASP_ISTIO_REVISION=cp-v113x.istio-system",
		"NASP_NETWORK=network2",
		"NASP_SEARCH_DOMAINS=demo.svc.cluster.local",
		"NASP_TYPE=sidecar",
		"NASP_POD_NAME=grpc-client",
		"NASP_POD_NAMESPACE=default",
		"NASP_POD_OWNER=kubernetes://apis/apps/v1/namespaces/default/deployments/alpine",
		"NASP_WORKLOAD_NAME=alpine",
		"NASP_MESH_ID=mesh1",
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

func init() {
	klog.InitFlags(nil)
	flag.Parse()
}

func main() {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterRemote,
	}, klog.TODO())
	if err != nil {
		log.Fatal(err)
	}

	grpcDialOptions, err := iih.GetGRPCDialOptions()
	if err != nil {
		log.Fatal(err)
	}

	client, err := grpc.Dial(
		"catalog:8082",
		grpcDialOptions...,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih.Run(ctx)

	c := pb.NewAllsparkClient(client)
	_, err = c.Incoming(ctx, &pb.Params{})
	// fmt.Printf("%d -- %s\n", grpc.Code(err), grpc.ErrorDesc(err))

	//iih.RunMetricsServer(context.Background())

	os.Exit(0)

	greeter := pb.NewGreeterClient(client)

	reply, err := greeter.SayHello(context.Background(), &pb.HelloRequest{Name: "joska"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(reply.Message)

	reply, err = greeter.SayHello(context.Background(), &pb.HelloRequest{Name: "pista"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(reply.Message)
}
