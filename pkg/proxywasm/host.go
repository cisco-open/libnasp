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
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	klog "k8s.io/klog/v2"
	"mosn.io/proxy-wasm-go-host/proxywasm/common"
	v1 "mosn.io/proxy-wasm-go-host/proxywasm/v1"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type HostFunctions struct {
	v1.ImportsHandler

	logger logr.Logger

	properties api.PropertyHolder
	metrics    api.MetricHandler
}

type HostFunctionsOption func(hf *HostFunctions)

func SetHostFunctionsLogger(logger logr.Logger) HostFunctionsOption {
	return func(hf *HostFunctions) {
		hf.logger = logger
	}
}

func SetHostFunctionsMetricHandler(h api.MetricHandler) HostFunctionsOption {
	return func(hf *HostFunctions) {
		hf.metrics = h
	}
}

func NewHostFunctions(properties api.PropertyHolder, options ...HostFunctionsOption) *HostFunctions {
	hf := &HostFunctions{
		ImportsHandler: NewDefaultHostFunctions(),

		properties: properties,
	}

	for _, o := range options {
		o(hf)
	}

	if hf.logger == (logr.Logger{}) {
		hf.logger = klog.Background()
	}

	return hf
}

func (f *HostFunctions) Logger() logr.Logger {
	return f.logger
}

// wasm host functions

func (f *HostFunctions) GetPluginConfig() common.IoBuffer {
	if val, found := f.properties.Get("plugin_config.bytes"); found {
		if content, ok := val.([]byte); ok {
			return common.NewIoBufferBytes(content)
		}
	}

	return nil
}

func (f *HostFunctions) Log(level v1.LogLevel, msg string) v1.WasmResult {
	logLevel := 0
	if level < 2 {
		logLevel = 2 - int(level)
	}

	f.Logger().V(logLevel).WithName("wasm filter").Info(msg, "level", level)

	return v1.WasmResultOk
}

func (f *HostFunctions) CallForeignFunction(funcName string, param []byte) ([]byte, v1.WasmResult) {
	return nil, v1.WasmResultUnimplemented
}

func (f *HostFunctions) GetProperty(key string) (string, v1.WasmResult) {
	key = strings.TrimRight(strings.ReplaceAll(key, "\x00", "."), ".")

	if v, ok := f.properties.Get(key); ok {
		f.Logger().V(2).Info("get property", "key", key, "value", v)

		return Stringify(v), v1.WasmResultOk
	}

	f.Logger().V(2).Info("get property", "key", key, "value", "(MISSING)")

	return "", v1.WasmResultNotFound
}

func (f *HostFunctions) SetProperty(key, value string) v1.WasmResult {
	f.Logger().V(2).Info("set property", "key", key, "value", value)

	f.properties.Set(key, value)

	return v1.WasmResultOk
}

func (f *HostFunctions) GetHttpRequestHeader() common.HeaderMap {
	var value interface{}
	var ok bool

	value, ok = f.properties.Get("http.request")
	if !ok {
		return nil
	}

	if req, ok := value.(api.HTTPRequest); ok {
		return NewHeaders(req.Header(), f.Logger())
	}

	return nil
}

func (f *HostFunctions) GetHttpRequestBody() common.IoBuffer {
	if val, ok := f.properties.Get("request.body"); ok {
		if content, ok := val.(common.IoBuffer); ok {
			return content
		}
	}

	return nil
}

func (f *HostFunctions) GetHttpResponseBody() common.IoBuffer {
	if val, ok := f.properties.Get("response.body"); ok {
		if content, ok := val.(common.IoBuffer); ok {
			return content
		}
	}

	return nil
}

func (f *HostFunctions) GetHttpResponseHeader() common.HeaderMap {
	var value interface{}
	var ok bool

	value, ok = f.properties.Get("http.response")
	if !ok {
		return nil
	}

	if resp, ok := value.(api.HTTPResponse); ok {
		return NewHeaders(resp.Header(), f.Logger())
	}

	return nil
}

func (f *HostFunctions) GetDownStreamData() common.IoBuffer {
	if val, found := f.properties.Get("downstream.data"); found {
		if content, ok := val.(common.IoBuffer); ok {
			return content
		}
	}

	return nil
}

func (f *HostFunctions) GetUpstreamData() common.IoBuffer {
	if val, found := f.properties.Get("upstream.data"); found {
		if content, ok := val.(common.IoBuffer); ok {
			return content
		}
	}

	return nil
}

