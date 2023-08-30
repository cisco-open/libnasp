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
package describedgroup

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var describedGroupMemberMemberId = fields.Context{
	SpecName:                    "MemberId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}
var describedGroupMemberGroupInstanceId = fields.Context{
	SpecName:           "GroupInstanceId",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          4,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      5,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  4,
	HighestSupportedNullableVersion: 32767,
}
var describedGroupMemberClientId = fields.Context{
	SpecName:                    "ClientId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}
var describedGroupMemberClientHost = fields.Context{
	SpecName:                    "ClientHost",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}
var describedGroupMemberMemberMetadata = fields.Context{
	SpecName:                    "MemberMetadata",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}
var describedGroupMemberMemberAssignment = fields.Context{
	SpecName:                    "MemberAssignment",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  5,
	HighestSupportedFlexVersion: 32767,
}

type DescribedGroupMember struct {
	memberMetadata      []byte
	memberAssignment    []byte
	unknownTaggedFields []fields.RawTaggedField
	memberId            fields.NullableString
	groupInstanceId     fields.NullableString
	clientId            fields.NullableString
	clientHost          fields.NullableString
	isNil               bool
}

func (o *DescribedGroupMember) MemberId() fields.NullableString {
	return o.memberId
}

func (o *DescribedGroupMember) SetMemberId(val fields.NullableString) {
	o.isNil = false
	o.memberId = val
}

func (o *DescribedGroupMember) GroupInstanceId() fields.NullableString {
	return o.groupInstanceId
}

func (o *DescribedGroupMember) SetGroupInstanceId(val fields.NullableString) {
	o.isNil = false
	o.groupInstanceId = val
}

func (o *DescribedGroupMember) ClientId() fields.NullableString {
	return o.clientId
}

func (o *DescribedGroupMember) SetClientId(val fields.NullableString) {
	o.isNil = false
	o.clientId = val
}

func (o *DescribedGroupMember) ClientHost() fields.NullableString {
	return o.clientHost
}

func (o *DescribedGroupMember) SetClientHost(val fields.NullableString) {
	o.isNil = false
	o.clientHost = val
}

func (o *DescribedGroupMember) MemberMetadata() []byte {
	return o.memberMetadata
}

func (o *DescribedGroupMember) SetMemberMetadata(val []byte) {
	o.isNil = false
	o.memberMetadata = val
}

func (o *DescribedGroupMember) MemberAssignment() []byte {
	return o.memberAssignment
}

func (o *DescribedGroupMember) SetMemberAssignment(val []byte) {
	o.isNil = false
	o.memberAssignment = val
}

func (o *DescribedGroupMember) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *DescribedGroupMember) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *DescribedGroupMember) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	memberIdField := fields.String{Context: describedGroupMemberMemberId}
	if err := memberIdField.Read(buf, version, &o.memberId); err != nil {
		return errors.WrapIf(err, "couldn't set \"memberId\" field")
	}

	groupInstanceIdField := fields.String{Context: describedGroupMemberGroupInstanceId}
	if err := groupInstanceIdField.Read(buf, version, &o.groupInstanceId); err != nil {
		return errors.WrapIf(err, "couldn't set \"groupInstanceId\" field")
	}

	clientIdField := fields.String{Context: describedGroupMemberClientId}
	if err := clientIdField.Read(buf, version, &o.clientId); err != nil {
		return errors.WrapIf(err, "couldn't set \"clientId\" field")
	}

	clientHostField := fields.String{Context: describedGroupMemberClientHost}
	if err := clientHostField.Read(buf, version, &o.clientHost); err != nil {
		return errors.WrapIf(err, "couldn't set \"clientHost\" field")
	}

	memberMetadataField := fields.Bytes{Context: describedGroupMemberMemberMetadata}
	if err := memberMetadataField.Read(buf, version, &o.memberMetadata); err != nil {
		return errors.WrapIf(err, "couldn't set \"memberMetadata\" field")
	}

	memberAssignmentField := fields.Bytes{Context: describedGroupMemberMemberAssignment}
	if err := memberAssignmentField.Read(buf, version, &o.memberAssignment); err != nil {
		return errors.WrapIf(err, "couldn't set \"memberAssignment\" field")
	}

	// process tagged fields

	if version < DescribedGroupMemberLowestSupportedFlexVersion() || version > DescribedGroupMemberHighestSupportedFlexVersion() {
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

func (o *DescribedGroupMember) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	memberIdField := fields.String{Context: describedGroupMemberMemberId}
	if err := memberIdField.Write(buf, version, o.memberId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"memberId\" field")
	}
	groupInstanceIdField := fields.String{Context: describedGroupMemberGroupInstanceId}
	if err := groupInstanceIdField.Write(buf, version, o.groupInstanceId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"groupInstanceId\" field")
	}
	clientIdField := fields.String{Context: describedGroupMemberClientId}
	if err := clientIdField.Write(buf, version, o.clientId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"clientId\" field")
	}
	clientHostField := fields.String{Context: describedGroupMemberClientHost}
	if err := clientHostField.Write(buf, version, o.clientHost); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"clientHost\" field")
	}
	memberMetadataField := fields.Bytes{Context: describedGroupMemberMemberMetadata}
	if err := memberMetadataField.Write(buf, version, o.memberMetadata); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"memberMetadata\" field")
	}
	memberAssignmentField := fields.Bytes{Context: describedGroupMemberMemberAssignment}
	if err := memberAssignmentField.Write(buf, version, o.memberAssignment); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"memberAssignment\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < DescribedGroupMemberLowestSupportedFlexVersion() || version > DescribedGroupMemberHighestSupportedFlexVersion() {
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

