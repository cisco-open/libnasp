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
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"

	"emperror.dev/errors"
	"github.com/bytecodealliance/wasmtime-go/v3"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"mosn.io/proxy-wasm-go-host/proxywasm/common"
)

var (
	ErrAddrOverflow         = errors.New("addr overflow")
	ErrInstanceNotStart     = errors.New("instance has not started")
	ErrInstanceAlreadyStart = errors.New("instance has already started")
	ErrInvalidParam         = errors.New("invalid param")
	ErrRegisterNotFunc      = errors.New("register a non-func object")
	ErrRegisterArgNum       = errors.New("register func with invalid arg num")
	ErrRegisterArgType      = errors.New("register func with invalid arg type")

	ErrPutUint32                   = errors.New("could not put uint32 to memory")
	ErrGetUint32                   = errors.New("could not get uint32 from memory")
	ErrPutByte                     = errors.New("could not put byte to memory")
	ErrGetByte                     = errors.New("could not get byte from memory")
	ErrPutData                     = errors.New("could not put data to memory")
	ErrGetData                     = errors.New("could not get data to memory")
	ErrGetExportsMemNotImplemented = errors.New("GetExportsMem not implemented")
	ErrMallocFunctionNotFound      = errors.New("could not find memory allocate function")
	ErrMalloc                      = errors.New("could not allocate memory")
)

type Instance struct {
	lock     sync.Mutex
	started  uint32
	refCount int
	stopCond *sync.Cond

	// user-defined data
	data interface{}

	store    *wasmtime.Store
	module   *wasmtime.Module
	instance *wasmtime.Instance
	memory   *wasmtime.Memory

	hostModules map[string][]hostFunc

	startFunctions  []string
	mallocFunctions []string

	logger logr.Logger
}

type InstanceOptions func(instance *Instance)

func StartFunctionsOption(names ...string) InstanceOptions {
	return func(instance *Instance) {
		instance.startFunctions = names
	}
}

func MallocFunctionsOption(names ...string) InstanceOptions {
	return func(instance *Instance) {
		instance.mallocFunctions = names
	}
}

func SetLoggerOption(logger logr.Logger) InstanceOptions {
	return func(instance *Instance) {
		instance.logger = logger
	}
}

func NewWasmtimeInstance(ctx context.Context, engine *wasmtime.Engine, module *wasmtime.Module, options ...InstanceOptions) *Instance {
	ins := &Instance{
		store:           wasmtime.NewStore(engine),
		module:          module,
		lock:            sync.Mutex{},
		hostModules:     make(map[string][]hostFunc),
		startFunctions:  []string{"_start", "_initialize"},
		mallocFunctions: []string{"proxy_on_memory_allocate", "malloc"},
	}
	ins.stopCond = sync.NewCond(&ins.lock)

	for _, option := range options {
		option(ins)
	}

	if ins.logger == (logr.Logger{}) {
		ins.logger = klog.Background()
	}

	return ins
}

func (w *Instance) Logger() logr.Logger {
	return w.logger
}

func (w *Instance) GetData() interface{} {
	return w.data
}

func (w *Instance) SetData(data interface{}) {
	w.data = data
}

func (w *Instance) Acquire() bool {
	w.lock.Lock()
	defer w.lock.Unlock()

	if !w.checkStart() {
		return false
	}

	w.refCount++

	return true
}

func (w *Instance) Release() {
	w.lock.Lock()
	if w.refCount > 0 {
		w.refCount--
	}

	if w.refCount <= 0 {
		w.stopCond.Broadcast()
	}
	w.lock.Unlock()
}

func (w *Instance) Lock(data interface{}) {
	w.lock.Lock()
	w.data = data
}

func (w *Instance) Unlock() {
	w.data = nil
	w.lock.Unlock()
}

func (w *Instance) GetModule() common.WasmModule {
	return nil
}

func (w *Instance) Init(f func(*Instance)) error {
	f(w)

	linker := wasmtime.NewLinker(w.store.Engine)

	err := linker.DefineWasi()
	if err != nil {
		return err
	}

	w.store.SetWasi(wasmtime.NewWasiConfig())

	for moduleName, hostModule := range w.hostModules {
		for _, hostFnc := range hostModule {
			err := linker.DefineFunc(w.store, moduleName, hostFnc.funcName, hostFnc.f)
			if err != nil {
				return err
			}
		}
	}

	w.instance, err = linker.Instantiate(w.store, w.module)
	if err != nil {
		return err
	}

	w.memory = w.instance.GetExport(w.store, "memory").Memory()

	return nil
}

