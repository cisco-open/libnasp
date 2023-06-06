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
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type contextKey struct {
	name string
}

var connectionStateContextKey = contextKey{"network.connection.state"}

type ConnectionState interface { //nolint:interfacebloat
	GetTimeToFirstByte() time.Time
	SetTimeToFirstByte(time.Time)
	ResetTimeToFirstByte()
	GetLocalCertificate() Certificate
	SetLocalCertificate(cert tls.Certificate)
	GetPeerCertificate() Certificate
	GetTLSConnectionState() tls.ConnectionState
	SetTLSConnectionState(tls.ConnectionState)
	SetTLSConnectionStateAsync(func() tls.ConnectionState)
	LocalAddr() net.Addr
	SetLocalAddr(net.Addr)
	RemoteAddr() net.Addr
	SetRemoteAddr(net.Addr)
	GetOriginalAddress() string
	ID() string
}

type ConnectionStateGetter interface {
	GetConnectionState() ConnectionState
}

type ConnectionStateHolder interface {
	Get() ConnectionState
	Set(ConnectionState)
	ID() string
}

type connectionStateHolder struct {
	connectionState ConnectionState
	id              string
}

func (h *connectionStateHolder) ID() string {
	return h.id
}

func (h *connectionStateHolder) Get() ConnectionState {
	return h.connectionState
}

func (h *connectionStateHolder) Set(stat ConnectionState) {
	h.connectionState = stat
}

func ConnectionStateToContextFromNetConn(ctx context.Context, conn net.Conn) context.Context {
	if state, ok := ConnectionStateFromNetConn(conn); ok {
		state.SetTimeToFirstByte(time.Time{})

		return ConnectionStateToContext(ctx, state)
	}

	return ctx
}

// ConnectionStateToContext creates a new context object with the provided
// ConnectionState object added to it as a value using the
// connectionStateContextKey constant as the key.
func ConnectionStateToContext(ctx context.Context, state ConnectionState) context.Context {
	return context.WithValue(ctx, connectionStateContextKey, &connectionStateHolder{
		connectionState: state,
		id:              uuid.NewString(),
	})
}

// ConnectionStateFromContext extracts the ConnectionState object from the
// provided context and returns it, along with a boolean indicating whether the
// extraction was successful. If the ConnectionState object is not found in the
// context or is nil, the function returns false.
func ConnectionStateFromContext(ctx context.Context) (ConnectionState, bool) {
	if holder, ok := ctx.Value(connectionStateContextKey).(ConnectionStateHolder); ok && holder.Get() != nil {
		return holder.Get(), true
	}

	return nil, false
}

// NewConnectionStateHolderToContext creates a new connectionStateHolder object
// with a new UUID string as the ID and adds it to the provided context using
// the connectionStateContextKey constant as the key.
func NewConnectionStateHolderToContext(ctx context.Context) context.Context {
	return ConnectionStateHolderToContext(ctx, &connectionStateHolder{id: uuid.NewString()})
}

func ConnectionStateHolderToContext(ctx context.Context, holder ConnectionStateHolder) context.Context {
	return context.WithValue(ctx, connectionStateContextKey, holder)
}

func ConnectionStateHolderFromContext(ctx context.Context) (ConnectionStateHolder, bool) {
	if holder, ok := ctx.Value(connectionStateContextKey).(ConnectionStateHolder); ok {
		return holder, true
	}

	return nil, false
}

func ConnectionStateFromHTTPRequest(req *http.Request) (ConnectionState, bool) {
	if state, ok := ConnectionStateFromContext(req.Context()); ok {
		if req.TLS != nil {
			state.SetTLSConnectionState(*req.TLS)
		}

		return state, true
	}

	return nil, false
}

func NewConnectionState() ConnectionState {
	return NewConnectionStateWithID(uuid.NewString())
}

func NewConnectionStateWithID(id string) ConnectionState {
	s := &connectionState{
		id: id,
		connectionState: tls.ConnectionState{
			HandshakeComplete: false,
		},
	}

	return s
}

type connectionState struct {
	timeToFirstByte  time.Time
	localCertificate *x509.Certificate
	connectionState  tls.ConnectionState
	localAddr        net.Addr
	remoteAddr       net.Addr
	originalAddress  string

	id string

	mux  sync.RWMutex
	mux2 sync.Mutex
}

func (s *connectionState) ID() string {
	return s.id
}

func (s *connectionState) GetOriginalAddress() string {
	return s.originalAddress
}

func (s *connectionState) GetTimeToFirstByte() time.Time {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.timeToFirstByte
}

func (s *connectionState) ResetTimeToFirstByte() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.timeToFirstByte = time.Time{}
}

func (s *connectionState) SetTimeToFirstByte(t time.Time) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.timeToFirstByte.IsZero() {
		s.timeToFirstByte = t
	}
}

func (s *connectionState) GetLocalCertificate() Certificate {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if s.localCertificate == nil {
		return nil
	}

	return &certificate{
		Certificate: s.localCertificate,
	}
}

func (s *connectionState) SetLocalCertificate(cert tls.Certificate) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
		s.localCertificate = cert
	}
}

func (s *connectionState) GetPeerCertificate() Certificate {
	cs := s.GetTLSConnectionState()
	if len(cs.PeerCertificates) < 1 {
		return nil
	}

	return &certificate{
		Certificate: cs.PeerCertificates[0],
	}
}

func (s *connectionState) GetTLSConnectionState() tls.ConnectionState {
	s.mux.RLock()
	s.mux2.Lock()
	defer func() {
		s.mux.RUnlock()
		s.mux2.Unlock()
	}()

	return s.connectionState
}

func (s *connectionState) SetTLSConnectionStateAsync(f func() tls.ConnectionState) {
	s.mux2.Lock()
	go func() {
		defer s.mux2.Unlock()
		s.connectionState = f()
	}()
}

func (s *connectionState) SetTLSConnectionState(state tls.ConnectionState) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.connectionState = state
}

func (s *connectionState) LocalAddr() net.Addr {
	return s.localAddr
}

func (s *connectionState) SetLocalAddr(addr net.Addr) {
	s.localAddr = addr
}

func (s *connectionState) RemoteAddr() net.Addr {
	return s.remoteAddr
}

func (s *connectionState) SetRemoteAddr(addr net.Addr) {
	s.remoteAddr = addr
}
