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

package client

import (
	"encoding/json"
	"io"
	"net"

	"emperror.dev/errors"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/tunnel/api"
)

type session struct {
	session *smux.Session
	client  *client
	stream  api.ControlStream
}

func NewSession(client *client, sess *smux.Session) api.Session {
	return &session{
		session: sess,
		client:  client,
	}
}

func (s *session) Close() error {
	if s.stream != nil {
		if err := s.stream.Close(); err != nil {
			s.client.logger.Error(err, "could not gracefully close control stream")
		}
	}

	return errors.WrapIf(s.session.Close(), "could not gracefully close session")
}

func (s *session) Handle() error {
	if err := s.createControlStream(); err != nil {
		return errors.WrapIf(err, "could not create control stream")
	}

	s.client.onControlStreamConnected()

	return s.stream.Handle()
}

func (s *session) GetControlStream() io.ReadWriter {
	return s.stream
}

func (s *session) OpenTCPStream(port int, id string) (net.Conn, error) {
	stream, err := s.session.OpenStream()
	if err != nil {
		return nil, errors.WrapIf(err, "could not open stream")
	}

	if _, _, err := api.SendMessage(stream, api.OpenTCPStreamMessageType, api.OpenTCPStreamMessage{
		ID: id,
	}); err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not send message", "type", api.OpenTCPStreamMessageType)
	}

	var msg api.Message
	if err := json.NewDecoder(stream).Decode(&msg); err != nil {
		return nil, errors.WrapIf(err, "could not decode message")
	}

	switch msg.Type { //nolint:exhaustive
	case api.ErrInvalidStreamIDMessageType:
		stream.Close()

		return nil, errors.WithStack(api.ErrInvalidStreamID)
	case api.StreamOpenedResponseMessageType:
		return stream, nil
	default:
		return nil, errors.WithDetails(api.ErrInvalidMessageType, "type", msg.Type)
	}
}

func (s *session) createControlStream() error {
	rawStream, err := s.session.OpenStream()
	if err != nil {
		return errors.WrapIf(err, "could not open stream")
	}

	_, _, err = api.SendMessage(rawStream, api.OpenControlStreamMessageType, nil)
	if err != nil {
		return errors.WrapIfWithDetails(err, "could not send message", "type", api.OpenControlStreamMessageType)
	}

	var msg api.Message
	if err := json.NewDecoder(rawStream).Decode(&msg); err != nil {
		return errors.WrapIf(err, "could not decode message")
	}

	switch msg.Type { //nolint:exhaustive
	case api.StreamOpenedResponseMessageType:
	default:
		return errors.WithStackIf(api.ErrCtrlInvalidResponse)
	}

	s.stream = NewControlStream(s.client, rawStream)

	return nil
}
