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
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var describableLogDirTopicTopic = fields.Context{
	SpecName:                        "Topic",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}
var describableLogDirTopicPartitions = fields.Context{
	SpecName:                        "Partitions",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}

type DescribableLogDirTopic struct {
	partitions          []int32
	unknownTaggedFields []fields.RawTaggedField
	topic               fields.NullableString
	isNil               bool
}

func (o *DescribableLogDirTopic) Topic() fields.NullableString {
	return o.topic
}

func (o *DescribableLogDirTopic) SetTopic(val fields.NullableString) {
	o.isNil = false
	o.topic = val
}

func (o *DescribableLogDirTopic) Partitions() []int32 {
	return o.partitions
}

func (o *DescribableLogDirTopic) SetPartitions(val []int32) {
	o.isNil = false
	o.partitions = val
}

func (o *DescribableLogDirTopic) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *DescribableLogDirTopic) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *DescribableLogDirTopic) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	topicField := fields.String{Context: describableLogDirTopicTopic}
	if err := topicField.Read(buf, version, &o.topic); err != nil {
		return errors.WrapIf(err, "couldn't set \"topic\" field")
	}

	partitionsField := fields.Array[int32, *fields.Int32]{
		Context:          describableLogDirTopicPartitions,
		ElementProcessor: &fields.Int32{Context: describableLogDirTopicPartitions}}

	partitions, err := partitionsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"partitions\" field")
	}
	o.partitions = partitions

	// process tagged fields

	if version < DescribableLogDirTopicLowestSupportedFlexVersion() || version > DescribableLogDirTopicHighestSupportedFlexVersion() {
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

func (o *DescribableLogDirTopic) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	topicField := fields.String{Context: describableLogDirTopicTopic}
	if err := topicField.Write(buf, version, o.topic); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topic\" field")
	}
	partitionsField := fields.Array[int32, *fields.Int32]{
		Context:          describableLogDirTopicPartitions,
		ElementProcessor: &fields.Int32{Context: describableLogDirTopicPartitions}}
	if err := partitionsField.Write(buf, version, o.partitions); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitions\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < DescribableLogDirTopicLowestSupportedFlexVersion() || version > DescribableLogDirTopicHighestSupportedFlexVersion() {
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

func (o *DescribableLogDirTopic) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *DescribableLogDirTopic) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.topic); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"topic\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayMarshalJSON("partitions", o.partitions); err != nil {
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

func (o *DescribableLogDirTopic) IsNil() bool {
	return o.isNil
}

func (o *DescribableLogDirTopic) Clear() {
	o.Release()
	o.isNil = true

	o.partitions = nil
	o.unknownTaggedFields = nil
}

func (o *DescribableLogDirTopic) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.topic.SetValue("")
	o.partitions = nil

	o.isNil = false
}

func (o *DescribableLogDirTopic) Equal(that *DescribableLogDirTopic) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.topic.Equal(&that.topic) {
		return false
	}
	if !fields.PrimitiveTypeSliceEqual(o.partitions, that.partitions) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *DescribableLogDirTopic) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	topicField := fields.String{Context: describableLogDirTopicTopic}
	fieldSize, err = topicField.SizeInBytes(version, o.topic)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topic\" field")
	}
	size += fieldSize

	partitionsField := fields.Array[int32, *fields.Int32]{
		Context:          describableLogDirTopicPartitions,
		ElementProcessor: &fields.Int32{Context: describableLogDirTopicPartitions}}
	fieldSize, err = partitionsField.SizeInBytes(version, o.partitions)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitions\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < DescribableLogDirTopicLowestSupportedFlexVersion() || version > DescribableLogDirTopicHighestSupportedFlexVersion() {
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
func (o *DescribableLogDirTopic) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.topic.Release()
	o.partitions = nil
}

func (o *DescribableLogDirTopic) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *DescribableLogDirTopic) validateNonIgnorableFields(version int16) error {
	return nil
}

func DescribableLogDirTopicLowestSupportedVersion() int16 {
	return 0
}

func DescribableLogDirTopicHighestSupportedVersion() int16 {
	return 32767
}

func DescribableLogDirTopicLowestSupportedFlexVersion() int16 {
	return 2
}

func DescribableLogDirTopicHighestSupportedFlexVersion() int16 {
	return 32767
}

func DescribableLogDirTopicDefault() DescribableLogDirTopic {
	var d DescribableLogDirTopic
	d.SetDefault()

	return d
}
