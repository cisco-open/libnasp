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

	"github.com/cisco-open/libnasp/pkg/istio/discovery"
	"github.com/cisco-open/libnasp/pkg/network"
	"github.com/cisco-open/libnasp/pkg/network/pool"
	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
	"github.com/cisco-open/libnasp/pkg/proxywasm/tcp"
	"github.com/cisco-open/libnasp/pkg/tls/verify"
)

type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

type tcpDialer struct {
	streamHandler   api.StreamHandler
	tlsConfig       *tls.Config
	discoveryClient discovery.DiscoveryClient
	dialer          *net.Dialer

	connectionPoolRegistry pool.Registry
}

func NewTCPDialer(streamHandler api.StreamHandler, tlsConfig *tls.Config, discoveryClient discovery.DiscoveryClient, dialer *net.Dialer) Dialer {
	return &tcpDialer{
		streamHandler:   streamHandler,
		tlsConfig:       tlsConfig,
		discoveryClient: discoveryClient,
		dialer:          dialer,

		connectionPoolRegistry: pool.NewSyncMapPoolRegistry(pool.SyncMapPoolRegistryWithLogger(streamHandler.Logger())),
	}
}

func (d *tcpDialer) DialContext(ctx context.Context, _net string, address string) (net.Conn, error) {
	var tlsConfig *tls.Config

	prop, _ := d.discoveryClient.GetTCPClientPropertiesByHost(context.Background(), address)
	if prop != nil {
		if prop.UseTLS() {
			tlsConfig = d.tlsConfig.Clone()
			tlsConfig.ServerName = prop.ServerName()
			verify.SetCertVerifierToTLSConfig(prop, tlsConfig)
		}
		if endpointAddr, err := prop.Address(); err != nil {
			return nil, err
		} else {
			address = endpointAddr.String()
		}
	}

	opts := []network.DialerOption{
		network.DialerWithConnectionOptions(network.ConnectionWithCloserWrapper(d.discoveryClient.NewConnectionCloseWrapper())),
		network.DialerWithDialerWrapper(d.discoveryClient.NewDialWrapper()),
		network.DialerWithNetDialer(d.dialer),
		network.DialerWithTLSConfig(tlsConfig),
	}

	connectionDialer := network.NewDialer(opts...)

	f := func() (net.Conn, error) {
		var conn net.Conn
		var err error

		conn, err = connectionDialer.DialContext(ctx, _net, address)
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

	var cp pool.Pool
	if !d.connectionPoolRegistry.HasPool(address) {
		d.streamHandler.Logger().V(2).Info("new pool created", "address", address)
		p, err := pool.NewChannelPool(
			f,
			pool.ChannelPoolWithLogger(d.streamHandler.Logger()),
		)
		if err != nil {
			return nil, err
		}

		d.connectionPoolRegistry.AddPool(address, p)
	}

	cp, err := d.connectionPoolRegistry.GetPool(address)
	if err != nil {
		return nil, err
	}

	d.streamHandler.Logger().V(2).Info("pool stat", "address", address, "len", cp.Len())

	return cp.Get()
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
