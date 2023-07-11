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
	"golang.ngrok.com/muxado"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
)

type session struct {
	session muxado.Session
	client  *client
	stream  api.ControlStream
}

func NewSession(client *client, sess muxado.Session) api.Session {
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

type buffered struct {
	net.Conn

	reader io.Reader
}

func (bd *buffered) Read(b []byte) (n int, err error) {
	return bd.reader.Read(b)
}

func (s *session) OpenTCPStream(id string) (net.Conn, error) {
	stream, err := s.session.OpenStream()
	if err != nil {
		return nil, errors.WrapIf(err, "could not open stream")
	}

	if _, _, err := api.SendMessage(stream, api.OpenTCPStreamMessageType, api.OpenTCPStreamRequest{
		ID: id,
	}); err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not send message", "type", api.OpenTCPStreamMessageType)
	}

	var msg api.Message
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&msg); err != nil {
		return nil, errors.WrapIf(err, "could not decode message")
	}

	if msg.Type == api.OpenTCPStreamMessageType {
		var resp api.OpenTCPStreamResponse
		if err := msg.Decode(&resp); err != nil {
			return nil, err
		}

		if resp.Error != nil {
			if err := stream.Close(); err != nil {
				s.client.logger.Error(err, "error during stream close")
			}

			return nil, errors.WithStack(resp.Error)
		}

		return &buffered{
			Conn:   stream,
			reader: io.MultiReader(decoder.Buffered(), stream),
		}, nil
	}

	return nil, errors.WithDetails(api.ErrInvalidMessageType, "type", msg.Type)
}

func (s *session) createControlStream() error {
	rawStream, err := s.session.OpenStream()
	if err != nil {
		return errors.WrapIf(err, "could not open stream")
	}

	if _, _, err = api.SendMessage(rawStream, api.OpenControlStreamMessageType, api.OpenControlStreamRequest{
		Metadata:    s.client.metadata,
		BearerToken: s.client.bearerToken,
	}); err != nil {
		return errors.WrapIfWithDetails(err, "could not send message", "type", api.OpenControlStreamMessageType)
	}

	var msg api.Message
	decoder := json.NewDecoder(rawStream)
	if err := decoder.Decode(&msg); err != nil {
		return errors.WrapIf(err, "could not decode message")
	}

	if msg.Type == api.OpenControlStreamMessageType {
		var resp api.OpenControlStreamResponse
		if err := msg.Decode(&resp); err != nil {
			return err
		}

		if resp.ErrorMessage != "" {
			return errors.New(resp.ErrorMessage)
		}

		s.stream = NewControlStream(s.client, &buffered{
			Conn:   rawStream,
			reader: io.MultiReader(decoder.Buffered(), rawStream),
		})

		return nil
	}

	return errors.WithStackIf(api.ErrCtrlInvalidResponse)
}
