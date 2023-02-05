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
	"fmt"
	"net"

	"github.com/go-logr/logr"
)

type contextKey struct {
	name string
}

var connectionContextKey = contextKey{"network.connection"}
var ConnectionTrackerLogger logr.Logger = logr.Discard()

type ConnectionHolder interface {
	SetConn(net.Conn)
	NetConn() net.Conn
}

type connectionHolder struct {
	net.Conn
}

func (h *connectionHolder) SetConn(conn net.Conn) {
	h.Conn = conn
}

func (h *connectionHolder) NetConn() net.Conn {
	return h.Conn
}

type connectionTracker struct {
	logger logr.Logger

	connection                 Connection
	connWithTLSConnectionState ConnWithTLSConnectionState
}

func WrappedConnectionToContext(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, connectionContextKey, &connectionHolder{conn})
}

func ConnectionHolderFromContext(ctx context.Context) (ConnectionHolder, bool) {
	if c, ok := ctx.Value(connectionContextKey).(ConnectionHolder); ok {
		return c, true
	}

	return nil, false
}

func SetConnectionToContextConnectionHolder(ctx context.Context, conn net.Conn) bool {
	if h, ok := ConnectionHolderFromContext(ctx); ok {
		h.SetConn(conn)

		return true
	}

	return false
}

func WrappedConnectionFromContext(ctx context.Context) (Connection, bool) {
	if c, ok := ctx.Value(connectionContextKey).(ConnectionHolder); ok {
		return WrappedConnectionFromNetConn(c.NetConn())
	}

	return nil, false
}

func WrappedConnectionFromNetConn(conn net.Conn) (Connection, bool) {
	t := &connectionTracker{
		logger: ConnectionTrackerLogger,
	}

	connection := t.getConnection(conn)

	if connection != nil {
		connection.SetConnWithTLSConnectionState(t.connWithTLSConnectionState)
	}

	return connection, connection != nil
}

func (t *connectionTracker) getConnection(c net.Conn) Connection {
	t.logger.Info("check connection", "type", fmt.Sprintf("%T", c))

	if c, ok := c.(ConnWithTLSConnectionState); ok {
		t.connWithTLSConnectionState = c
		if t.connection != nil {
			return t.connection
		}
	}

	if conn, ok := c.(Connection); ok && t.connection == nil {
		t.connection = conn
	}

	if conn, ok := c.(interface {
		NetConn() net.Conn
	}); ok {
		return t.getConnection(conn.NetConn())
	} else if conn, ok := c.(interface {
		GetConn() net.Conn
	}); ok {
		return t.getConnection(conn.GetConn())
	}

	return t.connection
}
