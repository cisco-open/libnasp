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
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/listoffsets/request/listoffsetstopic"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var listOffsetsTopicName = fields.Context{
	SpecName:                    "Name",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var listOffsetsTopicPartitions = fields.Context{
	SpecName:                    "Partitions",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}

type ListOffsetsTopic struct {
	partitions          []listoffsetstopic.ListOffsetsPartition
	unknownTaggedFields []fields.RawTaggedField
	name                fields.NullableString
	isNil               bool
}

func (o *ListOffsetsTopic) Name() fields.NullableString {
	return o.name
}

func (o *ListOffsetsTopic) SetName(val fields.NullableString) {
	o.isNil = false
	o.name = val
}

func (o *ListOffsetsTopic) Partitions() []listoffsetstopic.ListOffsetsPartition {
	return o.partitions
}

func (o *ListOffsetsTopic) SetPartitions(val []listoffsetstopic.ListOffsetsPartition) {
	o.isNil = false
	o.partitions = val
}

func (o *ListOffsetsTopic) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *ListOffsetsTopic) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *ListOffsetsTopic) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	nameField := fields.String{Context: listOffsetsTopicName}
	if err := nameField.Read(buf, version, &o.name); err != nil {
		return errors.WrapIf(err, "couldn't set \"name\" field")
	}

	partitionsField := fields.ArrayOfStruct[listoffsetstopic.ListOffsetsPartition, *listoffsetstopic.ListOffsetsPartition]{Context: listOffsetsTopicPartitions}
	partitions, err := partitionsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"partitions\" field")
	}
	o.partitions = partitions

	// process tagged fields

	if version < ListOffsetsTopicLowestSupportedFlexVersion() || version > ListOffsetsTopicHighestSupportedFlexVersion() {
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

func (o *ListOffsetsTopic) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	nameField := fields.String{Context: listOffsetsTopicName}
	if err := nameField.Write(buf, version, o.name); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"name\" field")
	}

	partitionsField := fields.ArrayOfStruct[listoffsetstopic.ListOffsetsPartition, *listoffsetstopic.ListOffsetsPartition]{Context: listOffsetsTopicPartitions}
	if err := partitionsField.Write(buf, version, o.partitions); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitions\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ListOffsetsTopicLowestSupportedFlexVersion() || version > ListOffsetsTopicHighestSupportedFlexVersion() {
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

func (o *ListOffsetsTopic) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *ListOffsetsTopic) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.name); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"name\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("partitions", o.partitions); err != nil {
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

func (o *ListOffsetsTopic) IsNil() bool {
	return o.isNil
}

func (o *ListOffsetsTopic) Clear() {
	o.Release()
	o.isNil = true

	o.partitions = nil
	o.unknownTaggedFields = nil
}

func (o *ListOffsetsTopic) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.name.SetValue("")
	for i := range o.partitions {
		o.partitions[i].Release()
	}
	o.partitions = nil

	o.isNil = false
}

func (o *ListOffsetsTopic) Equal(that *ListOffsetsTopic) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.name.Equal(&that.name) {
		return false
	}
	if len(o.partitions) != len(that.partitions) {
		return false
	}
	for i := range o.partitions {
		if !o.partitions[i].Equal(&that.partitions[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *ListOffsetsTopic) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	nameField := fields.String{Context: listOffsetsTopicName}
	fieldSize, err = nameField.SizeInBytes(version, o.name)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"name\" field")
	}
	size += fieldSize

	partitionsField := fields.ArrayOfStruct[listoffsetstopic.ListOffsetsPartition, *listoffsetstopic.ListOffsetsPartition]{Context: listOffsetsTopicPartitions}
	fieldSize, err = partitionsField.SizeInBytes(version, o.partitions)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitions\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < ListOffsetsTopicLowestSupportedFlexVersion() || version > ListOffsetsTopicHighestSupportedFlexVersion() {
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
func (o *ListOffsetsTopic) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.name.Release()
	for i := range o.partitions {
		o.partitions[i].Release()
	}
	o.partitions = nil
}

func (o *ListOffsetsTopic) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *ListOffsetsTopic) validateNonIgnorableFields(version int16) error {
	return nil
}

func ListOffsetsTopicLowestSupportedVersion() int16 {
	return 0
}

func ListOffsetsTopicHighestSupportedVersion() int16 {
	return 32767
}

func ListOffsetsTopicLowestSupportedFlexVersion() int16 {
	return 6
}

func ListOffsetsTopicHighestSupportedFlexVersion() int16 {
	return 32767
}

func ListOffsetsTopicDefault() ListOffsetsTopic {
	var d ListOffsetsTopic
	d.SetDefault()

	return d
}
