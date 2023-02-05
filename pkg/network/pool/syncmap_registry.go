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

//nolint:forcetypeassert
package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
)

var (
	// ErrPoolNotFound is the error to indicate when a requested pool is not found.
	ErrPoolNotFound = errors.New("pool is not found")
)

type syncMapPoolRegistry struct {
	store  sync.Map
	count  uint64
	closed bool

	cleanupCheckDuration   time.Duration
	cleanupCheckCancelFunc context.CancelFunc

	name   string
	logger logr.Logger
}

// SyncMapPoolRegistryOption is a function for sync map pool registry options.
type SyncMapPoolRegistryOption func(*syncMapPoolRegistry)

// SyncMapPoolRegistryWithLogger sets the logger of the pool.
func SyncMapPoolRegistryWithLogger(logger logr.Logger) SyncMapPoolRegistryOption {
	return func(r *syncMapPoolRegistry) {
		r.logger = logger
	}
}

// SyncMapPoolRegistryWithName sets the name of the registry.
func SyncMapPoolRegistryWithName(name string) SyncMapPoolRegistryOption {
	return func(r *syncMapPoolRegistry) {
		r.name = name
	}
}

func NewSyncMapPoolRegistry(opts ...SyncMapPoolRegistryOption) Registry {
	r := &syncMapPoolRegistry{
		store:                sync.Map{},
		count:                0,
		cleanupCheckDuration: time.Second * 1,
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.logger == (logr.Logger{}) {
		r.logger = logr.Discard()
	}

	r.logger = r.logger.WithName(r.name).V(3)

	return r
}

func (r *syncMapPoolRegistry) GetPool(id string) (Pool, error) {
	val, ok := r.store.Load(id)
	if !ok {
		return nil, ErrPoolNotFound
	}

	return val.(Pool), nil
}

func (r *syncMapPoolRegistry) Close() (err error) {
	r.store.Range(func(key, value any) bool {
		err = value.(Pool).Close()
		if err != nil {
			return false
		}

		r.RemovePool(key.(string))

		return true
	})

	if err != nil {
		return err
	}

	r.closed = true

	return nil
}

func (r *syncMapPoolRegistry) IsClosed() bool {
	return r.closed
}

func (r *syncMapPoolRegistry) RemovePool(id string) {
	_, loaded := r.store.LoadAndDelete(id)
	if loaded {
		atomic.AddUint64(&r.count, ^uint64(0))

		if atomic.LoadUint64(&r.count) == 0 {
			r.stopCleanupCheck()
		}
	}
}

func (r *syncMapPoolRegistry) HasPool(id string) bool {
	_, ok := r.store.Load(id)

	return ok
}

func (r *syncMapPoolRegistry) AddPool(id string, pool Pool) {
	if pool.IsClosed() {
		return
	}

	_, loaded := r.store.LoadOrStore(id, pool)
	if !loaded {
		atomic.AddUint64(&r.count, 1)

		if atomic.LoadUint64(&r.count) == 1 {
			go r.startCleanupCheck()
		}
	}
}

func (r *syncMapPoolRegistry) Len() int {
	return int(atomic.LoadUint64(&r.count))
}

func (r *syncMapPoolRegistry) stopCleanupCheck() {
	if r.cleanupCheckCancelFunc == nil {
		return
	}

	r.cleanupCheckCancelFunc()
}

func (r *syncMapPoolRegistry) startCleanupCheck() {
	r.stopCleanupCheck()

	r.logger.Info("start pool cleanup check")

	var ctx context.Context
	ctx, r.cleanupCheckCancelFunc = context.WithCancel(context.Background())

	cleanup := func() {
		r.logger.Info("check closed pools", "poolCount", r.Len())
		r.store.Range(func(key, value any) bool {
			if pool, ok := value.(Pool); ok && pool.IsClosed() {
				r.logger.Info("remove closed pool", "id", key)
				r.RemovePool(key.(string))
			}

			return true
		})
	}

	t := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-t.C:
			cleanup()
		case <-ctx.Done():
			r.logger.Info("stop pool cleanup check")
			t.Stop()
			return
		}
	}
}
