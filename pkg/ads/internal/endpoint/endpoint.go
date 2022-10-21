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
	"math"
	"net"
	"sort"
	"strings"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
)

// GetClusterLoadAssignmentReferences returns the ClusterLoadAssignment names that references the provided clusters
func GetClusterLoadAssignmentReferences(clusters []*envoy_config_cluster_v3.Cluster) []string {
	endpointResourceNames := make(map[string]struct{})

	for _, cluster := range clusters {
		endpointResourceName := GetClusterLoadAssignmentReference(cluster)
		if endpointResourceName == "" {
			continue
		}
		endpointResourceNames[endpointResourceName] = struct{}{}
	}

	resourceNames := make([]string, 0, len(endpointResourceNames))
	for k := range endpointResourceNames {
		resourceNames = append(resourceNames, k)
	}

	sort.Strings(resourceNames)

	return resourceNames
}

func GetClusterLoadAssignmentReference(cluster *envoy_config_cluster_v3.Cluster) string {
	if cluster == nil {
		return ""
	}

	clusterDiscoveryType, ok := cluster.ClusterDiscoveryType.(*envoy_config_cluster_v3.Cluster_Type)
	if !ok {
		return ""
	}

	if clusterDiscoveryType.Type == envoy_config_cluster_v3.Cluster_EDS {
		if cluster.EdsClusterConfig != nil && cluster.EdsClusterConfig.ServiceName != "" {
			return cluster.GetEdsClusterConfig().GetServiceName()
		} else {
			return cluster.GetName()
		}
	}

	return ""
}

func GetAddress(ep *envoy_config_endpoint_v3.Endpoint) net.Addr {
	address := ep.GetAddress().GetSocketAddress()
	if address == nil {
		return nil
	}

	return &net.TCPAddr{
		IP:   net.ParseIP(address.GetAddress()),
		Port: int(address.GetPortValue()),
	}
}

type LbEndpointsFilterOption interface {
	Filter(ep *envoy_config_endpoint_v3.LbEndpoint) bool
}

type endpointWithSocketAddressFilter struct {
	address *net.TCPAddr
}

func (f *endpointWithSocketAddressFilter) Filter(ep *envoy_config_endpoint_v3.LbEndpoint) bool {
	socketAddress := ep.GetEndpoint().GetAddress().GetSocketAddress()
	if socketAddress == nil {
		return false
	}

	address := net.TCPAddr{
		IP:   net.ParseIP(socketAddress.GetAddress()),
		Port: int(socketAddress.GetPortValue()),
	}
	if f.address == nil || f.address.String() == address.String() {
		return true
	}
	return false
}

func HasSocketAddress() *endpointWithSocketAddressFilter {
	return &endpointWithSocketAddressFilter{}
}

func WithSocketAddress(address *net.TCPAddr) *endpointWithSocketAddressFilter {
	return &endpointWithSocketAddressFilter{address: address}
}

type healthyEndpointFilter struct{}

func (f *healthyEndpointFilter) Filter(ep *envoy_config_endpoint_v3.LbEndpoint) bool {
	return IsHealthy(ep)
}

func WithHealthyStatus() *healthyEndpointFilter { return &healthyEndpointFilter{} }

func Filter(localityLbEndpoints []*envoy_config_endpoint_v3.LocalityLbEndpoints, opts ...LbEndpointsFilterOption) []*envoy_config_endpoint_v3.Endpoint {
	var eps []*envoy_config_endpoint_v3.Endpoint
	for _, ep := range localityLbEndpoints {
		for _, lbEndpoint := range ep.GetLbEndpoints() {
			add := true
			for _, opt := range opts {
				if opt != nil {
					if !opt.Filter(lbEndpoint) {
						add = false
						break
					}
				}
			}

			if add {
				eps = append(eps, lbEndpoint.GetEndpoint())
			}
		}
	}

	return eps
}

// IsHealthy returns true if the given endpoint is healthy
func IsHealthy(ep *envoy_config_endpoint_v3.LbEndpoint) bool {
	//nolint:exhaustive
	switch ep.GetHealthStatus() {
	case envoy_config_core_v3.HealthStatus_UNKNOWN,
		envoy_config_core_v3.HealthStatus_HEALTHY,
		envoy_config_core_v3.HealthStatus_DEGRADED:
		return true
	default:
		// all other health statuses are considered by Envoy as unhealthy
		return false
	}
}