func (o *DescribedGroupMember) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *DescribedGroupMember) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 7)
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
	if b, err := fields.MarshalPrimitiveTypeJSON(o.clientId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"clientId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.clientHost); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"clientHost\""), b}, []byte(": ")))
	}
	if b, err := fields.BytesMarshalJSON("memberMetadata", o.memberMetadata); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.BytesMarshalJSON("memberAssignment", o.memberAssignment); err != nil {
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

func (o *DescribedGroupMember) IsNil() bool {
	return o.isNil
}

func (o *DescribedGroupMember) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *DescribedGroupMember) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.memberId.SetValue("")
	o.groupInstanceId.SetValue("null")
	o.clientId.SetValue("")
	o.clientHost.SetValue("")
	if o.memberMetadata != nil {
		pools.ReleaseByteSlice(o.memberMetadata)
	}
	o.memberMetadata = nil
	if o.memberAssignment != nil {
		pools.ReleaseByteSlice(o.memberAssignment)
	}
	o.memberAssignment = nil

	o.isNil = false
}

func (o *DescribedGroupMember) Equal(that *DescribedGroupMember) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.memberId.Equal(&that.memberId) {
		return false
	}
	if !o.groupInstanceId.Equal(&that.groupInstanceId) {
		return false
	}
	if !o.clientId.Equal(&that.clientId) {
		return false
	}
	if !o.clientHost.Equal(&that.clientHost) {
		return false
	}
	if !bytes.Equal(o.memberMetadata, that.memberMetadata) {
		return false
	}
	if !bytes.Equal(o.memberAssignment, that.memberAssignment) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *DescribedGroupMember) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	memberIdField := fields.String{Context: describedGroupMemberMemberId}
	fieldSize, err = memberIdField.SizeInBytes(version, o.memberId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"memberId\" field")
	}
	size += fieldSize

	groupInstanceIdField := fields.String{Context: describedGroupMemberGroupInstanceId}
	fieldSize, err = groupInstanceIdField.SizeInBytes(version, o.groupInstanceId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"groupInstanceId\" field")
	}
	size += fieldSize

	clientIdField := fields.String{Context: describedGroupMemberClientId}
	fieldSize, err = clientIdField.SizeInBytes(version, o.clientId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"clientId\" field")
	}
	size += fieldSize

	clientHostField := fields.String{Context: describedGroupMemberClientHost}
	fieldSize, err = clientHostField.SizeInBytes(version, o.clientHost)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"clientHost\" field")
	}
	size += fieldSize

	memberMetadataField := fields.Bytes{Context: describedGroupMemberMemberMetadata}
	fieldSize, err = memberMetadataField.SizeInBytes(version, o.memberMetadata)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"memberMetadata\" field")
	}
	size += fieldSize

	memberAssignmentField := fields.Bytes{Context: describedGroupMemberMemberAssignment}
	fieldSize, err = memberAssignmentField.SizeInBytes(version, o.memberAssignment)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"memberAssignment\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < DescribedGroupMemberLowestSupportedFlexVersion() || version > DescribedGroupMemberHighestSupportedFlexVersion() {
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
func (o *DescribedGroupMember) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.memberId.Release()
	o.groupInstanceId.Release()
	o.clientId.Release()
	o.clientHost.Release()
	if o.memberMetadata != nil {
		pools.ReleaseByteSlice(o.memberMetadata)
	}
	if o.memberAssignment != nil {
		pools.ReleaseByteSlice(o.memberAssignment)
	}
}

func (o *DescribedGroupMember) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *DescribedGroupMember) validateNonIgnorableFields(version int16) error {
	return nil
}

func DescribedGroupMemberLowestSupportedVersion() int16 {
	return 0
}

func DescribedGroupMemberHighestSupportedVersion() int16 {
	return 32767
}

func DescribedGroupMemberLowestSupportedFlexVersion() int16 {
	return 5
}

func DescribedGroupMemberHighestSupportedFlexVersion() int16 {
	return 32767
}

func DescribedGroupMemberDefault() DescribedGroupMember {
	var d DescribedGroupMember
	d.SetDefault()

	return d
}
