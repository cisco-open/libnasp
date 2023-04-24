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

package middleware

import (
	"net"
	"strconv"
	"time"

	"github.com/blend/go-sdk/envoyutil"

	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/proxywasm"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type envoyHttpHandlerMiddleware struct {
}

func NewEnvoyHTTPHandlerMiddleware() lhttp.HandleMiddleware {
	return &envoyHttpHandlerMiddleware{}
}

func (m *envoyHttpHandlerMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) {
	if stream.Direction() == api.ListenerDirectionOutbound {
		m.beforeOutboundRequest(req, stream)
		return
	}

	m.beforeInboundRequest(req, stream)
}

func (m *envoyHttpHandlerMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) {
	if stream.Direction() == api.ListenerDirectionOutbound {
		m.afterOutboundRequest(req, stream)
		return
	}

	m.afterInboundRequest(req, stream)
}

func (m *envoyHttpHandlerMiddleware) BeforeResponse(resp api.HTTPResponse, stream api.Stream) {
	if stream.Direction() == api.ListenerDirectionOutbound {
		m.beforeOutboundResponse(resp, stream)
		return
	}

	m.beforeInboundResponse(resp, stream)
}

func (m *envoyHttpHandlerMiddleware) AfterResponse(resp api.HTTPResponse, stream api.Stream) {
	if stream.Direction() == api.ListenerDirectionOutbound {
		m.afterOutboundResponse(resp, stream)
		return
	}

	m.afterInboundResponse(resp, stream)
}

// INBOUND

func (m *envoyHttpHandlerMiddleware) setXForwardedHeaders(req api.HTTPRequest, stream api.Stream) {
	getStringProperty := func(p api.PropertyHolder, key string) string {
		if raw, ok := p.Get(key); ok {
			if str, ok := raw.(string); ok {
				return str
			}
		}

		return ""
	}

	xfccHeader := envoyutil.XFCCElement{
		By:      getStringProperty(stream, "connection.uri_san_local_certificate"),
		Hash:    getStringProperty(stream, "connection.sha256_peer_certificate_digest"),
		Subject: getStringProperty(stream, "connection.uri_san_peer_certificate"),
		DNS:     []string{getStringProperty(stream, "dns_san_peer_certificate")},
	}

	req.Header().Add(envoyutil.HeaderXFCC, xfccHeader.String())
	req.Header().Add("X-Forwarded-Proto", "https")
}

func (m *envoyHttpHandlerMiddleware) beforeInboundRequest(req api.HTTPRequest, stream api.Stream) {
	if connection := req.Connection(); connection != nil {
		m.setRequestProperties(req, stream, connection.GetTimeToFirstByte())
		SetEnvoyConnectionInfo(connection, stream)
		m.setXForwardedHeaders(req, stream)
	}
}

func (m *envoyHttpHandlerMiddleware) afterInboundRequest(req api.HTTPRequest, stream api.Stream) {
}

func (m *envoyHttpHandlerMiddleware) beforeInboundResponse(resp api.HTTPResponse, stream api.Stream) {
}

func (m *envoyHttpHandlerMiddleware) afterInboundResponse(resp api.HTTPResponse, stream api.Stream) {
	m.setResponseInfo(resp, stream)

	responseSize := uint64(resp.ContentLength()) + (proxywasm.NewHeaders(resp.Trailer(), stream.Logger())).ByteSize() + (proxywasm.NewHeaders(resp.Header(), stream.Logger())).ByteSize()
	stream.Logger().V(3).Info("response size", "size", responseSize)

	stream.Set("response.total_size", responseSize)
}

// OUTBOUND

func (m *envoyHttpHandlerMiddleware) beforeOutboundRequest(req api.HTTPRequest, stream api.Stream) {
	m.setRequestProperties(req, stream, time.Now())
}

func (m *envoyHttpHandlerMiddleware) afterOutboundRequest(req api.HTTPRequest, stream api.Stream) {
}

func (m *envoyHttpHandlerMiddleware) beforeOutboundResponse(resp api.HTTPResponse, stream api.Stream) {
	if connection := resp.Connection(); connection != nil {
		SetEnvoyConnectionInfo(connection, stream)
	}
}

func (m *envoyHttpHandlerMiddleware) afterOutboundResponse(resp api.HTTPResponse, stream api.Stream) {
	m.setResponseInfo(resp, stream)

	responseSize := uint64(resp.ContentLength()) + (proxywasm.NewHeaders(resp.Trailer(), stream.Logger())).ByteSize() + (proxywasm.NewHeaders(resp.Header(), stream.Logger())).ByteSize()
	stream.Logger().V(3).Info("response size", "size", responseSize)

	stream.Set("response.total_size", responseSize)
}

