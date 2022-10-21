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

package istio

import (
	"fmt"

	"wwwin-github.cisco.com/eti/nasp/pkg/istio/fb"
	"wwwin-github.cisco.com/eti/nasp/pkg/proxywasm/api"
	pwhttp "wwwin-github.cisco.com/eti/nasp/pkg/proxywasm/http"
)

const (
	FilterMetadataServiceName = "cluster_metadata.filter_metadata.istio.services.0.name"
	FilterMetadataServiceHost = "cluster_metadata.filter_metadata.istio.services.0.host"
)

type istioHttpHandlerMiddleware struct {
}

func NewIstioHTTPHandlerMiddleware() pwhttp.HandleMiddleware {
	return &istioHttpHandlerMiddleware{}
}

func (m *istioHttpHandlerMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) {
	if stream.Direction() != api.ListenerDirectionInbound {
		return
	}

	if serviceName, found := stream.Get("node.metadata.LABELS.service\\.istio\\.io/canonical-name"); found {
		if sn, ok := serviceName.(string); ok {
			stream.Set(FilterMetadataServiceName, sn)
		}

		if workloadNamespace, found := stream.Get("node.metadata.NAMESPACE"); found {
			stream.Set(FilterMetadataServiceHost, fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, workloadNamespace))
		}
	}
}

func (m *istioHttpHandlerMiddleware) AfterResponse(resp api.HTTPResponse, stream api.Stream) {
	if stream.Direction() != api.ListenerDirectionOutbound {
		return
	}

	if v, ok := stream.Get("upstream_peer"); ok {
		if value, ok := v.(string); ok {
			node := fb.GetRootAsFlatNode([]byte(value), 0)

			labels := make(map[string]string)
			label := new(fb.KeyVal)
			for i := 0; i < node.LabelsLength(); i++ {
				if node.Labels(label, i) {
					labels[string(label.Key())] = string(label.Value())
				}
			}

			stream.Set(FilterMetadataServiceName, labels["service.istio.io/canonical-name"])
			stream.Set(FilterMetadataServiceHost, fmt.Sprintf("%s.%s.svc.cluster.local", labels["service.istio.io/canonical-name"], node.Namespace()))
		}
	}
}

func (m *istioHttpHandlerMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) {
}
func (m *istioHttpHandlerMiddleware) BeforeResponse(resp api.HTTPResponse, stream api.Stream) {
}
