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
	"net"

	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network/proxy"
)

var heimdallURL string
var serverAddress string
var clientAddress string
var naspEnabled bool
var bufferSize int

func init() {
	flag.StringVar(&serverAddress, "server-address", ":5002", "tcp server address")
	flag.StringVar(&clientAddress, "client-address", ":5001", "tcp client address")
	flag.BoolVar(&naspEnabled, "nasp-enabled", true, "whether Nasp is enabled")
	flag.IntVar(&bufferSize, "buffer-size", 4096, "length of buffer in bytes to read or write")
	klog.InitFlags(nil)
	flag.Parse()
}

func getIIH(ctx context.Context) (istio.IstioIntegrationHandler, error) {
	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
	istioHandlerConfig.Enabled = naspEnabled
	iih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, klog.TODO())
	if err != nil {
		return nil, err
	}

	if err := iih.Run(context.Background()); err != nil {
		return nil, err
	}

	return iih, nil
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	iih, err := getIIH(ctx)
	if err != nil {
		panic(err)
	}

	l, err = iih.GetTCPListener(l)
	if err != nil {
		panic(err)
	}

	d, err := iih.GetTCPDialer()
	if err != nil {
		panic(err)
	}

	for {
		localConnection, err := l.Accept()
		if err != nil {
			panic(err)
		}
		klog.Info("Accepting connection on:", serverAddress)

		remoteConnection, err := d.DialContext(ctx, "tcp", clientAddress)
		if err != nil {
			panic(err)
		}

		go proxy.New(localConnection, remoteConnection, proxy.WithBufferSize(bufferSize)).Start()
	}
}
