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

package virtualhost

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/cisco-open/libnasp/pkg/ads/internal/route"
)

type FilterOption interface {
	Filter(*envoy_config_route_v3.VirtualHost) bool
}

type domainFilter struct {
	domain string
}

func (f *domainFilter) Filter(vh *envoy_config_route_v3.VirtualHost) bool {
	for _, domain := range vh.GetDomains() {
		if domain == f.domain {
			return true
		}
	}
	return false
}

func MatchingDomain(domain string) *domainFilter {
	return &domainFilter{domain: domain}
}

type routeFilter struct {
	routeFilters []route.FilterOption
}

func (f *routeFilter) Filter(vh *envoy_config_route_v3.VirtualHost) bool {
	routes := route.Filter(vh.GetRoutes(), f.routeFilters...)
	return len(routes) != 0
}
func HasRoutesMatching(routeFilters ...route.FilterOption) *routeFilter {
	return &routeFilter{routeFilters: routeFilters}
}

// Filter returns elements from the given virtualHosts which match the given filters
func Filter(virtualHosts []*envoy_config_route_v3.VirtualHost, opts ...FilterOption) []*envoy_config_route_v3.VirtualHost {
	if opts == nil {
		return virtualHosts
	}

	var filteredVirtualHosts []*envoy_config_route_v3.VirtualHost
	for _, vh := range virtualHosts {
		if vh == nil {
			continue
		}

		add := true
		for _, opt := range opts {
			if opt != nil {
				if !opt.Filter(vh) {
					add = false
					break
				}
			}
		}

		if add {
			filteredVirtualHosts = append(filteredVirtualHosts, vh)
		}
	}
	return filteredVirtualHosts
}
