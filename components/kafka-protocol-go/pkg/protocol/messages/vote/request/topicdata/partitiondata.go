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
package topicdata

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var partitionDataPartitionIndex = fields.Context{
	SpecName:                    "PartitionIndex",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionDataCandidateEpoch = fields.Context{
	SpecName:                    "CandidateEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionDataCandidateId = fields.Context{
	SpecName:                    "CandidateId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionDataLastOffsetEpoch = fields.Context{
	SpecName:                    "LastOffsetEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var partitionDataLastOffset = fields.Context{
	SpecName:                    "LastOffset",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}

type PartitionData struct {
	unknownTaggedFields []fields.RawTaggedField
	lastOffset          int64
	partitionIndex      int32
	candidateEpoch      int32
	candidateId         int32
	lastOffsetEpoch     int32
	isNil               bool
}

func (o *PartitionData) PartitionIndex() int32 {
	return o.partitionIndex
}

func (o *PartitionData) SetPartitionIndex(val int32) {
	o.isNil = false
	o.partitionIndex = val
}

func (o *PartitionData) CandidateEpoch() int32 {
	return o.candidateEpoch
}

func (o *PartitionData) SetCandidateEpoch(val int32) {
	o.isNil = false
	o.candidateEpoch = val
}

func (o *PartitionData) CandidateId() int32 {
	return o.candidateId
}

func (o *PartitionData) SetCandidateId(val int32) {
	o.isNil = false
	o.candidateId = val
}

func (o *PartitionData) LastOffsetEpoch() int32 {
	return o.lastOffsetEpoch
}

func (o *PartitionData) SetLastOffsetEpoch(val int32) {
	o.isNil = false
	o.lastOffsetEpoch = val
}

func (o *PartitionData) LastOffset() int64 {
	return o.lastOffset
}

func (o *PartitionData) SetLastOffset(val int64) {
	o.isNil = false
	o.lastOffset = val
}

func (o *PartitionData) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *PartitionData) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *PartitionData) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionIndexField := fields.Int32{Context: partitionDataPartitionIndex}
	if err := partitionIndexField.Read(buf, version, &o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndex\" field")
	}

	candidateEpochField := fields.Int32{Context: partitionDataCandidateEpoch}
	if err := candidateEpochField.Read(buf, version, &o.candidateEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"candidateEpoch\" field")
	}

	candidateIdField := fields.Int32{Context: partitionDataCandidateId}
	if err := candidateIdField.Read(buf, version, &o.candidateId); err != nil {
		return errors.WrapIf(err, "couldn't set \"candidateId\" field")
	}

	lastOffsetEpochField := fields.Int32{Context: partitionDataLastOffsetEpoch}
	if err := lastOffsetEpochField.Read(buf, version, &o.lastOffsetEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"lastOffsetEpoch\" field")
	}

	lastOffsetField := fields.Int64{Context: partitionDataLastOffset}
	if err := lastOffsetField.Read(buf, version, &o.lastOffset); err != nil {
		return errors.WrapIf(err, "couldn't set \"lastOffset\" field")
	}

	// process tagged fields

	if version < PartitionDataLowestSupportedFlexVersion() || version > PartitionDataHighestSupportedFlexVersion() {
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

func (o *PartitionData) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionIndexField := fields.Int32{Context: partitionDataPartitionIndex}
	if err := partitionIndexField.Write(buf, version, o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndex\" field")
	}
	candidateEpochField := fields.Int32{Context: partitionDataCandidateEpoch}
	if err := candidateEpochField.Write(buf, version, o.candidateEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"candidateEpoch\" field")
	}
	candidateIdField := fields.Int32{Context: partitionDataCandidateId}
	if err := candidateIdField.Write(buf, version, o.candidateId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"candidateId\" field")
	}
	lastOffsetEpochField := fields.Int32{Context: partitionDataLastOffsetEpoch}
	if err := lastOffsetEpochField.Write(buf, version, o.lastOffsetEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"lastOffsetEpoch\" field")
	}
	lastOffsetField := fields.Int64{Context: partitionDataLastOffset}
	if err := lastOffsetField.Write(buf, version, o.lastOffset); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"lastOffset\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < PartitionDataLowestSupportedFlexVersion() || version > PartitionDataHighestSupportedFlexVersion() {
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

func (o *PartitionData) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *PartitionData) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 6)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.candidateEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"candidateEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.candidateId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"candidateId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.lastOffsetEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"lastOffsetEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.lastOffset); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"lastOffset\""), b}, []byte(": ")))
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

func (o *PartitionData) IsNil() bool {
	return o.isNil
}

func (o *PartitionData) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *PartitionData) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partitionIndex = 0
	o.candidateEpoch = 0
	o.candidateId = 0
	o.lastOffsetEpoch = 0
	o.lastOffset = 0

	o.isNil = false
}

func (o *PartitionData) Equal(that *PartitionData) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partitionIndex != that.partitionIndex {
		return false
	}
	if o.candidateEpoch != that.candidateEpoch {
		return false
	}
	if o.candidateId != that.candidateId {
		return false
	}
	if o.lastOffsetEpoch != that.lastOffsetEpoch {
		return false
	}
	if o.lastOffset != that.lastOffset {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *PartitionData) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionIndexField := fields.Int32{Context: partitionDataPartitionIndex}
	fieldSize, err = partitionIndexField.SizeInBytes(version, o.partitionIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndex\" field")
	}
	size += fieldSize

	candidateEpochField := fields.Int32{Context: partitionDataCandidateEpoch}
	fieldSize, err = candidateEpochField.SizeInBytes(version, o.candidateEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"candidateEpoch\" field")
	}
	size += fieldSize

	candidateIdField := fields.Int32{Context: partitionDataCandidateId}
	fieldSize, err = candidateIdField.SizeInBytes(version, o.candidateId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"candidateId\" field")
	}
	size += fieldSize

	lastOffsetEpochField := fields.Int32{Context: partitionDataLastOffsetEpoch}
	fieldSize, err = lastOffsetEpochField.SizeInBytes(version, o.lastOffsetEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"lastOffsetEpoch\" field")
	}
	size += fieldSize

	lastOffsetField := fields.Int64{Context: partitionDataLastOffset}
	fieldSize, err = lastOffsetField.SizeInBytes(version, o.lastOffset)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"lastOffset\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < PartitionDataLowestSupportedFlexVersion() || version > PartitionDataHighestSupportedFlexVersion() {
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
func (o *PartitionData) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *PartitionData) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *PartitionData) validateNonIgnorableFields(version int16) error {
	return nil
}

func PartitionDataLowestSupportedVersion() int16 {
	return 0
}

func PartitionDataHighestSupportedVersion() int16 {
	return 32767
}

func PartitionDataLowestSupportedFlexVersion() int16 {
	return 0
}

func PartitionDataHighestSupportedFlexVersion() int16 {
	return 32767
}

func PartitionDataDefault() PartitionData {
	var d PartitionData
	d.SetDefault()

	return d
}