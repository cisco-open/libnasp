//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package protocol

import (
	jsoniter "github.com/json-iterator/go"

	"emperror.dev/errors"
	"github.com/tailscale/hujson"

	"io/fs"
)

func LoadProtocolSpec(kafkaProtocolSpec fs.FS) (map[string]*MessageSpec, error) {
	msgSpecs := make(map[string]*MessageSpec)

	dirEntries, err := fs.ReadDir(kafkaProtocolSpec, ".")
	if err != nil {
		return nil, err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			subDir, err := fs.Sub(kafkaProtocolSpec, dirEntry.Name())
			if err != nil {
				return nil, err
			}

			specs, err := LoadProtocolSpec(subDir)
			if err != nil {
				return nil, err
			}

			for k, v := range specs {
				msgSpecs[k] = v
			}

			continue
		}

		msgSpecRaw, err := fs.ReadFile(kafkaProtocolSpec, dirEntry.Name())
		if err != nil {
			return nil, err
		}
		msgSpecJson, err := hujson.Minimize(msgSpecRaw)
		if err != nil {
			return nil, errors.WrapIff(err, "couldn't convert kafka message specification to standard JSON")
		}

		msgSpec, err := loadMessageSpec(msgSpecJson)
		if err != nil {
			return nil, errors.WrapIff(err, "couldn't parse kafka message specification from %q", dirEntry.Name())
		}

		msgSpecs[msgSpec.Name] = msgSpec
	}

	return msgSpecs, nil
}

func loadMessageSpec(specJson []byte) (*MessageSpec, error) {
	var msgSpec MessageSpec

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(specJson, &msgSpec); err != nil {
		return nil, err
	}

	return &msgSpec, nil
}
