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
	"os"
	"strings"

	"k8s.io/klog/v2"

	istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
	"github.com/cisco-open/nasp/pkg/environment"
	"github.com/cisco-open/nasp/pkg/istio"
)

func init() {
	for _, env := range []string{
		"NASP_ISTIO_CA_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_ISTIO_VERSION=1.13.5",
		"NASP_ISTIO_REVISION=cp-v113x.istio-system",
		"NASP_TYPE=sidecar",
		"NASP_POD_NAME=alpine-efefefef-f5wwf",
		"NASP_POD_NAMESPACE=default",
		"NASP_POD_OWNER=kubernetes://apis/apps/v1/namespaces/default/deployments/alpine",
		"NASP_POD_SERVICE_ACCOUNT=alpine",
		"NASP_WORKLOAD_NAME=alpine",
		"NASP_MESH_ID=mesh1",
		"NASP_CLUSTER_ID=waynz0r-0626-01",
		"NASP_NETWORK=network2",
		"NASP_APP_CONTAINERS=alpine",
		"NASP_INSTANCE_IP=10.20.4.75",
		"NASP_LABELS=security.istio.io/tlsMode:istio, pod-template-hash:efefefef, service.istio.io/canonical-revision:latest, istio.io/rev:cp-v111x.istio-system, topology.istio.io/network:network1, k8s-app:alpine, service.istio.io/canonical-name:alpine, app:alpine, version:v11",
		"NASP_SEARCH_DOMAINS=demo.svc.cluster.local,svc.cluster.local,cluster.local",
	} {
		p := strings.Split(env, "=")
		if len(p) != 2 {
			continue
		}
		os.Setenv(p[0], p[1])
	}
}

var mode string

func init() {
	klog.InitFlags(nil)
	flag.StringVar(&mode, "mode", "server", "mode")
	flag.Parse()
}

func getIIH() (*istio.IstioIntegrationHandler, error) {
	istioHandlerConfig := &istio.IstioIntegrationHandlerConfig{
		MetricsAddress: ":15090",
		UseTLS:         true,
		IstioCAConfigGetter: func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
			return istio_ca.GetIstioCAClientConfig(e.ClusterID, e.IstioRevision)
		},
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
				line := fmt.Sprintf("%s [%s]", "<< ", bytes)
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

	conn, err := d.DialContext(ctx, "tcp", "192.168.255.31:10000")
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

	fmt.Printf("response: %s\n", reply[:n])
}
