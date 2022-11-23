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
	v2 "mosn.io/proxy-wasm-go-host/proxywasm/v2"
)

type ImportsV2 struct {
	instance *Instance
}

func RegisterV2Imports(instance *Instance) {
	imports := &ImportsV2{
		instance: instance,
	}

	_ = instance.RegisterFunc("env", "proxy_log", imports.ProxyLog)

	_ = instance.RegisterFunc("env", "proxy_set_effective_context", imports.ProxySetEffectiveContext)
	_ = instance.RegisterFunc("env", "proxy_context_finalize", imports.ProxyContextFinalize)

	_ = instance.RegisterFunc("env", "proxy_resume_stream", imports.ProxyResumeStream)
	_ = instance.RegisterFunc("env", "proxy_close_stream", imports.ProxyCloseStream)

	_ = instance.RegisterFunc("env", "proxy_send_http_response", imports.ProxySendHttpResponse)
	_ = instance.RegisterFunc("env", "proxy_resume_http_stream", imports.ProxyResumeHttpStream)
	_ = instance.RegisterFunc("env", "proxy_close_http_stream", imports.ProxyCloseHttpStream)

	_ = instance.RegisterFunc("env", "proxy_get_buffer", imports.ProxyGetBuffer)
	_ = instance.RegisterFunc("env", "proxy_set_buffer", imports.ProxySetBuffer)

	_ = instance.RegisterFunc("env", "proxy_get_map_values", imports.ProxyGetMapValues)
	_ = instance.RegisterFunc("env", "proxy_set_map_values", imports.ProxySetMapValues)

	_ = instance.RegisterFunc("env", "proxy_open_shared_kvstore", imports.ProxyOpenSharedKvstore)
	_ = instance.RegisterFunc("env", "proxy_get_shared_kvstore_key_values", imports.ProxyGetSharedKvstoreKeyValues)
	_ = instance.RegisterFunc("env", "proxy_set_shared_kvstore_key_values", imports.ProxySetSharedKvstoreKeyValues)
	_ = instance.RegisterFunc("env", "proxy_add_shared_kvstore_key_values", imports.ProxyAddSharedKvstoreKeyValues)
	_ = instance.RegisterFunc("env", "proxy_remove_shared_kvstore_key", imports.ProxyRemoveSharedKvstoreKey)
	_ = instance.RegisterFunc("env", "proxy_delete_shared_kvstore", imports.ProxyDeleteSharedKvstore)

	_ = instance.RegisterFunc("env", "proxy_open_shared_queue", imports.ProxyOpenSharedQueue)
	_ = instance.RegisterFunc("env", "proxy_dequeue_shared_queue_item", imports.ProxyDequeueSharedQueueItem)
	_ = instance.RegisterFunc("env", "proxy_enqueue_shared_queue_item", imports.ProxyEnqueueSharedQueueItem)
	_ = instance.RegisterFunc("env", "proxy_delete_shared_queue", imports.ProxyDeleteSharedQueue)

	_ = instance.RegisterFunc("env", "proxy_create_timer", imports.ProxyCreateTimer)
	_ = instance.RegisterFunc("env", "proxy_delete_timer", imports.ProxyDeleteTimer)

	_ = instance.RegisterFunc("env", "proxy_create_metric", imports.ProxyCreateMetric)
	_ = instance.RegisterFunc("env", "proxy_get_metric_value", imports.ProxyGetMetricValue)
	_ = instance.RegisterFunc("env", "proxy_set_metric_value", imports.ProxySetMetricValue)
	_ = instance.RegisterFunc("env", "proxy_increment_metric_value", imports.ProxyIncrementMetricValue)
	_ = instance.RegisterFunc("env", "proxy_delete_metric", imports.ProxyDeleteMetric)

	_ = instance.RegisterFunc("env", "proxy_dispatch_http_call", imports.ProxyDispatchHttpCall)

	_ = instance.RegisterFunc("env", "proxy_dispatch_grpc_call", imports.ProxyDispatchGrpcCall)
	_ = instance.RegisterFunc("env", "proxy_open_grpc_stream", imports.ProxyOpenGrpcStream)
	_ = instance.RegisterFunc("env", "proxy_send_grpc_stream_message", imports.ProxySendGrpcStreamMessage)
	_ = instance.RegisterFunc("env", "proxy_cancel_grpc_call", imports.ProxyCancelGrpcCall)
	_ = instance.RegisterFunc("env", "proxy_close_grpc_call", imports.ProxyCloseGrpcCall)

	_ = instance.RegisterFunc("env", "proxy_call_custom_function", imports.ProxyCallCustomFunction)
}

func (i *ImportsV2) ProxyLog(logLevel v2.LogLevel, messageData int32, messageSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_log")

	return v2.ProxyLog(i.instance, int32(logLevel), messageData, messageSize)
}

func (i *ImportsV2) ProxySetEffectiveContext(contextID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_set_effective_context")

	return v2.ProxySetEffectiveContext(i.instance, contextID)
}

