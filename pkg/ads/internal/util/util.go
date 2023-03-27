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

package util

import (
	"encoding/json"
	"sync"

	"google.golang.org/protobuf/types/known/anypb"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_network_tcp_proxy_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

// KeyValueCollection thread-safe collection of <K, V> items
type KeyValueCollection[K comparable, V comparable] struct {
	items map[K]V

	mu sync.RWMutex // sync access to fields being read and updated by multiple go routines
}

func NewKeyValueCollection[K comparable, V comparable]() *KeyValueCollection[K, V] {
	return &KeyValueCollection[K, V]{
		items: make(map[K]V),
	}
}

func (c *KeyValueCollection[K, S]) Get(key K) (S, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats, ok := c.items[key]
	return stats, ok
}

func (c *KeyValueCollection[K, S]) Set(key K, stats S) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = stats
}

func (c *KeyValueCollection[K, S]) Keep(items map[K]struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var remove []K

	for k := range c.items {
		if _, ok := items[k]; !ok {
			remove = append(remove, k)
		}
	}

	for _, r := range remove {
		delete(c.items, r)
	}
}

func (c *KeyValueCollection[K, S]) Lock() {
	c.mu.Lock()
}

func (c *KeyValueCollection[K, S]) Unlock() {
	c.mu.Unlock()
}

func (c *KeyValueCollection[K, S]) RLock() {
	c.mu.RLock()
}

func (c *KeyValueCollection[K, S]) RUnlock() {
	c.mu.RUnlock()
}

func (c *KeyValueCollection[K, S]) Items() map[K]S {
	return c.items
}

// AckSotWResponse creates a 'DiscoveryRequest' to tell the management server that the config change for the given resource type
// was accepted
func AckSotWResponse(resourceType resource_v3.Type, version, nonce string, resourceNames []string) proto.Message {
	r := &discovery_v3.DiscoveryRequest{
		TypeUrl:       resourceType,
		VersionInfo:   version,
		ResourceNames: resourceNames,
		ResponseNonce: nonce,
	}

	return r
}

// NackSotWResponse creates a 'DiscoveryRequest' to tell the management server that the config change for the given resource type
// was rejected due to the given error
func NackSotWResponse(resourceType resource_v3.Type, version, nonce string, resourceNames []string, err error) *discovery_v3.DiscoveryRequest {
	r := &discovery_v3.DiscoveryRequest{
		TypeUrl:       resourceType,
		VersionInfo:   version,
		ResourceNames: resourceNames,
		ResponseNonce: nonce,
		ErrorDetail: &rpcstatus.Status{
			Code:    int32(codes.Internal),
			Message: err.Error(),
		},
	}

	return r
}

// AckDeltaResponse creates a 'DeltaDiscoveryRequest' to tell the management server that the config change for the given resource type
// was accepted
func AckDeltaResponse(resourceType resource_v3.Type, nonce string) proto.Message {
	r := &discovery_v3.DeltaDiscoveryRequest{
		TypeUrl:       resourceType,
		ResponseNonce: nonce,
	}

	return r
}

// NackDeltaResponse creates a 'DeltaDiscoveryRequest' to tell the management server that the config change for the given resource type
// was rejected due to the given error
func NackDeltaResponse(resourceType resource_v3.Type, nonce string, err error) proto.Message {
	r := &discovery_v3.DeltaDiscoveryRequest{
		TypeUrl:       resourceType,
		ResponseNonce: nonce,
		ErrorDetail: &rpcstatus.Status{
			Code:    int32(codes.Internal),
			Message: err.Error(),
		},
	}

	return r
}

func GetTcpProxy(filter *envoy_config_listener_v3.Filter) *envoy_extensions_filters_network_tcp_proxy_v3.TcpProxy {
	if typedConfig := filter.GetTypedConfig(); typedConfig != nil {
		if config, err := filter.GetTypedConfig().UnmarshalNew(); err == nil {
			if tcpProxy, ok := config.(*envoy_extensions_filters_network_tcp_proxy_v3.TcpProxy); ok {
				return tcpProxy
			}
		}
	}
	return nil
}

func GetDownstreamTlsContext(ts *envoy_config_core_v3.TransportSocket) *auth.DownstreamTlsContext {
	if typedConfig := ts.GetTypedConfig(); typedConfig != nil {
		dsc := &auth.DownstreamTlsContext{}
		if err := anypb.UnmarshalTo(typedConfig, dsc, proto.UnmarshalOptions{}); err == nil {
			return dsc
		}
	}

	return nil
}

// GetUnifiedFilterMetadata returns the filter metadata stored under the 'metadata.typed_filter_metadata'
// and 'metadata.filter_metadata' of the given metadat.
// If a key is present on both the one from 'metadata.typed_filter_metadata' will be taken into account.
func GetUnifiedFilterMetadata(md *envoy_config_core_v3.Metadata) (map[string]interface{}, error) {
	rawMetadata := make(map[string]interface{})

	for k, v := range md.GetTypedFilterMetadata() {
		typedMetadataValue, err := v.UnmarshalNew()
		if err != nil {
			return nil, err
		}
		rawMetadata[k] = typedMetadataValue
	}

	untypedFilterMetadata := md.GetFilterMetadata()
	for k := range untypedFilterMetadata {
		if _, ok := rawMetadata[k]; !ok {
			rawMetadata[k] = untypedFilterMetadata[k]
		}
	}

	b, err := json.Marshal(rawMetadata)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]interface{})
	err = json.Unmarshal(b, &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// StringSliceSubtract returns the Get of a-b
func StringSliceSubtract(a, b []string) []string {
	if a == nil {
		return nil
	}

	if b == nil {
		return a
	}

	var onlyInA []string
	for i := range a {
		found := false
		for j := range b {
			if a[i] == b[j] {
				found = true
				break
			}
		}
		if !found {
			onlyInA = append(onlyInA, a[i])
		}
	}
	if len(onlyInA) == 0 {
		return nil
	}

	return onlyInA
}

// SortedStringSliceEqual returns true if the provided sorted strings are equal
func SortedStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
