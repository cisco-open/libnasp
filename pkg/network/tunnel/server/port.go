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
)

type port struct {
	session *session

	targetPort  int
	servicePort int

	ctx  context.Context
	ctxc context.CancelFunc
}

func NewPort(session *session, servicePort, targetPort int) *port {
	port := &port{
		session: session,

		servicePort: servicePort,
		targetPort:  targetPort,
	}

	port.ctx, port.ctxc = context.WithCancel(context.Background())

	return port
}

func (p *port) Close() error {
	p.ctxc()

	return nil
}

func (p *port) Listen() error {
	address := fmt.Sprintf(":%d", p.servicePort)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer l.Close()

	p.session.server.logger.V(2).Info("listening on address", "address", address)

	for {
		select {
		case <-p.ctx.Done():
			p.session.server.logger.V(2).Info("stop listening on address", "address", address)

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
			if os.IsTimeout(err) {
				continue
			}

			if err != nil {
				p.session.server.logger.Error(err, "could not accept")
				continue
			}

			if err := p.session.RequestConn(p.targetPort, c); err != nil {
				p.session.server.logger.Error(err, "could not request connection")
				c.Close()
			}
		}
	}
}
