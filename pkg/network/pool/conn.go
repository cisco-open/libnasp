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
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// poolConnection is a wrapper around net.Conn.
type poolConnection struct {
	net.Conn
	mu   sync.RWMutex
	pool Pool

	idleTimeout time.Duration
	idleTimer   *time.Timer

	failed bool
	closed atomic.Bool
	used   atomic.Bool
}

// Close puts the given connects back to the pool instead of closing it.
func (c *poolConnection) Close() error {
	if c.failed {
		return c.Discard()
	}

	return c.pool.Put(c)
}

func (c *poolConnection) NetConn() net.Conn {
	return c.Conn
}

func (c *poolConnection) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)

	// return closed error instead of timeout
	// if timeout is caused by closing connection
	// and release it back to the pool
	if errors.Is(err, os.ErrDeadlineExceeded) && !c.used.Load() {
		if e, ok := err.(*net.OpError); ok {
			e.Err = net.ErrClosed
		}

		return
	}

	if err != nil {
		c.failed = true
	}

	return
}

// Discard closes the underlying net.Conn and mark the connection closed.
func (c *poolConnection) Discard() error {
	if c.IsClosed() {
		return nil
	}

	c.stopIdleTimer()
	c.closed.Store(true)

	return c.Conn.Close()
}

// IsClosed checks whether the connection is closed already.
func (c *poolConnection) IsClosed() bool {
	return c.closed.Load()
}

// Acquire is used to signal when the connection is served from the connection pool.
func (c *poolConnection) Acquire() {
	c.used.Store(true)

	// set deadline to zero
	_ = c.Conn.SetDeadline(time.Time{})

	c.stopIdleTimer()
}

// Release is used to signal when the connection is put back into the connection pool.
func (c *poolConnection) Release() {
	c.used.Store(false)

	if c.idleTimeout > 0 {
		c.setIdleTimer(c.idleTimeout)
	}

	// set deadline to past to unblock read/writes
	_ = c.Conn.SetDeadline(time.Now().Add(-time.Millisecond))
}

func (c *poolConnection) setIdleTimer(idleTimeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.idleTimer != nil {
		c.idleTimer.Reset(idleTimeout)
	} else {
		c.idleTimer = time.AfterFunc(idleTimeout, func() {
			_ = c.Discard()
		})
	}
}

func (c *poolConnection) stopIdleTimer() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.idleTimer == nil {
		return
	}

	c.idleTimer.Stop()
	c.idleTimer = nil
}
