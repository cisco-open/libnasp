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

	"github.com/sethvargo/go-envconfig"
)

type IstioEnvironment struct {
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

	ClusterID string `env:"CLUSTER_ID"`
	DNSDomain string `env:"DNS_DOMAIN"`
	MeshID    string `env:"MESH_ID"`

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

func (e *IstioEnvironment) GetNodePropertiesFromEnvironment() map[string]interface{} {
	labels := map[string]interface{}{}
	for k, v := range e.Labels {
		labels[k] = v
	}

	platformMetadata := map[string]interface{}{}
	for k, v := range e.PlatformMetadata {
		platformMetadata[k] = v
	}

	return map[string]interface{}{
		"id":      e.GetNodeID(),
		"cluster": e.GetClusterName(),
		"metadata": map[string]interface{}{
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
		},
	}
}
