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

	"emperror.dev/errors"
)

type connection struct {
	net.Conn

	localAddress  net.Addr
	remoteAddress net.Addr
}

func newconn(conn net.Conn, network string, laddr, raddr string) (*connection, error) {
	c := &connection{
		Conn: conn,
	}

	var err error

	switch network {
	case "tcp":
		if c.localAddress, err = net.ResolveTCPAddr(network, laddr); err != nil {
			return nil, err
		}
		if c.remoteAddress, err = net.ResolveTCPAddr(network, raddr); err != nil {
			return nil, err
		}

		return c, nil
	default:
		return nil, errors.NewWithDetails("unsupported network type", "type", network)
	}
}

func (c *connection) Read(b []byte) (n int, err error) {
	// fmt.Printf("reading from [%s]\n", c.remoteAddress)
	buff := make([]byte, cap(b))

	n, err = c.Conn.Read(buff)

	// fmt.Printf("read\n|%s|\n", string(buff))

	n = copy(b, buff[:n])

	return n, err
}

func (c *connection) Write(b []byte) (n int, err error) {
	// fmt.Printf("write\n|%s|\n", string(b))

	return c.Conn.Write(b)
}

func (c *connection) NetConn() net.Conn {
	return c.Conn
}

func (c *connection) LocalAddr() net.Addr {
	return c.localAddress
}

func (c *connection) RemoteAddr() net.Addr {
	return c.remoteAddress
}
