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

package network

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/http/httptrace"

	ltls "github.com/cisco-open/nasp/pkg/tls"
)

type Server interface {
	GetUnifiedListener() ltls.UnifiedListener
	Serve(l net.Listener) error
	ServeTLS(l net.Listener, certFile, keyFile string) error
	ServeWithTLSConfig(l net.Listener, config *tls.Config) error
	GetHTTPServer() *http.Server
}

type server struct {
	*http.Server

	unifiedListener ltls.UnifiedListener
}

func WrapHTTPTransport(rt http.RoundTripper, dialer ConnectionDialer) http.RoundTripper {
	if t, ok := rt.(*http.Transport); ok {
		t.DialContext = dialer.DialContext
		t.DialTLSContext = dialer.DialTLSContext

		rt = t
	}

	return &transport{
		RoundTripper: rt,
	}
}

func WrapHTTPServer(srv *http.Server) Server {
	srv.ConnContext = WrappedConnectionToContext

	return &server{
		Server: srv,
	}
}

func (s *server) GetHTTPServer() *http.Server {
	return s.Server
}

func (s *server) GetUnifiedListener() ltls.UnifiedListener {
	return s.unifiedListener
}

func (s *server) Serve(l net.Listener) error {
	s.unifiedListener = ltls.NewUnifiedListener(NewWrappedListener(l), nil, ltls.TLSModeDisabled)

	return s.Server.Serve(s.unifiedListener)
}

func (s *server) ServeWithTLSConfig(l net.Listener, config *tls.Config) error {
	s.unifiedListener = ltls.NewUnifiedListener(NewWrappedListener(l), WrapTLSConfig(config), ltls.TLSModePermissive)

	return s.Server.Serve(s.unifiedListener)
}

func (s *server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{tlsCert},
		RootCAs:      certPool,
	}

	return s.ServeWithTLSConfig(l, tlsConfig)
}

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := NewConnectionToContext(req.Context())
	req = req.WithContext(httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GotConn: func(i httptrace.GotConnInfo) {
			SetConnectionToContextConnectionHolder(ctx, i.Conn)
		},
	}))

	return t.RoundTripper.RoundTrip(req)
}
