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

package config

import (
	"crypto/tls"
	"net"

	"google.golang.org/protobuf/types/known/structpb"
)

// NodeInfo stores information that identifies the client instance at management server
type NodeInfo struct {
	// Id of the client
	Id string
	// ClusterName is the name of the upstream cluster where the client is running
	ClusterName string

	// Metadata is metadata extending the client's identifier
	Metadata *structpb.Struct
}

// ClientConfig holds the data needed to connect to a management server
type ClientConfig struct {
	// ManagementServerAddress is the address of the management server
	ManagementServerAddress net.Addr

	// ClusterID is the name of the K8s cluster that hosts the management server
	ClusterID string

	// TLSConfig is the TLS configuration to connect with to the management server
	TLSConfig *tls.Config

	NodeInfo *NodeInfo

	// search domains available on the node where the caller workload resides
	SearchDomains []string
	// Note: istiod only supports incremental xDS partially. It still sends state of the word data but over the
	// incremental xDS protocol.See https://istio.io/latest/docs/reference/commands/pilot-discovery/ -> ISTIO_DELTA_XDS
	IncrementalXDSEnabled bool

	ManagementServerCapabilities *ManagementServerCapabilities
}

// ManagementServerCapabilities holds data that describe the capabilities of the management server
type ManagementServerCapabilities struct {

	// SotWWildcardSubscriptionSupported denotes whether '*' subscription is supported for LDS, CDS and NDS
	// when the management server runs in SotW mode
	SotWWildcardSubscriptionSupported bool

	// DeltaWildcardSubscriptionSupported denotes whether '*' subscription is supported for LDS, CDS and NDS
	// when the management server runs in Delta mode
	DeltaWildcardSubscriptionSupported bool

	// ClusterDiscoveryServiceProvided returns true if management server provides CDS
	ClusterDiscoveryServiceProvided bool

	// EndpointDiscoveryServiceProvided returns true if management server provides EDS
	EndpointDiscoveryServiceProvided bool

	// ListenerDiscoveryServiceProvided returns true if management server provides LDS
	ListenerDiscoveryServiceProvided bool

	// RouteConfigurationDiscoveryServiceProvided returns true if management server provides RDS
	RouteConfigurationDiscoveryServiceProvided bool

	// NameTableDiscoveryServiceSupported returns true if management server provides NDS
	NameTableDiscoveryServiceProvided bool
}
