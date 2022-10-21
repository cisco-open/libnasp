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

package stream

import (
	"net"
	"net/url"
	"os"
)

type SocketFS struct {
	openedSockets map[string]*Socket

	FDProvider FDProvider
}

type Socket struct {
	net.Conn
	fd int
}

func (s *Socket) Fd() int {
	return s.fd
}

func NewSocketFS(fdProvider FDProvider) *SocketFS {
	return &SocketFS{
		openedSockets: make(map[string]*Socket, 0),
		FDProvider:    fdProvider,
	}
}

func (fs *SocketFS) AddSocket(id string, conn net.Conn) {
	fs.openedSockets[id] = &Socket{
		Conn: conn,
	}
}

func (fs *SocketFS) Open(path string, flag int, perm uint32) (Stream, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if u.Path == "" {
		return nil, os.ErrNotExist
	}

	path = u.Path[1:]

	if socket, ok := fs.openedSockets[path]; ok {
		return socket, nil
	}

	return nil, os.ErrNotExist
}

func (h *SocketFS) Close() error {
	return nil
}
