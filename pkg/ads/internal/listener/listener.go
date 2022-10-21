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

package listener

import (
	"net"
	"sort"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	"github.com/cisco-open/nasp/pkg/ads/internal/util"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	http_connection_manager_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type FilterOption interface {
	Filter(*envoy_config_listener_v3.Listener) bool
}

type trafficDirectionFilter struct {
	trafficDirection envoy_config_core_v3.TrafficDirection
}

func (f *trafficDirectionFilter) Filter(listener *envoy_config_listener_v3.Listener) bool {
	return listener.GetTrafficDirection() == f.trafficDirection
}

func MatchingTrafficDirection(trafficDirection envoy_config_core_v3.TrafficDirection) *trafficDirectionFilter {
	return &trafficDirectionFilter{trafficDirection: trafficDirection}
}

type listeningOnFilter struct {
	address net.TCPAddr
}

func (f *listeningOnFilter) Filter(listener *envoy_config_listener_v3.Listener) bool {
	// TODO: should we consider other address types as well beside socket address like Pipe or EnvoyInternalAddress ?
	sockAddr := listener.GetAddress().GetSocketAddress()
	ip := net.ParseIP(sockAddr.GetAddress())
	port := sockAddr.GetPortValue()
	listenAddr := net.TCPAddr{IP: ip, Port: int(port)}

	return listenAddr.String() == f.address.String()
}

func ListeningOn(address net.TCPAddr) *listeningOnFilter {
	return &listeningOnFilter{address: address}
}

type hasNetworkFilter struct {
	filterName string
}

func (f *hasNetworkFilter) Filter(listener *envoy_config_listener_v3.Listener) bool {
	for _, fcm := range listener.GetFilterChains() {
		for _, filter := range fcm.GetFilters() {
			if filter.GetName() == f.filterName {
				return true
			}
		}
	}

	return false
}

func HasNetworkFilter(filterName string) *hasNetworkFilter {
	return &hasNetworkFilter{filterName: filterName}
}

type orFilterOption struct {
	filters []FilterOption
}

func (f *orFilterOption) Filter(listener *envoy_config_listener_v3.Listener) bool {
	for _, filter := range f.filters {
		if filter.Filter(listener) {
			return true
		}
	}
	return false
}

func Or(filters ...FilterOption) *orFilterOption {
	return &orFilterOption{filters: filters}
}

// Filter returns elements from the given listeners which match the given filters
func Filter(listeners []*envoy_config_listener_v3.Listener, opts ...FilterOption) []*envoy_config_listener_v3.Listener {
	if opts == nil {
		return listeners
	}

	filteredListeners := make([]*envoy_config_listener_v3.Listener, 0, len(listeners))
	for _, listener := range listeners {
		if listener == nil {
			continue
		}

		add := true
		for _, opt := range opts {
			if opt != nil {
				if !opt.Filter(listener) {
					add = false
					break
				}
			}
		}

		if add {
			filteredListeners = append(filteredListeners, listener)
		}
	}
	return filteredListeners
}

// GetRouteReferences returns the route configuration names that references the current listener resources
func GetRouteReferences(listeners []*envoy_config_listener_v3.Listener) []string {
	routeResourceNames := make(map[string]struct{})

	for _, listener := range listeners {
		routeConfigName := GetRouteConfigName(listener)
		if routeConfigName != "" {
			routeResourceNames[routeConfigName] = struct{}{}
		}
	}

	resourceNames := make([]string, 0, len(routeResourceNames))
	for k := range routeResourceNames {
		resourceNames = append(resourceNames, k)
	}

	sort.Strings(resourceNames)

	return resourceNames
}

// GetRouteConfigName returns the name of the route configuration in RDS that the specified http listener references
func GetRouteConfigName(listener *envoy_config_listener_v3.Listener) string {
	for _, chain := range listener.FilterChains {
		for _, filter := range chain.Filters {
			if filter.Name != wellknown.HTTPConnectionManager {
				continue
			}

			config := resource_v3.GetHTTPConnectionManager(filter)

			if config == nil {
				continue
			}

			if rds, ok := config.RouteSpecifier.(*http_connection_manager_v3.HttpConnectionManager_Rds); ok && rds != nil && rds.Rds != nil {
				return rds.Rds.GetRouteConfigName()
			}
		}
	}

	return ""
}

func GetMetadata(listener *envoy_config_listener_v3.Listener) (map[string]interface{}, error) {
	return util.GetUnifiedMetadata(listener.GetMetadata())
}
