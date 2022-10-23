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

	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/examples/grpc/pb"
	"github.com/cisco-open/nasp/pkg/istio"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var heimdallURL string

func init() {
	flag.StringVar(&heimdallURL, "heimdall-url", "https://localhost:16443/config", "Heimdall URL")
	klog.InitFlags(nil)
	flag.Parse()
}

type greeterServer struct {
	pb.UnimplementedGreeterServer
}

func (gs *greeterServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, "test-grpc-16362813-F46B-41AC-B191-A390DB1F6BDF", "16362813-F46B-41AC-B191-A390DB1F6BDF", "v1"),
	}, klog.TODO())
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih.Run(ctx)

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	pb.RegisterGreeterServer(grpcServer, &greeterServer{})

	//////// standard HTTP library version with TLS

	err = iih.ListenAndServe(context.Background(), ":8082", grpcServer)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
