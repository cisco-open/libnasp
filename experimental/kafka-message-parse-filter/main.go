// Copyright 2020-2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/binary"
	"errors"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"strconv"
	"wwwin-github.cisco.com/eti/kafka-protocol-go/pkg/request"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
}

// Override types.DefaultPluginContext.
func (ctx *pluginContext) NewTcpContext(contextID uint32) types.TcpContext {
	return &networkContext{}
}

type networkContext struct {
	// Embed the default tcp context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultTcpContext
}

// Override types.DefaultTcpContext.
func (ctx *networkContext) OnNewConnection() types.Action {
	proxywasm.LogInfo("new connection!")
	return types.ActionContinue
}

// Override types.DefaultTcpContext.
func (ctx *networkContext) OnDownstreamData(dataSize int, endOfStream bool) types.Action {
	if dataSize == 0 || dataSize < 4 {
		return types.ActionContinue
	}

	kafkaMessageLengthRaw, err := proxywasm.GetDownstreamData(0, 4)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to get downstream data: %v", err)
		return types.ActionContinue
	}

	kafkaMessageLength := binary.BigEndian.Uint32(kafkaMessageLengthRaw)

	proxywasm.LogInfof(">>>>>> downstream data received, kafka message length >>>>>>\n%s", strconv.Itoa(int(kafkaMessageLength)))

	if dataSize < int(kafkaMessageLength) {
		return types.ActionPause
	}

	kafkaMessage, err := proxywasm.GetDownstreamData(4, int(kafkaMessageLength))
	if err != nil && !errors.Is(err, types.ErrorStatusNotFound) {
		proxywasm.LogCriticalf("failed to get downstream data: %v", err)
		return types.ActionContinue
	}

	proxywasm.LogInfof(">>>>>> downstream data received, kafka message >>>>>>\n%s", string(kafkaMessage))

	kafkaRequest, err := request.Parse(kafkaMessage)

	proxywasm.LogInfof(">>>>>> kafka message parsed, kafka message contents >>>>>>\n%s", kafkaRequest.String())
	proxywasm.LogInfof(">>>>>> kafka message parsed error, error >>>>>>\n%s", err.Error())

	kafkaRequest.Release()

	return types.ActionContinue
}

// Override types.DefaultTcpContext.
func (ctx *networkContext) OnDownstreamClose(types.PeerType) {
	proxywasm.LogInfo("downstream connection close!")
	return
}

// Override types.DefaultTcpContext.
func (ctx *networkContext) OnUpstreamData(dataSize int, endOfStream bool) types.Action {
	return types.ActionContinue
}

// Override types.DefaultTcpContext.
func (ctx *networkContext) OnStreamDone() {
	proxywasm.LogInfo("connection complete!")
}
