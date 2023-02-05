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
	"bytes"
	"fmt"
	"io"
	"net/http"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"

	"github.com/cisco-open/nasp/pkg/proxywasm"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type Handler interface {
	http.Handler

	Logger() logr.Logger
	AddMiddleware(middleware HandleMiddleware)
	RemoveMiddleware(middleware HandleMiddleware)
}

type httpHandler struct {
	MiddlewareHandler

	handler       http.Handler
	streamHandler api.StreamHandler
	direction     api.ListenerDirection
}

func NewHandler(handler http.Handler, streamHandler api.StreamHandler, direction api.ListenerDirection) Handler {
	return &httpHandler{
		MiddlewareHandler: NewMiddlewareHandler(),
		handler:           handler,
		streamHandler:     streamHandler,
		direction:         direction,
	}
}

func (h *httpHandler) Logger() logr.Logger {
	return h.streamHandler.Logger()
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	stream, err := h.streamHandler.NewStream(h.direction)
	if err != nil {
		h.streamHandler.Logger().Error(err, "could not create new stream")
		return
	}

	defer func() {
		err = stream.Close()
		if err != nil {
			h.streamHandler.Logger().Error(err, "could not close stream")
		}
	}()

	h.serve(w, req, stream)
}

func (h *httpHandler) serve(responseWriter http.ResponseWriter, request *http.Request, stream api.Stream) {
	if _, ok := h.handler.(*grpc.Server); ok {
		if _, ok := responseWriter.(http.Flusher); !ok {
			h.Logger().Error(errors.New("gRPC requires a ResponseWriter supporting http.Flusher"), "")
			return
		}
	}

	wrappedRequest := WrapHTTPRequest(request)

	crw := &customResponseWriter{
		responseWriter: responseWriter,
	}

	resp := &http.Response{}
	resp.Header = responseWriter.Header()

	wrappedResponse := WrapHTTPResponse(resp)

	request.Header.Add(":method", request.Method)

	proxywasm.HTTPRequestProperty(stream).Set(wrappedRequest)
	proxywasm.HTTPResponseProperty(stream).Set(wrappedResponse)

	h.BeforeRequest(wrappedRequest, stream)

	if err := stream.HandleHTTPRequest(wrappedRequest); err != nil {
		h.Logger().Error(err, "could not handle request")
		return
	}

	if resp.StatusCode > 0 {
		h.writeResponse(crw, responseWriter, resp)
		return
	}

	h.handler.ServeHTTP(crw, request)

	h.AfterRequest(wrappedRequest, stream)

	resp.ContentLength = int64(len(crw.body))
	if crw.statusCode == 0 {
		crw.statusCode = 200
	}
	resp.StatusCode = crw.statusCode
	resp.Body = io.NopCloser(bytes.NewReader(crw.body))

	h.BeforeResponse(wrappedResponse, stream)

	if err := stream.HandleHTTPResponse(wrappedResponse); err != nil {
		h.Logger().Error(err, "could not handle response")
		return
	}

	h.AfterResponse(wrappedResponse, stream)

	h.writeResponse(crw, responseWriter, resp)
}

func (h *httpHandler) writeResponse(crw *customResponseWriter, responseWriter http.ResponseWriter, resp *http.Response) {
	h.convertTrailers(resp.Header, resp.Trailer)

	responseWriter.WriteHeader(resp.StatusCode)

	crw.flushable = true
	crw.Flush()

	_, err := io.Copy(responseWriter, resp.Body)
	if err != nil {
		h.Logger().Error(err, "could not write response")
	}
}

func (h *httpHandler) convertTrailers(header http.Header, trailer http.Header) {
	for k, v := range trailer {
		header.Add("trailer", k)
		for _, vv := range v {
			header.Add(fmt.Sprintf("%s%s", http.TrailerPrefix, k), vv)
		}
	}
}

type customResponseWriter struct {
	responseWriter http.ResponseWriter
	body           []byte
	statusCode     int
	flushable      bool
}

func (w *customResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)

	return len(b), nil
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *customResponseWriter) Flush() {
	if !w.flushable {
		return
	}

	if flusher, ok := w.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
