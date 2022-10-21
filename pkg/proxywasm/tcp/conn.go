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

package tcp

import (
	"net"

	"mosn.io/proxy-wasm-go-host/proxywasm/common"

	"wwwin-github.cisco.com/eti/nasp/pkg/network"
	"wwwin-github.cisco.com/eti/nasp/pkg/proxywasm"
	"wwwin-github.cisco.com/eti/nasp/pkg/proxywasm/api"
	"wwwin-github.cisco.com/eti/nasp/pkg/proxywasm/middleware"
)

type wrappedConn struct {
	net.Conn

	stream api.Stream

	readBuffer  common.IoBuffer
	writeBuffer common.IoBuffer

	connectionInfoSet bool
}

func NewWrappedConn(conn net.Conn, stream api.Stream) net.Conn {
	c := &wrappedConn{
		Conn: conn,

		stream:      stream,
		readBuffer:  proxywasm.NewBuffer([]byte{}),
		writeBuffer: proxywasm.NewBuffer([]byte{}),
	}

	if stream.Direction() == api.ListenerDirectionOutbound {
		stream.Set("upstream.data", c.readBuffer)
		stream.Set("downstream.data", c.writeBuffer)
	} else {
		stream.Set("downstream.data", c.readBuffer)
		stream.Set("upstream.data", c.writeBuffer)
	}

	return c
}

func (c *wrappedConn) NetConn() net.Conn {
	return c.Conn
}

func (c *wrappedConn) Close() error {
	if err := c.stream.HandleTCPCloseConnection(c); err != nil {
		return err
	}

	if err := c.stream.Close(); err != nil {
		return err
	}

	return c.Conn.Close()
}

func (c *wrappedConn) Read(b []byte) (n int, err error) {
	buf := make([]byte, len(b))
	n, err = c.Conn.Read(buf)
	if err != nil {
		return
	}

	if !c.connectionInfoSet {
		if connection, ok := network.WrappedConnectionFromNetConn(c); ok {
			middleware.SetEnvoyConnectionInfo(connection, c.stream)
			c.connectionInfoSet = true
		}
	}

	c.readBuffer.Drain(c.readBuffer.Len())
	if n, err := c.readBuffer.Write(buf[:n]); err != nil {
		return n, err
	}

	if c.stream.Direction() == api.ListenerDirectionOutbound {
		if err := c.stream.HandleUpstreamData(c, n, c.readBuffer.Bytes()); err != nil {
			return 0, err
		}
	} else {
		if err := c.stream.HandleDownstreamData(c, n, c.readBuffer.Bytes()); err != nil {
			return 0, err
		}
	}

	n = copy(b, c.readBuffer.Bytes())

	return n, err
}

func (c *wrappedConn) Write(b []byte) (n int, err error) {
	if !c.connectionInfoSet {
		if connection, ok := network.WrappedConnectionFromNetConn(c); ok {
			middleware.SetEnvoyConnectionInfo(connection, c.stream)
			c.connectionInfoSet = true
		}
	}

	if n, err := c.writeBuffer.Write(b); err != nil {
		return n, err
	}

	if c.stream.Direction() == api.ListenerDirectionOutbound {
		if err := c.stream.HandleDownstreamData(c, c.writeBuffer.Len(), c.writeBuffer.Bytes()); err != nil {
			return 0, err
		}
	} else {
		if err := c.stream.HandleUpstreamData(c, c.writeBuffer.Len(), c.writeBuffer.Bytes()); err != nil {
			return 0, err
		}
	}

	n, err = c.Conn.Write(c.writeBuffer.Bytes())
	if err != nil {
		return
	}

	c.writeBuffer.Drain(n)

	return n, err
}
