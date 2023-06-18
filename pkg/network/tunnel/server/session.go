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
	"encoding/json"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/pborman/uuid"
	"golang.ngrok.com/muxado"

	"github.com/cisco-open/nasp/pkg/network/proxy"
	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
	"github.com/cisco-open/nasp/pkg/network/tunnel/common"
)

type session struct {
	common.MessageHandlerRegistry

	server     *server
	session    muxado.Session
	ctrlStream api.ControlStream

	id     string
	logger logr.Logger

	ports             sync.Map
	backChannels      sync.Map
	activeConnections uint32

	ctx  context.Context
	ctxc context.CancelFunc
}

func NewSession(srv *server, sess muxado.Session) *session {
	id := uuid.New()

	s := &session{
		MessageHandlerRegistry: common.NewMessageHandlerRegistry(),

		server:       srv,
		session:      sess,
		ports:        sync.Map{},
		backChannels: sync.Map{},

		id:     id,
		logger: srv.logger.WithValues("remoteAddr", sess.RemoteAddr().String(), "sessionID", id),
	}

	s.AddMessageHandler(api.OpenControlStreamMessageType, s.openControlStream)
	s.AddMessageHandler(api.OpenTCPStreamMessageType, s.openTCPStream)

	return s
}

func (s *session) ID() string {
	return s.id
}

func (s *session) Logger() logr.Logger {
	return s.logger
}

func (s *session) Handle(ctx context.Context) error {
	s.ctx, s.ctxc = context.WithCancel(ctx)

	go func() {
		t := time.NewTicker(time.Second * 1)
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-t.C:
				s.logger.Info("active connections", "count", atomic.LoadUint32(&s.activeConnections))
			}
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			stream, err := s.session.Accept()
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				return nil
			}
			if err != nil {
				return errors.WrapIf(err, "could not accept stream")
			}

			go func() {
				if err := s.handleStream(stream); err != nil && !errors.Is(err, io.EOF) {
					s.logger.Error(err, "error during session handling")
				}
			}()
		}
	}
}

func (s *session) Close() error {
	if s.ctx == nil {
		return nil
	}

	s.logger.V(2).Info("close session")

	s.ports.Range(func(key, val any) bool {
		if mp, ok := val.(*port); ok {
			s.logger.V(2).Info("release port", "portID", mp.req.ID, "assignedPort", mp.assignedPort)
			s.server.portProvider.ReleasePort(mp.assignedPort)
			if err := mp.Close(); err != nil {
				s.logger.Error(err, "could not gracefully close managed port")
			}
			s.ports.Delete(key)
		}

		return true
	})

	if s.ctrlStream != nil {
		if err := s.ctrlStream.Close(); err != nil {
			s.logger.Error(err, "could not close control stream")
		}
	}

	s.ctxc()
	s.ctx = nil
	s.ctxc = nil

	return s.session.Close()
}

func (s *session) requestPort(req api.RequestPort) int {
	var assignedPort int

	if req.RequestedPort > 0 {
		if s.server.portProvider.GetPort(req.RequestedPort) {
			assignedPort = req.RequestedPort
		}
	} else {
		assignedPort = s.server.portProvider.GetFreePort()
	}

	if assignedPort == 0 {
		return 0
	}

	s.logger.V(2).Info("get free port", "portID", req.ID, "assignedPort", assignedPort)

	mp := NewPort(s, req, assignedPort)

	s.ports.Store(req.Name, mp)

	go func() {
		if err := mp.Listen(); err != nil {
			s.logger.Error(err, "could not listen")
		}
	}()

	return assignedPort
}

func (s *session) RequestConn(portID string, c net.Conn) error {
	id := uuid.NewUUID().String()

	logger := s.logger.WithValues("connectionID", id, "portID", portID, "remoteAddr", c.RemoteAddr().String())

	ch := make(chan io.ReadWriteCloser)
	s.backChannels.Store(id, ch)

	logger.V(2).Info("request connection")

	m, _, err := api.SendMessage(s.ctrlStream, api.RequestConnectionMessageType, &api.ConnectionRequest{
		PortID:        portID,
		Identifier:    id,
		RemoteAddress: c.RemoteAddr().String(),
		LocalAddress:  c.LocalAddr().String(),
	})
	if err != nil {
		s.backChannels.Delete(id)

		return errors.WrapIfWithDetails(err, "could not send message", "type", m.Type)
	}

	var tcpStream io.ReadWriteCloser

	select {
	case tcpStream = <-ch:
	case <-time.After(s.server.sessionTimeout):
		s.backChannels.Delete(id)

		return errors.WithStackIf(api.ErrSessionTimeout)
	}

	go func() {
		atomic.AddUint32(&s.activeConnections, 1)
		logger.V(2).Info("start proxying")
		defer func() {
			logger.V(2).Info("proxying stopped")
			logger.V(1).Info("stream closed")
			atomic.AddUint32(&s.activeConnections, ^uint32(0))
		}()

		proxy.New(c, tcpStream).Start()
	}()

	return err
}

