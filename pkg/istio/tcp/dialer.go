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

package istio

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/cisco-open/nasp/pkg/istio/discovery"
	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	"github.com/cisco-open/nasp/pkg/proxywasm/tcp"
)

type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

type tcpDialer struct {
	streamHandler   api.StreamHandler
	tlsConfig       *tls.Config
	discoveryClient discovery.DiscoveryClient
}

func NewTCPDialer(streamHandler api.StreamHandler, tlsConfig *tls.Config, discoveryClient discovery.DiscoveryClient) Dialer {
	return &tcpDialer{
		streamHandler:   streamHandler,
		tlsConfig:       tlsConfig,
		discoveryClient: discoveryClient,
	}
}

func (d *tcpDialer) DialContext(ctx context.Context, _net string, address string) (net.Conn, error) {
	ctx = network.NewConnectionToContext(ctx)

	tlsConfig := d.tlsConfig.Clone()

	prop, _ := d.discoveryClient.GetTCPClientPropertiesByHost(context.Background(), address)
	if prop != nil {
		if !prop.UseTLS() {
			tlsConfig = nil
		} else {
			tlsConfig.ServerName = prop.ServerName()
		}
		if endpointAddr, err := prop.Address(); err != nil {
			return nil, err
		} else {
			address = endpointAddr.String()
		}
	}

	useTLS := tlsConfig != nil

	opts := []network.DialerOption{
		network.DialerWithWrappedConnectionOptions(network.WrappedConnectionWithCloserWrapper(d.discoveryClient.NewConnectionCloseWrapper())),
		network.DialerWithDialerWrapper(d.discoveryClient.NewDialWrapper()),
	}

	var connectionDialer network.ConnectionDialer
	if useTLS {
		connectionDialer = network.NewDialerWithTLSConfig(tlsConfig, opts...)
	} else {
		connectionDialer = network.NewDialer(opts...)
	}

	var conn net.Conn
	var err error

	if useTLS {
		conn, err = connectionDialer.DialContext(ctx, _net, address)
	} else {
		conn, err = connectionDialer.DialTLSContext(ctx, _net, address)
	}
	if err != nil {
		return nil, err
	}

	wrappedConn, err := d.handleNewConnection(conn, prop)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return wrappedConn, nil
}

func (d *tcpDialer) handleNewConnection(conn net.Conn, prop discovery.ClientProperties) (net.Conn, error) {
	stream, err := d.streamHandler.NewStream(api.ListenerDirectionOutbound)
	if err != nil {
		return nil, err
	}

	conn = tcp.NewWrappedConn(conn, stream)

	if prop != nil {
		if md, ok := prop.Metadata()["cluster_metadata"].(map[string]interface{}); ok {
			stream.Set("cluster_metadata.filter_metadata", md)
		}
		if n, ok := prop.Metadata()["cluster_name"]; ok {
			stream.Set("cluster_name", n)
		}
	}

	stream.Set("response.flags", 0)

	if err := stream.HandleTCPNewConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}
