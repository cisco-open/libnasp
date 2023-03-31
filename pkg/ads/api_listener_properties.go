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

	"github.com/cisco-open/nasp/pkg/ads/internal/filterchain"

	"github.com/cisco-open/nasp/pkg/ads/internal/listener"
	"github.com/cisco-open/nasp/pkg/ads/internal/util"

	"emperror.dev/errors"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	xds_filters "istio.io/istio/pilot/pkg/xds/filters"
)

type listenerProperties struct {
	useTLS                   bool
	permissive               bool
	requireClientCertificate bool
	metadata                 map[string]interface{}
	inboundListener          *envoy_config_listener_v3.Listener
}

func (lp *listenerProperties) UseTLS() bool {
	return lp.useTLS
}

func (lp *listenerProperties) Permissive() bool {
	return lp.permissive
}

func (lp *listenerProperties) IsClientCertificateRequired() bool {
	return lp.requireClientCertificate
}

func (lp *listenerProperties) Metadata() map[string]interface{} {
	return lp.metadata
}

func (lp *listenerProperties) NetworkFilters(networkFilterSelectOpts ...NetworkFilterSelectOption) ([]NetworkFilter, error) {
	return listenerNetworkFilters(lp.inboundListener, networkFilterSelectOpts...)
}

func (lp *listenerProperties) String() string {
	return fmt.Sprintf("{useTLS=%t, permissive=%t, isClientCertificateRequired=%t}", lp.useTLS, lp.permissive, lp.requireClientCertificate)
}

type listenerPropertiesResponse struct {
	result ListenerProperties
	err    error
}

func (p *listenerPropertiesResponse) ListenerProperties() ListenerProperties {
	return p.result
}

func (p *listenerPropertiesResponse) Error() error {
	return p.err
}

// listenerPropertiesObservable emits listenerProperties changes wrapped in listenerPropertiesResponse to
// interested clients (observers)
type listenerPropertiesObservable struct {
	apiResultObservable[getListenerPropertiesInput, *listenerPropertiesResponse]
}

// GetListenerProperties emits ListenerProperties changes wrapped in ListenerPropertiesResponse to the caller via
// the returned channel.
// The caller must signal through the passed in Context when it's not interested in changes any more and the channel
// can be closed. The channel is closed on ctx.Done()
func (c *client) GetListenerProperties(ctx context.Context, address string) (<-chan ListenerPropertiesResponse, error) {
	url, err := url.Parse("tcp://" + address)
	if err != nil {
		return nil, err
	}
	host := url.Hostname()
	port := url.Port()
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" {
		port = "80"
	}

	// validate port number
	portValue, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return nil, errors.Errorf("wrong port number format: %s", port)
	}

	input := getListenerPropertiesInput{
		host: host,
		port: uint32(portValue),
	}

	ch := make(chan ListenerPropertiesResponse, 1)
	go func() {
		defer close(ch)

		resultUpdated, cancel := c.listenerPropertiesResults().registerForUpdates(input)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case <-resultUpdated:
				if r, ok := c.listenerPropertiesResults().get(input); ok {
					sendLatest[ListenerPropertiesResponse](ctx, r, ch)
				}
			}
		}
	}()

	return ch, nil
}

// listenerPropertiesResults gets existing or creates a new empty listenerPropertiesObservable instance
func (c *client) listenerPropertiesResults() *listenerPropertiesObservable {
	var observable *listenerPropertiesObservable
	typ := reflect.TypeOf(observable)

	if x, ok := c.apiResults.Load(typ); !ok || x == nil {
		observable = &listenerPropertiesObservable{}
		observable.computeResponse = c.getListenerPropertiesResponse
		observable.recomputeResultFuncName = "getListenerProperties"

		c.apiResults.Store(typ, observable)
	} else {
		//nolint:forcetypeassert
		observable = x.(*listenerPropertiesObservable)
	}

	return observable
}

type getListenerPropertiesInput struct {
	host string
	port uint32
}

func (c *client) getListenerPropertiesResponse(input getListenerPropertiesInput) *listenerPropertiesResponse {
	if !c.isInitialized() {
		return nil
	}

	r, err := c.getListenerProperties(input)
	if err != nil {
		return &listenerPropertiesResponse{
			err: err,
		}
	}

	return &listenerPropertiesResponse{
		result: r,
		err:    err,
	}
}

func (c *client) getListenerProperties(input getListenerPropertiesInput) (ListenerProperties, error) {
	var lp *listenerProperties
	var matchedListener *envoy_config_listener_v3.Listener

	if !c.isInitialized() {
		return nil, nil
	}

	listeners, err := c.listListeners(listener.MatchingTrafficDirection(envoy_config_core_v3.TrafficDirection_INBOUND))
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't list inbound listeners for address")
	}
	for _, lstnr := range listeners {
		// find listener's filter chains that are matching the first 2 steps of the rules described here
		// https://github.com/envoyproxy/go-control-plane/blob/v0.9.9/envoy/config/listener/v3/listener_components.pb.go#L211
		// which is enough to determine the properties of a workload listener
		filterChains, err := filterchain.Filter(lstnr,
			filterchain.WithDestinationPort(input.port),
			filterchain.WithDestinationIP(net.ParseIP(input.host)))

		if err != nil {
			return nil, err
		}

		if len(filterChains) == 0 && lstnr.GetDefaultFilterChain() != nil {
			// if no filter chains found, use default filter chain of the listener
			filterChains = append(filterChains, lstnr.GetDefaultFilterChain())
		}

		if len(filterChains) > 0 {
			if matchedListener != nil {
				return nil, errors.New("multiple listeners found")
			}
			matchedListener = lstnr

			tlsTransportProto := false
			rawBufferTransProto := false
			requireClientCertificate := false

			for _, fc := range filterChains {
				fcm := fc.GetFilterChainMatch()

				switch fcm.GetTransportProtocol() {
				case xds_filters.TLSTransportProtocol:
					tlsTransportProto = true
					if ts := util.GetDownstreamTlsContext(fc.TransportSocket); ts != nil {
						requireClientCertificate = ts.RequireClientCertificate.GetValue()
					}
				case xds_filters.RawBufferTransportProtocol:
					rawBufferTransProto = true
				}
			}

			metadata, err := listener.GetFilterMetadata(lstnr)
			if err != nil {
				return nil, errors.WrapIf(err, "couldn't get listener metadata")
			}

			lp = &listenerProperties{
				// if at least one of the filter chain uses TLS proto for matching than the listener support TLS communication
				useTLS: tlsTransportProto,
				// if there are filter chains for both TLS and raw_buffer proto than the listener is PERMISSIVE
				permissive: tlsTransportProto && rawBufferTransProto,
				// shows whether client certificate is required
				requireClientCertificate: requireClientCertificate,
				metadata:                 metadata,
				inboundListener:          matchedListener,
			}
		}
	}

	if lp == nil {
		return nil, errors.Errorf("couldn't find listener")
	}

	return lp, nil
}
