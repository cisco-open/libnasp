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

	"github.com/go-logr/logr"
	"github.com/tetratelabs/wazero"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type wazeroVM struct {
	runtime wazero.Runtime
	ctx     context.Context
	logger  logr.Logger
}

func NewVM(ctx context.Context, logger logr.Logger) api.WasmRuntime {
	runtime := wazero.NewRuntime(ctx)

	return &wazeroVM{
		runtime: runtime,
		ctx:     ctx,
		logger:  logger.WithName("wazeroVM"),
	}
}

func (vm *wazeroVM) Init() {}

func (vm *wazeroVM) Name() string {
	return "wazero"
}

func (vm *wazeroVM) Runtime() wazero.Runtime {
	return vm.runtime
}

func (vm *wazeroVM) NewModule(wasmBytes []byte) api.WasmModule {
	compiledModule, err := vm.runtime.CompileModule(context.Background(), wasmBytes)
	if err != nil {
		vm.logger.Error(err, "could not compile module")
		return nil
	}

	ctx, ctxCancel := context.WithCancel(vm.ctx)

	return &wazeroWasmModule{
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		compiledModule: compiledModule,
		rawBytes:       wasmBytes,
		vm:             vm,
		logger:         vm.logger,
	}
}
