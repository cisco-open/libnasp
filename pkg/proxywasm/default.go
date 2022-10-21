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
	"time"

	"mosn.io/proxy-wasm-go-host/proxywasm/common"
	v1 "mosn.io/proxy-wasm-go-host/proxywasm/v1"
)

type DefaultHostFunctions struct {
	*v1.DefaultImportsHandler
}

func NewDefaultHostFunctions() v1.ImportsHandler {
	return &DefaultHostFunctions{
		DefaultImportsHandler: &v1.DefaultImportsHandler{},
	}
}

func (d *DefaultHostFunctions) Wait() v1.Action {
	return v1.ActionContinue
}

// utils
func (d *DefaultHostFunctions) GetRootContextID() int32 {
	panic("GetRootContextID not implemented")
}

func (d *DefaultHostFunctions) GetVmConfig() common.IoBuffer {
	panic("GetVmConfig not implemented")
}

func (d *DefaultHostFunctions) GetPluginConfig() common.IoBuffer {
	panic("GetPluginConfig not implemented")
}

func (d *DefaultHostFunctions) Log(level v1.LogLevel, msg string) v1.WasmResult {
	panic("Log not implemented")
}

func (d *DefaultHostFunctions) SetEffectiveContextID(contextID int32) v1.WasmResult {
	panic("SetEffectiveContextID not implemented")
}

func (d *DefaultHostFunctions) SetTickPeriodMilliseconds(tickPeriodMilliseconds int32) v1.WasmResult {
	panic("SetTickPeriodMilliseconds not implemented")
}

func (d *DefaultHostFunctions) GetCurrentTimeNanoseconds() (int32, v1.WasmResult) {
	nano := time.Now().Nanosecond()

	return int32(nano), v1.WasmResultOk
}

func (d *DefaultHostFunctions) Done() v1.WasmResult {
	panic("Done not implemented")
}

// l4

func (d *DefaultHostFunctions) GetDownStreamData() common.IoBuffer {
	panic("GetDownStreamData not implemented")
}

func (d *DefaultHostFunctions) GetUpstreamData() common.IoBuffer {
	panic("GetUpstreamData not implemented")
}

func (d *DefaultHostFunctions) ResumeDownstream() v1.WasmResult {
	panic("ResumeDownstream not implemented")
}

func (d *DefaultHostFunctions) ResumeUpstream() v1.WasmResult {
	panic("ResumeUpstream not implemented")
}

// http

func (d *DefaultHostFunctions) GetHttpRequestHeader() common.HeaderMap {
	panic("GetHttpRequestHeader not implemented")
}

func (d *DefaultHostFunctions) GetHttpRequestBody() common.IoBuffer {
	panic("GetHttpRequestBody not implemented")
}

func (d *DefaultHostFunctions) GetHttpRequestTrailer() common.HeaderMap {
	panic("GetHttpRequestTrailer not implemented")
}

func (d *DefaultHostFunctions) GetHttpResponseHeader() common.HeaderMap {
	panic("GetHttpResponseHeader not implemented")
}

func (d *DefaultHostFunctions) GetHttpResponseBody() common.IoBuffer {
	panic("GetHttpResponseBody not implemented")
}

func (d *DefaultHostFunctions) GetHttpResponseTrailer() common.HeaderMap {
	panic("GetHttpResponseTrailer not implemented")
}

func (d *DefaultHostFunctions) GetHttpCallResponseHeaders() common.HeaderMap {
	panic("GetHttpCallResponseHeaders not implemented")
}

func (d *DefaultHostFunctions) GetHttpCallResponseBody() common.IoBuffer {
	panic("GetHttpCallResponseBody not implemented")
}

func (d *DefaultHostFunctions) GetHttpCallResponseTrailer() common.HeaderMap {
	panic("GetHttpCallResponseTrailer not implemented")
}

