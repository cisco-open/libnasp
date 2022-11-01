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

package main

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

const tickMilliseconds uint32 = 1000

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &rootContext{
		lock:         &sync.Mutex{},
		httpContexts: map[uint32]*httpContext{},
		contextID:    contextID,
	}
}

type rootContext struct {
	types.DefaultPluginContext

	contextID    uint32
	httpContexts map[uint32]*httpContext
	lock         *sync.Mutex
}

func (ctx *rootContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	rand.Seed(time.Now().UnixNano())

	if err := proxywasm.SetTickPeriodMilliSeconds(tickMilliseconds); err != nil {
		proxywasm.LogCriticalf("failed to set tick period: %v", err)
	}

	return types.OnPluginStartStatusOK
}

func (ctx *rootContext) OnTick() {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	if len(ctx.httpContexts) == 0 {
		proxywasm.LogDebugf("no active http context found")

		return
	}

	proxywasm.LogDebugf("there are %d active http context", len(ctx.httpContexts))

	ids := []string{}
	for _, hctx := range ctx.httpContexts {
		hctx := hctx
		proxywasm.LogDebugf("Set effective context to: %d", hctx.ContextID)
		if err := proxywasm.SetEffectiveContext(hctx.ContextID); err != nil {
			proxywasm.LogError(err.Error())
		}

		proxywasm.LogDebugf("[cid: %d] Get test id", hctx.ContextID)
		str, err := proxywasm.GetProperty([]string{"test-id"})
		if err == nil {
			proxywasm.LogDebugf("[cid: %d] test id is '%s'", hctx.ContextID, string(str))
			ids = append(ids, string(str))
		} else {
			proxywasm.LogError(err.Error())
		}
	}

	if err := proxywasm.SetEffectiveContext(ctx.contextID); err != nil {
		proxywasm.LogError(err.Error())
	}
	if err := proxywasm.SetProperty([]string{"test-ids"}, []byte(strings.Join(ids, ","))); err != nil {
		proxywasm.LogError(err.Error())
	}
}

func (ctx *rootContext) AddContext(hctx *httpContext) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	ctx.httpContexts[hctx.ContextID] = hctx
}

func (ctx *rootContext) RemoveContext(hctx *httpContext) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	delete(ctx.httpContexts, hctx.ContextID)
}

type httpContext struct {
	types.DefaultHttpContext
	ContextID uint32

	rootContext *rootContext
}

func (rootContext *rootContext) NewHttpContext(id uint32) types.HttpContext {
	ctx := &httpContext{
		ContextID: id,

		rootContext: rootContext,
	}

	rootContext.AddContext(ctx)

	return ctx
}

func (ctx *httpContext) OnHttpStreamDone() {
	ctx.rootContext.RemoveContext(ctx)
}
