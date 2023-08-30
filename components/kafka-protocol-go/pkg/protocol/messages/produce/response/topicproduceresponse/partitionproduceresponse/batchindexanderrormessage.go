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
package partitionproduceresponse

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var batchIndexAndErrorMessageBatchIndex = fields.Context{
	SpecName:                    "BatchIndex",
	LowestSupportedVersion:      8,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var batchIndexAndErrorMessageBatchIndexErrorMessage = fields.Context{
	SpecName:           "BatchIndexErrorMessage",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          8,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      9,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  8,
	HighestSupportedNullableVersion: 32767,
}

type BatchIndexAndErrorMessage struct {
	unknownTaggedFields    []fields.RawTaggedField
	batchIndexErrorMessage fields.NullableString
	batchIndex             int32
	isNil                  bool
}

func (o *BatchIndexAndErrorMessage) BatchIndex() int32 {
	return o.batchIndex
}

func (o *BatchIndexAndErrorMessage) SetBatchIndex(val int32) {
	o.isNil = false
	o.batchIndex = val
}

func (o *BatchIndexAndErrorMessage) BatchIndexErrorMessage() fields.NullableString {
	return o.batchIndexErrorMessage
}

func (o *BatchIndexAndErrorMessage) SetBatchIndexErrorMessage(val fields.NullableString) {
	o.isNil = false
	o.batchIndexErrorMessage = val
}

func (o *BatchIndexAndErrorMessage) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *BatchIndexAndErrorMessage) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *BatchIndexAndErrorMessage) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	batchIndexField := fields.Int32{Context: batchIndexAndErrorMessageBatchIndex}
	if err := batchIndexField.Read(buf, version, &o.batchIndex); err != nil {
		return errors.WrapIf(err, "couldn't set \"batchIndex\" field")
	}

	batchIndexErrorMessageField := fields.String{Context: batchIndexAndErrorMessageBatchIndexErrorMessage}
	if err := batchIndexErrorMessageField.Read(buf, version, &o.batchIndexErrorMessage); err != nil {
		return errors.WrapIf(err, "couldn't set \"batchIndexErrorMessage\" field")
	}

	// process tagged fields

	if version < BatchIndexAndErrorMessageLowestSupportedFlexVersion() || version > BatchIndexAndErrorMessageHighestSupportedFlexVersion() {
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

func (o *BatchIndexAndErrorMessage) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	batchIndexField := fields.Int32{Context: batchIndexAndErrorMessageBatchIndex}
	if err := batchIndexField.Write(buf, version, o.batchIndex); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"batchIndex\" field")
	}
	batchIndexErrorMessageField := fields.String{Context: batchIndexAndErrorMessageBatchIndexErrorMessage}
	if err := batchIndexErrorMessageField.Write(buf, version, o.batchIndexErrorMessage); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"batchIndexErrorMessage\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < BatchIndexAndErrorMessageLowestSupportedFlexVersion() || version > BatchIndexAndErrorMessageHighestSupportedFlexVersion() {
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

func (o *BatchIndexAndErrorMessage) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *BatchIndexAndErrorMessage) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.batchIndex); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"batchIndex\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.batchIndexErrorMessage); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"batchIndexErrorMessage\""), b}, []byte(": ")))
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

func (o *BatchIndexAndErrorMessage) IsNil() bool {
	return o.isNil
}

func (o *BatchIndexAndErrorMessage) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *BatchIndexAndErrorMessage) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.batchIndex = 0
	o.batchIndexErrorMessage.SetValue("null")

	o.isNil = false
}

func (o *BatchIndexAndErrorMessage) Equal(that *BatchIndexAndErrorMessage) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.batchIndex != that.batchIndex {
		return false
	}
	if !o.batchIndexErrorMessage.Equal(&that.batchIndexErrorMessage) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *BatchIndexAndErrorMessage) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	batchIndexField := fields.Int32{Context: batchIndexAndErrorMessageBatchIndex}
	fieldSize, err = batchIndexField.SizeInBytes(version, o.batchIndex)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"batchIndex\" field")
	}
	size += fieldSize

	batchIndexErrorMessageField := fields.String{Context: batchIndexAndErrorMessageBatchIndexErrorMessage}
	fieldSize, err = batchIndexErrorMessageField.SizeInBytes(version, o.batchIndexErrorMessage)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"batchIndexErrorMessage\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < BatchIndexAndErrorMessageLowestSupportedFlexVersion() || version > BatchIndexAndErrorMessageHighestSupportedFlexVersion() {
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
func (o *BatchIndexAndErrorMessage) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.batchIndexErrorMessage.Release()
}

func (o *BatchIndexAndErrorMessage) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *BatchIndexAndErrorMessage) validateNonIgnorableFields(version int16) error {
	if !batchIndexAndErrorMessageBatchIndex.IsSupportedVersion(version) {
		if o.batchIndex != 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"batchIndex\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !batchIndexAndErrorMessageBatchIndexErrorMessage.IsSupportedVersion(version) {
		if !o.batchIndexErrorMessage.IsNil() {
			return errors.New(strings.Join([]string{"attempted to write non-default \"batchIndexErrorMessage\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func BatchIndexAndErrorMessageLowestSupportedVersion() int16 {
	return 8
}

func BatchIndexAndErrorMessageHighestSupportedVersion() int16 {
	return 32767
}

func BatchIndexAndErrorMessageLowestSupportedFlexVersion() int16 {
	return 9
}

func BatchIndexAndErrorMessageHighestSupportedFlexVersion() int16 {
	return 32767
}

func BatchIndexAndErrorMessageDefault() BatchIndexAndErrorMessage {
	var d BatchIndexAndErrorMessage
	d.SetDefault()

	return d
}
