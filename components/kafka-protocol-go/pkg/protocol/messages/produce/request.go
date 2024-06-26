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
package produce

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/produce/request"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var requestTransactionalId = fields.Context{
	SpecName:           "TransactionalId",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          3,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      9,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  3,
	HighestSupportedNullableVersion: 32767,
}
var requestAcks = fields.Context{
	SpecName:                    "Acks",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var requestTimeoutMs = fields.Context{
	SpecName:                    "TimeoutMs",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var requestTopicData = fields.Context{
	SpecName:                    "TopicData",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}

type Request struct {
	topicData           []request.TopicProduceData
	unknownTaggedFields []fields.RawTaggedField
	transactionalId     fields.NullableString
	timeoutMs           int32
	acks                int16
	isNil               bool
}

func (o *Request) TransactionalId() fields.NullableString {
	return o.transactionalId
}

func (o *Request) SetTransactionalId(val fields.NullableString) {
	o.isNil = false
	o.transactionalId = val
}

func (o *Request) Acks() int16 {
	return o.acks
}

func (o *Request) SetAcks(val int16) {
	o.isNil = false
	o.acks = val
}

func (o *Request) TimeoutMs() int32 {
	return o.timeoutMs
}

func (o *Request) SetTimeoutMs(val int32) {
	o.isNil = false
	o.timeoutMs = val
}

func (o *Request) TopicData() []request.TopicProduceData {
	return o.topicData
}

func (o *Request) SetTopicData(val []request.TopicProduceData) {
	o.isNil = false
	o.topicData = val
}

func (o *Request) ApiKey() int16 {
	return 0
}

func (o *Request) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Request) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Request) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	transactionalIdField := fields.String{Context: requestTransactionalId}
	if err := transactionalIdField.Read(buf, version, &o.transactionalId); err != nil {
		return errors.WrapIf(err, "couldn't set \"transactionalId\" field")
	}

	acksField := fields.Int16{Context: requestAcks}
	if err := acksField.Read(buf, version, &o.acks); err != nil {
		return errors.WrapIf(err, "couldn't set \"acks\" field")
	}

	timeoutMsField := fields.Int32{Context: requestTimeoutMs}
	if err := timeoutMsField.Read(buf, version, &o.timeoutMs); err != nil {
		return errors.WrapIf(err, "couldn't set \"timeoutMs\" field")
	}

	topicDataField := fields.ArrayOfStruct[request.TopicProduceData, *request.TopicProduceData]{Context: requestTopicData}
	topicData, err := topicDataField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"topicData\" field")
	}
	o.topicData = topicData

	// process tagged fields

	if version < RequestLowestSupportedFlexVersion() || version > RequestHighestSupportedFlexVersion() {
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

func (o *Request) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	transactionalIdField := fields.String{Context: requestTransactionalId}
	if err := transactionalIdField.Write(buf, version, o.transactionalId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"transactionalId\" field")
	}
	acksField := fields.Int16{Context: requestAcks}
	if err := acksField.Write(buf, version, o.acks); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"acks\" field")
	}
	timeoutMsField := fields.Int32{Context: requestTimeoutMs}
	if err := timeoutMsField.Write(buf, version, o.timeoutMs); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"timeoutMs\" field")
	}

	topicDataField := fields.ArrayOfStruct[request.TopicProduceData, *request.TopicProduceData]{Context: requestTopicData}
	if err := topicDataField.Write(buf, version, o.topicData); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topicData\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < RequestLowestSupportedFlexVersion() || version > RequestHighestSupportedFlexVersion() {
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

func (o *Request) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *Request) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.transactionalId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"transactionalId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.acks); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"acks\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.timeoutMs); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"timeoutMs\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("topicData", o.topicData); err != nil {
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

func (o *Request) IsNil() bool {
	return o.isNil
}

func (o *Request) Clear() {
	o.Release()
	o.isNil = true

	o.topicData = nil
	o.unknownTaggedFields = nil
}

func (o *Request) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.transactionalId.SetValue("null")
	o.acks = 0
	o.timeoutMs = 0
	for i := range o.topicData {
		o.topicData[i].Release()
	}
	o.topicData = nil

	o.isNil = false
}

func (o *Request) Equal(that *Request) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.transactionalId.Equal(&that.transactionalId) {
		return false
	}
	if o.acks != that.acks {
		return false
	}
	if o.timeoutMs != that.timeoutMs {
		return false
	}
	if len(o.topicData) != len(that.topicData) {
		return false
	}
	for i := range o.topicData {
		if !o.topicData[i].Equal(&that.topicData[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *Request) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	transactionalIdField := fields.String{Context: requestTransactionalId}
	fieldSize, err = transactionalIdField.SizeInBytes(version, o.transactionalId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"transactionalId\" field")
	}
	size += fieldSize

	acksField := fields.Int16{Context: requestAcks}
	fieldSize, err = acksField.SizeInBytes(version, o.acks)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"acks\" field")
	}
	size += fieldSize

	timeoutMsField := fields.Int32{Context: requestTimeoutMs}
	fieldSize, err = timeoutMsField.SizeInBytes(version, o.timeoutMs)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"timeoutMs\" field")
	}
	size += fieldSize

	topicDataField := fields.ArrayOfStruct[request.TopicProduceData, *request.TopicProduceData]{Context: requestTopicData}
	fieldSize, err = topicDataField.SizeInBytes(version, o.topicData)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topicData\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < RequestLowestSupportedFlexVersion() || version > RequestHighestSupportedFlexVersion() {
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
func (o *Request) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.transactionalId.Release()
	for i := range o.topicData {
		o.topicData[i].Release()
	}
	o.topicData = nil
}

func (o *Request) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Request) validateNonIgnorableFields(version int16) error {
	if !requestTransactionalId.IsSupportedVersion(version) {
		if !o.transactionalId.IsNil() {
			return errors.New(strings.Join([]string{"attempted to write non-default \"transactionalId\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func RequestLowestSupportedVersion() int16 {
	return 0
}

func RequestHighestSupportedVersion() int16 {
	return 9
}

func RequestLowestSupportedFlexVersion() int16 {
	return 9
}

func RequestHighestSupportedFlexVersion() int16 {
	return 32767
}

func RequestDefault() Request {
	var d Request
	d.SetDefault()

	return d
}
