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
package describeacls

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var requestResourceTypeFilter = fields.Context{
	SpecName:                    "ResourceTypeFilter",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var requestResourceNameFilter = fields.Context{
	SpecName:                        "ResourceNameFilter",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}
var requestPatternTypeFilter = fields.Context{
	SpecName:                    "PatternTypeFilter",
	CustomDefaultValue:          int8(3),
	LowestSupportedVersion:      1,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var requestPrincipalFilter = fields.Context{
	SpecName:                        "PrincipalFilter",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}
var requestHostFilter = fields.Context{
	SpecName:                        "HostFilter",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}
var requestOperation = fields.Context{
	SpecName:                    "Operation",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var requestPermissionType = fields.Context{
	SpecName:                    "PermissionType",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type Request struct {
	unknownTaggedFields []fields.RawTaggedField
	resourceNameFilter  fields.NullableString
	principalFilter     fields.NullableString
	hostFilter          fields.NullableString
	resourceTypeFilter  int8
	patternTypeFilter   int8
	operation           int8
	permissionType      int8
	isNil               bool
}

func (o *Request) ResourceTypeFilter() int8 {
	return o.resourceTypeFilter
}

func (o *Request) SetResourceTypeFilter(val int8) {
	o.isNil = false
	o.resourceTypeFilter = val
}

func (o *Request) ResourceNameFilter() fields.NullableString {
	return o.resourceNameFilter
}

func (o *Request) SetResourceNameFilter(val fields.NullableString) {
	o.isNil = false
	o.resourceNameFilter = val
}

func (o *Request) PatternTypeFilter() int8 {
	return o.patternTypeFilter
}

func (o *Request) SetPatternTypeFilter(val int8) {
	o.isNil = false
	o.patternTypeFilter = val
}

func (o *Request) PrincipalFilter() fields.NullableString {
	return o.principalFilter
}

func (o *Request) SetPrincipalFilter(val fields.NullableString) {
	o.isNil = false
	o.principalFilter = val
}

func (o *Request) HostFilter() fields.NullableString {
	return o.hostFilter
}

func (o *Request) SetHostFilter(val fields.NullableString) {
	o.isNil = false
	o.hostFilter = val
}

func (o *Request) Operation() int8 {
	return o.operation
}

func (o *Request) SetOperation(val int8) {
	o.isNil = false
	o.operation = val
}

func (o *Request) PermissionType() int8 {
	return o.permissionType
}

func (o *Request) SetPermissionType(val int8) {
	o.isNil = false
	o.permissionType = val
}

func (o *Request) ApiKey() int16 {
	return 29
}

func (o *Request) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Request) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Request) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	resourceTypeFilterField := fields.Int8{Context: requestResourceTypeFilter}
	if err := resourceTypeFilterField.Read(buf, version, &o.resourceTypeFilter); err != nil {
		return errors.WrapIf(err, "couldn't set \"resourceTypeFilter\" field")
	}

	resourceNameFilterField := fields.String{Context: requestResourceNameFilter}
	if err := resourceNameFilterField.Read(buf, version, &o.resourceNameFilter); err != nil {
		return errors.WrapIf(err, "couldn't set \"resourceNameFilter\" field")
	}

	patternTypeFilterField := fields.Int8{Context: requestPatternTypeFilter}
	if err := patternTypeFilterField.Read(buf, version, &o.patternTypeFilter); err != nil {
		return errors.WrapIf(err, "couldn't set \"patternTypeFilter\" field")
	}

	principalFilterField := fields.String{Context: requestPrincipalFilter}
	if err := principalFilterField.Read(buf, version, &o.principalFilter); err != nil {
		return errors.WrapIf(err, "couldn't set \"principalFilter\" field")
	}

	hostFilterField := fields.String{Context: requestHostFilter}
	if err := hostFilterField.Read(buf, version, &o.hostFilter); err != nil {
		return errors.WrapIf(err, "couldn't set \"hostFilter\" field")
	}

	operationField := fields.Int8{Context: requestOperation}
	if err := operationField.Read(buf, version, &o.operation); err != nil {
		return errors.WrapIf(err, "couldn't set \"operation\" field")
	}

	permissionTypeField := fields.Int8{Context: requestPermissionType}
	if err := permissionTypeField.Read(buf, version, &o.permissionType); err != nil {
		return errors.WrapIf(err, "couldn't set \"permissionType\" field")
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

	resourceTypeFilterField := fields.Int8{Context: requestResourceTypeFilter}
	if err := resourceTypeFilterField.Write(buf, version, o.resourceTypeFilter); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"resourceTypeFilter\" field")
	}
	resourceNameFilterField := fields.String{Context: requestResourceNameFilter}
	if err := resourceNameFilterField.Write(buf, version, o.resourceNameFilter); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"resourceNameFilter\" field")
	}
	patternTypeFilterField := fields.Int8{Context: requestPatternTypeFilter}
	if err := patternTypeFilterField.Write(buf, version, o.patternTypeFilter); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"patternTypeFilter\" field")
	}
	principalFilterField := fields.String{Context: requestPrincipalFilter}
	if err := principalFilterField.Write(buf, version, o.principalFilter); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"principalFilter\" field")
	}
	hostFilterField := fields.String{Context: requestHostFilter}
	if err := hostFilterField.Write(buf, version, o.hostFilter); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"hostFilter\" field")
	}
	operationField := fields.Int8{Context: requestOperation}
	if err := operationField.Write(buf, version, o.operation); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"operation\" field")
	}
	permissionTypeField := fields.Int8{Context: requestPermissionType}
	if err := permissionTypeField.Write(buf, version, o.permissionType); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"permissionType\" field")
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

	s := make([][]byte, 0, 8)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.resourceTypeFilter); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"resourceTypeFilter\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.resourceNameFilter); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"resourceNameFilter\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.patternTypeFilter); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"patternTypeFilter\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.principalFilter); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"principalFilter\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.hostFilter); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"hostFilter\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.operation); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"operation\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.permissionType); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"permissionType\""), b}, []byte(": ")))
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
	o.resourceTypeFilter = 0
	o.resourceNameFilter.SetValue("")
	o.patternTypeFilter = 3
	o.principalFilter.SetValue("")
	o.hostFilter.SetValue("")
	o.operation = 0
	o.permissionType = 0

	o.isNil = false
}

