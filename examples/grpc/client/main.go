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

	"google.golang.org/grpc"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/examples/grpc/pb"
	"github.com/cisco-open/nasp/pkg/istio"
)

var heimdallURL string

func init() {
	flag.StringVar(&heimdallURL, "heimdall-url", "https://localhost:16443/config", "Heimdall URL")
	klog.InitFlags(nil)
	flag.Parse()
}

func main() {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, "test-grpc-16362813-F46B-41AC-B191-A390DB1F6BDF", "16362813-F46B-41AC-B191-A390DB1F6BDF", "v1"),
	}, klog.TODO())
	if err != nil {
		log.Fatal(err)
	}

	grpcDialOptions, err := iih.GetGRPCDialOptions()
	if err != nil {
		log.Fatal(err)
	}

	client, err := grpc.Dial(
		"localhost:8082",
		grpcDialOptions...,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih.Run(ctx)

	greeter := pb.NewGreeterClient(client)

	reply, err := greeter.SayHello(context.Background(), &pb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(reply.Message)
}
