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

//nolint:goerr113
package tcp_test

import (
	"bytes"
	"errors"
	"net"
	"sync"
	"testing"

	"golang.org/x/net/nettest"

	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	"github.com/cisco-open/nasp/pkg/proxywasm/tcp"
)

func BenchmarkNetConn(b *testing.B) {
	b.Run("plain", func(b *testing.B) {
		benchmark(b, pipemaker())
	})
	b.Run("wrapped", func(b *testing.B) {
		benchmark(b, wrappedpipemaker())
	})
}

func benchmark(b *testing.B, piper nettest.MakePipe) {
	b.Helper()

	c1, c2, stop, _ := piper()
	defer stop()

	content := bytes.Repeat([]byte("."), 1024)
	readbuff := make([]byte, len(content))

	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errs := make(chan error, 1)

		wg.Add(1)
		go func() {
			defer wg.Done()

			_, _ = c1.Write(content)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			l, err := c2.Read(readbuff)
			if err != nil {
				errs <- err
				return
			}

			if !bytes.Equal(content, readbuff[:l]) {
				err = errors.New("transmitted content differs")
			}

			errs <- err
		}()

		wg.Wait()

		err := <-errs
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestNetConn(t *testing.T) {
	t.Parallel()

	nettest.TestConn(t, pipemaker())
}

func TestWrappedNetConn(t *testing.T) {
	t.Parallel()

	nettest.TestConn(t, wrappedpipemaker())
}

func pipemaker() nettest.MakePipe {
	return func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2 = net.Pipe()
		stop = func() {
			c1.Close()
			c2.Close()
		}

		return
	}
}

func wrappedpipemaker() nettest.MakePipe {
	return func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2 = net.Pipe()
		stop = func() {
			c1.Close()
			c2.Close()
		}

		c1 = tcp.NewWrappedConn(c1, wasmstream)
		c2 = tcp.NewWrappedConn(c2, wasmstream)

		return
	}
}

var wasmstream api.Stream = NewTestWASMStream()

func NewTestWASMStream() api.Stream {
	return &teststream{}
}

type teststream struct {
}

func (s *teststream) Get(key string) (interface{}, bool) {
	return nil, false
}

func (s *teststream) Set(key string, value interface{}) {

}

func (s *teststream) Logger() logr.Logger {
	return logr.Discard()
}

func (s *teststream) Close() error {
	return nil
}

func (s *teststream) Direction() api.ListenerDirection {
	return api.ListenerDirectionInbound
}

func (s *teststream) HandleHTTPRequest(req api.HTTPRequest) error {
	return nil
}

func (s *teststream) HandleHTTPResponse(resp api.HTTPResponse) error {
	return nil
}

func (s *teststream) HandleTCPNewConnection(conn net.Conn) error {
	return nil
}

func (s *teststream) HandleTCPCloseConnection(conn net.Conn) error {
	return nil
}

func (s *teststream) HandleDownstreamData(conn net.Conn, n int) error {
	return nil
}

func (s *teststream) HandleUpstreamData(conn net.Conn, n int) error {
	return nil
}
