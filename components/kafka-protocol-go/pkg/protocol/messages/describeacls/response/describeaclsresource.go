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
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/describeacls/response/describeaclsresource"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var describeAclsResourceResourceType = fields.Context{
	SpecName:                    "ResourceType",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describeAclsResourceResourceName = fields.Context{
	SpecName:                    "ResourceName",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describeAclsResourcePatternType = fields.Context{
	SpecName:                    "PatternType",
	CustomDefaultValue:          int8(3),
	LowestSupportedVersion:      1,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describeAclsResourceAcls = fields.Context{
	SpecName:                    "Acls",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type DescribeAclsResource struct {
	acls                []describeaclsresource.AclDescription
	unknownTaggedFields []fields.RawTaggedField
	resourceName        fields.NullableString
	resourceType        int8
	patternType         int8
	isNil               bool
}

func (o *DescribeAclsResource) ResourceType() int8 {
	return o.resourceType
}

func (o *DescribeAclsResource) SetResourceType(val int8) {
	o.isNil = false
	o.resourceType = val
}

func (o *DescribeAclsResource) ResourceName() fields.NullableString {
	return o.resourceName
}

func (o *DescribeAclsResource) SetResourceName(val fields.NullableString) {
	o.isNil = false
	o.resourceName = val
}

func (o *DescribeAclsResource) PatternType() int8 {
	return o.patternType
}

func (o *DescribeAclsResource) SetPatternType(val int8) {
	o.isNil = false
	o.patternType = val
}

func (o *DescribeAclsResource) Acls() []describeaclsresource.AclDescription {
	return o.acls
}

func (o *DescribeAclsResource) SetAcls(val []describeaclsresource.AclDescription) {
	o.isNil = false
	o.acls = val
}

func (o *DescribeAclsResource) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *DescribeAclsResource) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *DescribeAclsResource) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	resourceTypeField := fields.Int8{Context: describeAclsResourceResourceType}
	if err := resourceTypeField.Read(buf, version, &o.resourceType); err != nil {
		return errors.WrapIf(err, "couldn't set \"resourceType\" field")
	}

	resourceNameField := fields.String{Context: describeAclsResourceResourceName}
	if err := resourceNameField.Read(buf, version, &o.resourceName); err != nil {
		return errors.WrapIf(err, "couldn't set \"resourceName\" field")
	}

	patternTypeField := fields.Int8{Context: describeAclsResourcePatternType}
	if err := patternTypeField.Read(buf, version, &o.patternType); err != nil {
		return errors.WrapIf(err, "couldn't set \"patternType\" field")
	}

	aclsField := fields.ArrayOfStruct[describeaclsresource.AclDescription, *describeaclsresource.AclDescription]{Context: describeAclsResourceAcls}
	acls, err := aclsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"acls\" field")
	}
	o.acls = acls

	// process tagged fields

	if version < DescribeAclsResourceLowestSupportedFlexVersion() || version > DescribeAclsResourceHighestSupportedFlexVersion() {
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

func (o *DescribeAclsResource) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	resourceTypeField := fields.Int8{Context: describeAclsResourceResourceType}
	if err := resourceTypeField.Write(buf, version, o.resourceType); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"resourceType\" field")
	}
	resourceNameField := fields.String{Context: describeAclsResourceResourceName}
	if err := resourceNameField.Write(buf, version, o.resourceName); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"resourceName\" field")
	}
	patternTypeField := fields.Int8{Context: describeAclsResourcePatternType}
	if err := patternTypeField.Write(buf, version, o.patternType); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"patternType\" field")
	}

	aclsField := fields.ArrayOfStruct[describeaclsresource.AclDescription, *describeaclsresource.AclDescription]{Context: describeAclsResourceAcls}
	if err := aclsField.Write(buf, version, o.acls); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"acls\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < DescribeAclsResourceLowestSupportedFlexVersion() || version > DescribeAclsResourceHighestSupportedFlexVersion() {
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

func (o *DescribeAclsResource) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *DescribeAclsResource) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.resourceType); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"resourceType\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.resourceName); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"resourceName\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.patternType); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"patternType\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("acls", o.acls); err != nil {
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

func (o *DescribeAclsResource) IsNil() bool {
	return o.isNil
}

func (o *DescribeAclsResource) Clear() {
	o.Release()
	o.isNil = true

	o.acls = nil
	o.unknownTaggedFields = nil
}

func (o *DescribeAclsResource) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.resourceType = 0
	o.resourceName.SetValue("")
	o.patternType = 3
	for i := range o.acls {
		o.acls[i].Release()
	}
	o.acls = nil

	o.isNil = false
}

func (o *DescribeAclsResource) Equal(that *DescribeAclsResource) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.resourceType != that.resourceType {
		return false
	}
	if !o.resourceName.Equal(&that.resourceName) {
		return false
	}
	if o.patternType != that.patternType {
		return false
	}
	if len(o.acls) != len(that.acls) {
		return false
	}
	for i := range o.acls {
		if !o.acls[i].Equal(&that.acls[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *DescribeAclsResource) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	resourceTypeField := fields.Int8{Context: describeAclsResourceResourceType}
	fieldSize, err = resourceTypeField.SizeInBytes(version, o.resourceType)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"resourceType\" field")
	}
	size += fieldSize

	resourceNameField := fields.String{Context: describeAclsResourceResourceName}
	fieldSize, err = resourceNameField.SizeInBytes(version, o.resourceName)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"resourceName\" field")
	}
	size += fieldSize

	patternTypeField := fields.Int8{Context: describeAclsResourcePatternType}
	fieldSize, err = patternTypeField.SizeInBytes(version, o.patternType)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"patternType\" field")
	}
	size += fieldSize

	aclsField := fields.ArrayOfStruct[describeaclsresource.AclDescription, *describeaclsresource.AclDescription]{Context: describeAclsResourceAcls}
	fieldSize, err = aclsField.SizeInBytes(version, o.acls)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"acls\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < DescribeAclsResourceLowestSupportedFlexVersion() || version > DescribeAclsResourceHighestSupportedFlexVersion() {
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
func (o *DescribeAclsResource) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.resourceName.Release()
	for i := range o.acls {
		o.acls[i].Release()
	}
	o.acls = nil
}

func (o *DescribeAclsResource) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *DescribeAclsResource) validateNonIgnorableFields(version int16) error {
	if !describeAclsResourcePatternType.IsSupportedVersion(version) {
		if o.patternType != 3 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"patternType\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func DescribeAclsResourceLowestSupportedVersion() int16 {
	return 0
}

func DescribeAclsResourceHighestSupportedVersion() int16 {
	return 32767
}

func DescribeAclsResourceLowestSupportedFlexVersion() int16 {
	return 2
}

func DescribeAclsResourceHighestSupportedFlexVersion() int16 {
	return 32767
}

func DescribeAclsResourceDefault() DescribeAclsResource {
	var d DescribeAclsResource
	d.SetDefault()

	return d
}