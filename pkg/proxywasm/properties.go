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
	"encoding/binary"

	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
)

type propertyHolderWrapper struct {
	parentProperties api.PropertyHolder
	properties       api.PropertyHolder
}

func NewPropertyHolderWrapper(properties api.PropertyHolder, parentProperties api.PropertyHolder) api.WrappedPropertyHolder {
	return &propertyHolderWrapper{
		properties:       properties,
		parentProperties: parentProperties,
	}
}

func (w *propertyHolderWrapper) Properties() api.PropertyHolder {
	return w.properties
}

func (w *propertyHolderWrapper) ParentProperties() api.PropertyHolder {
	return w.parentProperties
}

func (w *propertyHolderWrapper) Get(key string) (interface{}, bool) {
	if v, found := w.properties.Get(key); found {
		return v, found
	}

	if w.parentProperties != nil {
		return w.parentProperties.Get(key)
	}

	return nil, false
}

func (w *propertyHolderWrapper) Set(key string, value interface{}) {
	w.properties.Set(key, value)
}

func Stringify(value interface{}) string {
	switch v := value.(type) {
	case bool:
		b := make([]byte, 1)
		val := 0
		if v {
			val = 1
		}
		b[0] = byte(val)
		return string(b)
	case int64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
		return string(b)
	case uint64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, v)
		return string(b)
	case int:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
		return string(b)
	case string:
		return v
	case map[string]interface{}: // convert if possible to map[string]string
		m := map[string]string{}
		for k, v := range v {
			if s, ok := v.(string); ok {
				m[k] = s
			}
		}
		return serializeMapToPairs(m)
	case map[string]string:
		return serializeMapToPairs(v)
	}

	return ""
}
