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
package describelogdirstopic

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var describeLogDirsPartitionPartitionIndex = fields.Context{
	SpecName:                    "PartitionIndex",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describeLogDirsPartitionPartitionSize = fields.Context{
	SpecName:                    "PartitionSize",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describeLogDirsPartitionOffsetLag = fields.Context{
	SpecName:                    "OffsetLag",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describeLogDirsPartitionIsFutureKey = fields.Context{
	SpecName:                    "IsFutureKey",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type DescribeLogDirsPartition struct {
	unknownTaggedFields []fields.RawTaggedField
	partitionSize       int64
	offsetLag           int64
	partitionIndex      int32
	isFutureKey         bool
	isNil               bool
}

func (o *DescribeLogDirsPartition) PartitionIndex() int32 {
	return o.partitionIndex
}

func (o *DescribeLogDirsPartition) SetPartitionIndex(val int32) {
	o.isNil = false
	o.partitionIndex = val
}

func (o *DescribeLogDirsPartition) PartitionSize() int64 {
	return o.partitionSize
}

func (o *DescribeLogDirsPartition) SetPartitionSize(val int64) {
	o.isNil = false
	o.partitionSize = val
}

func (o *DescribeLogDirsPartition) OffsetLag() int64 {
	return o.offsetLag
}

func (o *DescribeLogDirsPartition) SetOffsetLag(val int64) {
	o.isNil = false
	o.offsetLag = val
}

func (o *DescribeLogDirsPartition) IsFutureKey() bool {
	return o.isFutureKey
}

func (o *DescribeLogDirsPartition) SetIsFutureKey(val bool) {
	o.isNil = false
	o.isFutureKey = val
}

func (o *DescribeLogDirsPartition) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *DescribeLogDirsPartition) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *DescribeLogDirsPartition) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	partitionIndexField := fields.Int32{Context: describeLogDirsPartitionPartitionIndex}
	if err := partitionIndexField.Read(buf, version, &o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionIndex\" field")
	}

	partitionSizeField := fields.Int64{Context: describeLogDirsPartitionPartitionSize}
	if err := partitionSizeField.Read(buf, version, &o.partitionSize); err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionSize\" field")
	}

	offsetLagField := fields.Int64{Context: describeLogDirsPartitionOffsetLag}
	if err := offsetLagField.Read(buf, version, &o.offsetLag); err != nil {
		return errors.WrapIf(err, "couldn't set \"offsetLag\" field")
	}

	isFutureKeyField := fields.Bool{Context: describeLogDirsPartitionIsFutureKey}
	if err := isFutureKeyField.Read(buf, version, &o.isFutureKey); err != nil {
		return errors.WrapIf(err, "couldn't set \"isFutureKey\" field")
	}

	// process tagged fields

	if version < DescribeLogDirsPartitionLowestSupportedFlexVersion() || version > DescribeLogDirsPartitionHighestSupportedFlexVersion() {
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

func (o *DescribeLogDirsPartition) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	partitionIndexField := fields.Int32{Context: describeLogDirsPartitionPartitionIndex}
	if err := partitionIndexField.Write(buf, version, o.partitionIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionIndex\" field")
	}
	partitionSizeField := fields.Int64{Context: describeLogDirsPartitionPartitionSize}
	if err := partitionSizeField.Write(buf, version, o.partitionSize); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionSize\" field")
	}
	offsetLagField := fields.Int64{Context: describeLogDirsPartitionOffsetLag}
	if err := offsetLagField.Write(buf, version, o.offsetLag); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"offsetLag\" field")
	}
	isFutureKeyField := fields.Bool{Context: describeLogDirsPartitionIsFutureKey}
	if err := isFutureKeyField.Write(buf, version, o.isFutureKey); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"isFutureKey\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < DescribeLogDirsPartitionLowestSupportedFlexVersion() || version > DescribeLogDirsPartitionHighestSupportedFlexVersion() {
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

func (o *DescribeLogDirsPartition) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *DescribeLogDirsPartition) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.partitionSize); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"partitionSize\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.offsetLag); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"offsetLag\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.isFutureKey); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"isFutureKey\""), b}, []byte(": ")))
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

func (o *DescribeLogDirsPartition) IsNil() bool {
	return o.isNil
}

func (o *DescribeLogDirsPartition) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *DescribeLogDirsPartition) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.partitionIndex = 0
	o.partitionSize = 0
	o.offsetLag = 0
	o.isFutureKey = false

	o.isNil = false
}

func (o *DescribeLogDirsPartition) Equal(that *DescribeLogDirsPartition) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.partitionIndex != that.partitionIndex {
		return false
	}
	if o.partitionSize != that.partitionSize {
		return false
	}
	if o.offsetLag != that.offsetLag {
		return false
	}
	if o.isFutureKey != that.isFutureKey {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *DescribeLogDirsPartition) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	partitionIndexField := fields.Int32{Context: describeLogDirsPartitionPartitionIndex}
	fieldSize, err = partitionIndexField.SizeInBytes(version, o.partitionIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionIndex\" field")
	}
	size += fieldSize

	partitionSizeField := fields.Int64{Context: describeLogDirsPartitionPartitionSize}
	fieldSize, err = partitionSizeField.SizeInBytes(version, o.partitionSize)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionSize\" field")
	}
	size += fieldSize

	offsetLagField := fields.Int64{Context: describeLogDirsPartitionOffsetLag}
	fieldSize, err = offsetLagField.SizeInBytes(version, o.offsetLag)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"offsetLag\" field")
	}
	size += fieldSize

	isFutureKeyField := fields.Bool{Context: describeLogDirsPartitionIsFutureKey}
	fieldSize, err = isFutureKeyField.SizeInBytes(version, o.isFutureKey)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"isFutureKey\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < DescribeLogDirsPartitionLowestSupportedFlexVersion() || version > DescribeLogDirsPartitionHighestSupportedFlexVersion() {
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
func (o *DescribeLogDirsPartition) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *DescribeLogDirsPartition) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *DescribeLogDirsPartition) validateNonIgnorableFields(version int16) error {
	return nil
}

func DescribeLogDirsPartitionLowestSupportedVersion() int16 {
	return 0
}

func DescribeLogDirsPartitionHighestSupportedVersion() int16 {
	return 32767
}

func DescribeLogDirsPartitionLowestSupportedFlexVersion() int16 {
	return 2
}

func DescribeLogDirsPartitionHighestSupportedFlexVersion() int16 {
	return 32767
}

func DescribeLogDirsPartitionDefault() DescribeLogDirsPartition {
	var d DescribeLogDirsPartition
	d.SetDefault()

	return d
}
