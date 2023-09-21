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

	"emperror.dev/errors"
	"github.com/go-logr/logr"

	"github.com/cisco-open/libnasp/pkg/network/tunnel/api"
)

type messageHandler struct {
	io.ReadWriteCloser
	MessageHandlerRegistry

	logger logr.Logger
}

func NewMessageHandler(str io.ReadWriteCloser, logger logr.Logger) api.MessageHandler {
	s := &messageHandler{
		ReadWriteCloser:        str,
		MessageHandlerRegistry: NewMessageHandlerRegistry(),

		logger: logger,
	}

	return s
}

func (s *messageHandler) Handle() error {
	s.logger.V(3).Info("handle control stream")

	var msg api.Message

	d := json.NewDecoder(s)
	for {
		if err := d.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return errors.WrapIf(err, "could not decode message")
		}

		go func(msg api.Message) {
			if err := s.HandleMessage(msg); err != nil {
				s.logger.Error(err, "error during message handling", append([]interface{}{"type", msg.Type}, errors.GetDetails(err)...)...)
			}
		}(msg)
	}
}
