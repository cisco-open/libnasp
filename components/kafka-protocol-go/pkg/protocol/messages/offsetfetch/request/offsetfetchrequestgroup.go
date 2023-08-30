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
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/offsetfetch/request/offsetfetchrequestgroup"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var offsetFetchRequestGroupGroupId = fields.Context{
	SpecName:                    "groupId",
	LowestSupportedVersion:      8,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var offsetFetchRequestGroupTopics = fields.Context{
	SpecName:                        "Topics",
	LowestSupportedVersion:          8,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      6,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  8,
	HighestSupportedNullableVersion: 32767,
}

type OffsetFetchRequestGroup struct {
	topics              []offsetfetchrequestgroup.OffsetFetchRequestTopics
	unknownTaggedFields []fields.RawTaggedField
	groupId             fields.NullableString
	isNil               bool
}

func (o *OffsetFetchRequestGroup) GroupId() fields.NullableString {
	return o.groupId
}

func (o *OffsetFetchRequestGroup) SetGroupId(val fields.NullableString) {
	o.isNil = false
	o.groupId = val
}

func (o *OffsetFetchRequestGroup) Topics() []offsetfetchrequestgroup.OffsetFetchRequestTopics {
	return o.topics
}

func (o *OffsetFetchRequestGroup) SetTopics(val []offsetfetchrequestgroup.OffsetFetchRequestTopics) {
	o.isNil = false
	o.topics = val
}

func (o *OffsetFetchRequestGroup) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *OffsetFetchRequestGroup) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *OffsetFetchRequestGroup) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	groupIdField := fields.String{Context: offsetFetchRequestGroupGroupId}
	if err := groupIdField.Read(buf, version, &o.groupId); err != nil {
		return errors.WrapIf(err, "couldn't set \"groupId\" field")
	}

	topicsField := fields.ArrayOfStruct[offsetfetchrequestgroup.OffsetFetchRequestTopics, *offsetfetchrequestgroup.OffsetFetchRequestTopics]{Context: offsetFetchRequestGroupTopics}
	topics, err := topicsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"topics\" field")
	}
	o.topics = topics

	// process tagged fields

	if version < OffsetFetchRequestGroupLowestSupportedFlexVersion() || version > OffsetFetchRequestGroupHighestSupportedFlexVersion() {
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

func (o *OffsetFetchRequestGroup) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	groupIdField := fields.String{Context: offsetFetchRequestGroupGroupId}
	if err := groupIdField.Write(buf, version, o.groupId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"groupId\" field")
	}

	topicsField := fields.ArrayOfStruct[offsetfetchrequestgroup.OffsetFetchRequestTopics, *offsetfetchrequestgroup.OffsetFetchRequestTopics]{Context: offsetFetchRequestGroupTopics}
	if err := topicsField.Write(buf, version, o.topics); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topics\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < OffsetFetchRequestGroupLowestSupportedFlexVersion() || version > OffsetFetchRequestGroupHighestSupportedFlexVersion() {
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

func (o *OffsetFetchRequestGroup) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *OffsetFetchRequestGroup) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 3)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.groupId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"groupId\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("topics", o.topics); err != nil {
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

func (o *OffsetFetchRequestGroup) IsNil() bool {
	return o.isNil
}

func (o *OffsetFetchRequestGroup) Clear() {
	o.Release()
	o.isNil = true

	o.topics = nil
	o.unknownTaggedFields = nil
}

func (o *OffsetFetchRequestGroup) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.groupId.SetValue("")
	for i := range o.topics {
		o.topics[i].Release()
	}
	o.topics = nil

	o.isNil = false
}

func (o *OffsetFetchRequestGroup) Equal(that *OffsetFetchRequestGroup) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.groupId.Equal(&that.groupId) {
		return false
	}
	if len(o.topics) != len(that.topics) {
		return false
	}
	for i := range o.topics {
		if !o.topics[i].Equal(&that.topics[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *OffsetFetchRequestGroup) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	groupIdField := fields.String{Context: offsetFetchRequestGroupGroupId}
	fieldSize, err = groupIdField.SizeInBytes(version, o.groupId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"groupId\" field")
	}
	size += fieldSize

	topicsField := fields.ArrayOfStruct[offsetfetchrequestgroup.OffsetFetchRequestTopics, *offsetfetchrequestgroup.OffsetFetchRequestTopics]{Context: offsetFetchRequestGroupTopics}
	fieldSize, err = topicsField.SizeInBytes(version, o.topics)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topics\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < OffsetFetchRequestGroupLowestSupportedFlexVersion() || version > OffsetFetchRequestGroupHighestSupportedFlexVersion() {
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
func (o *OffsetFetchRequestGroup) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.groupId.Release()
	for i := range o.topics {
		o.topics[i].Release()
	}
	o.topics = nil
}

func (o *OffsetFetchRequestGroup) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *OffsetFetchRequestGroup) validateNonIgnorableFields(version int16) error {
	if !offsetFetchRequestGroupGroupId.IsSupportedVersion(version) {
		if o.groupId.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"groupId\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !offsetFetchRequestGroupTopics.IsSupportedVersion(version) {
		if len(o.topics) > 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"topics\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func OffsetFetchRequestGroupLowestSupportedVersion() int16 {
	return 8
}

func OffsetFetchRequestGroupHighestSupportedVersion() int16 {
	return 32767
}

func OffsetFetchRequestGroupLowestSupportedFlexVersion() int16 {
	return 6
}

func OffsetFetchRequestGroupHighestSupportedFlexVersion() int16 {
	return 32767
}

func OffsetFetchRequestGroupDefault() OffsetFetchRequestGroup {
	var d OffsetFetchRequestGroup
	d.SetDefault()

	return d
}
