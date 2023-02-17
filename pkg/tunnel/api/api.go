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

package api

import (
	"context"
	"errors"
	"io"
	"net"
)

var (
	ErrConnectionClosed    = errors.New("connection closed")
	ErrCtrlInvalidResponse = errors.New("invalid response on control channel")
	ErrInvalidMessageType  = errors.New("invalid message type")
	ErrInvalidPort         = errors.New("invalid port")
	ErrInvalidStreamType   = errors.New("invalid stream type")
	ErrInvalidStreamID     = errors.New("invalid stream id")
	ErrSessionTimeout      = errors.New("timeout getting session")
)

type Client interface {
	Connect(ctx context.Context) error
	AddTCPPort(port int) (net.Listener, error)
}

type Server interface {
	Start(ctx context.Context) error
}

type Session interface {
	Close() error
	Handle() error
	GetControlStream() io.ReadWriter
	OpenTCPStream(port int, id string) (net.Conn, error)
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type MessageHandler func(msg []byte) error

type ControlStream interface {
	Stream
	AddMessageHandler(messageType MessageType, handler MessageHandler)
}

type Stream interface {
	io.ReadWriteCloser
	Handle() error
}
