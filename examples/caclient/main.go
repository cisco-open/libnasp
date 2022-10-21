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
	"flag"
	"fmt"
	"time"

	klog "k8s.io/klog/v2"

	"wwwin-github.cisco.com/eti/nasp/pkg/ca"
	istio_ca "wwwin-github.cisco.com/eti/nasp/pkg/ca/istio"
	"wwwin-github.cisco.com/eti/nasp/pkg/ca/selfsigned"
)

var (
	caType         string
	certTTL        time.Duration
	istioClusterID string
	istioRevision  string
	hostname       string
)

func init() {
	flag.StringVar(&caType, "ca", "istio", "Which CA implementation to use (selfsigned|istio)")
	flag.StringVar(&istioClusterID, "istio-cluster-id", "waynz0r-0626-01", "Istio cluster ID")
	flag.StringVar(&istioRevision, "istio-revision", "cp-v113x.istio-system", "Istio revision")
	flag.StringVar(&hostname, "hostname", "example.cisco.com", "Hostname of the certificate")
	flag.DurationVar(&certTTL, "cert-ttl", time.Hour*24, "Certificate TTLS")

	flag.Parse()
}

func main() {
	var caClient ca.Client
	var err error

	switch caType {
	case "selfsigned":
		caClient, err = selfsigned.NewSelfSignedCAClient(selfsigned.WithKeySize(2048))
		if err != nil {
			panic(err)
		}
	default:
		klog.Info("getting config for istio")

		istioCAClientConfig, err := istio_ca.GetIstioCAClientConfig(istioClusterID, istioRevision)
		if err != nil {
			panic(err)
		}

		caClient = istio_ca.NewIstioCAClient(istioCAClientConfig, klog.TODO())
	}

	cert, err := caClient.GetCertificate(hostname, certTTL)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", cert.GetPrivateKeyPEM())
	fmt.Printf("%s\n", cert.GetCertificatePEM())
}
