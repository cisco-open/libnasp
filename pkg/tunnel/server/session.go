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
	"io"
	"net"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/pborman/uuid"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/network/proxy"
	"github.com/cisco-open/nasp/pkg/tunnel/api"
)

type session struct {
	server     *server
	session    *smux.Session
	ctrlStream api.ControlStream

	ports map[int]*port

	backChannels map[string]chan io.ReadWriteCloser

	mu  sync.Mutex
	mu2 sync.RWMutex
}

func NewSession(srv *server, sess *smux.Session) *session {
	return &session{
		server:       srv,
		session:      sess,
		ports:        make(map[int]*port),
		mu:           sync.Mutex{},
		backChannels: make(map[string]chan io.ReadWriteCloser),
	}
}

func (s *session) Handle() error {
	for {
		stream, err := s.session.AcceptStream()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errors.WrapIf(err, "could not accept stream")
		}

		go func() {
			if err := s.handleStream(stream); err != nil && !errors.Is(err, io.EOF) {
				s.server.logger.Error(err, "error during session handling")
			}
		}()
	}
}

func (s *session) Close() error {
	s.server.logger.V(2).Info("close session")

	s.mu.Lock()
	for p, mp := range s.ports {
		s.server.logger.V(2).Info("release port", "port", mp.servicePort)
		s.server.portProvider.ReleasePort(p)
		if err := mp.Close(); err != nil {
			s.server.logger.Error(err, "could not gracefully close managed port")
		}
		delete(s.ports, p)
	}
	s.mu.Unlock()

	if s.ctrlStream != nil {
		if err := s.ctrlStream.Close(); err != nil {
			s.server.logger.Error(err, "could not close control stream")
		}
	}

	return s.session.Close()
}

func (s *session) handleStream(stream *smux.Stream) error {
	s.server.logger.V(3).Info("handle stream", "id", stream.ID())

	var msg api.Message
	if err := json.NewDecoder(stream).Decode(&msg); err != nil {
		return err
	}

	switch msg.Type { //nolint:exhaustive
	case api.OpenControlStreamMessageType:
		if s.ctrlStream != nil {
			if _, _, err := api.SendMessage(stream, api.ErrCtrlStreamAlreadyExistsMessageType, nil); err != nil {
				return errors.WrapIfWithDetails(err, "could not send message", "type", api.ErrCtrlStreamAlreadyExistsMessageType)
			}

			return errors.New("control stream already established")
		}

		if _, _, err := api.SendMessage(stream, api.StreamOpenedResponseMessageType, nil); err != nil {
			return errors.WrapIfWithDetails(err, "could not send message", "type", api.StreamOpenedResponseMessageType)
		}

		s.ctrlStream = NewControlStream(s, stream)

		return s.ctrlStream.Handle()
	case api.OpenTCPStreamMessageType:
		var m api.OpenTCPStreamMessage
		if err := msg.Decode(&m); err != nil {
			return errors.WrapIfWithDetails(err, "could not decode message", "type", api.OpenTCPStreamMessageType)
		}

		var ch chan io.ReadWriteCloser
		var ok bool
		s.mu2.Lock()
		if ch, ok = s.backChannels[m.ID]; ok {
			delete(s.backChannels, m.ID)
		}
		s.mu2.Unlock()

		if !ok {
			if _, _, err := api.SendMessage(stream, api.ErrInvalidStreamIDMessageType, nil); err != nil {
				return errors.WrapIfWithDetails(err, "could not send message", "type", api.ErrInvalidStreamIDMessageType)
			}

			return errors.WithStack(api.ErrInvalidStreamID)
		}

		if _, _, err := api.SendMessage(stream, api.StreamOpenedResponseMessageType, nil); err != nil {
			return errors.WrapIfWithDetails(err, "could not send message", "type", api.StreamOpenedResponseMessageType)
		}

		ch <- stream

		return nil
	default:
		if _, _, err := api.SendMessage(stream, api.ErrInvalidStreamTypeMessageType, nil); err != nil {
			return errors.WrapIfWithDetails(err, "could not send message", "type", api.ErrInvalidStreamTypeMessageType)
		}

		return errors.WithStack(api.ErrInvalidStreamType)
	}
}

func (s *session) AddPort(port int) int {
	p := s.server.portProvider.GetFreePort()

	s.server.logger.V(2).Info("get free port", "targetPort", port, "servicePort", p)

	mp := NewPort(s, p, port)

	s.mu.Lock()
	s.ports[p] = mp
	s.mu.Unlock()

	go func() {
		if err := mp.Listen(); err != nil {
			s.server.logger.Error(err, "could not listen")
		}
	}()

	return p
}

func (s *session) RequestConn(port int, c net.Conn) error {
	id := uuid.NewUUID().String()
	_, _, err := api.SendMessage(s.ctrlStream, api.RequestConnectionMessageType, &api.RequestConnectionMessage{
		Port:       port,
		Identifier: id,
	})
	if err != nil {
		return errors.WrapIfWithDetails(err, "could not send message", "type", api.RequestConnectionMessageType)
	}

	ch := make(chan io.ReadWriteCloser, 1)

	s.mu2.Lock()
	s.backChannels[id] = ch
	s.mu2.Unlock()

	var tcpStream io.ReadWriteCloser

	select {
	case tcpStream = <-ch:
	case <-time.After(s.server.sessionTimeout):
		s.mu2.Lock()
		delete(s.backChannels, id)
		s.mu2.Unlock()

		return errors.WithStackIf(api.ErrSessionTimeout)
	}

	go proxy.New(c, tcpStream).Start()

	return err
}
