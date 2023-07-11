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

package client

import (
	"net"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
)

var _ net.Listener = &managedPort{}

type managedPort struct {
	id            string
	options       api.ManagedPortOptions
	remoteAddress string

	connChan chan net.Conn

	initialized bool
}

type ManagedPortOptions struct {
	name          string
	requestedPort int
	targetPort    int
}

func (o *ManagedPortOptions) SetTargetPort(port int) api.ManagedPortOptions {
	o.targetPort = port

	return o
}

func (o *ManagedPortOptions) GetTargetPort() int {
	return o.targetPort
}

func (o *ManagedPortOptions) SetRequestedPort(port int) api.ManagedPortOptions {
	o.requestedPort = port

	return o
}

func (o *ManagedPortOptions) GetRequestedPort() int {
	return o.requestedPort
}

func (o *ManagedPortOptions) SetName(name string) api.ManagedPortOptions {
	o.name = name

	return o
}

func (o *ManagedPortOptions) GetName() string {
	return o.name
}

func NewManagedPort(id string, options api.ManagedPortOptions) *managedPort {
	return &managedPort{
		id:      id,
		options: options,

		connChan: make(chan net.Conn),
	}
}

func (p *managedPort) ID() string {
	return p.id
}

func (p *managedPort) GetConnChannel() chan<- net.Conn {
	return p.connChan
}

func (p *managedPort) Accept() (net.Conn, error) {
	conn, open := <-p.connChan

	if !open {
		return nil, api.ErrListenerStopped
	}

	if conn == nil {
		return nil, api.ErrInvalidConnection
	}

	return conn, nil
}

func (p *managedPort) Close() error {
	close(p.connChan)

	return nil
}

func (p *managedPort) Addr() net.Addr {
	if p.remoteAddress != "" {
		if addr, err := net.ResolveTCPAddr("tcp", p.remoteAddress); err == nil {
			return addr
		}
	}

	return &net.TCPAddr{
		Port: p.options.GetRequestedPort(),
	}
}
