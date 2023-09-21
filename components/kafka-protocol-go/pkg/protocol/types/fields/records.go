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
	"math"
	"strconv"
	"strings"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"

	"emperror.dev/errors"
)

type Records struct {
	Context
}

func (f *Records) Read(buf *bytes.Reader, version int16, out *RecordBatches) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	if f.IsFlexibleVersion(version) {
		return f.readCompactRecords(buf, version, out)
	}

	return f.readRecords(buf, version, out)
}

func (f *Records) SizeInBytes(version int16, data *RecordBatches) (int, error) {
	if !f.IsSupportedVersion(version) {
		return 0, nil
	}

	if data.IsNil() {
		if !f.IsNullableVersion(version) {
			return 0, errors.New("non-nullable records field was set to null")
		}

		if f.IsFlexibleVersion(version) {
			return varint.Uint32Size(0), nil // bytes needed to serialize the value 0 as varint
		}

		return 4, nil // bytes needed to serialize the value -1 as int32
	}

	dataLen := data.SizeInBytes()
	if f.IsFlexibleVersion(version) {
		if int64(dataLen)+1 > math.MaxUint32 {
			return 0, errors.New(strings.Join([]string{"field of type array has invalid length:", strconv.Itoa(dataLen)}, " "))
		}
	} else {
		if int64(dataLen) > math.MaxInt32 {
			return 0, errors.New(strings.Join([]string{"field of type array has invalid length:", strconv.Itoa(dataLen)}, " "))
		}
	}

	if f.IsFlexibleVersion(version) {
		dataLenLength := varint.Uint32Size(uint32(dataLen) + 1)
		return dataLenLength + dataLen, nil
	}

	dataLenLength := 4
	return dataLenLength + dataLen, nil
}

func (f *Records) Write(buf *typesbytes.SliceWriter, version int16, data *RecordBatches) error {
	if !f.IsSupportedVersion(version) {
		return nil
	}

	//nolint:nestif
	if data.IsNil() {
		if !f.IsNullableVersion(version) {
			return errors.New("non-nullable records field was set to null")
		}

		if f.IsFlexibleVersion(version) {
			err := varint.WriteUint32(buf, 0)
			if err != nil {
				return errors.WrapIf(err, "couldn't write length of null records field")
			}
			return nil
		}

		err := types.WriteInt32(buf, -1)
		if err != nil {
			return errors.WrapIf(err, "couldn't write length of null records field")
		}
		return nil
	}

	baseWrPos := buf.Len()
	err := data.Write(buf)
	if err != nil {
		return errors.WrapIf(err, "couldn't serialize record batches")
	}
	length := buf.Len() - baseWrPos

	if f.IsFlexibleVersion(version) {
		if int64(length)+1 > math.MaxUint32 {
			return errors.New(strings.Join([]string{"field of type array has invalid length:", strconv.Itoa(length)}, " "))
		}
	} else {
		if int64(length) > math.MaxInt32 {
			return errors.New(strings.Join([]string{"field of type array has invalid length:", strconv.Itoa(length)}, " "))
		}
	}

	//nolint: nestif
	if f.IsFlexibleVersion(version) {
		var b [5]byte
		n, err := varint.PutUint32(b[:], uint32(length)+1)
		if err != nil {
			return errors.WrapIf(err, "couldn't serialize length of records field")
		}

		m, err := buf.InsertAt(b[:n], int64(baseWrPos))
		if err != nil {
			return errors.WrapIf(err, "couldn't write record length")
		}
		if n != m {
			return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes of serialized records length, but wrote", strconv.Itoa(n), "bytes"}, " "))
		}
	} else {
		var b [4]byte

		binary.BigEndian.PutUint32(b[:], uint32(length))
		m, err := buf.InsertAt(b[:], int64(baseWrPos))
		if err != nil {
			return errors.WrapIf(err, "couldn't write record length")
		}
		if m != 4 {
			return errors.New(strings.Join([]string{"expected to write 4 bytes of serialized records length, but wrote", strconv.Itoa(m), "bytes"}, " "))
		}
	}

	return nil
}

func (f *Records) readRecords(buf *bytes.Reader, version int16, out *RecordBatches) error {
	var length int32
	err := types.ReadInt32(buf, &length)
	if err != nil {
		return errors.WrapIf(err, "unable to read records field length")
	}

	return f.readRecordsValue(buf, version, int(length), out)
}

func (f *Records) readCompactRecords(buf *bytes.Reader, version int16, out *RecordBatches) error {
	length, err := varint.ReadUint32(buf)
	if err != nil {
		return errors.WrapIf(err, "unable to read compact records field length")
	}

	return f.readRecordsValue(buf, version, int(length)-1, out)
}

// readRecordsValue reads the value of type "records" from the given byte buffer
// according to the serialization format described at https://kafka.apache.org/documentation/#messages
func (f *Records) readRecordsValue(buf *bytes.Reader, version int16, length int, out *RecordBatches) error {
	if length < 0 {
		if f.IsNullableVersion(version) {
			out.Clear()
			return nil
		}

		return errors.New("non-nullable records field was serialized as null")
	}

	if length == 0 {
		out.Complete()
		return nil
	}

	if buf.Len() < length {
		return errors.New(strings.Join([]string{"expected", strconv.Itoa(length), "bytes, but got only", strconv.Itoa(buf.Len())}, " "))
	}

	r := typesbytes.NewChunkReader(buf, int64(length))

	return out.Read(&r)
}
