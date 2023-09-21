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
	"io"
	"net"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"

	"github.com/cisco-open/libnasp/pkg/network/tunnel/api"
	"github.com/cisco-open/libnasp/pkg/network/tunnel/common"
)

type ctrlStream struct {
	api.MessageHandler

	session *session
	stream  io.ReadWriteCloser
	logger  logr.Logger

	lastPong time.Time

	keepaliveTimer *time.Timer

	user     api.User
	metadata map[string]string
}

func NewControlStream(session *session, str io.ReadWriteCloser, user api.User, metadata map[string]string) api.ControlStream {
	mh := common.NewMessageHandler(str, session.Logger())

	s := &ctrlStream{
		MessageHandler: mh,

		session: session,
		stream:  str,
		logger:  session.Logger(),

		user:     user,
		metadata: metadata,
	}

	mh.AddMessageHandler(api.RequestPortMessageType, s.requestPort)
	mh.AddMessageHandler(api.PongMessageType, s.pong)

	go func() {
		if err := s.keepalive(session.server.keepaliveTimeout); err != nil {
			s.logger.Error(err, "keepalive error")
		}
	}()

	return s
}

func (s *ctrlStream) GetMetadata() map[string]string {
	return s.metadata
}

func (s *ctrlStream) GetUser() api.User {
	return s.user
}

func (s *ctrlStream) Close() error {
	if s.keepaliveTimer != nil {
		stopped := s.keepaliveTimer.Stop()
		s.logger.V(3).Info("stop keepalive timer", "stopped", stopped)
	}

	return s.MessageHandler.Close()
}

func (s *ctrlStream) keepalive(timeout time.Duration) error {
	if timeout == 0 {
		return nil
	}

	if s.lastPong.IsZero() {
		s.lastPong = time.Now()
	}

	s.logger.V(3).Info("send ping")

	s.keepaliveTimer = time.AfterFunc(time.Second*1, func() {
		if err := s.keepalive(timeout); err != nil {
			s.logger.Error(err, "keepalive error")
		}
	})

	if time.Since(s.lastPong) > timeout {
		s.logger.Info("ponged out", "timeout", timeout, "lastPont", s.lastPong)
		if err := s.Close(); err != nil {
			s.logger.Error(err, "error during stream close")
		}

		return nil
	}

	_, _, err := api.SendMessage(s, api.PingMessageType, nil)

	return err
}

func (s *ctrlStream) pong(msg api.Message, params ...any) error {
	s.logger.V(3).Info("pong arrived")

	s.lastPong = time.Now()

	return nil
}

func (s *ctrlStream) requestPort(msg api.Message, params ...any) error {
	var req api.RequestPort
	if err := msg.Decode(&req); err != nil {
		return err
	}

	assignedPort, addPortErr := s.session.requestPort(req)
	if addPortErr != nil {
		addPortErr = errors.WrapIfWithDetails(addPortErr, "could not assign port", "portID", req.ID, "requestedPort", req.RequestedPort)
	}

	_, _, err := api.SendMessage(s.stream, api.RequestPortMessageType, &api.RequestPortResponse{
		Type:         "tcp",
		ID:           req.ID,
		Name:         req.Name,
		AssignedPort: assignedPort,
		Address: (&net.TCPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: assignedPort,
		}).String(),
		RequestedPort: req.RequestedPort,
	})
	if err != nil {
		return errors.WrapIf(err, "could not send addPortResponse message")
	}

	return addPortErr
}
