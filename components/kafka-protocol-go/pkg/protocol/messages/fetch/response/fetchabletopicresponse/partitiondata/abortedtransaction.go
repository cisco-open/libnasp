// Code generated by kafka-protocol-go. DO NOT EDIT.

// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	     https://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package partitiondata

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var abortedTransactionProducerId = fields.Context{
	SpecName:                        "ProducerId",
	LowestSupportedVersion:          4,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      12,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  4,
	HighestSupportedNullableVersion: 32767,
}
var abortedTransactionFirstOffset = fields.Context{
	SpecName:                        "FirstOffset",
	LowestSupportedVersion:          4,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      12,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  4,
	HighestSupportedNullableVersion: 32767,
}

type AbortedTransaction struct {
	unknownTaggedFields []fields.RawTaggedField
	producerId          int64
	firstOffset         int64
	isNil               bool
}

func (o *AbortedTransaction) ProducerId() int64 {
	return o.producerId
}

func (o *AbortedTransaction) SetProducerId(val int64) {
	o.isNil = false
	o.producerId = val
}

func (o *AbortedTransaction) FirstOffset() int64 {
	return o.firstOffset
}

func (o *AbortedTransaction) SetFirstOffset(val int64) {
	o.isNil = false
	o.firstOffset = val
}

func (o *AbortedTransaction) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *AbortedTransaction) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *AbortedTransaction) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	producerIdField := fields.Int64{Context: abortedTransactionProducerId}
	if err := producerIdField.Read(buf, version, &o.producerId); err != nil {
		return errors.WrapIf(err, "couldn't set \"producerId\" field")
	}

	firstOffsetField := fields.Int64{Context: abortedTransactionFirstOffset}
	if err := firstOffsetField.Read(buf, version, &o.firstOffset); err != nil {
		return errors.WrapIf(err, "couldn't set \"firstOffset\" field")
	}

	// process tagged fields

	if version < AbortedTransactionLowestSupportedFlexVersion() || version > AbortedTransactionHighestSupportedFlexVersion() {
		// tagged fields are only supported by flexible versions
		o.isNil = false
		return nil
	}

	if buf.Len() == 0 {
		o.isNil = false
		return nil
	}

	rawTaggedFields, err := fields.ReadRawTaggedFields(buf)
	if err != nil {
		return err
	}

	o.unknownTaggedFields = rawTaggedFields

	o.isNil = false
	return nil
}

func (o *AbortedTransaction) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	producerIdField := fields.Int64{Context: abortedTransactionProducerId}
	if err := producerIdField.Write(buf, version, o.producerId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"producerId\" field")
	}
	firstOffsetField := fields.Int64{Context: abortedTransactionFirstOffset}
	if err := firstOffsetField.Write(buf, version, o.firstOffset); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"firstOffset\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < AbortedTransactionLowestSupportedFlexVersion() || version > AbortedTransactionHighestSupportedFlexVersion() {
		if numTaggedFields > 0 {
			return errors.New(strings.Join([]string{"tagged fields were set, but version", strconv.Itoa(int(version)), "of this message does not support them"}, " "))
		}

		return nil
	}

	rawTaggedFields := make([]fields.RawTaggedField, 0, numTaggedFields)
	rawTaggedFields = append(rawTaggedFields, o.unknownTaggedFields...)

	if err := fields.WriteRawTaggedFields(buf, rawTaggedFields); err != nil {
		return errors.WrapIf(err, "couldn't serialize tagged fields")
	}

	return nil
}

func (o *AbortedTransaction) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *AbortedTransaction) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.producerId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"producerId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.firstOffset); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"firstOffset\""), b}, []byte(": ")))
	}

	if b, err := fields.ArrayOfStructMarshalJSON("unknownTaggedFields", o.unknownTaggedFields); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}

	var b bytes.Buffer
	if err := b.WriteByte('{'); err != nil {
		return nil, err
	}
	if _, err := b.Write(bytes.Join(s, []byte(", "))); err != nil {
		return nil, err
	}
	if err := b.WriteByte('}'); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (o *AbortedTransaction) IsNil() bool {
	return o.isNil
}

func (o *AbortedTransaction) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *AbortedTransaction) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.producerId = 0
	o.firstOffset = 0

	o.isNil = false
}

func (o *AbortedTransaction) Equal(that *AbortedTransaction) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.producerId != that.producerId {
		return false
	}
	if o.firstOffset != that.firstOffset {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *AbortedTransaction) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	producerIdField := fields.Int64{Context: abortedTransactionProducerId}
	fieldSize, err = producerIdField.SizeInBytes(version, o.producerId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"producerId\" field")
	}
	size += fieldSize

	firstOffsetField := fields.Int64{Context: abortedTransactionFirstOffset}
	fieldSize, err = firstOffsetField.SizeInBytes(version, o.firstOffset)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"firstOffset\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < AbortedTransactionLowestSupportedFlexVersion() || version > AbortedTransactionHighestSupportedFlexVersion() {
		if numTaggedFields > 0 {
			return 0, errors.New(strings.Join([]string{"tagged fields were set, but version", strconv.Itoa(int(version)), "of this message does not support them"}, " "))
		}

		return size, nil
	}

	taggedFieldsSize := varint.Uint32Size(uint32(numTaggedFields)) // bytes for serializing the number of tagged fields

	for i := range o.unknownTaggedFields {
		length := len(o.unknownTaggedFields[i].Value())
		if int64(length) > 0xffffffff {
			return 0, errors.New(strings.Join([]string{"invalid field value length:", strconv.Itoa(length), ", tag:", strconv.Itoa(int(o.unknownTaggedFields[i].Tag()))}, " "))
		}
		taggedFieldsSize += varint.Uint32Size(o.unknownTaggedFields[i].Tag()) // bytes for serializing the tag of the unknown tag
		taggedFieldsSize += varint.Uint32Size(uint32(length))                 // bytes for serializing the length of the unknown tagged field
		taggedFieldsSize += length
	}

	size += taggedFieldsSize

	return size, nil
}

// Release releases the dynamically allocated fields of this object by returning then to object pools
func (o *AbortedTransaction) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *AbortedTransaction) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *AbortedTransaction) validateNonIgnorableFields(version int16) error {
	if !abortedTransactionProducerId.IsSupportedVersion(version) {
		if o.producerId != 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"producerId\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !abortedTransactionFirstOffset.IsSupportedVersion(version) {
		if o.firstOffset != 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"firstOffset\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func AbortedTransactionLowestSupportedVersion() int16 {
	return 4
}

func AbortedTransactionHighestSupportedVersion() int16 {
	return 32767
}

func AbortedTransactionLowestSupportedFlexVersion() int16 {
	return 12
}

func AbortedTransactionHighestSupportedFlexVersion() int16 {
	return 32767
}

func AbortedTransactionDefault() AbortedTransaction {
	var d AbortedTransaction
	d.SetDefault()

	return d
}
