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

package discovery

import (
	"context"
	"net"

	"github.com/cisco-open/nasp/pkg/network"
)

type DialWrapper = network.DialWrapper

type discoveryClientDialWrapper struct {
	discoveryClient   DiscoveryClient
	parentDialWrapper DialWrapper
}

func (d *xdsDiscoveryClient) NewDialWrapper() DialWrapper {
	return &discoveryClientDialWrapper{
		discoveryClient: d,
	}
}

func (c *discoveryClientDialWrapper) AddParentDialerWrapper(w DialWrapper) {
	c.parentDialWrapper = w
}

func (c *discoveryClientDialWrapper) BeforeDial(ctx context.Context, network, addr string) error {
	if c.parentDialWrapper != nil {
		return c.parentDialWrapper.BeforeDial(ctx, network, addr)
	}

	return nil
}

func (c *discoveryClientDialWrapper) AfterDial(ctx context.Context, conn net.Conn, network, addr string) error {
	c.discoveryClient.Logger().V(0).Info("increment active request count", "address", addr)
	c.discoveryClient.IncrementActiveRequestsCount(addr)

	if c.parentDialWrapper != nil {
		return c.parentDialWrapper.AfterDial(ctx, conn, network, addr)
	}

	return nil
}

type ConnectionCloseWrapper = network.ConnectionCloseWrapper

type discoveryClientCloser struct {
	discoveryClient    DiscoveryClient
	parentCloseWrapper ConnectionCloseWrapper
}

func (d *xdsDiscoveryClient) NewConnectionCloseWrapper() ConnectionCloseWrapper {
	return &discoveryClientCloser{
		discoveryClient: d,
	}
}

func (d *discoveryClientCloser) AddParentCloseWrapper(parentCloseWrapper ConnectionCloseWrapper) {
	d.parentCloseWrapper = parentCloseWrapper
}

func (c *discoveryClientCloser) AfterClose(conn net.Conn) error {
	if c.parentCloseWrapper != nil {
		return c.parentCloseWrapper.AfterClose(conn)
	}

	return nil
}

func (c *discoveryClientCloser) BeforeClose(conn net.Conn) error {
	address := conn.RemoteAddr().String()
	if res, ok := conn.(interface {
		GetOriginalAddress() string
	}); ok {
		address = res.GetOriginalAddress()
	}

	c.discoveryClient.Logger().V(0).Info("decrement active request count", "address", address)
	c.discoveryClient.DecrementActiveRequestsCount(address)

	if c.parentCloseWrapper != nil {
		return c.parentCloseWrapper.BeforeClose(conn)
	}

	return nil
}
