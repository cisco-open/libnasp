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
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
)

type port struct {
	session *session

	req          api.RequestPort
	assignedPort int

	logger logr.Logger

	ctx  context.Context
	ctxc context.CancelFunc
}

func NewPort(session *session, req api.RequestPort, assignedPort int) *port {
	port := &port{
		session: session,

		req:          req,
		assignedPort: assignedPort,

		logger: session.Logger().WithValues("name", req.Name, "assignedPort", assignedPort),
	}

	port.ctx, port.ctxc = context.WithCancel(context.Background())

	return port
}

func (p *port) Close() error {
	p.ctxc()

	return nil
}

func (p *port) Listen() error {
	address := fmt.Sprintf(":%d", p.assignedPort)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer l.Close()

	p.logger.V(2).Info("listening on address", "address", address)
	if err := p.session.server.eventBus.Publish(string(api.PortListenEventName), api.PortListenEvent{
		PortData: api.PortData{
			Name:       p.req.Name,
			Address:    address,
			Port:       p.assignedPort,
			TargetPort: p.req.TargetPort,
			User:       p.session.ctrlStream.GetUser(),
			Metadata:   p.session.ctrlStream.GetMetadata(),
		},
		SessionID: p.session.id,
	}); err != nil {
		p.logger.Error(err, "could not publish message")
	}

	for {
		select {
		case <-p.ctx.Done():
			p.logger.V(2).Info("stop listening on address", "address", address)
			if err := p.session.server.eventBus.Publish(string(api.PortReleaseEventName), api.PortReleaseEvent{
				PortData: api.PortData{
					Name:       p.req.ID,
					Address:    address,
					Port:       p.assignedPort,
					TargetPort: p.req.TargetPort,
					User:       p.session.ctrlStream.GetUser(),
					Metadata:   p.session.ctrlStream.GetMetadata(),
				},
				SessionID: p.session.id,
			}); err != nil {
				p.logger.Error(err, "could not publish message")
			}

			return nil
		default:
			if l, ok := l.(interface {
				SetDeadline(t time.Time) error
			}); ok {
				err := l.SetDeadline(time.Now().Add(time.Second * 1))
				if err != nil {
					return err
				}
			}

			c, err := l.Accept()
			if err != nil {
				if !os.IsTimeout(err) {
					p.logger.Error(err, "could not accept")
				}
				continue
			}

			if err := p.session.RequestConn(p.req.ID, c); err != nil {
				p.logger.Error(err, "could not request connection")
				if err := c.Close(); err != nil {
					p.logger.Error(err, "error during closing connection")
				}
			}
		}
	}
}
