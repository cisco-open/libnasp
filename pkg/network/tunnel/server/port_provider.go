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

package server

import (
	"sync"

	"emperror.dev/errors"
)

var ErrInvalidPortRange = errors.New("invalid port range")

type portProvider struct {
	portMin int
	portMax int
	ports   map[int]bool

	mu sync.Mutex
}

type PortProvider interface {
	GetFreePort() int
	ReleasePort(int)
}

func NewPortProvider(portMin, portMax int) (PortProvider, error) {
	if portMin > portMax {
		return nil, errors.WithStackIf(ErrInvalidPortRange)
	}

	return &portProvider{
		ports:   make(map[int]bool),
		portMin: portMin,
		portMax: portMax,
		mu:      sync.Mutex{},
	}, nil
}

func (p *portProvider) GetFreePort() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := p.portMin; i < p.portMax; i++ {
		if _, ok := p.ports[i]; !ok {
			p.ports[i] = true
			return i
		}
	}

	return 0
}

func (p *portProvider) ReleasePort(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.ports, port)
}
