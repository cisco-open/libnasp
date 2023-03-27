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
	"net/url"
	"reflect"
	"strconv"

	"github.com/cisco-open/nasp/pkg/ads/internal/listener"

	"github.com/cisco-open/nasp/pkg/ads/internal/virtualhost"

	"github.com/cisco-open/nasp/pkg/ads/internal/route"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"emperror.dev/errors"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type httpClientPropertiesResponse struct {
	result ClientProperties
	err    error
}

func (p *httpClientPropertiesResponse) ClientProperties() ClientProperties {
	return p.result
}

func (p *httpClientPropertiesResponse) Error() error {
	return p.err
}

// httpClientPropertiesObservable emits httpClientProperties changes wrapped in httpClientPropertiesResponse to
// interested clients (observers)
type httpClientPropertiesObservable struct {
	apiResultObservable[getHTTPClientPropertiesByHostInput, *httpClientPropertiesResponse]
}

func (c *client) GetHTTPClientPropertiesByHost(ctx context.Context, address string) (<-chan HTTPClientPropertiesResponse, error) {
	url, err := url.Parse("http://" + address)
	if err != nil {
		return nil, err
	}
	host := url.Hostname()
	port := url.Port()
	if host == "" {
		return nil, errors.New("missing host or IP")
	}
	if port == "" {
		port = "80"
	}
	portValue, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return nil, errors.Errorf("wrong port number format: %s", port)
	}

	input := getHTTPClientPropertiesByHostInput{
		host: host,
		port: uint32(portValue),
	}

	ch := make(chan HTTPClientPropertiesResponse, 1)
	go func() {
		defer close(ch)

		resultUpdated, cancel := c.httpClientPropertiesResults().registerForUpdates(input)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case <-resultUpdated:
				if r, ok := c.httpClientPropertiesResults().get(input); ok {
					sendLatest[HTTPClientPropertiesResponse](ctx, r, ch)
					// Exit after the first successful response as callers are not interested in updates.
					// They'll invoke the GetHTTPClientPropertiesByHost again when interested to get an updated result
					if r.Error() == nil {
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// httpClientPropertiesResults gets existing or creates a new empty httpClientPropertiesObservable instance
func (c *client) httpClientPropertiesResults() *httpClientPropertiesObservable {
	var observable *httpClientPropertiesObservable
	typ := reflect.TypeOf(observable)

	if x, ok := c.apiResults.Load(typ); !ok || x == nil {
		observable = &httpClientPropertiesObservable{}
		observable.computeResponse = c.getHTTPClientPropertiesByHostResponse
		observable.recomputeResultFuncName = "getHTTPClientPropertiesByHost"

		c.apiResults.Store(typ, observable)
	} else {
		//nolint:forcetypeassert
		observable = x.(*httpClientPropertiesObservable)
	}

	return observable
}

type getHTTPClientPropertiesByHostInput struct {
	host string
	port uint32
}

func (c *client) getHTTPClientPropertiesByHostResponse(input getHTTPClientPropertiesByHostInput) *httpClientPropertiesResponse {
	if !c.isInitialized() {
		return nil
	}

	r, err := c.getHTTPClientPropertiesByHost(input)
	if err != nil {
		return &httpClientPropertiesResponse{
			err: err,
		}
	}

	return &httpClientPropertiesResponse{
		result: r,
		err:    err,
	}
}

func (c *client) getHTTPClientPropertiesByHost(input getHTTPClientPropertiesByHostInput) (ClientProperties, error) {
	if !c.isInitialized() {
		return nil, nil
	}

	hostIPs, err := c.resolveHost(input.host)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't resolve host %q", input.host)
	}

	listener, err := c.getHTTPOutboundListener(hostIPs, int(input.port))
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't get HTTP outbound listener for address: %s:%d", input.host, input.port)
	}

	cluster, route, err := c.getHttpClientTargetCluster(input.host, int(input.port), hostIPs, listener)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't get target cluster for address: %s:%d", input.host, input.port)
	}

	clientProps, err := c.newClientProperties(cluster, listener, route)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't create client properties for target service at %s:%d", input.host, input.port)
	}

	return clientProps, nil
}

