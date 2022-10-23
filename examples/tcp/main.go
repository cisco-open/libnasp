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
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"

	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
)

var mode string
var heimdallURL string

func init() {
	flag.StringVar(&heimdallURL, "heimdall-url", "https://localhost:16443/config", "Heimdall URL")
	flag.StringVar(&mode, "mode", "server", "mode")
	klog.InitFlags(nil)
	flag.Parse()
}

func getIIH() (*istio.IstioIntegrationHandler, error) {
	istioHandlerConfig := &istio.IstioIntegrationHandlerConfig{
		MetricsAddress:      ":15090",
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, "test-grpc-16362813-F46B-41AC-B191-A390DB1F6BDF", "16362813-F46B-41AC-B191-A390DB1F6BDF", "v1"),
	}

	iih, err := istio.NewIstioIntegrationHandler(istioHandlerConfig, klog.TODO())
	if err != nil {
		return nil, err
	}

	return iih, nil
}

func main() {
	if mode == "client" {
		client()
		return
	}

	server()
}

func server() {
	iih, err := getIIH()
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", ":10000")
	if err != nil {
		panic(err)
	}

	l, err = iih.GetTCPListener(l)
	if err != nil {
		panic(err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go func(conn net.Conn) {
			defer conn.Close()
			reader := bufio.NewReader(conn)
			for {
				// read client request data
				bytes, err := reader.ReadBytes(byte('\n'))
				if err != nil {
					if err != io.EOF {
						fmt.Println("failed to read data, err:", err)
					}
					return
				}

				fmt.Printf("request: %s", bytes)

				// prepend prefix and send as response
				line := fmt.Sprintf("%s %s", "<< ", bytes)
				fmt.Printf("response: %s", line)
				conn.Write([]byte(line))
			}
		}(c)
	}
}

func client() {
	iih, err := getIIH()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	d, err := iih.GetTCPDialer()
	if err != nil {
		panic(err)
	}

	conn, err := d.DialContext(ctx, "tcp", "localhost:10000")
	if err != nil {
		panic(err)
	}

	bytes := []byte("hello\n")
	fmt.Printf("request: %s", bytes)

	_, err = conn.Write(bytes)
	if err != nil {
		panic(err)
	}

	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	if err != nil {
		panic(err)
	}

	conn.Close()

	fmt.Printf("response: %s", reply[:n])
}
