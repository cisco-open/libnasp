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
package brokerheartbeat

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var responseThrottleTimeMs = fields.Context{
	SpecName:                    "ThrottleTimeMs",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var responseErrorCode = fields.Context{
	SpecName:                    "ErrorCode",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var responseIsCaughtUp = fields.Context{
	SpecName:                    "IsCaughtUp",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var responseIsFenced = fields.Context{
	SpecName:                    "IsFenced",
	CustomDefaultValue:          bool(true),
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var responseShouldShutDown = fields.Context{
	SpecName:                    "ShouldShutDown",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}

type Response struct {
	unknownTaggedFields []fields.RawTaggedField
	throttleTimeMs      int32
	errorCode           int16
	isCaughtUp          bool
	isFenced            bool
	shouldShutDown      bool
	isNil               bool
}

func (o *Response) ThrottleTimeMs() int32 {
	return o.throttleTimeMs
}

func (o *Response) SetThrottleTimeMs(val int32) {
	o.isNil = false
	o.throttleTimeMs = val
}

func (o *Response) ErrorCode() int16 {
	return o.errorCode
}

func (o *Response) SetErrorCode(val int16) {
	o.isNil = false
	o.errorCode = val
}

func (o *Response) IsCaughtUp() bool {
	return o.isCaughtUp
}

func (o *Response) SetIsCaughtUp(val bool) {
	o.isNil = false
	o.isCaughtUp = val
}

func (o *Response) IsFenced() bool {
	return o.isFenced
}

func (o *Response) SetIsFenced(val bool) {
	o.isNil = false
	o.isFenced = val
}

func (o *Response) ShouldShutDown() bool {
	return o.shouldShutDown
}

func (o *Response) SetShouldShutDown(val bool) {
	o.isNil = false
	o.shouldShutDown = val
}

func (o *Response) ApiKey() int16 {
	return 63
}

func (o *Response) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Response) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Response) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	throttleTimeMsField := fields.Int32{Context: responseThrottleTimeMs}
	if err := throttleTimeMsField.Read(buf, version, &o.throttleTimeMs); err != nil {
		return errors.WrapIf(err, "couldn't set \"throttleTimeMs\" field")
	}

	errorCodeField := fields.Int16{Context: responseErrorCode}
	if err := errorCodeField.Read(buf, version, &o.errorCode); err != nil {
		return errors.WrapIf(err, "couldn't set \"errorCode\" field")
	}

	isCaughtUpField := fields.Bool{Context: responseIsCaughtUp}
	if err := isCaughtUpField.Read(buf, version, &o.isCaughtUp); err != nil {
		return errors.WrapIf(err, "couldn't set \"isCaughtUp\" field")
	}

	isFencedField := fields.Bool{Context: responseIsFenced}
	if err := isFencedField.Read(buf, version, &o.isFenced); err != nil {
		return errors.WrapIf(err, "couldn't set \"isFenced\" field")
	}

	shouldShutDownField := fields.Bool{Context: responseShouldShutDown}
	if err := shouldShutDownField.Read(buf, version, &o.shouldShutDown); err != nil {
		return errors.WrapIf(err, "couldn't set \"shouldShutDown\" field")
	}

	// process tagged fields

	if version < ResponseLowestSupportedFlexVersion() || version > ResponseHighestSupportedFlexVersion() {
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

func (o *Response) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	throttleTimeMsField := fields.Int32{Context: responseThrottleTimeMs}
	if err := throttleTimeMsField.Write(buf, version, o.throttleTimeMs); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"throttleTimeMs\" field")
	}
	errorCodeField := fields.Int16{Context: responseErrorCode}
	if err := errorCodeField.Write(buf, version, o.errorCode); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"errorCode\" field")
	}
	isCaughtUpField := fields.Bool{Context: responseIsCaughtUp}
	if err := isCaughtUpField.Write(buf, version, o.isCaughtUp); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"isCaughtUp\" field")
	}
	isFencedField := fields.Bool{Context: responseIsFenced}
	if err := isFencedField.Write(buf, version, o.isFenced); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"isFenced\" field")
	}
	shouldShutDownField := fields.Bool{Context: responseShouldShutDown}
	if err := shouldShutDownField.Write(buf, version, o.shouldShutDown); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"shouldShutDown\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ResponseLowestSupportedFlexVersion() || version > ResponseHighestSupportedFlexVersion() {
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

func (o *Response) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *Response) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 6)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.throttleTimeMs); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"throttleTimeMs\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.errorCode); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"errorCode\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.isCaughtUp); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"isCaughtUp\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.isFenced); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"isFenced\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.shouldShutDown); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"shouldShutDown\""), b}, []byte(": ")))
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

func (o *Response) IsNil() bool {
	return o.isNil
}

func (o *Response) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *Response) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.throttleTimeMs = 0
	o.errorCode = 0
	o.isCaughtUp = false
	o.isFenced = true
	o.shouldShutDown = false

	o.isNil = false
}

func (o *Response) Equal(that *Response) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.throttleTimeMs != that.throttleTimeMs {
		return false
	}
	if o.errorCode != that.errorCode {
		return false
	}
	if o.isCaughtUp != that.isCaughtUp {
		return false
	}
	if o.isFenced != that.isFenced {
		return false
	}
	if o.shouldShutDown != that.shouldShutDown {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *Response) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	throttleTimeMsField := fields.Int32{Context: responseThrottleTimeMs}
	fieldSize, err = throttleTimeMsField.SizeInBytes(version, o.throttleTimeMs)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"throttleTimeMs\" field")
	}
	size += fieldSize

	errorCodeField := fields.Int16{Context: responseErrorCode}
	fieldSize, err = errorCodeField.SizeInBytes(version, o.errorCode)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"errorCode\" field")
	}
	size += fieldSize

	isCaughtUpField := fields.Bool{Context: responseIsCaughtUp}
	fieldSize, err = isCaughtUpField.SizeInBytes(version, o.isCaughtUp)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"isCaughtUp\" field")
	}
	size += fieldSize

	isFencedField := fields.Bool{Context: responseIsFenced}
	fieldSize, err = isFencedField.SizeInBytes(version, o.isFenced)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"isFenced\" field")
	}
	size += fieldSize

	shouldShutDownField := fields.Bool{Context: responseShouldShutDown}
	fieldSize, err = shouldShutDownField.SizeInBytes(version, o.shouldShutDown)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"shouldShutDown\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < ResponseLowestSupportedFlexVersion() || version > ResponseHighestSupportedFlexVersion() {
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
func (o *Response) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *Response) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Response) validateNonIgnorableFields(version int16) error {
	return nil
}

func ResponseLowestSupportedVersion() int16 {
	return 0
}

func ResponseHighestSupportedVersion() int16 {
	return 0
}

func ResponseLowestSupportedFlexVersion() int16 {
	return 0
}

func ResponseHighestSupportedFlexVersion() int16 {
	return 32767
}

func ResponseDefault() Response {
	var d Response
	d.SetDefault()

	return d
}
