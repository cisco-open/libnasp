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
	"errors"
	"io"
	"net"

	"github.com/banzaicloud/proxy-wasm-go-host/pkg/utils"
	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	"github.com/cisco-open/nasp/pkg/proxywasm/middleware"
)

type wrappedConn struct {
	net.Conn

	stream api.Stream

	readCacheBuffer    api.IoBuffer
	readBuffer         api.IoBuffer
	writeBuffer        api.IoBuffer
	internalReadBuffer []byte

	connectionInfoSet bool
}

func NewWrappedConn(conn net.Conn, stream api.Stream) net.Conn {
	c := &wrappedConn{
		Conn: conn,

		stream:          stream,
		readCacheBuffer: utils.NewIoBufferBytes([]byte{}),
		readBuffer:      utils.NewIoBufferBytes([]byte{}),
		writeBuffer:     utils.NewIoBufferBytes([]byte{}),

		internalReadBuffer: make([]byte, 16384),
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

func (c *wrappedConn) Read(b []byte) (int, error) {
	// flush the remaining read buffer
	if c.readCacheBuffer.Len() > 0 {
		x := copy(b, c.readCacheBuffer.Bytes())
		c.readCacheBuffer.Drain(x)

		return x, nil
	}

	// read from the underlying reader
	n, err := c.Conn.Read(c.internalReadBuffer)
	if n == 0 { // return if there is no more data
		return n, err
	}
	if err != nil && !errors.Is(err, io.EOF) { // or got a non-EOF error
		return n, err
	}

	if n, err := c.readBuffer.Write(c.internalReadBuffer[:n]); err == nil {
		switch c.stream.Direction() {
		case api.ListenerDirectionOutbound:
			if err := c.stream.HandleUpstreamData(c, n); err != nil {
				return 0, err
			}
		case api.ListenerDirectionUnspecified, api.ListenerDirectionInbound:
			if err := c.stream.HandleDownstreamData(c, n); err != nil {
				return 0, err
			}
		}
	}

	// put the read data to the read buffer
	if n, err := c.readCacheBuffer.Write(c.readBuffer.Bytes()); err != nil {
		return n, err
	} else {
		c.readBuffer.Drain(n)
	}

	// copy from read buffer to the upstream buffer and drain the read buffer until the read len
	n = copy(b, c.readCacheBuffer.Bytes())
	c.readCacheBuffer.Drain(n)

	return n, err
}

func (c *wrappedConn) Write(b []byte) (int, error) {
	if !c.connectionInfoSet {
		if connection, ok := network.WrappedConnectionFromNetConn(c); ok {
			middleware.SetEnvoyConnectionInfo(connection, c.stream)
			c.connectionInfoSet = true
		}
	}

	n, err := c.writeBuffer.Write(b)
	if err != nil {
		return n, err
	}

	if c.stream.Direction() == api.ListenerDirectionOutbound {
		if err := c.stream.HandleDownstreamData(c, c.writeBuffer.Len()); err != nil {
			return 0, err
		}
	} else {
		if err := c.stream.HandleUpstreamData(c, c.writeBuffer.Len()); err != nil {
			return 0, err
		}
	}

	l, err := c.Conn.Write(c.writeBuffer.Bytes())
	if err != nil {
		return n, err
	}
	c.writeBuffer.Drain(l)

	return n, nil
}
