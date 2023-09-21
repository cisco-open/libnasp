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

package loadbalancer_test

import (
	"net"
	"testing"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/cisco-open/libnasp/pkg/ads/internal/endpoint"
	"github.com/cisco-open/libnasp/pkg/ads/internal/loadbalancer"
)

func Test_leastRequest_P2C_LoadBalancer_nextEndpoint(t *testing.T) {
	t.Parallel()

	stats := endpoint.NewEndpointsStats()
	endpoints := []*envoy_config_endpoint_v3.Endpoint{
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.1"), Port: 0}),
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.2"), Port: 0}),
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.3"), Port: 0}),
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.4"), Port: 0}),
	}

	// all endpoint weights are equal
	localityLbEndpoints := []*envoy_config_endpoint_v3.LocalityLbEndpoints{
		{
			Locality: &envoy_config_core_v3.Locality{
				Region: "test_region",
				Zone:   "test_zone",
			},
			Priority:            0,
			LoadBalancingWeight: &wrappers.UInt32Value{Value: 1},
			LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
				newLbEndpoint(endpoints[0], 4, envoy_config_core_v3.HealthStatus_UNKNOWN),
				newLbEndpoint(endpoints[1], 4, envoy_config_core_v3.HealthStatus_UNKNOWN),
				newLbEndpoint(endpoints[2], 4, envoy_config_core_v3.HealthStatus_UNKNOWN),
				newLbEndpoint(endpoints[3], 4, envoy_config_core_v3.HealthStatus_UNKNOWN),
			},
		},
	}

	for _, ep := range endpoints {
		address := net.TCPAddr{
			IP:   net.ParseIP(ep.GetAddress().GetSocketAddress().GetAddress()),
			Port: int(ep.GetAddress().GetSocketAddress().GetPortValue()),
		}
		stats.Set(address.String(), &endpoint.Stats{})
	}
	stats.SetActiveRequestsCount("10.10.0.1:0", 1)
	stats.SetActiveRequestsCount("10.10.0.2:0", 2)
	stats.SetActiveRequestsCount("10.10.0.3:0", 7)
	stats.SetActiveRequestsCount("10.10.0.4:0", 0)

	endpointsLoad := endpoint.GetLoadDistribution(localityLbEndpoints, nil, 1.4, -1.0)

	lb := loadbalancer.NewWeightedLeastLoadBalancer(endpoints, endpointsLoad, stats, 2, true)

	counter := make([]int, len(endpoints))
	// run 5 iterations
	for i := 0; i < 5; i++ {
		addr := lb.NextEndpoint()
		if addr == nil {
			continue
		}

		stats.IncActiveRequestsCount(addr.String())

		for j, ep := range endpoints {
			if endpoint.GetAddress(ep).String() == addr.String() {
				counter[j]++
			}
		}
	}

	// "10.10.0.1:0" expected to have [0, 5] count
	if !(counter[0] >= 0 && counter[0] <= 5) {
		t.Error("Expected count to be in range [0, 5], got: ", counter[0], " for endpoint ", endpoints[0])
	}

	// "10.10.0.2:0" expected to have [0, 5] count
	if !(counter[1] >= 0 && counter[1] <= 5) {
		t.Error("Expected count to be in range [0, 5], got: ", counter[1], " for endpoint ", endpoints[1])
	}

	// "10.10.0.3:0" expected to have 0 count as it already has 7 active requests which is more than the max active requests
	// count potentially the other endpoints can reach through  5 iterations
	if counter[2] > 0 {
		t.Error("Expected count 0, got: ", counter[2], " for endpoint ", endpoints[2])
	}

	// "10.10.0.3:0" expected to have [0, 5] count
	if !(counter[3] >= 0 && counter[3] <= 5) {
		t.Error("Expected count to be in range [0, 5], got: ", counter[3], " for endpoint ", endpoints[3])
	}
}
