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
package txnoffsetcommitrequesttopic

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var txnOffsetCommitRequestPartitionPartitionIndex = fields.Context{
	SpecName:                    "PartitionIndex",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var txnOffsetCommitRequestPartitionCommittedOffset = fields.Context{
	SpecName:                    "CommittedOffset",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var txnOffsetCommitRequestPartitionCommittedLeaderEpoch = fields.Context{
	SpecName:                    "CommittedLeaderEpoch",
	CustomDefaultValue:          int32(-1),
	LowestSupportedVersion:      2,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var txnOffsetCommitRequestPartitionCommittedMetadata = fields.Context{
	SpecName:                        "CommittedMetadata",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      3,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}

type TxnOffsetCommitRequestPartition struct {
	unknownTaggedFields  []fields.RawTaggedField
	committedMetadata    fields.NullableString
	committedOffset      int64
	partitionIndex       int32
	committedLeaderEpoch int32
	isNil                bool
}

func (o *TxnOffsetCommitRequestPartition) PartitionIndex() int32 {
	return o.partitionIndex
}

func (o *TxnOffsetCommitRequestPartition) SetPartitionIndex(val int32) {
	o.isNil = false
	o.partitionIndex = val
}

func (o *TxnOffsetCommitRequestPartition) CommittedOffset() int64 {
	return o.committedOffset
}

func (o *TxnOffsetCommitRequestPartition) SetCommittedOffset(val int64) {
	o.isNil = false
	o.committedOffset = val
}

func (o *TxnOffsetCommitRequestPartition) CommittedLeaderEpoch() int32 {
	return o.committedLeaderEpoch
}

func (o *TxnOffsetCommitRequestPartition) SetCommittedLeaderEpoch(val int32) {
	o.isNil = false
	o.committedLeaderEpoch = val
}

func (o *TxnOffsetCommitRequestPartition) CommittedMetadata() fields.NullableString {
	return o.committedMetadata
}

func (o *TxnOffsetCommitRequestPartition) SetCommittedMetadata(val fields.NullableString) {
	o.isNil = false
	o.committedMetadata = val
}

func (o *TxnOffsetCommitRequestPartition) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *TxnOffsetCommitRequestPartition) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *TxnOffsetCommitRequestPartition) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionIndexField := fields.Int32{Context: txnOffsetCommitRequestPartitionPartitionIndex}
	if err := partitionIndexField.Read(buf, version, &o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndex\" field")
	}

	committedOffsetField := fields.Int64{Context: txnOffsetCommitRequestPartitionCommittedOffset}
	if err := committedOffsetField.Read(buf, version, &o.committedOffset); err != nil {
		return errors.WrapIf(err, "couldn't set \"committedOffset\" field")
	}

	committedLeaderEpochField := fields.Int32{Context: txnOffsetCommitRequestPartitionCommittedLeaderEpoch}
	if err := committedLeaderEpochField.Read(buf, version, &o.committedLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"committedLeaderEpoch\" field")
	}

	committedMetadataField := fields.String{Context: txnOffsetCommitRequestPartitionCommittedMetadata}
	if err := committedMetadataField.Read(buf, version, &o.committedMetadata); err != nil {
		return errors.WrapIf(err, "couldn't set \"committedMetadata\" field")
	}

	// process tagged fields

	if version < TxnOffsetCommitRequestPartitionLowestSupportedFlexVersion() || version > TxnOffsetCommitRequestPartitionHighestSupportedFlexVersion() {
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

func (o *TxnOffsetCommitRequestPartition) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionIndexField := fields.Int32{Context: txnOffsetCommitRequestPartitionPartitionIndex}
	if err := partitionIndexField.Write(buf, version, o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndex\" field")
	}
	committedOffsetField := fields.Int64{Context: txnOffsetCommitRequestPartitionCommittedOffset}
	if err := committedOffsetField.Write(buf, version, o.committedOffset); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"committedOffset\" field")
	}
	committedLeaderEpochField := fields.Int32{Context: txnOffsetCommitRequestPartitionCommittedLeaderEpoch}
	if err := committedLeaderEpochField.Write(buf, version, o.committedLeaderEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"committedLeaderEpoch\" field")
	}
	committedMetadataField := fields.String{Context: txnOffsetCommitRequestPartitionCommittedMetadata}
	if err := committedMetadataField.Write(buf, version, o.committedMetadata); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"committedMetadata\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < TxnOffsetCommitRequestPartitionLowestSupportedFlexVersion() || version > TxnOffsetCommitRequestPartitionHighestSupportedFlexVersion() {
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

