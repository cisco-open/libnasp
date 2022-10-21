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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"

	istio_ca "wwwin-github.cisco.com/eti/nasp/pkg/ca/istio"
	"wwwin-github.cisco.com/eti/nasp/pkg/ca/selfsigned"
	"wwwin-github.cisco.com/eti/nasp/pkg/environment"
	"wwwin-github.cisco.com/eti/nasp/pkg/network"
	ltls "wwwin-github.cisco.com/eti/nasp/pkg/tls"
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

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "joska", "joska")
	ctx = context.WithValue(ctx, "pista", "pista")
	ctx = context.WithValue(ctx, "geza", "geza")
	ctx = context.WithValue(ctx, "asdsa", "asda")

	fmt.Printf("%#v\n", ctx.Value("pista"))
}

func main123() {
	if len(os.Args) > 1 && os.Args[1] == "client" {
		httpclient()
		// tcpclient()
		return
	}

	// httpserver()
	// tcpserver()
}

func tcpclient() {
	caClient, err := selfsigned.NewSelfSignedCAClient(selfsigned.WithKeySize(2048))
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caClient.GetCAPem())

	ctx := context.Background()
	ctx = network.NewConnectionToContext(ctx)

	var d network.ConnectionDialer
	// d, err = network.NewDialerWithTLS("/Users/wayne/tmp/cert.pem", "/Users/wayne/tmp/key.pem", true)
	// if err != nil {
	// 	panic(err)
	// }

	d = network.NewDialerWithTLSConfig(&tls.Config{
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := caClient.GetCertificate("spiffe://cluster.local/joska/pista2", time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	})
	if err != nil {
		panic(err)
	}

	conn, err := d.DialTLSContext(ctx, "tcp", "192.168.255.31:10000")
	if err != nil {
		panic(err)
	}

	_, err = conn.Write([]byte("hello\n"))
	if err != nil {
		panic(err)
	}

	reply := make([]byte, 1024)

	n, err := conn.Read(reply)
	if err != nil {
		panic(err)
	}

	if connection, ok := network.WrappedConnectionFromNetConn(conn); ok {
		printconnection(connection)
	}

	fmt.Println(string(reply[:n]))
}

func httpclient() {
	e, _ := environment.GetIstioEnvironment("NASP_")
	ic, _ := istio_ca.GetIstioCAClientConfig(e.ClusterID, e.IstioRevision)

	caClient := istio_ca.NewIstioCAClient(ic, klog.TODO())

	// caClient, err := selfsigned.NewSelfSignedCAClient(selfsigned.WithKeySize(2048))
	// if err != nil {
	// 	panic(err)
	// }

	var certPool *x509.CertPool
	if scp, err := x509.SystemCertPool(); err == nil {
		certPool = scp
	} else {
		certPool = x509.NewCertPool()
	}
	certPool.AppendCertsFromPEM(caClient.GetCAPem())

	tlsConfig := &tls.Config{
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := caClient.GetCertificate("spiffe://cluster.local/joska/pista2", time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			fmt.Printf("%s\n", string(cert.GetCertificatePEM()))

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
		ServerName:         "outbound_.8080_._.echo.demo.svc.cluster.local",
		NextProtos: []string{
			"istio",
		},
	}
	// tlsConfig = nil

	// d, err := network.NewDialerWithTLS("/Users/wayne/tmp/cert.pem", "/Users/wayne/tmp/key.pem", true)
	// if err != nil {
	// 	panic(err)
	// }

	d := network.NewDialerWithTLSConfig(tlsConfig)
	// d = &network.Dialer{}

	c := http.Client{
		Transport: network.WrapHTTPTransport(http.DefaultTransport, d),
	}

	// r, err := c.Get("https://192.168.255.31:8080/")
	// r, err = c.Get("https://httpbin.org/headers")
	r, err := c.Get("http://echo:8080")
	if err != nil {
		panic(err)
	}

	if connection, ok := network.WrappedConnectionFromContext(r.Request.Context()); ok {
		printconnection(connection)
	}

	b, err := httputil.DumpResponse(r, true)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

func printconnection(connection network.Connection) {
	fmt.Printf("local addr: %s\n", connection.LocalAddr().String())
	fmt.Printf("remote addr: %s\n", connection.RemoteAddr().String())

	if cert := connection.GetLocalCertificate(); cert != nil {
		fmt.Printf("LOCALURI %#v\n", cert.GetFirstURI())
	}

	if cert := connection.GetPeerCertificate(); cert != nil {
		fmt.Printf("PEERURI %#v\n", cert.GetFirstURI())
		fmt.Printf("PEERSUB %#v\n", cert.GetSubject())
		fmt.Printf("PEERDNS %#v\n", cert.GetFirstDNSName())
	}

	fmt.Printf("TTFB: %#v\n", connection.GetTimeToFirstByte().Format(time.RFC3339Nano))
}

func tcpserver() {
	caClient, err := selfsigned.NewSelfSignedCAClient(selfsigned.WithKeySize(2048))
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", "192.168.255.31:10000")
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caClient.GetCAPem())

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequestClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := caClient.GetCertificate("spiffe://cluster.local/ns/custom/sa/customsa", time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}

	// l = tls.NewListener(network.WrapListener(l), network.WrapTLSConfig(tlsConfig))

	l = ltls.NewUnifiedListener(network.WrapListener(l), network.WrapTLSConfig(tlsConfig), ltls.TLSModePermissive)

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

				if connection, ok := network.WrappedConnectionFromNetConn(conn); ok {
					printconnection(connection)
				}

				// prepend prefix and send as response
				line := fmt.Sprintf("%s %s", "<< ", bytes)
				fmt.Printf("response: %s", line)
				conn.Write([]byte(line))
			}
		}(c)
	}
}

func httpserver() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("joska2")
		if connection, ok := network.WrappedConnectionFromContext(r.Context()); ok {
			printconnection(connection)
		}
	})

	// server := &http.Server{
	// 	// TLSConfig: tlsConfig,
	// 	Handler: mux,
	// }
	// // ioutil.Discard.Write([]byte(fmt.Sprintf("%#v", server)))

	server := network.WrapHTTPServer(&http.Server{
		Handler: mux,
	})

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	caClient, err := selfsigned.NewSelfSignedCAClient(selfsigned.WithKeySize(2048))
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caClient.GetCAPem())

	server.ServeWithTLSConfig(listener, &tls.Config{
		ClientAuth: tls.RequestClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := caClient.GetCertificate("spiffe://cluster.local/ns/joska/sa/pista", time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2"},
	})
	// server.ServeTLS(listener, "/Users/wayne/tmp/cert.pem", "/Users/wayne/tmp/key.pem")
}
