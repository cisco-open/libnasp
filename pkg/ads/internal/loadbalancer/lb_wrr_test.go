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

package loadbalancer_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/cisco-open/libnasp/pkg/ads/internal/endpoint"
	"github.com/cisco-open/libnasp/pkg/ads/internal/loadbalancer"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

func newEnvoyEndpoint(address *net.TCPAddr) *envoy_config_endpoint_v3.Endpoint {
	return &envoy_config_endpoint_v3.Endpoint{
		Address: &envoy_config_core_v3.Address{
			Address: &envoy_config_core_v3.Address_SocketAddress{
				SocketAddress: &envoy_config_core_v3.SocketAddress{
					Address:       address.IP.String(),
					PortSpecifier: &envoy_config_core_v3.SocketAddress_PortValue{PortValue: uint32(address.Port)},
				},
			},
		},
	}
}

func newLbEndpoint(endpoint *envoy_config_endpoint_v3.Endpoint, lbWeight uint32, healthStatus envoy_config_core_v3.HealthStatus) *envoy_config_endpoint_v3.LbEndpoint {
	return &envoy_config_endpoint_v3.LbEndpoint{
		HostIdentifier:      &envoy_config_endpoint_v3.LbEndpoint_Endpoint{Endpoint: endpoint},
		LoadBalancingWeight: &wrappers.UInt32Value{Value: lbWeight},
		HealthStatus:        healthStatus,
	}
}

