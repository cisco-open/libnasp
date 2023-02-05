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

//nolint:forcetypeassert,gosec,gocritic
package pool

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	emperror "emperror.dev/errors"
	"k8s.io/klog/v2"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

var (
	address    = "127.0.0.1:7777"
	factory    = func() (net.Conn, error) { return net.Dial("tcp", address) }
	errFactory = errors.New("factory error")
	badFactory = func() (net.Conn, error) { return nil, errFactory }
)

func init() {
	// used for factory function
	go simpleTCPServer()
	time.Sleep(time.Millisecond * 300) // wait until tcp server has been settled

	rand.Seed(time.Now().UTC().UnixNano())
}

func TestChannelPool_New(t *testing.T) {
	t.Parallel()

	_, err := NewChannelPool(factory)
	require.Nil(t, err)
}

func TestChannelPool_InvalidCapacityError(t *testing.T) {
	t.Parallel()

	_, err := NewChannelPool(factory, ChannelPoolWithInitialCap(5), ChannelPoolWithMaxCap(3))
	require.ErrorIs(t, ErrInvalidCapacity, err)
}

func TestChannelPool_FactoryErrorAtInit(t *testing.T) {
	t.Parallel()

	_, err := NewChannelPool(badFactory, ChannelPoolWithInitialCap(5), ChannelPoolWithMaxCap(10))
	require.ErrorIs(t, errFactory, emperror.Cause(err))
}

func TestChannelPool_NameOption(t *testing.T) {
	t.Parallel()

	name := "name"
	p, _ := NewChannelPool(badFactory, ChannelPoolWithName(name))
	require.Equal(t, name, p.(*channelPool).name)
}

func TestChannelPool_IdleTimeoutOption(t *testing.T) {
	t.Parallel()

	timeout := time.Second * 99
	p, _ := NewChannelPool(badFactory, ChannelPoolWithIdleTimeout(timeout))
	require.Equal(t, timeout, p.(*channelPool).idleTimeout)
}

func TestChannelPool_ConnectionIdleTimeoutOption(t *testing.T) {
	t.Parallel()

	timeout := time.Second * 99
	p, _ := NewChannelPool(badFactory, ChannelPoolWithConnectionIdleTimeout(timeout))
	require.Equal(t, timeout, p.(*channelPool).connectionIdleTimeout)
}

func TestChannelPool_InitialCapOption(t *testing.T) {
	t.Parallel()

	initialCap := uint32(13)
	p, _ := NewChannelPool(factory, ChannelPoolWithInitialCap(initialCap))
	require.Equal(t, initialCap, p.(*channelPool).initialCap)
}

func TestChannelPool_MaxCapOption(t *testing.T) {
	t.Parallel()

	initialCap := uint32(13)
	p, _ := NewChannelPool(factory, ChannelPoolWithMaxCap(initialCap))
	require.Equal(t, initialCap, p.(*channelPool).maxCap)
}

func TestChannelPool_LoggerOption(t *testing.T) {
	t.Parallel()

	logger := logr.Discard()
	p, _ := NewChannelPool(factory)
	require.Equal(t, logger.V(3), p.(*channelPool).logger)

	logger = klog.Background()
	p, _ = NewChannelPool(factory, ChannelPoolWithLogger(logger))
	require.Equal(t, logger.V(3), p.(*channelPool).logger)
}

func TestChannelPool_BadFactoryGet(t *testing.T) {
	t.Parallel()

	p, err := NewChannelPool(badFactory)
	require.Nil(t, err)

	c, err := p.Get()
	require.ErrorIs(t, errFactory, emperror.Cause(err))
	require.Nil(t, c)
}

func TestChannelPool_Get(t *testing.T) {
	t.Parallel()

	initialCap := 5

	p, err := NewChannelPool(factory, ChannelPoolWithInitialCap(uint32(initialCap)))
	require.Nil(t, err)
	defer p.Close()

	c, err := p.Get()
	require.Nil(t, err)
	require.IsType(t, &poolConnection{}, c)

	// check current capacity
	require.Equal(t, initialCap-1, p.Len())

	// get the remaining connections
	conns := make([]net.Conn, initialCap-1)
	for i := 0; i < 4; i++ {
		conn, _ := p.Get()
		conns[i] = conn
	}

	// close previously got connection
	c.Close()

	// check current capacity
	require.Equal(t, 1, p.Len())

	// another get creates a new connection with factory
	c, err = p.Get()
	require.Nil(t, err)
	require.IsType(t, &poolConnection{}, c)
	c.Close()

	// close stored connections
	for _, conn := range conns {
		conn.Close()
	}

	// check current capacity
	require.Equal(t, initialCap, p.Len())
}

