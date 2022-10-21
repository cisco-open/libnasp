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
	"sync"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type runtimeCreatorStore struct {
	engines sync.Map
}

func NewRuntimeCreatorStore() api.RuntimeCreatorStore {
	return &runtimeCreatorStore{
		engines: sync.Map{},
	}
}

func (s *runtimeCreatorStore) Get(name string) (api.WasmRuntimeCreator, bool) {
	if val, ok := s.engines.Load(name); ok {
		if creator, ok := val.(api.WasmRuntimeCreator); ok {
			return creator, true
		}
	}

	return nil, false
}

func (s *runtimeCreatorStore) Set(name string, creator api.WasmRuntimeCreator) {
	s.engines.Store(name, creator)
}
