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

package proxywasm

import (
	"sync"
	"sync/atomic"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/dotn"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

var contexts sync.Map

const (
	BaseContextID = "nasp-root"
)

type context struct {
	api.PropertyHolder

	id              int32
	latestContextID int32
	logger          logr.Logger
	parentContext   *context
}

func StoreContext(rootID string, ctx api.Context) {
	contexts.Store(rootID, ctx)
}

func GetBaseContext() api.Context {
	if ctx, found := GetContext(BaseContextID); found {
		return ctx
	}

	return NewContext(BaseContextID)
}

func GetContext(rootID string) (api.Context, bool) {
	if val, ok := contexts.Load(rootID); ok {
		if ctx, ok := val.(api.Context); ok {
			return ctx, true
		}
	}

	return nil, false
}

func DeleteContext(rootID string) {
	contexts.Delete(rootID)
}

func NewContext(rootID string) api.Context {
	c := &context{
		PropertyHolder: dotn.New(),

		id:              1,
		latestContextID: 1,
		logger:          klog.Background(),
	}

	StoreContext(rootID, c)

	return c
}

func (c *context) ID() int32 {
	return c.id
}

func (c *context) NewContextID() int32 {
	return atomic.AddInt32(&c.latestContextID, 1)
}

func (c *context) GetOrCreateContext(rootID string) api.Context {
	if ctx, found := GetContext(rootID); found {
		return ctx
	}

	return c.newContext(rootID)
}

func (c *context) newContext(rootID string) api.Context {
	ctx := NewContext(rootID)
	if ctxImpl, ok := ctx.(*context); ok {
		ctxImpl.parentContext = c
		ctxImpl.logger = c.logger
		ctxImpl.id = c.NewContextID()
		ctxImpl.latestContextID = ctx.ID()
		ctxImpl.PropertyHolder = NewPropertyHolderWrapper(dotn.New(), c.PropertyHolder)
	}

	return ctx
}
