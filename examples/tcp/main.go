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
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/util"
)

var mode string
var heimdallURL string
var serverAddress string

func init() {
	flag.StringVar(&heimdallURL, "heimdall-url", "https://localhost:16443/config", "Heimdall URL")
	flag.StringVar(&mode, "mode", "server", "mode")
	flag.StringVar(&serverAddress, "server-address", "echo.testing:9000", "tcp server address")
	klog.InitFlags(nil)
	flag.Parse()
}

func getIIH(ctx context.Context) (istio.IstioIntegrationHandler, error) {
	authToken := os.Getenv("NASP_AUTH_TOKEN")
	if authToken == "" {
		panic(errors.New("NASP_AUTH_TOKEN env var must be specified."))
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
	if mode == "client" {
		client()
		return
	}

	server()
}

func server() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih, err := getIIH(ctx)
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
			defer func() {
				if s, ok := network.ConnectionStateFromNetConn(conn); ok {
					util.PrintConnectionState(s, klog.Background())
				}
				conn.Close()
			}()
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iih, err := getIIH(ctx)
	if err != nil {
		panic(err)
	}

	d, err := iih.GetTCPDialer()
	if err != nil {
		panic(err)
	}

	conn, err := d.DialContext(ctx, "tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	defer func() {
		if s, ok := network.ConnectionStateFromNetConn(conn); ok {
			util.PrintConnectionState(s, klog.Background())
		}
		conn.Close()
	}()

	for i := 0; i < 5; i++ {
		if err := sendReceive(conn, fmt.Sprintf("name %d", i)); err != nil {
			panic(err)
		}
	}
}

func sendReceive(conn net.Conn, msg string) error {
	bytes := []byte(msg + "\n")

	fmt.Printf("request: %s", bytes)

	if _, err := conn.Write(bytes); err != nil {
		return err
	}

	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	if err != nil {
		return err
	}

	fmt.Printf("response: %s", reply[:n])

	return nil
}
