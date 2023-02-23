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
	"time"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(ctx, heimdallURL, "ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNkltcElUbk50ZFVsaVkyUkhkRTQzWHpKalIwTlRWMHBQTFU5RFgyWnlNM1JFY2pJM2VYUjNibkZOVTNNaWZRLmV5SmhkV1FpT2xzaWFYTjBhVzh0WTJFaVhTd2laWGh3SWpveE5qYzNNalEzT0RReUxDSnBZWFFpT2pFMk56Y3hOakUwTkRJc0ltbHpjeUk2SW1oMGRIQnpPaTh2YTNWaVpYSnVaWFJsY3k1a1pXWmhkV3gwTG5OMll5NWpiSFZ6ZEdWeUxteHZZMkZzSWl3aWEzVmlaWEp1WlhSbGN5NXBieUk2ZXlKdVlXMWxjM0JoWTJVaU9pSm9aV2x0WkdGc2JDSXNJbk5sY25acFkyVmhZMk52ZFc1MElqcDdJbTVoYldVaU9pSmtaV1poZFd4MElpd2lkV2xrSWpvaU1UQXdPVGt6TlRVdFptRmxNaTAwWVdVd0xUZzFPVGd0TVdOak5XVTNaV1l5WVRWa0luMTlMQ0p1WW1ZaU9qRTJOemN4TmpFME5ESXNJbk4xWWlJNkluTjVjM1JsYlRwelpYSjJhV05sWVdOamIzVnVkRHBvWldsdFpHRnNiRHBrWldaaGRXeDBJbjAuV2NLbjVMY051YmNhMEYxRXBYVU1RV0RsRmNqOUQxeVBfSmFpUTUyRTJNSmhCQUd0MU1GYV96QlM3dVVDSExmQmctc0E0SUstYUYzWW5rYnFJdzlzNEpIa252b0hIaWlYY1l6SEZJUnBMZzg3NEhTSmpfOWt6SERHYkNoSVhfSy1GSnlMYUNHSzlnSWFZOFpvUEFRZWJnVDZ5QVRQbVhsN2ZrRUsxblBWdFpFS3B0LU1sSXVNZDBHWDVIcUVZQWNGeGZpTTZQU1hBaWstVU9OUXhMWnBSRDl3c1pfM0k4YXRjZi1UcEdkaW5QTXlPb01DWW10cFZNb3U3VFZFd3RodmF3eWZPVXFWTzN5NXRGQXFQSlZjaHllTHVRU20zcHc5a1FFRE40NFdHNU1KRDFITmgxeldBaDN6RW9ma2pOSXFNUnNlZGRNY2J3VjBYY1Y4aWtlOVRB", "v1"),
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

	iih.Run(ctx)

	func() {
		defer cancel()
		defer client.Close()

		for i := 0; i < 10; i++ {
			reply, err := pb.NewGreeterClient(client).SayHello(ctx, &pb.HelloRequest{Name: "world"})
			if err != nil {
				log.Fatal(err)
			}

			log.Println(reply.Message)
		}
	}()

	time.Sleep(time.Millisecond * 100)
}
