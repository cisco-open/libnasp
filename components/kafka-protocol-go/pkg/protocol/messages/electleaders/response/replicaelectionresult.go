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
package response

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/electleaders/response/replicaelectionresult"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var replicaElectionResultTopic = fields.Context{
	SpecName:                    "Topic",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var replicaElectionResultPartitionResult = fields.Context{
	SpecName:                    "PartitionResult",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type ReplicaElectionResult struct {
	partitionResult     []replicaelectionresult.PartitionResult
	unknownTaggedFields []fields.RawTaggedField
	topic               fields.NullableString
	isNil               bool
}

func (o *ReplicaElectionResult) Topic() fields.NullableString {
	return o.topic
}

func (o *ReplicaElectionResult) SetTopic(val fields.NullableString) {
	o.isNil = false
	o.topic = val
}

func (o *ReplicaElectionResult) PartitionResult() []replicaelectionresult.PartitionResult {
	return o.partitionResult
}

func (o *ReplicaElectionResult) SetPartitionResult(val []replicaelectionresult.PartitionResult) {
	o.isNil = false
	o.partitionResult = val
}

func (o *ReplicaElectionResult) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *ReplicaElectionResult) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *ReplicaElectionResult) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	topicField := fields.String{Context: replicaElectionResultTopic}
	if err := topicField.Read(buf, version, &o.topic); err != nil {
		return errors.WrapIf(err, "couldn't set \"topic\" field")
	}

	partitionResultField := fields.ArrayOfStruct[replicaelectionresult.PartitionResult, *replicaelectionresult.PartitionResult]{Context: replicaElectionResultPartitionResult}
	partitionResult, err := partitionResultField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"partitionResult\" field")
	}
	o.partitionResult = partitionResult

	// process tagged fields

	if version < ReplicaElectionResultLowestSupportedFlexVersion() || version > ReplicaElectionResultHighestSupportedFlexVersion() {
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

func (o *ReplicaElectionResult) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	topicField := fields.String{Context: replicaElectionResultTopic}
	if err := topicField.Write(buf, version, o.topic); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topic\" field")
	}

	partitionResultField := fields.ArrayOfStruct[replicaelectionresult.PartitionResult, *replicaelectionresult.PartitionResult]{Context: replicaElectionResultPartitionResult}
	if err := partitionResultField.Write(buf, version, o.partitionResult); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"partitionResult\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ReplicaElectionResultLowestSupportedFlexVersion() || version > ReplicaElectionResultHighestSupportedFlexVersion() {
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

func (o *ReplicaElectionResult) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *ReplicaElectionResult) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.topic); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"topic\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("partitionResult", o.partitionResult); err != nil {
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

func (o *ReplicaElectionResult) IsNil() bool {
	return o.isNil
}

func (o *ReplicaElectionResult) Clear() {
	o.Release()
	o.isNil = true

	o.partitionResult = nil
	o.unknownTaggedFields = nil
}

func (o *ReplicaElectionResult) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.topic.SetValue("")
	for i := range o.partitionResult {
		o.partitionResult[i].Release()
	}
	o.partitionResult = nil

	o.isNil = false
}

func (o *ReplicaElectionResult) Equal(that *ReplicaElectionResult) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.topic.Equal(&that.topic) {
		return false
	}
	if len(o.partitionResult) != len(that.partitionResult) {
		return false
	}
	for i := range o.partitionResult {
		if !o.partitionResult[i].Equal(&that.partitionResult[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *ReplicaElectionResult) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	topicField := fields.String{Context: replicaElectionResultTopic}
	fieldSize, err = topicField.SizeInBytes(version, o.topic)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topic\" field")
	}
	size += fieldSize

	partitionResultField := fields.ArrayOfStruct[replicaelectionresult.PartitionResult, *replicaelectionresult.PartitionResult]{Context: replicaElectionResultPartitionResult}
	fieldSize, err = partitionResultField.SizeInBytes(version, o.partitionResult)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"partitionResult\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < ReplicaElectionResultLowestSupportedFlexVersion() || version > ReplicaElectionResultHighestSupportedFlexVersion() {
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
func (o *ReplicaElectionResult) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.topic.Release()
	for i := range o.partitionResult {
		o.partitionResult[i].Release()
	}
	o.partitionResult = nil
}

func (o *ReplicaElectionResult) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *ReplicaElectionResult) validateNonIgnorableFields(version int16) error {
	return nil
}

func ReplicaElectionResultLowestSupportedVersion() int16 {
	return 0
}

func ReplicaElectionResultHighestSupportedVersion() int16 {
	return 32767
}

func ReplicaElectionResultLowestSupportedFlexVersion() int16 {
	return 2
}

func ReplicaElectionResultHighestSupportedFlexVersion() int16 {
	return 32767
}

func ReplicaElectionResultDefault() ReplicaElectionResult {
	var d ReplicaElectionResult
	d.SetDefault()

	return d
}
