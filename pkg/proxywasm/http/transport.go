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
	"net/http"

	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

type HTTPTransport struct {
	MiddlewareHandler

	streamHandler api.StreamHandler
	transport     http.RoundTripper
	logger        logr.Logger
}

func NewHTTPTransport(transport http.RoundTripper, streamHandler api.StreamHandler, logger logr.Logger) *HTTPTransport {
	return &HTTPTransport{
		MiddlewareHandler: NewMiddlewareHandler(),

		streamHandler: streamHandler,
		logger:        logger,
		transport:     transport,
	}
}

func (t *HTTPTransport) Logger() logr.Logger {
	return t.logger
}

func (t *HTTPTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	origURL := r.URL.String()
	t.logger.V(2).Info("sending request", "url", origURL)

	stream, err := t.streamHandler.NewStream(api.ListenerDirectionOutbound)
	if err != nil {
		t.logger.Error(err, "could not get new stream")
		return nil, err
	}
	defer stream.Close()

	req := r.Clone(context.Background())

	wrappedRequest := WrapHTTPRequest(req)

	wrappedRequest, stream = t.BeforeRequest(wrappedRequest, stream)

	if err := stream.HandleHTTPRequest(wrappedRequest); err != nil {
		t.logger.Error(err, "could not handle request")
		return nil, err
	}

	t.AfterRequest(wrappedRequest, stream)

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		t.logger.Error(err, "error at roundtrip")
		return nil, err
	}

	wrappedResponse := WrapHTTPResponse(resp)

	wrappedResponse, stream = t.BeforeResponse(req.Context(), wrappedResponse, stream)

	if err := stream.HandleHTTPResponse(wrappedResponse); err != nil {
		t.logger.Error(err, "could not handle response")
		return nil, err
	}

	t.AfterResponse(req.Context(), wrappedResponse, stream)

	t.logger.V(2).Info("response received", "url", origURL)

	return resp, nil
}
