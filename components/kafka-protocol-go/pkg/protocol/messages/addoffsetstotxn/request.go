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
package addoffsetstotxn

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var requestTransactionalId = fields.Context{
	SpecName:                    "TransactionalId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var requestProducerId = fields.Context{
	SpecName:                    "ProducerId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var requestProducerEpoch = fields.Context{
	SpecName:                    "ProducerEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var requestGroupId = fields.Context{
	SpecName:                    "GroupId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}

type Request struct {
	unknownTaggedFields []fields.RawTaggedField
	transactionalId     fields.NullableString
	groupId             fields.NullableString
	producerId          int64
	producerEpoch       int16
	isNil               bool
}

func (o *Request) TransactionalId() fields.NullableString {
	return o.transactionalId
}

func (o *Request) SetTransactionalId(val fields.NullableString) {
	o.isNil = false
	o.transactionalId = val
}

func (o *Request) ProducerId() int64 {
	return o.producerId
}

func (o *Request) SetProducerId(val int64) {
	o.isNil = false
	o.producerId = val
}

func (o *Request) ProducerEpoch() int16 {
	return o.producerEpoch
}

func (o *Request) SetProducerEpoch(val int16) {
	o.isNil = false
	o.producerEpoch = val
}

func (o *Request) GroupId() fields.NullableString {
	return o.groupId
}

func (o *Request) SetGroupId(val fields.NullableString) {
	o.isNil = false
	o.groupId = val
}

func (o *Request) ApiKey() int16 {
	return 25
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

	producerIdField := fields.Int64{Context: requestProducerId}
	if err := producerIdField.Read(buf, version, &o.producerId); err != nil {
		return errors.WrapIf(err, "couldn't set \"producerId\" field")
	}

	producerEpochField := fields.Int16{Context: requestProducerEpoch}
	if err := producerEpochField.Read(buf, version, &o.producerEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"producerEpoch\" field")
	}

	groupIdField := fields.String{Context: requestGroupId}
	if err := groupIdField.Read(buf, version, &o.groupId); err != nil {
		return errors.WrapIf(err, "couldn't set \"groupId\" field")
	}

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
	producerIdField := fields.Int64{Context: requestProducerId}
	if err := producerIdField.Write(buf, version, o.producerId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"producerId\" field")
	}
	producerEpochField := fields.Int16{Context: requestProducerEpoch}
	if err := producerEpochField.Write(buf, version, o.producerEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"producerEpoch\" field")
	}
	groupIdField := fields.String{Context: requestGroupId}
	if err := groupIdField.Write(buf, version, o.groupId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"groupId\" field")
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
	if b, err := fields.MarshalPrimitiveTypeJSON(o.producerId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"producerId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.producerEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"producerEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.groupId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"groupId\""), b}, []byte(": ")))
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

	o.unknownTaggedFields = nil
}

func (o *Request) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.transactionalId.SetValue("")
	o.producerId = 0
	o.producerEpoch = 0
	o.groupId.SetValue("")

	o.isNil = false
}

func (o *Request) Equal(that *Request) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.transactionalId.Equal(&that.transactionalId) {
		return false
	}
	if o.producerId != that.producerId {
		return false
	}
	if o.producerEpoch != that.producerEpoch {
		return false
	}
	if !o.groupId.Equal(&that.groupId) {
		return false
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

	producerIdField := fields.Int64{Context: requestProducerId}
	fieldSize, err = producerIdField.SizeInBytes(version, o.producerId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"producerId\" field")
	}
	size += fieldSize

	producerEpochField := fields.Int16{Context: requestProducerEpoch}
	fieldSize, err = producerEpochField.SizeInBytes(version, o.producerEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"producerEpoch\" field")
	}
	size += fieldSize

	groupIdField := fields.String{Context: requestGroupId}
	fieldSize, err = groupIdField.SizeInBytes(version, o.groupId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"groupId\" field")
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
	o.groupId.Release()
}

func (o *Request) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Request) validateNonIgnorableFields(version int16) error {
	return nil
}

func RequestLowestSupportedVersion() int16 {
	return 0
}

func RequestHighestSupportedVersion() int16 {
	return 3
}

func RequestLowestSupportedFlexVersion() int16 {
	return 3
}

func RequestHighestSupportedFlexVersion() int16 {
	return 32767
}

func RequestDefault() Request {
	var d Request
	d.SetDefault()

	return d
}
