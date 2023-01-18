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
	"net"

	"github.com/cisco-open/nasp/pkg/ads/internal/endpoint"

	"emperror.dev/errors"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/go-logr/logr"
)

type listClusterLoadAssignmentsFilterOption interface {
	Filter(*envoy_config_endpoint_v3.ClusterLoadAssignment) bool
}

func (c *client) listClusterLoadAssignments(opts ...listClusterLoadAssignmentsFilterOption) ([]*envoy_config_endpoint_v3.ClusterLoadAssignment, error) {
	var clusterLoadAssignments []*envoy_config_endpoint_v3.ClusterLoadAssignment
	for _, proto := range c.resources[resource_v3.EndpointType].GetResources() {
		if proto == nil {
			continue
		}
		cla, ok := proto.(*envoy_config_endpoint_v3.ClusterLoadAssignment)
		if !ok {
			return nil, errors.Errorf("couldn't convert ClusterLoadAssignment %q from protobuf message format to typed object", proto)
		}

		add := true
		for _, opt := range opts {
			if opt != nil {
				if !opt.Filter(cla) {
					add = false
					break
				}
			}
		}

		if add {
			clusterLoadAssignments = append(clusterLoadAssignments, cla)
		}
	}

	return clusterLoadAssignments, nil
}

func (c *client) getClusterLoadAssignment(name string) (*envoy_config_endpoint_v3.ClusterLoadAssignment, error) {
	proto := c.resources[resource_v3.EndpointType].GetResource(name)
	if proto == nil {
		return nil, errors.Errorf("couldn't find ClusterLoadAssignment %q", name)
	}

	cla, ok := proto.(*envoy_config_endpoint_v3.ClusterLoadAssignment)
	if !ok {
		return nil, errors.Errorf("couldn't convert ClusterLoadAssignment %q from protobuf message format to typed object", proto)
	}

	return cla, nil
}

func (c *client) getClusterLoadAssignmentForCluster(cluster *envoy_config_cluster_v3.Cluster) (*envoy_config_endpoint_v3.ClusterLoadAssignment, error) {
	name := endpoint.GetClusterLoadAssignmentReference(cluster)

	cla, err := c.getClusterLoadAssignment(name)
	if err != nil {
		return nil, err
	}

	return cla, nil
}

func (c *client) onEndpointsUpdated(ctx context.Context) {
	log := logr.FromContextOrDiscard(ctx)

	// update endpoint stats
	clusterLoadAssignments, err := c.listClusterLoadAssignments()
	if err != nil {
		log.Error(err, "couldn't list all cluster load assignments")
		return
	}

	// remove endpoint stats that correspond to removed endpoints
	keepEndpointsStats := make(map[string]struct{})
	for _, cla := range clusterLoadAssignments {
		for _, ep := range cla.GetEndpoints() {
			for _, lbEndpoint := range ep.GetLbEndpoints() {
				address := lbEndpoint.GetEndpoint().GetAddress().GetSocketAddress()
				if address == nil {
					continue
				}

				netAddr := &net.TCPAddr{
					IP:   net.ParseIP(address.GetAddress()),
					Port: int(address.GetPortValue()),
				}

				if _, ok := c.endpointsStats.Get(netAddr.String()); !ok {
					c.endpointsStats.Set(netAddr.String(), &endpoint.Stats{})
				}

				keepEndpointsStats[netAddr.String()] = struct{}{}
			}
		}
	}
	c.endpointsStats.Keep(keepEndpointsStats)
}
