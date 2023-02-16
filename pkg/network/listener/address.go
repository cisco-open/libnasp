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

package listener

import (
	"net"
	"strings"
)

type addr struct {
	net  string
	addr string
}

func (a addr) Network() string {
	return a.net
}

func (a addr) String() string {
	return a.addr
}

func (l *multipleListeners) Addr() net.Addr {
	networks := make([]string, 0)
	addresses := make([]string, 0)

	l.listeners.Range(func(k, v any) bool {
		if l, ok := v.(net.Listener); ok {
			networks = append(networks, l.Addr().Network())
			addresses = append(addresses, l.Addr().String())
		}

		return true
	})

	return addr{
		net:  strings.Join(networks, ","),
		addr: strings.Join(addresses, ","),
	}
}
