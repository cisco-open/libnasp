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

package tls

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
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

type UnifiedListener interface {
	net.Listener

	SetTLSMode(mode TLSMode)
	SetTLSClientAuthMode(mode tls.ClientAuthType)
}

type unifiedListener struct {
	net.Listener

	tlsConfig *tls.Config
	mode      TLSMode
}

func NewUnifiedListener(listener net.Listener, tlsConfig *tls.Config, mode TLSMode) UnifiedListener {
	return &unifiedListener{
		Listener: listener,

		tlsConfig: tlsConfig,
		mode:      mode,
	}
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

	if l.mode == TLSModeDisabled {
		return c, nil
	}

	if l.mode == TLSModeStrict {
		return tls.Server(c, l.tlsConfig), nil
	}

	// buffer reads on our conn
	conn := &unifiedConn{
		Conn: c,
		buf:  bufio.NewReader(c),
	}

	// inspect the first few bytes
	hdr, err := conn.buf.Peek(3)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if isTLSHandhsake(hdr) {
		return tls.Server(conn, l.tlsConfig), nil
	}

	return conn, nil
}