func (w *Instance) Start() error {
	for _, fn := range w.startFunctions {
		f := w.instance.GetFunc(w.store, fn)
		if f == nil {
			continue
		}

		if _, err := f.Call(w.store); err != nil {
			w.HandleError(err)
			return err
		}

		atomic.StoreUint32(&w.started, 1)

		return nil
	}

	return errors.NewWithDetails("could not start instance: start function is not exported", "functions", w.startFunctions)
}

func (w *Instance) Stop() {
	go func() {
		w.lock.Lock()
		for w.refCount > 0 {
			w.stopCond.Wait()
		}
		_ = atomic.CompareAndSwapUint32(&w.started, 1, 0)
		// if err := w.instance.Close(w.moduleCtx); err != nil { // TODO?
		// 	w.Logger().Error(err, "could not close module", "module", w.module.Name())
		// }
		w.lock.Unlock()
	}()
}

// return true is Instance is started, false if not started.
func (w *Instance) checkStart() bool {
	return atomic.LoadUint32(&w.started) == 1
}

type hostFunc struct {
	funcName string
	f        interface{}
}

func (w *Instance) RegisterFunc(namespace string, funcName string, f interface{}) error {
	if w.checkStart() {
		return ErrInstanceAlreadyStart
	}

	if _, ok := w.hostModules[namespace]; !ok {
		w.hostModules[namespace] = []hostFunc{}
	}

	w.hostModules[namespace] = append(w.hostModules[namespace], hostFunc{funcName: funcName, f: f})

	return nil
}

func (w *Instance) Malloc(size int32) (uint64, error) {
	if !w.checkStart() {
		return 0, ErrInstanceNotStart
	}

	var f *wasmtime.Func
	mallocFuncNames := w.mallocFunctions
	var mallocFuncName string
	for _, fn := range mallocFuncNames {
		if f == nil {
			f = w.instance.GetFunc(w.store, fn)
		}
		if f != nil {
			mallocFuncName = fn
			break
		}
	}

	if f == nil {
		return 0, ErrMallocFunctionNotFound
	}

	malloc := &Call{
		Func:   f,
		name:   mallocFuncName,
		store:  w.store,
		logger: w.logger,
	}

	addr, err := malloc.Call(size)
	if err != nil {
		w.HandleError(err)
		return 0, err
	}

	if v, ok := addr.(int32); ok {
		return uint64(v), nil
	}

	return 0, ErrMalloc
}

type Call struct {
	*wasmtime.Func

	name string

	store *wasmtime.Store

	logger logr.Logger
}

func (c *Call) Call(args ...interface{}) (interface{}, error) {
	c.logger.V(3).Info(fmt.Sprintf("call module function %s", c.name))

	ret, err := c.Func.Call(c.store, args...)
	if err != nil {
		return 0, err
	}

	switch r := ret.(type) {
	case []wasmtime.Val:
		return r[0].I32(), nil
	default:
		return r, nil
	}
}

func (w *Instance) GetExportsFunc(funcName string) (common.WasmFunction, error) {
	if !w.checkStart() {
		return nil, ErrInstanceNotStart
	}

	f := &Call{
		Func:   w.instance.GetFunc(w.store, funcName),
		name:   funcName,
		store:  w.store,
		logger: w.logger,
	}

	return f, nil
}

func (w *Instance) GetExportsMem(memName string) ([]byte, error) {
	return nil, ErrGetExportsMemNotImplemented
}

func (w *Instance) GetMemory(addr uint64, size uint64) ([]byte, error) {
	memory := w.memory.UnsafeData(w.store)

	if w.memory.DataSize(w.store) <= uintptr(addr+size) {
		return nil, ErrGetData
	}

	return memory[uint32(addr) : uint32(addr)+uint32(size)], nil
}

func (w *Instance) PutMemory(addr uint64, size uint64, content []byte) error {
	if n := len(content); n < int(size) {
		size = uint64(n)
	}

	if w.memory.DataSize(w.store) <= uintptr(addr+size) {
		return ErrPutData
	}

	memory := w.memory.UnsafeData(w.store)

	copy(memory[uint32(addr):], content[:size])

	return nil
}

func (w *Instance) GetByte(addr uint64) (byte, error) {
	bytes, err := w.GetMemory(addr, 1)
	if err != nil {
		return 0, err
	}
	return bytes[0], nil
}

func (w *Instance) PutByte(addr uint64, b byte) error {
	return w.PutMemory(addr, 1, []byte{b})
}

func (w *Instance) GetUint32(addr uint64) (uint32, error) {
	data, err := w.GetMemory(addr, 4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(data), nil
}

func (w *Instance) PutUint32(addr uint64, value uint32) error {
	data, err := w.GetMemory(addr, 4)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(data, value)

	return nil
}

func (w *Instance) HandleError(err error) {
	w.Logger().Error(err, "wasm error")
}
