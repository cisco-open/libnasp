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

package proxywasm

import (
	"regexp"
	"strings"
)

var statTags = []map[string]string{
	{
		"tag_name": "reporter",
		"regex":    "(reporter=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "cluster_name",
		"regex":    "^cluster\\.((.+?(\\..+?\\.svc\\.cluster\\.local)?)\\.)",
	},
	{
		"tag_name": "tcp_prefix",
		"regex":    "^tcp\\.((.*?)\\.)\\w+?$",
	},
	{
		"tag_name": "response_code",
		"regex":    "(response_code=\\.=(.+?);\\.;)|_rq(_(\\.d{3}))$",
	},
	{
		"tag_name": "response_code_class",
		"regex":    "_rq(_(\\dxx))$",
	},
	{
		"tag_name": "listener_address",
		"regex":    "^listener\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)",
	},
	{
		"tag_name": "mongo_prefix",
		"regex":    "^mongo\\.(.+?)\\.(collection|cmd|cx_|op_|delays_|decoding_)(.*?)$",
	},
	{
		"tag_name": "mysql_cluster",
		"regex":    "^mysql\\.((.*?)\\.)\\w+?$",
	},
	{
		"tag_name": "postgres_cluster",
		"regex":    "^postgres\\.((.*?)\\.)\\w+?$",
	},
	{
		"tag_name": "source_namespace",
		"regex":    "(source_namespace=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_workload",
		"regex":    "(source_workload=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_workload_namespace",
		"regex":    "(source_workload_namespace=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_principal",
		"regex":    "(source_principal=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_app",
		"regex":    "(source_app=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_version",
		"regex":    "(source_version=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_cluster",
		"regex":    "(source_cluster=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_namespace",
		"regex":    "(destination_namespace=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_workload",
		"regex":    "(destination_workload=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_workload_namespace",
		"regex":    "(destination_workload_namespace=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_principal",
		"regex":    "(destination_principal=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_app",
		"regex":    "(destination_app=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_version",
		"regex":    "(destination_version=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_service",
		"regex":    "(destination_service=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_service_name",
		"regex":    "(destination_service_name=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_service_namespace",
		"regex":    "(destination_service_namespace=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_port",
		"regex":    "(destination_port=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_cluster",
		"regex":    "(destination_cluster=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "request_protocol",
		"regex":    "(request_protocol=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "request_operation",
		"regex":    "(request_operation=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "request_host",
		"regex":    "(request_host=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "response_flags",
		"regex":    "(response_flags=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "grpc_response_status",
		"regex":    "(grpc_response_status=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "connection_security_policy",
		"regex":    "(connection_security_policy=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_canonical_service",
		"regex":    "(source_canonical_service=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_canonical_service",
		"regex":    "(destination_canonical_service=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "source_canonical_revision",
		"regex":    "(source_canonical_revision=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "destination_canonical_revision",
		"regex":    "(destination_canonical_revision=\\.=(.*?);\\.;)",
	},
	{
		"tag_name": "cache",
		"regex":    "(cache\\.(.+?)\\.)",
	},
	{
		"tag_name": "component",
		"regex":    "(component\\.(.+?)\\.)",
	},
	{
		"tag_name": "tag",
		"regex":    "(tag\\.(.+?);\\.)",
	},
	{
		"tag_name": "wasm_filter",
		"regex":    "(wasm_filter\\.(.+?)\\.)",
	},
	{
		"tag_name": "authz_enforce_result",
		"regex":    "rbac(\\.(allowed|denied))",
	},
	{
		"tag_name": "authz_dry_run_action",
		"regex":    "(\\.istio_dry_run_(allow|deny)_)",
	},
	{
		"tag_name": "authz_dry_run_result",
		"regex":    "(\\.shadow_(allowed|denied))",
	},
}

type ParsedMetric struct {
	Name   string
	Values map[string]string
}

func ParseEnvoyStatTag(tag string) ParsedMetric {
	tags := make(map[string]string)

	for _, st := range statTags {
		tagName := st["tag_name"]
		regex := st["regex"]

		r, _ := regexp.Compile(regex)

		res := r.FindStringSubmatch(tag)
		if len(res) > 1 {
			tag = strings.Replace(tag, res[1], "", 1)
			tags[tagName] = res[2]
		}
	}

	return ParsedMetric{
		Name:   tag,
		Values: tags,
	}
}
