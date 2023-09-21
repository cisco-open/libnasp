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
	"sort"
	"strconv"
	"strings"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/pools"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"

	"emperror.dev/errors"
)

func ReadRawTaggedFields(buf *bytes.Reader) ([]RawTaggedField, error) {
	numTaggedFields, err := varint.ReadUint32(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.WrapIf(err, "couldn't read the number of tagged fields")
	}
	if numTaggedFields == 0 {
		return nil, nil
	}
	taggedFields := make([]RawTaggedField, numTaggedFields)
	for i := range taggedFields {
		if taggedFields[i].value != nil {
			pools.ReleaseByteSlice(taggedFields[i].value)
		}
		tag, val, err := readTaggedField(buf)
		if err != nil {
			return nil, errors.WrapIf(err, strings.Join([]string{"expected to read", strconv.Itoa(int(numTaggedFields)), "tagged fields but got", strconv.Itoa(i)}, " "))
		}

		taggedFields[i].tag = tag
		taggedFields[i].value = val
	}

	return taggedFields, nil
}

func WriteRawTaggedFields(buf *typesbytes.SliceWriter, taggedFields []RawTaggedField) error {
	numTaggedFields := len(taggedFields)
	if int64(numTaggedFields) > math.MaxUint32 {
		return errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(numTaggedFields)}, " "))
	}
	if err := varint.WriteUint32(buf, uint32(numTaggedFields)); err != nil {
		return errors.WrapIf(err, "couldn't serialize tagged fields count")
	}

	if numTaggedFields == 0 {
		return nil
	}

	sort.Slice(taggedFields, func(i, j int) bool {
		return taggedFields[i].Tag() < taggedFields[j].Tag()
	})

	for i := 0; i < numTaggedFields-1; i++ {
		if taggedFields[i].Tag() == taggedFields[i+1].Tag() {
			return errors.New(strings.Join([]string{"duplicate tag", strconv.Itoa(int(taggedFields[i].Tag())), "found among tagged fields"}, " "))
		}
	}

	for i := range taggedFields {
		if err := writeTaggedField(buf, &taggedFields[i]); err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't write tagged field, tag:", strconv.Itoa(int(taggedFields[i].Tag()))}, " "))
		}
	}

	return nil
}

func readTaggedField(buf *bytes.Reader) (uint32, []byte, error) {
	tag, err := varint.ReadUint32(buf)
	if err != nil {
		return 0, nil, err
	}
	size, err := varint.ReadUint32(buf)
	if err != nil {
		return 0, nil, err
	}

	val := pools.GetByteSlice(int(size), int(size))
	n, err := buf.Read(val)
	if err != nil {
		return 0, nil, err
	}

	if n != int(size) {
		return 0, nil, errors.New(strings.Join([]string{"expected tagged field value length is", strconv.Itoa(int(size)), "bytes but got", strconv.Itoa(n), "bytes"}, " "))
	}

	return tag, val, nil
}

func writeTaggedField(w *typesbytes.SliceWriter, taggedField *RawTaggedField) error {
	if err := varint.WriteUint32(w, taggedField.Tag()); err != nil {
		return errors.WrapIf(err, strings.Join([]string{"couldn't write tag", strconv.Itoa(int(taggedField.Tag()))}, " "))
	}

	length := len(taggedField.Value())
	if int64(length) > math.MaxUint32 {
		return errors.New(strings.Join([]string{"invalid tagged field length:", strconv.Itoa(length)}, " "))
	}
	if err := varint.WriteUint32(w, uint32(length)); err != nil {
		return errors.WrapIf(err, strings.Join([]string{"couldn't write tagged field length", strconv.Itoa(length)}, " "))
	}

	if length == 0 {
		return nil
	}

	n, err := w.Write(taggedField.Value())
	if err != nil {
		return errors.New("couldn't write tagged field value")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

// RawTaggedField provides setter/getter methods for raw tagged fields
type RawTaggedField struct {
	value []byte
	tag   uint32
}

func (r *RawTaggedField) Tag() uint32 {
	return r.tag
}

func (r *RawTaggedField) SetTag(tag uint32) {
	r.tag = tag
}

func (r *RawTaggedField) Value() []byte {
	return r.value
}

func (r *RawTaggedField) SetValue(value []byte) {
	r.value = value
}

func (r *RawTaggedField) Equal(that *RawTaggedField) bool {
	return r.tag == that.tag && bytes.Equal(r.value, that.value)
}

func (r *RawTaggedField) String() string {
	s, err := r.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (r *RawTaggedField) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}

	var s [2][]byte
	s[0] = []byte("{\"tag\": " + strconv.FormatUint(uint64(r.tag), 10))

	if r.value == nil {
		s[1] = []byte("\"value\": null}")
	} else {
		s[1] = []byte("\"value\": \"" + base64.StdEncoding.EncodeToString(r.value) + "\"}")
	}

	return bytes.Join(s[:], []byte(", ")), nil
}

func (r *RawTaggedField) Release() {
	if r.value != nil {
		pools.ReleaseByteSlice(r.value)
		r.value = nil
	}
}