func (i *ImportsV2) ProxyContextFinalize() v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_context_finalize")

	return v2.ProxyContextFinalize(i.instance)
}

func (i *ImportsV2) ProxyResumeStream(streamType v2.StreamType) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_resume_stream")

	return v2.ProxyResumeStream(i.instance, streamType)
}

func (i *ImportsV2) ProxyCloseStream(streamType v2.StreamType) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_close_stream")

	return v2.ProxyCloseStream(i.instance, streamType)
}

func (i *ImportsV2) ProxySendHttpResponse(responseCode int32, responseCodeDetailsData int32, responseCodeDetailsSize int32,
	responseBodyData int32, responseBodySize int32, additionalHeadersMapData int32, additionalHeadersSize int32,
	grpcStatus int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_send_http_response")

	return v2.ProxySendHttpResponse(i.instance, responseCode, responseCodeDetailsData, responseCodeDetailsSize, responseBodyData, responseBodySize, additionalHeadersMapData, additionalHeadersSize, grpcStatus)
}

func (i *ImportsV2) ProxyResumeHttpStream(streamType v2.StreamType) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_resume_http_stream")

	return v2.ProxyResumeHttpStream(i.instance, streamType)
}

func (i *ImportsV2) ProxyCloseHttpStream(streamType v2.StreamType) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_close_http_stream")

	return v2.ProxyCloseHttpStream(i.instance, streamType)
}

func (i *ImportsV2) ProxyGetBuffer(bufferType v2.BufferType, offset int32, maxSize int32,
	returnBufferData int32, returnBufferSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_get_buffer")

	return v2.ProxyGetBuffer(i.instance, int32(bufferType), offset, maxSize, returnBufferData, returnBufferSize)
}

func (i *ImportsV2) ProxySetBuffer(bufferType v2.BufferType, offset int32, size int32,
	bufferData int32, bufferSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_set_buffer")

	return v2.ProxySetBuffer(i.instance, bufferType, offset, size, bufferData, bufferSize)
}

func (i *ImportsV2) ProxyGetMapValues(mapType v2.MapType, keysData int32, keysSize int32,
	returnMapData int32, returnMapSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_get_map_values")

	return v2.ProxyGetMapValues(i.instance, mapType, keysData, keysSize, returnMapData, returnMapSize)
}

func (i *ImportsV2) ProxySetMapValues(mapType v2.MapType, removeKeysData int32, removeKeysSize int32,
	mapData int32, mapSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_set_map_values")

	return v2.ProxySetMapValues(i.instance, mapType, removeKeysData, removeKeysSize, mapData, mapSize)
}

func (i *ImportsV2) ProxyOpenSharedKvstore(kvstoreNameData int32, kvstoreNameSize int32, createIfNotExist int32,
	returnKvstoreID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_open_shared_kvstore")

	return v2.ProxyOpenSharedKvstore(i.instance, kvstoreNameData, kvstoreNameSize, createIfNotExist, returnKvstoreID)
}

func (i *ImportsV2) ProxyGetSharedKvstoreKeyValues(kvstoreID int32, keyData int32, keySize int32,
	returnValuesData int32, returnValuesSize int32, returnCas int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_get_shared_kvstore_key_values")

	return v2.ProxyGetSharedKvstoreKeyValues(i.instance, kvstoreID, keyData, keySize, returnValuesData, returnValuesSize, returnCas)
}

func (i *ImportsV2) ProxySetSharedKvstoreKeyValues(kvstoreID int32, keyData int32, keySize int32,
	valuesData int32, valuesSize int32, cas int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_set_shared_kvstore_key_values")

	return v2.ProxySetSharedKvstoreKeyValues(i.instance, kvstoreID, keyData, keySize, valuesData, valuesSize, cas)
}

func (i *ImportsV2) ProxyAddSharedKvstoreKeyValues(kvstoreID int32, keyData int32, keySize int32,
	valuesData int32, valuesSize int32, cas int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_add_shared_kvstore_key_values")

	return v2.ProxyAddSharedKvstoreKeyValues(i.instance, kvstoreID, keyData, keySize, valuesData, valuesSize, cas)
}

func (i *ImportsV2) ProxyRemoveSharedKvstoreKey(kvstoreID int32, keyData int32, keySize int32, cas int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_remove_shared_kvstore_key")

	return v2.ProxyRemoveSharedKvstoreKey(i.instance, kvstoreID, keyData, keySize, cas)
}

func (i *ImportsV2) ProxyDeleteSharedKvstore(kvstoreID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_delete_shared_kvstore")

	return v2.ProxyDeleteSharedKvstore(i.instance, kvstoreID)
}

func (i *ImportsV2) ProxyOpenSharedQueue(queueNameData int32, queueNameSize int32, createIfNotExist int32,
	returnQueueID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_open_shared_queue")

	return v2.ProxyOpenSharedQueue(i.instance, queueNameData, queueNameSize, createIfNotExist, returnQueueID)
}

