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

	"emperror.dev/errors"
	"github.com/go-logr/logr"

	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
)

type vmStore struct {
	runtimeCreators api.RuntimeCreatorStore
	logger          logr.Logger

	vms sync.Map
}

type VMKey string

func NewVMStore(runtimeCreators api.RuntimeCreatorStore, logger logr.Logger) api.VMStore {
	return &vmStore{
		runtimeCreators: runtimeCreators,
		logger:          logger.WithName("vmStore"),

		vms: sync.Map{},
	}
}

func (s *vmStore) GetOrCreateVM(vmConfig api.WasmVMConfig) (api.WasmVM, error) {
	details := []interface{}{"id", vmConfig.ID, "runtime", vmConfig.Runtime}

	s.logger.V(2).Info("get or create vm", details...)

	key, err := vmConfig.GetVMKey()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if value, ok := s.vms.Load(key); ok {
		if vm, ok := value.(api.WasmVM); ok {
			s.logger.V(2).Info("vm found in store", details...)
			return vm, nil
		}
	}

	if vmConfig.Runtime == "" {
		return nil, errors.New("runtime must be specified")
	}

	creator, ok := s.runtimeCreators.Get(vmConfig.Runtime)
	if !ok {
		return nil, errors.NewWithDetails("could not find engine", "runtime", vmConfig.Runtime)
	}

	vm := NewWasmVM(creator())
	s.vms.Store(key, vm)

	s.logger.V(2).Info("vm created and stored", details...)

	return vm, nil
}