func (o *TxnOffsetCommitRequestPartition) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *TxnOffsetCommitRequestPartition) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.committedOffset); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"committedOffset\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.committedLeaderEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"committedLeaderEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.committedMetadata); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"committedMetadata\""), b}, []byte(": ")))
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

func (o *TxnOffsetCommitRequestPartition) IsNil() bool {
	return o.isNil
}

func (o *TxnOffsetCommitRequestPartition) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *TxnOffsetCommitRequestPartition) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partitionIndex = 0
	o.committedOffset = 0
	o.committedLeaderEpoch = -1
	o.committedMetadata.SetValue("")

	o.isNil = false
}

func (o *TxnOffsetCommitRequestPartition) Equal(that *TxnOffsetCommitRequestPartition) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partitionIndex != that.partitionIndex {
		return false
	}
	if o.committedOffset != that.committedOffset {
		return false
	}
	if o.committedLeaderEpoch != that.committedLeaderEpoch {
		return false
	}
	if !o.committedMetadata.Equal(&that.committedMetadata) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *TxnOffsetCommitRequestPartition) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionIndexField := fields.Int32{Context: txnOffsetCommitRequestPartitionPartitionIndex}
	fieldSize, err = partitionIndexField.SizeInBytes(version, o.partitionIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndex\" field")
	}
	size += fieldSize

	committedOffsetField := fields.Int64{Context: txnOffsetCommitRequestPartitionCommittedOffset}
	fieldSize, err = committedOffsetField.SizeInBytes(version, o.committedOffset)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"committedOffset\" field")
	}
	size += fieldSize

	committedLeaderEpochField := fields.Int32{Context: txnOffsetCommitRequestPartitionCommittedLeaderEpoch}
	fieldSize, err = committedLeaderEpochField.SizeInBytes(version, o.committedLeaderEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"committedLeaderEpoch\" field")
	}
	size += fieldSize

	committedMetadataField := fields.String{Context: txnOffsetCommitRequestPartitionCommittedMetadata}
	fieldSize, err = committedMetadataField.SizeInBytes(version, o.committedMetadata)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"committedMetadata\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < TxnOffsetCommitRequestPartitionLowestSupportedFlexVersion() || version > TxnOffsetCommitRequestPartitionHighestSupportedFlexVersion() {
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
func (o *TxnOffsetCommitRequestPartition) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.committedMetadata.Release()
}

func (o *TxnOffsetCommitRequestPartition) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *TxnOffsetCommitRequestPartition) validateNonIgnorableFields(version int16) error {
	return nil
}

func TxnOffsetCommitRequestPartitionLowestSupportedVersion() int16 {
	return 0
}

func TxnOffsetCommitRequestPartitionHighestSupportedVersion() int16 {
	return 32767
}

func TxnOffsetCommitRequestPartitionLowestSupportedFlexVersion() int16 {
	return 3
}

func TxnOffsetCommitRequestPartitionHighestSupportedFlexVersion() int16 {
	return 32767
}

func TxnOffsetCommitRequestPartitionDefault() TxnOffsetCommitRequestPartition {
	var d TxnOffsetCommitRequestPartition
	d.SetDefault()

	return d
}