func (o *Request) Equal(that *Request) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.resourceTypeFilter != that.resourceTypeFilter {
		return false
	}
	if !o.resourceNameFilter.Equal(&that.resourceNameFilter) {
		return false
	}
	if o.patternTypeFilter != that.patternTypeFilter {
		return false
	}
	if !o.principalFilter.Equal(&that.principalFilter) {
		return false
	}
	if !o.hostFilter.Equal(&that.hostFilter) {
		return false
	}
	if o.operation != that.operation {
		return false
	}
	if o.permissionType != that.permissionType {
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

	resourceTypeFilterField := fields.Int8{Context: requestResourceTypeFilter}
	fieldSize, err = resourceTypeFilterField.SizeInBytes(version, o.resourceTypeFilter)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"resourceTypeFilter\" field")
	}
	size += fieldSize

	resourceNameFilterField := fields.String{Context: requestResourceNameFilter}
	fieldSize, err = resourceNameFilterField.SizeInBytes(version, o.resourceNameFilter)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"resourceNameFilter\" field")
	}
	size += fieldSize

	patternTypeFilterField := fields.Int8{Context: requestPatternTypeFilter}
	fieldSize, err = patternTypeFilterField.SizeInBytes(version, o.patternTypeFilter)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"patternTypeFilter\" field")
	}
	size += fieldSize

	principalFilterField := fields.String{Context: requestPrincipalFilter}
	fieldSize, err = principalFilterField.SizeInBytes(version, o.principalFilter)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"principalFilter\" field")
	}
	size += fieldSize

	hostFilterField := fields.String{Context: requestHostFilter}
	fieldSize, err = hostFilterField.SizeInBytes(version, o.hostFilter)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"hostFilter\" field")
	}
	size += fieldSize

	operationField := fields.Int8{Context: requestOperation}
	fieldSize, err = operationField.SizeInBytes(version, o.operation)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"operation\" field")
	}
	size += fieldSize

	permissionTypeField := fields.Int8{Context: requestPermissionType}
	fieldSize, err = permissionTypeField.SizeInBytes(version, o.permissionType)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"permissionType\" field")
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

	o.resourceNameFilter.Release()
	o.principalFilter.Release()
	o.hostFilter.Release()
}

func (o *Request) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Request) validateNonIgnorableFields(version int16) error {
	if !requestPatternTypeFilter.IsSupportedVersion(version) {
		if o.patternTypeFilter != 3 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"patternTypeFilter\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func RequestLowestSupportedVersion() int16 {
	return 0
}

func RequestHighestSupportedVersion() int16 {
	return 3
}

func RequestLowestSupportedFlexVersion() int16 {
	return 2
}

func RequestHighestSupportedFlexVersion() int16 {
	return 32767
}

func RequestDefault() Request {
	var d Request
	d.SetDefault()

	return d
}