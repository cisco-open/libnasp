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

	"wwwin-github.cisco.com/eti/nasp/pkg/ads/internal/routeconfig"

	"emperror.dev/errors"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/go-logr/logr"
)

func (c *client) listRouteConfigurations(opts ...routeconfig.FilterOption) ([]*envoy_config_route_v3.RouteConfiguration, error) {
	resources := c.resources[resource_v3.RouteType].GetResources()
	routeConfigurations := make([]*envoy_config_route_v3.RouteConfiguration, 0, len(resources))
	for _, proto := range resources {
		if proto == nil {
			continue
		}
		routeConfiguration, ok := proto.(*envoy_config_route_v3.RouteConfiguration)
		if !ok {
			return nil, errors.Errorf("couldn't convert RouteConfiguration %q from protobuf message format to typed object", proto)
		}
		routeConfigurations = append(routeConfigurations, routeConfiguration)
	}

	return routeconfig.Filter(routeConfigurations, opts...), nil
}

func (c *client) getRouteConfig(name string) (*envoy_config_route_v3.RouteConfiguration, error) {
	resource := c.resources[resource_v3.RouteType].GetResource(name)
	if resource == nil {
		return nil, nil
	}

	routeConfig, ok := resource.(*envoy_config_route_v3.RouteConfiguration)
	if !ok {
		return nil, errors.Errorf("couldn't convert Route configuration %q from protobuf message format to typed object", resource)
	}

	return routeConfig, nil
}

func (c *client) onRouteConfigurationsUpdated(ctx context.Context) {
	log := logr.FromContextOrDiscard(ctx)

	routeConfigurations, err := c.listRouteConfigurations()
	if err != nil {
		log.Error(err, "couldn't list all route configurations")
		return
	}

	// update cluster stats for clusters the route configurations point to
	for _, r := range routeConfigurations {
		if r == nil {
			continue
		}

		for _, vh := range r.GetVirtualHosts() {
			for _, route := range vh.GetRoutes() {
				if route.GetRoute().GetClusterSpecifier() == nil {
					continue
				}
				if cluster, ok := route.GetRoute().GetClusterSpecifier().(*envoy_config_route_v3.RouteAction_Cluster); ok && cluster != nil {
					// if route config references a single upstream cluster and not multiple weighted upstream clusters
					// then reset weighted cluster selection counter
					log.V(2).Info("reset weighted upstream cluster selection counter as traffic is routed to single upstream cluster",
						"route configuration", r.GetName(),
						"route", route.GetName(),
						"cluster", cluster.Cluster)
					c.clustersStats.ResetWeightedClusterSelectionCounter(cluster.Cluster)
				}
			}
		}
	}
}
