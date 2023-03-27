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

package route

import (
	"regexp"
	"strings"

	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/cisco-open/nasp/pkg/ads/internal/cluster"
	"github.com/cisco-open/nasp/pkg/ads/internal/util"
)

type FilterOption interface {
	Filter(*envoy_config_route_v3.Route) bool
}

type routeToClusterActionFilter struct{}

func (f *routeToClusterActionFilter) Filter(route *envoy_config_route_v3.Route) bool {
	return route.GetRoute().GetCluster() != "" || route.GetRoute().GetWeightedClusters() != nil
}

func WithRoutingToUpstreamCluster() *routeToClusterActionFilter {
	return &routeToClusterActionFilter{}
}

type routePathFilter struct {
	path string
}

func (f *routePathFilter) Filter(route *envoy_config_route_v3.Route) bool {
	if route.GetMatch().GetPrefix() != "" || route.GetMatch().GetPath() != "" {
		ignoreCase := false
		if route.GetMatch().GetCaseSensitive() != nil {
			ignoreCase = !route.GetMatch().GetCaseSensitive().GetValue()
		}

		switch ignoreCase {
		case true:
			if route.GetMatch().GetPrefix() != "" {
				return strings.HasPrefix(strings.ToLower(f.path), strings.ToLower(route.GetMatch().GetPrefix()))
			}

			return strings.EqualFold(f.path, route.GetMatch().GetPath())

		case false:
			if route.GetMatch().GetPrefix() != "" {
				return strings.HasPrefix(f.path, route.GetMatch().GetPrefix())
			}

			return f.path == route.GetMatch().GetPath()
		}
	}

	if route.GetMatch().GetSafeRegex() != nil {
		match, _ := regexp.MatchString(route.GetMatch().GetSafeRegex().GetRegex(), f.path)
		return match
	}

	return false
}

func MatchingPath(path string) *routePathFilter {
	return &routePathFilter{path: path}
}

// Filter returns elements from the given routes which match the given filters
func Filter(routes []*envoy_config_route_v3.Route, opts ...FilterOption) []*envoy_config_route_v3.Route {
	if opts == nil {
		return routes
	}

	var filteredRoutes []*envoy_config_route_v3.Route
	for _, route := range routes {
		if route == nil {
			continue
		}

		add := true
		for _, opt := range opts {
			if opt != nil {
				if !opt.Filter(route) {
					add = false
					break
				}
			}
		}

		if add {
			filteredRoutes = append(filteredRoutes, route)
		}
	}
	return filteredRoutes
}

// GetTargetClusterName returns the name of the upstream cluster which the given route points to
func GetTargetClusterName(route *envoy_config_route_v3.Route, clustersStats *cluster.ClustersStats) string {
	routeAction := route.GetRoute()
	if routeAction != nil {
		if routeAction.GetCluster() != "" {
			return routeAction.GetCluster() // traffic is routed to single upstream cluster
		}

		if routeAction.GetWeightedClusters() != nil {
			// traffic is routed to multiple upstream clusters according to cluster weights
			clustersWeightMap := make(map[string]uint32)
			for _, weightedCluster := range routeAction.GetWeightedClusters().GetClusters() {
				clustersWeightMap[weightedCluster.GetName()] = weightedCluster.GetWeight().GetValue()
			}
			return cluster.SelectCluster(clustersWeightMap, clustersStats)
		}
	}

	return ""
}

// GetFilterMetadata returns the route configuration metadata stored under the 'metadata.typed_filter_metadata'
// and 'metadata.filter_metadata' of the given route configuration.
// If a key is present on both the one from 'metadata.typed_filter_metadata' will be taken into account.
func GetFilterMetadata(route *envoy_config_route_v3.Route) (map[string]interface{}, error) {
	return util.GetUnifiedFilterMetadata(route.GetMetadata())
}
