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
package createdelegationtoken

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/createdelegationtoken/request"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var requestOwnerPrincipalType = fields.Context{
	SpecName:                        "OwnerPrincipalType",
	LowestSupportedVersion:          3,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  3,
	HighestSupportedNullableVersion: 32767,
}
var requestOwnerPrincipalName = fields.Context{
	SpecName:                        "OwnerPrincipalName",
	LowestSupportedVersion:          3,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      2,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  3,
	HighestSupportedNullableVersion: 32767,
}
var requestRenewers = fields.Context{
	SpecName:                    "Renewers",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var requestMaxLifetimeMs = fields.Context{
	SpecName:                    "MaxLifetimeMs",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type Request struct {
	renewers            []request.CreatableRenewers
	unknownTaggedFields []fields.RawTaggedField
	ownerPrincipalType  fields.NullableString
	ownerPrincipalName  fields.NullableString
	maxLifetimeMs       int64
	isNil               bool
}

func (o *Request) OwnerPrincipalType() fields.NullableString {
	return o.ownerPrincipalType
}

func (o *Request) SetOwnerPrincipalType(val fields.NullableString) {
	o.isNil = false
	o.ownerPrincipalType = val
}

func (o *Request) OwnerPrincipalName() fields.NullableString {
	return o.ownerPrincipalName
}

func (o *Request) SetOwnerPrincipalName(val fields.NullableString) {
	o.isNil = false
	o.ownerPrincipalName = val
}

func (o *Request) Renewers() []request.CreatableRenewers {
	return o.renewers
}

func (o *Request) SetRenewers(val []request.CreatableRenewers) {
	o.isNil = false
	o.renewers = val
}

func (o *Request) MaxLifetimeMs() int64 {
	return o.maxLifetimeMs
}

func (o *Request) SetMaxLifetimeMs(val int64) {
	o.isNil = false
	o.maxLifetimeMs = val
}

func (o *Request) ApiKey() int16 {
	return 38
}

func (o *Request) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Request) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Request) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	ownerPrincipalTypeField := fields.String{Context: requestOwnerPrincipalType}
	if err := ownerPrincipalTypeField.Read(buf, version, &o.ownerPrincipalType); err != nil {
		return errors.WrapIf(err, "couldn't set \"ownerPrincipalType\" field")
	}

	ownerPrincipalNameField := fields.String{Context: requestOwnerPrincipalName}
	if err := ownerPrincipalNameField.Read(buf, version, &o.ownerPrincipalName); err != nil {
		return errors.WrapIf(err, "couldn't set \"ownerPrincipalName\" field")
	}

	renewersField := fields.ArrayOfStruct[request.CreatableRenewers, *request.CreatableRenewers]{Context: requestRenewers}
	renewers, err := renewersField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"renewers\" field")
	}
	o.renewers = renewers

	maxLifetimeMsField := fields.Int64{Context: requestMaxLifetimeMs}
	if err := maxLifetimeMsField.Read(buf, version, &o.maxLifetimeMs); err != nil {
		return errors.WrapIf(err, "couldn't set \"maxLifetimeMs\" field")
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

	ownerPrincipalTypeField := fields.String{Context: requestOwnerPrincipalType}
	if err := ownerPrincipalTypeField.Write(buf, version, o.ownerPrincipalType); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"ownerPrincipalType\" field")
	}
	ownerPrincipalNameField := fields.String{Context: requestOwnerPrincipalName}
	if err := ownerPrincipalNameField.Write(buf, version, o.ownerPrincipalName); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"ownerPrincipalName\" field")
	}

	renewersField := fields.ArrayOfStruct[request.CreatableRenewers, *request.CreatableRenewers]{Context: requestRenewers}
	if err := renewersField.Write(buf, version, o.renewers); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"renewers\" field")
	}

	maxLifetimeMsField := fields.Int64{Context: requestMaxLifetimeMs}
	if err := maxLifetimeMsField.Write(buf, version, o.maxLifetimeMs); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"maxLifetimeMs\" field")
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
	if b, err := fields.MarshalPrimitiveTypeJSON(o.ownerPrincipalType); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"ownerPrincipalType\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.ownerPrincipalName); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"ownerPrincipalName\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("renewers", o.renewers); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.maxLifetimeMs); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"maxLifetimeMs\""), b}, []byte(": ")))
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

	o.renewers = nil
	o.unknownTaggedFields = nil
}

func (o *Request) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.ownerPrincipalType.SetValue("")
	o.ownerPrincipalName.SetValue("")
	for i := range o.renewers {
		o.renewers[i].Release()
	}
	o.renewers = nil
	o.maxLifetimeMs = 0

	o.isNil = false
}

func (o *Request) Equal(that *Request) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.ownerPrincipalType.Equal(&that.ownerPrincipalType) {
		return false
	}
	if !o.ownerPrincipalName.Equal(&that.ownerPrincipalName) {
		return false
	}
	if len(o.renewers) != len(that.renewers) {
		return false
	}
	for i := range o.renewers {
		if !o.renewers[i].Equal(&that.renewers[i]) {
			return false
		}
	}
	if o.maxLifetimeMs != that.maxLifetimeMs {
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

	ownerPrincipalTypeField := fields.String{Context: requestOwnerPrincipalType}
	fieldSize, err = ownerPrincipalTypeField.SizeInBytes(version, o.ownerPrincipalType)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"ownerPrincipalType\" field")
	}
	size += fieldSize

	ownerPrincipalNameField := fields.String{Context: requestOwnerPrincipalName}
	fieldSize, err = ownerPrincipalNameField.SizeInBytes(version, o.ownerPrincipalName)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"ownerPrincipalName\" field")
	}
	size += fieldSize

	renewersField := fields.ArrayOfStruct[request.CreatableRenewers, *request.CreatableRenewers]{Context: requestRenewers}
	fieldSize, err = renewersField.SizeInBytes(version, o.renewers)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"renewers\" field")
	}
	size += fieldSize

	maxLifetimeMsField := fields.Int64{Context: requestMaxLifetimeMs}
	fieldSize, err = maxLifetimeMsField.SizeInBytes(version, o.maxLifetimeMs)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"maxLifetimeMs\" field")
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

	o.ownerPrincipalType.Release()
	o.ownerPrincipalName.Release()
	for i := range o.renewers {
		o.renewers[i].Release()
	}
	o.renewers = nil
}

func (o *Request) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Request) validateNonIgnorableFields(version int16) error {
	if !requestOwnerPrincipalType.IsSupportedVersion(version) {
		if o.ownerPrincipalType.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"ownerPrincipalType\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !requestOwnerPrincipalName.IsSupportedVersion(version) {
		if o.ownerPrincipalName.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"ownerPrincipalName\" at version", strconv.Itoa(int(version))}, " "))
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