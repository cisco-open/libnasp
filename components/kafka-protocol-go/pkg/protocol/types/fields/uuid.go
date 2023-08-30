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
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types"

	"emperror.dev/errors"
)

type UUID [16]byte

func (u *UUID) SetZero() {
	*u = [16]byte{}
}

func (u *UUID) IsZero() bool {
	return *u == UUID{}
}

type Uuid struct {
	Context
}

func (f *Uuid) Read(buf *bytes.Reader, version int16, out *UUID) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	var mostSignificantBits int64
	var leastSignificantBits int64

	err := types.ReadInt64(buf, &mostSignificantBits)
	if err != nil {
		return errors.WrapIf(err, "couldn't read the most significant bits of UUID value")
	}

	err = types.ReadInt64(buf, &leastSignificantBits)
	if err != nil {
		return errors.WrapIf(err, "couldn't read the least significant bits of UUID value")
	}

	b := pools.GetByteSlice(0, 16)
	b = binary.BigEndian.AppendUint64(b, uint64(mostSignificantBits))
	b = binary.BigEndian.AppendUint64(b, uint64(leastSignificantBits))

	copy(out[:], b)
	pools.ReleaseByteSlice(b)

	return nil
}

// SizeInBytes returns the size of this field in bytes when it's serialized
func (f *Uuid) SizeInBytes(version int16, _ UUID) (int, error) {
	if !f.IsSupportedVersion(version) {
		return 0, nil
	}

	return 16, nil
}

func (f *Uuid) Write(w *typesbytes.SliceWriter, version int16, data UUID) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	n, err := w.Write(data[:])
	if err != nil {
		return errors.New("couldn't write string to byte buffer")
	}
	if n != len(data) {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(len(data)), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}
