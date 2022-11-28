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

package nasp

import (
	"context"
	"net"

	"github.com/cisco-open/nasp/pkg/istio"
	itcp "github.com/cisco-open/nasp/pkg/istio/tcp"
	"k8s.io/klog/v2"
)

var logger = klog.NewKlogr()

type TCPListener struct {
	iih      *istio.IstioIntegrationHandler
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
}

type Connection struct {
	conn net.Conn
}

func NewTCPListener(heimdallURL, clientID, clientSecret string) (*TCPListener, error) {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, clientID, clientSecret, "v1"),
		PushgatewayConfig: &istio.PushgatewayConfig{
			Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
			UseUniqueIDLabel: true,
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", ":10000")
	if err != nil {
		return nil, err
	}

	listener, err = iih.GetTCPListener(listener)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	return &TCPListener{
		iih:      iih,
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
	}, nil

}

func (l *TCPListener) Accept() (*Connection, error) {
	c, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &Connection{c}, nil
}

func (c *Connection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *Connection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

type TCPDialer struct {
	iih    *istio.IstioIntegrationHandler
	dialer itcp.Dialer
	ctx    context.Context
	cancel context.CancelFunc
}

func NewTCPDialer(heimdallURL, clientID, clientSecret string) (*TCPDialer, error) {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, clientID, clientSecret, "v1"),
		PushgatewayConfig: &istio.PushgatewayConfig{
			Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
			UseUniqueIDLabel: true,
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	dialer, err := iih.GetTCPDialer()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	return &TCPDialer{
		iih:    iih,
		dialer: dialer,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (d *TCPDialer) Dial() (*Connection, error) {
	c, err := d.dialer.DialContext(context.Background(), "tcp", "localhost:10000")
	if err != nil {
		return nil, err
	}
	return &Connection{c}, nil
}
