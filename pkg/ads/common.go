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
	"sort"
	"sync"

	"emperror.dev/errors"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	ptypes "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/proto"
	istio_xds_v3 "istio.io/istio/pilot/pkg/xds/v3"
	istio_networking_nds_v1 "istio.io/istio/pkg/dns/proto"
)

type resourceSubscription struct {
	resourceNames map[string]struct{} // resource names the being subscribed for

	// used only by the SotW xDS protocol; these fields are ignored in case of incremental xDS protocol
	// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#ack-nack-and-resource-type-instance-version
	sotwResponseNonce   string // the nonce of the most recent successfully processed SotW subscription response
	sotwResponseVersion string // the version of the most recent successfully processed SotW subscription response
}

// resourceCache stores the latest information successfully received from Management server about subscribed resources of type T
type resourceCache struct {
	subscription resourceSubscription
	resources    map[string]proto.Message // resources received from management server

	initialized bool

	mu sync.RWMutex // sync access to fields being read and updated by multiple go routines
}

// Subscribe initializes subscription data with the given resource names to start receiving
// updates from the Management server for given resources.
func (s *resourceCache) Subscribe(resourceNames []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(resourceNames) == 0 {
		return
	}

	s.subscription.resourceNames = make(map[string]struct{}, len(resourceNames))
	for i := range resourceNames {
		s.subscription.resourceNames[resourceNames[i]] = struct{}{}
	}
}

func (s *resourceCache) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.initialized
}

func (s *resourceCache) ResetSubscription() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscription.resourceNames = nil
}

func (s *resourceCache) ResetInitialized() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.initialized = false
}

// GetSubscribedResourceNames returns the list of subscribed resources names
func (s *resourceCache) GetSubscribedResourceNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.subscription.resourceNames) == 0 {
		return nil
	}

	res := make([]string, 0, len(s.subscription.resourceNames))
	for k := range s.subscription.resourceNames {
		res = append(res, k)
	}

	sort.Strings(res)

	return res
}

// GetLastAckSotWResponseVersionInfo returns the version info of the most recent successfully processed subscription response
// not used with incremental xDS
func (s *resourceCache) GetLastAckSotWResponseVersionInfo() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.subscription.sotwResponseVersion
}

// StoreSotWResponseVersionInfo stores the version info of the most recent successfully processed subscription response
// not used with incremental xDS
func (s *resourceCache) StoreSotWResponseVersionInfo(versionInfo string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscription.sotwResponseVersion = versionInfo
}

// GetLastAckSotWResponseNonce returns the nonce of the most recent successfully processed subscription response
// not used with incremental xDS
func (s *resourceCache) GetLastAckSotWResponseNonce() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.subscription.sotwResponseNonce
}

// StoreSotWResponseNonce stores the nonce of the most recent successfully processed subscription response
// not used with incremental xDS
func (s *resourceCache) StoreSotWResponseNonce(nonce string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscription.sotwResponseNonce = nonce
}

// RemoveResources removes the resources identified by the given resource names
// from internal state
func (s *resourceCache) RemoveResources(resourceNames []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range resourceNames {
		// update resources
		delete(s.resources, resourceNames[i])
	}
}

// RemoveAllResources removes all resources
// from internal state
func (s *resourceCache) RemoveAllResources() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resources = nil
}

// AddResources unmarshals the given resources collection into typed objects
// and adds them to internal state
func (s *resourceCache) AddResources(resourceType resource_v3.Type, resources []*discovery_v3.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(resources) == 0 {
		return nil
	}

	typedResources := make(map[string]proto.Message, len(resources))
	for _, r := range resources {
		if r == nil || r.GetResource() == nil {
			continue
		}
		res := newEmptyResource(resourceType)
		err := r.GetResource().UnmarshalTo(res)
		if err != nil {
			return err
		}

		// ignore resources we are not interested in case of explicit resource subscription
		if !s.IsSubscribedFor(r.GetName()) {
			continue
		}

		if _, ok := typedResources[r.GetName()]; ok {
			return errors.Errorf("duplicate resource name %q", r.GetName())
		}

		typedResources[r.GetName()] = res
	}

	// update resources
	if s.resources == nil {
		s.resources = typedResources
	} else {
		for k, v := range typedResources {
			s.resources[k] = v
		}
	}

	if !s.initialized {
		s.initialized = true
	}

	return nil
}

func (s *resourceCache) AddIstioResources(resourceType resource_v3.Type, resources []*discovery_v3.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(resources) == 0 {
		return nil
	}

	typedResources := make(map[string]proto.Message, len(resources))
	for _, r := range resources {
		if r == nil || r.GetResource() == nil {
			continue
		}
		res := newEmptyResource(resourceType)
		err := r.GetResource().UnmarshalTo(res)
		if err != nil {
			return err
		}

		switch typRes := res.(type) {
		case *istio_networking_nds_v1.NameTable:
			for name, nameTableResource := range typRes.GetTable() {
				if nameTableResource == nil {
					continue
				}
				// ignore resources we are not interested in case of explicit resource subscription
				if !s.IsSubscribedFor(name) {
					continue
				}

				if _, ok := typedResources[name]; ok {
					return errors.Errorf("duplicate resource name %q", name)
				}

				typedResources[name] = nameTableResource
			}
		default:
			return errors.Errorf("unsupported Istio resource type: %T", typRes)
		}
	}

	// update resources
	if s.resources == nil {
		s.resources = typedResources
	} else {
		for k, v := range typedResources {
			s.resources[k] = v
		}
	}

	if !s.initialized {
		s.initialized = true
	}

	return nil
}

