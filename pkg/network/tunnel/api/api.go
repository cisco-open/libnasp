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
	ErrConnectionClosed        = errors.New("connection closed")
	ErrCtrlInvalidResponse     = errors.New("invalid response on control channel")
	ErrInvalidMessageType      = errors.New("invalid message type")
	ErrInvalidPort             = errors.New("invalid port")
	ErrInvalidStreamType       = errors.New("invalid stream type")
	ErrInvalidStreamID         = errors.New("invalid stream id")
	ErrSessionTimeout          = errors.New("getting session timed out")
	ErrListenerStopped         = errors.New("listener stopped")
	ErrInvalidConnection       = errors.New("invalid connection")
	ErrAcceptTimeout           = errors.New("accepting connection timed out")
	ErrInvalidParam            = errors.New("invalid parameter")
	ErrCtrlStreamAlreadyExists = errors.New("control stream already exists")
	ErrPortAlreadyExists       = errors.New("port already exists")
)

type Client interface {
	Connect(ctx context.Context) error
	GetTCPListener(id string, options ManagedPortOptions) (net.Listener, error)
}

type ManagedPortOptions interface {
	SetName(string) ManagedPortOptions
	GetName() string
	SetRequestedPort(int) ManagedPortOptions
	GetRequestedPort() int
	SetTargetPort(int) ManagedPortOptions
	GetTargetPort() int
}

type Server interface {
	Start(ctx context.Context) error
}

type Session interface {
	Close() error
	Handle() error
	GetControlStream() io.ReadWriter
	OpenTCPStream(id string) (net.Conn, error)
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type MessageHandlerFunc func(msg Message, params ...any) error

type ControlStream interface {
	Stream
	GetMetadata() map[string]string
	GetUser() User
}

type MessageHandler interface {
	Stream
	AddMessageHandler(messageType MessageType, handler MessageHandlerFunc)
}

type Stream interface {
	io.ReadWriteCloser
	Handle() error
}

type PortProvider interface {
	GetFreePort() int
	GetPort(port int) bool
	ReleasePort(int)
}

type Authenticator interface {
	Authenticate(ctx context.Context, bearerToken string) (bool, User, error)
}

type User interface {
	UID() string
	Name() string
	Groups() []string
	Metadata() map[string]string
}

type EventName string

const (
	PortListenEventName  EventName = "port.listen"
	PortReleaseEventName EventName = "port.release"
)

type PortData struct {
	Name       string
	Address    string
	Port       int
	TargetPort int

	User     User
	Metadata map[string]string
}

type PortListenEvent struct {
	SessionID string
	PortData
}

type PortReleaseEvent struct {
	SessionID string
	PortData
}
