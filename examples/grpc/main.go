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
	"fmt"
	"log"
	"os"
	"strings"

	"k8s.io/klog/v2"

	"wwwin-github.cisco.com/eti/nasp/examples/grpc/pb"
	"wwwin-github.cisco.com/eti/nasp/pkg/istio"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func init() {
	for _, env := range []string{
		"NASP_ISTIO_CA_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_ISTIO_VERSION=1.13.5",
		"NASP_ISTIO_REVISION=cp-v113x.istio-system",
		"NASP_TYPE=sidecar",
		"NASP_POD_NAME=grpc-server",
		"NASP_POD_NAMESPACE=default",
		"NASP_POD_OWNER=kubernetes://apis/apps/v1/namespaces/default/deployments/alpine",
		"NASP_WORKLOAD_NAME=alpine",
		"NASP_MESH_ID=mesh1",
		"NASP_CLUSTER_ID=waynz0r-0626-01",
		"NASP_APP_CONTAINERS=alpine",
		"NASP_INSTANCE_IP=10.20.4.75",
		"NASP_LABELS=app:grpc-server,version:v99,security.istio.io/tlsMode:istio, pod-template-hash:efefefef, service.istio.io/canonical-revision:latest, istio.io/rev:cp-v111x.istio-system, topology.istio.io/network:network1, k8s-app:alpine, service.istio.io/canonical-name:alpine",
	} {
		p := strings.Split(env, "=")
		if len(p) != 2 {
			continue
		}
		os.Setenv(p[0], p[1])
	}
}

type greeterServer struct {
	pb.UnimplementedGreeterServer
}

func (gs *greeterServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func interceptorFunc(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("Intercepted: %+v", req)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("Metadata: %v", md)
	}
	resp, err := handler(ctx, req)
	if err != nil {
		return resp, err
	}

	grpc.SendHeader(ctx, metadata.Pairs("c", "d"))
	grpc.SetTrailer(ctx, metadata.Pairs("t1", "v1", "Grpc-New-Status", "0"))

	fmt.Println("server interceptor run")

	return resp, err
}

func main() {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterRemote,
	}, klog.TODO())
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih.Run(ctx)

	logInterceptor := grpc.UnaryInterceptor(interceptorFunc)

	grpcServer := grpc.NewServer(logInterceptor)
	reflection.Register(grpcServer)
	pb.RegisterGreeterServer(grpcServer, &greeterServer{})

	//////// standard HTTP library version with TLS

	err = iih.ListenAndServe(context.Background(), ":8080", grpcServer)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	// lis, err := net.Listen("tcp", "localhost:1234")
	// if err != nil {
	// 	log.Fatalf("failed to listen: %v", err)
	// }

	// err = grpcServer.Serve(lis)
	// if err != nil {
	// 	log.Fatalf("failed to serve: %v", err)
	// }
}
