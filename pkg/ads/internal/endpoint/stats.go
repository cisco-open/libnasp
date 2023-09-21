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

package endpoint

import (
	"github.com/cisco-open/libnasp/pkg/ads/internal/util"
)

// Stats stores various items related to a service endpoint
type Stats struct {
	// roundRobinCount is the number of times the endpoint was selected by weighted round-robin load balancing
	roundRobinCount uint32

	// activeRequestsCount is the number of active requests the endpoint has.
	// Note: A request completes when the upstream response reaches its end-of-stream, i.e. when trailers
	// or the response header/body with end-stream set are received. (see https://www.envoyproxy.io/docs/envoy/latest/intro/life_of_a_request#response-path-and-http-lifecycle)
	activeRequestsCount uint32
}

// ActiveRequestsCount returns the number of active requests the endpoint has.
func (s *Stats) ActiveRequestsCount() uint32 {
	if s == nil {
		return 0
	}

	return s.activeRequestsCount
}

// RoundRobinCount returns the number of times the endpoint was selected by weighted round-robin load balancing
func (s *Stats) RoundRobinCount() uint32 {
	if s == nil {
		return 0
	}

	return s.roundRobinCount
}

// EndpointsStats stores Stats for a collection of endpoints
type EndpointsStats struct {
	*util.KeyValueCollection[string, *Stats]
}

func (eps *EndpointsStats) IncRoundRobinCount(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.roundRobinCount++
	}
}

// RoundRobinCount returns the number of times the endpoint with the given address was selected by weighted round-robin load balancing
func (eps *EndpointsStats) RoundRobinCount(endpointAddress string) uint32 {
	stats, _ := eps.Get(endpointAddress)

	return stats.RoundRobinCount()
}

// ActiveRequestsCount returns the number of active requests the endpoint with given address has.
func (eps *EndpointsStats) ActiveRequestsCount(endpointAddress string) uint32 {
	stats, _ := eps.Get(endpointAddress)

	return stats.ActiveRequestsCount()
}

func (eps *EndpointsStats) SetActiveRequestsCount(endpointAddress string, activeRequestsCount uint32) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.activeRequestsCount = activeRequestsCount
	}
}

func (eps *EndpointsStats) IncActiveRequestsCount(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.activeRequestsCount++
	}
}

func (eps *EndpointsStats) DecActiveRequestsCount(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.activeRequestsCount--
	}
}

func (eps *EndpointsStats) ResetRoundRobinCount(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.roundRobinCount = 0
	}
}

func (eps *EndpointsStats) ResetActiveRequestsCount(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.activeRequestsCount = 0
	}
}

func NewEndpointsStats() *EndpointsStats {
	return &EndpointsStats{
		KeyValueCollection: util.NewKeyValueCollection[string, *Stats](),
	}
}
