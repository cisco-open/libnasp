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
	"github.com/cisco-open/nasp/pkg/ads/internal/util"
)

// Stats stores various items related to a service endpoint
type Stats struct {
	// RoundRobinCount is the number of times the endpoint was selected by round-robin load balancing
	RoundRobinCount uint32
}

// EndpointsStats stores Stats for a collection of endpoints
type EndpointsStats struct {
	*util.KeyValueCollection[string, Stats]
}

func (eps *EndpointsStats) IncRoundRobinCounter(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.RoundRobinCount++
		eps.Items()[endpointAddress] = stats
	}
}

func (eps *EndpointsStats) ResetRoundRobinCounter(endpointAddress string) {
	eps.Lock()
	defer eps.Unlock()

	if stats, ok := eps.Items()[endpointAddress]; ok {
		stats.RoundRobinCount = 0
		eps.Items()[endpointAddress] = stats
	}
}

func NewEndpointsStats() *EndpointsStats {
	return &EndpointsStats{
		KeyValueCollection: util.NewKeyValueCollection[string, Stats](),
	}
}