func (d *DefaultHostFunctions) HttpCall(url string, headers common.HeaderMap, body common.IoBuffer, trailer common.HeaderMap, timeoutMilliseconds int32) (int32, v1.WasmResult) {
	panic("HttpCall not implemented")
}

func (d *DefaultHostFunctions) ResumeHttpRequest() v1.WasmResult {
	panic("ResumeHttpRequest not implemented")
}

func (d *DefaultHostFunctions) ResumeHttpResponse() v1.WasmResult {
	panic("ResumeHttpResponse not implemented")
}

func (d *DefaultHostFunctions) SendHttpResp(respCode int32, respCodeDetail common.IoBuffer, respBody common.IoBuffer, additionalHeaderMap common.HeaderMap, grpcCode int32) v1.WasmResult {
	panic("SendHttpResp not implemented")
}

// grpc

func (d *DefaultHostFunctions) OpenGrpcStream(grpcService string, serviceName string, method string) (int32, v1.WasmResult) {
	panic("OpenGrpcStream not implemented")
}

func (d *DefaultHostFunctions) SendGrpcCallMsg(token int32, data common.IoBuffer, endOfStream int32) v1.WasmResult {
	panic("SendGrpcCallMsg not implemented")
}

func (d *DefaultHostFunctions) CancelGrpcCall(token int32) v1.WasmResult {
	panic("CancelGrpcCall not implemented")
}

func (d *DefaultHostFunctions) CloseGrpcCall(token int32) v1.WasmResult {
	panic("CloseGrpcCall not implemented")
}

func (d *DefaultHostFunctions) GrpcCall(grpcService string, serviceName string, method string, data common.IoBuffer, timeoutMilliseconds int32) (int32, v1.WasmResult) {
	panic("GrpcCall not implemented")
}

func (d *DefaultHostFunctions) GetGrpcReceiveInitialMetaData() common.HeaderMap {
	panic("GetGrpcReceiveInitialMetaData not implemented")
}

func (d *DefaultHostFunctions) GetGrpcReceiveBuffer() common.IoBuffer {
	panic("GetGrpcReceiveBuffer not implemented")
}

func (d *DefaultHostFunctions) GetGrpcReceiveTrailerMetaData() common.HeaderMap {
	panic("GetGrpcReceiveTrailerMetaData not implemented")
}

// foreign

func (d *DefaultHostFunctions) CallForeignFunction(funcName string, param []byte) ([]byte, v1.WasmResult) {
	panic("CallForeignFunction not implemented")
}

func (d *DefaultHostFunctions) GetFuncCallData() common.IoBuffer {
	panic("GetFuncCallData not implemented")
}

// property

func (d *DefaultHostFunctions) GetProperty(key string) (string, v1.WasmResult) {
	panic("GetProperty not implemented")
}

func (d *DefaultHostFunctions) SetProperty(key string, value string) v1.WasmResult {
	panic("SetProperty not implemented")
}

// metric

func (d *DefaultHostFunctions) DefineMetric(metricType v1.MetricType, name string) (int32, v1.WasmResult) {
	panic("DefineMetric not implemented")
}

func (d *DefaultHostFunctions) IncrementMetric(metricID int32, offset int64) v1.WasmResult {
	panic("IncrementMetric not implemented")
}

func (d *DefaultHostFunctions) RecordMetric(metricID int32, value int64) v1.WasmResult {
	panic("RecordMetric not implemented")
}

func (d *DefaultHostFunctions) GetMetric(metricID int32) (int64, v1.WasmResult) {
	panic("GetMetric not implemented")
}

func (d *DefaultHostFunctions) RemoveMetric(metricID int32) v1.WasmResult {
	panic("RemoveMetric not implemented")
}

// custom extension

func (d *DefaultHostFunctions) GetCustomBuffer(bufferType v1.BufferType) common.IoBuffer {
	panic("GetCustomBuffer not implemented")
}

func (d *DefaultHostFunctions) GetCustomHeader(mapType v1.MapType) common.HeaderMap {
	panic("GetCustomHeader not implemented")
}
