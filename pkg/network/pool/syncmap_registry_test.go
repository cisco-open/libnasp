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
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"k8s.io/klog/v2"
)

var (
	id = "test"
)

func TestSyncMapRegistry_New(t *testing.T) {
	t.Parallel()

	r := NewSyncMapPoolRegistry()
	require.IsType(t, &syncMapPoolRegistry{}, r)
}

func TestSyncMapRegistry_NameOption(t *testing.T) {
	t.Parallel()

	name := "name"
	r := NewSyncMapPoolRegistry(SyncMapPoolRegistryWithName(name))
	require.Equal(t, name, r.(*syncMapPoolRegistry).name)
}

func TestSyncMapRegistry_LoggerOption(t *testing.T) {
	t.Parallel()

	r := NewSyncMapPoolRegistry()
	require.Equal(t, logr.Discard().V(3), r.(*syncMapPoolRegistry).logger)

	r = NewSyncMapPoolRegistry(SyncMapPoolRegistryWithLogger(logger))
	require.Equal(t, klog.Background().V(3), r.(*syncMapPoolRegistry).logger)
}

func TestSyncMapRegistry_CRUD(t *testing.T) {
	t.Parallel()

	r := NewSyncMapPoolRegistry()
	require.Equal(t, false, r.HasPool(id))

	require.Equal(t, 0, r.Len())

	_, err := r.GetPool(id)
	require.ErrorIs(t, ErrPoolNotFound, err)

	p, err := NewChannelPool(factory)
	require.Nil(t, err)

	r.AddPool(id, p)
	require.Equal(t, true, r.HasPool(id))

	require.Equal(t, 1, r.Len())

	rp, err := r.GetPool(id)
	require.Nil(t, err)
	require.EqualValues(t, p, rp)

	r.RemovePool(id)

	require.Equal(t, false, r.HasPool(id))

	_, err = r.GetPool(id)
	require.ErrorIs(t, ErrPoolNotFound, err)

	require.Equal(t, 0, r.Len())
}

func TestSyncMapRegistry_Close(t *testing.T) {
	t.Parallel()

	r := NewSyncMapPoolRegistry()

	p, err := NewChannelPool(factory)
	require.Nil(t, err)
	r.AddPool(id, p)

	p, err = NewChannelPool(factory)
	require.Nil(t, err)
	r.AddPool(id+"2", p)

	require.Equal(t, 2, r.Len())

	require.Equal(t, false, r.IsClosed())
	err = r.Close()
	require.Nil(t, err)
	require.Equal(t, true, r.IsClosed())
	require.Equal(t, 0, r.Len())
}

func TestSyncMapRegistry_PoolIdleTimeout(t *testing.T) {
	t.Parallel()

	r := NewSyncMapPoolRegistry()

	p, err := NewChannelPool(factory, ChannelPoolWithIdleTimeout(time.Millisecond*500))
	require.Nil(t, err)
	r.AddPool(id, p)

	p, err = NewChannelPool(factory, ChannelPoolWithIdleTimeout(time.Second))
	require.Nil(t, err)
	r.AddPool(id+"2", p)

	require.Equal(t, 2, r.Len())

	time.Sleep(time.Second * 3)

	require.Equal(t, 0, r.Len())
}
