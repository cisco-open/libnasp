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
	v1 "mosn.io/proxy-wasm-go-host/proxywasm/v1"
)

type ImportsV1 struct {
	instance *Instance
}

func RegisterV1Imports(instance *Instance) {
	imports := &ImportsV1{
		instance: instance,
	}

	_ = instance.RegisterFunc("env", "proxy_log", imports.ProxyLog)

	_ = instance.RegisterFunc("env", "proxy_get_property", imports.ProxyGetProperty)
	_ = instance.RegisterFunc("env", "proxy_set_property", imports.ProxySetProperty)

	_ = instance.RegisterFunc("env", "proxy_get_header_map_value", imports.ProxyGetHeaderMapValue)
	_ = instance.RegisterFunc("env", "proxy_remove_header_map_value", imports.ProxyRemoveHeaderMapValue)
	_ = instance.RegisterFunc("env", "proxy_replace_header_map_value", imports.ProxyReplaceHeaderMapValue)
	_ = instance.RegisterFunc("env", "proxy_add_header_map_value", imports.ProxyAddHeaderMapValue)

	_ = instance.RegisterFunc("env", "proxy_call_foreign_function", imports.ProxyCallForeignFunction)

	_ = instance.RegisterFunc("env", "proxy_get_buffer_bytes", imports.ProxyGetBufferBytes)
	_ = instance.RegisterFunc("env", "proxy_set_buffer_bytes", imports.ProxySetBufferBytes)

	_ = instance.RegisterFunc("env", "proxy_define_metric", imports.ProxyDefineMetric)
	_ = instance.RegisterFunc("env", "proxy_increment_metric", imports.ProxyIncrementMetric)
	_ = instance.RegisterFunc("env", "proxy_record_metric", imports.ProxyRecordMetric)

	_ = instance.RegisterFunc("env", "proxy_set_tick_period_milliseconds", imports.ProxySetTickPeriodMilliseconds)

	_ = instance.RegisterFunc("env", "proxy_set_effective_context", imports.ProxySetEffectiveContext)

	_ = instance.RegisterFunc("env", "proxy_done", imports.ProxyDone)

	_ = instance.RegisterFunc("env", "proxy_send_local_response", imports.ProxySendHttpResponse)
	_ = instance.RegisterFunc("env", "proxy_get_header_map_pairs", imports.ProxyGetHeaderMapPairs)

	_ = instance.RegisterFunc("env", "proxy_http_call", imports.ProxyHttpCall)

	_ = instance.RegisterFunc("env", "proxy_remove_metric", imports.ProxyRemoveMetric)
	_ = instance.RegisterFunc("env", "proxy_get_metric", imports.ProxyGetMetric)
	_ = instance.RegisterFunc("env", "proxy_set_header_map_pairs", imports.ProxySetHeaderMapPairs)
	_ = instance.RegisterFunc("env", "proxy_get_current_time_nanoseconds", imports.ProxyGetCurrentTimeNanoseconds)
	_ = instance.RegisterFunc("env", "proxy_resume_downstream", imports.ProxyResumeDownstream)
	_ = instance.RegisterFunc("env", "proxy_resume_upstream", imports.ProxyResumeUpstream)
	_ = instance.RegisterFunc("env", "proxy_continue_stream", imports.ProxyContinueStream)
	_ = instance.RegisterFunc("env", "proxy_close_stream", imports.ProxyCloseStream)

	_ = instance.RegisterFunc("env", "proxy_open_grpc_stream", imports.ProxyOpenGrpcStream)
	_ = instance.RegisterFunc("env", "proxy_send_grpc_call_message", imports.ProxySendGrpcCallMessage)
	_ = instance.RegisterFunc("env", "proxy_cancel_grpc_call", imports.ProxyCancelGrpcCall)
	_ = instance.RegisterFunc("env", "proxy_close_grpc_call", imports.ProxyCloseGrpcCall)
	_ = instance.RegisterFunc("env", "proxy_grpc_call", imports.ProxyGrpcCall)
	_ = instance.RegisterFunc("env", "proxy_dispatch_grpc_call", imports.ProxyGrpcCall)

	_ = instance.RegisterFunc("env", "proxy_grpc_stream", imports.ProxyOpenGrpcStream)
	_ = instance.RegisterFunc("env", "proxy_grpc_send", imports.ProxySendGrpcCallMessage)
	_ = instance.RegisterFunc("env", "proxy_grpc_cancel", imports.ProxyCancelGrpcCall)
	_ = instance.RegisterFunc("env", "proxy_grpc_close", imports.ProxyCloseGrpcCall)

	_ = instance.RegisterFunc("env", "proxy_get_status", imports.ProxyGetStatus)

	_ = instance.RegisterFunc("env", "proxy_resume_http_request", imports.ProxyResumeHttpRequest)
	_ = instance.RegisterFunc("env", "proxy_resume_http_response", imports.ProxyResumeHttpResponse)
	_ = instance.RegisterFunc("env", "proxy_send_http_response", imports.ProxySendHttpResponse)
	_ = instance.RegisterFunc("env", "proxy_dispatch_http_call", imports.ProxyHttpCall)
	_ = instance.RegisterFunc("env", "proxy_register_shared_queue", imports.ProxyRegisterSharedQueue)
	_ = instance.RegisterFunc("env", "proxy_remove_shared_queue", imports.ProxyRemoveSharedQueue)
	_ = instance.RegisterFunc("env", "proxy_resolve_shared_queue", imports.ProxyResolveSharedQueue)
	_ = instance.RegisterFunc("env", "proxy_dequeue_shared_queue", imports.ProxyDequeueSharedQueue)
	_ = instance.RegisterFunc("env", "proxy_enqueue_shared_queue", imports.ProxyEnqueueSharedQueue)
	_ = instance.RegisterFunc("env", "proxy_get_shared_data", imports.ProxyGetSharedData)
	_ = instance.RegisterFunc("env", "proxy_set_shared_data", imports.ProxySetSharedData)
}

