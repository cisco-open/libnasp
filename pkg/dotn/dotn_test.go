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

package dotn_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cisco-open/libnasp/pkg/dotn"
)

func TestDotnGetterSetter(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	m := dotn.NewConcurrent()

	nodeID := "sidecar~10.20.4.77~echo-7ffdff66-f5wwf.default~default.svc.cluster.local"
	clusterName := "echo.default"
	nodeMetadata := map[string]interface{}{
		"NAME":              "echo-7ffdff66-f5wwf",
		"NAMESPACE":         "default",
		"OWNER":             "kubernetes://apis/apps/v1/namespaces/default/deployments/echo",
		"WORKLOAD_NAME":     "echo",
		"ISTIO_VERSION":     "1.11.4",
		"MESH_ID":           "mesh1",
		"CLUSTER_ID":        "waynz0r-0101-gke",
		"PLATFORM_METADATA": map[string]string{},
		"APP_CONTAINERS":    []interface{}{"server", "client"},
		"INSTANCE_IPS":      "10.20.4.77,fe80::185a:3eff:fe83:a7a3",
	}
	nodeLabels := map[string]interface{}{
		"security.istio.io/tlsMode":           "istio",
		"pod-template-hash":                   "7ffdff66",
		"service.istio.io/canonical-revision": "latest",
		"istio.io/rev":                        "cp-v111x.istio-system",
		"topology.istio.io/network":           "network1",
		"k8s-app":                             "echo",
		"service.istio.io/canonical-name":     "echo",
	}

	m.Set("node.id", nodeID)
	m.Set("node.cluster", clusterName)
	m.Set("node.metadata", nodeMetadata)
	m.Set("node.metadata.LABELS", nodeLabels)

	type tcase struct {
		key      string
		expected interface{}
	}

	// test clone, it should not override the original values
	p := m.Clone()
	p.Set("node.id", "bela")
	p.Set("node.metadata.NAME", "geza")

	for _, tc := range []tcase{
		{"node.id", nodeID},
		{"node.cluster", clusterName},
		{"node.metadata.NAME", nodeMetadata["NAME"]},
		{"node.metadata.NAMESPACE", nodeMetadata["NAMESPACE"]},
		{"node.metadata.OWNER", nodeMetadata["OWNER"]},
		{"node.metadata.WORKLOAD_NAME", nodeMetadata["WORKLOAD_NAME"]},
		{"node.metadata.ISTIO_VERSION", nodeMetadata["ISTIO_VERSION"]},
		{"node.metadata.MESH_ID", nodeMetadata["MESH_ID"]},
		{"node.metadata.CLUSTER_ID", nodeMetadata["CLUSTER_ID"]},
		{"node.metadata.LABELS", nodeLabels},
		{"node.metadata.PLATFORM_METADATA", nodeMetadata["PLATFORM_METADATA"]},
		{"node.metadata.APP_CONTAINERS.1", nodeMetadata["APP_CONTAINERS"].([]interface{})[1]}, //nolint:forcetypeassert
		{"node.metadata.INSTANCE_IPS", nodeMetadata["INSTANCE_IPS"]},
		{"node.metadata.LABELS.security\\.istio\\.io/tlsMode", nodeLabels["security.istio.io/tlsMode"]},
	} {
		assert.Equal(tc.expected, func() interface{} {
			v, _ := m.Get(tc.key)

			return v
		}())
	}
}

func TestDotnMerge(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	m := dotn.New()

	m.Set("destination", map[string]interface{}{
		"address": "127.0.0.1",
		"port":    56710,
	})

	m.Set("destination.address", "127.0.0.2")

	m.Set("destination", map[string]interface{}{
		"protocol": []interface{}{"tcp"},
	})

	m.Set("destination", map[string]interface{}{
		"protocol": []interface{}{"udp"},
	})

	assert.Equal("udp", func() interface{} {
		v, _ := m.Get("destination.protocol.1")

		return v
	}())

	assert.Equal("127.0.0.2", func() interface{} {
		v, _ := m.Get("destination.address")

		return v
	}())
}
