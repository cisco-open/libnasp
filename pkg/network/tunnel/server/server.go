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

package server

import (
	"context"
	"net"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/werbenhu/eventbus"
	"golang.ngrok.com/muxado"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
)

var (
	defaultLogger           = logr.Discard()
	defaultSessionTimeout   = time.Second * 10
	defaultKeepaliveTimeout = time.Second * 60
)

type server struct {
	listenAddress string

	sessions     sync.Map
	portProvider api.PortProvider

	logger              logr.Logger
	sessionTimeout      time.Duration
	keepaliveTimeout    time.Duration
	listenerWrapperFunc ListenerWrapperFunc

	muxConfig *muxado.Config

	eventBus      *eventbus.EventBus
	authenticator api.Authenticator
}

type ServerOption func(*server)

type ListenerWrapperFunc func(l net.Listener) (net.Listener, error)

func ServerWithLogger(logger logr.Logger) ServerOption {
	return func(s *server) {
		s.logger = logger
	}
}

func ServerWithPortProvider(provider api.PortProvider) ServerOption {
	return func(s *server) {
		s.portProvider = provider
	}
}

func ServerWithSessionTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.sessionTimeout = timeout
	}
}

func ServerWithKeepaliveTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.keepaliveTimeout = timeout
	}
}

func ServerWithListenerWrapperFunc(listenerWrapperFunc ListenerWrapperFunc) ServerOption {
	return func(srv *server) {
		srv.listenerWrapperFunc = listenerWrapperFunc
	}
}

func ServerWithEventSubscribe(topic string, handler any) ServerOption {
	return func(srv *server) {
		if err := srv.eventBus.Subscribe(topic, handler); err != nil {
			panic(err)
		}
	}
}

func ServerWithAuthenticator(auth api.Authenticator) ServerOption {
	return func(srv *server) {
		srv.authenticator = auth
	}
}

func ServerWithMuxadoConfig(config *muxado.Config) ServerOption {
	return func(s *server) {
		s.muxConfig = config
	}
}

func NewServer(listenAddress string, options ...ServerOption) (api.Server, error) {
	s := &server{
		listenAddress:    listenAddress,
		sessions:         sync.Map{},
		sessionTimeout:   defaultSessionTimeout,
		keepaliveTimeout: defaultKeepaliveTimeout,
		listenerWrapperFunc: func(l net.Listener) (net.Listener, error) {
			return l, nil
		},

		eventBus: eventbus.New(),
	}

	for _, option := range options {
		option(s)
	}

	if s.portProvider == nil {
		if pp, err := NewPortProvider(50000, 55000); err != nil {
			return nil, err
		} else {
			s.portProvider = pp
		}
	}

	if s.logger == (logr.Logger{}) {
		s.logger = defaultLogger
	}

	return s, nil
}

func (s *server) Start(ctx context.Context) error {
	s.logger.Info("start server", "address", s.listenAddress)
	defer s.logger.Info("server stopped")

	l, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return errors.WrapIf(err, "could not listen")
	}

	l, err = s.listenerWrapperFunc(l)
	if err != nil {
		return errors.WrapIf(err, "could not wrap listener")
	}

	go func() {
		<-ctx.Done()
		s.logger.Info("context cancelled: closing listener")
		if err := l.Close(); err != nil {
			s.logger.Error(err, "error during listener close")
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return errors.WrapIf(err, "could not accept")
		}

		go func() {
			if err := s.handleConn(ctx, conn); err != nil {
				s.logger.Error(err, "error during connection handling")
			}
		}()
	}
}

func (s *server) handleConn(ctx context.Context, conn net.Conn) error {
	session := NewSession(s, muxado.Server(conn, s.muxConfig))
	s.sessions.Store(conn.RemoteAddr().String(), session)
	session.Logger().V(1).Info("session opened")

	defer func() {
		if err := session.Close(); err != nil {
			session.Logger().Error(err, "error during session close")
		} else {
			session.Logger().V(1).Info("session closed")
		}
		s.sessions.Delete(conn.RemoteAddr().String())
	}()

	return session.Handle(ctx)
}
