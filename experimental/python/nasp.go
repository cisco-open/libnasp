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

import "C"
import (
	"fmt"
	"hash/fnv"
	"k8s.io/klog/v2"
	"strconv"
)

/*
#include <stdlib.h>

struct GoError {
	char *error_msg;
};

inline struct GoError *create_go_error(char *error_msg) {
	struct GoError *p;

	p = malloc(sizeof(struct GoError));
	p->error_msg = error_msg;

	return p;
};

struct GoHttpHeader {
	char *key;
	char *value;
};

struct GoHttpHeaders {
	struct GoHttpHeader *items;
	unsigned int len;
};

struct GoHttpResponse {
	int status_code;
	char *version;
	struct GoHttpHeaders headers;
    char *body;
};

inline struct GoHttpResponse *create_go_http_response(int status_code, char *version, struct GoHttpHeaders headers, char *body) {
	struct GoHttpResponse *p;

	p = malloc(sizeof(struct GoHttpResponse));
	p->status_code = status_code;
	p->version = version;
	p->headers = headers;
	p->body = body;

	return p;
};

inline struct GoHttpHeader *new_http_headers(unsigned int count) {
	struct GoHttpHeader *p;

	p = malloc(sizeof(struct GoHttpHeader) * count);

	return p;
};


inline void set_http_header_at(struct GoHttpHeader *headers, struct GoHttpHeader header, unsigned int index) {
	headers[index] = header;
};

inline void free_http_headers(struct GoHttpHeader *headers, unsigned int size) {
	if(headers == NULL) {
		return;
	}

	for(int i = 0; i < size; i++) {
		free(headers[i].key);
		free(headers[i].value);
	}

	free(headers);
};

*/
import "C"

import (
	"context"
	"emperror.dev/errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"unsafe"

	"github.com/cisco-open/nasp/pkg/istio"
)

var logger = klog.NewKlogr()

var httpTransportPool = httpTransportPoolManager{}

//export NewHTTPTransport
func NewHTTPTransport(heimdallURLPtr, clientIDPtr, clientSecretPtr *C.char) (C.ulonglong, *C.struct_GoError) {
	heimdallURL := C.GoString(heimdallURLPtr)
	clientID := C.GoString(clientIDPtr)
	clientSecret := C.GoString(clientSecretPtr)

	hash := fnv.New64()
	hash.Write([]byte(heimdallURL))
	hash.Write([]byte(clientID))
	hash.Write([]byte(clientSecret))
	httpTransportIDHex := fmt.Sprintf("%x", hash.Sum(nil))
	httpTransportID, _ := strconv.ParseUint(httpTransportIDHex, 16, 64)

	if _, ok := httpTransportPool.Get(httpTransportID); ok {
		return C.ulonglong(httpTransportID), nil
	}

	istioHandlerConfig := &istio.IstioIntegrationHandlerConfig{
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, clientID, clientSecret, "v1"),
		PushgatewayConfig: &istio.PushgatewayConfig{
			Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
			UseUniqueIDLabel: true,
		},
	}

	iih, err := istio.NewIstioIntegrationHandler(istioHandlerConfig, logger)

	if err != nil {
		return C.ulonglong(0), cGoError(err)
	}

	transport, err := iih.GetHTTPTransport(http.DefaultTransport)
	if err != nil {
		return C.ulonglong(0), cGoError(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	client := &http.Client{
		Transport: transport,
	}

	httpTransport := &HTTPTransport{iih: iih, client: client, cancel: cancel}
	httpTransportPool.Set(httpTransportID, httpTransport)

	return C.ulonglong(httpTransportID), nil
}

//export SendHTTPRequest
func SendHTTPRequest(httpTransportID C.ulonglong, method, url, body *C.char) (*C.struct_GoHttpResponse, *C.struct_GoError) {
	httpTransport, ok := httpTransportPool.Get(uint64(httpTransportID))
	if !ok {
		return cGoInternalServerErrorHttpResponse(), cGoError(errors.NewPlain("transport not found"))
	}

	httpResp, err := httpTransport.Request(
		C.GoString(method),
		C.GoString(url),
		C.GoString(body))

	if err != nil {
		return cGoInternalServerErrorHttpResponse(), cGoError(err)
	}

	var headers map[string]string
	if httpResp.Headers != nil && httpResp.Headers != nil && len(httpResp.Headers.header) > 0 {
		headers = make(map[string]string, len(httpResp.Headers.header))
		for k, v := range httpResp.Headers.header {
			headers[k] = strings.Join(v, ",")
		}
	}

	return cGoHttpResponse(httpResp.StatusCode, httpResp.Version, headers, httpResp.Body), nil
}

//export CloseHTTPTransport
func CloseHTTPTransport(httpTransportID C.ulonglong) {
	if httpTransport, ok := httpTransportPool.GetAndDelete(uint64(httpTransportID)); ok {
		httpTransport.Close()
	}
}

type HTTPTransport struct {
	iih    *istio.IstioIntegrationHandler
	client *http.Client
	cancel context.CancelFunc
}

type HTTPHeader struct {
	header http.Header
}

func (h *HTTPHeader) Get(key string) string {
	return h.header.Get(key)
}

func (h *HTTPHeader) Set(key, value string) {
	h.header.Set(key, value)
}

type HTTPResponse struct {
	StatusCode int
	Version    string
	Headers    *HTTPHeader
	Body       []byte
}

func (t *HTTPTransport) Request(method, url, body string) (*HTTPResponse, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	request, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	response, err := t.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: response.StatusCode,
		Version:    response.Proto,
		Headers:    &HTTPHeader{response.Header},
		Body:       bodyBytes,
	}, nil
}

