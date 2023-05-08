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

package grpc

import (
	"io"
	"net/url"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type GRPCRequest struct {
	url             *url.URL
	header          metadata.MD
	connectionState network.ConnectionState
}

type GRPCResponse struct {
	statusCode      codes.Code
	header          metadata.MD
	trailer         metadata.MD
	connectionState network.ConnectionState
}

func WrapGRPCResponse(statusCode codes.Code, header metadata.MD, trailer metadata.MD, connectionState network.ConnectionState) api.HTTPResponse {
	return &GRPCResponse{
		statusCode:      statusCode,
		header:          header,
		trailer:         trailer,
		connectionState: connectionState,
	}
}

func (r *GRPCResponse) Header() api.HeaderMap {
	return WrapGRPCMetadata(r.header)
}

func (r *GRPCResponse) Trailer() api.HeaderMap {
	return WrapGRPCMetadata(r.trailer)
}

func (r *GRPCResponse) Body() io.ReadCloser {
	return nil
}

func (r *GRPCResponse) SetBody(io.ReadCloser) {
}

func (r *GRPCResponse) ContentLength() int64 {
	return 0
}

func (r *GRPCResponse) ConnectionState() network.ConnectionState {
	return r.connectionState
}

func (r *GRPCResponse) StatusCode() int {
	switch r.statusCode {
	case codes.OK:
		return 200
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return 500
	case codes.InvalidArgument:
		return 400
	case codes.DeadlineExceeded:
		return 504
	case codes.NotFound:
		return 404
	case codes.AlreadyExists:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.ResourceExhausted:
		return 429
	case codes.FailedPrecondition:
		return 400
	case codes.Aborted:
		return 409
	case codes.OutOfRange:
		return 400
	case codes.Unimplemented:
		return 501
	case codes.Internal:
		return 500
	case codes.Unavailable:
		return 503
	case codes.DataLoss:
		return 500
	case codes.Unauthenticated:
		return 401
	default:
		return 500
	}
}

func WrapGRPCRequest(rawURL string, header metadata.MD, connectionState network.ConnectionState) api.HTTPRequest {
	r := &GRPCRequest{
		url:             &url.URL{},
		header:          header,
		connectionState: connectionState,
	}

	if u, err := url.Parse(rawURL); err != nil {
		r.url = u
	}

	return r
}

func (r *GRPCRequest) URL() *url.URL {
	return r.url
}

func (r *GRPCRequest) Header() api.HeaderMap {
	return WrapGRPCMetadata(r.header)
}

func (r *GRPCRequest) Trailer() api.HeaderMap {
	return WrapGRPCMetadata(metadata.MD{})
}

func (r *GRPCRequest) Body() io.ReadCloser {
	return nil
}

func (r *GRPCRequest) SetBody(io.ReadCloser) {
}

func (r *GRPCRequest) HTTPProtocol() string {
	return ""
}

func (r *GRPCRequest) Host() string {
	return r.url.Host
}

func (r *GRPCRequest) Method() string {
	return ""
}

func (r *GRPCRequest) ConnectionState() network.ConnectionState {
	return r.connectionState
}

type grpcHeaderMap struct {
	headers metadata.MD
}

func WrapGRPCMetadata(headers metadata.MD) api.HeaderMap {
	return &grpcHeaderMap{
		headers: headers,
	}
}

func (m *grpcHeaderMap) Get(key string) (string, bool) {
	value := m.headers.Get(key)
	if len(value) > 0 {
		return value[0], true
	}

	return "", false
}

func (m *grpcHeaderMap) Set(key, value string) {
	m.headers.Set(key, value)
}

func (m *grpcHeaderMap) Add(key, value string) {
	m.headers.Append(key, value)
}

func (m *grpcHeaderMap) Del(key string) {
	m.headers.Delete(key)
}

func (m *grpcHeaderMap) Keys() []string {
	keys := []string{}

	for k := range m.headers {
		keys = append(keys, k)
	}

	return keys
}

func (m *grpcHeaderMap) Len() int {
	return len(m.headers)
}

func (m *grpcHeaderMap) Raw() map[string][]string {
	return m.headers
}

func (m *grpcHeaderMap) Clone() api.HeaderMap {
	return WrapGRPCMetadata(m.headers.Copy())
}