func TestChannelPool_Put(t *testing.T) {
	t.Parallel()

	initialCap := 5
	maxCap := 30

	p, err := NewChannelPool(factory, ChannelPoolWithInitialCap(uint32(initialCap)), ChannelPoolWithMaxCap(uint32(maxCap)))
	require.Nil(t, err)
	defer p.Close()

	// get/create from the pool
	conns := make([]net.Conn, maxCap)
	for i := 0; i < maxCap; i++ {
		conn, _ := p.Get()
		conns[i] = conn
	}

	// now put them all back
	for _, conn := range conns {
		conn.Close()
	}

	require.Equal(t, maxCap, p.Len())

	conn, err := p.Get()
	require.Nil(t, err)

	// close pool
	p.Close()

	// connection should be closed since the pool is closed already
	conn.Close()
	require.Equal(t, 0, p.Len())
}

func TestChannelPool_Full(t *testing.T) {
	t.Parallel()

	initialCap := 3
	maxCap := 5
	connCount := 10

	p, err := NewChannelPool(factory, ChannelPoolWithInitialCap(uint32(initialCap)), ChannelPoolWithMaxCap(uint32(maxCap)))
	require.Nil(t, err)
	defer p.Close()

	// get the remaining connections
	conns := make([]net.Conn, connCount)
	for i := 0; i < connCount; i++ {
		conn, _ := p.Get()
		conns[i] = conn
	}

	// close stored connections
	for _, conn := range conns {
		conn.Close()
	}

	require.Equal(t, maxCap, p.Len())
}

func TestChannelPool_Len(t *testing.T) {
	t.Parallel()

	initialCap := 5
	p, err := NewChannelPool(factory, ChannelPoolWithInitialCap(uint32(initialCap)))
	require.Nil(t, err)
	defer p.Close()

	require.Equal(t, initialCap, p.Len())
}

func TestChannelPool_Close(t *testing.T) {
	t.Parallel()

	initialCap := 5
	p, err := NewChannelPool(factory, ChannelPoolWithInitialCap(uint32(initialCap)))
	require.Nil(t, err)
	p.Close()

	require.Equal(t, true, p.IsClosed())

	_, err = p.Get()
	require.ErrorIs(t, ErrClosed, emperror.Cause(err))
}

func TestChannelPool_Concurrency(t *testing.T) {
	t.Parallel()

	p, err := NewChannelPool(factory)
	require.Nil(t, err)

	var wg sync.WaitGroup

	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				conn, _ := p.Get()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				conn.Close()
				wg.Done()
			}()
		}
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			conn, _ := p.Get()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			conn.Close()
			wg.Done()
		}()
	}

	wg.Wait()

	p.Close()

	require.Equal(t, 0, p.Len())
	require.Equal(t, true, p.IsClosed())
}

func TestChannelPool_IdleTimeout(t *testing.T) {
	t.Parallel()

	timeout := time.Second * 1

	p, err := NewChannelPool(factory, ChannelPoolWithIdleTimeout(timeout))
	require.Nil(t, err)

	_, err = p.Get()
	require.Nil(t, err)

	// wait a second longer than idle timeout
	time.Sleep(timeout + time.Second)

	// pool len should be 0
	require.Equal(t, 0, p.Len())
	// pool should be closed
	require.Equal(t, true, p.IsClosed())
}

func TestChannelPool_ConnectionIdleTimeout(t *testing.T) {
	t.Parallel()

	timeout := time.Second * 1

	p, err := NewChannelPool(factory, ChannelPoolWithConnectionIdleTimeout(timeout))
	require.Nil(t, err)

	// create new connection
	c, err := p.Get()
	require.Nil(t, err)
	// close connection to put it back to the pool
	c.Close()

	// wait a second longer than idle timeout
	time.Sleep(timeout + time.Second)

	// pool len should be 0 as the idle connection should be closed
	require.Equal(t, 0, p.Len())
	// pool should NOT be closed
	require.Equal(t, false, p.IsClosed())

	b := make([]byte, 1024)
	_, err = c.Read(b)
	require.ErrorIs(t, err, net.ErrClosed)
}

func simpleTCPServer() {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			buffer := make([]byte, 256)
			_, _ = conn.Read(buffer)
		}()

		_, _ = conn.Write([]byte("example"))
	}
}
