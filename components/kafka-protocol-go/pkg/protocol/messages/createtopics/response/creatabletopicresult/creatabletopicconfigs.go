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
package creatabletopicresult

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var creatableTopicConfigsName = fields.Context{
	SpecName:                        "Name",
	LowestSupportedVersion:          5,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      5,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  5,
	HighestSupportedNullableVersion: 32767,
}
var creatableTopicConfigsValue = fields.Context{
	SpecName:                        "Value",
	LowestSupportedVersion:          5,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      5,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  5,
	HighestSupportedNullableVersion: 32767,
}
var creatableTopicConfigsReadOnly = fields.Context{
	SpecName:                        "ReadOnly",
	LowestSupportedVersion:          5,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      5,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  5,
	HighestSupportedNullableVersion: 32767,
}
var creatableTopicConfigsConfigSource = fields.Context{
	SpecName:                        "ConfigSource",
	CustomDefaultValue:              int8(-1),
	LowestSupportedVersion:          5,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      5,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  5,
	HighestSupportedNullableVersion: 32767,
}
var creatableTopicConfigsIsSensitive = fields.Context{
	SpecName:                        "IsSensitive",
	LowestSupportedVersion:          5,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      5,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  5,
	HighestSupportedNullableVersion: 32767,
}

type CreatableTopicConfigs struct {
	unknownTaggedFields []fields.RawTaggedField
	name                fields.NullableString
	value               fields.NullableString
	readOnly            bool
	configSource        int8
	isSensitive         bool
	isNil               bool
}

func (o *CreatableTopicConfigs) Name() fields.NullableString {
	return o.name
}

func (o *CreatableTopicConfigs) SetName(val fields.NullableString) {
	o.isNil = false
	o.name = val
}

func (o *CreatableTopicConfigs) Value() fields.NullableString {
	return o.value
}

func (o *CreatableTopicConfigs) SetValue(val fields.NullableString) {
	o.isNil = false
	o.value = val
}

func (o *CreatableTopicConfigs) ReadOnly() bool {
	return o.readOnly
}

func (o *CreatableTopicConfigs) SetReadOnly(val bool) {
	o.isNil = false
	o.readOnly = val
}

func (o *CreatableTopicConfigs) ConfigSource() int8 {
	return o.configSource
}

func (o *CreatableTopicConfigs) SetConfigSource(val int8) {
	o.isNil = false
	o.configSource = val
}

func (o *CreatableTopicConfigs) IsSensitive() bool {
	return o.isSensitive
}

func (o *CreatableTopicConfigs) SetIsSensitive(val bool) {
	o.isNil = false
	o.isSensitive = val
}

func (o *CreatableTopicConfigs) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *CreatableTopicConfigs) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *CreatableTopicConfigs) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	nameField := fields.String{Context: creatableTopicConfigsName}
	if err := nameField.Read(buf, version, &o.name); err != nil {
		return errors.WrapIf(err, "couldn't set \"name\" field")
	}

	valueField := fields.String{Context: creatableTopicConfigsValue}
	if err := valueField.Read(buf, version, &o.value); err != nil {
		return errors.WrapIf(err, "couldn't set \"value\" field")
	}

	readOnlyField := fields.Bool{Context: creatableTopicConfigsReadOnly}
	if err := readOnlyField.Read(buf, version, &o.readOnly); err != nil {
		return errors.WrapIf(err, "couldn't set \"readOnly\" field")
	}

	configSourceField := fields.Int8{Context: creatableTopicConfigsConfigSource}
	if err := configSourceField.Read(buf, version, &o.configSource); err != nil {
		return errors.WrapIf(err, "couldn't set \"configSource\" field")
	}

	isSensitiveField := fields.Bool{Context: creatableTopicConfigsIsSensitive}
	if err := isSensitiveField.Read(buf, version, &o.isSensitive); err != nil {
		return errors.WrapIf(err, "couldn't set \"isSensitive\" field")
	}

	// process tagged fields

	if version < CreatableTopicConfigsLowestSupportedFlexVersion() || version > CreatableTopicConfigsHighestSupportedFlexVersion() {
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

func (o *CreatableTopicConfigs) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	nameField := fields.String{Context: creatableTopicConfigsName}
	if err := nameField.Write(buf, version, o.name); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"name\" field")
	}
	valueField := fields.String{Context: creatableTopicConfigsValue}
	if err := valueField.Write(buf, version, o.value); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"value\" field")
	}
	readOnlyField := fields.Bool{Context: creatableTopicConfigsReadOnly}
	if err := readOnlyField.Write(buf, version, o.readOnly); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"readOnly\" field")
	}
	configSourceField := fields.Int8{Context: creatableTopicConfigsConfigSource}
	if err := configSourceField.Write(buf, version, o.configSource); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"configSource\" field")
	}
	isSensitiveField := fields.Bool{Context: creatableTopicConfigsIsSensitive}
	if err := isSensitiveField.Write(buf, version, o.isSensitive); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"isSensitive\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < CreatableTopicConfigsLowestSupportedFlexVersion() || version > CreatableTopicConfigsHighestSupportedFlexVersion() {
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

