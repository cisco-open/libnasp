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

package dotn

import (
	"strconv"
	"strings"
)

type store map[string]interface{}
type array []interface{}

func New() *store {
	return &store{}
}

func (m store) Get(key string) (value interface{}, ok bool) {
	pieces := strings.Split(strings.ReplaceAll(key, "\\.", "\x00"), ".")

	value = m

	var val interface{}
	for i, p := range pieces {
		p = strings.ReplaceAll(p, "\x00", ".")
		ok = false
		switch v := value.(type) {
		case store:
			if val, ok = v[p]; ok {
				value = val
			}
		case map[string]interface{}:
			if val, ok = v[p]; ok {
				value = val
			}
		case map[string]string:
			if val, ok = v[p]; ok {
				value = val
			}
		case array:
			ip, err := strconv.ParseInt(p, 0, 8)
			if err != nil || int64(len(v)) < ip+1 {
				continue
			}
			value = v[ip]
			ok = true
		}

		if !ok {
			return nil, false
		}

		if len(pieces) == i+1 {
			switch value := value.(type) {
			case array:
				return []interface{}(value), true
			case store:
				return map[string]interface{}(value), true
			default:
				return value, true
			}
		}
	}

	return nil, false
}

func (m store) Set(key string, value interface{}) {
	pieces := strings.Split(key, ".")

	val := m

	for i, p := range pieces {
		p = strings.ReplaceAll(p, "\x00", ".")
		latest := len(pieces) == i+1
		if v, ok := val[p]; !ok { //nolint:nestif
			if latest {
				switch value := value.(type) {
				case []interface{}:
					val[p] = array(value)
				case map[string]interface{}:
					val[p] = store(map[string]interface{}{})
					m.Set(strings.Join(pieces[:i+1], "."), value)
				default:
					val[p] = value
				}
			} else {
				val[p] = store{}
				if v, ok := val[p].(store); ok {
					val = v
				}
			}
		} else {
			if latest {
				switch value := value.(type) {
				case map[string]interface{}: // iterate over map to merge
					for k, v := range value {
						m.Set(strings.Join(append(pieces[:i+1], strings.ReplaceAll(k, ".", "\x00")), "."), v)
					}
				case []interface{}: // merge array
					if v, ok := val[p].(array); ok {
						v = append(v, value...)
						val[p] = v
					}
				default:
					val[p] = value
				}
			} else {
				switch m := v.(type) {
				case store:
					val = m
				case map[string]interface{}:
					val = store(m)
				}
			}
		}
	}
}

func (m store) Convert() map[string]interface{} {
	d := make(map[string]interface{})
	for k, v := range m {
		if s, ok := v.(array); ok {
			v = []interface{}(s)
		}
		if s, ok := v.(store); ok {
			v = s.Convert()
		}
		if s, ok := v.(map[string]interface{}); ok {
			v = store(s).Convert()
		}
		d[k] = v
	}

	return d
}
