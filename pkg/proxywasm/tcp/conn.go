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
	"bytes"
	"io"
	"net"
	"sync"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/proxywasm"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	"github.com/cisco-open/nasp/pkg/proxywasm/middleware"
)

type wrappedConn struct {
	net.Conn

	stream api.Stream

	readCacheBuffer api.IoBuffer
	readBuffer      api.IoBuffer
	writeBuffer     api.IoBuffer
	internalBuffer  []byte

	connectionInfoSet bool

	writeMu sync.Mutex
	readMu  sync.Mutex
}

func NewWrappedConn(conn net.Conn, stream api.Stream) net.Conn {
	c := &wrappedConn{
		Conn: conn,

		stream:          stream,
		readCacheBuffer: new(bytes.Buffer),
		readBuffer:      new(bytes.Buffer),
		writeBuffer:     new(bytes.Buffer),
		internalBuffer:  make([]byte, 1024),
	}

	if stream.Direction() == api.ListenerDirectionOutbound {
		proxywasm.UpstreamDataProperty(stream).Set(c.readBuffer)
		proxywasm.DownstreamDataProperty(stream).Set(c.writeBuffer)
	} else {
		proxywasm.DownstreamDataProperty(stream).Set(c.readBuffer)
		proxywasm.UpstreamDataProperty(stream).Set(c.writeBuffer)
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
	c.readMu.Lock()
	defer c.readMu.Unlock()

	// flush the remaining read buffer
	x, err := c.readCacheBuffer.Read(b)
	if err != nil && !errors.Is(err, io.EOF) {
		return x, err
	}
	if x == len(b) {
		return x, nil
	}

	// read from the underlying reader
	n, err := c.Conn.Read(c.internalBuffer)
	if n == 0 { // return if there is no more data
		return n, err
	}
	if err != nil && !errors.Is(err, io.EOF) { // or got a non-EOF error
		return n, err
	}

	n, err = c.readBuffer.Write(c.internalBuffer[:n])
	if err != nil {
		return n, err
	}

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

	defer func() {
		// put the read data to the read buffer
		_, _ = io.Copy(c.readCacheBuffer, c.readBuffer)
	}()

	if n, err := c.readBuffer.Read(b[x:]); err != nil {
		return n, err
	} else {
		x += n
	}

	return x, err
}

func (c *wrappedConn) Write(b []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

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
		if err := c.stream.HandleDownstreamData(c, n); err != nil {
			return 0, err
		}
	} else {
		if err := c.stream.HandleUpstreamData(c, n); err != nil {
			return 0, err
		}
	}

	for {
		n1, err := io.Copy(c.Conn, io.LimitReader(c.writeBuffer, int64(cap(b))))
		if err != nil {
			return 0, err
		}
		if n1 == 0 {
			return n, nil
		}
	}
}
