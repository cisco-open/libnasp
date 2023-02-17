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
	"encoding/json"
	"net"
	"time"

	"emperror.dev/errors"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
	"github.com/cisco-open/nasp/pkg/network/tunnel/common"
)

type ctrlStream struct {
	api.ControlStream

	session *session
	stream  *smux.Stream

	lastPong time.Time

	keepaliveTimer *time.Timer
}

func NewControlStream(session *session, str *smux.Stream) api.ControlStream {
	cs := common.NewControlStream(str, session.server.logger)

	s := &ctrlStream{
		ControlStream: cs,

		session: session,
		stream:  str,
	}

	cs.AddMessageHandler(api.AddPortMessageType, s.addPort)
	cs.AddMessageHandler("pong", s.pong)

	go func() {
		if err := s.keepalive(session.server.keepaliveTimeout); err != nil {
			session.server.logger.Error(err, "keepalive error")
		}
	}()

	return s
}

func (s *ctrlStream) Close() error {
	if s.keepaliveTimer != nil {
		stopped := s.keepaliveTimer.Stop()
		s.session.server.logger.V(3).Info("stop keepalive timer", "stopped", stopped)
	}

	return s.ControlStream.Close()
}

func (s *ctrlStream) keepalive(timeout time.Duration) error {
	if timeout == 0 {
		return nil
	}

	if s.lastPong.IsZero() {
		s.lastPong = time.Now()
	}

	s.session.server.logger.V(3).Info("send ping")

	s.keepaliveTimer = time.AfterFunc(time.Second*1, func() {
		if err := s.keepalive(timeout); err != nil {
			s.session.server.logger.Error(err, "keepalive error")
		}
	})

	if time.Since(s.lastPong) > timeout {
		s.session.server.logger.Info("ponged out", "timeout", timeout, "lastPont", s.lastPong)
		s.Close()

		return nil
	}

	_, _, err := api.SendMessage(s, api.PingMessageType, nil)

	return err
}

func (s *ctrlStream) pong(msg []byte) error {
	s.session.server.logger.V(3).Info("pong arrived")

	s.lastPong = time.Now()

	return nil
}

func (s *ctrlStream) addPort(msg []byte) error {
	var a api.AddPortRequestMessage
	if err := json.Unmarshal(msg, &a); err != nil {
		return err
	}

	if a.Port > 0 {
		pp := s.session.AddPort(a.Port)

		_, _, err := api.SendMessage(s.stream, api.AddPortResponseMessageType, &api.AddPortResponseMessage{
			Type: "tcp",
			Port: a.Port,
			Address: (&net.TCPAddr{
				IP:   net.ParseIP("0.0.0.0"),
				Port: pp,
			}).String(),
		})
		if err != nil {
			return errors.WrapIf(err, "could not send addPortResponse message")
		}
	}

	return nil
}
