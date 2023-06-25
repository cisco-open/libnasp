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

package proxy

import (
	"errors"
	"io"
	"net"
	"sync"
)

type proxy struct {
	lconn, rconn io.ReadWriteCloser

	mu     sync.Mutex
	closel sync.Once
	closer sync.Once

	bufferSize int
}

type ProxyOption func(*proxy)

func WithBufferSize(size int) ProxyOption {
	return func(p *proxy) {
		p.bufferSize = size
	}
}

func New(lconn, rconn io.ReadWriteCloser, options ...ProxyOption) *proxy {
	p := &proxy{
		lconn: lconn,
		rconn: rconn,

		mu: sync.Mutex{},

		bufferSize: 4096,
	}

	for _, option := range options {
		option(p)
	}

	return p
}

func (p *proxy) Start() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = p.proxy(p.lconn, p.rconn)

		p.closer.Do(func() {
			_ = p.close(p.rconn)
		})
	}()

	go func() {
		defer wg.Done()
		_ = p.proxy(p.rconn, p.lconn)

		p.closel.Do(func() {
			_ = p.close(p.lconn)
		})
	}()

	wg.Wait()
}

func (p *proxy) close(conn io.ReadWriteCloser) error {
	if d, ok := conn.(interface {
		CloseWrite() error
	}); ok {
		return d.CloseWrite()
	}

	return conn.Close()
}

func (p *proxy) proxy(src, dst io.ReadWriteCloser) error {
	_, err := io.CopyBuffer(dst, src, make([]byte, p.bufferSize))
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
		return err
	}

	return nil
}
