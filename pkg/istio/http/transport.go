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

package http

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/go-logr/logr"

	"github.com/cisco-open/libnasp/pkg/istio/discovery"
	"github.com/cisco-open/libnasp/pkg/network"
	"github.com/cisco-open/libnasp/pkg/tls/verify"
)

type istioHTTPRequestTransport struct {
	transport       http.RoundTripper
	tlsConfig       *tls.Config
	discoveryClient discovery.DiscoveryClient
	logger          logr.Logger
}

func NewIstioHTTPRequestTransport(transport http.RoundTripper, tlsConfig *tls.Config, discoveryClient discovery.DiscoveryClient, logger logr.Logger) http.RoundTripper {
	return &istioHTTPRequestTransport{
		transport:       transport,
		tlsConfig:       tlsConfig,
		discoveryClient: discoveryClient,
		logger:          logger,
	}
}

func (t *istioHTTPRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var tlsConfig *tls.Config

	if prop, _ := t.discoveryClient.GetHTTPClientPropertiesByHost(req.Context(), req.URL.Host); prop != nil { //nolint:nestif
		t.logger.V(3).Info("discovered overrides", "overrides", prop)
		if endpointAddr, err := prop.Address(); err != nil {
			return nil, err
		} else {
			req.URL.Host = endpointAddr.String()
			t.logger.V(3).Info("address override", "address", req.URL.Host)
		}

		if prop.UseTLS() {
			tlsConfig = t.tlsConfig
			tlsConfig.ServerName = prop.ServerName()
			req.URL.Scheme = "https"
			if !strings.Contains(req.URL.Host, ":") {
				req.URL.Host += ":80"
			}
			verify.SetCertVerifierToTLSConfig(prop, tlsConfig)
		}
	}

	t.discoveryClient.IncrementActiveRequestsCount(req.URL.Host)
	defer t.discoveryClient.DecrementActiveRequestsCount(req.URL.Host)

	dialerOpts := []network.DialerOption{
		network.DialerWithConnectionOptions(network.ConnectionWithCloserWrapper(t.discoveryClient.NewConnectionCloseWrapper())),
		network.DialerWithDialerWrapper(t.discoveryClient.NewDialWrapper()),
		network.DialerWithTLSConfig(tlsConfig),
	}

	return network.WrapHTTPTransport(t.transport, network.NewDialer(dialerOpts...)).RoundTrip(req)
}
