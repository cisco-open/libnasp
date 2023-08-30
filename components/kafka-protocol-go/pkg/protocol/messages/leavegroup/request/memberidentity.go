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

var memberIdentityMemberId = fields.Context{
	SpecName:                    "MemberId",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  4,
	HighestSupportedFlexVersion: 32767,
}
var memberIdentityGroupInstanceId = fields.Context{
	SpecName:           "GroupInstanceId",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          3,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      4,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  3,
	HighestSupportedNullableVersion: 32767,
}
var memberIdentityReason = fields.Context{
	SpecName:           "Reason",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          5,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      4,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  5,
	HighestSupportedNullableVersion: 32767,
}

type MemberIdentity struct {
	unknownTaggedFields []fields.RawTaggedField
	memberId            fields.NullableString
	groupInstanceId     fields.NullableString
	reason              fields.NullableString
	isNil               bool
}

func (o *MemberIdentity) MemberId() fields.NullableString {
	return o.memberId
}

func (o *MemberIdentity) SetMemberId(val fields.NullableString) {
	o.isNil = false
	o.memberId = val
}

func (o *MemberIdentity) GroupInstanceId() fields.NullableString {
	return o.groupInstanceId
}

func (o *MemberIdentity) SetGroupInstanceId(val fields.NullableString) {
	o.isNil = false
	o.groupInstanceId = val
}

func (o *MemberIdentity) Reason() fields.NullableString {
	return o.reason
}

func (o *MemberIdentity) SetReason(val fields.NullableString) {
	o.isNil = false
	o.reason = val
}

func (o *MemberIdentity) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *MemberIdentity) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *MemberIdentity) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	memberIdField := fields.String{Context: memberIdentityMemberId}
	if err := memberIdField.Read(buf, version, &o.memberId); err != nil {
		return errors.WrapIf(err, "couldn't set \"memberId\" field")
	}

	groupInstanceIdField := fields.String{Context: memberIdentityGroupInstanceId}
	if err := groupInstanceIdField.Read(buf, version, &o.groupInstanceId); err != nil {
		return errors.WrapIf(err, "couldn't set \"groupInstanceId\" field")
	}

	reasonField := fields.String{Context: memberIdentityReason}
	if err := reasonField.Read(buf, version, &o.reason); err != nil {
		return errors.WrapIf(err, "couldn't set \"reason\" field")
	}

	// process tagged fields

	if version < MemberIdentityLowestSupportedFlexVersion() || version > MemberIdentityHighestSupportedFlexVersion() {
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

func (o *MemberIdentity) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	memberIdField := fields.String{Context: memberIdentityMemberId}
	if err := memberIdField.Write(buf, version, o.memberId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"memberId\" field")
	}
	groupInstanceIdField := fields.String{Context: memberIdentityGroupInstanceId}
	if err := groupInstanceIdField.Write(buf, version, o.groupInstanceId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"groupInstanceId\" field")
	}
	reasonField := fields.String{Context: memberIdentityReason}
	if err := reasonField.Write(buf, version, o.reason); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"reason\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < MemberIdentityLowestSupportedFlexVersion() || version > MemberIdentityHighestSupportedFlexVersion() {
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

func (o *MemberIdentity) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *MemberIdentity) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 4)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.memberId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"memberId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.groupInstanceId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"groupInstanceId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.reason); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"reason\""), b}, []byte(": ")))
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

func (o *MemberIdentity) IsNil() bool {
	return o.isNil
}

func (o *MemberIdentity) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *MemberIdentity) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.memberId.SetValue("")
	o.groupInstanceId.SetValue("null")
	o.reason.SetValue("null")

	o.isNil = false
}

func (o *MemberIdentity) Equal(that *MemberIdentity) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.memberId.Equal(&that.memberId) {
		return false
	}
	if !o.groupInstanceId.Equal(&that.groupInstanceId) {
		return false
	}
	if !o.reason.Equal(&that.reason) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *MemberIdentity) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	memberIdField := fields.String{Context: memberIdentityMemberId}
	fieldSize, err = memberIdField.SizeInBytes(version, o.memberId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"memberId\" field")
	}
	size += fieldSize

	groupInstanceIdField := fields.String{Context: memberIdentityGroupInstanceId}
	fieldSize, err = groupInstanceIdField.SizeInBytes(version, o.groupInstanceId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"groupInstanceId\" field")
	}
	size += fieldSize

	reasonField := fields.String{Context: memberIdentityReason}
	fieldSize, err = reasonField.SizeInBytes(version, o.reason)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"reason\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < MemberIdentityLowestSupportedFlexVersion() || version > MemberIdentityHighestSupportedFlexVersion() {
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
func (o *MemberIdentity) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.memberId.Release()
	o.groupInstanceId.Release()
	o.reason.Release()
}

func (o *MemberIdentity) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *MemberIdentity) validateNonIgnorableFields(version int16) error {
	if !memberIdentityMemberId.IsSupportedVersion(version) {
		if o.memberId.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"memberId\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !memberIdentityGroupInstanceId.IsSupportedVersion(version) {
		if !o.groupInstanceId.IsNil() {
			return errors.New(strings.Join([]string{"attempted to write non-default \"groupInstanceId\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func MemberIdentityLowestSupportedVersion() int16 {
	return 3
}

func MemberIdentityHighestSupportedVersion() int16 {
	return 32767
}

func MemberIdentityLowestSupportedFlexVersion() int16 {
	return 4
}

func MemberIdentityHighestSupportedFlexVersion() int16 {
	return 32767
}

func MemberIdentityDefault() MemberIdentity {
	var d MemberIdentity
	d.SetDefault()

	return d
}
