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
	"sync"

	"google.golang.org/grpc/credentials/insecure"

	"wwwin-github.cisco.com/eti/nasp/pkg/ads/config"

	client2 "wwwin-github.cisco.com/eti/nasp/pkg/ads/client"

	"wwwin-github.cisco.com/eti/nasp/pkg/ads/internal/apiresult"
	"wwwin-github.cisco.com/eti/nasp/pkg/ads/internal/util"

	"wwwin-github.cisco.com/eti/nasp/pkg/ads/internal/endpoint"

	"wwwin-github.cisco.com/eti/nasp/pkg/ads/internal/cluster"

	"emperror.dev/errors"
	core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
	istio_xds_v3 "istio.io/istio/pilot/pkg/xds/v3"
)

type client struct {
	config *config.ClientConfig

	resourceStream grpc.ClientStream

	requestsChan chan proto.Message // requests to be sent to management server

	resources            map[resource_v3.Type]*client2.ResourceCache // map of resource caches by resource type
	resourcesUpdatedChan chan struct{}                               // channel to notify about updates received from management server

	apiResults     *sync.Map
	endpointsStats *endpoint.EndpointsStats
	clustersStats  *cluster.ClustersStats

	mu sync.RWMutex // sync access to fields being read and updated by multiple go routines
}

// newClient returns a new `client` instance
func newClient(config *config.ClientConfig) *client {
	c := &client{
		config: config,
		resources: map[resource_v3.Type]*client2.ResourceCache{
			resource_v3.ClusterType:    {}, // envoy clusters resources received from management server
			resource_v3.EndpointType:   {}, // envoy cluster members (envoy endpoints resources) received from management server
			resource_v3.ListenerType:   {}, // envoy listeners resources received from management server
			resource_v3.RouteType:      {}, // envoy routes resources received from management server
			istio_xds_v3.NameTableType: {}, // istio name table resources - list of service entries and their associated IPs (including k8s services) - received from NDS
		},
		apiResults:     &sync.Map{},
		endpointsStats: endpoint.NewEndpointsStats(),
		clustersStats:  cluster.NewClustersStats(),
	}

	return c
}

// SetSearchDomains updates the search domain list specified at creation with domains
func (c *client) SetSearchDomains(domains []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.SearchDomains = domains
}

// GetSearchDomains returns the currently configured search domain list for this client
func (c *client) GetSearchDomains() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config.SearchDomains
}

func (c *client) isInitialized() bool {
	for resType, resourceCache := range c.resources {
		switch resType {
		case resource_v3.ClusterType:
			if !c.config.ManagementServerCapabilities.ClusterDiscoveryServiceProvided {
				continue
			}
		case resource_v3.EndpointType:
			if !c.config.ManagementServerCapabilities.EndpointDiscoveryServiceProvided {
				continue
			}
		case resource_v3.ListenerType:
			if !c.config.ManagementServerCapabilities.ListenerDiscoveryServiceProvided {
				continue
			}
		case resource_v3.RouteType:
			if !c.config.ManagementServerCapabilities.RouteConfigurationDiscoveryServiceProvided {
				continue
			}
		case istio_xds_v3.NameTableType:
			if !c.config.ManagementServerCapabilities.NameTableDiscoveryServiceProvided {
				continue
			}
		}

		if !resourceCache.IsInitialized() {
			return false
		}
	}

	return true
}

func (c *client) reset() {
	c.resourceStream = nil
	c.resources[resource_v3.ClusterType].ResetInitialized()
	c.resources[resource_v3.EndpointType].ResetInitialized()
	c.resources[resource_v3.ListenerType].ResetInitialized()
	c.resources[resource_v3.RouteType].ResetInitialized()
	c.resources[istio_xds_v3.NameTableType].ResetInitialized()

	c.resources[resource_v3.EndpointType].ResetSubscription() // reset endpoints subscribe list as will be set after CDS SotW is received
	c.resources[resource_v3.RouteType].ResetSubscription()    // reset routes config subscribe list as will be set after LDS SotW is received
}

