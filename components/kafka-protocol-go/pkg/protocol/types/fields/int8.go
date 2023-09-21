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

package fields

import (
	"bytes"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types"
)

type Int8 struct {
	Context
}

func (f *Int8) Write(w *typesbytes.SliceWriter, version int16, data int8) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	return types.WriteInt8(w, data)
}

// SizeInBytes returns the size of this field in bytes when it's serialized
func (f *Int8) SizeInBytes(version int16, _ int8) (int, error) {
	if !f.IsSupportedVersion(version) {
		return 0, nil
	}

	return 1, nil
}

func (f *Int8) Read(buf *bytes.Reader, version int16, out *int8) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	return types.ReadInt8(buf, out)
}
