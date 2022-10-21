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
	"time"

	"k8s.io/klog/v2"

	istio_ca "wwwin-github.cisco.com/eti/nasp/pkg/ca/istio"
	"wwwin-github.cisco.com/eti/nasp/pkg/environment"
	"wwwin-github.cisco.com/eti/nasp/pkg/istio"
)

func init() {
	for _, env := range []string{
		"NASP_ISTIO_CA_ADDR=istiod-cp-v113x.istio-system.svc:15012",
		"NASP_ISTIO_VERSION=1.13.5",
		"NASP_ISTIO_REVISION=cp-v113x.istio-system",
		"NASP_TYPE=sidecar",
		"NASP_POD_NAME=tcp-client-efefefef-5989",
		"NASP_POD_NAMESPACE=demo",
		"NASP_POD_OWNER=kubernetes://apis/apps/v1/namespaces/default/deployments/tcp-client",
		"NASP_POD_SERVICE_ACCOUNT=tcp-client",
		"NASP_WORKLOAD_NAME=tcp-client",
		"NASP_MESH_ID=mesh1",
		"NASP_CLUSTER_ID=waynz0r-0626-01",
		"NASP_NETWORK=network2",
		"NASP_APP_CONTAINERS=alpine",
		"NASP_INSTANCE_IP=10.1.1.1",
		"NASP_LABELS=security.istio.io/tlsMode:istio, pod-template-hash:efefefef, service.istio.io/canonical-revision:latest, istio.io/rev:cp-v111x.istio-system, topology.istio.io/network:network1, k8s-app:tcp-client, service.istio.io/canonical-name:tcp-client, app:tcp-client, version:v11",
		"NASP_SEARCH_DOMAINS=demo.svc.cluster.local,svc.cluster.local,cluster.local",
	} {
		p := strings.Split(env, "=")
		if len(p) != 2 {
			continue
		}
		os.Setenv(p[0], p[1])
	}
}

func init() {
	klog.InitFlags(nil)
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
	iih, err := getIIH()
	if err != nil {
		panic(err)
	}

	iih.Run(context.Background())

	ctx := context.Background()

	d, err := iih.GetTCPDialer()
	if err != nil {
		panic(err)
	}

	conn, err := d.DialContext(ctx, "tcp", "tcp-echo-service:2701")
	if err != nil {
		panic(err)
	}

	go reader(ctx, conn)
	writer(ctx, conn)

	conn.Close()

	fmt.Println("conn closed")

	time.Sleep(time.Second * 30)
}

func reader(ctx context.Context, conn net.Conn) {
	defer func() {
		fmt.Println("reader closed")
	}()

	for {
		buff := make([]byte, 4096)
		n, err := conn.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		fmt.Printf("%s", buff[:n])
	}
}

func writer(ctx context.Context, conn net.Conn) {
	_, cf := context.WithCancel(ctx)
	defer func() {
		fmt.Println("writer closed")
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(string(text)) == "STOP" {
			cf()
			break
		}
		fmt.Fprintf(conn, text)
	}
}