func (c *client) connectToStream(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx)

	c.reset()

	managementServerAddress := c.config.ManagementServerAddress.String()
	log.Info("connecting to ADS management server", "address", managementServerAddress)

	var dialOpts []grpc.DialOption
	if c.config.TLSConfig != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(c.config.TLSConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.DialContext(
		ctx,
		c.config.ManagementServerAddress.String(),
		dialOpts...,
	)
	if err != nil {
		return errors.WrapIfWithDetails(err, "couldn't create dail context for management server",
			"address", managementServerAddress)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error(err, "couldn't close connection to ADS management server", "address", managementServerAddress)
		}
	}()

	adsClient := discovery_v3.NewAggregatedDiscoveryServiceClient(conn)
	var resourceStream grpc.ClientStream
	if c.config.IncrementalXDSEnabled {
		resourceStream, err = adsClient.DeltaAggregatedResources(ctx)
	} else {
		resourceStream, err = adsClient.StreamAggregatedResources(ctx)
	}
	if err != nil {
		return errors.WrapIfWithDetails(err, "couldn't connect to ADS management server", "address", managementServerAddress)
	}

	log.Info("connected to ADS management server", "address", managementServerAddress)

	requestsChan := make(chan proto.Message, 100)
	defer close(requestsChan)

	resourcesUpdatedChan := make(chan struct{})
	defer close(resourcesUpdatedChan)

	c.requestsChan = requestsChan
	c.resourcesUpdatedChan = resourcesUpdatedChan
	c.resourceStream = resourceStream

	// Subscribe for to all resources of Cluster and Listener type; we'll subscribe for Endpoint and RouteConfig resource types
	// when Cluster and Listener data is received from the Management server as those have a dependency on Cluster and Listener resources.
	// This is the flow what Envoy implements: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#api-flow

	// note: only CDS and LDS supports wildcard: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#how-the-client-specifies-what-resources-to-return
	c.resources[resource_v3.ClusterType].Subscribe([]string{"*"})
	c.resources[resource_v3.ListenerType].Subscribe([]string{"*"})
	c.resources[istio_xds_v3.NameTableType].Subscribe([]string{"*"}) // note: Istio's NDS seems to support wildcard subscription

	if c.config.IncrementalXDSEnabled {
		c.enqueueDeltaSubscriptionRequest(ctx, resource_v3.ClusterType, []string{"*"}, nil)
		c.enqueueDeltaSubscriptionRequest(ctx, resource_v3.ListenerType, []string{"*"}, nil)
		c.enqueueDeltaSubscriptionRequest(ctx, istio_xds_v3.NameTableType, []string{"*"}, nil)
	} else {
		c.enqueueSotWSubscriptionRequest(ctx, resource_v3.ClusterType, []string{"*"}, c.resources[resource_v3.ClusterType].GetLastAckSotWResponseVersionInfo(), c.resources[resource_v3.ClusterType].GetLastAckSotWResponseNonce())
		c.enqueueSotWSubscriptionRequest(ctx, resource_v3.ListenerType, []string{"*"}, c.resources[resource_v3.ListenerType].GetLastAckSotWResponseVersionInfo(), c.resources[resource_v3.ListenerType].GetLastAckSotWResponseNonce())
		c.enqueueSotWSubscriptionRequest(ctx, istio_xds_v3.NameTableType, []string{"*"}, c.resources[istio_xds_v3.NameTableType].GetLastAckSotWResponseVersionInfo(), c.resources[istio_xds_v3.NameTableType].GetLastAckSotWResponseNonce())
	}

	return c.handleStream(ctx)
}

