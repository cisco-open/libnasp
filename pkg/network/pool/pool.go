// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
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

package pool

import (
	"errors"
	"net"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// Pool interface describes a pool implementation.
type Pool interface {
	// Get returns a connection from the pool or create new one at request.
	Get() (net.Conn, error)

	// Put puts the connection back to the pool.
	Put(net.Conn) error

	// Len returns the current number of connections of the pool.
	Len() int

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close() error

	// IsClosed returns whether the pool is closed.
	IsClosed() bool
}

// Registry interface describes a pool registry.
type Registry interface {
	// GetPool returns a registered pool for the given id or an error.
	GetPool(id string) (Pool, error)

	// HasPool returns whether a pool is registered with the given id.
	HasPool(id string) bool

	// AddPool registers the given pool for the given id.
	AddPool(id string, pool Pool)

	// RemovePool removes the pool registered for the given id.
	RemovePool(id string)

	// Len returns the current number of registered pools.
	Len() int

	// Close closes the registry and also every pool registered into it.
	// The registry is unusable after this call.
	Close() error

	// IsClosed returns whether the registry is closed.
	IsClosed() bool
}

// Connection interface describes a pool connection.
type Connection interface {
	net.Conn

	// Discard closes the underlying net.Conn and mark the connection closed.
	Discard() error

	// IsClosed returns whether the connection is closed.
	IsClosed() bool

	// Acquire signals when the connection is served from the connection pool.
	Acquire()

	// Release signals when the connection is put back into the connection pool.
	Release()
}
