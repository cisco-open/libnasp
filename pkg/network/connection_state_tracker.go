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
	"fmt"
	"net"

	"github.com/go-logr/logr"
)

var discardLogger = logr.Discard()
var ConnectionStateTrackerLogger logr.Logger = discardLogger

type connectionStateTracker struct {
	connectionState            ConnectionState
	connWithTLSConnectionState ConnWithTLSConnectionState
}

func ConnectionStateFromNetConn(conn net.Conn) (ConnectionState, bool) {
	t := &connectionStateTracker{}

	state := t.getConnectionState(conn)

	if state != nil && t.connWithTLSConnectionState != nil {
		state.SetTLSConnectionState(t.connWithTLSConnectionState.ConnectionState())
	}

	return state, state != nil
}

func (t *connectionStateTracker) getTLSConnection(c net.Conn) net.Conn {
	if ConnectionStateTrackerLogger != discardLogger {
		ConnectionStateTrackerLogger.Info("check connection", "type", fmt.Sprintf("%T", c))
	}

	if _, ok := c.(ConnWithTLSConnectionState); ok {
		return c
	}

	if conn, ok := c.(interface {
		NetConn() net.Conn
	}); ok {
		return t.getTLSConnection(conn.NetConn())
	} else if conn, ok := c.(interface {
		GetConn() net.Conn
	}); ok {
		return t.getTLSConnection(conn.GetConn())
	}

	return nil
}

func (t *connectionStateTracker) getConnectionState(c net.Conn) ConnectionState {
	if ConnectionStateTrackerLogger != discardLogger {
		ConnectionStateTrackerLogger.Info("check connection", "type", fmt.Sprintf("%T", c))
	}

	if c, ok := c.(ConnWithTLSConnectionState); ok {
		t.connWithTLSConnectionState = c
		if t.connectionState != nil {
			return t.connectionState
		}
	}

	if conn, ok := c.(ConnectionStateGetter); ok && t.connectionState == nil {
		t.connectionState = conn.GetConnectionState()
	}

	if conn, ok := c.(interface {
		NetConn() net.Conn
	}); ok {
		return t.getConnectionState(conn.NetConn())
	} else if conn, ok := c.(interface {
		GetConn() net.Conn
	}); ok {
		return t.getConnectionState(conn.GetConn())
	}

	return t.connectionState
}
