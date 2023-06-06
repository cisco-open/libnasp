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

package listener

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type TLSMode string

const (
	TLSModeStrict     TLSMode = "STRICT"
	TLSModePermissive TLSMode = "PERMISSIVE"
	TLSModeDisabled   TLSMode = "DISABLED"
)

const tlsHandshakeRecord = 22

func isTLSHandhsake(b []byte) bool {
	if b[0] != tlsHandshakeRecord {
		return false
	}
	tlsVersion := binary.BigEndian.Uint16(b[1:])
	switch tlsVersion {
	case tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12, tls.VersionTLS13:
		return true
	default:
		return false
	}
}

// here's a buffered conn for peeking into the connection
type unifiedConn struct {
	net.Conn
	buf *bufio.Reader
}

func (c *unifiedConn) Read(b []byte) (int, error) {
	return c.buf.Read(b)
}

func (c *unifiedConn) NetConn() net.Conn {
	return c.Conn
}

func (c *unifiedConn) SetTLSConn(conn *tls.Conn) {
	if c, ok := c.Conn.(interface {
		SetTLSConn(conn *tls.Conn)
	}); ok {
		c.SetTLSConn(conn)
	}
}

type UnifiedListener interface {
	net.Listener

	SetTLSMode(mode TLSMode)
	SetTLSClientAuthMode(mode tls.ClientAuthType)
}

type unifiedListener struct {
	net.Listener

	tlsConfig *tls.Config
	mode      TLSMode

	tlsConnCreator    func(net.Conn, *tls.Config) *tls.Conn
	connectionWrapper ConnectionWrapper
}

type UnifiedListenerOption func(*unifiedListener)

func UnifiedListenerWithTLSConnectionCreator(f func(net.Conn, *tls.Config) *tls.Conn) UnifiedListenerOption {
	return func(l *unifiedListener) {
		l.tlsConnCreator = f
	}
}

func UnifiedListenerWithConnectionWrapper(wrapper ConnectionWrapper) UnifiedListenerOption {
	return func(l *unifiedListener) {
		l.connectionWrapper = wrapper
	}
}

func NewUnifiedListener(listener net.Listener, tlsConfig *tls.Config, mode TLSMode, opts ...UnifiedListenerOption) UnifiedListener {
	l := &unifiedListener{
		Listener: listener,

		tlsConfig: tlsConfig,
		mode:      mode,
	}

	for _, o := range opts {
		o(l)
	}

	if l.tlsConnCreator == nil {
		l.tlsConnCreator = tls.Server
	}

	return l
}

func (l *unifiedListener) SetTLSMode(mode TLSMode) {
	l.mode = mode
}

func (l *unifiedListener) SetTLSClientAuthMode(mode tls.ClientAuthType) {
	l.tlsConfig.ClientAuth = mode
}

func (l *unifiedListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if l.mode == TLSModePermissive {
		// buffer reads on our conn
		var conn net.Conn
		conn = &unifiedConn{
			Conn: c,
			buf:  bufio.NewReader(c),
		}

		// inspect the first few bytes
		hdr, err := conn.(*unifiedConn).buf.Peek(3)
		// close connection in case of EOF without reporting error
		if errors.Is(err, io.EOF) {
			conn.Close()
			return conn, nil
		}
		if err != nil {
			conn.Close()
			return nil, err
		}

		if l.connectionWrapper != nil {
			conn = l.connectionWrapper(conn)
		}

		if isTLSHandhsake(hdr) {
			return l.tlsConnCreator(conn, l.tlsConfig), nil
		}

		return conn, nil
	}

	if l.connectionWrapper != nil {
		c = l.connectionWrapper(c)
	}

	if l.mode == TLSModeDisabled {
		return c, nil
	}

	return l.tlsConnCreator(c, l.tlsConfig), nil
}
