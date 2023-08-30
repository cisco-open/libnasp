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

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"

	"emperror.dev/errors"
)

// RecordHeader provides setter/getter methods for kafka record headers https://kafka.apache.org/documentation/#recordheader
type RecordHeader struct {
	key   []byte
	value []byte
}

func (h *RecordHeader) Key() string {
	return string(h.key)
}

func (h *RecordHeader) SetKey(key string) {
	if h.key != nil {
		pools.ReleaseByteSlice(h.key)
	}
	h.key = pools.GetByteSlice(len(key), len(key))
	copy(h.key, key)
}

func (h *RecordHeader) Value() []byte {
	return h.value
}

func (h *RecordHeader) SetValue(value []byte) {
	h.value = value
}

func (h *RecordHeader) String() string {
	s, err := h.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (h *RecordHeader) MarshalJSON() ([]byte, error) {
	if h == nil {
		return nullStringBytes, nil
	}

	var s [2][]byte
	var b bytes.Buffer

	b.WriteString("{\"key\": \"")
	b.Write(h.key)
	b.WriteString("\"")

	s[0] = b.Bytes()
	b.Reset()

	if h.value == nil {
		b.WriteString("\"value\": null}")
		s[1] = b.Bytes()
	} else {
		b.WriteString("\"value\": \"")
		b.WriteString(base64.StdEncoding.EncodeToString(h.value))
		b.WriteString("\"}")
		s[1] = b.Bytes()
	}

	return bytes.Join(s[:], []byte(", ")), nil
}

func (h *RecordHeader) Equal(that *RecordHeader) bool {
	return bytes.Equal(h.key, that.key) && bytes.Equal(h.value, that.value)
}

func (h *RecordHeader) Read(r *serialization.Reader) error {
	keyLength, err := varint.ReadInt32_(r)
	if err != nil {
		return errors.WrapIf(err, "couldn't read BatchRecord header \"headerKeyLength\" field")
	}

	var key []byte
	if keyLength > 0 {
		key = pools.GetByteSlice(int(keyLength), int(keyLength))
		n, err := r.Read(key)
		if err != nil {
			return errors.WrapIf(err, "couldn't read BatchRecord header \"headerKey\" field")
		}
		if n != int(keyLength) {
			return errors.WrapIf(err, strings.Join([]string{"expected to read BatchRecord header key of size", strconv.Itoa(int(keyLength)), "bytes but got", strconv.Itoa(n), "bytes"}, " "))
		}
	}

	var value []byte
	valueLength, err := varint.ReadInt32_(r)
	if err != nil {
		return errors.WrapIf(err, "couldn't read BatchRecord header \"headerValueLength\" field")
	}

	//nolint:nestif
	if valueLength >= 0 {
		value = pools.GetByteSlice(int(valueLength), int(valueLength))
		if valueLength > 0 {
			n, err := r.Read(value)
			if err != nil {
				return errors.WrapIf(err, "couldn't read BatchRecord header \"value\" field")
			}
			if n != int(valueLength) {
				return errors.WrapIf(err, strings.Join([]string{"expected to read BatchRecord header value of size", strconv.Itoa(int(valueLength)), "bytes but got", strconv.Itoa(n), "bytes"}, " "))
			}
		}
	}

	h.key = key
	h.value = value

	return nil
}

func (h *RecordHeader) Write(w *serialization.Writer) error {
	var err error
	if err = varint.WriteInt32(w, int32(len(h.key))); err != nil {
		return errors.WrapIf(err, "couldn't serialize BatchRecord header \"key\" field")
	}

	if len(h.key) > 0 {
		n, err := w.Write(h.key)
		if err != nil {
			return errors.WrapIf(err, "couldn't serialize BatchRecord header \"key\" field")
		}
		if n != len(h.key) {
			return errors.WrapIf(err, strings.Join([]string{"expected to write BatchRecord header key of", strconv.Itoa(len(h.key)), "bytes size but wrote", strconv.Itoa(n), "bytes"}, " "))
		}
	}

	if h.value == nil {
		err = varint.WriteInt32(w, -1)
	} else {
		err = varint.WriteInt32(w, int32(len(h.value)))
	}
	if err != nil {
		return errors.WrapIf(err, "couldn't serialize BatchRecord header \"valueLength\" field")
	}

	if len(h.value) > 0 {
		n, err := w.Write(h.value)
		if err != nil {
			return errors.WrapIf(err, "couldn't serialize BatchRecord header \"value\" field")
		}
		if n != len(h.value) {
			return errors.WrapIf(err, strings.Join([]string{"expected to write BatchRecord header value of", strconv.Itoa(len(h.value)), "bytes size but wrote", strconv.Itoa(n), "bytes"}, " "))
		}
	}
	return nil
}

func (h *RecordHeader) SizeInBytes() int {
	keyLengthLength := varint.Int32Size(int32(len(h.key)))
	keyLength := len(h.key)
	valueLengthLength := varint.Int32Size(int32(len(h.value)))
	if h.value == nil {
		valueLengthLength = varint.Int32Size(-1)
	}
	valueLength := len(h.value)

	return keyLengthLength + keyLength + valueLengthLength + valueLength
}

func (h *RecordHeader) Release() {
	if h.key != nil {
		pools.ReleaseByteSlice(h.key)
		h.key = nil
	}

	if h.value != nil {
		pools.ReleaseByteSlice(h.value)
		h.value = nil
	}
}
