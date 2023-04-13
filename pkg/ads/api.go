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
	"time"

	"emperror.dev/errors"
	backoffv4 "github.com/cenkalti/backoff/v4"
	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cisco-open/nasp/pkg/ads/config"
)

// HostNotFoundError is returned when the given host name is not found in Istio's NDS
type HostNotFoundError struct {
	HostName string
}

func (e *HostNotFoundError) Error() string {
	return fmt.Sprintf("hostname %q not found in Istio NDS", e.HostName)
}

func (e *HostNotFoundError) Is(target error) bool {
	//nolint:errorlint
	t, ok := target.(*HostNotFoundError)
	if !ok {
		return false
	}
	return t.HostName == e.HostName
}

// NoEndpointFoundError is returned when no service endpoint can be selected
// either due to there is no healthy endpoint or according to the load balancing weight
// no endpoint that accepts traffic can be selected
type NoEndpointFoundError struct{}

func (e *NoEndpointFoundError) Error() string {
	return "no healthy endpoint available"
}

func (e *NoEndpointFoundError) Is(target error) bool {
	//nolint:errorlint
	_, ok := target.(*NoEndpointFoundError)
	return ok
}

// Client knows how to interact with a Management server through ADS API over gRPC stream using either SotW or the
// incremental variant (see: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#four-variants)
// of the xDS protocol to provide resource configurations received from the management server to consumer applications.
type Client interface {
	GetHTTPClientPropertiesByHost(ctx context.Context, address string) (<-chan HTTPClientPropertiesResponse, error)
	GetTCPClientPropertiesByHost(ctx context.Context, address string) (<-chan ClientPropertiesResponse, error)
	GetListenerProperties(ctx context.Context, address string) (<-chan ListenerPropertiesResponse, error)

	// SetSearchDomains updates the search domain list specified at creation with domains
	SetSearchDomains(domains []string)

	// GetSearchDomains returns the currently configured search domain list for this client
	GetSearchDomains() []string

	// IncrementActiveRequestsCount increments the active requests count for the endpoint given its address.
	// Should be invoked when a connection is successfully established to the endpoint and a request is sent to it.
	IncrementActiveRequestsCount(address string)

	// DecrementActiveRequestsCount decrements the active requests count for the endpoint given its address
	// Should be invoked when the endpoint finished processing a request
	DecrementActiveRequestsCount(address string)

	ResolveHost(hostName string) ([]net.IP, error)
}

// ListenerPropertiesResponse contains the result for the API call
// to retrieve ListenerProperties for a given address.
type ListenerPropertiesResponse interface {
	ListenerProperties() ListenerProperties
	// Error in case retrieving ListenerProperties failed returns the error causing the failure
	Error() error
}

// ListenerProperties represents the properties of a listener
// of a workload
type ListenerProperties interface {
	// UseTLS returns true if the workload accepts TLS traffic
	UseTLS() bool

	// Permissive returns true if the workload accepts both plaintext and TLS traffic
	Permissive() bool

	// IsClientCertificateRequired returns true if client certificate is required
	IsClientCertificateRequired() bool

	// Metadata returns the metadata associated with this listener
	Metadata() map[string]interface{}
}

// ClientPropertiesResponse contains the result for the API call
// to retrieve ClientProperties for a given destination TCP service.
type ClientPropertiesResponse interface {
	ClientProperties() ClientProperties

	// Error in case retrieving ClientProperties failed returns the error causing the failure
	Error() error
}

// ClientProperties represents the properties for a workload to connect to a target service
type ClientProperties interface {
	// Address returns the address of the endpoint
	Address() (net.Addr, error)

	// UseTLS returns true if the target service accept TLS traffic
	UseTLS() bool

	// Permissive returns true if the target service accepts both TLS and plaintext traffic
	Permissive() bool

	// ServerName returns the SNI of the target HTTP service when connection to one of its endpoints
	// is made through TLS
	ServerName() string

	// Metadata returns the metadata associated with the target service
	Metadata() map[string]interface{}
}

// HTTPClientPropertiesResponse contains the result for the API call
// to retrieve HTTPClientProperties for a given destination HTTP service.
type HTTPClientPropertiesResponse interface {
	ClientProperties() ClientProperties

	// Error in case retrieving HTTPClientProperties failed returns the error causing the failure
	Error() error
}

// IgnoreHostNotFound returns nil if the root cause of the given err is of type HostNotFoundError, otherwise returns
// the original err
func IgnoreHostNotFound(err error) error {
	if err == nil {
		return nil
	}
	cause := errors.Cause(err)

	//nolint:errorlint
	if _, ok := cause.(*HostNotFoundError); ok {
		return nil
	}
	return err
}

func Connect(ctx context.Context, config *config.ClientConfig) (Client, error) {
	log := logr.FromContextOrDiscard(ctx)

	if config.ManagementServerCapabilities == nil {
		config.ManagementServerCapabilities = istioControlPlaneCapabilities()
	}
	apiClient := newClient(config)

	go func() {
		b := backoffv4.WithContext(backoffv4.NewConstantBackOff(15*time.Second), ctx)
		//nolint:errcheck
		backoffv4.Retry(func() error {
			err := apiClient.connectToStream(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled {
					return nil // exiting as the context has been cancelled
				}
				log.Error(err, "connection to stream broken, retrying ...")
			}

			return err
		}, b)
	}()

	return apiClient, nil
}

func istioControlPlaneCapabilities() *config.ManagementServerCapabilities {
	return &config.ManagementServerCapabilities{
		SotWWildcardSubscriptionSupported: true,

		// apparently istiod doesn't support '*' for CDS and LDS to subscribe to all resources
		// in case of CDS and LDS when incremental XDS is enabled but rather empty resources names needs to be passed to it
		// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#how-the-client-specifies-what-resources-to-return
		DeltaWildcardSubscriptionSupported: false,

		ListenerDiscoveryServiceProvided:           true,
		EndpointDiscoveryServiceProvided:           true,
		ClusterDiscoveryServiceProvided:            true,
		RouteConfigurationDiscoveryServiceProvided: true,
		NameTableDiscoveryServiceProvided:          true,
	}
}
