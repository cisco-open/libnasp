//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

package main

import (
	"crypto/md5"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
)

// main - test WASM filter which waits for multiple data chunks to accumulate in the host before processing the data
// stream and returning with `Continue` from `OnDownstreamData` or `OnUpstreamData`. It waits until at least `reqDataSize`
// specified in the configuration is received by the host; it modifies the received data stream by appending the number
// of chunks seen before reaching at least `reqDataSize` and md5 checksum of the original data
// it
func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{contextID: contextID}
}

type pluginContext struct {
	types.DefaultPluginContext

	contextID uint32

	reqDataSize uint64
}

func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	pluginConfig, err := proxywasm.GetPluginConfiguration()
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	if !gjson.ValidBytes(pluginConfig) {
		proxywasm.LogCriticalf("the plugin configuration is not a valid json: %q", string(pluginConfig))
		return types.OnPluginStartStatusFailed
	}

	ctx.reqDataSize = gjson.ParseBytes(pluginConfig).Get("req_data_size").Uint()

	return types.OnPluginStartStatusOK
}

func (ctx *pluginContext) NewTcpContext(contextID uint32) types.TcpContext {
	return &networkContext{contextID: contextID, reqDataSize: ctx.reqDataSize}
}

type networkContext struct {
	types.DefaultTcpContext

	contextID uint32

	reqDataSize uint64

	counter int
}

func (ctx *networkContext) OnNewConnection() types.Action {
	proxywasm.LogInfo("new connection!")

	ctx.counter = 0

	return types.ActionContinue
}

func (ctx *networkContext) OnDownstreamData(dataSize int, endOfStream bool) types.Action {
	if endOfStream {
		return types.ActionContinue
	}

	ctx.counter++
	if uint64(dataSize) < ctx.reqDataSize {
		return types.ActionPause
	}

	data, err := proxywasm.GetDownstreamData(0, int(ctx.reqDataSize))
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to get downstream data: %v", err)
		return types.ActionContinue
	}

	proxywasm.LogInfof(">>>>>> downstream data received >>>>>>\n%s", string(data))

	checkSum := md5.Sum(data)
	var b []byte
	b = append(b, byte(ctx.counter))
	b = append(b, checkSum[:]...)
	err = proxywasm.AppendDownstreamData(b)
	if err != nil {
		proxywasm.LogCritical("couldn't append the MD5 checksum of received data")
		ctx.counter = 0
		return types.ActionContinue
	}

	ctx.counter = 0
	return types.ActionContinue
}
