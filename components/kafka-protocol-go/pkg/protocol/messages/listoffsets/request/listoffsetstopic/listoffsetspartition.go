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
package listoffsetstopic

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var listOffsetsPartitionPartitionIndex = fields.Context{
	SpecName:                    "PartitionIndex",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var listOffsetsPartitionCurrentLeaderEpoch = fields.Context{
	SpecName:                    "CurrentLeaderEpoch",
	CustomDefaultValue:          int32(-1),
	LowestSupportedVersion:      4,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var listOffsetsPartitionTimestamp = fields.Context{
	SpecName:                    "Timestamp",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var listOffsetsPartitionMaxNumOffsets = fields.Context{
	SpecName:                    "MaxNumOffsets",
	CustomDefaultValue:          int32(1),
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     0,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}

type ListOffsetsPartition struct {
	unknownTaggedFields []fields.RawTaggedField
	timestamp           int64
	partitionIndex      int32
	currentLeaderEpoch  int32
	maxNumOffsets       int32
	isNil               bool
}

func (o *ListOffsetsPartition) PartitionIndex() int32 {
	return o.partitionIndex
}

func (o *ListOffsetsPartition) SetPartitionIndex(val int32) {
	o.isNil = false
	o.partitionIndex = val
}

func (o *ListOffsetsPartition) CurrentLeaderEpoch() int32 {
	return o.currentLeaderEpoch
}

func (o *ListOffsetsPartition) SetCurrentLeaderEpoch(val int32) {
	o.isNil = false
	o.currentLeaderEpoch = val
}

func (o *ListOffsetsPartition) Timestamp() int64 {
	return o.timestamp
}

func (o *ListOffsetsPartition) SetTimestamp(val int64) {
	o.isNil = false
	o.timestamp = val
}

func (o *ListOffsetsPartition) MaxNumOffsets() int32 {
	return o.maxNumOffsets
}

func (o *ListOffsetsPartition) SetMaxNumOffsets(val int32) {
	o.isNil = false
	o.maxNumOffsets = val
}

func (o *ListOffsetsPartition) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *ListOffsetsPartition) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *ListOffsetsPartition) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionIndexField := fields.Int32{Context: listOffsetsPartitionPartitionIndex}
	if err := partitionIndexField.Read(buf, version, &o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndex\" field")
	}

	currentLeaderEpochField := fields.Int32{Context: listOffsetsPartitionCurrentLeaderEpoch}
	if err := currentLeaderEpochField.Read(buf, version, &o.currentLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"currentLeaderEpoch\" field")
	}

	timestampField := fields.Int64{Context: listOffsetsPartitionTimestamp}
	if err := timestampField.Read(buf, version, &o.timestamp); err != nil {
		return errors.WrapIf(err, "couldn't set \"timestamp\" field")
	}

	maxNumOffsetsField := fields.Int32{Context: listOffsetsPartitionMaxNumOffsets}
	if err := maxNumOffsetsField.Read(buf, version, &o.maxNumOffsets); err != nil {
		return errors.WrapIf(err, "couldn't set \"maxNumOffsets\" field")
	}

	// process tagged fields

	if version < ListOffsetsPartitionLowestSupportedFlexVersion() || version > ListOffsetsPartitionHighestSupportedFlexVersion() {
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

func (o *ListOffsetsPartition) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionIndexField := fields.Int32{Context: listOffsetsPartitionPartitionIndex}
	if err := partitionIndexField.Write(buf, version, o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndex\" field")
	}
	currentLeaderEpochField := fields.Int32{Context: listOffsetsPartitionCurrentLeaderEpoch}
	if err := currentLeaderEpochField.Write(buf, version, o.currentLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"currentLeaderEpoch\" field")
	}
	timestampField := fields.Int64{Context: listOffsetsPartitionTimestamp}
	if err := timestampField.Write(buf, version, o.timestamp); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"timestamp\" field")
	}
	maxNumOffsetsField := fields.Int32{Context: listOffsetsPartitionMaxNumOffsets}
	if err := maxNumOffsetsField.Write(buf, version, o.maxNumOffsets); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"maxNumOffsets\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ListOffsetsPartitionLowestSupportedFlexVersion() || version > ListOffsetsPartitionHighestSupportedFlexVersion() {
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

func (o *ListOffsetsPartition) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *ListOffsetsPartition) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.currentLeaderEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"currentLeaderEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.timestamp); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"timestamp\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.maxNumOffsets); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"maxNumOffsets\""), b}, []byte(": ")))
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

func (o *ListOffsetsPartition) IsNil() bool {
	return o.isNil
}

func (o *ListOffsetsPartition) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *ListOffsetsPartition) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partitionIndex = 0
	o.currentLeaderEpoch = -1
	o.timestamp = 0
	o.maxNumOffsets = 1

	o.isNil = false
}

func (o *ListOffsetsPartition) Equal(that *ListOffsetsPartition) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partitionIndex != that.partitionIndex {
		return false
	}
	if o.currentLeaderEpoch != that.currentLeaderEpoch {
		return false
	}
	if o.timestamp != that.timestamp {
		return false
	}
	if o.maxNumOffsets != that.maxNumOffsets {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *ListOffsetsPartition) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionIndexField := fields.Int32{Context: listOffsetsPartitionPartitionIndex}
	fieldSize, err = partitionIndexField.SizeInBytes(version, o.partitionIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndex\" field")
	}
	size += fieldSize

	currentLeaderEpochField := fields.Int32{Context: listOffsetsPartitionCurrentLeaderEpoch}
	fieldSize, err = currentLeaderEpochField.SizeInBytes(version, o.currentLeaderEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"currentLeaderEpoch\" field")
	}
	size += fieldSize

	timestampField := fields.Int64{Context: listOffsetsPartitionTimestamp}
	fieldSize, err = timestampField.SizeInBytes(version, o.timestamp)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"timestamp\" field")
	}
	size += fieldSize

	maxNumOffsetsField := fields.Int32{Context: listOffsetsPartitionMaxNumOffsets}
	fieldSize, err = maxNumOffsetsField.SizeInBytes(version, o.maxNumOffsets)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"maxNumOffsets\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < ListOffsetsPartitionLowestSupportedFlexVersion() || version > ListOffsetsPartitionHighestSupportedFlexVersion() {
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
func (o *ListOffsetsPartition) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *ListOffsetsPartition) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *ListOffsetsPartition) validateNonIgnorableFields(version int16) error {
	return nil
}

func ListOffsetsPartitionLowestSupportedVersion() int16 {
	return 0
}

func ListOffsetsPartitionHighestSupportedVersion() int16 {
	return 32767
}

func ListOffsetsPartitionLowestSupportedFlexVersion() int16 {
	return 6
}

func ListOffsetsPartitionHighestSupportedFlexVersion() int16 {
	return 32767
}

func ListOffsetsPartitionDefault() ListOffsetsPartition {
	var d ListOffsetsPartition
	d.SetDefault()

	return d
}