func (i *ImportsV2) ProxyDequeueSharedQueueItem(queueID int32, returnPayloadData int32, returnPayloadSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_dequeue_shared_queue_item")

	return v2.ProxyDequeueSharedQueueItem(i.instance, queueID, returnPayloadData, returnPayloadSize)
}

func (i *ImportsV2) ProxyEnqueueSharedQueueItem(queueID int32, payloadData int32, payloadSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_enqueue_shared_queue_item")

	return v2.ProxyEnqueueSharedQueueItem(i.instance, queueID, payloadData, payloadSize)
}

func (i *ImportsV2) ProxyDeleteSharedQueue(queueID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_delete_shared_queue")

	return v2.ProxyDeleteSharedQueue(i.instance, queueID)
}

func (i *ImportsV2) ProxyCreateTimer(period int32, oneTime int32, returnTimerID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_create_timer")

	return v2.ProxyCreateTimer(i.instance, period, oneTime, returnTimerID)
}

func (i *ImportsV2) ProxyDeleteTimer(timerID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_delete_timer")

	return v2.ProxyDeleteTimer(i.instance, timerID)
}

func (i *ImportsV2) ProxyCreateMetric(metricType v2.MetricType,
	metricNameData int32, metricNameSize int32, returnMetricID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_create_metric")

	return v2.ProxyCreateMetric(i.instance, metricType, metricNameData, metricNameSize, returnMetricID)
}

func (i *ImportsV2) ProxyGetMetricValue(metricID int32, returnValue int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_get_metric_value")

	return v2.ProxyGetMetricValue(i.instance, metricID, returnValue)
}

func (i *ImportsV2) ProxySetMetricValue(metricID int32, value int64) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_set_metric_value")

	return v2.ProxySetMetricValue(i.instance, metricID, value)
}

func (i *ImportsV2) ProxyIncrementMetricValue(metricID int32, offset int64) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_increment_metric_value")

	return v2.ProxyIncrementMetricValue(i.instance, metricID, offset)
}

func (i *ImportsV2) ProxyDeleteMetric(metricID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_delete_metric")

	return v2.ProxyDeleteMetric(i.instance, metricID)
}

func (i *ImportsV2) ProxyDispatchHttpCall(upstreamNameData int32, upstreamNameSize int32, headersMapData int32, headersMapSize int32,
	bodyData int32, bodySize int32, trailersMapData int32, trailersMapSize int32, timeoutMilliseconds int32,
	returnCalloutID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_dispatch_http_call")

	return v2.ProxyDispatchHttpCall(i.instance, upstreamNameData, upstreamNameSize, headersMapData, headersMapSize, bodyData, bodySize, trailersMapData, trailersMapSize, timeoutMilliseconds, returnCalloutID)
}

func (i *ImportsV2) ProxyDispatchGrpcCall(upstreamNameData int32, upstreamNameSize int32, serviceNameData int32, serviceNameSize int32,
	serviceMethodData int32, serviceMethodSize int32, initialMetadataMapData int32, initialMetadataMapSize int32,
	grpcMessageData int32, grpcMessageSize int32, timeoutMilliseconds int32, returnCalloutID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_dispatch_grpc_call")

	return v2.ProxyDispatchGrpcCall(i.instance, upstreamNameData, upstreamNameSize, serviceNameData, serviceNameSize, serviceMethodData, serviceMethodSize, initialMetadataMapData, initialMetadataMapSize, grpcMessageData, grpcMessageSize, timeoutMilliseconds, returnCalloutID)
}

func (i *ImportsV2) ProxyOpenGrpcStream(upstreamNameData int32, upstreamNameSize int32, serviceNameData int32, serviceNameSize int32,
	serviceMethodData int32, serviceMethodSize int32, initialMetadataMapData int32, initialMetadataMapSize int32,
	returnCalloutID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_open_grpc_stream")

	return v2.ProxyOpenGrpcStream(i.instance, upstreamNameData, upstreamNameSize, serviceNameData, serviceNameSize, serviceMethodData, serviceMethodSize, initialMetadataMapData, initialMetadataMapSize, returnCalloutID)
}

func (i *ImportsV2) ProxySendGrpcStreamMessage(calloutID int32, grpcMessageData int32, grpcMessageSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_send_grpc_stream_message")

	return v2.ProxySendGrpcStreamMessage(i.instance, calloutID, grpcMessageData, grpcMessageSize)
}

func (i *ImportsV2) ProxyCancelGrpcCall(calloutID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_cancel_grpc_call")

	return v2.ProxyCancelGrpcCall(i.instance, calloutID)
}

func (i *ImportsV2) ProxyCloseGrpcCall(calloutID int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_close_grpc_call")

	return v2.ProxyCloseGrpcCall(i.instance, calloutID)
}

func (i *ImportsV2) ProxyCallCustomFunction(customFunctionID int32, parametersData int32, parametersSize int32,
	returnResultsData int32, returnResultsSize int32) v2.Result {
	i.instance.Logger().V(4).Info("call host function proxy_call_custom_function")

	return v2.ProxyCallCustomFunction(i.instance, customFunctionID, parametersData, parametersSize, returnResultsData, returnResultsSize)
}
