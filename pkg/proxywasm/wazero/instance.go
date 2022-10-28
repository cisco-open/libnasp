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
	"fmt"
	"sync"
	"sync/atomic"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
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
	ErrMalloc                      = errors.New("could not malloc")
)

type Instance struct {
	lock     sync.Mutex
	started  uint32
	refCount int
	stopCond *sync.Cond

	// user-defined data
	data interface{}

	r         wazero.Runtime
	module    api.Module
	moduleCtx context.Context
	ctx       context.Context

	hostModules map[string]wazero.HostModuleBuilder

	startFunctions []string

	logger logr.Logger
}

type InstanceOptions func(instance *Instance)

func StartFunctionsOption(names ...string) InstanceOptions {
	return func(instance *Instance) {
		instance.startFunctions = names
	}
}

func SetLoggerOption(logger logr.Logger) InstanceOptions {
	return func(instance *Instance) {
		instance.logger = logger
	}
}

func NewWazeroInstance(ctx context.Context, r wazero.Runtime, options ...InstanceOptions) *Instance {
	ins := &Instance{
		r:              r,
		lock:           sync.Mutex{},
		ctx:            ctx,
		hostModules:    make(map[string]wazero.HostModuleBuilder),
		startFunctions: []string{"_start", "_initialize"},
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

func (w *Instance) SetModule(ctx context.Context, module api.Module) {
	w.module = module
	w.moduleCtx = ctx
}

func (w *Instance) Init(f func(*Instance), ns wazero.Namespace) error {
	f(w)

	for _, hm := range w.hostModules {
		_, err := hm.Instantiate(w.ctx, ns)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Instance) Start() error {
	for _, fn := range w.startFunctions {
		f := w.module.ExportedFunction(fn)
		if f == nil {
			continue
		}

		if _, err := f.Call(context.Background()); err != nil {
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
		if err := w.module.Close(w.moduleCtx); err != nil {
			w.Logger().Error(err, "could not close module", "module", w.module.Name())
		}
		w.lock.Unlock()
	}()
}

// return true is Instance is started, false if not started.
func (w *Instance) checkStart() bool {
	return atomic.LoadUint32(&w.started) == 1
}

func (w *Instance) RegisterFunc(namespace string, funcName string, f interface{}) error {
	if w.checkStart() {
		return ErrInstanceAlreadyStart
	}

	if _, ok := w.hostModules[namespace]; !ok {
		w.hostModules[namespace] = w.r.NewHostModuleBuilder(namespace)
	}

	w.hostModules[namespace].NewFunctionBuilder().WithFunc(f).Export(funcName)

	return nil
}

func (w *Instance) Malloc(size int32) (uint64, error) {
	if !w.checkStart() {
		return 0, ErrInstanceNotStart
	}

	malloc, err := w.GetExportsFunc("malloc")
	if err != nil {
		return 0, err
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
	api.Function

	logger logr.Logger
}

func (c *Call) Call(args ...interface{}) (interface{}, error) {
	params := make([]uint64, 0)
	for _, ff := range args {
		switch v := ff.(type) {
		case uint64:
			params = append(params, v)
		case uint32:
			params = append(params, uint64(v))
		case int:
			params = append(params, uint64(v))
		case int64:
			params = append(params, uint64(v))
		case int32:
			params = append(params, uint64(v))
		default:
		}
	}

	if len(c.Function.Definition().ExportNames()) > 0 {
		c.logger.V(3).Info(fmt.Sprintf("call module function %s", c.Function.Definition().ExportNames()[0]))
	}

	ret, err := c.Function.Call(context.Background(), params...)
	if err != nil {
		return 0, err
	}

	if len(ret) > 0 {
		return int32(ret[0]), nil
	}

	return 0, nil
}

func (w *Instance) GetExportsFunc(funcName string) (common.WasmFunction, error) {
	if !w.checkStart() {
		return nil, ErrInstanceNotStart
	}

	f := &Call{
		Function: w.module.ExportedFunction(funcName),

		logger: w.logger,
	}

	return f, nil
}

func (w *Instance) GetExportsMem(memName string) ([]byte, error) {
	return nil, ErrGetExportsMemNotImplemented
}

func (w *Instance) GetMemory(addr uint64, size uint64) ([]byte, error) {
	mem, ok := w.module.Memory().Read(w.ctx, uint32(addr), uint32(size))
	if !ok {
		return nil, ErrGetData
	}

	return mem, nil
}

func (w *Instance) PutMemory(addr uint64, size uint64, content []byte) error {
	if n := len(content); n < int(size) {
		size = uint64(n)
	}

	ok := w.module.Memory().Write(w.ctx, uint32(addr), content[:size])
	if !ok {
		return ErrPutData
	}

	return nil
}

func (w *Instance) GetByte(addr uint64) (byte, error) {
	b, ok := w.module.Memory().ReadByte(w.ctx, uint32(addr))
	if !ok {
		return 0, ErrGetByte
	}

	return b, nil
}

func (w *Instance) PutByte(addr uint64, b byte) error {
	ok := w.module.Memory().WriteByte(w.ctx, uint32(addr), b)
	if !ok {
		return ErrPutByte
	}

	return nil
}

func (w *Instance) GetUint32(addr uint64) (uint32, error) {
	ret, ok := w.module.Memory().ReadUint32Le(w.ctx, uint32(addr))
	if !ok {
		return 0, ErrGetUint32
	}

	return ret, nil
}

func (w *Instance) PutUint32(addr uint64, value uint32) error {
	ok := w.module.Memory().WriteUint32Le(w.ctx, uint32(addr), value)
	if !ok {
		return ErrPutUint32
	}

	return nil
}

func (w *Instance) HandleError(err error) {
	w.Logger().Error(err, "wasm error")
}
