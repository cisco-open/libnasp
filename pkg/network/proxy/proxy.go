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

	lclosed, rclosed bool

	mu sync.Mutex
}

func New(lconn, rconn io.ReadWriteCloser) *proxy {
	return &proxy{
		lconn: lconn,
		rconn: rconn,

		mu: sync.Mutex{},
	}
}

func (p *proxy) Start() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		p.proxy(&wg, p.lconn, p.rconn)
		p.closeConnections()
	}()

	go func() {
		p.proxy(&wg, p.rconn, p.lconn)
		p.closeConnections()
	}()

	wg.Wait()
}

func (p *proxy) closeConnections() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.lclosed {
		if err := p.lconn.Close(); err != nil {
			return err
		}
		p.lclosed = true
	}

	if !p.rclosed {
		if err := p.rconn.Close(); err != nil {
			return err
		}
		p.rclosed = true
	}

	return nil
}

func (p *proxy) proxy(wg *sync.WaitGroup, src, dst io.ReadWriteCloser) error {
	defer wg.Done()

	err := p.copy(src, dst)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
		return err
	}

	return nil
}

func (p *proxy) copy(src, dst io.ReadWriter) error {
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			return err
		}
		b := buff[:n]

		n, err = dst.Write(b)
		if err != nil {
			return err
		}
	}
}
