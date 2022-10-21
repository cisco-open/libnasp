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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
)

type WasmVM interface {
	WasmRuntime
	Acquire(user interface{})
	Release(user interface{})
	Close()
}

type VMStore interface {
	GetOrCreateVM(vmConfig WasmVMConfig) (WasmVM, error)
}

type WasmVMConfig struct {
	ID            string                 `json:"id,omitempty"`
	Runtime       string                 `json:"runtime,omitempty"`
	Code          DataSource             `json:"code,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

type VMKey string

func (c *WasmVMConfig) GetVMKey() (VMKey, error) {
	code, err := c.Code.Get()
	if err != nil {
		return "", errors.WrapIf(err, "could not get wasm code")
	}

	config, err := json.Marshal(c.Configuration)
	if err != nil {
		return "", errors.WrapIf(err, "could not marshal vm configuration")
	}

	id := c.ID
	if id == "" {
		id = "vm-id-not-specified"
	}

	return VMKey(fmt.Sprintf("%s.%x.%x", id, sha256.Sum256(code), sha256.Sum256(config))), nil
}
