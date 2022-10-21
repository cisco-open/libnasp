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

package dotn

import (
	"sync"

	"github.com/gohobby/deepcopy"
)

type ConcurrentMap struct {
	data  store
	mlock *sync.RWMutex
}

func New() *ConcurrentMap {
	return &ConcurrentMap{
		mlock: new(sync.RWMutex),
		data:  store{},
	}
}

func NewWithData(data map[string]interface{}) *ConcurrentMap {
	m := New()
	m.data = store(data)

	return m
}

func (mm *ConcurrentMap) Set(key string, value interface{}) {
	mm.mlock.Lock()
	mm.data.Set(key, value)
	mm.mlock.Unlock()
}

func (mm *ConcurrentMap) Reset(data store) {
	mm.mlock.Lock()
	mm.data = data
	mm.mlock.Unlock()
}

func (mm *ConcurrentMap) Get(key string) (interface{}, bool) {
	mm.mlock.RLock()
	r, ok := mm.data.Get(key)
	mm.mlock.RUnlock()
	return r, ok
}

func (mm *ConcurrentMap) Clone() *ConcurrentMap {
	return NewWithData(deepcopy.Map(mm.data.Convert()).Clone())
}