// healthScore computes the health score of the given endpoints collection as
// healthy_endpoints count / total endpoints count
func healthScore(endpoints []*envoy_config_endpoint_v3.LbEndpoint) float64 {
	healthyEndpoints := 0
	endpointsCount := 0

	for _, ep := range endpoints {
		if ep == nil {
			continue
		}

		endpointsCount++
		if IsHealthy(ep) {
			healthyEndpoints++
		}
	}

	if endpointsCount == 0 {
		return 0.0
	}

	return float64(healthyEndpoints) / float64(endpointsCount)
}

// getPriorityLevelsLoadDistribution computes the load for locality priority levels
// according to https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/priority#priority-levels
// It returns a map where the key is the priority level and the value the priority level's computed load
func getPriorityLevelsLoadDistribution(localityLbEndpoints []*envoy_config_endpoint_v3.LocalityLbEndpoints, overProvisioningFactor float64) map[uint32]float64 {
	if len(localityLbEndpoints) == 0 {
		return nil
	}

	endpointsByPriorityLevel := make(map[uint32][]*envoy_config_endpoint_v3.LbEndpoint)
	for _, localityLbEndpoint := range localityLbEndpoints {
		endpoints := endpointsByPriorityLevel[localityLbEndpoint.GetPriority()]
		endpoints = append(endpoints, localityLbEndpoint.GetLbEndpoints()...)

		endpointsByPriorityLevel[localityLbEndpoint.GetPriority()] = endpoints
	}

	priorityLevels := make([]uint32, 0, len(endpointsByPriorityLevel))

	priorityLevelsHealthScore := make(map[uint32]float64, len(endpointsByPriorityLevel))
	normalizedTotalPriorityLevelHealthScore := 0.0

	for priorityLevel, endpoints := range endpointsByPriorityLevel {
		priorityLevels = append(priorityLevels, priorityLevel)

		priorityLevelsHealthScore[priorityLevel] = overProvisioningFactor * healthScore(endpoints)
		if priorityLevelsHealthScore[priorityLevel] > 1.0 {
			priorityLevelsHealthScore[priorityLevel] = 1.0
		}

		normalizedTotalPriorityLevelHealthScore += priorityLevelsHealthScore[priorityLevel]
	}

	if normalizedTotalPriorityLevelHealthScore > 1.0 {
		normalizedTotalPriorityLevelHealthScore = 1.0
	}

	sort.Slice(priorityLevels, func(i, j int) bool {
		return priorityLevels[i] < priorityLevels[j]
	})

	priorityLevelsLoad := make(map[uint32]float64, len(priorityLevelsHealthScore))
	prevPriorityLevelsLoadSum := 0.0

	for _, priorityLevel := range priorityLevels {
		priorityLevelLoad := priorityLevelsHealthScore[priorityLevel] / normalizedTotalPriorityLevelHealthScore
		if priorityLevelLoad > (1.0 - prevPriorityLevelsLoadSum) {
			priorityLevelLoad = 1.0 - prevPriorityLevelsLoadSum
		}
		priorityLevelLoad = math.Round(priorityLevelLoad*10000) / 10000 // truncate precision to 4 decimals to avoid floating point number precision errors

		priorityLevelsLoad[priorityLevel] = priorityLevelLoad
		prevPriorityLevelsLoadSum += priorityLevelLoad
	}

	return priorityLevelsLoad
}

