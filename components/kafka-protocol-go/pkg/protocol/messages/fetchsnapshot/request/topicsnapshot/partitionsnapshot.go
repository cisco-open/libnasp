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
package topicsnapshot

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/fetchsnapshot/request/topicsnapshot/partitionsnapshot"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var partitionSnapshotPartition = fields.Context{
	SpecName:                    "Partition",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionSnapshotCurrentLeaderEpoch = fields.Context{
	SpecName:                    "CurrentLeaderEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionSnapshotSnapshotId = fields.Context{
	SpecName:                    "SnapshotId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionSnapshotPosition = fields.Context{
	SpecName:                    "Position",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}

type PartitionSnapshot struct {
	unknownTaggedFields []fields.RawTaggedField
	snapshotId          partitionsnapshot.SnapshotId
	position            int64
	partition           int32
	currentLeaderEpoch  int32
	isNil               bool
}

func (o *PartitionSnapshot) Partition() int32 {
	return o.partition
}

func (o *PartitionSnapshot) SetPartition(val int32) {
	o.isNil = false
	o.partition = val
}

func (o *PartitionSnapshot) CurrentLeaderEpoch() int32 {
	return o.currentLeaderEpoch
}

func (o *PartitionSnapshot) SetCurrentLeaderEpoch(val int32) {
	o.isNil = false
	o.currentLeaderEpoch = val
}

func (o *PartitionSnapshot) SnapshotId() partitionsnapshot.SnapshotId {
	return o.snapshotId
}

func (o *PartitionSnapshot) SetSnapshotId(val partitionsnapshot.SnapshotId) {
	o.isNil = false
	o.snapshotId = val
}

func (o *PartitionSnapshot) Position() int64 {
	return o.position
}

func (o *PartitionSnapshot) SetPosition(val int64) {
	o.isNil = false
	o.position = val
}

func (o *PartitionSnapshot) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *PartitionSnapshot) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *PartitionSnapshot) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionField := fields.Int32{Context: partitionSnapshotPartition}
	if err := partitionField.Read(buf, version, &o.partition); err != nil {
		return errors.WrapIf(err, "couldn't set \"partition\" field")
	}

	currentLeaderEpochField := fields.Int32{Context: partitionSnapshotCurrentLeaderEpoch}
	if err := currentLeaderEpochField.Read(buf, version, &o.currentLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"currentLeaderEpoch\" field")
	}

	if partitionSnapshotSnapshotId.IsSupportedVersion(version) {
		if err := o.snapshotId.Read(buf, version); err != nil {
			return errors.WrapIf(err, "couldn't set \"snapshotId\" field")
		}
	}

	positionField := fields.Int64{Context: partitionSnapshotPosition}
	if err := positionField.Read(buf, version, &o.position); err != nil {
		return errors.WrapIf(err, "couldn't set \"position\" field")
	}

	// process tagged fields

	if version < PartitionSnapshotLowestSupportedFlexVersion() || version > PartitionSnapshotHighestSupportedFlexVersion() {
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

func (o *PartitionSnapshot) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionField := fields.Int32{Context: partitionSnapshotPartition}
	if err := partitionField.Write(buf, version, o.partition); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partition\" field")
	}
	currentLeaderEpochField := fields.Int32{Context: partitionSnapshotCurrentLeaderEpoch}
	if err := currentLeaderEpochField.Write(buf, version, o.currentLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"currentLeaderEpoch\" field")
	}

	if partitionSnapshotSnapshotId.IsSupportedVersion(version) {
		if err := o.snapshotId.Write(buf, version); err != nil {
			return errors.WrapIf(err, "couldn't serialize \"snapshotId\" field")
		}
	}

	positionField := fields.Int64{Context: partitionSnapshotPosition}
	if err := positionField.Write(buf, version, o.position); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"position\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < PartitionSnapshotLowestSupportedFlexVersion() || version > PartitionSnapshotHighestSupportedFlexVersion() {
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

func (o *PartitionSnapshot) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *PartitionSnapshot) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
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
	if b, err := o.snapshotId.MarshalJSON(); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"snapshotId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.position); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"position\""), b}, []byte(": ")))
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

func (o *PartitionSnapshot) IsNil() bool {
	return o.isNil
}

func (o *PartitionSnapshot) Clear() {
	o.Release()
	o.isNil = true

	o.snapshotId.Clear()
	o.unknownTaggedFields = nil
}

func (o *PartitionSnapshot) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partition = 0
	o.currentLeaderEpoch = 0
	o.snapshotId.SetDefault()
	o.position = 0

	o.isNil = false
}

func (o *PartitionSnapshot) Equal(that *PartitionSnapshot) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partition != that.partition {
		return false
	}
	if o.currentLeaderEpoch != that.currentLeaderEpoch {
		return false
	}
	if !o.snapshotId.Equal(&that.snapshotId) {
		return false
	}
	if o.position != that.position {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *PartitionSnapshot) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionField := fields.Int32{Context: partitionSnapshotPartition}
	fieldSize, err = partitionField.SizeInBytes(version, o.partition)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partition\" field")
	}
	size += fieldSize

	currentLeaderEpochField := fields.Int32{Context: partitionSnapshotCurrentLeaderEpoch}
	fieldSize, err = currentLeaderEpochField.SizeInBytes(version, o.currentLeaderEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"currentLeaderEpoch\" field")
	}
	size += fieldSize

	fieldSize, err = o.snapshotId.SizeInBytes(version)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"snapshotId\" field")
	}
	size += fieldSize

	positionField := fields.Int64{Context: partitionSnapshotPosition}
	fieldSize, err = positionField.SizeInBytes(version, o.position)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"position\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < PartitionSnapshotLowestSupportedFlexVersion() || version > PartitionSnapshotHighestSupportedFlexVersion() {
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
func (o *PartitionSnapshot) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.snapshotId.Release()
}

func (o *PartitionSnapshot) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *PartitionSnapshot) validateNonIgnorableFields(version int16) error {
	return nil
}

func PartitionSnapshotLowestSupportedVersion() int16 {
	return 0
}

func PartitionSnapshotHighestSupportedVersion() int16 {
	return 32767
}

func PartitionSnapshotLowestSupportedFlexVersion() int16 {
	return 0
}

func PartitionSnapshotHighestSupportedFlexVersion() int16 {
	return 32767
}

func PartitionSnapshotDefault() PartitionSnapshot {
	var d PartitionSnapshot
	d.SetDefault()

	return d
}