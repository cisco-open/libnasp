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
	"io"
	"strconv"
	"strings"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"

	"emperror.dev/errors"
)

var nullStringBytes = []byte("null")

type NullableString struct {
	value []byte
	isNil bool
}

func (s *NullableString) Value() string {
	return string(s.value)
}

func (s *NullableString) Bytes() []byte {
	return s.value
}

func (s *NullableString) SetValue(value string) {
	s.Release()

	switch value {
	case "null":
		s.isNil = true
	case "":
		s.isNil = false
	default:
		s.value = []byte(value)
		s.isNil = false
	}
}

func (s *NullableString) Equal(that *NullableString) bool {
	return bytes.Equal(s.value, that.value)
}

func (s *NullableString) Clear() {
	s.Release()
	s.isNil = true
}

func (s *NullableString) IsNil() bool {
	return s.isNil
}

func (s *NullableString) Release() {
	if s.value != nil {
		pools.ReleaseByteSlice(s.value)
		s.value = nil
	}
}

func (s *NullableString) String() string {
	return s.Value()
}

func (s *NullableString) MarshalJSON() ([]byte, error) {
	if s == nil || s.isNil {
		return nullStringBytes, nil
	}

	var b bytes.Buffer
	b.WriteString("\"")
	b.Write(s.value)
	b.WriteString("\"")

	return b.Bytes(), nil
}

type String struct {
	Context
}

func (s *String) Read(buf *bytes.Reader, version int16, out *NullableString) error {
	if !s.IsSupportedVersion(version) {
		return nil
	}

	if s.IsFlexibleVersion(version) {
		return s.readCompactString(buf, version, out)
	}
	return s.readString(buf, version, out)
}

// SizeInBytes returns the size of data in bytes when it's serialized
func (s *String) SizeInBytes(version int16, data NullableString) (int, error) {
	if !s.IsSupportedVersion(version) {
		return 0, nil
	}

	if data.IsNil() {
		if !s.IsNullableVersion(version) {
			return 0, errors.New("non-nullable string field was set to null")
		}

		if s.IsFlexibleVersion(version) {
			return varint.Uint32Size(0), nil // bytes needed to serialize the value 0 as varint
		}

		return 2, nil // bytes needed to serialize the value -1 as int16
	}

	length := len(data.value) // bytes needed to serialize data
	if length > 0x7fff {
		return 0, errors.New(strings.Join([]string{"field of type string has invalid length:", strconv.Itoa(length)}, " "))
	}

	if s.IsFlexibleVersion(version) {
		length += varint.Uint32Size(uint32(length) + 1) // add bytes needed to serialize the length of data as varint
		return length, nil
	}

	length += 2 // add bytes needed to serialize the length of data as int16
	return length, nil
}

func (s *String) Write(w *typesbytes.SliceWriter, version int16, data NullableString) error {
	if !s.IsSupportedVersion(version) {
		return nil
	}

	//nolint:nestif
	if data.IsNil() {
		if !s.IsNullableVersion(version) {
			return errors.New("non-nullable string field was set to null")
		}

		if s.IsFlexibleVersion(version) {
			err := varint.WriteUint32(w, 0)
			if err != nil {
				return errors.WrapIf(err, "couldn't write null string length")
			}
			return nil
		}

		err := types.WriteInt16(w, -1)
		if err != nil {
			return errors.WrapIf(err, "couldn't write null string length")
		}
		return nil
	}

	length := len(data.value)
	if length > 0x7fff {
		return errors.New(strings.Join([]string{"field of type string has invalid length:", strconv.Itoa(length)}, " "))
	}

	if s.IsFlexibleVersion(version) {
		return s.writeCompactString(w, data.value)
	}

	return s.writeString(w, data.value)
}

func (s *String) readStringValue(r *bytes.Reader, length int, version int16, out *NullableString) error {
	if length < 0 {
		if s.IsNullableVersion(version) {
			out.Clear()
			return nil
		}

		return errors.New("non-nullable string field was serialized as null")
	}
	if length > 0x7fff {
		return errors.New(strings.Join([]string{"field of type string has invalid length:", strconv.Itoa(length)}, " "))
	}

	out.value = nil
	out.isNil = false

	if length > 0 {
		out.value = pools.GetByteSlice(length, length)

		n, err := r.Read(out.value)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't read", strconv.Itoa(length), "bytes holding string value"}, " "))
		}

		if n < length {
			return errors.WrapIf(io.ErrUnexpectedEOF, strings.Join([]string{"expected", strconv.Itoa(length), "bytes holding string value but got only", strconv.Itoa(n), "bytes"}, " "))
		}

		if isNullString(out.value) {
			out.isNil = true
			return nil
		}
	}

	return nil
}

func (s *String) readString(r *bytes.Reader, version int16, out *NullableString) error {
	var length int16

	err := types.ReadInt16(r, &length)
	if err != nil {
		return errors.WrapIf(err, "unable to read string field length")
	}

	return s.readStringValue(r, int(length), version, out)
}

func (s *String) readCompactString(r *bytes.Reader, version int16, out *NullableString) error {
	length, err := varint.ReadUint32(r)
	if err != nil {
		return errors.WrapIf(err, "unable to read compact string field length")
	}

	return s.readStringValue(r, int(length)-1, version, out)
}

func (s *String) writeCompactString(w *typesbytes.SliceWriter, data []byte) error {
	length := len(data)
	err := varint.WriteUint32(w, uint32(length)+1)
	if err != nil {
		return errors.WrapIf(err, "couldn't write length to byte buffer")
	}
	if length == 0 {
		return nil
	}

	n, err := w.Write(data)
	if err != nil {
		return errors.New("couldn't write string to byte buffer")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func (s *String) writeString(w *typesbytes.SliceWriter, data []byte) error {
	length := len(data)
	err := types.WriteInt16(w, int16(length))
	if err != nil {
		return errors.WrapIf(err, "couldn't write length to byte buffer")
	}
	if length == 0 {
		return nil
	}

	n, err := w.Write(data)
	if err != nil {
		return errors.New("couldn't write string to byte buffer")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func isNullString(data []byte) bool {
	return bytes.Equal(data, nullStringBytes)
}
