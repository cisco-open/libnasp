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
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var metadataResponseBrokerNodeId = fields.Context{
	SpecName:                    "NodeId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var metadataResponseBrokerHost = fields.Context{
	SpecName:                    "Host",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var metadataResponseBrokerPort = fields.Context{
	SpecName:                    "Port",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var metadataResponseBrokerRack = fields.Context{
	SpecName:           "Rack",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          1,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      9,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  1,
	HighestSupportedNullableVersion: 32767,
}

type MetadataResponseBroker struct {
	unknownTaggedFields []fields.RawTaggedField
	host                fields.NullableString
	rack                fields.NullableString
	nodeId              int32
	port                int32
	isNil               bool
}

func (o *MetadataResponseBroker) NodeId() int32 {
	return o.nodeId
}

func (o *MetadataResponseBroker) SetNodeId(val int32) {
	o.isNil = false
	o.nodeId = val
}

func (o *MetadataResponseBroker) Host() fields.NullableString {
	return o.host
}

func (o *MetadataResponseBroker) SetHost(val fields.NullableString) {
	o.isNil = false
	o.host = val
}

func (o *MetadataResponseBroker) Port() int32 {
	return o.port
}

func (o *MetadataResponseBroker) SetPort(val int32) {
	o.isNil = false
	o.port = val
}

func (o *MetadataResponseBroker) Rack() fields.NullableString {
	return o.rack
}

func (o *MetadataResponseBroker) SetRack(val fields.NullableString) {
	o.isNil = false
	o.rack = val
}

func (o *MetadataResponseBroker) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *MetadataResponseBroker) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *MetadataResponseBroker) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	nodeIdField := fields.Int32{Context: metadataResponseBrokerNodeId}
	if err := nodeIdField.Read(buf, version, &o.nodeId); err != nil {
		return errors.WrapIf(err, "couldn't set \"nodeId\" field")
	}

	hostField := fields.String{Context: metadataResponseBrokerHost}
	if err := hostField.Read(buf, version, &o.host); err != nil {
		return errors.WrapIf(err, "couldn't set \"host\" field")
	}

	portField := fields.Int32{Context: metadataResponseBrokerPort}
	if err := portField.Read(buf, version, &o.port); err != nil {
		return errors.WrapIf(err, "couldn't set \"port\" field")
	}

	rackField := fields.String{Context: metadataResponseBrokerRack}
	if err := rackField.Read(buf, version, &o.rack); err != nil {
		return errors.WrapIf(err, "couldn't set \"rack\" field")
	}

	// process tagged fields

	if version < MetadataResponseBrokerLowestSupportedFlexVersion() || version > MetadataResponseBrokerHighestSupportedFlexVersion() {
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

func (o *MetadataResponseBroker) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	nodeIdField := fields.Int32{Context: metadataResponseBrokerNodeId}
	if err := nodeIdField.Write(buf, version, o.nodeId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"nodeId\" field")
	}
	hostField := fields.String{Context: metadataResponseBrokerHost}
	if err := hostField.Write(buf, version, o.host); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"host\" field")
	}
	portField := fields.Int32{Context: metadataResponseBrokerPort}
	if err := portField.Write(buf, version, o.port); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"port\" field")
	}
	rackField := fields.String{Context: metadataResponseBrokerRack}
	if err := rackField.Write(buf, version, o.rack); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"rack\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < MetadataResponseBrokerLowestSupportedFlexVersion() || version > MetadataResponseBrokerHighestSupportedFlexVersion() {
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

func (o *MetadataResponseBroker) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *MetadataResponseBroker) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.nodeId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"nodeId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.host); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"host\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.port); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"port\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.rack); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"rack\""), b}, []byte(": ")))
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

func (o *MetadataResponseBroker) IsNil() bool {
	return o.isNil
}

func (o *MetadataResponseBroker) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *MetadataResponseBroker) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.nodeId = 0
	o.host.SetValue("")
	o.port = 0
	o.rack.SetValue("null")

	o.isNil = false
}

func (o *MetadataResponseBroker) Equal(that *MetadataResponseBroker) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.nodeId != that.nodeId {
		return false
	}
	if !o.host.Equal(&that.host) {
		return false
	}
	if o.port != that.port {
		return false
	}
	if !o.rack.Equal(&that.rack) {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *MetadataResponseBroker) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	nodeIdField := fields.Int32{Context: metadataResponseBrokerNodeId}
	fieldSize, err = nodeIdField.SizeInBytes(version, o.nodeId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"nodeId\" field")
	}
	size += fieldSize

	hostField := fields.String{Context: metadataResponseBrokerHost}
	fieldSize, err = hostField.SizeInBytes(version, o.host)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"host\" field")
	}
	size += fieldSize

	portField := fields.Int32{Context: metadataResponseBrokerPort}
	fieldSize, err = portField.SizeInBytes(version, o.port)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"port\" field")
	}
	size += fieldSize

	rackField := fields.String{Context: metadataResponseBrokerRack}
	fieldSize, err = rackField.SizeInBytes(version, o.rack)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"rack\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < MetadataResponseBrokerLowestSupportedFlexVersion() || version > MetadataResponseBrokerHighestSupportedFlexVersion() {
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
func (o *MetadataResponseBroker) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.host.Release()
	o.rack.Release()
}

func (o *MetadataResponseBroker) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *MetadataResponseBroker) validateNonIgnorableFields(version int16) error {
	return nil
}

func MetadataResponseBrokerLowestSupportedVersion() int16 {
	return 0
}

func MetadataResponseBrokerHighestSupportedVersion() int16 {
	return 32767
}

func MetadataResponseBrokerLowestSupportedFlexVersion() int16 {
	return 9
}

func MetadataResponseBrokerHighestSupportedFlexVersion() int16 {
	return 32767
}

func MetadataResponseBrokerDefault() MetadataResponseBroker {
	var d MetadataResponseBroker
	d.SetDefault()

	return d
}
