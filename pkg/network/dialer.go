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

package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
)

type Dialer struct {
	NetDialer *net.Dialer

	TLSConfig *tls.Config
}

type ConnectionDialer interface {
	connectionDialer
	DialTLSContext(ctx context.Context, network, address string) (net.Conn, error)
}

type connectionDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

func NewDialer() ConnectionDialer {
	return &Dialer{}
}

func NewDialerWithTLS(certFile, keyFile string, insecure bool) (ConnectionDialer, error) {
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	return &Dialer{
		TLSConfig: &tls.Config{
			Certificates:       []tls.Certificate{tlsCert},
			RootCAs:            certPool,
			InsecureSkipVerify: insecure,
		},
	}, nil
}

func NewDialerWithTLSConfig(config *tls.Config) ConnectionDialer {
	return &Dialer{
		TLSConfig: config,
	}
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dialContext(d.netDialer(), ctx, network, addr)
}

func (d *Dialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dialContext(d.tlsDialer(), ctx, network, addr)
}

func (d *Dialer) dialContext(dialer connectionDialer, ctx context.Context, network, addr string) (net.Conn, error) {
	c, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	if wc, ok := WrappedConnectionFromContext(ctx); ok {
		wc.SetNetConn(c)

		return wc, nil
	}

	return NewWrappedConn(c), nil
}

func (d *Dialer) tlsDialer() *tls.Dialer {
	return &tls.Dialer{
		NetDialer: d.NetDialer,
		Config:    WrapTLSConfig(d.TLSConfig),
	}
}

func (d *Dialer) netDialer() *net.Dialer {
	if d.NetDialer != nil {
		return d.NetDialer
	}
	return new(net.Dialer)
}