func (i *ImportsV1) ProxyLog(logLevel int32, messageData int32, messageSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_log")

	return v1.ProxyLog(i.instance, logLevel, messageData, messageSize)
}

func (i *ImportsV1) ProxySetEffectiveContext(contextID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_set_effective_context")

	return v1.ProxySetEffectiveContext(i.instance, contextID)
}

func (i *ImportsV1) ProxyGetProperty(keyPtr int32, keySize int32, returnValueData int32, returnValueSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_property")

	return v1.ProxyGetProperty(i.instance, keyPtr, keySize, returnValueData, returnValueSize)
}

func (i *ImportsV1) ProxySetProperty(keyPtr int32, keySize int32, valuePtr int32, valueSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_set_property")

	return v1.ProxySetProperty(i.instance, keyPtr, keySize, valuePtr, valueSize)
}

func (i *ImportsV1) ProxyGetBufferBytes(bufferType int32, start int32, length int32, returnBufferData int32, returnBufferSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_buffer_bytes", "type", bufferType)

	return v1.ProxyGetBufferBytes(i.instance, bufferType, start, length, returnBufferData, returnBufferSize)
}

func (i *ImportsV1) ProxySetBufferBytes(bufferType int32, start int32, length int32, dataPtr int32, dataSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_set_buffer_bytes")

	return v1.ProxySetBufferBytes(i.instance, bufferType, start, length, dataPtr, dataSize)
}

func (i *ImportsV1) ProxyGetHeaderMapPairs(mapType int32, returnDataPtr int32, returnDataSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_header_map_pairs")

	return v1.ProxyGetHeaderMapPairs(i.instance, mapType, returnDataPtr, returnDataSize)
}

func (i *ImportsV1) ProxySetHeaderMapPairs(mapType int32, ptr int32, size int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_set_header_map_pairs")

	return v1.ProxySetHeaderMapPairs(i.instance, mapType, ptr, size)
}

func (i *ImportsV1) ProxyGetHeaderMapValue(mapType int32, keyDataPtr int32, keySize int32, valueDataPtr int32, valueSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_header_map_value", "type", mapType)

	return v1.ProxyGetHeaderMapValue(i.instance, mapType, keyDataPtr, keySize, valueDataPtr, valueSize)
}

func (i *ImportsV1) ProxyReplaceHeaderMapValue(mapType int32, keyDataPtr int32, keySize int32, valueDataPtr int32, valueSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_replace_header_map_value")

	return v1.ProxyReplaceHeaderMapValue(i.instance, mapType, keyDataPtr, keySize, valueDataPtr, valueSize)
}

