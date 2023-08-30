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
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/leaderandisr/common"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var leaderAndIsrTopicStateTopicName = fields.Context{
	SpecName:                    "TopicName",
	LowestSupportedVersion:      2,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}
var leaderAndIsrTopicStateTopicId = fields.Context{
	SpecName:                    "TopicId",
	LowestSupportedVersion:      5,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}
var leaderAndIsrTopicStatePartitionStates = fields.Context{
	SpecName:                    "PartitionStates",
	LowestSupportedVersion:      2,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}

type LeaderAndIsrTopicState struct {
	partitionStates     []common.LeaderAndIsrPartitionState
	unknownTaggedFields []fields.RawTaggedField
	topicName           fields.NullableString
	topicId             fields.UUID
	isNil               bool
}

func (o *LeaderAndIsrTopicState) TopicName() fields.NullableString {
	return o.topicName
}

func (o *LeaderAndIsrTopicState) SetTopicName(val fields.NullableString) {
	o.isNil = false
	o.topicName = val
}

func (o *LeaderAndIsrTopicState) TopicId() fields.UUID {
	return o.topicId
}

func (o *LeaderAndIsrTopicState) SetTopicId(val fields.UUID) {
	o.isNil = false
	o.topicId = val
}

func (o *LeaderAndIsrTopicState) PartitionStates() []common.LeaderAndIsrPartitionState {
	return o.partitionStates
}

func (o *LeaderAndIsrTopicState) SetPartitionStates(val []common.LeaderAndIsrPartitionState) {
	o.isNil = false
	o.partitionStates = val
}

func (o *LeaderAndIsrTopicState) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *LeaderAndIsrTopicState) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *LeaderAndIsrTopicState) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	topicNameField := fields.String{Context: leaderAndIsrTopicStateTopicName}
	if err := topicNameField.Read(buf, version, &o.topicName); err != nil {
		return errors.WrapIf(err, "couldn't set \"topicName\" field")
	}

	topicIdField := fields.Uuid{Context: leaderAndIsrTopicStateTopicId}
	if err := topicIdField.Read(buf, version, &o.topicId); err != nil {
		return errors.WrapIf(err, "couldn't set \"topicId\" field")
	}

	partitionStatesField := fields.ArrayOfStruct[common.LeaderAndIsrPartitionState, *common.LeaderAndIsrPartitionState]{Context: leaderAndIsrTopicStatePartitionStates}
	partitionStates, err := partitionStatesField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionStates\" field")
	}
	o.partitionStates = partitionStates

	// process tagged fields

	if version < LeaderAndIsrTopicStateLowestSupportedFlexVersion() || version > LeaderAndIsrTopicStateHighestSupportedFlexVersion() {
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

func (o *LeaderAndIsrTopicState) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	topicNameField := fields.String{Context: leaderAndIsrTopicStateTopicName}
	if err := topicNameField.Write(buf, version, o.topicName); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topicName\" field")
	}
	topicIdField := fields.Uuid{Context: leaderAndIsrTopicStateTopicId}
	if err := topicIdField.Write(buf, version, o.topicId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topicId\" field")
	}

	partitionStatesField := fields.ArrayOfStruct[common.LeaderAndIsrPartitionState, *common.LeaderAndIsrPartitionState]{Context: leaderAndIsrTopicStatePartitionStates}
	if err := partitionStatesField.Write(buf, version, o.partitionStates); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionStates\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < LeaderAndIsrTopicStateLowestSupportedFlexVersion() || version > LeaderAndIsrTopicStateHighestSupportedFlexVersion() {
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

func (o *LeaderAndIsrTopicState) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *LeaderAndIsrTopicState) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 4)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.topicName); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"topicName\""), b}, []byte(": ")))
	}
	if b, err := fields.BytesMarshalJSON("topicId", o.topicId[:]); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.ArrayOfStructMarshalJSON("partitionStates", o.partitionStates); err != nil {
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

func (o *LeaderAndIsrTopicState) IsNil() bool {
	return o.isNil
}

func (o *LeaderAndIsrTopicState) Clear() {
	o.Release()
	o.isNil = true

	o.partitionStates = nil
	o.unknownTaggedFields = nil
}

func (o *LeaderAndIsrTopicState) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.topicName.SetValue("")
	o.topicId.SetZero()
	for i := range o.partitionStates {
		o.partitionStates[i].Release()
	}
	o.partitionStates = nil

	o.isNil = false
}

func (o *LeaderAndIsrTopicState) Equal(that *LeaderAndIsrTopicState) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.topicName.Equal(&that.topicName) {
		return false
	}
	if o.topicId != that.topicId {
		return false
	}
	if len(o.partitionStates) != len(that.partitionStates) {
		return false
	}
	for i := range o.partitionStates {
		if !o.partitionStates[i].Equal(&that.partitionStates[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *LeaderAndIsrTopicState) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	topicNameField := fields.String{Context: leaderAndIsrTopicStateTopicName}
	fieldSize, err = topicNameField.SizeInBytes(version, o.topicName)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topicName\" field")
	}
	size += fieldSize

	topicIdField := fields.Uuid{Context: leaderAndIsrTopicStateTopicId}
	fieldSize, err = topicIdField.SizeInBytes(version, o.topicId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topicId\" field")
	}
	size += fieldSize

	partitionStatesField := fields.ArrayOfStruct[common.LeaderAndIsrPartitionState, *common.LeaderAndIsrPartitionState]{Context: leaderAndIsrTopicStatePartitionStates}
	fieldSize, err = partitionStatesField.SizeInBytes(version, o.partitionStates)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionStates\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < LeaderAndIsrTopicStateLowestSupportedFlexVersion() || version > LeaderAndIsrTopicStateHighestSupportedFlexVersion() {
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
func (o *LeaderAndIsrTopicState) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.topicName.Release()
	for i := range o.partitionStates {
		o.partitionStates[i].Release()
	}
	o.partitionStates = nil
}

func (o *LeaderAndIsrTopicState) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *LeaderAndIsrTopicState) validateNonIgnorableFields(version int16) error {
	if !leaderAndIsrTopicStateTopicName.IsSupportedVersion(version) {
		if o.topicName.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"topicName\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !leaderAndIsrTopicStatePartitionStates.IsSupportedVersion(version) {
		if len(o.partitionStates) > 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"partitionStates\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func LeaderAndIsrTopicStateLowestSupportedVersion() int16 {
	return 2
}

func LeaderAndIsrTopicStateHighestSupportedVersion() int16 {
	return 32767
}

func LeaderAndIsrTopicStateLowestSupportedFlexVersion() int16 {
	return 4
}

func LeaderAndIsrTopicStateHighestSupportedFlexVersion() int16 {
	return 32767
}

func LeaderAndIsrTopicStateDefault() LeaderAndIsrTopicState {
	var d LeaderAndIsrTopicState
	d.SetDefault()

	return d
}
