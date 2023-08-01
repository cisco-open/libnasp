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
	"sync/atomic"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/proxywasm"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	"github.com/cisco-open/nasp/pkg/proxywasm/middleware"
)

const (
	internalBufferSize      = 2048
	internalWriteBufferSize = 4096
)

type wrappedConn struct {
	net.Conn

	stream api.Stream

	readCacheBuffer api.IoBuffer
	readBuffer      api.IoBuffer
	writeBuffer     api.IoBuffer
	internalBuffer  []byte

	connectionInfoSet atomic.Bool

	writeMu sync.Mutex
	readMu  sync.Mutex

	closed bool
}

func NewWrappedConn(conn net.Conn, stream api.Stream) net.Conn {
	c := &wrappedConn{
		Conn: conn,

		stream:          stream,
		readCacheBuffer: new(bytes.Buffer),
		readBuffer:      new(bytes.Buffer),
		writeBuffer:     new(bytes.Buffer),
		internalBuffer:  make([]byte, internalBufferSize),
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
	if c.closed {
		return nil
	}

	// close underlying connection first
	if err := c.Conn.Close(); err != nil {
		return err
	}

	c.readMu.Lock()
	c.writeMu.Lock()
	defer c.readMu.Unlock()
	defer c.writeMu.Unlock()

	c.closed = true

	if err := c.stream.HandleTCPCloseConnection(c); err != nil {
		return err
	}

	return c.stream.Close()
}

func (c *wrappedConn) Read(b []byte) (int, error) {
	n, err := c.read(b)

	// try another read if the return len is zero without EOF
	// most probably a wasm module drained the read buffer
	if n == 0 && err == nil {
		return c.read(b)
	}

	return n, err
}

func (c *wrappedConn) read(b []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if c.closed {
		return 0, io.EOF
	}

	// flush the remaining read buffer
	x, err := c.readCacheBuffer.Read(b)
	if err != nil && !errors.Is(err, io.EOF) {
		return x, err
	}
	if x == len(b) {
		return x, nil
	}

	if diff := cap(b) - cap(c.internalBuffer); diff > 0 {
		c.internalBuffer = append(c.internalBuffer, make([]byte, diff)...)
	}

	done := false
	dataLength := 0
	endOfStream := false

	var netErr error
	var n int

	for !done {
		// read from the underlying reader
		n, netErr = c.Conn.Read(c.internalBuffer)
		c.setConnectionState()

		if netErr != nil && !errors.Is(netErr, io.EOF) { // or got a non-EOF error
			return x, netErr
		}

		endOfStream = errors.Is(netErr, io.EOF)
		if n > 0 {
			n, err = c.readBuffer.Write(c.internalBuffer[:n])
			if err != nil {
				return x, err
			}
			dataLength += n
		}

		switch c.stream.Direction() {
		case api.ListenerDirectionOutbound:
			if done, err = c.stream.HandleUpstreamData(c, dataLength, endOfStream); err != nil {
				return x, err
			}
		case api.ListenerDirectionUnspecified, api.ListenerDirectionInbound:
			if done, err = c.stream.HandleDownstreamData(c, dataLength, endOfStream); err != nil {
				return x, err
			}
		}

		if endOfStream || netErr != nil {
			done = true
		}
	}

	defer func() {
		// put the read data to the read buffer
		_, _ = io.Copy(c.readCacheBuffer, c.readBuffer)
	}()

	// do not return eof on readbuffer here
	if n, err := c.readBuffer.Read(b[x:]); err != nil && !errors.Is(err, io.EOF) {
		return x + n, err
	} else {
		x += n
	}

	return x, netErr
}

func (c *wrappedConn) Write(b []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.closed {
		return 0, io.EOF
	}

	c.setConnectionState()

	n, err := c.writeBuffer.Write(b)
	if err != nil {
		return n, err
	}

	done := false
	if c.stream.Direction() == api.ListenerDirectionOutbound {
		if done, err = c.stream.HandleDownstreamData(c, n, false); err != nil {
			return n, err
		}
	} else {
		if done, err = c.stream.HandleUpstreamData(c, n, false); err != nil {
			return n, err
		}
	}

	if !done {
		return n, nil
	}

	for {
		n1, err := c.writeBuffer.(*bytes.Buffer).WriteTo(c.Conn)
		if err != nil {
			return 0, err
		}
		if n1 == 0 {
			return n, nil
		}
	}
}

func (c *wrappedConn) setConnectionState() {
	if !c.connectionInfoSet.Load() {
		if connectionState, ok := network.ConnectionStateFromNetConn(c); ok {
			middleware.SetEnvoyConnectionInfo(connectionState, c.stream)
			c.connectionInfoSet.Store(true)
		}
	}
}
