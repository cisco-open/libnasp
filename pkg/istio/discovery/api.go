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

package discovery

import (
	"context"

	"net"
)

type DiscoveryClient interface {
	Connect(ctx context.Context) error
	ListenerDiscoveryClient
	ClientDiscoveryClient
}

type ClientDiscoveryClient interface {
	GetHTTPClientPropertiesByHost(ctx context.Context, address string, callbacks ...func(HTTPClientProperties)) (HTTPClientProperties, error)
	GetTCPClientPropertiesByHost(ctx context.Context, address string, callbacks ...func(ClientProperties)) (ClientProperties, error)
}

type ListenerDiscoveryClient interface {
	GetListenerProperties(ctx context.Context, address string, callbacks ...func(ListenerProperties)) (ListenerProperties, error)
}

type ListenerProperties interface {
	UseTLS() bool
	Permissive() bool
	IsClientCertificateRequired() bool
	Metadata() map[string]interface{}
}

type ClientProperties interface {
	Address() (net.Addr, error)
	UseTLS() bool
	Permissive() bool
	ServerName() string
	Metadata() map[string]interface{}
}

type HTTPClientProperties interface {
	ClientProperties
	ServerName() string
}
