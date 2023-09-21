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

package http

import (
	"io"
	"net/http"
	"net/url"

	"github.com/cisco-open/libnasp/pkg/network"
	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
)

type HTTPRequest struct {
	Request *http.Request
}

type HTTPResponse struct {
	Response *http.Response
}

func WrapHTTPRequest(req *http.Request) api.HTTPRequest {
	return &HTTPRequest{
		Request: req,
	}
}

func WrapHTTPResponse(resp *http.Response) api.HTTPResponse {
	return &HTTPResponse{
		Response: resp,
	}
}

func (r *HTTPRequest) URL() *url.URL {
	return r.Request.URL
}

func (r *HTTPRequest) Header() api.HeaderMap {
	return WrapHTTPHeader(r.Request.Header)
}

func (r *HTTPRequest) Trailer() api.HeaderMap {
	return WrapHTTPHeader(r.Request.Trailer)
}

func (r *HTTPRequest) Body() io.ReadCloser {
	return r.Request.Body
}

func (r *HTTPRequest) SetBody(body io.ReadCloser) {
	r.Request.Body = body
}

func (r *HTTPRequest) HTTPProtocol() string {
	return r.Request.Proto
}

func (r *HTTPRequest) Host() string {
	return r.Request.Host
}

func (r *HTTPRequest) Method() string {
	return r.Request.Method
}

func (r *HTTPRequest) ConnectionState() network.ConnectionState {
	if connectionState, ok := network.ConnectionStateFromContext(r.Request.Context()); ok {
		return connectionState
	}

	return nil
}

func (r *HTTPResponse) GetHTTPResponse() *http.Response {
	return r.Response
}

func (r *HTTPResponse) Header() api.HeaderMap {
	return WrapHTTPHeader(r.Response.Header)
}

func (r *HTTPResponse) Trailer() api.HeaderMap {
	return WrapHTTPHeader(r.Response.Header)
}

func (r *HTTPResponse) Body() io.ReadCloser {
	return r.Response.Body
}

func (r *HTTPResponse) SetBody(body io.ReadCloser) {
	r.Response.Body = body
}

func (r *HTTPResponse) ContentLength() int64 {
	return r.Response.ContentLength
}

func (r *HTTPResponse) StatusCode() int {
	return r.Response.StatusCode
}

func (r *HTTPResponse) ConnectionState() network.ConnectionState {
	if connectionState, ok := network.ConnectionStateFromContext(r.Response.Request.Context()); ok {
		return connectionState
	}

	return nil
}

type httpHeaderMap struct {
	headers http.Header
}

func WrapHTTPHeader(headers http.Header) api.HeaderMap {
	return &httpHeaderMap{
		headers: headers,
	}
}

func (m *httpHeaderMap) Get(key string) (string, bool) {
	value := m.headers.Get(key)
	if value == "" {
		return value, false
	}

	return value, true
}

func (m *httpHeaderMap) Set(key, value string) {
	m.headers.Set(key, value)
}

func (m *httpHeaderMap) Add(key, value string) {
	m.headers.Add(key, value)
}

func (m *httpHeaderMap) Del(key string) {
	m.headers.Del(key)
}

func (m *httpHeaderMap) Keys() []string {
	keys := []string{}

	for k := range m.headers {
		keys = append(keys, k)
	}

	return keys
}

func (m *httpHeaderMap) Len() int {
	return len(m.headers)
}

func (m *httpHeaderMap) Raw() map[string][]string {
	return m.headers
}

func (m *httpHeaderMap) Clone() api.HeaderMap {
	return WrapHTTPHeader(m.headers.Clone())
}
