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
	"encoding/base64"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"

	"emperror.dev/errors"
)

type Bytes struct {
	Context
}

func (f *Bytes) Read(r *bytes.Reader, version int16, out *[]byte) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	if f.IsFlexibleVersion(version) {
		return f.readCompactBytes(r, version, out)
	}
	return f.readBytes(r, version, out)
}

// SizeInBytes returns the size of data in bytes when it's serialized
func (f *Bytes) SizeInBytes(version int16, data []byte) (int, error) {
	if !f.IsSupportedVersion(version) {
		return 0, nil
	}

	if data == nil && !f.IsNullableVersion(version) {
		return 0, errors.New("non-nullable bytes field was set to null")
	}

	if len(data) == 0 {
		var length int32
		if data == nil {
			length = -1
		}

		if f.IsFlexibleVersion(version) {
			return varint.Uint32Size(uint32(length + 1)), nil // bytes needed to serialize the value -1 or 0 as varint
		}

		return 4, nil // bytes needed to serialize the value -1 or 0 as int32
	}

	length := len(data)
	if f.IsFlexibleVersion(version) {
		if int64(length)+1 > math.MaxInt32 {
			return 0, errors.New(strings.Join([]string{"field of type bytes has invalid length:", strconv.Itoa(length)}, " "))
		}
	} else {
		if int64(length) > math.MaxInt32 {
			return 0, errors.New(strings.Join([]string{"field of type bytes has invalid length:", strconv.Itoa(length)}, " "))
		}
	}

	if f.IsFlexibleVersion(version) {
		length += varint.Uint32Size(uint32(length) + 1) // add bytes needed to serialize the length of data as varint
		return length, nil
	}

	length += 4 // add bytes needed to serialize the length of data as int32
	return length, nil
}

func (f *Bytes) Write(w *typesbytes.SliceWriter, version int16, data []byte) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	if data == nil && !f.IsNullableVersion(version) {
		return errors.New("non-nullable bytes field was set to null")
	}

	if len(data) == 0 {
		var length int32
		if data == nil {
			length = -1
		}

		var err error
		if f.IsFlexibleVersion(version) {
			err = varint.WriteUint32(w, uint32(length+1))
		} else {
			err = types.WriteInt32(w, length)
		}

		if err != nil {
			return errors.WrapIf(err, "couldn't write length of null bytes field")
		}
		return nil
	}

	length := len(data)
	if int64(length) > math.MaxInt32 {
		return errors.New(strings.Join([]string{"field of type bytes has invalid length:", strconv.Itoa(length)}, " "))
	}

	if f.IsFlexibleVersion(version) {
		return f.writeCompactBytes(w, data)
	}

	return f.writeBytes(w, data)
}

func (f *Bytes) readBytesValue(r *bytes.Reader, length int, version int16, out *[]byte) error {
	if length < 0 {
		if f.IsNullableVersion(version) {
			return nil
		}

		return errors.New("non-nullable bytes field was serialized as null")
	}
	if int64(length) > math.MaxInt32 {
		return errors.New(strings.Join([]string{"field of type bytes has invalid length:", strconv.Itoa(length)}, " "))
	}

	valBytes := pools.GetByteSlice(length, length)
	if length == 0 {
		*out = valBytes
		return nil
	}

	n, err := r.Read(valBytes)
	if err != nil {
		return errors.WrapIf(err, strings.Join([]string{"couldn't read", strconv.Itoa(length), "bytes holding bytes value"}, " "))
	}

	if n < length {
		return errors.WrapIf(io.ErrUnexpectedEOF, strings.Join([]string{"expected", strconv.Itoa(length), "bytes holding bytes value but got only", strconv.Itoa(n), "bytes"}, " "))
	}

	*out = valBytes

	return nil
}

func (f *Bytes) readBytes(r *bytes.Reader, version int16, out *[]byte) error {
	var length int32
	err := types.ReadInt32(r, &length)
	if err != nil {
		return errors.WrapIf(err, "unable to read bytes field length")
	}

	return f.readBytesValue(r, int(length), version, out)
}

func (f *Bytes) readCompactBytes(r *bytes.Reader, version int16, out *[]byte) error {
	length, err := varint.ReadUint32(r)
	if err != nil {
		return errors.WrapIf(err, "unable to read compact bytes field length")
	}

	return f.readBytesValue(r, int(length)-1, version, out)
}

func (f *Bytes) writeCompactBytes(w *typesbytes.SliceWriter, data []byte) error {
	length := len(data)
	err := varint.WriteUint32(w, uint32(length+1))
	if err != nil {
		return errors.WrapIf(err, "couldn't write bytes length to byte buffer")
	}
	if length == 0 {
		return nil
	}

	n, err := w.Write(data)
	if err != nil {
		return errors.New("couldn't write bytes to byte buffer")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func (f *Bytes) writeBytes(w *typesbytes.SliceWriter, data []byte) error {
	length := len(data)
	err := types.WriteInt32(w, int32(length))
	if err != nil {
		return errors.WrapIf(err, "couldn't write length to byte buffer")
	}
	if length == 0 {
		return nil
	}

	n, err := w.Write(data)
	if err != nil {
		return errors.New("couldn't write bytes to byte buffer")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func BytesMarshalJSON(key string, data []byte) ([]byte, error) {
	if data == nil {
		return []byte("\"" + key + "\": null"), nil
	}

	j := []byte("\"" + base64.StdEncoding.EncodeToString(data) + "\"")

	return bytes.Join([][]byte{[]byte("\"" + key + "\""), j}, []byte(": ")), nil
}