func (c *client) enqueueDeltaSubscriptionRequest(
	ctx context.Context,
	resourceType resource_v3.Type,
	resourceNamesSubscribe, resourceNamesUnsubscribe []string) {
	log := logr.FromContextOrDiscard(ctx)

	log.V(1).Info("subscribing to resources", "type", resourceType, "resource names", resourceNamesSubscribe)

	if len(resourceNamesSubscribe) == 1 && resourceNamesSubscribe[0] == "*" {
		if !c.config.ManagementServerCapabilities.DeltaWildcardSubscriptionSupported {
			resourceNamesSubscribe = nil
		}
	}
	c.requestsChan <- &discovery_v3.DeltaDiscoveryRequest{
		TypeUrl: resourceType,
		Node: &core_v3.Node{
			Id:       c.config.NodeInfo.Id,
			Cluster:  c.config.NodeInfo.ClusterName,
			Metadata: c.config.NodeInfo.Metadata,
		},
		ResourceNamesSubscribe:   resourceNamesSubscribe,
		ResourceNamesUnsubscribe: resourceNamesUnsubscribe,
	}
}

func (c *client) enqueueSotWSubscriptionRequest(
	ctx context.Context,
	resourceType resource_v3.Type,
	resourceNamesSubscribe []string,
	version, nonce string) {
	log := logr.FromContextOrDiscard(ctx)

	log.V(1).Info("subscribing to resources", "type", resourceType, "resource names", resourceNamesSubscribe)

	if len(resourceNamesSubscribe) == 1 && resourceNamesSubscribe[0] == "*" {
		if !c.config.ManagementServerCapabilities.SotWWildcardSubscriptionSupported {
			resourceNamesSubscribe = nil
		}
	}

	c.requestsChan <- &discovery_v3.DiscoveryRequest{
		TypeUrl: resourceType,
		Node: &core_v3.Node{
			Id:       c.config.NodeInfo.Id,
			Cluster:  c.config.NodeInfo.ClusterName,
			Metadata: c.config.NodeInfo.Metadata,
		},
		VersionInfo:   version,
		ResourceNames: resourceNamesSubscribe,
		ResponseNonce: nonce,
	}
}

func (c *client) handleStream(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.SetLimit(3)
	g.Go(func() error {
		return c.handleStreamReceive(ctx)
	})
	g.Go(func() error {
		return c.handleStreamSend(ctx)
	})
	g.Go(func() error {
		return c.apiResultsRefresher(ctx)
	})

	return g.Wait()
}

// recv blocks until it receives a message into m or the stream is
// done. It returns io.EOF when the stream completes successfully. On
// any other error, the stream is aborted and the error contains the RPC
// status.
func (c *client) recv() (proto.Message, error) {
	switch stream := c.resourceStream.(type) {
	case discovery_v3.AggregatedDiscoveryService_StreamAggregatedResourcesClient:
		return stream.Recv()
	case discovery_v3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient:
		return stream.Recv()
	default:
		return nil, nil
	}
}

// handleStreamReceive waits for response messages from management server and processes them
// It exists when one of the following occurs:
// - the passed in context is cancelled
// - an error occurs during te processing of the response message
// - the gRPC stream to management server is broken/terminated thus no more messages can be received through the stream
func (c *client) handleStreamReceive(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := c.recv()
			if err != nil {
				return err
			}

			resourceTyp := c.getTypeUrl(resp)
			if err := c.validateResourceType(resourceTyp); err != nil {
				log.Info(err.Error())
				return nil
			}

			// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#the-xds-transport-protocol
			// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#incremental-xds
			var r proto.Message
			if err := c.handleResponse(ctx, resp); err != nil {
				// NACK
				r = c.nack(resp, err)
			} else {
				// ACK
				r = c.ack(resp)
			}

			// enqueue ACK/NACK request
			c.requestsChan <- r

			if err == nil {
				switch resourceTyp {
				case resource_v3.ClusterType:
					c.onClustersUpdated(ctx)
				case resource_v3.ListenerType:
					c.onListenersUpdated(ctx)
				case resource_v3.EndpointType:
					c.onEndpointsUpdated(ctx)
				case resource_v3.RouteType:
					c.onRouteConfigurationsUpdated(ctx)
				}

				c.resourcesUpdatedChan <- struct{}{}
			}
		}
	}
}

