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

package wasmtime

import (
	"context"
	"strings"

	"emperror.dev/errors"
	"github.com/bytecodealliance/wasmtime-go/v3"
	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type wasmtimeWasmModule struct {
	vm             *wasmtimeVM
	compiledModule *wasmtime.Module
	ctx            context.Context
	ctxCancel      context.CancelFunc
	logger         logr.Logger
}

func (m *wasmtimeWasmModule) Init() {}

func (m *wasmtimeWasmModule) SetLogger(logger logr.Logger) {
	m.logger = logger
}

func (m *wasmtimeWasmModule) NewInstance() (api.WasmInstance, error) {
	instance := NewWasmtimeInstance(m.ctx, m.vm.Runtime(), m.compiledModule, SetLoggerOption(m.logger))

	if err := instance.Init(RegisterV1Imports); err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not init module", "module")
	}

	return instance, nil
}

func (m *wasmtimeWasmModule) GetABINameList() []string {
	abiNameList := make([]string, 0)

	exportList := m.compiledModule.Exports()

	for _, export := range exportList {
		name := export.Name()
		if strings.HasPrefix(name, "proxy_abi") {
			abiNameList = append(abiNameList, name)
		}
	}

	return abiNameList
}
