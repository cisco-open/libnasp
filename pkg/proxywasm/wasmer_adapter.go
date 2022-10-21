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
	"github.com/go-logr/logr"
	"mosn.io/proxy-wasm-go-host/proxywasm/common"
	proxywasm "mosn.io/proxy-wasm-go-host/proxywasm/v1"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type wasmerAdapter struct {
	common.WasmVM
}

func NewWasmerAdapter(wasmerVM common.WasmVM) api.WasmRuntime {
	return &wasmerAdapter{
		WasmVM: wasmerVM,
	}
}

func (s *wasmerAdapter) NewModule(wasmBytes []byte) api.WasmModule {
	return &wasmModuleAdapter{
		WasmModule: s.WasmVM.NewModule(wasmBytes),
	}
}

type wasmModuleAdapter struct {
	common.WasmModule
}

func (m *wasmModuleAdapter) NewInstance() (common.WasmInstance, error) {
	ins := m.WasmModule.NewInstance()
	proxywasm.RegisterImports(ins)

	return ins, nil
}

func (m *wasmModuleAdapter) SetLogger(logger logr.Logger) {
}
