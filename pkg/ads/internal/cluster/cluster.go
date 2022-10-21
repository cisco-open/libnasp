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

package cluster

import (
	"math"

	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/cisco-open/nasp/pkg/ads/internal/util"
)

type FilterOption interface {
	Filter(*envoy_config_cluster_v3.Cluster) bool
}

// Filter returns elements from the given clusters which match the given filters
func Filter(clusters []*envoy_config_cluster_v3.Cluster, opts ...FilterOption) []*envoy_config_cluster_v3.Cluster {
	if opts == nil {
		return clusters
	}

	var filteredClusters []*envoy_config_cluster_v3.Cluster
	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}

		add := true
		for _, opt := range opts {
			if opt != nil {
				if !opt.Filter(cluster) {
					add = false
					break
				}
			}
		}

		if add {
			filteredClusters = append(filteredClusters, cluster)
		}
	}
	return filteredClusters
}

// SelectCluster selects a cluster from a list of clusters specified in weightedClusters when traffic is routed
// to multiple clusters according to their weight
func SelectCluster(weightedClusters map[string]uint32, clustersStats *ClustersStats) string {
	if len(weightedClusters) == 0 {
		return ""
	}

	var totalSelectionCount, totalWeight uint32
	for clusterName, clusterWeight := range weightedClusters {
		totalWeight += clusterWeight
		if stats, ok := clustersStats.Get(clusterName); ok {
			totalSelectionCount += stats.WeightedClusterSelectionCount
		}
	}

	totalSelectionCount++ // we want to retrieve the next cluster for the next upcoming connection

	// select cluster with the least relative load
	var selectedCluster string
	var selectedClusterWeight uint32
	minLoad := math.MaxFloat64

	for clusterName, clusterWeight := range weightedClusters {
		// the maximum number of times this cluster can be selected by the algorithm to satisfy cluster's weights
		maxLoad := float64(clusterWeight) / float64(totalWeight) * float64(totalSelectionCount)

		load := float64(0) // the current load relative to maxLoad
		if stats, ok := clustersStats.Get(clusterName); ok {
			if float64(stats.WeightedClusterSelectionCount) >= maxLoad {
				continue // skip this cluster as was already selected maxLoad times
			}
			load = float64(stats.WeightedClusterSelectionCount) / maxLoad
		}

		if load < minLoad {
			minLoad = load
			selectedCluster = clusterName
			selectedClusterWeight = clusterWeight
			continue
		}

		if minLoad == load {
			if clusterWeight > selectedClusterWeight {
				selectedCluster = clusterName
				selectedClusterWeight = clusterWeight
				continue
			}

			if clusterWeight == selectedClusterWeight && selectedCluster == "" {
				selectedCluster = clusterName
				selectedClusterWeight = clusterWeight
			}
		}
	}

	clustersStats.IncWeightedClusterSelectionCounter(selectedCluster)

	return selectedCluster
}

// GetLoadBalancingPolicy returns the load balancing policy of the given cluster
func GetLoadBalancingPolicy(cluster *envoy_config_cluster_v3.Cluster) envoy_config_cluster_v3.Cluster_LbPolicy {
	// TODO: check cluster.GetLoadBalancingPolicy() as if it is set than it supersedes cluster.GetLbPolicy()
	return cluster.GetLbPolicy()
}

func GetMetadata(cluster *envoy_config_cluster_v3.Cluster) (map[string]interface{}, error) {
	return util.GetUnifiedMetadata(cluster.GetMetadata())
}

// GetTlsServerName returns the SNI to be used when connecting to endpoints in the selectedCluster using TLS
func GetTlsServerName(cluster *envoy_config_cluster_v3.Cluster) string {
	if cluster.GetTransportSocketMatches() != nil {
		for _, tsm := range cluster.GetTransportSocketMatches() {
			if tsm.GetTransportSocket().GetName() != wellknown.TransportSocketTLS {
				continue
			}

			config, err := tsm.GetTransportSocket().GetTypedConfig().UnmarshalNew()
			if err == nil {
				if upstreamTlsCtx, ok := config.(*envoy_extensions_transport_sockets_tls_v3.UpstreamTlsContext); ok {
					return upstreamTlsCtx.GetSni()
				}
			}
		}

		return ""
	}

	if cluster.GetTransportSocket().GetName() != wellknown.TransportSocketTLS {
		return ""
	}
	config, err := cluster.GetTransportSocket().GetTypedConfig().UnmarshalNew()
	if err == nil {
		if upstreamTlsCtx, ok := config.(*envoy_extensions_transport_sockets_tls_v3.UpstreamTlsContext); ok {
			return upstreamTlsCtx.GetSni()
		}
	}

	return ""
}

// UsesTls returns true if the endpoints in the given selectedCluster accepts TLS traffic
func UsesTls(cluster *envoy_config_cluster_v3.Cluster) bool {
	if cluster.GetTransportSocketMatches() != nil {
		for _, tsm := range cluster.GetTransportSocketMatches() {
			if tsm.GetTransportSocket().GetName() == wellknown.TransportSocketTLS {
				return true
			}
		}
		return false
	}

	// If no transport socket matches configuration is specified check transport socket configuration
	return cluster.GetTransportSocket().GetName() == wellknown.TransportSocketTLS
}

// IsPermissive returns true if the endpoints in the given selectedCluster can accept both plaintext and mutual TLS traffic
func IsPermissive(cluster *envoy_config_cluster_v3.Cluster) bool {
	if cluster.GetTransportSocketMatches() != nil {
		tlsTransportSocket := false
		rawBufferTransportSocket := false
		for _, tsm := range cluster.GetTransportSocketMatches() {
			switch tsm.GetTransportSocket().GetName() {
			case wellknown.TransportSocketTLS:
				tlsTransportSocket = true
			case wellknown.TransportSocketRawBuffer:
				rawBufferTransportSocket = true
			}
		}

		// If there are transport socket matches for both tls and raw_buffer than the
		// endpoints in the given selectedCluster can accept both plaintext and mutual TLS traffic
		return tlsTransportSocket && rawBufferTransportSocket
	}

	return false
}
