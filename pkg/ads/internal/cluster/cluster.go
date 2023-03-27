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

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

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
		totalSelectionCount += clustersStats.WeightedClusterSelectionCount(clusterName)
	}

	totalSelectionCount++ // we want to retrieve the next cluster for the next upcoming connection

	// select cluster with the least relative load
	var selectedCluster string
	var selectedClusterWeight uint32
	minLoad := math.MaxFloat64

	for clusterName, clusterWeight := range weightedClusters {
		// the maximum number of times this cluster can be selected by the algorithm to satisfy cluster's weights
		maxLoad := float64(clusterWeight) / float64(totalWeight) * float64(totalSelectionCount)

		if float64(clustersStats.WeightedClusterSelectionCount(clusterName)) >= maxLoad {
			continue // skip this cluster as was already selected maxLoad times
		}
		load := float64(clustersStats.WeightedClusterSelectionCount(clusterName)) / maxLoad // the current load relative to maxLoad

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

	clustersStats.IncWeightedClusterSelectionCount(selectedCluster)

	return selectedCluster
}

// GetLoadBalancingPolicy returns the load balancing policy of the given cluster
func GetLoadBalancingPolicy(cluster *envoy_config_cluster_v3.Cluster) envoy_config_cluster_v3.Cluster_LbPolicy {
	// TODO: check cluster.GetLoadBalancingPolicy() as if it is set than it supersedes cluster.GetLbPolicy()
	return cluster.GetLbPolicy()
}

// GetFilterMetadata returns the cluster metadata stored under the 'metadata.typed_filter_metadata'
// and 'metadata.filter_metadata' of the given envoy cluster.
// If a key is present on both the one from 'metadata.typed_filter_metadata' will be taken into account.
func GetFilterMetadata(cluster *envoy_config_cluster_v3.Cluster) (map[string]interface{}, error) {
	return util.GetUnifiedFilterMetadata(cluster.GetMetadata())
}

// GetTlsServerName returns the SNI to be used when connecting to endpoints in the selectedCluster using TLS
func GetTlsServerName(transportSockets []*envoy_config_core_v3.TransportSocket) string {
	// the first match from the matching transport socket config is used when a connection is made to the endpoint
	if len(transportSockets) == 0 || transportSockets[0].GetName() != wellknown.TransportSocketTLS {
		return ""
	}

	config, err := transportSockets[0].GetTypedConfig().UnmarshalNew()
	if err == nil {
		if upstreamTlsCtx, ok := config.(*envoy_extensions_transport_sockets_tls_v3.UpstreamTlsContext); ok {
			return upstreamTlsCtx.GetSni()
		}
	}

	return ""
}

// UsesTls returns true if the endpoint matched by the given endpoint metadate match fields accepts TLS traffic
func UsesTls(transportSockets []*envoy_config_core_v3.TransportSocket) bool {
	// the first match from the matching transport socket config is used when a connection is made to the endpoint
	if len(transportSockets) == 0 {
		return false
	}

	return transportSockets[0].GetName() == wellknown.TransportSocketTLS
}

// IsPermissive returns true if the endpoints in the given cluster can accept both plaintext and mutual TLS traffic
func IsPermissive(transportSockets []*envoy_config_core_v3.TransportSocket) bool {
	tlsTransportSocket := false
	rawBufferTransportSocket := false

	for _, transportSocket := range transportSockets {
		switch transportSocket.GetName() {
		case wellknown.TransportSocketTLS:
			tlsTransportSocket = true
		case wellknown.TransportSocketRawBuffer:
			rawBufferTransportSocket = true
		}
	}

	// If there are transport socket matches for both tls and raw_buffer than the
	// endpoints in the given cluster can accept both plaintext and mutual TLS traffic
	return tlsTransportSocket && rawBufferTransportSocket
}

// GetMatchingTransportSockets returns the list of TransportSocket configs that matches the given endpoint
// metadata match criteria
func GetMatchingTransportSockets(cluster *envoy_config_cluster_v3.Cluster, endpointMetadataMatch map[string]interface{}) []*envoy_config_core_v3.TransportSocket {
	var transportSockets []*envoy_config_core_v3.TransportSocket
	for _, tsm := range cluster.GetTransportSocketMatches() {
		var sm map[string]interface{}
		if tsm.GetMatch() != nil {
			sm = tsm.GetMatch().AsMap()
		}

		if len(sm) == 0 {
			// empty transport socket match criteria matches any endpoint
			if tsm.GetTransportSocket() != nil {
				transportSockets = append(transportSockets, tsm.GetTransportSocket())
			}
			continue // check next tsm
		}

		if len(endpointMetadataMatch) == len(sm) {
			ok := true
			for k, v := range sm {
				//nolint:forcetypeassert
				if endpointMetadataMatch[k].(string) != v.(string) {
					ok = false
					break
				}
			}
			if ok && tsm.GetTransportSocket() != nil {
				transportSockets = append(transportSockets, tsm.GetTransportSocket())
			}
		}
	}

	// add cluster's default transport socket to the end of the list of defined
	if cluster.GetTransportSocket() != nil {
		transportSockets = append(transportSockets, cluster.GetTransportSocket())
	}
	return transportSockets
}
