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
	"sync"
	"sync/atomic"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
)

var (
	ErrListenerClosed        = errors.New("listener is closed")
	ErrListenerAlreadyExists = errors.New("listener already exists")
	ErrInvalidNilListener    = errors.New("nil listener is not allowed")
)

type MultiListener interface {
	net.Listener
	Addrs() []net.Addr

	AddListener(lis net.Listener) error
	AddListeners(lis ...net.Listener) error
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
	closed     atomic.Bool

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

func NewMultiListener(options ...MultiListenerOption) MultiListener {
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

	return lstn
}

func (l *multipleListeners) AddListeners(listeners ...net.Listener) error {
	if l.closed.Load() {
		return errors.WithStack(ErrListenerClosed)
	}

	var e error

	for _, lis := range listeners {
		if err := l.AddListener(lis); err != nil {
			e = errors.Append(e, errors.WithStack(err))
		}
	}

	return e
}

func (l *multipleListeners) AddListener(lis net.Listener) error {
	if l.closed.Load() {
		return errors.WithStack(ErrListenerClosed)
	}

	if lis == nil {
		return errors.WithStackIf(ErrInvalidNilListener)
	}

	id := l.getID(lis)

	if _, ok := l.listeners.Load(id); ok {
		return errors.WithStackIf(ErrListenerAlreadyExists)
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
	id := l.getID(lis)

	if v, loaded := l.listeners.LoadAndDelete(id); loaded {
		if lis, ok := v.(*listener); ok {
			l.logger.V(2).Info("removing listener", "id", id)
			lis.ctxc()
		}
	}
}

func (l *multipleListeners) Accept() (net.Conn, error) {
	if l.closed.Load() {
		return nil, errors.WithStack(ErrListenerClosed)
	}

	for {
		accept, ok := <-l.acceptChan
		if ok {
			l.logger.V(3).Info("connection on listener", "id", accept.listenerID, "address", accept.listener.Addr().String())

			return accept.conn, accept.err
		} else {
			return nil, errors.WithStack(ErrListenerClosed)
		}
	}
}

func (l *multipleListeners) Close() error {
	if l.closed.Load() {
		return errors.WithStack(ErrListenerClosed)
	}

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

	close(l.acceptChan)

	l.closed.Store(true)

	l.logger.V(2).Info("multi listener closed")

	return nil
}

func (l *multipleListeners) startListener(id string, lis *listener) {
	l.logger.V(2).Info("start listener", "id", id)
	defer l.logger.V(2).Info("listener stopped", "id", id)

	go func(lis *listener) {
		<-lis.ctx.Done()
		if err := lis.Listener.Close(); err != nil {
			l.logger.Error(err, "could not close listener")
		}
	}(lis)

	for {
		c, err := lis.Accept()
		// break on net close error
		if errors.Is(err, net.ErrClosed) {
			break
		}

		l.acceptChan <- accept{
			listenerID: id,
			listener:   lis,
			conn:       c,
			err:        err,
		}
	}
}

func (l *multipleListeners) getID(lis net.Listener) string {
	id := lis.Addr().String()
	if withIDGetter, ok := lis.(interface {
		ID() string
	}); ok {
		return withIDGetter.ID()
	}

	return id
}