func Test_weightedRoundRobinLoadBalancer_nextEndpoint(t *testing.T) {
	t.Parallel()

	stats := endpoint.NewEndpointsStats()
	endpoints := []*envoy_config_endpoint_v3.Endpoint{
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.1"), Port: 0}),
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.2"), Port: 0}),
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.3"), Port: 0}),
		newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.4"), Port: 0}),
	}
	localityLbEndpoints := []*envoy_config_endpoint_v3.LocalityLbEndpoints{
		{
			Locality: &envoy_config_core_v3.Locality{
				Region: "test_region",
				Zone:   "test_zone",
			},
			Priority:            0,
			LoadBalancingWeight: &wrappers.UInt32Value{Value: 1},
			LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
				newLbEndpoint(endpoints[0], 1, envoy_config_core_v3.HealthStatus_UNKNOWN), // 10% load
				newLbEndpoint(endpoints[1], 2, envoy_config_core_v3.HealthStatus_UNKNOWN), // 20% load
				newLbEndpoint(endpoints[2], 7, envoy_config_core_v3.HealthStatus_UNKNOWN), // 70% load
				newLbEndpoint(endpoints[3], 0, envoy_config_core_v3.HealthStatus_UNKNOWN), // 0% load
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
	endpointsLoad := endpoint.GetLoadDistribution(localityLbEndpoints, stats, 1.4, -1.0)

	lb := loadbalancer.NewWeightedRoundRobinLoadBalancer(endpoints, endpointsLoad, stats)

	expectedEndpoints := []net.Addr{
		endpoint.GetAddress(endpoints[2]),
		endpoint.GetAddress(endpoints[1]),
		endpoint.GetAddress(endpoints[0]),
		endpoint.GetAddress(endpoints[2]),
		endpoint.GetAddress(endpoints[2]),
		endpoint.GetAddress(endpoints[2]),
		endpoint.GetAddress(endpoints[1]),
		endpoint.GetAddress(endpoints[2]),
		endpoint.GetAddress(endpoints[2]),
		endpoint.GetAddress(endpoints[2]),
	}
	expectedCount := []uint32{1, 2, 7, 0}
	counter := make([]uint32, len(endpoints))

	// first 10 round
	for _, expected := range expectedEndpoints {
		actual := lb.NextEndpoint()

		if !reflect.DeepEqual(actual, expected) {
			t.Error("Expected: ", expected, ", got: ", actual)
			return
		}

		for i, ep := range endpoints {
			if reflect.DeepEqual(endpoint.GetAddress(ep), actual) {
				counter[i]++
			}
		}
	}
	for i := range counter {
		if counter[i] != expectedCount[i] {
			t.Error("Expected count: ", expectedCount[i], ", got: ", counter[i], " for endpoint ", endpoints[i])
		}
	}

	var totWeight uint32
	for _, localityLbEndpoint := range localityLbEndpoints {
		for _, lbEndpoint := range localityLbEndpoint.GetLbEndpoints() {
			totWeight += endpoint.GetLoadBalancingWeight(lbEndpoint)
		}
	}

	// next 1000 round
	for i := 0; i < 1000; i++ {
		next := lb.NextEndpoint()
		for j, ep := range endpoints {
			if reflect.DeepEqual(endpoint.GetAddress(ep), next) {
				counter[j]++
			}
		}
	}

	expectedCount = []uint32{101, 202, 707, 0}
	for i := range counter {
		if counter[i] != expectedCount[i] {
			t.Error("Expected count: ", expectedCount[i], ", got: ", counter[i], " for endpoint ", endpoints[i])
		}
	}

	// change weights and run another 1000 rounds
	localityLbEndpoints[0].LbEndpoints[2].LoadBalancingWeight.Value = 4 // 7 -> 4 (70% -> 40%)
	localityLbEndpoints[0].LbEndpoints[3].LoadBalancingWeight.Value = 3 // 0 -> 3 (0% -> 30%)

	lb.SetEndpointsLoad(endpoint.GetLoadDistribution(localityLbEndpoints, stats, 1.4, -1.0))

	for i := 0; i < 1000; i++ {
		next := lb.NextEndpoint()
		for j, ep := range endpoints {
			if reflect.DeepEqual(endpoint.GetAddress(ep), next) {
				counter[j]++
			}
		}
	}
	expectedCount = []uint32{201, 402, 804, 603}
	for i := range counter {
		if counter[i] != expectedCount[i] {
			t.Error("Expected count: ", expectedCount[i], ", got: ", counter[i], " for endpoint ", endpoints[i])
		}
	}
}

//nolint:maintidx,funlen
func Test_endpointsLoadDistribution(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	tests := []struct {
		name                              string
		localityLbEndpoints               []*envoy_config_endpoint_v3.LocalityLbEndpoints
		expectedEndpointsLoadDistribution map[string]float64
	}{
		{
			name: "all endpoints in priority level 0 are healthy",
			localityLbEndpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_1",
						Zone:   "test_zone_1",
					},
					Priority:            0,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 6}, // test_region_1/test_zone_1 locality weight 60% of all traffic in the priority level 0
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.1"), Port: 8080}), 6, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.2"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.3"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.4"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_1",
						Zone:   "test_zone_2",
					},
					Priority:            0,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 4}, // test_region_1/test_zone_2 locality weight 40% of all traffic in the priority level 0
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.5"), Port: 8080}), 6, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.6"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.7"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.8"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_2",
						Zone:   "test_zone_1",
					},
					Priority:            1,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 6}, // test_region_2/test_zone_1 locality weight 60% of all traffic in the priority level 1
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.9"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.10"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.11"), Port: 8080}), 3, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.12"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_2",
						Zone:   "test_zone_2",
					},
					Priority:            1,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 4}, // test_region_2/test_zone_2 locality weight 40% of all traffic in the priority level 1
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.13"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.14"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.15"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.16"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
			},
			expectedEndpointsLoadDistribution: map[string]float64{
				// P0
				// test_region_1/test_zone_1
				"10.10.0.1:8080": 0.36, // P0: 100%, locality: 60%, endpoint weight: 60% ==> 36%
				"10.10.0.2:8080": 0.12, // P0: 100%, locality: 60%, endpoint weight: 20% ==> 12%
				"10.10.0.3:8080": 0.06, // P0: 100%, locality: 60%, endpoint weight: 10% ==> 6%
				"10.10.0.4:8080": 0.06, // P0: 100%, locality: 60%, endpoint weight: 10% ==> 6%
				// test_region_1/test_zone_2
				"10.10.0.5:8080": 0.24, // P0: 100%, locality: 40%, endpoint weight: 60% ==> 24%
				"10.10.0.6:8080": 0.08, // P0: 100%, locality: 40%, endpoint weight: 20% ==> 8%
				"10.10.0.7:8080": 0.04, // P0: 100%, locality: 40%, endpoint weight: 10% ==> 4%
				"10.10.0.8:8080": 0.04, // P0: 100%, locality: 40%, endpoint weight: 10% ==> 4%
				// P1
				// all endpoints in P1 should not receive traffic as endpoints in P0 are healthy
				// test_region_2/test_zone_1
				"10.10.0.9:8080":  0.0, // P1: 0%, locality: 60%, endpoint weight: 25% ==> 0%
				"10.10.0.10:8080": 0.0, // P1: 0%, locality: 60%, endpoint weight: 25% ==> 0%
				"10.10.0.11:8080": 0.0, // P1: 0%, locality: 60%, endpoint weight: 37.5% ==> 0%
				"10.10.0.12:8080": 0.0, // P1: 0%, locality: 60%, endpoint weight: 12.5% ==> 0%
				// test_region_1/test_zone_2
				"10.10.0.13:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
				"10.10.0.14:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
				"10.10.0.15:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
				"10.10.0.16:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
			},
		},
		{
			name: "more than 71% of the endpoints in priority level 0 are healthy",
			localityLbEndpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_1",
						Zone:   "test_zone_1",
					},
					Priority:            0,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 6}, // test_region_1/test_zone_1 locality weight 60% of all traffic in the priority level 0
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.1"), Port: 8080}), 6, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.2"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNHEALTHY),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.3"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.4"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_1",
						Zone:   "test_zone_2",
					},
					Priority:            0,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 4}, // test_region_1/test_zone_2 locality weight 40% of all traffic in the priority level 0
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.5"), Port: 8080}), 6, envoy_config_core_v3.HealthStatus_UNHEALTHY),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.6"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.7"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.8"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_2",
						Zone:   "test_zone_1",
					},
					Priority:            1,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 6}, // test_region_2/test_zone_1 locality weight 60% of all traffic in the priority level 1
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.9"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.10"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.11"), Port: 8080}), 3, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.12"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_2",
						Zone:   "test_zone_2",
					},
					Priority:            1,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 4}, // test_region_2/test_zone_2 locality weight 40% of all traffic in the priority level 1
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.13"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.14"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.15"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.16"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
			},
			expectedEndpointsLoadDistribution: map[string]float64{
				// P0
				// test_region_1/test_zone_1
				"10.10.0.1:8080": 0.45, // P0: 100%, locality: 60%, endpoint weight: 75% ==> 45%
				"10.10.0.2:8080": 0.0,  // P0: 100%, locality: 60%, endpoint weight: 0% ==> 0% (unhealthy endpoint)
				"10.10.0.3:8080": 0.08, // P0: 100%, locality: 60%, endpoint weight: 12.5% ==> 8%
				"10.10.0.4:8080": 0.08, // P0: 100%, locality: 60%, endpoint weight: 12.5% ==> 8%
				// test_region_1/test_zone_2
				"10.10.0.5:8080": 0.0, // P0: 100%, locality: 40%, endpoint weight: 0% ==> 0% (unhealthy endpoint)
				"10.10.0.6:8080": 0.2, // P0: 100%, locality: 40%, endpoint weight: 50% ==> 20%
				"10.10.0.7:8080": 0.1, // P0: 100%, locality: 40%, endpoint weight: 25% ==> 10%
				"10.10.0.8:8080": 0.1, // P0: 100%, locality: 40%, endpoint weight: 25% ==> 10%
				// P1
				// all endpoints in P1 should not receive traffic as endpoints in P0 are healthy
				// test_region_2/test_zone_1
				"10.10.0.9:8080":  0.0, // P1: 0%, locality: 60%, endpoint weight: 25% ==> 0%
				"10.10.0.10:8080": 0.0, // P1: 0%, locality: 60%, endpoint weight: 25% ==> 0%
				"10.10.0.11:8080": 0.0, // P1: 0%, locality: 60%, endpoint weight: 37.5% ==> 0%
				"10.10.0.12:8080": 0.0, // P1: 0%, locality: 60%, endpoint weight: 12.5% ==> 0%
				// test_region_1/test_zone_2
				"10.10.0.13:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
				"10.10.0.14:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
				"10.10.0.15:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
				"10.10.0.16:8080": 0.0, // P1: 0%, locality: 40%, endpoint weight: 25% ==> 0%
			},
		},
		{
			name: "less than 71% of the endpoints in priority level 0 are healthy",
			localityLbEndpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_1",
						Zone:   "test_zone_1",
					},
					Priority:            0,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 6}, // test_region_1/test_zone_1 locality weight 60% of all traffic in the priority level 0
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.1"), Port: 8080}), 6, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.2"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNHEALTHY),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.3"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNHEALTHY),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.4"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNHEALTHY),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_1",
						Zone:   "test_zone_2",
					},
					Priority:            0,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 4}, // test_region_1/test_zone_2 locality weight 40% of all traffic in the priority level 0
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.5"), Port: 8080}), 6, envoy_config_core_v3.HealthStatus_UNHEALTHY),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.6"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.7"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.8"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_2",
						Zone:   "test_zone_1",
					},
					Priority:            1,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 6}, // test_region_2/test_zone_1 locality weight 60% of all traffic in the priority level 1
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.9"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.10"), Port: 8080}), 2, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.11"), Port: 8080}), 3, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.12"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
				{
					Locality: &envoy_config_core_v3.Locality{
						Region: "test_region_2",
						Zone:   "test_zone_2",
					},
					Priority:            1,
					LoadBalancingWeight: &wrappers.UInt32Value{Value: 4}, // test_region_2/test_zone_2 locality weight 40% of all traffic in the priority level 1
					LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.13"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.14"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.15"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
						newLbEndpoint(newEnvoyEndpoint(&net.TCPAddr{IP: net.ParseIP("10.10.0.16"), Port: 8080}), 1, envoy_config_core_v3.HealthStatus_UNKNOWN),
					},
				},
			},
			expectedEndpointsLoadDistribution: map[string]float64{
				// P0 is not healthy thus will receive only 70% of the traffic the rest goes to P1
				// test_region_1/test_zone_1
				"10.10.0.1:8080": 0.20, // P0: 70%, locality: 28.88%, endpoint weight: 100% ==> 20%
				"10.10.0.2:8080": 0.0,  // P0: 70%, locality: 28.88%, endpoint weight: 0% ==> 0% (unhealthy endpoint)
				"10.10.0.3:8080": 0.0,  // P0: 70%, locality: 28.88%, endpoint weight: 0% ==> 0% (unhealthy endpoint)
				"10.10.0.4:8080": 0.0,  // P0: 70%, locality: 28.88%, endpoint weight: 0% ==> 0% (unhealthy endpoint)
				// test_region_1/test_zone_2
				"10.10.0.5:8080": 0.0,  // P0: 70%, locality: 55.01%, endpoint weight: 0% ==> 0% (unhealthy endpoint)
				"10.10.0.6:8080": 0.19, // P0: 70%, locality: 55.01%, endpoint weight: 50% ==> 19%
				"10.10.0.7:8080": 0.1,  // P0: 70%, locality: 55.01%, endpoint weight: 25% ==> 1%
				"10.10.0.8:8080": 0.1,  // P0: 70%, locality: 55.01%, endpoint weight: 25% ==> 1%
				// P1
				// less the 71% of endpoint in P0 are healthy thus P1 should take over some traffic
				// test_region_2/test_zone_1
				"10.10.0.9:8080":  0.06, // P1: 30%, locality: 82.53%, endpoint weight: 25% ==> 6%
				"10.10.0.10:8080": 0.06, // P1: 30%, locality: 82.53%, endpoint weight: 25% ==> 6%
				"10.10.0.11:8080": 0.09, // P1: 30%, locality: 82.53%, endpoint weight: 37.5% ==> 9%
				"10.10.0.12:8080": 0.03, // P1: 30%, locality: 82.53%, endpoint weight: 12.5% ==> 3%
				// test_region_1/test_zone_2
				"10.10.0.13:8080": 0.04, // P1: 30%, locality: 55.03%, endpoint weight: 25% ==> 4%
				"10.10.0.14:8080": 0.04, // P1: 30%, locality: 55.03%, endpoint weight: 25% ==> 4%
				"10.10.0.15:8080": 0.04, // P1: 30%, locality: 55.03%, endpoint weight: 25% ==> 4%
				"10.10.0.16:8080": 0.04, // P1: 30%, locality: 55.03%, endpoint weight: 25% ==> 4%
			},
		},
	}

	for _, test := range tests {
		tc := test
		stats := endpoint.NewEndpointsStats()

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actualEndpointsLoadDistribution := endpoint.GetLoadDistribution(tc.localityLbEndpoints, nil, 1.4, -1.0)
			if assert.Equal(tc.expectedEndpointsLoadDistribution, actualEndpointsLoadDistribution) == false {
				return
			}

			// if active request bias is 0.0 than load distribution should not change
			actualEndpointsLoadDistribution = endpoint.GetLoadDistribution(tc.localityLbEndpoints, stats, 1.4, 0.0)
			if assert.Equal(tc.expectedEndpointsLoadDistribution, actualEndpointsLoadDistribution) == false {
				return
			}
		})
	}
}
