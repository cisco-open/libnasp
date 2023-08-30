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
	"strconv"
	"strings"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/serialization"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

// BatchRecord provides setter/getter methods for kafka records https://kafka.apache.org/documentation/#record
type BatchRecord struct {
	key            []byte
	value          []byte
	headers        []RecordHeader
	timestampDelta int64
	offsetDelta    int32
	attributes     int8
}

func (r *BatchRecord) TimestampDelta() int64 {
	return r.timestampDelta
}

func (r *BatchRecord) SetTimestampDelta(timestampDelta int64) {
	r.timestampDelta = timestampDelta
}

func (r *BatchRecord) OffsetDelta() int32 {
	return r.offsetDelta
}

func (r *BatchRecord) SetOffsetDelta(offsetDelta int32) {
	r.offsetDelta = offsetDelta
}

func (r *BatchRecord) Key() []byte {
	return r.key
}

func (r *BatchRecord) SetKey(key []byte) {
	r.key = key
}

func (r *BatchRecord) Value() []byte {
	return r.value
}

func (r *BatchRecord) SetValue(value []byte) {
	r.value = value
}

func (r *BatchRecord) Headers() []RecordHeader {
	return r.headers
}

func (r *BatchRecord) SetHeaders(headers []RecordHeader) {
	r.headers = headers
}

func (r *BatchRecord) Equal(that *BatchRecord) bool {
	if !(r.timestampDelta == that.timestampDelta &&
		r.offsetDelta == that.offsetDelta &&
		bytes.Equal(r.key, that.key) &&
		bytes.Equal(r.value, that.value)) {
		return false
	}

	if len(r.headers) != len(that.headers) {
		return false
	}
	for i := range r.headers {
		if !r.headers[i].Equal(&that.headers[i]) {
			return false
		}
	}

	return true
}

func (r *BatchRecord) String() string {
	s, err := r.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (r *BatchRecord) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}

	var s [5][]byte
	s[0] = []byte("{\"timestampDelta\": " + strconv.FormatInt(r.timestampDelta, 10))
	s[1] = []byte("\"offsetDelta\": " + strconv.FormatInt(int64(r.offsetDelta), 10))
	if r.key == nil {
		s[2] = []byte("\"key\": null")
	} else {
		s[2] = []byte("\"key\": \"" + base64.StdEncoding.EncodeToString(r.key) + "\"")
	}
	if r.value == nil {
		s[3] = []byte("\"value\": null")
	} else {
		s[3] = []byte("\"value\": \"" + base64.StdEncoding.EncodeToString(r.value) + "\"")
	}
	//nolint:nestif
	if r.headers == nil {
		s[4] = []byte("\"headers\": null}")
	} else {
		h := make([][]byte, 0, len(r.headers))
		for i := range r.headers {
			j, err := r.headers[i].MarshalJSON()
			if err != nil {
				return nil, err
			}
			h = append(h, j)
		}

		var arr bytes.Buffer
		if _, err := arr.WriteString("\"headers\": "); err != nil {
			return nil, err
		}
		if err := arr.WriteByte('['); err != nil {
			return nil, err
		}
		if _, err := arr.Write(bytes.Join(h, []byte(", "))); err != nil {
			return nil, err
		}
		if _, err := arr.WriteString("]}"); err != nil {
			return nil, err
		}

		s[4] = arr.Bytes()
	}

	return bytes.Join(s[:], []byte(", ")), nil
}

func (r *BatchRecord) Read(rd *serialization.Reader) error {
	recordLength, err := varint.ReadInt32_(rd) // Batch record length unused
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record \"length\" field")
	}

	var attr int8
	err = types.ReadInt8_(rd, &attr) // Batch record attributes unused
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record \"attributes\" field")
	}

	timestampDelta, err := varint.ReadInt64_(rd)
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record \"timestampDelta\" field")
	}

	offsetDelta, err := varint.ReadInt32_(rd)
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record \"offsetDelta\" field")
	}

	var key []byte
	keyLength, err := varint.ReadInt32_(rd)
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record \"keyLength\" field")
	}

	//nolint:nestif
	if keyLength >= 0 {
		key = pools.GetByteSlice(int(keyLength), int(keyLength))
		if keyLength > 0 {
			n, err := rd.Read(key)
			if err != nil {
				return errors.WrapIf(err, "couldn't read batch record \"key\" field")
			}
			if n != int(keyLength) {
				return errors.WrapIf(err, strings.Join([]string{"expected to read batch record key of size", strconv.Itoa(int(keyLength)), "bytes but got", strconv.Itoa(n), "bytes"}, " "))
			}
		}
	}

	var value []byte
	valueLength, err := varint.ReadInt32_(rd)
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record \"valueLength\" field")
	}

	//nolint:nestif
	if valueLength >= 0 {
		value = pools.GetByteSlice(int(valueLength), int(valueLength))
		if valueLength > 0 {
			n, err := rd.Read(value)
			if err != nil {
				return errors.WrapIf(err, "couldn't read batch record \"value\" field")
			}
			if n != int(valueLength) {
				return errors.WrapIf(err, strings.Join([]string{"expected to read batch record value of size", strconv.Itoa(int(valueLength)), "bytes but got", strconv.Itoa(n), "bytes"}, " "))
			}
		}
	}

	numHeaders, err := varint.ReadInt32_(rd)
	if err != nil {
		return errors.WrapIf(err, "couldn't read batch record headers count")
	}

	headers := make([]RecordHeader, numHeaders)
	for i := range headers {
		err := headers[i].Read(rd)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't read batch record header item", strconv.Itoa(i)}, " "))
		}
	}

	r.attributes = attr
	r.timestampDelta = timestampDelta
	r.offsetDelta = offsetDelta
	r.key = key
	r.value = value
	r.headers = headers

	actualRecordLength := r.recordLength()
	if actualRecordLength != int(recordLength) {
		return errors.New(strings.Join([]string{"invalid record size: expected to read", strconv.Itoa(int(recordLength)), "bytes in record payload, but instead read", strconv.Itoa(actualRecordLength)}, " "))
	}

	return nil
}

