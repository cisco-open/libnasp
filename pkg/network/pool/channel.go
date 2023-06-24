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
	"net"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

var logger = klog.Background()

var (
	ErrInvalidCapacity = errors.New("invalid capacity settings")
	ErrNilConnection   = errors.New("could not put nil connection")
)

var (
	DefaultInitialCap            uint32        = 0
	DefaultMaxCap                uint32        = 20
	DefaultConnectionIdleTimeout time.Duration = time.Second * 90
	DefaultIdleTimeout           time.Duration = time.Second * 90
)

type ConnectionChannel chan Connection

// channelPool implements the Pool interface based on buffered channels.
type channelPool struct {
	mu    sync.Mutex
	conns ConnectionChannel

	factory Factory

	connectionIdleTimeout time.Duration
	lastUsedAt            time.Time
	idleTimeout           time.Duration

	name       string
	initialCap uint32
	maxCap     uint32
	logger     logr.Logger
}

// Factory is a function to create new connections.
type Factory func() (net.Conn, error)

// ChannelPoolOption is a function for channel pool options.
type ChannelPoolOption func(*channelPool)

// ChannelPoolWithLogger sets the logger of the pool.
func ChannelPoolWithLogger(logger logr.Logger) ChannelPoolOption {
	return func(p *channelPool) {
		p.logger = logger
	}
}

// ChannelPoolWithInitialCap sets initial capability of the pool.
func ChannelPoolWithInitialCap(capacity uint32) ChannelPoolOption {
	return func(p *channelPool) {
		p.initialCap = capacity
	}
}

// ChannelPoolWithMaxCap sets maximum connection count of the pool.
func ChannelPoolWithMaxCap(capacity uint32) ChannelPoolOption {
	return func(p *channelPool) {
		p.maxCap = capacity
	}
}

// ChannelPoolWithConnectionIdleTimeout sets connection idle timeout of the pool.
func ChannelPoolWithConnectionIdleTimeout(timeout time.Duration) ChannelPoolOption {
	return func(p *channelPool) {
		p.connectionIdleTimeout = timeout
	}
}

// ChannelPoolWithIdleTimeout sets idle timeout of the pool.
func ChannelPoolWithIdleTimeout(timeout time.Duration) ChannelPoolOption {
	return func(p *channelPool) {
		p.idleTimeout = timeout
	}
}

// ChannelPoolWithName sets the name of the pool.
func ChannelPoolWithName(name string) ChannelPoolOption {
	return func(p *channelPool) {
		p.name = name
	}
}

// NewChannelPool returns a new pool based on buffered channels.
// A new connection will be created during a Get() via the Factory(),
// if there is no new connection available in the pool.
func NewChannelPool(factory Factory, opts ...ChannelPoolOption) (Pool, error) {
	c := &channelPool{
		factory:               factory,
		lastUsedAt:            time.Now(),
		initialCap:            DefaultInitialCap,
		maxCap:                DefaultMaxCap,
		connectionIdleTimeout: DefaultConnectionIdleTimeout,
		idleTimeout:           DefaultIdleTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.logger == (logr.Logger{}) {
		c.logger = logr.Discard()
	}

	c.logger = c.logger.WithName(c.name).V(3)

	if c.initialCap > c.maxCap {
		return nil, ErrInvalidCapacity
	}

	c.conns = make(ConnectionChannel, c.maxCap)

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < int(c.initialCap); i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, errors.WrapIf(err, "could not create new connection with factory")
		}
		c.conns <- c.wrapConn(conn)
	}

	go c.handleIdleTimeout()

	return c, nil
}

// Get returns a connection from the pool. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) Get() (net.Conn, error) {
	c.lastUsedAt = time.Now()

	conns, factory := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}

	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		conn.Acquire()

		return conn, nil
	default:
		conn, err := factory()
		if err != nil {
			return nil, errors.WrapIf(err, "could not create new connection with factory")
		}

		return c.wrapConn(conn), nil
	}
}

// Put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) Put(conn net.Conn) error {
	c.lastUsedAt = time.Now()

	if conn == nil {
		return ErrNilConnection
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.removeClosedConnections()

	if c.IsClosed() {
		// pool is closed, close passed connection
		return conn.Close()
	}

	wc := c.wrapConn(conn)

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- wc:
		c.logger.Info("put connection back to the pool")
		wc.Release()
		return nil
	default:
		// pool is full, close passed connection
		c.logger.Info("pool is full, close connection")
		return conn.Close()
	}
}

// Close closes the pool and all its connections. After Close() the pool is
// no longer usable.
func (c *channelPool) Close() error {
	if c.IsClosed() {
		return nil
	}

	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

	c.logger.Info("pool closed")

	if conns == nil {
		return nil
	}

	close(conns)
	for conn := range conns {
		err := conn.Discard()
		if err != nil {
			return err
		}
	}

	return nil
}

// IsClosed returns whether the pool is closed.
func (c *channelPool) IsClosed() bool {
	return c.conns == nil
}

// Len returns the current number of connections of the pool.
func (c *channelPool) Len() int {
	conns, _ := c.getConnsAndFactory()

	return len(conns)
}

// getConnsAndFactory returns the connection channel and the factory.
func (c *channelPool) getConnsAndFactory() (ConnectionChannel, Factory) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removeClosedConnections()
	conns := c.conns
	factory := c.factory

	return conns, factory
}

// removeClosedConnections removes the closed connections from the channel.
func (c *channelPool) removeClosedConnections() {
	max := len(c.conns)
	if max == 0 {
		return
	}

	c.logger.Info("remove closed connections")
	for i := 0; i < max; i++ {
		conn := <-c.conns
		if !conn.IsClosed() {
			c.conns <- conn
		}
	}
}

// wrapConn wraps a standard net.Conn to a PoolConnection.
func (c *channelPool) wrapConn(conn net.Conn) Connection {
	p := &poolConnection{
		Conn:        conn,
		pool:        c,
		idleTimeout: c.connectionIdleTimeout,
	}

	return p
}

// handleIdleTimeout handles idle timeout management
func (c *channelPool) handleIdleTimeout() {
	handle := func() {
		c.logger.Info("check idle timeout", "lastUsedAt", c.lastUsedAt)
		if !c.lastUsedAt.IsZero() && time.Since(c.lastUsedAt) > c.idleTimeout {
			c.logger.Info("idle timeout elapsed")
			c.Close()
		}
	}

	t := time.NewTicker(time.Second * 1)
	for {
		<-t.C
		if c.IsClosed() {
			c.logger.Info("stop idle timeout checking")
			t.Stop()
			return
		}
		handle()
	}
}
