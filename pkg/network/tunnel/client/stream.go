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
	"io"
	"time"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
	"github.com/cisco-open/nasp/pkg/network/tunnel/common"
)

type ctrlStream struct {
	api.MessageHandler

	client *client
}

func NewControlStream(client *client, str io.ReadWriteCloser) api.ControlStream {
	mh := common.NewMessageHandler(str, client.logger)

	s := &ctrlStream{
		MessageHandler: mh,
		client:         client,
	}

	mh.AddMessageHandler(api.RequestPortMessageType, s.requestPortResponse)
	mh.AddMessageHandler(api.RequestConnectionMessageType, s.requestConnection)
	mh.AddMessageHandler(api.PingMessageType, s.ping)

	return s
}

func (s *ctrlStream) GetMetadata() map[string]string {
	return nil
}

func (s *ctrlStream) GetUser() api.User {
	return nil
}

func (s *ctrlStream) ping(msg api.Message, params ...any) error {
	s.client.logger.V(3).Info("ping arrived, send pong")

	_, _, err := api.SendMessage(s, api.PongMessageType, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *ctrlStream) requestConnection(msg api.Message, params ...any) error {
	var req api.ConnectionRequest
	if err := msg.Decode(&req); err != nil {
		return err
	}

	logger := s.client.logger.WithValues("portID", req.PortID, "connectionID", req.Identifier)

	logger.V(2).Info("incoming connection request")

	var mp *managedPort
	if v, ok := s.client.managedPorts.Load(req.PortID); ok {
		if p, ok := v.(*managedPort); ok {
			mp = p
		}
	}

	if mp == nil {
		return errors.WithStackIf(api.ErrInvalidPort)
	}

	conn, err := s.client.session.OpenTCPStream(req.Identifier)
	if err != nil {
		return errors.WrapIfWithDetails(err, "could not open tcp stream", "portID", req.PortID, "id", req.Identifier)
	}

	logger.V(2).Info("stream is successfully created")

	if conn, err := newconn(conn, "tcp", req.LocalAddress, req.RemoteAddress); err != nil {
		return errors.WrapIf(err, "could not create wrapped connection")
	} else {
		select {
		case mp.GetConnChannel() <- conn:
			logger.V(2).Info("stream is accepted")
		case <-time.After(s.client.acceptTimeout):
			logger.Error(api.ErrAcceptTimeout, "could not establish connection")
			if err := conn.Close(); err != nil {
				logger.Error(err, "error during closing stream")
			} else {
				logger.V(2).Info("stream is closed")
			}
		}
	}

	return nil
}

func (s *ctrlStream) requestPortResponse(msg api.Message, params ...any) error {
	var resp api.RequestPortResponse
	if err := msg.Decode(&resp); err != nil {
		return err
	}

	if v, ok := s.client.managedPorts.Load(resp.ID); ok {
		if mp, ok := v.(*managedPort); ok {
			if resp.AssignedPort == 0 {
				_ = mp.Close()

				s.client.managedPorts.Delete(resp.ID)

				return errors.NewWithDetails("could not assign port", "portID", resp.ID, "requestedPort", resp.RequestedPort)
			}

			mp.remoteAddress = resp.Address
			mp.initialized = true

			s.client.logger.Info("port assigned", "portID", resp.ID, "remoteAddress", resp.Address, "targetPort", mp.options.GetTargetPort())

			return nil
		}
	}

	return errors.WithStackIf(api.ErrInvalidPort)
}
