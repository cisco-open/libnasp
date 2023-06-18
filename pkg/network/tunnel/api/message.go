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

package api

import (
	"encoding/json"
	"io"

	"emperror.dev/errors"
	"github.com/google/uuid"
)

type Message struct {
	ID   string      `json:"id,omitempty"`
	Type MessageType `json:"type,omitempty"`
	Data []byte      `json:"data,omitempty"`
}

func (m Message) Decode(v any) error {
	return errors.WrapIf(json.Unmarshal(m.Data, &v), "could not unmarshal message")
}

type MessageType string

const (
	OpenTCPStreamMessageType     MessageType = "openTCPStream"
	OpenControlStreamMessageType MessageType = "openControlStream"
	RequestPortMessageType       MessageType = "requestPort"
	PortRegisteredMessageType    MessageType = "portRegistered"
	PongMessageType              MessageType = "pong"
	PingMessageType              MessageType = "ping"
	RequestConnectionMessageType MessageType = "requestConnection"
)

type ErrorType string

const (
	ErrCtrlStreamAlreadyExistsErrorType ErrorType = "errCtrlStreamAlreadyExists"
	ErrInvalidStreamTypeErrorType       ErrorType = "errInvalidStreamType"
	ErrInvalidStreamIDErrorType         ErrorType = "errInvalidStreamID"
	ErrAuthenticationFailureErrorType   ErrorType = "errAuthenticationFailure"
)

type BaseReponse struct {
	ErrorType    ErrorType `json:"errorType,omitempty"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	Error        error     `json:"-"`
}

func (r *BaseReponse) SetError(t ErrorType, e error) {
	r.Error = e
	r.ErrorType = t
	r.ErrorMessage = e.Error()
}

type OpenTCPStreamRequest struct {
	ID string `json:"id,omitempty"`
}

type OpenTCPStreamResponse struct {
	BaseReponse `json:",inline"`
}
type OpenControlStreamRequest struct {
	Metadata    map[string]string `json:"metadata,omitempty"`
	BearerToken string            `json:"token,omitempty"`
}

type OpenControlStreamResponse struct {
	BaseReponse `json:",inline"`
}

type RequestPort struct {
	Type          string `json:"type,omitempty"`
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	RequestedPort int    `json:"requestedPort,omitempty"`
	TargetPort    int    `json:"targetPort,omitempty"`
}

type PortRegistered struct {
	Type         string `json:"tcp,omitempty"`
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	AssignedPort int    `json:"port,omitempty"`
	Address      string `json:"address,omitempty"`
}

type ConnectionRequest struct {
	PortID        string `json:"portID,omitempty"`
	Identifier    string `json:"identifier,omitempty"`
	RemoteAddress string `json:"remoteAddress,omitempty"`
	LocalAddress  string `json:"localAddress,omitempty"`
}

func CreateMessage(messageType MessageType, message interface{}) (Message, []byte, error) {
	return CreateMessageWithID(uuid.New().String(), messageType, message)
}

func CreateMessageWithID(id string, messageType MessageType, message interface{}) (Message, []byte, error) {
	msg := Message{
		ID:   id,
		Type: messageType,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return msg, nil, err
	}

	msg.Data = data

	j, err := json.Marshal(msg)
	if err != nil {
		return msg, nil, err
	}

	return msg, j, nil
}

func SendMessage(writeTo io.Writer, messageType MessageType, message interface{}) (Message, []byte, error) {
	msg, j, err := CreateMessage(messageType, message)
	if err != nil {
		return msg, nil, err
	}

	_, err = writeTo.Write(j)

	return msg, j, err
}
