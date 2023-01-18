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

package ads

import (
	"context"

	"github.com/cisco-open/nasp/pkg/ads/internal/endpoint"

	"github.com/cisco-open/nasp/pkg/ads/internal/cluster"
	"github.com/cisco-open/nasp/pkg/ads/internal/util"

	"emperror.dev/errors"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/go-logr/logr"
)

func (c *client) onClustersUpdated(ctx context.Context) {
	log := logr.FromContextOrDiscard(ctx)

	clusters, err := c.listClusters()
	if err != nil {
		log.Error(err, "couldn't list all upstream clusters")
		return
	}

	// update Endpoint subscription to reflect the updated cl list
	resourceNames := endpoint.GetClusterLoadAssignmentReferences(clusters)
	subscribedResourceNames := c.resources[resource_v3.EndpointType].GetSubscribedResourceNames()

	if util.SortedStringSliceEqual(resourceNames, subscribedResourceNames) {
		return
	}

	if len(resourceNames) > 0 {
		log.V(1).Info("subscribing to resources", "type", resource_v3.EndpointType, "resource names", resourceNames)
	}

	resourceNamesUnsubscribe := util.StringSliceSubtract(subscribedResourceNames, resourceNames)
	if len(resourceNamesUnsubscribe) > 0 {
		log.V(1).Info("unsubscribing from resources", "type", resource_v3.EndpointType, "resource names", resourceNamesUnsubscribe)
		if !c.config.IncrementalXDSEnabled {
			// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#deleting-resources
			// in case of SotW API we need to delete dependent resources explicitly
			c.resources[resource_v3.EndpointType].RemoveResources(resourceNamesUnsubscribe)
		}
	}

	c.resources[resource_v3.EndpointType].Subscribe(resourceNames)
	// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#how-the-client-specifies-what-resources-to-return
	if c.config.IncrementalXDSEnabled {
		c.enqueueDeltaSubscriptionRequest(
			ctx,
			resource_v3.EndpointType,
			resourceNames,
			resourceNamesUnsubscribe,
		)
	} else {
		c.enqueueSotWSubscriptionRequest(
			ctx,
			resource_v3.EndpointType,
			resourceNames,
			c.resources[resource_v3.EndpointType].GetLastAckSotWResponseVersionInfo(),
			c.resources[resource_v3.EndpointType].GetLastAckSotWResponseNonce())
	}

	//  update cl stats and endpoint stats
	keepClustersStats := make(map[string]struct{})
	for _, cl := range clusters {
		if cl == nil {
			continue
		}

		if _, ok := c.clustersStats.Get(cl.GetName()); !ok {
			c.clustersStats.Set(cl.GetName(), &cluster.Stats{})
		}

		keepClustersStats[cl.GetName()] = struct{}{}

		name := endpoint.GetClusterLoadAssignmentReference(cl)
		if name == "" || !c.resources[resource_v3.EndpointType].IsInitialized() {
			continue
		}

		cla, err := c.getClusterLoadAssignment(name)
		if err != nil {
			log.Error(err, "couldn't get cl load assignment", "name", name)
			continue
		}

		endpointWeightsAreEqual := endpoint.LoadBalancingWeightsAreEqual(cla.GetEndpoints())
		if (cluster.GetLoadBalancingPolicy(cl) != envoy_config_cluster_v3.Cluster_ROUND_ROBIN && cluster.GetLoadBalancingPolicy(cl) != envoy_config_cluster_v3.Cluster_LEAST_REQUEST) ||
			(cluster.GetLoadBalancingPolicy(cl) == envoy_config_cluster_v3.Cluster_LEAST_REQUEST && endpointWeightsAreEqual) {
			for _, localityLbEndpoints := range cla.GetEndpoints() {
				for _, lbEndpoint := range localityLbEndpoints.GetLbEndpoints() {
					address := endpoint.GetAddress(lbEndpoint.GetEndpoint())
					if address == nil {
						continue
					}
					c.endpointsStats.ResetRoundRobinCount(address.String())
				}
			}
		}
	}

	// remove clusters stats that correspond to removed clusters
	c.clustersStats.Keep(keepClustersStats)
}

func (c *client) listClusters(opts ...cluster.FilterOption) ([]*envoy_config_cluster_v3.Cluster, error) {
	clusterResources := c.resources[resource_v3.ClusterType].GetResources()
	clusters := make([]*envoy_config_cluster_v3.Cluster, 0, len(clusterResources))
	for _, proto := range clusterResources {
		if proto == nil {
			continue
		}
		cluster, ok := proto.(*envoy_config_cluster_v3.Cluster)
		if !ok {
			return nil, errors.Errorf("couldn't convert Cluster %q from protobuf message format to typed object", proto)
		}
		clusters = append(clusters, cluster)
	}

	return cluster.Filter(clusters, opts...), nil
}

func (c *client) getCluster(name string) (*envoy_config_cluster_v3.Cluster, error) {
	resource := c.resources[resource_v3.ClusterType].GetResource(name)
	if resource == nil {
		return nil, nil
	}

	cluster, ok := resource.(*envoy_config_cluster_v3.Cluster)
	if !ok {
		return nil, errors.Errorf("couldn't convert Cluster  %q from protobuf message format to typed object", resource)
	}
	return cluster, nil
}
