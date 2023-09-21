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
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/createpartitions/request/createpartitionstopic"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var createPartitionsTopicName = fields.Context{
	SpecName:                    "Name",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var createPartitionsTopicCount = fields.Context{
	SpecName:                    "Count",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var createPartitionsTopicAssignments = fields.Context{
	SpecName:                        "Assignments",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}

type CreatePartitionsTopic struct {
	assignments         []createpartitionstopic.CreatePartitionsAssignment
	unknownTaggedFields []fields.RawTaggedField
	name                fields.NullableString
	count               int32
	isNil               bool
}

func (o *CreatePartitionsTopic) Name() fields.NullableString {
	return o.name
}

func (o *CreatePartitionsTopic) SetName(val fields.NullableString) {
	o.isNil = false
	o.name = val
}

func (o *CreatePartitionsTopic) Count() int32 {
	return o.count
}

func (o *CreatePartitionsTopic) SetCount(val int32) {
	o.isNil = false
	o.count = val
}

func (o *CreatePartitionsTopic) Assignments() []createpartitionstopic.CreatePartitionsAssignment {
	return o.assignments
}

func (o *CreatePartitionsTopic) SetAssignments(val []createpartitionstopic.CreatePartitionsAssignment) {
	o.isNil = false
	o.assignments = val
}

func (o *CreatePartitionsTopic) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *CreatePartitionsTopic) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *CreatePartitionsTopic) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	nameField := fields.String{Context: createPartitionsTopicName}
	if err := nameField.Read(buf, version, &o.name); err != nil {
		return errors.WrapIf(err, "couldn't set \"name\" field")
	}

	countField := fields.Int32{Context: createPartitionsTopicCount}
	if err := countField.Read(buf, version, &o.count); err != nil {
		return errors.WrapIf(err, "couldn't set \"count\" field")
	}

	assignmentsField := fields.ArrayOfStruct[createpartitionstopic.CreatePartitionsAssignment, *createpartitionstopic.CreatePartitionsAssignment]{Context: createPartitionsTopicAssignments}
	assignments, err := assignmentsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"assignments\" field")
	}
	o.assignments = assignments

	// process tagged fields

	if version < CreatePartitionsTopicLowestSupportedFlexVersion() || version > CreatePartitionsTopicHighestSupportedFlexVersion() {
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

func (o *CreatePartitionsTopic) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	nameField := fields.String{Context: createPartitionsTopicName}
	if err := nameField.Write(buf, version, o.name); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"name\" field")
	}
	countField := fields.Int32{Context: createPartitionsTopicCount}
	if err := countField.Write(buf, version, o.count); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"count\" field")
	}

	assignmentsField := fields.ArrayOfStruct[createpartitionstopic.CreatePartitionsAssignment, *createpartitionstopic.CreatePartitionsAssignment]{Context: createPartitionsTopicAssignments}
	if err := assignmentsField.Write(buf, version, o.assignments); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"assignments\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < CreatePartitionsTopicLowestSupportedFlexVersion() || version > CreatePartitionsTopicHighestSupportedFlexVersion() {
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

func (o *CreatePartitionsTopic) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *CreatePartitionsTopic) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 4)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.name); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"name\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.count); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"count\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("assignments", o.assignments); err != nil {
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

func (o *CreatePartitionsTopic) IsNil() bool {
	return o.isNil
}

func (o *CreatePartitionsTopic) Clear() {
	o.Release()
	o.isNil = true

	o.assignments = nil
	o.unknownTaggedFields = nil
}

func (o *CreatePartitionsTopic) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.name.SetValue("")
	o.count = 0
	for i := range o.assignments {
		o.assignments[i].Release()
	}
	o.assignments = nil

	o.isNil = false
}

func (o *CreatePartitionsTopic) Equal(that *CreatePartitionsTopic) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.name.Equal(&that.name) {
		return false
	}
	if o.count != that.count {
		return false
	}
	if len(o.assignments) != len(that.assignments) {
		return false
	}
	for i := range o.assignments {
		if !o.assignments[i].Equal(&that.assignments[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *CreatePartitionsTopic) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	nameField := fields.String{Context: createPartitionsTopicName}
	fieldSize, err = nameField.SizeInBytes(version, o.name)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"name\" field")
	}
	size += fieldSize

	countField := fields.Int32{Context: createPartitionsTopicCount}
	fieldSize, err = countField.SizeInBytes(version, o.count)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"count\" field")
	}
	size += fieldSize

	assignmentsField := fields.ArrayOfStruct[createpartitionstopic.CreatePartitionsAssignment, *createpartitionstopic.CreatePartitionsAssignment]{Context: createPartitionsTopicAssignments}
	fieldSize, err = assignmentsField.SizeInBytes(version, o.assignments)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"assignments\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < CreatePartitionsTopicLowestSupportedFlexVersion() || version > CreatePartitionsTopicHighestSupportedFlexVersion() {
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
func (o *CreatePartitionsTopic) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.name.Release()
	for i := range o.assignments {
		o.assignments[i].Release()
	}
	o.assignments = nil
}

func (o *CreatePartitionsTopic) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *CreatePartitionsTopic) validateNonIgnorableFields(version int16) error {
	return nil
}

func CreatePartitionsTopicLowestSupportedVersion() int16 {
	return 0
}

func CreatePartitionsTopicHighestSupportedVersion() int16 {
	return 32767
}

func CreatePartitionsTopicLowestSupportedFlexVersion() int16 {
	return 2
}

func CreatePartitionsTopicHighestSupportedFlexVersion() int16 {
	return 32767
}

func CreatePartitionsTopicDefault() CreatePartitionsTopic {
	var d CreatePartitionsTopic
	d.SetDefault()

	return d
}