func (f *HostFunctions) SendHttpResp(respCode int32, respCodeDetail common.IoBuffer, respBody common.IoBuffer, additionalHeaderMap common.HeaderMap, grpcCode int32) v1.WasmResult {
	if value, ok := f.properties.Get("http.response"); ok {
		if resp, ok := value.(interface {
			GetHTTPResponse() *http.Response
		}); ok {
			hresp := resp.GetHTTPResponse() //nolint:bodyclose
			hresp.StatusCode = int(respCode)
			hresp.Status = http.StatusText(int(respCode))
			hresp.Body = io.NopCloser(bytes.NewReader(respBody.Bytes()))
			additionalHeaderMap.Range(func(key, value string) bool {
				hresp.Header.Add(key, value)
				return true
			})
		}

		return v1.WasmResultOk
	}

	return v1.WasmResultNotFound
}

func (f *HostFunctions) SetEffectiveContextID(contextID int32) v1.WasmResult {
	var rootContext v1.ContextHandler
	if ctx, ok := RootABIContextProperty(f.properties).Get(); ok {
		rootContext = ctx
	}
	if rootContext == nil {
		return v1.WasmResultInternalFailure
	}

	var plugin api.WasmPlugin
	if plug, ok := PluginProperty(f.properties).Get(); ok {
		plugin = plug
	}
	if plugin == nil {
		return v1.WasmResultInternalFailure
	}

	// root context
	if contextID == plugin.Context().ID() {
		rootContext.GetInstance().SetData(rootContext)

		return v1.WasmResultOk
	}

	// filter context
	fc, found := plugin.GetFilterContext(rootContext.GetInstance(), contextID)
	if found {
		rootContext.GetInstance().SetData(fc.GetABIContext())

		return v1.WasmResultOk
	}

	return v1.WasmResultNotFound
}

func (f *HostFunctions) SetTickPeriodMilliseconds(tickPeriodMilliseconds int32) v1.WasmResult {
	var rootContext v1.ContextHandler
	if ctx, ok := RootABIContextProperty(f.properties).Get(); ok {
		rootContext = ctx
	}
	if rootContext == nil {
		return v1.WasmResultInternalFailure
	}

	var plugin api.WasmPlugin
	if plug, ok := PluginProperty(f.properties).Get(); ok {
		plugin = plug
	}
	if plugin == nil {
		return v1.WasmResultInternalFailure
	}

	logger := f.logger.WithValues("contextID", plugin.Context().ID())

	go func() {
		period := time.Duration(tickPeriodMilliseconds) * time.Millisecond
		ticker := time.NewTicker(period)
		defer func() {
			logger.V(2).Info("stop ticker")
			ticker.Stop()
		}()

		logger.V(2).Info("start ticker", "period", period)

		done := make(chan bool)
		TickerDoneChannelProperty(f.properties).Set(done)

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				func() {
					rootContext.GetInstance().Lock(rootContext)
					defer func() {
						rootContext.GetInstance().Unlock()
					}()

					if err := rootContext.GetExports().ProxyOnTick(plugin.Context().ID()); err != nil {
						logger.Error(err, "error at proxy_on_tick")
					}
				}()
			}
		}
	}()

	return v1.WasmResultOk
}

func (f *HostFunctions) Done() v1.WasmResult {
	if wph, ok := f.properties.(api.WrappedPropertyHolder); ok {
		if done, ok := TickerDoneChannelProperty(wph.Properties()).Get(); ok {
			done <- true
		}
	}

	return v1.WasmResultOk
}

func (f *HostFunctions) DefineMetric(metricType v1.MetricType, name string) (int32, v1.WasmResult) {
	if f.metrics == nil {
		return 0, v1.WasmResultUnimplemented
	}

	retval := f.metrics.DefineMetric(int32(metricType), name)
	if retval < 0 {
		return retval, v1.WasmResultInternalFailure
	}

	return retval, v1.WasmResultOk
}

func (f *HostFunctions) RecordMetric(metricID int32, value int64) v1.WasmResult {
	if f.metrics == nil {
		return v1.WasmResultNotFound
	}

	err := f.metrics.RecordMetric(metricID, value)
	if err != nil {
		return v1.WasmResultNotFound
	}

	return v1.WasmResultOk
}

func (f *HostFunctions) IncrementMetric(metricID int32, offset int64) v1.WasmResult {
	if f.metrics == nil {
		return v1.WasmResultNotFound
	}

	err := f.metrics.IncrementMetric(metricID, offset)
	if err != nil {
		return v1.WasmResultNotFound
	}

	return v1.WasmResultOk
}
