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
package creatabletopic

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var creatableReplicaAssignmentPartitionIndex = fields.Context{
	SpecName:                    "PartitionIndex",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}
var creatableReplicaAssignmentBrokerIds = fields.Context{
	SpecName:                    "BrokerIds",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}

type CreatableReplicaAssignment struct {
	brokerIds           []int32
	unknownTaggedFields []fields.RawTaggedField
	partitionIndex      int32
	isNil               bool
}

func (o *CreatableReplicaAssignment) PartitionIndex() int32 {
	return o.partitionIndex
}

func (o *CreatableReplicaAssignment) SetPartitionIndex(val int32) {
	o.isNil = false
	o.partitionIndex = val
}

func (o *CreatableReplicaAssignment) BrokerIds() []int32 {
	return o.brokerIds
}

func (o *CreatableReplicaAssignment) SetBrokerIds(val []int32) {
	o.isNil = false
	o.brokerIds = val
}

func (o *CreatableReplicaAssignment) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *CreatableReplicaAssignment) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *CreatableReplicaAssignment) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionIndexField := fields.Int32{Context: creatableReplicaAssignmentPartitionIndex}
	if err := partitionIndexField.Read(buf, version, &o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndex\" field")
	}

	brokerIdsField := fields.Array[int32, *fields.Int32]{
		Context:          creatableReplicaAssignmentBrokerIds,
		ElementProcessor: &fields.Int32{Context: creatableReplicaAssignmentBrokerIds}}

	brokerIds, err := brokerIdsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"brokerIds\" field")
	}
	o.brokerIds = brokerIds

	// process tagged fields

	if version < CreatableReplicaAssignmentLowestSupportedFlexVersion() || version > CreatableReplicaAssignmentHighestSupportedFlexVersion() {
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

func (o *CreatableReplicaAssignment) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionIndexField := fields.Int32{Context: creatableReplicaAssignmentPartitionIndex}
	if err := partitionIndexField.Write(buf, version, o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndex\" field")
	}
	brokerIdsField := fields.Array[int32, *fields.Int32]{
		Context:          creatableReplicaAssignmentBrokerIds,
		ElementProcessor: &fields.Int32{Context: creatableReplicaAssignmentBrokerIds}}
	if err := brokerIdsField.Write(buf, version, o.brokerIds); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"brokerIds\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < CreatableReplicaAssignmentLowestSupportedFlexVersion() || version > CreatableReplicaAssignmentHighestSupportedFlexVersion() {
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

func (o *CreatableReplicaAssignment) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *CreatableReplicaAssignment) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayMarshalJSON("brokerIds", o.brokerIds); err != nil {
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

func (o *CreatableReplicaAssignment) IsNil() bool {
	return o.isNil
}

func (o *CreatableReplicaAssignment) Clear() {
	o.Release()
	o.isNil = true

	o.brokerIds = nil
	o.unknownTaggedFields = nil
}

func (o *CreatableReplicaAssignment) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partitionIndex = 0
	o.brokerIds = nil

	o.isNil = false
}

func (o *CreatableReplicaAssignment) Equal(that *CreatableReplicaAssignment) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partitionIndex != that.partitionIndex {
		return false
	}
	if !fields.PrimitiveTypeSliceEqual(o.brokerIds, that.brokerIds) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *CreatableReplicaAssignment) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionIndexField := fields.Int32{Context: creatableReplicaAssignmentPartitionIndex}
	fieldSize, err = partitionIndexField.SizeInBytes(version, o.partitionIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndex\" field")
	}
	size += fieldSize

	brokerIdsField := fields.Array[int32, *fields.Int32]{
		Context:          creatableReplicaAssignmentBrokerIds,
		ElementProcessor: &fields.Int32{Context: creatableReplicaAssignmentBrokerIds}}
	fieldSize, err = brokerIdsField.SizeInBytes(version, o.brokerIds)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"brokerIds\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < CreatableReplicaAssignmentLowestSupportedFlexVersion() || version > CreatableReplicaAssignmentHighestSupportedFlexVersion() {
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
func (o *CreatableReplicaAssignment) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.brokerIds = nil
}

func (o *CreatableReplicaAssignment) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *CreatableReplicaAssignment) validateNonIgnorableFields(version int16) error {
	return nil
}

func CreatableReplicaAssignmentLowestSupportedVersion() int16 {
	return 0
}

func CreatableReplicaAssignmentHighestSupportedVersion() int16 {
	return 32767
}

func CreatableReplicaAssignmentLowestSupportedFlexVersion() int16 {
	return 5
}

func CreatableReplicaAssignmentHighestSupportedFlexVersion() int16 {
	return 32767
}

func CreatableReplicaAssignmentDefault() CreatableReplicaAssignment {
	var d CreatableReplicaAssignment
	d.SetDefault()

	return d
}