// handleStreamSend picks up requests from the 'client.requestsChan' channel and sends them to the management server
// It exists when one of the following occurs:
// - the passed in context is cancelled
// - the 'client.requestsChan' channel is closed
// - the gRPC stream to management server is broken/terminated thus no more request can be sent through the stream
func (c *client) handleStreamSend(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx)
	defer func() {
		if err := c.resourceStream.CloseSend(); err != nil {
			log.Error(err, "couldn't close send direction of the stream", "address", c.config.ManagementServerAddress)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case request := <-c.requestsChan:
			log.V(1).Info("sending request", "request", request)

			var err error
			switch req := request.(type) {
			case *discovery_v3.DiscoveryRequest:
				adsResourceStream, ok := c.resourceStream.(discovery_v3.AggregatedDiscoveryService_StreamAggregatedResourcesClient)
				if ok {
					err = adsResourceStream.Send(req)
				}
			case *discovery_v3.DeltaDiscoveryRequest:
				adsResourceStream, ok := c.resourceStream.(discovery_v3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient)
				if ok {
					err = adsResourceStream.Send(req)
				}
			}
			if err != nil {
				return err
			}
		}
	}
}

// handleResponse processes the  response message
// It exists when one of the following occurs:
// - an error occurs during te processing of the response message
// - the gRPC stream to management server is broken/terminated thus no more messages can be received through the stream
func (c *client) handleResponse(ctx context.Context, response proto.Message) error {
	switch resp := response.(type) {
	case *discovery_v3.DiscoveryResponse:
		if err := resp.Validate(); err != nil {
			// received response is not valid. return error to nack the response
			return err
		}

		if err := c.handleSotWResponse(ctx, resp); err != nil {
			return err
		}

		return nil

	case *discovery_v3.DeltaDiscoveryResponse:
		if err := resp.Validate(); err != nil {
			// received response is not valid. return error to nack the response
			return err
		}

		if err := c.handleDeltaResponse(ctx, resp); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (c *client) handleSotWResponse(_ context.Context, resp *discovery_v3.DiscoveryResponse) error {
	if r := c.resources[resp.GetTypeUrl()]; r != nil {
		switch resp.GetTypeUrl() {
		case resource_v3.ClusterType, resource_v3.EndpointType, resource_v3.ListenerType, resource_v3.RouteType:
			err := r.SetResources(resp.GetTypeUrl(), resp.GetResources())
			if err != nil {
				return err
			}
		case istio_xds_v3.NameTableType:
			err := r.SetIstioResources(istio_xds_v3.NameTableType, resp.GetResources())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *client) handleDeltaResponse(_ context.Context, resp *discovery_v3.DeltaDiscoveryResponse) error {
	if r := c.resources[resp.GetTypeUrl()]; r != nil {
		switch resp.GetTypeUrl() {
		case resource_v3.ClusterType, resource_v3.EndpointType, resource_v3.ListenerType, resource_v3.RouteType:
			if err := r.AddResources(resp.GetTypeUrl(), resp.GetResources()); err != nil {
				return err
			}
			r.RemoveResources(resp.GetRemovedResources())
		case istio_xds_v3.NameTableType:
			r.RemoveAllResources()
			if err := r.AddIstioResources(istio_xds_v3.NameTableType, resp.GetResources()); err != nil {
				return err
			}
		}
	}

	return nil
}

// apiResultsRefresher re-compustes all stored API results when resource updates
// are received from management server.
func (c *client) apiResultsRefresher(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.resourcesUpdatedChan:
			if c.isInitialized() {
				c.apiResults.Range(func(key any, value any) bool {
					if refresher, ok := value.(apiresult.Refresher); ok {
						refresher.Refresh(ctx)
					}

					return true
				})
			}
		}
	}
}

func (c *client) validateResourceType(resourceType resource_v3.Type) error {
	switch resourceType {
	case resource_v3.ClusterType,
		resource_v3.EndpointType,
		resource_v3.ListenerType,
		resource_v3.RouteType,
		istio_xds_v3.NameTableType:
		return nil
	default:
		return errors.Errorf("unsupported resource type %q", resourceType)
	}
}

func (c *client) getTypeUrl(response proto.Message) resource_v3.Type {
	switch resp := response.(type) {
	case *discovery_v3.DiscoveryResponse:
		return resp.GetTypeUrl()
	case *discovery_v3.DeltaDiscoveryResponse:
		return resp.GetTypeUrl()
	default:
		return ""
	}
}

func (c *client) getSubscribedResourceNames(resourceType resource_v3.Type) []string {
	if r := c.resources[resourceType]; r != nil {
		return r.GetSubscribedResourceNames()
	}

	return nil
}

// getLastAckSotWResponseVersionInfo returns the version info of the most recent successfully processed subscription response
// not used with incremental xDS
func (c *client) getLastAckSotWResponseVersionInfo(resourceType resource_v3.Type) string {
	if r := c.resources[resourceType]; r != nil {
		return r.GetLastAckSotWResponseVersionInfo()
	}

	return ""
}

// storeSotWResponseVersionInfo stores the version info of the most recent successfully processed subscription response
// not used with incremental xDS
func (c *client) storeSotWResponseVersionInfo(resourceType resource_v3.Type, versionInfo string) {
	if r := c.resources[resourceType]; r != nil {
		r.StoreSotWResponseVersionInfo(versionInfo)
	}
}

// getLastAckSotWResponseNonce returns the nonce of the most recent successfully processed subscription response
// not used with incremental xDS
func (c *client) getLastAckSotWResponseNonce(resourceType resource_v3.Type) string {
	if r := c.resources[resourceType]; r != nil {
		return r.GetLastAckSotWResponseNonce()
	}

	return ""
}

// storeSotWResponseNonce stores the nonce of the most recent successfully processed subscription response
// not used with incremental xDS
func (c *client) storeSotWResponseNonce(resourceType resource_v3.Type, nonce string) {
	if r := c.resources[resourceType]; r != nil {
		r.StoreSotWResponseNonce(nonce)
	}
}

// ack creates a request to tell the management server that the config change for the given resource type
// was accepted
func (c *client) ack(response proto.Message) proto.Message {
	// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#ack-nack-and-resource-type-instance-version
	switch resp := response.(type) {
	case *discovery_v3.DiscoveryResponse:
		c.storeSotWResponseVersionInfo(resp.GetTypeUrl(), resp.GetVersionInfo())
		c.storeSotWResponseNonce(resp.GetTypeUrl(), resp.GetNonce())

		return util.AckSotWResponse(
			resp.GetTypeUrl(),
			resp.GetVersionInfo(),
			resp.GetNonce(),
			c.getSubscribedResourceNames(resp.GetTypeUrl()))
	case *discovery_v3.DeltaDiscoveryResponse:
		return util.AckDeltaResponse(resp.GetTypeUrl(), resp.GetNonce())
	default:
		return nil
	}
}

// nack creates a request to tell the management server that the config change for the given resource type
// was rejected
func (c *client) nack(response proto.Message, err error) proto.Message {
	// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#ack-nack-and-resource-type-instance-version
	switch resp := response.(type) {
	case *discovery_v3.DiscoveryResponse:
		return util.NackSotWResponse(
			resp.GetTypeUrl(),
			c.getLastAckSotWResponseVersionInfo(resp.GetTypeUrl()),
			c.getLastAckSotWResponseNonce(resp.GetNonce()),
			c.getSubscribedResourceNames(resp.GetTypeUrl()),
			err)
	case *discovery_v3.DeltaDiscoveryResponse:
		return util.NackDeltaResponse(resp.GetTypeUrl(), resp.GetNonce(), err)
	default:
		return nil
	}
}