// getHttpClientTargetCluster returns the upstream cluster which HTTP traffic is directed to when
// clients connect to host:port
func (c *client) getHttpClientTargetCluster(host string, port int, hostIPs []net.IP, targetListener *envoy_config_listener_v3.Listener) (*envoy_config_cluster_v3.Cluster, *envoy_config_route_v3.Route, error) {
	routeConfigName := listener.GetRouteConfigName(targetListener)
	if routeConfigName == "" {
		return nil, nil, errors.New("couldn't determine route config")
	}
	routeConfig, err := c.getRouteConfig(routeConfigName)
	if err != nil {
		return nil, nil, errors.WrapIff(err, "no route configuration with name %q found", routeConfigName)
	}

	if routeConfig == nil {
		return nil, nil, errors.Errorf("no route configuration with name %q found", routeConfigName)
	}

	// get the name of the upstream cluster client traffic is routed to from the route config
	virtualHostDomainMatches := make([]string, 0, 2*len(hostIPs))
	// 1. match by IP and port
	for _, hostIP := range hostIPs {
		virtualHostDomainMatches = append(virtualHostDomainMatches, net.JoinHostPort(hostIP.String(), strconv.Itoa(port)))
	}
	// 2. match by IP
	for _, hostIP := range hostIPs {
		virtualHostDomainMatches = append(virtualHostDomainMatches, hostIP.String())
	}

	if net.ParseIP(host) == nil {
		// 3. match by host:port
		virtualHostDomainMatches = append(virtualHostDomainMatches, net.JoinHostPort(host, strconv.Itoa(port)))

		// 4. match by host
		virtualHostDomainMatches = append(virtualHostDomainMatches, host)
	}

	var virtualHosts []*envoy_config_route_v3.VirtualHost
	routeFilters := []route.FilterOption{
		route.WithRoutingToUpstreamCluster(),
		route.MatchingPath("/"),
	}
	for _, vhDomainMatch := range virtualHostDomainMatches {
		virtualHosts = virtualhost.Filter(routeConfig.GetVirtualHosts(),
			virtualhost.MatchingDomain(vhDomainMatch),
			virtualhost.HasRoutesMatching(routeFilters...),
		)

		if len(virtualHosts) > 0 {
			break
		}
	}
	if len(virtualHosts) > 1 {
		return nil, nil, errors.Errorf("multiple matching virtual host found for address: %s:%d", host, port)
	}
	if len(virtualHosts) == 0 {
		return nil, nil, errors.Errorf("no matching virtual host found for address: %s:%d", host, port)
	}

	routes := route.Filter(virtualHosts[0].GetRoutes(), routeFilters...)
	if len(routes) > 1 {
		return nil, nil, errors.Errorf("multiple matching routes to upstream cluster found for outbound traffic to address: %s:%d", host, port)
	}
	if len(routes) == 0 {
		return nil, nil, errors.Errorf("no route to upstream cluster found for outbound traffic to address: %s:%d", host, port)
	}
	r := routes[0]

	clusterName := route.GetTargetClusterName(r, c.clustersStats)
	if clusterName == "" {
		return nil, nil, errors.Errorf("no upstream cluster found for outbound traffic to address: %s:%d", host, port)
	}

	cluster, err := c.getCluster(clusterName)
	if err != nil {
		return nil, nil, errors.WrapIff(err, "no upstream cluster with name %q found", clusterName)
	}
	if cluster == nil {
		return nil, nil, errors.Errorf("no upstream cluster with name %q found", clusterName)
	}

	return cluster, r, nil
}

func (c *client) getHTTPOutboundListener(hostIPs []net.IP, port int) (*envoy_config_listener_v3.Listener, error) {
	listenAddrFilters := make([]listener.FilterOption, len(hostIPs))
	for i := range hostIPs {
		listenAddrFilters[i] = listener.ListeningOn(net.TCPAddr{IP: hostIPs[i], Port: port})
	}

	listeners, err := c.listListeners(
		listener.MatchingTrafficDirection(envoy_config_core_v3.TrafficDirection_OUTBOUND),
		listener.HasNetworkFilter(wellknown.HTTPConnectionManager),
		listener.Or(listenAddrFilters...),
	)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't list HTTP outbound listeners for IPs %s and port %d", hostIPs, port)
	}

	if len(listeners) == 0 {
		// search for listener that listens on the catch-all IP: 0.0.0.0:<port>
		listeners, err = c.listListeners(
			listener.MatchingTrafficDirection(envoy_config_core_v3.TrafficDirection_OUTBOUND),
			listener.HasNetworkFilter(wellknown.HTTPConnectionManager),
			listener.ListeningOn(net.TCPAddr{IP: net.IPv4zero, Port: port}),
		)
		if err != nil {
			return nil, errors.WrapIff(err, "couldn't list HTTP outbound listeners for address: 0.0.0.0:%d", port)
		}
	}

	if len(listeners) > 1 {
		return nil, errors.New("multiple HTTP outbound listeners found")
	}

	if len(listeners) == 0 {
		return nil, errors.New("no HTTP outbound listener found")
	}

	return listeners[0], nil
}