func (s *session) handleStream(stream io.ReadWriteCloser) error {
	s.logger.V(3).Info("handle stream")

	var msg api.Message
	if err := json.NewDecoder(stream).Decode(&msg); err != nil {
		return err
	}

	if err := s.HandleMessage(msg, stream); err != nil {
		if s.ctrlStream == nil {
			if err := s.Close(); err != nil {
				s.logger.Error(err, "error during stream close")
			}
		}
		s.logger.Error(err, "error during message handling", append([]interface{}{"type", msg.Type}, errors.GetDetails(err)...)...)
	}

	return nil
}

func (s *session) openTCPStream(msg api.Message, params ...any) error {
	var m api.OpenTCPStreamRequest
	if err := msg.Decode(&m); err != nil {
		if err := s.Close(); err != nil {
			s.logger.Error(err, "error during stream close")
		}

		return errors.WrapIfWithDetails(err, "could not decode message", "type", api.OpenTCPStreamMessageType)
	}

	logger := s.logger.WithValues("connectionID", m.ID)

	logger.V(1).Info("open tcp stream")

	var stream io.ReadWriteCloser
	if p0, ok := params[0].(io.ReadWriteCloser); ok {
		stream = p0
	} else {
		return errors.WrapIf(api.ErrInvalidParam, "could not open tcp stream")
	}

	var ch chan io.ReadWriteCloser
	var ok bool

	if rch, found := s.backChannels.LoadAndDelete(m.ID); found {
		ch, ok = rch.(chan io.ReadWriteCloser)
	}

	response := api.OpenTCPStreamResponse{}

	if !ok {
		response.SetError(api.ErrInvalidStreamIDErrorType, api.ErrInvalidStreamID)

		if m, _, err := api.SendMessage(stream, api.OpenTCPStreamMessageType, response); err != nil {
			return errors.WrapIfWithDetails(err, "could not send response", "type", m.Type)
		}

		logger.Error(response.Error, "miafasz")

		return errors.WithStack(response.Error)
	}

	if m, _, err := api.SendMessage(stream, api.OpenTCPStreamMessageType, response); err != nil {
		return errors.WrapIfWithDetails(err, "could not send response", "type", m.Type)
	}

	ch <- stream

	logger.V(1).Info("tcp stream opened")

	return nil
}

func (s *session) openControlStream(msg api.Message, params ...any) error {
	var m api.OpenControlStreamRequest
	if err := msg.Decode(&m); err != nil {
		if err := s.Close(); err != nil {
			s.logger.Error(err, "error during stream close")
		}

		return errors.WrapIfWithDetails(err, "could not decode message", "type", api.OpenControlStreamMessageType)
	}

	s.logger.V(1).Info("open control stream")

	var stream io.ReadWriteCloser
	if p0, ok := params[0].(io.ReadWriteCloser); ok {
		stream = p0
	} else {
		return errors.WrapIf(api.ErrInvalidParam, "could not open control stream")
	}

	response := api.OpenControlStreamResponse{}

	if s.ctrlStream != nil {
		response.SetError(api.ErrCtrlStreamAlreadyExistsErrorType, api.ErrCtrlStreamAlreadyExists)
	}

	var authError error
	var user api.User
	if response.Error == nil && s.server.authenticator != nil {
		ok, u, err := s.server.authenticator.Authenticate(context.Background(), m.BearerToken)
		if err != nil {
			authError = errors.WrapIfWithDetails(err, "could not authenticate", "type", api.OpenControlStreamMessageType)
		} else if !ok {
			authError = errors.WrapIfWithDetails(err, "auth failed", "type", api.OpenControlStreamMessageType)
		}

		user = u
	}

	if authError != nil {
		response.SetError(api.ErrAuthenticationFailureErrorType, authError)
	}

	if m, _, err := api.SendMessage(stream, api.OpenControlStreamMessageType, response); err != nil {
		return errors.WrapIfWithDetails(err, "could not send message", "type", m.Type)
	}

	if response.Error != nil {
		if err := s.Close(); err != nil {
			s.logger.Error(err, "error during stream close")
		}

		return response.Error
	}

	s.logger.Info("control stream opened")

	s.ctrlStream = NewControlStream(s, stream, user, m.Metadata)

	return s.ctrlStream.Handle()
}
