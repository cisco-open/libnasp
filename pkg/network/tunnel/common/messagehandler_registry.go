// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
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
	"sync"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
)

type messageHandlerRegistry struct {
	messageHandlers sync.Map
}

type MessageHandlerRegistry interface {
	AddMessageHandler(messageType api.MessageType, handler api.MessageHandlerFunc)
	HandleMessage(msg api.Message, params ...any) error
}

func NewMessageHandlerRegistry() MessageHandlerRegistry {
	return &messageHandlerRegistry{
		messageHandlers: sync.Map{},
	}
}

func (h *messageHandlerRegistry) AddMessageHandler(messageType api.MessageType, handler api.MessageHandlerFunc) {
	h.messageHandlers.Store(messageType, handler)
}

func (h *messageHandlerRegistry) HandleMessage(msg api.Message, params ...any) error {
	if v, ok := h.messageHandlers.Load(msg.Type); ok {
		if handler, ok := v.(api.MessageHandlerFunc); ok {
			return handler(msg, params...)
		}
	}

	return errors.WithDetails(api.ErrInvalidMessageType, "type", msg.Type)
}
