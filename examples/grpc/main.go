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
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog/v2"

	"github.com/cisco-open/libnasp/examples/grpc/pb"
	"github.com/cisco-open/libnasp/pkg/istio"
	"github.com/cisco-open/libnasp/pkg/network"
	"github.com/cisco-open/libnasp/pkg/util"
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
	if s, ok := network.ConnectionStateFromContext(ctx); ok {
		util.PrintConnectionState(s, klog.Background())
	}

	log.Printf("Received: %v", in.GetName())

	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authToken := os.Getenv("NASP_AUTH_TOKEN")
	if authToken == "" {
		panic(errors.New("NASP_AUTH_TOKEN env var must be specified."))
	}

	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
	istioHandlerConfig.IstioCAConfigGetter = istio.IstioCAConfigGetterHeimdall(ctx, heimdallURL, authToken, "v1")

	iih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, klog.TODO())
	if err != nil {
		panic(err)
	}

	if err := iih.Run(ctx); err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	pb.RegisterGreeterServer(grpcServer, &greeterServer{})

	//////// standard HTTP library version with TLS

	err = iih.ListenAndServe(ctx, ":8082", grpcServer)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
