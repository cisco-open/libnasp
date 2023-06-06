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

	"github.com/cisco-open/nasp/pkg/network/listener"
)

type Server interface {
	GetUnifiedListener() listener.UnifiedListener
	Serve(l net.Listener) error
	ServeTLS(l net.Listener, certFile, keyFile string) error
	ServeWithTLSConfig(l net.Listener, config *tls.Config) error
	GetHTTPServer() *http.Server
}

type server struct {
	*http.Server

	unifiedListener listener.UnifiedListener
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
	srv.ConnContext = ConnectionStateToContextFromNetConn
	srv.ConnState = func(conn net.Conn, cstate http.ConnState) {
		if state, ok := ConnectionStateFromNetConn(conn); ok {
			if cstate != http.StateActive {
				state.ResetTimeToFirstByte()
			}
		}
	}

	return &server{
		Server: srv,
	}
}

func (s *server) GetHTTPServer() *http.Server {
	return s.Server
}

func (s *server) GetUnifiedListener() listener.UnifiedListener {
	return s.unifiedListener
}

func (s *server) Serve(l net.Listener) error {
	s.unifiedListener = listener.NewUnifiedListener(l, nil, listener.TLSModeDisabled,
		listener.UnifiedListenerWithTLSConnectionCreator(CreateTLSServerConn),
		listener.UnifiedListenerWithConnectionWrapper(func(c net.Conn) net.Conn {
			return WrapConnection(c)
		}),
	)

	return s.Server.Serve(s.unifiedListener)
}

func (s *server) ServeWithTLSConfig(l net.Listener, config *tls.Config) error {
	s.unifiedListener = listener.NewUnifiedListener(l, WrapTLSConfig(config), listener.TLSModePermissive,
		listener.UnifiedListenerWithTLSConnectionCreator(CreateTLSServerConn),
		listener.UnifiedListenerWithConnectionWrapper(func(c net.Conn) net.Conn {
			return WrapConnection(c)
		}),
	)

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
	ctx := NewConnectionStateHolderToContext(req.Context())

	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GotConn: func(i httptrace.GotConnInfo) {
			if connWithState, ok := i.Conn.(ConnectionStateGetter); ok {
				if h, ok := ConnectionStateHolderFromContext(ctx); ok {
					state := connWithState.GetConnectionState()
					state.ResetTimeToFirstByte()
					h.Set(state)
				}
			}
		},
	})

	req = req.WithContext(ctx)

	return t.RoundTripper.RoundTrip(req)
}
