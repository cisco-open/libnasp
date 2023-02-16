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

package common

import (
	"encoding/json"
	"io"
	"sync"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/tunnel/api"
)

type ctrlStream struct {
	io.ReadWriteCloser

	logger          logr.Logger
	messageHandlers sync.Map
}

func NewControlStream(str *smux.Stream, logger logr.Logger) api.ControlStream {
	s := &ctrlStream{
		ReadWriteCloser: str,

		logger:          logger,
		messageHandlers: sync.Map{},
	}

	return s
}

func (s *ctrlStream) AddMessageHandler(messageType api.MessageType, handler api.MessageHandler) {
	s.messageHandlers.Store(messageType, handler)
}

func (s *ctrlStream) Handle() error {
	s.logger.V(3).Info("handle control stream")

	var msg api.Message

	d := json.NewDecoder(s)
	for {
		if err := d.Decode(&msg); err != nil {
			return errors.WrapIf(err, "could not decode message")
		}

		if err := s.handleMessage(msg); err != nil {
			s.logger.Error(err, "could not handle message")
		}
	}
}

func (s *ctrlStream) handleMessage(msg api.Message) error {
	if v, ok := s.messageHandlers.Load(msg.Type); ok {
		if handler, ok := v.(api.MessageHandler); ok {
			return handler(msg.Data)
		}
	}

	return errors.WithDetails(api.ErrInvalidMessageType, "type", msg.Type)
}
