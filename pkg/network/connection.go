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
	"crypto/tls"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/cisco-open/nasp/pkg/network/listener"
)

type connection struct {
	net.Conn

	id              string
	connectionState ConnectionState

	closeWrapper ConnectionCloseWrapper

	tlsConnectionStateSetter sync.Once
	tlsConn                  *tls.Conn

	originalAddress string
}

type ConnectionCloseWrapper interface {
	AddParentCloseWrapper(ConnectionCloseWrapper)
	BeforeClose(net.Conn) error
	AfterClose(net.Conn) error
}

type ConnWithTLSConnectionState interface {
	ConnectionState() tls.ConnectionState
}

type ConnectionOption func(*connection)

// Make sure the connection implements ConnectionStateGetter interface
var _ ConnectionStateGetter = &connection{}

func ConnectionWithCloserWrapper(closeWrapper ConnectionCloseWrapper) ConnectionOption {
	return func(c *connection) {
		if c.closeWrapper != nil && !reflect.DeepEqual(c.closeWrapper, closeWrapper) {
			closeWrapper.AddParentCloseWrapper(c.closeWrapper)
		}

		c.closeWrapper = closeWrapper
	}
}

func ConnectionWithState(state ConnectionState) ConnectionOption {
	return func(c *connection) {
		c.connectionState = state
	}
}

func ConnectionWithOriginalAddress(address string) ConnectionOption {
	return func(c *connection) {
		c.originalAddress = address
	}
}

func WrapConnection(conn net.Conn, opts ...ConnectionOption) net.Conn {
	if _, ok := conn.(ConnectionStateGetter); ok {
		return conn
	}

	c := &connection{
		Conn: conn,
	}

	c.setOptions(opts...)

	if c.connectionState == nil {
		c.connectionState = NewConnectionState()
	}

	c.connectionState.SetLocalAddr(conn.LocalAddr())
	c.connectionState.SetRemoteAddr(conn.RemoteAddr())

	c.id = uuid.NewString()

	return c
}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) SetTLSConn(conn *tls.Conn) {
	if c.tlsConn == nil {
		c.tlsConn = conn
	}
}

func (c *connection) GetConnectionState() ConnectionState {
	return c.connectionState
}

func (c *connection) NetConn() net.Conn {
	return c.Conn
}

func (c *connection) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)

	c.connectionState.SetTimeToFirstByte(time.Now())

	c.tlsConnectionStateSetter.Do(c.setTLSConnectionState)

	return
}

func (c *connection) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)

	c.connectionState.SetTimeToFirstByte(time.Now())

	c.tlsConnectionStateSetter.Do(c.setTLSConnectionState)

	return
}

func (c *connection) Close() error {
	if c.closeWrapper != nil {
		if err := c.closeWrapper.BeforeClose(c); err != nil {
			return err
		}
	}

	if err := c.Conn.Close(); err != nil {
		return err
	}

	if c.closeWrapper != nil {
		if err := c.closeWrapper.AfterClose(c); err != nil {
			return err
		}
	}

	return nil
}

func (c *connection) setOptions(opts ...ConnectionOption) {
	for _, opt := range opts {
		opt(c)
	}
}

func (c *connection) setTLSConnectionState() {
	if c.connectionState.GetTLSConnectionState().HandshakeComplete {
		return
	}

	if c.tlsConn != nil {
		c.connectionState.SetTLSConnectionStateAsync(func() tls.ConnectionState {
			cs := c.tlsConn.ConnectionState()
			c.tlsConn = nil

			return cs
		})

		return
	}

	t := &connectionStateTracker{}

	tlsConn := t.getTLSConnection(c)
	if tlsConn != nil {
		if csc, ok := tlsConn.(interface {
			ConnectionState() tls.ConnectionState
		}); ok {
			c.connectionState.SetTLSConnectionState(csc.ConnectionState())
		}
	}
}

func WrappedConnectionListener(l net.Listener) net.Listener {
	return listener.NewListenerWithConnectionWrapper(l, func(c net.Conn) net.Conn {
		return WrapConnection(c)
	})
}
