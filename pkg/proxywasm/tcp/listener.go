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

package tcp

import (
	"net"

	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
)

type wrappedListener struct {
	net.Listener

	streamHandler api.StreamHandler
}

func WrapListener(l net.Listener, streamHandler api.StreamHandler) net.Listener {
	return &wrappedListener{
		Listener: l,

		streamHandler: streamHandler,
	}
}

func (l *wrappedListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return l.handleNewConnection(c)
}

func (l *wrappedListener) handleNewConnection(conn net.Conn) (net.Conn, error) {
	stream, err := l.streamHandler.NewStream(api.ListenerDirectionInbound)
	if err != nil {
		return nil, err
	}

	stream.Set("response.flags", 0)

	conn = NewWrappedConn(conn, stream)

	if err := stream.HandleTCPNewConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}
