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
	"crypto/tls"
	_ "embed"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"k8s.io/klog/v2"

	istio_ca "wwwin-github.cisco.com/eti/nasp/pkg/ca/istio"
	"wwwin-github.cisco.com/eti/nasp/pkg/wasm/stream"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <server|tls-server> <wasm module>\n", os.Args[0])
		os.Exit(0)
	}

	wasm, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}

	r := wazero.NewRuntimeWithConfig(context.Background(), wazero.NewRuntimeConfig().WithWasmCore2())
	defer r.Close(ctx)

	var useTls bool
	if os.Args[1] == "tls-server" {
		useTls = true
	}

	var istioCAClientConfig istio_ca.IstioCAClientConfig
	if useTls {
		fmt.Println("getting config for istio")
		istioCAClientConfig, err = istio_ca.GetIstioCAClientConfig("waynz0r-0626-01", "cp-v113x.istio-system")
		if err != nil {
			panic(err)
		}
	}

	if err := runTCPServer(ctx, r, wasm, &istioCAClientConfig); err != nil {
		panic(err)
	}
}

func runTCPServer(ctx context.Context, r wazero.Runtime, wasmPayload []byte, istioCAClientConfig *istio_ca.IstioCAClientConfig) error {
	code, err := r.CompileModule(ctx, wasmPayload, wazero.NewCompileConfig())
	if err != nil {
		return err
	}

	port := 8000

	var listen net.Listener
	if istioCAClientConfig != nil {
		port = 8443
		c := istio_ca.NewIstioCAClient(*istioCAClientConfig, klog.TODO())

		listen, err = tls.Listen("tcp", fmt.Sprintf(":%d", port), &tls.Config{
			GetCertificate: func(clientHelloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := c.GetCertificate("joskapista", time.Minute*60)
				if err != nil {
					return nil, err
				}
				return cert.GetTLSCertificate(), nil
			},
		})
	} else {
		listen, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	if err != nil {
		return err
	}
	defer listen.Close()

	fmt.Printf("started (ssl=%t) port(=%d)\n", istioCAClientConfig != nil, port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer func() {
				conn.Close()
			}()

			ctx, ctxCF := context.WithCancel(ctx)
			defer ctxCF()
			ns := r.NewNamespace(ctx)

			if _, err := wasi_snapshot_preview1.NewBuilder(r).Instantiate(ctx, ns); err != nil {
				log.Println(err)
				return
			}

			streamer := stream.New(r)
			socketFS := stream.NewSocketFS(streamer)
			socketFS.AddSocket("1", conn)
			if _, err := streamer.AddHandler("socket-accept", socketFS).Instantiate(ctx, ns); err != nil {
				log.Println(err)
				return
			}

			config := wazero.NewModuleConfig().
				WithStdout(os.Stdout).WithStderr(os.Stderr).
				WithArgs(append([]string{"wasi"}, os.Args[1:]...)...).
				WithName(time.Now().String())

			if _, err = ns.InstantiateModule(ctx, code, config); err != nil {
				if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
					fmt.Fprintf(os.Stderr, "exit_code: %d\n", exitErr.ExitCode())
				} else if !ok {
					log.Panicln(err)
				}
			}

			fmt.Println("finished")
		}()
	}
}