func (o *CreatableTopicConfigs) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *CreatableTopicConfigs) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 6)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.name); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"name\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.value); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"value\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.readOnly); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"readOnly\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.configSource); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"configSource\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.isSensitive); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"isSensitive\""), b}, []byte(": ")))
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

func (o *CreatableTopicConfigs) IsNil() bool {
	return o.isNil
}

func (o *CreatableTopicConfigs) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *CreatableTopicConfigs) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.name.SetValue("")
	o.value.SetValue("")
	o.readOnly = false
	o.configSource = -1
	o.isSensitive = false

	o.isNil = false
}

func (o *CreatableTopicConfigs) Equal(that *CreatableTopicConfigs) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.name.Equal(&that.name) {
		return false
	}
	if !o.value.Equal(&that.value) {
		return false
	}
	if o.readOnly != that.readOnly {
		return false
	}
	if o.configSource != that.configSource {
		return false
	}
	if o.isSensitive != that.isSensitive {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *CreatableTopicConfigs) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	nameField := fields.String{Context: creatableTopicConfigsName}
	fieldSize, err = nameField.SizeInBytes(version, o.name)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"name\" field")
	}
	size += fieldSize

	valueField := fields.String{Context: creatableTopicConfigsValue}
	fieldSize, err = valueField.SizeInBytes(version, o.value)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"value\" field")
	}
	size += fieldSize

	readOnlyField := fields.Bool{Context: creatableTopicConfigsReadOnly}
	fieldSize, err = readOnlyField.SizeInBytes(version, o.readOnly)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"readOnly\" field")
	}
	size += fieldSize

	configSourceField := fields.Int8{Context: creatableTopicConfigsConfigSource}
	fieldSize, err = configSourceField.SizeInBytes(version, o.configSource)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"configSource\" field")
	}
	size += fieldSize

	isSensitiveField := fields.Bool{Context: creatableTopicConfigsIsSensitive}
	fieldSize, err = isSensitiveField.SizeInBytes(version, o.isSensitive)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"isSensitive\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < CreatableTopicConfigsLowestSupportedFlexVersion() || version > CreatableTopicConfigsHighestSupportedFlexVersion() {
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
func (o *CreatableTopicConfigs) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.name.Release()
	o.value.Release()
}

func (o *CreatableTopicConfigs) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *CreatableTopicConfigs) validateNonIgnorableFields(version int16) error {
	if !creatableTopicConfigsName.IsSupportedVersion(version) {
		if o.name.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"name\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !creatableTopicConfigsValue.IsSupportedVersion(version) {
		if o.value.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"value\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !creatableTopicConfigsReadOnly.IsSupportedVersion(version) {
		if o.readOnly {
			return errors.New(strings.Join([]string{"attempted to write non-default \"readOnly\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !creatableTopicConfigsIsSensitive.IsSupportedVersion(version) {
		if o.isSensitive {
			return errors.New(strings.Join([]string{"attempted to write non-default \"isSensitive\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func CreatableTopicConfigsLowestSupportedVersion() int16 {
	return 5
}

func CreatableTopicConfigsHighestSupportedVersion() int16 {
	return 32767
}

func CreatableTopicConfigsLowestSupportedFlexVersion() int16 {
	return 5
}

func CreatableTopicConfigsHighestSupportedFlexVersion() int16 {
	return 32767
}

func CreatableTopicConfigsDefault() CreatableTopicConfigs {
	var d CreatableTopicConfigs
	d.SetDefault()

	return d
}
