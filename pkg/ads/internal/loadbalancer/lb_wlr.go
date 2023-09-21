//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

package loadbalancer

import (
	"math/rand"
	"net"
	"sort"
	"time"

	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	"github.com/cisco-open/libnasp/pkg/ads/internal/endpoint"
)

// weightedLeastRequestLoadBalancer load balances traffic to endpoints according to
// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#weighted-least-request
type weightedLeastRequestLoadBalancer struct {
	endpoints               []*envoy_config_endpoint_v3.Endpoint
	endpointsLoad           map[string]float64
	endpointsStats          *endpoint.EndpointsStats
	choiceCount             uint32
	endpointWeightsAreEqual bool
}

func NewWeightedLeastLoadBalancer(
	endpoints []*envoy_config_endpoint_v3.Endpoint,
	endpointsLoad map[string]float64,
	endpointsStats *endpoint.EndpointsStats,
	choiceCount uint32,
	endpointWeightsAreEqual bool) *weightedLeastRequestLoadBalancer {
	return &weightedLeastRequestLoadBalancer{
		endpoints:               endpoints,
		endpointsLoad:           endpointsLoad,
		endpointsStats:          endpointsStats,
		choiceCount:             choiceCount,
		endpointWeightsAreEqual: endpointWeightsAreEqual,
	}
}

// NextEndpoint returns the next Endpoint using the algorithm described by
// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#weighted-least-request
func (lb *weightedLeastRequestLoadBalancer) NextEndpoint() net.Addr {
	if lb.endpointWeightsAreEqual {
		return lb.p2cNextEndpoint()
	}

	// If two or more hosts in the cluster have different load balancing weights, the load balancer shifts into a mode
	// where it uses a weighted round-robin schedule in which weights are dynamically adjusted based on the host’s
	// request load at the time of selection.

	return NewWeightedRoundRobinLoadBalancer(lb.endpoints, lb.endpointsLoad, lb.endpointsStats).NextEndpoint()
}

// p2cNextEndpoint returns the next Endpoint using the Power of Two load balancing
// algorithm
func (lb *weightedLeastRequestLoadBalancer) p2cNextEndpoint() net.Addr {
	if len(lb.endpoints) == 0 {
		return nil
	}
	if len(lb.endpoints) == 1 {
		return endpoint.GetAddress(lb.endpoints[0])
	}
	endpoints := make([]*envoy_config_endpoint_v3.Endpoint, len(lb.endpoints))
	copy(endpoints, lb.endpoints)

	sort.SliceStable(endpoints, func(i, j int) bool {
		return endpoint.GetAddress(endpoints[i]).String() < endpoint.GetAddress(endpoints[j]).String()
	})

	// pick `lb.choiceCount` random elements from endpoints
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	for i := 0; i < int(lb.choiceCount); i++ {
		idx := rnd.Int31n(int32(len(endpoints) - i))

		// swap randomly selected element with the element at position (last - i)
		endpoints[idx], endpoints[len(endpoints)-i-1] = endpoints[len(endpoints)-i-1], endpoints[idx]
	}

	// the randomly picked `lb.choiceCount` elements are stored at the end of the array
	var selectedEndpoint *envoy_config_endpoint_v3.Endpoint
	var selectedEndpointActiveRequestsCount uint32

	for _, ep := range endpoints[len(endpoints)-int(lb.choiceCount):] {
		address := net.TCPAddr{
			IP:   net.ParseIP(ep.GetAddress().GetSocketAddress().GetAddress()),
			Port: int(ep.GetAddress().GetSocketAddress().GetPortValue()),
		}

		if selectedEndpoint == nil {
			selectedEndpoint = ep
			selectedEndpointActiveRequestsCount = lb.endpointsStats.ActiveRequestsCount(address.String())
			continue
		}

		if lb.endpointsStats.ActiveRequestsCount(address.String()) < selectedEndpointActiveRequestsCount {
			selectedEndpoint = ep
			selectedEndpointActiveRequestsCount = lb.endpointsStats.ActiveRequestsCount(address.String())
		}
	}

	return endpoint.GetAddress(selectedEndpoint)
}

func (lb *weightedLeastRequestLoadBalancer) SetEndpointsLoad(endpointsLoad map[string]float64) {
	lb.endpointsLoad = endpointsLoad
}