func (i *ImportsV1) ProxyAddHeaderMapValue(mapType int32, keyDataPtr int32, keySize int32, valueDataPtr int32, valueSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_add_header_map_value")

	return v1.ProxyAddHeaderMapValue(i.instance, mapType, keyDataPtr, keySize, valueDataPtr, valueSize)
}

func (i *ImportsV1) ProxyRemoveHeaderMapValue(mapType int32, keyDataPtr int32, keySize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_remove_header_map_value")

	return v1.ProxyRemoveHeaderMapValue(i.instance, mapType, keyDataPtr, keySize)
}

func (i *ImportsV1) ProxySetTickPeriodMilliseconds(tickPeriodMilliseconds int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_set_tick_period_milliseconds")

	return v1.ProxySetTickPeriodMilliseconds(i.instance, tickPeriodMilliseconds)
}

func (i *ImportsV1) ProxyGetCurrentTimeNanoseconds(resultUint64Ptr int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_current_time_nanoseconds")

	return v1.ProxyGetCurrentTimeNanoseconds(i.instance, resultUint64Ptr)
}

func (i *ImportsV1) ProxyResumeDownstream() int32 {
	i.instance.Logger().V(4).Info("call host function proxy_resume_downstream")

	return v1.ProxyResumeDownstream(i.instance)
}

func (i *ImportsV1) ProxyResumeUpstream() int32 {
	i.instance.Logger().V(4).Info("call host function proxy_resume_upstream")

	return v1.ProxyResumeUpstream(i.instance)
}

func (i *ImportsV1) ProxyContinueStream(streamType int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_continue_stream")

	if streamType == 1 {
		return v1.ProxyResumeDownstream(i.instance)
	} else {
		return v1.ProxyResumeUpstream(i.instance)
	}
}

func (i *ImportsV1) ProxyCloseStream(streamType int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_close_stream")

	return int32(v1.WasmResultUnimplemented)
}

func (i *ImportsV1) ProxyGetStatus(statusCode int32, messagePtr int32, messageSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_status")

	return int32(v1.WasmResultUnimplemented)
}

func (i *ImportsV1) ProxyOpenGrpcStream(grpcServiceData int32, grpcServiceSize int32,
	serviceNameData int32, serviceNameSize int32,
	methodData int32, methodSize int32,
	initialMetadata int32, initialMetadataSize int32,
	returnCalloutID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_open_grpc_stream")

	return v1.ProxyOpenGrpcStream(i.instance, grpcServiceData, grpcServiceSize, serviceNameData, serviceNameSize, methodData, methodSize, returnCalloutID)
}

func (i *ImportsV1) ProxySendGrpcCallMessage(calloutID int32, data int32, size int32, endOfStream int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_send_grpc_call_message")

	return v1.ProxySendGrpcCallMessage(i.instance, calloutID, data, size, endOfStream)
}

func (i *ImportsV1) ProxyCancelGrpcCall(calloutID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_cancel_grpc_call")

	return v1.ProxyCancelGrpcCall(i.instance, calloutID)
}

func (i *ImportsV1) ProxyCloseGrpcCall(calloutID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_close_grpc_call")

	return v1.ProxyCloseGrpcCall(i.instance, calloutID)
}

func (i *ImportsV1) ProxyGrpcCall(grpcServiceData int32, grpcServiceSize int32,
	serviceNameData int32, serviceNameSize int32,
	methodData int32, methodSize int32,
	initialMetadata int32, initialMetadataSize int32,
	grpcMessageData int32, grpcMessageSize int32,
	timeoutMilliseconds int32, returnCalloutID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_grpc_call")

	return v1.ProxyGrpcCall(i.instance, grpcServiceData, grpcServiceSize, serviceNameData, serviceNameSize, methodData, methodSize, grpcMessageData, grpcMessageSize, timeoutMilliseconds, returnCalloutID)
}

func (i *ImportsV1) ProxyResumeHttpRequest() int32 {
	i.instance.Logger().V(4).Info("call host function proxy_resume_http_request")

	return v1.ProxyResumeHttpRequest(i.instance)
}

func (i *ImportsV1) ProxyResumeHttpResponse() int32 {
	i.instance.Logger().V(4).Info("call host function proxy_resume_http_response")

	return v1.ProxyResumeHttpResponse(i.instance)
}

