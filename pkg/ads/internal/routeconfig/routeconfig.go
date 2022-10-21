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

package routeconfig

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type FilterOption interface {
	Filter(*envoy_config_route_v3.RouteConfiguration) bool
}

// Filter returns elements from the given listeners which match the given filters
func Filter(routeConfigs []*envoy_config_route_v3.RouteConfiguration, opts ...FilterOption) []*envoy_config_route_v3.RouteConfiguration {
	if opts == nil {
		return routeConfigs
	}

	filteredRouteConfigs := make([]*envoy_config_route_v3.RouteConfiguration, 0, len(routeConfigs))
	for _, routeConfig := range routeConfigs {
		if routeConfig == nil {
			continue
		}

		add := true
		for _, opt := range opts {
			if opt != nil {
				if !opt.Filter(routeConfig) {
					add = false
					break
				}
			}
		}

		if add {
			filteredRouteConfigs = append(filteredRouteConfigs, routeConfig)
		}
	}
	return filteredRouteConfigs
}
