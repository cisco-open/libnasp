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
	"runtime"

	"github.com/bytecodealliance/wasmtime-go/v3"
	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type wasmtimeVM struct {
	engine *wasmtime.Engine
	ctx    context.Context
	logger logr.Logger
}

func NewVM(ctx context.Context, logger logr.Logger) api.WasmRuntime {
	engine := wasmtime.NewEngine()

	return &wasmtimeVM{
		engine: engine,
		ctx:    ctx,
		logger: logger.WithName("wasmtimeVM"),
	}
}

func (vm *wasmtimeVM) Init() {}

func (vm *wasmtimeVM) Name() string {
	return "wasmtime"
}

func (vm *wasmtimeVM) Runtime() *wasmtime.Engine {
	return vm.engine
}

func (vm *wasmtimeVM) NewModule(wasmBytes []byte) api.WasmModule {
	engine := wasmtime.NewEngine()
	preCompiledModule, err := wasmtime.NewModule(engine, wasmBytes)
	if err != nil {
		panic(err)
	}

	moduleBytes, err := preCompiledModule.Serialize()
	if err != nil {
		panic(err)
	}

	compiledModule, err := wasmtime.NewModuleDeserialize(vm.engine, moduleBytes)
	if err != nil {
		vm.logger.Error(err, "could not compile module")
		return nil
	}

	engine = nil
	preCompiledModule = nil
	moduleBytes = nil

	runtime.GC()

	ctx, ctxCancel := context.WithCancel(vm.ctx)

	return &wasmtimeWasmModule{
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		compiledModule: compiledModule,
		vm:             vm,
		logger:         vm.logger,
	}
}
