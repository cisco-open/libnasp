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
	"context"
	"sync"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type HandleMiddleware interface {
	BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream)
	AfterRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream)
	BeforeResponse(requestContext context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream)
	AfterResponse(requestContext context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream)
}

type Middlewares map[HandleMiddleware]struct{}

type MiddlewareHandler interface {
	HandleMiddleware

	AddMiddleware(middleware HandleMiddleware)
	RemoveMiddleware(middleware HandleMiddleware)
	GetMiddlewares() Middlewares
}

type middlewareHandler struct {
	middlewares Middlewares
	mm          sync.RWMutex
}

func NewMiddlewareHandler() MiddlewareHandler {
	return &middlewareHandler{
		middlewares: Middlewares{},
	}
}

func (h *middlewareHandler) AddMiddleware(middleware HandleMiddleware) {
	h.mm.Lock()
	defer h.mm.Unlock()

	h.middlewares[middleware] = struct{}{}
}

func (h *middlewareHandler) RemoveMiddleware(middleware HandleMiddleware) {
	h.mm.Lock()
	defer h.mm.Unlock()

	delete(h.middlewares, middleware)
}

func (h *middlewareHandler) GetMiddlewares() Middlewares {
	h.mm.RLock()
	defer h.mm.RUnlock()

	return h.middlewares
}

func (h *middlewareHandler) BeforeResponse(requestContext context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	for mw := range h.GetMiddlewares() {
		resp, stream = mw.BeforeResponse(requestContext, resp, stream)
	}
	return resp, stream
}

func (h *middlewareHandler) AfterResponse(requestContext context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	for mw := range h.GetMiddlewares() {
		resp, stream = mw.AfterResponse(requestContext, resp, stream)
	}
	return resp, stream
}

func (h *middlewareHandler) BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	for mw := range h.GetMiddlewares() {
		req, stream = mw.BeforeRequest(req, stream)
	}
	return req, stream
}

func (h *middlewareHandler) AfterRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	for mw := range h.GetMiddlewares() {
		req, stream = mw.AfterRequest(req, stream)
	}
	return req, stream
}
