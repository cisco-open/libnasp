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
	"errors"
	"net"
	"time"
)

type Connection interface {
	net.Conn

	NetConn() net.Conn
	SetNetConn(net.Conn)
	GetTimeToFirstByte() time.Time
	GetLocalCertificate() Certificate
	GetPeerCertificate() Certificate
	SetLocalCertificate(cert *tls.Certificate)
	GetConnectionState() *tls.ConnectionState
	SetConnWithTLSConnectionState(ConnWithTLSConnectionState)
	SetOptions(opts ...WrappedConnectionOption)
}

type ConnectionCloseWrapper interface {
	AddParentCloseWrapper(ConnectionCloseWrapper)
	BeforeClose(net.Conn) error
	AfterClose(net.Conn) error
}

type ConnWithTLSConnectionState interface {
	ConnectionState() tls.ConnectionState
}

type wrappedListener struct {
	net.Listener
}

func NewWrappedListener(l net.Listener) net.Listener {
	return &wrappedListener{
		Listener: l,
	}
}

func (l *wrappedListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewWrappedConnection(c), nil
}

type wrappedConn struct {
	net.Conn

	connWithTLSConnectionState ConnWithTLSConnectionState
	timeToFirstByte            time.Time
	localCertificate           *x509.Certificate
	closeWrapper               ConnectionCloseWrapper
	beforeCloseRun             bool
	originalAddress            string
}

type WrappedConnectionOption func(*wrappedConn)

func WrappedConnectionWithCloserWrapper(closeWrapper ConnectionCloseWrapper) WrappedConnectionOption {
	return func(wc *wrappedConn) {
		if wc.closeWrapper != nil {
			closeWrapper.AddParentCloseWrapper(wc.closeWrapper)
		}
		wc.closeWrapper = closeWrapper
	}
}

func WrappedConnectionWithOriginalAddress(address string) WrappedConnectionOption {
	return func(wc *wrappedConn) {
		wc.originalAddress = address
	}
}

func NewWrappedConnection(conn net.Conn, opts ...WrappedConnectionOption) Connection {
	wc := &wrappedConn{
		Conn: conn,
	}

	wc.SetOptions(opts...)

	return wc
}

func NewConnectionToContext(ctx context.Context) context.Context {
	return WrappedConnectionToContext(ctx, NewWrappedConnection(NewNilConn()))
}

func (c *wrappedConn) NetConn() net.Conn {
	return c.Conn
}

func (c *wrappedConn) SetNetConn(conn net.Conn) {
	c.Conn = conn
}

func (c *wrappedConn) Read(b []byte) (n int, err error) {
	if c.timeToFirstByte.IsZero() {
		c.timeToFirstByte = time.Now()
	}

	return c.Conn.Read(b)
}

func (c *wrappedConn) Write(b []byte) (n int, err error) {
	if c.timeToFirstByte.IsZero() {
		c.timeToFirstByte = time.Now()
	}

	return c.Conn.Write(b)
}

func (c *wrappedConn) SetOptions(opts ...WrappedConnectionOption) {
	for _, opt := range opts {
		opt(c)
	}
}

func (c *wrappedConn) GetOriginalAddress() string {
	return c.originalAddress
}

func (c *wrappedConn) Close() error {
	if c.closeWrapper != nil && !c.beforeCloseRun {
		if err := c.closeWrapper.BeforeClose(c); err != nil {
			return err
		}

		c.beforeCloseRun = true
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

func (c *wrappedConn) GetTimeToFirstByte() time.Time {
	return c.timeToFirstByte
}

func (c *wrappedConn) GetLocalCertificate() Certificate {
	if c.localCertificate == nil {
		return nil
	}

	return &certificate{
		Certificate: c.localCertificate,
	}
}

func (c *wrappedConn) GetPeerCertificate() Certificate {
	cs := c.GetConnectionState()
	if cs == nil {
		return nil
	}

	if len(cs.PeerCertificates) < 1 {
		return nil
	}

	return &certificate{
		Certificate: cs.PeerCertificates[0],
	}
}

func (c *wrappedConn) SetLocalCertificate(cert *tls.Certificate) {
	if cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
		c.localCertificate = cert
	}
}

func (c *wrappedConn) SetConnWithTLSConnectionState(conn ConnWithTLSConnectionState) {
	c.connWithTLSConnectionState = conn
}

func (c *wrappedConn) GetConnectionState() *tls.ConnectionState {
	if c.connWithTLSConnectionState != nil {
		cs := c.connWithTLSConnectionState.ConnectionState()
		return &cs
	}

	return nil
}

var (
	ErrNilRead  = errors.New("cannot read from nilConn")
	ErrNilWrite = errors.New("cannot write to nilConn")
)

func NewNilConn() net.Conn {
	return &nilConn{}
}

type nilConn struct{}
type nilAddr struct{}

func (a *nilAddr) String() string {
	return "0.0.0.0:0"
}

func (a *nilAddr) Network() string {
	return "tcp"
}

func (c *nilConn) Read(b []byte) (n int, err error) {
	return 0, ErrNilRead
}

func (c *nilConn) Write(b []byte) (n int, err error) {
	return 0, ErrNilWrite
}

func (c *nilConn) Close() error {
	return nil
}

func (c *nilConn) LocalAddr() net.Addr {
	return &nilAddr{}
}

func (c *nilConn) RemoteAddr() net.Addr {
	return &nilAddr{}
}

func (c *nilConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *nilConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *nilConn) SetWriteDeadline(t time.Time) error {
	return nil
}
