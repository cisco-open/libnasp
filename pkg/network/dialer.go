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
	"reflect"

	"github.com/google/uuid"
)

type dialer struct {
	netDialer *net.Dialer
	tlsConfig *tls.Config

	dialWrapper       DialWrapper
	connectionOptions []ConnectionOption
}

type ConnectionDialer interface {
	connectionDialer
	DialTLSContext(ctx context.Context, network, address string) (net.Conn, error)
}

type DialWrapper interface {
	AddParentDialerWrapper(DialWrapper)
	BeforeDial(ctx context.Context, network, addr string) error
	AfterDial(ctx context.Context, conn net.Conn, network, addr string) error
}

type DialerOption func(*dialer)

func DialerWithNetDialer(nd *net.Dialer) DialerOption {
	return func(d *dialer) {
		d.netDialer = nd
	}
}

func DialerWithTLSConfig(config *tls.Config) DialerOption {
	return func(d *dialer) {
		d.tlsConfig = config
	}
}

func DialerWithDialerWrapper(w DialWrapper) DialerOption {
	return func(d *dialer) {
		if d.dialWrapper != nil && !reflect.DeepEqual(d.dialWrapper, w) {
			w.AddParentDialerWrapper(d.dialWrapper)
		}
		d.dialWrapper = w
	}
}

func DialerWithConnectionOptions(opts ...ConnectionOption) DialerOption {
	return func(d *dialer) {
		d.connectionOptions = opts
	}
}

type connectionDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

func NewDialer(opts ...DialerOption) ConnectionDialer {
	d := &dialer{}

	d.setOptions(opts...)

	return d
}

func NewDialerWithTLS(certFile, keyFile string, insecure bool, opts ...DialerOption) (ConnectionDialer, error) {
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	opts = append(opts, DialerWithTLSConfig(&tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		RootCAs:            certPool,
		InsecureSkipVerify: insecure,
	}))

	return NewDialer(opts...), nil
}

func NewDialerWithTLSConfig(config *tls.Config, opts ...DialerOption) ConnectionDialer {
	opts = append(opts, DialerWithTLSConfig(config))

	return NewDialer(opts...)
}

func (d *dialer) setOptions(opts ...DialerOption) {
	for _, o := range opts {
		o(d)
	}
}

func (d *dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if d.tlsConfig != nil {
		return d.DialTLSContext(ctx, network, addr)
	}

	return d.dialContext(d.getNetDialer(), ctx, network, addr)
}

func (d *dialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dialContext(d.getTLSDialer(), ctx, network, addr)
}

func (d *dialer) dialContext(dialer connectionDialer, ctx context.Context, network, addr string) (net.Conn, error) {
	if d.dialWrapper != nil {
		if err := d.dialWrapper.BeforeDial(ctx, network, addr); err != nil {
			return nil, err
		}
	}

	var state ConnectionState
	if s, ok := ConnectionStateFromContext(ctx); ok {
		state = s
	} else {
		state = NewConnectionStateWithID("connection-" + uuid.NewString())
	}

	innerCtx := ConnectionStateToContext(ctx, state)
	c, err := dialer.DialContext(innerCtx, network, addr)
	if err != nil {
		return nil, err
	}

	if s, ok := ConnectionStateHolderFromContext(ctx); ok && s.Get() == nil {
		s.Set(state)
	}

	opts := d.connectionOptions
	opts = append(opts, ConnectionWithOriginalAddress(addr), ConnectionWithState(state))

	wc := WrapConnection(c, opts...)

	if d.dialWrapper != nil {
		if err := d.dialWrapper.AfterDial(ctx, wc, network, addr); err != nil {
			return nil, err
		}
	}

	return wc, nil
}

func (d *dialer) getTLSDialer() *tls.Dialer {
	return &tls.Dialer{
		NetDialer: d.getNetDialer(),
		Config:    WrapTLSConfig(d.tlsConfig),
	}
}

func (d *dialer) getNetDialer() *net.Dialer {
	if d.netDialer != nil {
		return d.netDialer
	}
	return new(net.Dialer)
}