func (t *HTTPTransport) Close() {
	t.cancel()
}

type httpTransportPoolManager struct {
	pool map[uint64]*HTTPTransport
	mu   sync.RWMutex
}

func (m *httpTransportPoolManager) Get(id uint64) (*HTTPTransport, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	httpTransport, ok := m.pool[id]
	return httpTransport, ok
}

func (m *httpTransportPoolManager) GetAndDelete(id uint64) (*HTTPTransport, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	httpTransport, ok := m.pool[id]
	if ok {
		delete(m.pool, id)
	}
	return httpTransport, ok
}

func (m *httpTransportPoolManager) Set(id uint64, httpTransport *HTTPTransport) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pool == nil {
		m.pool = make(map[uint64]*HTTPTransport)
	}
	m.pool[id] = httpTransport
}

func cGoError(err error) *C.struct_GoError {
	goErr := C.create_go_error(C.CString(err.Error()))

	return goErr
}

//export free_go_error
func free_go_error(p *C.struct_GoError) {
	ptr := unsafe.Pointer(p)
	if ptr != nil {
		C.free(unsafe.Pointer(p.error_msg))
	}

	C.free(ptr)
}

func cGoHttpResponse(statusCode int, version string, httpHeaders map[string]string, body []byte) *C.struct_GoHttpResponse {
	var headers *C.struct_GoHttpHeader
	count := 0

	if len(httpHeaders) > 0 {
		headers = C.new_http_headers(C.uint(len(httpHeaders)))
		for k, v := range httpHeaders {
			hdr := C.struct_GoHttpHeader{
				key:   C.CString(k),
				value: C.CString(v),
			}

			C.set_http_header_at(headers, hdr, C.uint(count))
			count++
		}
	}

	httpResponse := C.create_go_http_response(
		C.int(statusCode),
		C.CString(version),
		C.struct_GoHttpHeaders{
			items: headers,
			len:   C.uint(count),
		},
		C.CString(string(body)),
	)

	return httpResponse
}

func cGoInternalServerErrorHttpResponse() *C.struct_GoHttpResponse {
	return cGoHttpResponse(http.StatusInternalServerError, "", nil, nil)
}

//export free_go_http_response
func free_go_http_response(p *C.struct_GoHttpResponse) {
	ptr := unsafe.Pointer(p)
	if ptr != nil {
		C.free(unsafe.Pointer(p.version))
		C.free_http_headers(p.headers.items, p.headers.len)
		C.free(unsafe.Pointer(p.body))
	}

	C.free(ptr)
}

func main() {}