// SetResources unmarshals the given resources collection into typed objects
// and assigns it to the internal state
func (s *resourceCache) SetResources(resourceType resource_v3.Type, resources []*ptypes.Any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(resources) == 0 {
		s.resources = nil
		return nil
	}

	typedResources := make(map[string]proto.Message, len(resources))
	for _, r := range resources {
		if r == nil {
			continue
		}

		res := newEmptyResource(resourceType)
		err := r.UnmarshalTo(res)
		if err != nil {
			return err
		}

		var name string
		switch typRes := res.(type) {
		case *envoy_config_cluster_v3.Cluster:
			name = typRes.GetName()
		case *envoy_config_endpoint_v3.ClusterLoadAssignment:
			name = typRes.GetClusterName()
		case *envoy_config_listener_v3.Listener:
			name = typRes.GetName()
		case *envoy_config_route_v3.RouteConfiguration:
			name = typRes.GetName()
		default:
			return errors.Errorf("couldn't Get resource name, resource: %q", typRes)
		}

		// ignore resources we are not interested in case of explicit resource subscription
		if !s.IsSubscribedFor(name) {
			continue
		}

		if _, ok := typedResources[name]; ok {
			return errors.Errorf("duplicate resource name %q", name)
		}

		typedResources[name] = res
	}

	// Set resources
	if s.resources == nil || s.IsSubscribedForAll() {
		// in case we are subscribed for all resources than we need to replace
		// the old resources list with new one in order to pick up deletes
		s.resources = typedResources
	} else {
		for k, v := range typedResources {
			s.resources[k] = v
		}
	}

	if !s.initialized {
		s.initialized = true
	}

	return nil
}

// SetIstioResources unmarshals the given resources collection into typed objects
// and assigns it to the internal state
func (s *resourceCache) SetIstioResources(resourceType resource_v3.Type, resources []*ptypes.Any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(resources) == 0 {
		s.resources = nil
		return nil
	}

	typedResources := make(map[string]proto.Message, len(resources))
	for _, r := range resources {
		if r == nil {
			continue
		}

		res := newEmptyResource(resourceType)
		err := r.UnmarshalTo(res)
		if err != nil {
			return err
		}

		switch typRes := res.(type) {
		case *istio_networking_nds_v1.NameTable:
			for name, nameTableResource := range typRes.GetTable() {
				if nameTableResource == nil {
					continue
				}
				// ignore resources we are not interested in case of explicit resource subscription
				if !s.IsSubscribedFor(name) {
					continue
				}

				if _, ok := typedResources[name]; ok {
					return errors.Errorf("duplicate resource name %q", name)
				}

				typedResources[name] = nameTableResource
			}
		default:
			return errors.Errorf("unsupported Istio resource type: %T", typRes)
		}
	}

	// Set resources
	if s.resources == nil || s.IsSubscribedForAll() {
		// in case we are subscribed for all resources than we need to replace
		// the old resources list with new one in order to pick up deletes
		s.resources = typedResources
	} else {
		for k, v := range typedResources {
			s.resources[k] = v
		}
	}

	if !s.initialized {
		s.initialized = true
	}

	return nil
}

// GetResources returns the stored resource instances
func (s *resourceCache) GetResources() map[string]proto.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.resources) == 0 {
		return nil
	}
	result := make(map[string]proto.Message, len(s.resources))
	for k := range s.resources {
		v := proto.Clone(s.resources[k])
		result[k] = v
	}

	return result
}

// GetResource returns the stored resource instance with the given resourceName
func (s *resourceCache) GetResource(resourceName string) proto.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.resources) == 0 {
		return nil
	}

	resource := s.resources[resourceName]
	return proto.Clone(resource)
}

func (s *resourceCache) IsSubscribedForAll() bool {
	if _, subscribeAll := s.subscription.resourceNames["*"]; subscribeAll {
		return true // subscribed to all resources
	}
	return false
}

func (s *resourceCache) IsSubscribedFor(resourceName string) bool {
	if s.IsSubscribedForAll() {
		return true
	}

	_, subscribed := s.subscription.resourceNames[resourceName]
	return subscribed
}

func newEmptyResource(resourceType resource_v3.Type) proto.Message {
	switch resourceType {
	case resource_v3.ClusterType:
		return &envoy_config_cluster_v3.Cluster{}
	case resource_v3.EndpointType:
		return &envoy_config_endpoint_v3.ClusterLoadAssignment{}
	case resource_v3.ListenerType:
		return &envoy_config_listener_v3.Listener{}
	case resource_v3.RouteType:
		return &envoy_config_route_v3.RouteConfiguration{}
	case istio_xds_v3.NameTableType:
		return &istio_networking_nds_v1.NameTable{}
	default:
		return nil
	}
}
