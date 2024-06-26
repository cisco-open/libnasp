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
package offsetforleadertopic

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var offsetForLeaderPartitionPartition = fields.Context{
	SpecName:                    "Partition",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}
var offsetForLeaderPartitionCurrentLeaderEpoch = fields.Context{
	SpecName:                    "CurrentLeaderEpoch",
	CustomDefaultValue:          int32(-1),
	LowestSupportedVersion:      2,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}
var offsetForLeaderPartitionLeaderEpoch = fields.Context{
	SpecName:                    "LeaderEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}

type OffsetForLeaderPartition struct {
	unknownTaggedFields []fields.RawTaggedField
	partition           int32
	currentLeaderEpoch  int32
	leaderEpoch         int32
	isNil               bool
}

func (o *OffsetForLeaderPartition) Partition() int32 {
	return o.partition
}

func (o *OffsetForLeaderPartition) SetPartition(val int32) {
	o.isNil = false
	o.partition = val
}

func (o *OffsetForLeaderPartition) CurrentLeaderEpoch() int32 {
	return o.currentLeaderEpoch
}

func (o *OffsetForLeaderPartition) SetCurrentLeaderEpoch(val int32) {
	o.isNil = false
	o.currentLeaderEpoch = val
}

func (o *OffsetForLeaderPartition) LeaderEpoch() int32 {
	return o.leaderEpoch
}

func (o *OffsetForLeaderPartition) SetLeaderEpoch(val int32) {
	o.isNil = false
	o.leaderEpoch = val
}

func (o *OffsetForLeaderPartition) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *OffsetForLeaderPartition) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *OffsetForLeaderPartition) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionField := fields.Int32{Context: offsetForLeaderPartitionPartition}
	if err := partitionField.Read(buf, version, &o.partition); err != nil {
		return errors.WrapIf(err, "couldn't set \"partition\" field")
	}

	currentLeaderEpochField := fields.Int32{Context: offsetForLeaderPartitionCurrentLeaderEpoch}
	if err := currentLeaderEpochField.Read(buf, version, &o.currentLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"currentLeaderEpoch\" field")
	}

	leaderEpochField := fields.Int32{Context: offsetForLeaderPartitionLeaderEpoch}
	if err := leaderEpochField.Read(buf, version, &o.leaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"leaderEpoch\" field")
	}

	// process tagged fields

	if version < OffsetForLeaderPartitionLowestSupportedFlexVersion() || version > OffsetForLeaderPartitionHighestSupportedFlexVersion() {
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

func (o *OffsetForLeaderPartition) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionField := fields.Int32{Context: offsetForLeaderPartitionPartition}
	if err := partitionField.Write(buf, version, o.partition); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partition\" field")
	}
	currentLeaderEpochField := fields.Int32{Context: offsetForLeaderPartitionCurrentLeaderEpoch}
	if err := currentLeaderEpochField.Write(buf, version, o.currentLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"currentLeaderEpoch\" field")
	}
	leaderEpochField := fields.Int32{Context: offsetForLeaderPartitionLeaderEpoch}
	if err := leaderEpochField.Write(buf, version, o.leaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"leaderEpoch\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < OffsetForLeaderPartitionLowestSupportedFlexVersion() || version > OffsetForLeaderPartitionHighestSupportedFlexVersion() {
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

func (o *OffsetForLeaderPartition) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *OffsetForLeaderPartition) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 4)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partition); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partition\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.currentLeaderEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"currentLeaderEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.leaderEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"leaderEpoch\""), b}, []byte(": ")))
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

func (o *OffsetForLeaderPartition) IsNil() bool {
	return o.isNil
}

func (o *OffsetForLeaderPartition) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *OffsetForLeaderPartition) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partition = 0
	o.currentLeaderEpoch = -1
	o.leaderEpoch = 0

	o.isNil = false
}

func (o *OffsetForLeaderPartition) Equal(that *OffsetForLeaderPartition) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partition != that.partition {
		return false
	}
	if o.currentLeaderEpoch != that.currentLeaderEpoch {
		return false
	}
	if o.leaderEpoch != that.leaderEpoch {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *OffsetForLeaderPartition) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionField := fields.Int32{Context: offsetForLeaderPartitionPartition}
	fieldSize, err = partitionField.SizeInBytes(version, o.partition)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partition\" field")
	}
	size += fieldSize

	currentLeaderEpochField := fields.Int32{Context: offsetForLeaderPartitionCurrentLeaderEpoch}
	fieldSize, err = currentLeaderEpochField.SizeInBytes(version, o.currentLeaderEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"currentLeaderEpoch\" field")
	}
	size += fieldSize

	leaderEpochField := fields.Int32{Context: offsetForLeaderPartitionLeaderEpoch}
	fieldSize, err = leaderEpochField.SizeInBytes(version, o.leaderEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"leaderEpoch\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < OffsetForLeaderPartitionLowestSupportedFlexVersion() || version > OffsetForLeaderPartitionHighestSupportedFlexVersion() {
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
func (o *OffsetForLeaderPartition) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *OffsetForLeaderPartition) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *OffsetForLeaderPartition) validateNonIgnorableFields(version int16) error {
	return nil
}

func OffsetForLeaderPartitionLowestSupportedVersion() int16 {
	return 0
}

func OffsetForLeaderPartitionHighestSupportedVersion() int16 {
	return 32767
}

func OffsetForLeaderPartitionLowestSupportedFlexVersion() int16 {
	return 4
}

func OffsetForLeaderPartitionHighestSupportedFlexVersion() int16 {
	return 32767
}

func OffsetForLeaderPartitionDefault() OffsetForLeaderPartition {
	var d OffsetForLeaderPartition
	d.SetDefault()

	return d
}
