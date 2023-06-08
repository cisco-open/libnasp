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

package api

import (
	"encoding/json"

	"github.com/go-logr/logr"
)

type WasmPluginConfig struct {
	Name          string       `json:"name,omitempty"`
	RootID        string       `json:"rootID,omitempty"`
	VMConfig      WasmVMConfig `json:"vmConfig,omitempty"`
	Configuration Marshallable `json:"configuration,omitempty"`
	InstanceCount uint32       `json:"instanceCount,omitempty"`
}

func (c *WasmPluginConfig) Key() string {
	b, _ := json.Marshal(c) //nolint:errchkjson

	return string(b)
}

type WasmPlugin interface { //nolint:interfacebloat // we don't care
	Name() string
	ID() string
	GetInstance() (WasmInstance, error)
	ReleaseInstance(instance WasmInstance)
	EnsureInstances(desired uint32) uint32
	Exec(f func(instance WasmInstance) bool)
	Close()
	Report()
	VM() WasmVM
	Context() Context
	Logger() logr.Logger

	RegisterFilterContext(instance WasmInstance, filterContext FilterContext)
	UnregisterFilterContext(instance WasmInstance, filterContext FilterContext)
	GetFilterContext(instance WasmInstance, id int32) (FilterContext, bool)

	GetWasmInstanceContext(instance WasmInstance) (WasmInstanceContext, error)
}

type WasmPluginManager interface {
	Delete(config WasmPluginConfig) error
	Get(config WasmPluginConfig) (WasmPlugin, error)
	GetOrCreate(config WasmPluginConfig) (WasmPlugin, error)
	Add(config WasmPluginConfig) (WasmPlugin, error)
	Logger() logr.Logger
	GetBaseContext() Context
}
