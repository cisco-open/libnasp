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

package loadbalancer

import (
	"math/rand"
	"net"
	"time"

	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	"github.com/cisco-open/libnasp/pkg/ads/internal/endpoint"
)

type randomLoadBalancer struct {
	endpoints []*envoy_config_endpoint_v3.Endpoint
}

func NewRandomLoadBalancer(endpoints []*envoy_config_endpoint_v3.Endpoint) *randomLoadBalancer {
	return &randomLoadBalancer{endpoints: endpoints}
}

// NextEndpoint returns randomly and Endpoint from endpoints
func (lb *randomLoadBalancer) NextEndpoint() net.Addr {
	if len(lb.endpoints) == 0 {
		return nil
	}

	if len(lb.endpoints) == 1 {
		return endpoint.GetAddress(lb.endpoints[0])
	}

	idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(lb.endpoints)) //nolint:gosec

	return endpoint.GetAddress(lb.endpoints[idx])
}