func (i *ImportsV1) ProxySendHttpResponse(respCode int32, respCodeDetailPtr int32, respCodeDetailSize int32,
	respBodyPtr int32, respBodySize int32, additionalHeaderMapDataPtr int32, additionalHeaderSize int32, grpcStatus int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_send_http_response")

	return v1.ProxySendHttpResponse(i.instance, respCode, respCodeDetailPtr, respCodeDetailSize, respBodyPtr, respBodySize, additionalHeaderMapDataPtr, additionalHeaderSize, grpcStatus)
}

func (i *ImportsV1) ProxyHttpCall(uriPtr int32, uriSize int32,
	headerPairsPtr int32, headerPairsSize int32, bodyPtr int32, bodySize int32, trailerPairsPtr int32, trailerPairsSize int32,
	timeoutMilliseconds int32, calloutIDPtr int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_http_call")

	return v1.ProxyHttpCall(i.instance, uriPtr, uriSize, headerPairsPtr, headerPairsSize, bodyPtr, bodySize, trailerPairsPtr, trailerPairsSize, timeoutMilliseconds, calloutIDPtr)
}

func (i *ImportsV1) ProxyDefineMetric(metricType int32, namePtr int32, nameSize int32, returnMetricId int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_define_metric")

	return v1.ProxyDefineMetric(i.instance, metricType, namePtr, nameSize, returnMetricId)
}

func (i *ImportsV1) ProxyRemoveMetric(metricID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_remove_metric")

	return v1.ProxyRemoveMetric(i.instance, metricID)
}

func (i *ImportsV1) ProxyIncrementMetric(metricId int32, offset int64) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_increment_metric")

	return v1.ProxyIncrementMetric(i.instance, metricId, offset)
}

func (i *ImportsV1) ProxyRecordMetric(metricId int32, value int64) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_record_metric")

	return v1.ProxyRecordMetric(i.instance, metricId, value)
}

func (i *ImportsV1) ProxyGetMetric(metricId int32, resultUint64Ptr int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_metric")

	return v1.ProxyGetMetric(i.instance, metricId, resultUint64Ptr)
}

func (i *ImportsV1) ProxyRegisterSharedQueue(queueNamePtr int32, queueNameSize int32, tokenPtr int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_register_shared_queue")

	return v1.ProxyRegisterSharedQueue(i.instance, queueNamePtr, queueNameSize, tokenPtr)
}

func (i *ImportsV1) ProxyRemoveSharedQueue(queueID int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_remove_shared_queue")

	return v1.ProxyRemoveSharedQueue(i.instance, queueID)
}

func (i *ImportsV1) ProxyResolveSharedQueue(vmIDPtr int32, vmIDSize int32, queueNamePtr int32, queueNameSize int32, tokenPtr int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_resolve_shared_queue")

	return v1.ProxyResolveSharedQueue(i.instance, queueNamePtr, queueNameSize, tokenPtr)
}

func (i *ImportsV1) ProxyDequeueSharedQueue(token int32, dataPtr int32, dataSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_dequeue_shared_queue")

	return v1.ProxyDequeueSharedQueue(i.instance, token, dataPtr, dataSize)
}

func (i *ImportsV1) ProxyEnqueueSharedQueue(token int32, dataPtr int32, dataSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_enqueue_shared_queue")

	return v1.ProxyEnqueueSharedQueue(i.instance, token, dataPtr, dataSize)
}

func (i *ImportsV1) ProxyGetSharedData(keyPtr int32, keySize int32, valuePtr int32, valueSizePtr int32, casPtr int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_get_shared_data")

	return v1.ProxyGetSharedData(i.instance, keyPtr, keySize, valuePtr, valueSizePtr, casPtr)
}

func (i *ImportsV1) ProxySetSharedData(keyPtr int32, keySize int32, valuePtr int32, valueSize int32, cas int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_set_shared_data")

	return v1.ProxySetSharedData(i.instance, keyPtr, keySize, valuePtr, valueSize, cas)
}

func (i *ImportsV1) ProxyDone() int32 {
	i.instance.Logger().V(4).Info("call host function proxy_done")

	return v1.ProxyDone(i.instance)
}

func (i *ImportsV1) ProxyCallForeignFunction(funcNamePtr int32, funcNameSize int32,
	paramPtr int32, paramSize int32, returnData int32, returnSize int32) int32 {
	i.instance.Logger().V(4).Info("call host function proxy_call_foreign_function")

	return v1.ProxyCallForeignFunction(i.instance, funcNamePtr, funcNameSize, paramPtr, paramSize, returnData, returnSize)
}