func (r *BatchRecord) Write(w *serialization.Writer) error {
	var err error
	if err = varint.WriteInt32(w, int32(r.recordLength())); err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record \"length\" field")
	}

	var attr = []byte{byte(r.attributes)}
	if _, err = w.Write(attr); err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record \"attributes\" field")
	}

	if err = varint.WriteInt64(w, r.timestampDelta); err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record \"timestampDelta\" field")
	}

	if err = varint.WriteInt32(w, r.offsetDelta); err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record \"offsetDelta\" field")
	}

	if r.key == nil {
		err = varint.WriteInt32(w, -1)
	} else {
		err = varint.WriteInt32(w, int32(len(r.key)))
	}
	if err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record \"keyLength\" field")
	}

	if len(r.key) > 0 {
		n, err := w.Write(r.key)
		if err != nil {
			return errors.WrapIf(err, "couldn't serialize batch record \"key\" field")
		}
		if n != len(r.key) {
			return errors.WrapIf(err, strings.Join([]string{"expected to write batch record key of", strconv.Itoa(len(r.key)), "bytes size but wrote", strconv.Itoa(n), "bytes"}, " "))
		}
	}

	if r.value == nil {
		err = varint.WriteInt32(w, -1)
	} else {
		err = varint.WriteInt32(w, int32(len(r.value)))
	}
	if err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record \"valueLength\" field")
	}

	if len(r.value) > 0 {
		n, err := w.Write(r.value)
		if err != nil {
			return errors.WrapIf(err, "couldn't serialize batch record \"value\" field")
		}
		if n != len(r.value) {
			return errors.WrapIf(err, strings.Join([]string{"expected to write batch record value of", strconv.Itoa(len(r.value)), "bytes size but wrote", strconv.Itoa(n), "bytes"}, " "))
		}
	}

	if r.headers == nil {
		return errors.New("invalid null header key found in headers")
	}

	if err = varint.WriteInt32(w, int32(len(r.headers))); err != nil {
		return errors.WrapIf(err, "couldn't serialize batch record headers count")
	}

	for i := range r.headers {
		err = r.headers[i].Write(w)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't serialize record header item:", strconv.Itoa(i)}, " "))
		}
	}

	return nil
}

func (r *BatchRecord) SizeInBytes() int {
	// https://kafka.apache.org/documentation/#record
	length := r.recordLength()
	recordLengthLength := varint.Int32Size(int32(length))

	return recordLengthLength + length
}

func (r *BatchRecord) Release() {
	if r.key != nil {
		pools.ReleaseByteSlice(r.key)
	}

	if r.value != nil {
		pools.ReleaseByteSlice(r.value)
	}
	for i := range r.headers {
		r.headers[i].Release()
	}
}

func (r *BatchRecord) recordLength() int {
	// https://kafka.apache.org/documentation/#record
	attributesLength := 1
	timestampDeltaLength := varint.Int64Size(r.timestampDelta)
	offsetDeltaLength := varint.Int32Size(r.offsetDelta)
	keyLengthLength := varint.Int32Size(int32(len(r.key))) // bytes needed to store the len of key field
	if r.key == nil {
		keyLengthLength = varint.Int32Size(-1)
	}
	keyLength := len(r.key)
	valueLengthLength := varint.Int32Size(int32(len(r.value))) // bytes needed to store the len of value field
	if r.value == nil {
		valueLengthLength = varint.Int32Size(-1)
	}
	valueLength := len(r.value)
	numHeadersLength := varint.Int32Size(int32(len(r.headers)))

	headersLength := 0
	for i := range r.headers {
		headersLength += r.headers[i].SizeInBytes()
	}

	length := attributesLength + timestampDeltaLength + offsetDeltaLength + keyLengthLength + keyLength +
		valueLengthLength + valueLength + numHeadersLength + headersLength

	return length
}
