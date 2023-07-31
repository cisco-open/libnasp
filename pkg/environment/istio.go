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

package environment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type IstioEnvironment struct {
	Enabled           bool              `env:"ENABLED"`
	Type              string            `env:"TYPE"`
	PodName           string            `env:"POD_NAME"`
	PodNamespace      string            `env:"POD_NAMESPACE"`
	PodOwner          string            `env:"POD_OWNER"`
	PodServiceAccount string            `env:"POD_SERVICE_ACCOUNT"`
	WorkloadName      string            `env:"WORKLOAD_NAME"`
	AppContainers     []string          `env:"APP_CONTAINERS"`
	InstanceIPs       []string          `env:"INSTANCE_IP"`
	Labels            map[string]string `env:"LABELS"`
	PlatformMetadata  map[string]string `env:"PLATFORM_METADATA"`
	Network           string            `env:"NETWORK"`
	SearchDomains     []string          `env:"SEARCH_DOMAINS"`
	SecretTTL         time.Duration     `env:"SECRET_TTL"`

	AdditionalMetadata map[string]string `env:"ADDITIONAL_METADATA"`

	ClusterID   string `env:"CLUSTER_ID"`
	DNSDomain   string `env:"DNS_DOMAIN"`
	TrustDomain string `env:"TRUST_DOMAIN"`
	MeshID      string `env:"MESH_ID"`

	IstioCAAddress string `env:"ISTIO_CA_ADDR"`
	IstioVersion   string `env:"ISTIO_VERSION"`
	IstioRevision  string `env:"ISTIO_REVISION"`
}

func GetIstioEnvironment(prefix string) (*IstioEnvironment, error) {
	ctx := context.Background()

	var c IstioEnvironment
	if err := envconfig.ProcessWith(ctx, &c, envconfig.PrefixLookuper(prefix, envconfig.OsLookuper())); err != nil {
		return nil, err
	}

	return &c, nil
}

func (e *IstioEnvironment) GetNodeID() string {
	ip := ""
	if len(e.InstanceIPs) > 0 {
		ip = e.InstanceIPs[0]
	}
	return strings.Join([]string{
		e.Type, ip, fmt.Sprintf("%s.%s", e.PodName, e.PodNamespace), e.GetDNSServiceDomain(),
	}, "~")
}

func (e *IstioEnvironment) GetClusterName() string {
	return fmt.Sprintf("%s.%s", e.WorkloadName, e.PodNamespace)
}

func (e *IstioEnvironment) GetSpiffeID() string {
	return fmt.Sprintf("spiffe://%s/ns/%s/sa/%s", e.GetDNSDomain(), e.PodNamespace, e.PodServiceAccount)
}

func (e *IstioEnvironment) GetDNSDomain() string {
	domain := e.DNSDomain
	if len(domain) == 0 {
		domain = "cluster.local"
	}

	return domain
}

func (e *IstioEnvironment) GetDNSServiceDomain() string {
	return e.PodNamespace + ".svc." + e.GetDNSDomain()
}

func (e *IstioEnvironment) GetLabels() map[string]string {
	if e.Labels == nil {
		e.Labels = map[string]string{}
	}

	return e.Labels
}

func (e *IstioEnvironment) GetPlatformMetadata() map[string]string {
	if e.PlatformMetadata == nil {
		e.PlatformMetadata = map[string]string{}
	}

	return e.PlatformMetadata
}

func (e *IstioEnvironment) GetAdditionalMetadata() map[string]string {
	if e.AdditionalMetadata == nil {
		e.AdditionalMetadata = map[string]string{}
	}

	return e.AdditionalMetadata
}

func (e *IstioEnvironment) GetAppContainers() []string {
	if e.AppContainers == nil {
		e.AppContainers = []string{}
	}

	return e.AppContainers
}

func (e *IstioEnvironment) GetInstanceIPs() []string {
	if e.InstanceIPs == nil {
		e.InstanceIPs = []string{}
	}

	return e.InstanceIPs
}

func (e *IstioEnvironment) GetSearchDomains() []string {
	if e.SearchDomains == nil {
		e.SearchDomains = []string{}
	}

	return e.SearchDomains
}

func (e *IstioEnvironment) GetSecretTTL() time.Duration {
	if e.SecretTTL == 0 {
		e.SecretTTL = time.Hour * 24
	}

	return e.SecretTTL
}

func (e *IstioEnvironment) Override(otherEnv IstioEnvironment) {
	e.Type = otherEnv.Type
	e.PodName = otherEnv.PodName
	e.PodNamespace = otherEnv.PodNamespace
	e.PodOwner = otherEnv.PodOwner
	e.PodServiceAccount = otherEnv.PodServiceAccount
	e.WorkloadName = otherEnv.WorkloadName
	e.AppContainers = otherEnv.AppContainers
	e.InstanceIPs = otherEnv.InstanceIPs
	e.Labels = otherEnv.Labels
	e.PlatformMetadata = otherEnv.PlatformMetadata
	e.Network = otherEnv.Network
	e.SearchDomains = otherEnv.SearchDomains

	e.AdditionalMetadata = otherEnv.AdditionalMetadata

	e.ClusterID = otherEnv.ClusterID
	e.DNSDomain = otherEnv.DNSDomain
	e.MeshID = otherEnv.MeshID

	e.IstioCAAddress = otherEnv.IstioCAAddress
	e.IstioRevision = otherEnv.IstioRevision
	e.IstioVersion = otherEnv.IstioVersion
}

func (e *IstioEnvironment) GetNodePropertiesFromEnvironment() map[string]interface{} {
	labels := map[string]interface{}{}
	for k, v := range e.Labels {
		labels[k] = v
	}

	platformMetadata := map[string]interface{}{}
	for k, v := range e.PlatformMetadata {
		platformMetadata[k] = v
	}

	md := map[string]interface{}{
		"NAME":              e.PodName,
		"NAMESPACE":         e.PodNamespace,
		"OWNER":             e.PodOwner,
		"WORKLOAD_NAME":     e.WorkloadName,
		"ISTIO_VERSION":     e.IstioVersion,
		"MESH_ID":           e.MeshID,
		"CLUSTER_ID":        e.ClusterID,
		"NETWORK":           e.Network,
		"LABELS":            labels,
		"PLATFORM_METADATA": platformMetadata,
		"APP_CONTAINERS":    strings.Join(e.AppContainers, ","),
		"INSTANCE_IPS":      strings.Join(e.InstanceIPs, ","),
	}

	for k, v := range e.AdditionalMetadata {
		md[k] = v
	}

	return map[string]interface{}{
		"id":       e.GetNodeID(),
		"cluster":  e.GetClusterName(),
		"metadata": md,
	}
}
