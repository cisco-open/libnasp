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

	"github.com/google/uuid"
)

type Message struct {
	ID   string      `json:"id"`
	Type MessageType `json:"type"`
	Data []byte      `json:"data"`
}

func (m Message) Decode(v any) error {
	return json.Unmarshal(m.Data, &v)
}

type MessageType string

const (
	OpenTCPStreamMessageType        MessageType = "openTCPStream"
	OpenControlStreamMessageType    MessageType = "openControlStream"
	AddPortMessageType              MessageType = "addPort"
	AddPortResponseMessageType      MessageType = "addPortResponse"
	PongMessageType                 MessageType = "pong"
	PingMessageType                 MessageType = "ping"
	RequestConnectionMessageType    MessageType = "requestConnection"
	StreamOpenedResponseMessageType MessageType = "streamOpened"

	ErrCtrlStreamAlreadyExistsMessageType MessageType = "errCtrlStreamAlreadyExists"
	ErrInvalidStreamTypeMessageType       MessageType = "errInvalidStreamType"
	ErrInvalidStreamIDMessageType         MessageType = "errInvalidStreamID"
)

type OpenTCPStreamMessage struct {
	ID string `json:"id"`
}

type AddPortRequestMessage struct {
	Type string `json:"type"`
	Port int    `json:"port"`
}

type AddPortResponseMessage struct {
	Type    string `json:"tcp"`
	Port    int    `json:"port"`
	Address string `json:"address"`
}

type RequestConnectionMessage struct {
	Port       int    `json:"port"`
	Identifier string `json:"identifier"`
}

func CreateMessage(messageType MessageType, message interface{}) (Message, []byte, error) {
	return CreateMessageWithID(uuid.New().String(), messageType, message)
}

func CreateMessageWithID(id string, messageType MessageType, message interface{}) (Message, []byte, error) {
	msg := Message{
		ID:   uuid.New().String(),
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
