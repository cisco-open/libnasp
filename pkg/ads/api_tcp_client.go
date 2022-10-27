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
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"

	"github.com/cisco-open/nasp/pkg/ads/internal/listener"
	"github.com/cisco-open/nasp/pkg/ads/internal/loadbalancer"

	"github.com/cisco-open/nasp/pkg/ads/internal/endpoint"

	"github.com/cisco-open/nasp/pkg/ads/internal/util"

	"github.com/cisco-open/nasp/pkg/ads/internal/cluster"
	routemeta "github.com/cisco-open/nasp/pkg/ads/internal/route"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"emperror.dev/errors"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type clientProperties struct {
	useTLS     bool
	permissive bool
	serverName string
	address    net.Addr
	metadata   map[string]interface{}
}

func (p *clientProperties) UseTLS() bool {
	return p.useTLS
}

func (p *clientProperties) Permissive() bool {
	return p.permissive
}

func (p *clientProperties) ServerName() string {
	return p.serverName
}

func (p *clientProperties) Address() (net.Addr, error) {
	if p.address == nil {
		return nil, &NoEndpointFoundError{}
	}
	return p.address, nil
}

func (p *clientProperties) Metadata() map[string]interface{} {
	return p.metadata
}

func (p *clientProperties) String() string {
	addr := ""
	if epAddr, _ := p.Address(); epAddr != nil {
		addr = epAddr.String()
	}
	return fmt.Sprintf("{serverName=%s, useTLS=%t, permissive=%t, address=%s}", p.ServerName(), p.UseTLS(), p.Permissive(), addr)
}

type clientPropertiesResponse struct {
	result ClientProperties
	err    error
}

func (p *clientPropertiesResponse) ClientProperties() ClientProperties {
	return p.result
}

func (p *clientPropertiesResponse) Error() error {
	return p.err
}

// tcpClientPropertiesObservable emits clientProperties changes wrapped in clientPropertiesResponse to
// interested clients (observers)
type tcpClientPropertiesObservable struct {
	apiResultObservable[getTCPClientPropertiesByHostInput, *clientPropertiesResponse]
}

func (c *client) GetTCPClientPropertiesByHost(ctx context.Context, address string) (<-chan ClientPropertiesResponse, error) {
	url, err := url.Parse("tcp://" + address)
	if err != nil {
		return nil, err
	}
	host := url.Hostname()
	port := url.Port()
	if host == "" {
		return nil, errors.New("missing host")
	}
	if port == "" {
		port = "80"
	}
	portValue, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return nil, errors.Errorf("wrong port number format: %s", port)
	}

	input := getTCPClientPropertiesByHostInput{
		host: host,
		port: uint32(portValue),
	}

	ch := make(chan ClientPropertiesResponse, 1)
	go func() {
		defer close(ch)

		resultUpdated, cancel := c.tcpClientPropertiesResults().registerForUpdates(input)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case <-resultUpdated:
				if r, ok := c.tcpClientPropertiesResults().get(input); ok {
					sendLatest[ClientPropertiesResponse](ctx, r, ch)
					// Exit after the first successful response as callers are not interested in updates.
					// They'll invoke the GetTCPClientPropertiesByHost again when interested to get an updated result
					if r.Error() == nil {
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// tcpClientPropertiesResults gets existing or creates a new empty tcpClientPropertiesObservable instance
func (c *client) tcpClientPropertiesResults() *tcpClientPropertiesObservable {
	var observable *tcpClientPropertiesObservable
	typ := reflect.TypeOf(observable)

	if x, ok := c.apiResults.Load(typ); !ok || x == nil {
		observable = &tcpClientPropertiesObservable{}
		observable.computeResponse = c.getTCPClientPropertiesByHostResponse
		observable.recomputeResultFuncName = "getTCPClientPropertiesByHost"

		c.apiResults.Store(typ, observable)
	} else {
		//nolint:forcetypeassert
		observable = x.(*tcpClientPropertiesObservable)
	}

	return observable
}

type getTCPClientPropertiesByHostInput struct {
	host string
	port uint32
}

func (c *client) getTCPClientPropertiesByHostResponse(input getTCPClientPropertiesByHostInput) *clientPropertiesResponse {
	if !c.isInitialized() {
		return nil
	}

	r, err := c.getTCPClientPropertiesByHost(input)
	if err != nil {
		return &clientPropertiesResponse{
			err: err,
		}
	}

	return &clientPropertiesResponse{
		result: r,
		err:    err,
	}
}

func (c *client) getTCPClientPropertiesByHost(input getTCPClientPropertiesByHostInput) (*clientProperties, error) {
	if !c.isInitialized() {
		return nil, nil
	}

	cluster, err := c.getTCPClientTargetCluster(input.host, int(input.port))
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't get target upstream cluster")
	}

	clientProps, err := c.newClientProperties(cluster, nil)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't create client properties for target service at %s:%d", input.host, input.port)
	}

	return clientProps, nil
}

// getTCPClientTargetCluster returns the Envoy upstream cluster which TCP traffic is directed to when
// clients connect to host:port
func (c *client) getTCPClientTargetCluster(host string, port int) (*envoy_config_cluster_v3.Cluster, error) {
	hostIPs, err := c.resolveHost(host)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't resolve host %q", host)
	}

	listener, err := c.getTCPOutboundListener(hostIPs, port)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't get TCP outbound listener for address: %s:%d", host, port)
	}

	var clusterName string
	for _, fc := range listener.GetFilterChains() {
		for _, f := range fc.GetFilters() {
			if f.GetName() != wellknown.TCPProxy {
				continue
			}

			tcpProxy := util.GetTcpProxy(f)
			if tcpProxy == nil {
				continue
			}

			if tcpProxy.GetCluster() != "" {
				clusterName = tcpProxy.GetCluster() // traffic is routed to single upstream cluster
				break
			}

			if tcpProxy.GetWeightedClusters() != nil {
				// traffic is routed to multiple upstream clusters according to cluster weights
				clustersWeightMap := make(map[string]uint32)
				for _, weightedCluster := range tcpProxy.GetWeightedClusters().GetClusters() {
					clustersWeightMap[weightedCluster.GetName()] = weightedCluster.GetWeight()
				}

				clusterName = cluster.SelectCluster(clustersWeightMap, c.clustersStats)
				if clusterName != "" {
					break
				}
			}
		}
	}

	if clusterName == "" {
		return nil, errors.Errorf("no cluster found for for outbound traffic for address: %s:%d", host, port)
	}

	cluster, err := c.getCluster(clusterName)
	if err != nil {
		return nil, errors.WrapIff(err, "no cluster with name %q found", clusterName)
	}
	if cluster == nil {
		return nil, errors.Errorf("no cluster with name %q found", clusterName)
	}

	return cluster, nil
}

