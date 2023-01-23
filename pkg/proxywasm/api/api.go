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

package api

import (
	"encoding/json"
	"net/http"

	"github.com/banzaicloud/proxy-wasm-go-host/api"
)

type IoBuffer = api.IoBuffer

type Context interface {
	PropertyHolder

	ID() int32
	NewContextID() int32
	GetOrCreateContext(rootID string) Context
}

type MetricHandler interface {
	HTTPHandler() http.Handler
	DefineMetric(metricType int32, name string) int32
	RecordMetric(metricID int32, value int64) error
	IncrementMetric(metricID int32, offset int64) error
}

type WrappedPropertyHolder interface {
	PropertyHolder
	Properties() PropertyHolder
	ParentProperties() PropertyHolder
}

type PropertyHolder interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type JsonnableMap map[string]interface{}

func (m JsonnableMap) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

type BinaryModule interface {
	Get() ([]byte, error)
}

type DataSource interface {
	Get() ([]byte, error)
}

type Closer interface {
	Close()
}

type Marshallable interface {
	Marshal() ([]byte, error)
}
