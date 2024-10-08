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
package request

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var offsetFetchRequestTopicName = fields.Context{
	SpecName:                        "Name",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         7,
	LowestSupportedFlexVersion:      6,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  2,
	HighestSupportedNullableVersion: 7,
}
var offsetFetchRequestTopicPartitionIndexes = fields.Context{
	SpecName:                        "PartitionIndexes",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         7,
	LowestSupportedFlexVersion:      6,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  2,
	HighestSupportedNullableVersion: 7,
}

type OffsetFetchRequestTopic struct {
	partitionIndexes    []int32
	unknownTaggedFields []fields.RawTaggedField
	name                fields.NullableString
	isNil               bool
}

func (o *OffsetFetchRequestTopic) Name() fields.NullableString {
	return o.name
}

func (o *OffsetFetchRequestTopic) SetName(val fields.NullableString) {
	o.isNil = false
	o.name = val
}

func (o *OffsetFetchRequestTopic) PartitionIndexes() []int32 {
	return o.partitionIndexes
}

func (o *OffsetFetchRequestTopic) SetPartitionIndexes(val []int32) {
	o.isNil = false
	o.partitionIndexes = val
}

func (o *OffsetFetchRequestTopic) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *OffsetFetchRequestTopic) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *OffsetFetchRequestTopic) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	nameField := fields.String{Context: offsetFetchRequestTopicName}
	if err := nameField.Read(buf, version, &o.name); err != nil {
		return errors.WrapIf(err, "couldn't set \"name\" field")
	}

	partitionIndexesField := fields.Array[int32, *fields.Int32]{
		Context:          offsetFetchRequestTopicPartitionIndexes,
		ElementProcessor: &fields.Int32{Context: offsetFetchRequestTopicPartitionIndexes}}

	partitionIndexes, err := partitionIndexesField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndexes\" field")
	}
	o.partitionIndexes = partitionIndexes

	// process tagged fields

	if version < OffsetFetchRequestTopicLowestSupportedFlexVersion() || version > OffsetFetchRequestTopicHighestSupportedFlexVersion() {
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

func (o *OffsetFetchRequestTopic) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	nameField := fields.String{Context: offsetFetchRequestTopicName}
	if err := nameField.Write(buf, version, o.name); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"name\" field")
	}
	partitionIndexesField := fields.Array[int32, *fields.Int32]{
		Context:          offsetFetchRequestTopicPartitionIndexes,
		ElementProcessor: &fields.Int32{Context: offsetFetchRequestTopicPartitionIndexes}}
	if err := partitionIndexesField.Write(buf, version, o.partitionIndexes); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndexes\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < OffsetFetchRequestTopicLowestSupportedFlexVersion() || version > OffsetFetchRequestTopicHighestSupportedFlexVersion() {
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

func (o *OffsetFetchRequestTopic) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *OffsetFetchRequestTopic) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.name); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"name\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayMarshalJSON("partitionIndexes", o.partitionIndexes); err != nil {
		return nil, err
	} else {
		s = append(s, b)
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

func (o *OffsetFetchRequestTopic) IsNil() bool {
	return o.isNil
}

func (o *OffsetFetchRequestTopic) Clear() {
	o.Release()
	o.isNil = true

	o.partitionIndexes = nil
	o.unknownTaggedFields = nil
}

func (o *OffsetFetchRequestTopic) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.name.SetValue("")
	o.partitionIndexes = nil

	o.isNil = false
}

func (o *OffsetFetchRequestTopic) Equal(that *OffsetFetchRequestTopic) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.name.Equal(&that.name) {
		return false
	}
	if !fields.PrimitiveTypeSliceEqual(o.partitionIndexes, that.partitionIndexes) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *OffsetFetchRequestTopic) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	nameField := fields.String{Context: offsetFetchRequestTopicName}
	fieldSize, err = nameField.SizeInBytes(version, o.name)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"name\" field")
	}
	size += fieldSize

	partitionIndexesField := fields.Array[int32, *fields.Int32]{
		Context:          offsetFetchRequestTopicPartitionIndexes,
		ElementProcessor: &fields.Int32{Context: offsetFetchRequestTopicPartitionIndexes}}
	fieldSize, err = partitionIndexesField.SizeInBytes(version, o.partitionIndexes)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndexes\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < OffsetFetchRequestTopicLowestSupportedFlexVersion() || version > OffsetFetchRequestTopicHighestSupportedFlexVersion() {
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
func (o *OffsetFetchRequestTopic) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.name.Release()
	o.partitionIndexes = nil
}

func (o *OffsetFetchRequestTopic) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *OffsetFetchRequestTopic) validateNonIgnorableFields(version int16) error {
	return nil
}

func OffsetFetchRequestTopicLowestSupportedVersion() int16 {
	return 0
}

func OffsetFetchRequestTopicHighestSupportedVersion() int16 {
	return 7
}

func OffsetFetchRequestTopicLowestSupportedFlexVersion() int16 {
	return 6
}

func OffsetFetchRequestTopicHighestSupportedFlexVersion() int16 {
	return 32767
}

func OffsetFetchRequestTopicDefault() OffsetFetchRequestTopic {
	var d OffsetFetchRequestTopic
	d.SetDefault()

	return d
}
