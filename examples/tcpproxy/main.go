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
	"net"
	"os"
	"time"

	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network/proxy"
)

var heimdallURL string
var serverAddress string
var clientAddress string

func init() {
	flag.StringVar(&heimdallURL, "heimdall-url", "https://localhost:16443/config", "Heimdall URL")
	flag.StringVar(&serverAddress, "server-address", ":5002", "tcp server address")
	flag.StringVar(&clientAddress, "client-address", ":5001", "tcp client address")
	klog.InitFlags(nil)
	flag.Parse()
}

func getIIH(ctx context.Context) (istio.IstioIntegrationHandler, error) {
	authToken := os.Getenv("NASP_AUTH_TOKEN")
	if authToken == "" {
		panic(errors.New("NASP_AUTH_TOKEN env var must be specified"))
	}

	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
	istioHandlerConfig.IstioCAConfigGetter = istio.IstioCAConfigGetterHeimdall(ctx, heimdallURL, authToken, "v1")

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

	iih, err := getIIH(ctx)
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	l, err = iih.GetTCPListener(l)
	if err != nil {
		panic(err)
	}

	netDialer := &net.Dialer{
		Timeout: time.Second * 10,
	}
	d, err := iih.GetTCPDialer(netDialer)
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

		proxy.New(localConnection, remoteConnection).Start()
	}
}
