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

	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
)

type wasmVM struct {
	api.WasmRuntime

	lock     sync.Mutex
	refCount uint32
	users    map[interface{}]struct{}

	closing bool
}

func NewWasmVM(runtime api.WasmRuntime) api.WasmVM {
	return &wasmVM{
		WasmRuntime: runtime,

		users: make(map[interface{}]struct{}),
	}
}

func (vm *wasmVM) Close() {
	vm.lock.Lock()
	defer vm.lock.Unlock()

	vm.closing = true

	for user := range vm.users {
		if closer, ok := user.(interface {
			Close()
		}); ok {
			closer.Close()
		}
		delete(vm.users, user)
		vm.refCount--
	}
}

func (vm *wasmVM) Acquire(user interface{}) {
	if vm.closing {
		return
	}

	vm.lock.Lock()
	defer vm.lock.Unlock()

	if _, ok := vm.users[user]; !ok {
		vm.users[user] = struct{}{}
		vm.refCount++
	}
}

func (vm *wasmVM) Release(user interface{}) {
	if vm.closing {
		return
	}

	vm.lock.Lock()
	defer vm.lock.Unlock()

	if _, ok := vm.users[user]; ok {
		delete(vm.users, user)
		vm.refCount--
	}
}
