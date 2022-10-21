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
	"math"
	"net"

	"github.com/cisco-open/nasp/pkg/ads/internal/endpoint"

	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

type weightedRoundRobinLoadBalancer struct {
	endpoints      []*envoy_config_endpoint_v3.Endpoint
	endpointsLoad  map[string]float64
	endpointsStats *endpoint.EndpointsStats
}

func NewWeightedRoundRobinLoadBalancer(
	endpoints []*envoy_config_endpoint_v3.Endpoint,
	endpointsLoad map[string]float64,
	endpointsStats *endpoint.EndpointsStats) *weightedRoundRobinLoadBalancer {
	return &weightedRoundRobinLoadBalancer{
		endpoints:      endpoints,
		endpointsLoad:  endpointsLoad,
		endpointsStats: endpointsStats,
	}
}

// NextEndpoint returns the next Endpoint using interleaved weighted round-robin algorithm
func (lb *weightedRoundRobinLoadBalancer) NextEndpoint() net.Addr {
	var totalRoundRobinCount uint32

	for endpointAddress := range lb.endpointsLoad {
		if stats, ok := lb.endpointsStats.Get(endpointAddress); ok {
			totalRoundRobinCount += stats.RoundRobinCount
		}
	}

	totalRoundRobinCount++ // we want to retrieve the next endpoint for the next upcoming connection

	// select endpoint with the least relative load
	var selectedEndpoint *envoy_config_endpoint_v3.Endpoint
	var selectedEndpointLoad float64
	var selectedEndpointAddress *net.TCPAddr
	minLoad := math.MaxFloat64

	for _, ep := range lb.endpoints {
		address := net.TCPAddr{
			IP:   net.ParseIP(ep.GetAddress().GetSocketAddress().GetAddress()),
			Port: int(ep.GetAddress().GetSocketAddress().GetPortValue()),
		}
		endpointLoad, ok := lb.endpointsLoad[address.String()]
		if !ok {
			continue
		}

		// the maximum number of times this endpoint can be selected by the round-robin
		// algorithm to satisfy locality's weight as well as endpoint's weights within the locality
		maxSelectCount := endpointLoad * float64(totalRoundRobinCount)

		relativeLoad := float64(0) // the current load relative to maxSelectCount
		stats, ok := lb.endpointsStats.Get(address.String())
		if ok {
			if float64(stats.RoundRobinCount) >= maxSelectCount {
				continue // skip this endpoint as was already selected by round-robin maxLoad times
			}
			relativeLoad = float64(stats.RoundRobinCount) / maxSelectCount
		}

		if relativeLoad < minLoad {
			minLoad = relativeLoad
			selectedEndpoint = ep
			selectedEndpointLoad = endpointLoad
			selectedEndpointAddress = &address
			continue
		}

		if minLoad == relativeLoad {
			if endpointLoad > selectedEndpointLoad {
				selectedEndpoint = ep
				selectedEndpointLoad = endpointLoad
				selectedEndpointAddress = &address
				continue
			}

			if endpointLoad == selectedEndpointLoad && selectedEndpoint == nil {
				selectedEndpoint = ep
				selectedEndpointAddress = &address
			}
		}
	}

	lb.endpointsStats.IncRoundRobinCounter(selectedEndpointAddress.String())

	return endpoint.GetAddress(selectedEndpoint)
}

func (lb *weightedRoundRobinLoadBalancer) SetEndpointsLoad(endpointsLoad map[string]float64) {
	lb.endpointsLoad = endpointsLoad
}
