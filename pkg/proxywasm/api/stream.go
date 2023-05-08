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

package api

import (
	"io"
	"net"
	"net/url"

	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/network"
)

type ListenerDirection uint64

const (
	ListenerDirectionUnspecified ListenerDirection = iota
	ListenerDirectionInbound
	ListenerDirectionOutbound
)

type StreamOption func(instance interface{})

type StreamHandler interface {
	NewStream(direction ListenerDirection, options ...StreamOption) (Stream, error)
	Logger() logr.Logger
}

type Stream interface {
	PropertyHolder

	Logger() logr.Logger
	Close() error
	Direction() ListenerDirection

	HandleHTTPRequest(req HTTPRequest) error
	HandleHTTPResponse(resp HTTPResponse) error

	HandleTCPNewConnection(conn net.Conn) error
	HandleTCPCloseConnection(conn net.Conn) error
	HandleDownstreamData(conn net.Conn, n int) error
	HandleUpstreamData(conn net.Conn, n int) error
}

type HTTPResponse interface {
	Header() HeaderMap
	Trailer() HeaderMap
	Body() io.ReadCloser
	SetBody(io.ReadCloser)

	ContentLength() int64
	StatusCode() int

	ConnectionState() network.ConnectionState
}

type HTTPRequest interface {
	URL() *url.URL
	Header() HeaderMap
	Trailer() HeaderMap
	Body() io.ReadCloser
	SetBody(io.ReadCloser)

	HTTPProtocol() string
	Host() string
	Method() string

	ConnectionState() network.ConnectionState
}