func (c *client) getTCPOutboundListener(hostIPs []net.IP, port int) (*envoy_config_listener_v3.Listener, error) {
	listenAddrFilters := make([]listener.FilterOption, len(hostIPs))
	for i := range hostIPs {
		listenAddrFilters[i] = listener.ListeningOn(net.TCPAddr{IP: hostIPs[i], Port: port})
	}

	listeners, err := c.listListeners(
		listener.MatchingTrafficDirection(envoy_config_core_v3.TrafficDirection_OUTBOUND),
		listener.HasNetworkFilter(wellknown.TCPProxy),
		listener.Or(listenAddrFilters...),
	)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't list TCP outbound listeners for IPs %s and port %d", hostIPs, port)
	}

	if len(listeners) == 0 {
		// search for listener that listens on the catch-all IP: 0.0.0.0:<port>
		listeners, err = c.listListeners(
			listener.MatchingTrafficDirection(envoy_config_core_v3.TrafficDirection_OUTBOUND),
			listener.HasNetworkFilter(wellknown.TCPProxy),
			listener.ListeningOn(net.TCPAddr{IP: net.IPv4zero, Port: port}),
		)
		if err != nil {
			return nil, errors.WrapIff(err, "couldn't list TCP outbound listeners for address: 0.0.0.0:%d", port)
		}
	}

	if len(listeners) > 1 {
		return nil, errors.New("multiple TCP outbound listeners found")
	}
	if len(listeners) == 0 {
		return nil, errors.New("no TCP outbound listener found")
	}

	return listeners[0], nil
}

// newClientProperties returns a new clientProperties instance populated with data from the given cluster and route
// and selecting endpoint address according to the LB policy of the cluster
func (c *client) newClientProperties(cl *envoy_config_cluster_v3.Cluster, route *envoy_config_route_v3.Route) (*clientProperties, error) {
	cla, err := c.getClusterLoadAssignmentForCluster(cl)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't list endpoints, cluster name=%q", cl.GetName())
	}
	clusterMetadata, err := cluster.GetMetadata(cl)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't get cluster metadata, cluster name=%q", cl.GetName())
	}

	var routeMetadata map[string]interface{}
	if route != nil {
		routeMetadata, err = routemeta.GetMetadata(route)
		if err != nil {
			return nil, errors.WrapIff(err, "couldn't get route metadata, route name=%q", route.GetName())
		}
	}

	var lb loadbalancer.LoadBalancer

	//nolint:exhaustive
	switch cluster.GetLoadBalancingPolicy(cl) {
	case envoy_config_cluster_v3.Cluster_RANDOM:
		endpoints := endpoint.Filter(cla.GetEndpoints(), endpoint.HasSocketAddress(), endpoint.WithHealthyStatus())
		lb = loadbalancer.NewRandomLoadBalancer(endpoints)
	case envoy_config_cluster_v3.Cluster_ROUND_ROBIN:
		overProvisioningFactor := 1.4 // default over provisioning factor
		if cla.GetPolicy().GetOverprovisioningFactor() != nil {
			overProvisioningFactor = float64(cla.GetPolicy().GetOverprovisioningFactor().GetValue()) / 100.0
		}
		endpointsLoad := endpoint.GetLoadDistribution(cla.GetEndpoints(), overProvisioningFactor)
		endpoints := endpoint.Filter(cla.GetEndpoints(), endpoint.HasSocketAddress(), endpoint.WithHealthyStatus())
		lb = loadbalancer.NewWeightedRoundRobinLoadBalancer(
			endpoints,
			endpointsLoad,
			c.endpointsStats)
	}

	var endpointAddress net.Addr
	if lb != nil {
		endpointAddress = lb.NextEndpoint()
	} else {
		// if unsupported LB policy is configured return first healthy endpoint from the list
		endpoints := endpoint.Filter(cla.GetEndpoints(), endpoint.HasSocketAddress(), endpoint.WithHealthyStatus())
		if len(endpoints) > 0 {
			endpointAddress = endpoint.GetAddress(endpoints[0])
		}
	}

	metadata := map[string]interface{}{
		"cluster_name":     cl.GetName(),
		"cluster_metadata": clusterMetadata,
		"route_name":       route.GetName(),
		"route_metadata":   routeMetadata,
	}

	clientProps := &clientProperties{
		permissive: cluster.IsPermissive(cl),
		serverName: cluster.GetTlsServerName(cl),
		useTLS:     cluster.UsesTls(cl),
		address:    endpointAddress,
		metadata:   metadata,
	}

	return clientProps, nil
}
