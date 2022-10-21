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

	"github.com/cisco-open/nasp/pkg/ads/internal/listener"
	"github.com/cisco-open/nasp/pkg/ads/internal/util"

	"emperror.dev/errors"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/go-logr/logr"
)

func (c *client) onListenersUpdated(ctx context.Context) {
	log := logr.FromContextOrDiscard(ctx)

	listeners, err := c.listListeners()
	if err != nil {
		log.Error(err, "couldn't list all listeners")
		return
	}

	// update routes subscription to reflect the updated listeners list
	resourceNames := listener.GetRouteReferences(listeners)
	subscribedResourceNames := c.resources[resource_v3.RouteType].GetSubscribedResourceNames()

	if util.SortedStringSliceEqual(resourceNames, subscribedResourceNames) {
		return
	}

	if len(resourceNames) > 0 {
		log.V(1).Info("subscribing to resources", "type", resource_v3.RouteType, "resource names", resourceNames)
	}

	resourceNamesUnsubscribe := util.StringSliceSubtract(subscribedResourceNames, resourceNames)
	if len(resourceNamesUnsubscribe) > 0 {
		log.V(1).Info("unsubscribing from resources", "type", resource_v3.RouteType, "resource names", resourceNamesUnsubscribe)
		if !c.config.IncrementalXDSEnabled {
			// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#deleting-resources
			// in case of SotW API we need to delete dependent resources explicitly
			c.resources[resource_v3.RouteType].RemoveResources(resourceNamesUnsubscribe)
		}
	}

	c.resources[resource_v3.RouteType].Subscribe(resourceNames)
	// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#how-the-client-specifies-what-resources-to-return
	if c.config.IncrementalXDSEnabled {
		c.enqueueDeltaSubscriptionRequest(
			ctx,
			resource_v3.RouteType,
			resourceNames,
			resourceNamesUnsubscribe,
		)
	} else {
		c.enqueueSotWSubscriptionRequest(
			ctx,
			resource_v3.RouteType,
			resourceNames,
			c.resources[resource_v3.RouteType].GetLastAckSotWResponseVersionInfo(),
			c.resources[resource_v3.RouteType].GetLastAckSotWResponseNonce())
	}

	//  update cluster stats which depend on listeners
	for _, listener := range listeners {
		for _, fc := range listener.GetFilterChains() {
			for _, f := range fc.GetFilters() {
				if f.GetName() == wellknown.TCPProxy {
					tcpProxy := util.GetTcpProxy(f)
					if tcpProxy == nil {
						continue
					}

					if tcpProxy.GetCluster() != "" {
						// traffic is routed to single upstream cluster thus reset weighted cluster stats
						c.clustersStats.ResetWeightedClusterSelectionCounter(tcpProxy.GetCluster())
					}
				}
			}
		}
	}
}

func (c *client) listListeners(opts ...listener.FilterOption) ([]*envoy_config_listener_v3.Listener, error) {
	resources := c.resources[resource_v3.ListenerType].GetResources()
	listeners := make([]*envoy_config_listener_v3.Listener, 0, len(resources))
	for _, proto := range resources {
		if proto == nil {
			continue
		}
		listener, ok := proto.(*envoy_config_listener_v3.Listener)
		if !ok {
			return nil, errors.Errorf("couldn't convert Listener %q from protobuf message format to typed object", proto)
		}
		listeners = append(listeners, listener)
	}
	return listener.Filter(listeners, opts...), nil
}
