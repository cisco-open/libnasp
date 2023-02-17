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

package listener

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
)

var ErrAlreadyExists = errors.New("listener already exists")

type MultiListener interface {
	net.Listener

	AddListener(lis net.Listener) error
	RemoveListener(lis net.Listener)
}

type accept struct {
	listenerID string
	listener   net.Listener
	conn       net.Conn
	err        error
}

type multipleListeners struct {
	listeners sync.Map

	acceptChan chan accept

	logger logr.Logger
}

type listener struct {
	net.Listener

	ctx  context.Context
	ctxc context.CancelFunc
}

type MultiListenerOption func(*multipleListeners)

func MultiListenerWithLogger(logger logr.Logger) MultiListenerOption {
	return func(l *multipleListeners) {
		l.logger = logger
	}
}

func NewListener(l net.Listener, options ...MultiListenerOption) MultiListener {
	lstn := &multipleListeners{
		listeners: sync.Map{},

		acceptChan: make(chan accept, 1),
	}

	for _, option := range options {
		option(lstn)
	}

	if lstn.logger == (logr.Logger{}) {
		lstn.logger = logr.Discard()
	}

	_ = lstn.AddListener(l)

	return lstn
}

func (l *multipleListeners) AddListener(lis net.Listener) error {
	id := lis.Addr().String()

	if _, ok := l.listeners.Load(id); ok {
		return errors.WithStackIf(ErrAlreadyExists)
	}

	lstn := &listener{
		Listener: lis,
	}
	lstn.ctx, lstn.ctxc = context.WithCancel(context.Background())

	l.listeners.Store(id, lstn)

	l.logger.V(2).Info("listener added", "id", id)

	go l.startListener(id, lstn)

	return nil
}

func (l *multipleListeners) RemoveListener(lis net.Listener) {
	id := lis.Addr().String()

	if v, loaded := l.listeners.LoadAndDelete(id); loaded {
		if lis, ok := v.(*listener); ok {
			l.logger.V(2).Info("removing listener", "id", id)
			lis.ctxc()
		}
	}
}

func (l *multipleListeners) Accept() (net.Conn, error) {
	for {
		select {
		case accept := <-l.acceptChan:
			l.logger.V(3).Info("connection on listener", "id", accept.listenerID, "address", accept.listener.Addr().String())

			return accept.conn, accept.err
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (l *multipleListeners) Close() (err error) {
	listeners := make([]net.Listener, 0)
	l.listeners.Range(func(k, v any) bool {
		if v, ok := v.(net.Listener); ok {
			listeners = append(listeners, v)
		}

		return true
	})

	for _, lis := range listeners {
		l.RemoveListener(lis)
	}

	return err
}

func (l *multipleListeners) startListener(id string, lis *listener) {
	l.logger.V(2).Info("start listener", "id", id)
	defer l.logger.V(2).Info("listener stopped", "id", id)

	for {
		select {
		case <-lis.ctx.Done():
			if err := lis.Listener.Close(); err != nil {
				l.logger.Error(err, "could not close listener")
			}

			return
		default:
			if lis, ok := lis.Listener.(interface {
				SetDeadline(t time.Time) error
			}); ok {
				err := lis.SetDeadline(time.Now().Add(time.Second * 1))
				if err != nil {
					l.logger.Error(err, "could not set deadline for listener", "id", id)
				}
			}

			c, err := lis.Accept()
			if os.IsTimeout(err) {
				continue
			}

			l.acceptChan <- accept{
				listenerID: id,
				listener:   lis,
				conn:       c,
				err:        err,
			}
		}
	}
}