// getLocalityLoadDistribution returns effective locality loads which are computed as described
// at https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/locality_weight#locality-weighted-load-balancing
func getLocalityLoadDistribution(localityLbEndpoints []*envoy_config_endpoint_v3.LocalityLbEndpoints, overProvisioningFactor float64) map[string]float64 {
	priorityLevelsLoad := getPriorityLevelsLoadDistribution(localityLbEndpoints, overProvisioningFactor)
	endpointsByLocality := make(map[string][]*envoy_config_endpoint_v3.LbEndpoint)

	for _, localityLbEndpoint := range localityLbEndpoints {
		locality := localityToString(localityLbEndpoint.GetLocality())
		endpoints := endpointsByLocality[locality]
		endpoints = append(endpoints, localityLbEndpoint.GetLbEndpoints()...)

		endpointsByLocality[locality] = endpoints
	}

	effectiveWeights := make(map[string]float64)
	effectiveWeightsSum := 0.0
	for _, localityLbEndpoint := range localityLbEndpoints {
		locality := localityToString(localityLbEndpoint.GetLocality())

		endpoints := endpointsByLocality[locality]
		localityAvailability := overProvisioningFactor * healthScore(endpoints)
		if localityAvailability > 1.0 {
			localityAvailability = 1.0
		}

		weight := getLocalityLoadBalancingWeight(localityLbEndpoint)
		effectiveWeights[locality] = float64(weight) * localityAvailability * priorityLevelsLoad[localityLbEndpoint.GetPriority()]
		effectiveWeightsSum += effectiveWeights[locality]
	}

	loads := make(map[string]float64)
	for _, localityLbEndpoint := range localityLbEndpoints {
		locality := localityToString(localityLbEndpoint.GetLocality())
		load := effectiveWeights[locality] / effectiveWeightsSum
		load = math.Round(load*10000) / 10000 // truncate precision to 4 decimals to avoid floating point number precision errors

		loads[locality] = load
	}

	return loads
}

// GetLoadDistribution computes the load distribution for healthy endpoints taking into account
// priority level and locality load distribution
func GetLoadDistribution(localityLbEndpoints []*envoy_config_endpoint_v3.LocalityLbEndpoints, overProvisioningFactor float64) map[string]float64 {
	totalWeightPerLocality := make(map[string]uint32)
	for _, localityLbEndpoint := range localityLbEndpoints {
		locality := localityToString(localityLbEndpoint.GetLocality())
		for _, lbEndpoint := range localityLbEndpoint.GetLbEndpoints() {
			if !IsHealthy(lbEndpoint) {
				continue
			}

			weight := GetLoadBalancingWeight(lbEndpoint)
			totalWeightPerLocality[locality] += weight
		}
	}

	loads := make(map[string]float64)
	localityLoad := getLocalityLoadDistribution(localityLbEndpoints, overProvisioningFactor)

	for _, localityLbEndpoint := range localityLbEndpoints {
		locality := localityToString(localityLbEndpoint.GetLocality())
		for _, lbEndpoint := range localityLbEndpoint.GetLbEndpoints() {
			socketAddress := lbEndpoint.GetEndpoint().GetAddress().GetSocketAddress()
			if socketAddress == nil {
				continue
			}

			address := net.TCPAddr{
				IP:   net.ParseIP(socketAddress.GetAddress()),
				Port: int(socketAddress.GetPortValue()),
			}

			weight := GetLoadBalancingWeight(lbEndpoint)
			if !IsHealthy(lbEndpoint) {
				weight = 0
			}
			load := float64(weight) / float64(totalWeightPerLocality[locality]) * localityLoad[locality]
			load = math.RoundToEven(load*100) / 100 // truncate precision to 2 decimals to avoid floating point number precision errors

			loads[address.String()] = load
		}
	}

	return loads
}

func GetLoadBalancingWeight(ep *envoy_config_endpoint_v3.LbEndpoint) uint32 {
	weight := uint32(1)

	if ep.GetLoadBalancingWeight() != nil {
		weight = ep.GetLoadBalancingWeight().GetValue()
	}

	return weight
}

func getLocalityLoadBalancingWeight(locality *envoy_config_endpoint_v3.LocalityLbEndpoints) uint32 {
	weight := uint32(1)

	if locality.GetLoadBalancingWeight() != nil {
		weight = locality.GetLoadBalancingWeight().GetValue()
	}

	return weight
}

func localityToString(locality *envoy_config_core_v3.Locality) string {
	if locality == nil {
		return ""
	}

	return strings.Join([]string{locality.GetRegion(), locality.GetZone(), locality.GetSubZone()}, "/")
}
