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
	"github.com/google/uuid"
	"golang.ngrok.com/muxado"

	"github.com/cisco-open/libnasp/pkg/network/tunnel/api"
)

type client struct {
	session api.Session

	managedPorts sync.Map

	connected bool

	serverAddress string
	logger        logr.Logger
	muxConfig     *muxado.Config
	dialer        api.Dialer
	backoff       backoff.BackOff
	metadata      map[string]string
	bearerToken   string
	acceptTimeout time.Duration
}

var (
	defaultBackoff       = backoff.NewConstantBackOff(time.Second * 5)
	defaultLogger        = logr.Discard()
	defaultDialer        = &net.Dialer{Timeout: time.Second * 10}
	defaultAcceptTimeout = time.Second * 10
)

type ClientOption func(*client)

func ClientWithAcceptTimeout(d time.Duration) ClientOption {
	return func(c *client) {
		c.acceptTimeout = d
	}
}

func ClientWithLogger(logger logr.Logger) ClientOption {
	return func(c *client) {
		c.logger = logger.WithName("bifrost")
	}
}

func ClientWithMuxadoConfig(config *muxado.Config) ClientOption {
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

func ClientWithMetadata(md map[string]string) ClientOption {
	return func(c *client) {
		c.metadata = md
	}
}

func ClientWithBearerToken(token string) ClientOption {
	return func(c *client) {
		c.bearerToken = token
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

	if c.acceptTimeout == 0 {
		c.acceptTimeout = defaultAcceptTimeout
	}

	return c
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
	c.logger.Info("connecting to server", "server", c.serverAddress)
	conn, err := c.dialer.DialContext(ctx, "tcp", c.serverAddress)
	if err != nil {
		return err
	}

	c.session = NewSession(c, muxado.Client(conn, c.muxConfig))
	defer c.session.Close()

	c.onConnected()
	defer c.onDisconnected()

	if err := c.session.Handle(); err != nil {
		c.logger.Error(err, "error during session handling")
	}

	return errors.WithStackIf(api.ErrConnectionClosed)
}

func (c *client) requestPort(id string, options api.ManagedPortOptions) error {
	c.logger.V(2).Info("request port", "portID", id, "name", options.GetName(), "requestedPort", options.GetRequestedPort())

	m, _, err := api.SendMessage(c.session.GetControlStream(), api.RequestPortMessageType, &api.RequestPort{
		Type:          "tcp",
		ID:            id,
		Name:          options.GetName(),
		RequestedPort: options.GetRequestedPort(),
		TargetPort:    options.GetTargetPort(),
	})

	return errors.WrapIfWithDetails(err, "could not send message", "type", m.Type)
}

func (c *client) GetTCPListener(options api.ManagedPortOptions) (net.Listener, error) {
	if options == nil {
		options = &ManagedPortOptions{}
	}

	id := uuid.NewString()

	if options.GetName() == "" {
		options.SetName(id)
	}

	mp := NewManagedPort(id, options)
	c.managedPorts.Store(id, mp)

	if c.connected && c.session.GetControlStream() != nil {
		if err := c.requestPort(id, options); err != nil {
			return nil, err
		}
	}

	return mp, nil
}

func (c *client) onConnected() {
	c.connected = true

	c.logger.Info("client connected", "server", c.serverAddress)
}

func (c *client) onControlStreamConnected() {
	c.logger.V(1).Info("control stream connected", "server", c.serverAddress)

	c.managedPorts.Range(func(key any, value any) bool {
		if mp, ok := value.(*managedPort); ok {
			if !mp.initialized {
				if err := c.requestPort(mp.id, mp.options); err != nil {
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

	c.logger.Info("client disconnected", "server", c.serverAddress)
}
