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
package stopreplicatopicstate

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var stopReplicaPartitionStatePartitionIndex = fields.Context{
	SpecName:                    "PartitionIndex",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var stopReplicaPartitionStateLeaderEpoch = fields.Context{
	SpecName:                    "LeaderEpoch",
	CustomDefaultValue:          int32(-1),
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var stopReplicaPartitionStateDeletePartition = fields.Context{
	SpecName:                    "DeletePartition",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type StopReplicaPartitionState struct {
	unknownTaggedFields []fields.RawTaggedField
	partitionIndex      int32
	leaderEpoch         int32
	deletePartition     bool
	isNil               bool
}

func (o *StopReplicaPartitionState) PartitionIndex() int32 {
	return o.partitionIndex
}

func (o *StopReplicaPartitionState) SetPartitionIndex(val int32) {
	o.isNil = false
	o.partitionIndex = val
}

func (o *StopReplicaPartitionState) LeaderEpoch() int32 {
	return o.leaderEpoch
}

func (o *StopReplicaPartitionState) SetLeaderEpoch(val int32) {
	o.isNil = false
	o.leaderEpoch = val
}

func (o *StopReplicaPartitionState) DeletePartition() bool {
	return o.deletePartition
}

func (o *StopReplicaPartitionState) SetDeletePartition(val bool) {
	o.isNil = false
	o.deletePartition = val
}

func (o *StopReplicaPartitionState) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *StopReplicaPartitionState) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *StopReplicaPartitionState) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionIndexField := fields.Int32{Context: stopReplicaPartitionStatePartitionIndex}
	if err := partitionIndexField.Read(buf, version, &o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndex\" field")
	}

	leaderEpochField := fields.Int32{Context: stopReplicaPartitionStateLeaderEpoch}
	if err := leaderEpochField.Read(buf, version, &o.leaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"leaderEpoch\" field")
	}

	deletePartitionField := fields.Bool{Context: stopReplicaPartitionStateDeletePartition}
	if err := deletePartitionField.Read(buf, version, &o.deletePartition); err != nil {
		return errors.WrapIf(err, "couldn't set \"deletePartition\" field")
	}

	// process tagged fields

	if version < StopReplicaPartitionStateLowestSupportedFlexVersion() || version > StopReplicaPartitionStateHighestSupportedFlexVersion() {
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

func (o *StopReplicaPartitionState) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionIndexField := fields.Int32{Context: stopReplicaPartitionStatePartitionIndex}
	if err := partitionIndexField.Write(buf, version, o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndex\" field")
	}
	leaderEpochField := fields.Int32{Context: stopReplicaPartitionStateLeaderEpoch}
	if err := leaderEpochField.Write(buf, version, o.leaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"leaderEpoch\" field")
	}
	deletePartitionField := fields.Bool{Context: stopReplicaPartitionStateDeletePartition}
	if err := deletePartitionField.Write(buf, version, o.deletePartition); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"deletePartition\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < StopReplicaPartitionStateLowestSupportedFlexVersion() || version > StopReplicaPartitionStateHighestSupportedFlexVersion() {
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

func (o *StopReplicaPartitionState) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *StopReplicaPartitionState) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 4)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.leaderEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"leaderEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.deletePartition); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"deletePartition\""), b}, []byte(": ")))
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

func (o *StopReplicaPartitionState) IsNil() bool {
	return o.isNil
}

func (o *StopReplicaPartitionState) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *StopReplicaPartitionState) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partitionIndex = 0
	o.leaderEpoch = -1
	o.deletePartition = false

	o.isNil = false
}

func (o *StopReplicaPartitionState) Equal(that *StopReplicaPartitionState) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partitionIndex != that.partitionIndex {
		return false
	}
	if o.leaderEpoch != that.leaderEpoch {
		return false
	}
	if o.deletePartition != that.deletePartition {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *StopReplicaPartitionState) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionIndexField := fields.Int32{Context: stopReplicaPartitionStatePartitionIndex}
	fieldSize, err = partitionIndexField.SizeInBytes(version, o.partitionIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndex\" field")
	}
	size += fieldSize

	leaderEpochField := fields.Int32{Context: stopReplicaPartitionStateLeaderEpoch}
	fieldSize, err = leaderEpochField.SizeInBytes(version, o.leaderEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"leaderEpoch\" field")
	}
	size += fieldSize

	deletePartitionField := fields.Bool{Context: stopReplicaPartitionStateDeletePartition}
	fieldSize, err = deletePartitionField.SizeInBytes(version, o.deletePartition)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"deletePartition\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < StopReplicaPartitionStateLowestSupportedFlexVersion() || version > StopReplicaPartitionStateHighestSupportedFlexVersion() {
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
func (o *StopReplicaPartitionState) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *StopReplicaPartitionState) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *StopReplicaPartitionState) validateNonIgnorableFields(version int16) error {
	if !stopReplicaPartitionStatePartitionIndex.IsSupportedVersion(version) {
		if o.partitionIndex != 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"partitionIndex\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !stopReplicaPartitionStateLeaderEpoch.IsSupportedVersion(version) {
		if o.leaderEpoch != -1 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"leaderEpoch\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !stopReplicaPartitionStateDeletePartition.IsSupportedVersion(version) {
		if o.deletePartition {
			return errors.New(strings.Join([]string{"attempted to write non-default \"deletePartition\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func StopReplicaPartitionStateLowestSupportedVersion() int16 {
	return 3
}

func StopReplicaPartitionStateHighestSupportedVersion() int16 {
	return 32767
}

func StopReplicaPartitionStateLowestSupportedFlexVersion() int16 {
	return 2
}

func StopReplicaPartitionStateHighestSupportedFlexVersion() int16 {
	return 32767
}

func StopReplicaPartitionStateDefault() StopReplicaPartitionState {
	var d StopReplicaPartitionState
	d.SetDefault()

	return d
}
