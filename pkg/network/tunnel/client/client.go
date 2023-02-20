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

package client

import (
	"context"
	"net"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-logr/logr"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
)

type client struct {
	session api.Session

	managedPorts sync.Map

	connected bool

	serverAddress string
	logger        logr.Logger
	muxConfig     *smux.Config
	dialer        api.Dialer
	backoff       backoff.BackOff
}

var (
	defaultBackoff = backoff.NewConstantBackOff(time.Second * 5)
	defaultLogger  = logr.Discard()
	defaultDialer  = &net.Dialer{Timeout: time.Second * 10}
)

type ClientOption func(*client)

func ClientWithLogger(logger logr.Logger) ClientOption {
	return func(c *client) {
		c.logger = logger
	}
}

func ClientWithSmuxConfig(config *smux.Config) ClientOption {
	return func(c *client) {
		c.muxConfig = config
	}
}

func ClientWithDialer(dialer api.Dialer) ClientOption {
	return func(c *client) {
		c.dialer = dialer
	}
}

func ClientWithBackoff(backoff backoff.BackOff) ClientOption {
	return func(c *client) {
		c.backoff = backoff
	}
}

func NewClient(serverAddress string, options ...ClientOption) api.Client {
	c := &client{
		serverAddress: serverAddress,

		managedPorts: sync.Map{},
	}

	for _, option := range options {
		option(c)
	}

	if c.logger == (logr.Logger{}) {
		c.logger = defaultLogger
	}

	if c.dialer == nil {
		c.dialer = defaultDialer
	}

	if c.backoff == nil {
		c.backoff = defaultBackoff
	}

	return c
}

func (c *client) AddTCPPort(id string, requestedPort int) (net.Listener, error) {
	c.logger.V(2).Info("add port", "id", id, "requestedPort", requestedPort)

	return c.addOrGetManagedPort(id, requestedPort)
}

func (c *client) Connect(ctx context.Context) error {
	return backoff.Retry(func() error {
		err := c.connect(ctx)
		if err != nil {
			c.logger.Error(err, "server connection error")
		}

		return err
	}, c.backoff)
}

func (c *client) connect(ctx context.Context) error {
	c.logger.V(2).Info("connecting to server", "server", c.serverAddress)
	conn, err := c.dialer.DialContext(ctx, "tcp", c.serverAddress)
	if err != nil {
		return err
	}

	session, err := smux.Client(conn, c.muxConfig)
	if err != nil {
		return errors.WrapIf(err, "could not initialize client connection")
	}

	c.session = NewSession(c, session)
	defer c.session.Close()

	c.onConnected()
	defer c.onDisconnected()

	if err := c.session.Handle(); err != nil {
		c.logger.Error(err, "error during session handling")
	}

	return errors.WithStackIf(api.ErrConnectionClosed)
}

func (c *client) sendAddPortMessage(id string, requestedPort int) error {
	_, _, err := api.SendMessage(c.session.GetControlStream(), api.AddPortMessageType, &api.AddPortRequestMessage{
		Type:          "tcp",
		ID:            id,
		RequestedPort: requestedPort,
	})

	return errors.WrapIfWithDetails(err, "could not send message", "type", api.AddPortMessageType)
}

func (c *client) addOrGetManagedPort(id string, requestedPort int) (net.Listener, error) {
	var mp *managedPort

	v, _ := c.managedPorts.Load(id)
	if p, ok := v.(*managedPort); ok {
		mp = p
	}
	if mp == nil {
		mp = NewManagedPort(id, requestedPort)
		c.managedPorts.Store(id, mp)

		if c.connected {
			if err := c.sendAddPortMessage(id, requestedPort); err != nil {
				return nil, err
			}
		}
	}

	return mp, nil
}

func (c *client) onConnected() {
	c.connected = true

	c.logger.V(2).Info("client connected", "server", c.serverAddress)
}

func (c *client) onControlStreamConnected() {
	c.logger.V(2).Info("control stream connected", "server", c.serverAddress)

	c.managedPorts.Range(func(key any, value any) bool {
		if mp, ok := value.(*managedPort); ok {
			if !mp.initialized {
				if err := c.sendAddPortMessage(mp.id, mp.requestedPort); err != nil {
					c.logger.Error(err, "")
				}
			}
		}

		return true
	})
}

func (c *client) onDisconnected() {
	c.connected = false

	c.managedPorts.Range(func(key any, value any) bool {
		if mp, ok := value.(*managedPort); ok {
			mp.initialized = false
		}

		return true
	})

	c.logger.V(2).Info("client disconnected", "server", c.serverAddress)
}
