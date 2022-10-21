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

package wazero

import (
	"context"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type wazeroWasmModule struct {
	vm             *wazeroVM
	compiledModule wazero.CompiledModule
	rawBytes       []byte
	ctx            context.Context
	ctxCancel      context.CancelFunc
	logger         logr.Logger
}

func (m *wazeroWasmModule) Init() {}

func (m *wazeroWasmModule) SetLogger(logger logr.Logger) {
	m.logger = logger
}

func (m *wazeroWasmModule) NewInstance() (api.WasmInstance, error) {
	instance := NewWazeroInstance(m.ctx, m.vm.Runtime(), SetLoggerOption(m.logger))
	ns := m.vm.Runtime().NewNamespace(m.ctx)

	if _, err := wasi_snapshot_preview1.NewBuilder(m.vm.Runtime()).Instantiate(m.ctx, ns); err != nil {
		return nil, errors.WrapIf(err, "could not instantiate wasi module")
	}

	if err := instance.Init(RegisterV1Imports, ns); err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not init module", "module", m.compiledModule.Name())
	}

	module, err := ns.InstantiateModule(m.ctx, m.compiledModule, wazero.NewModuleConfig().WithName(m.compiledModule.Name()))
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not instantiate module", "module", m.compiledModule.Name())
	}

	instance.SetModule(m.ctx, module)

	return instance, nil
}

func (m *wazeroWasmModule) GetABINameList() []string {
	abiNameList := make([]string, 0)

	exportList := m.compiledModule.ExportedFunctions()

	for _, export := range exportList {
		for _, name := range export.ExportNames() {
			if strings.HasPrefix(name, "proxy_abi") {
				abiNameList = append(abiNameList, name)
			}
		}
	}

	return abiNameList
}