func (m *envoyHttpHandlerMiddleware) setResponseInfo(resp api.HTTPResponse, stream api.Stream) {
	flags := 0
	grpcStatus := 0

	if val, ok := stream.Get("response.flags"); ok {
		if f, ok := val.(int); ok {
			flags = f
		}
	}

	if val, ok := stream.Get("grpc.status"); ok {
		if s, ok := val.(int); ok {
			grpcStatus = s
		}
	}

	stream.Set("response", map[string]interface{}{
		"flags":       flags,
		"code":        resp.StatusCode(),
		"grpc_status": grpcStatus,
		"headers":     proxywasm.NewHeaders(resp.Header(), stream.Logger()).Flatten(),
		"trailers":    proxywasm.NewHeaders(resp.Trailer(), stream.Logger()).Flatten(),
	})

	if val, ok := stream.Get("request.time"); ok {
		if ttfb, ok := val.(int64); ok {
			stream.Set("request.duration", time.Now().UnixNano()-ttfb)
		}
	}
}

func SetEnvoyConnectionInfo(conn network.Connection, stream api.Stream) {
	remoteKey := "source"
	connectionKey := "connection"
	if stream.Direction() == api.ListenerDirectionOutbound {
		remoteKey = "upstream"
		connectionKey = "upstream"
	}

	connection := map[string]interface{}{}
	connection["mtls"] = conn.GetPeerCertificate() != nil
	if peerCert := conn.GetPeerCertificate(); peerCert != nil {
		connection["subject_peer_certificate"] = peerCert.GetSubject()
		connection["dns_san_peer_certificate"] = peerCert.GetFirstDNSName()
		connection["uri_san_peer_certificate"] = peerCert.GetFirstURI()
		connection["sha256_peer_certificate_digest"] = peerCert.GetSHA256Digest()
	}

	if localCert := conn.GetLocalCertificate(); localCert != nil {
		connection["subject_local_certificate"] = localCert.GetSubject()
		connection["dns_san_local_certificate"] = localCert.GetFirstDNSName()
		connection["uri_san_local_certificate"] = localCert.GetFirstURI()
	}

	if cs := conn.GetConnectionState(); cs != nil {
		connection["requested_server_name"] = cs.ServerName
		connection["tls_version"] = cs.Version
		connection["negotiated_protocol"] = cs.NegotiatedProtocol
	}

	stream.Logger().V(3).Info("set connection key", "key", connectionKey, "value", connection)
	stream.Set(connectionKey, connection)

	if ip, port, err := net.SplitHostPort(conn.RemoteAddr().String()); err == nil {
		stream.Set(remoteKey+".address", ip)
		if port, err := strconv.Atoi(port); err == nil {
			stream.Set(remoteKey+".port", port)
		}
		stream.Set("upstream.address", ip)
		if port, err := strconv.Atoi(port); err == nil {
			stream.Set("upstream.port", port)
		}
	}

	if ip, port, err := net.SplitHostPort(conn.LocalAddr().String()); err == nil {
		stream.Set("destination.address", ip)
		if port, err := strconv.Atoi(port); err == nil {
			stream.Set("destination.port", port)
		}
	}
}

func (m *envoyHttpHandlerMiddleware) setRequestProperties(req api.HTTPRequest, stream api.Stream, ttfb time.Time) {
	headers := proxywasm.NewHeaders(req.Header(), stream.Logger())
	trailers := proxywasm.NewHeaders(req.Trailer(), stream.Logger())

	size := 0
	if val, found := req.Header().Get("content-length"); found {
		if s, err := strconv.Atoi(val); err != nil {
			size = s
		}
	}

	requestInfo := map[string]interface{}{
		"path":     req.URL().RequestURI(),
		"url_path": req.URL().Path,
		"host":     req.Host(),
		"scheme":   req.URL().Scheme,
		"method":   req.Method(),
		"headers":  headers.Flatten(),
		"trailers": trailers.Flatten(),
		"referer": func() string {
			value, _ := req.Header().Get("referer")
			return value
		},
		"useragent": func() string {
			value, _ := req.Header().Get("user-agent")
			return value
		},
		"time": ttfb.UnixNano(),
		"id": func() string {
			value, _ := req.Header().Get("x-request-id")
			return value
		},
		"protocol": req.HTTPProtocol(),
		"query":    req.URL().RawQuery,

		"size":       size,
		"total_size": uint64(size) + headers.ByteSize() + trailers.ByteSize(),
	}

	stream.Set("request", requestInfo)
}
