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
package updatemetadatabroker

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var updateMetadataEndpointPort = fields.Context{
	SpecName:                    "Port",
	LowestSupportedVersion:      1,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var updateMetadataEndpointHost = fields.Context{
	SpecName:                    "Host",
	LowestSupportedVersion:      1,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var updateMetadataEndpointListener = fields.Context{
	SpecName:                    "Listener",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}
var updateMetadataEndpointSecurityProtocol = fields.Context{
	SpecName:                    "SecurityProtocol",
	LowestSupportedVersion:      1,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  6,
	HighestSupportedFlexVersion: 32767,
}

type UpdateMetadataEndpoint struct {
	unknownTaggedFields []fields.RawTaggedField
	host                fields.NullableString
	listener            fields.NullableString
	port                int32
	securityProtocol    int16
	isNil               bool
}

func (o *UpdateMetadataEndpoint) Port() int32 {
	return o.port
}

func (o *UpdateMetadataEndpoint) SetPort(val int32) {
	o.isNil = false
	o.port = val
}

func (o *UpdateMetadataEndpoint) Host() fields.NullableString {
	return o.host
}

func (o *UpdateMetadataEndpoint) SetHost(val fields.NullableString) {
	o.isNil = false
	o.host = val
}

func (o *UpdateMetadataEndpoint) Listener() fields.NullableString {
	return o.listener
}

func (o *UpdateMetadataEndpoint) SetListener(val fields.NullableString) {
	o.isNil = false
	o.listener = val
}

func (o *UpdateMetadataEndpoint) SecurityProtocol() int16 {
	return o.securityProtocol
}

func (o *UpdateMetadataEndpoint) SetSecurityProtocol(val int16) {
	o.isNil = false
	o.securityProtocol = val
}

func (o *UpdateMetadataEndpoint) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *UpdateMetadataEndpoint) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *UpdateMetadataEndpoint) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	portField := fields.Int32{Context: updateMetadataEndpointPort}
	if err := portField.Read(buf, version, &o.port); err != nil {
		return errors.WrapIf(err, "couldn't set \"port\" field")
	}

	hostField := fields.String{Context: updateMetadataEndpointHost}
	if err := hostField.Read(buf, version, &o.host); err != nil {
		return errors.WrapIf(err, "couldn't set \"host\" field")
	}

	listenerField := fields.String{Context: updateMetadataEndpointListener}
	if err := listenerField.Read(buf, version, &o.listener); err != nil {
		return errors.WrapIf(err, "couldn't set \"listener\" field")
	}

	securityProtocolField := fields.Int16{Context: updateMetadataEndpointSecurityProtocol}
	if err := securityProtocolField.Read(buf, version, &o.securityProtocol); err != nil {
		return errors.WrapIf(err, "couldn't set \"securityProtocol\" field")
	}

	// process tagged fields

	if version < UpdateMetadataEndpointLowestSupportedFlexVersion() || version > UpdateMetadataEndpointHighestSupportedFlexVersion() {
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

func (o *UpdateMetadataEndpoint) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	portField := fields.Int32{Context: updateMetadataEndpointPort}
	if err := portField.Write(buf, version, o.port); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"port\" field")
	}
	hostField := fields.String{Context: updateMetadataEndpointHost}
	if err := hostField.Write(buf, version, o.host); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"host\" field")
	}
	listenerField := fields.String{Context: updateMetadataEndpointListener}
	if err := listenerField.Write(buf, version, o.listener); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"listener\" field")
	}
	securityProtocolField := fields.Int16{Context: updateMetadataEndpointSecurityProtocol}
	if err := securityProtocolField.Write(buf, version, o.securityProtocol); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"securityProtocol\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < UpdateMetadataEndpointLowestSupportedFlexVersion() || version > UpdateMetadataEndpointHighestSupportedFlexVersion() {
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

func (o *UpdateMetadataEndpoint) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *UpdateMetadataEndpoint) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.port); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"port\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.host); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"host\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.listener); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"listener\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.securityProtocol); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"securityProtocol\""), b}, []byte(": ")))
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

func (o *UpdateMetadataEndpoint) IsNil() bool {
	return o.isNil
}

func (o *UpdateMetadataEndpoint) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *UpdateMetadataEndpoint) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.port = 0
	o.host.SetValue("")
	o.listener.SetValue("")
	o.securityProtocol = 0

	o.isNil = false
}

func (o *UpdateMetadataEndpoint) Equal(that *UpdateMetadataEndpoint) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.port != that.port {
		return false
	}
	if !o.host.Equal(&that.host) {
		return false
	}
	if !o.listener.Equal(&that.listener) {
		return false
	}
	if o.securityProtocol != that.securityProtocol {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *UpdateMetadataEndpoint) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	portField := fields.Int32{Context: updateMetadataEndpointPort}
	fieldSize, err = portField.SizeInBytes(version, o.port)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"port\" field")
	}
	size += fieldSize

	hostField := fields.String{Context: updateMetadataEndpointHost}
	fieldSize, err = hostField.SizeInBytes(version, o.host)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"host\" field")
	}
	size += fieldSize

	listenerField := fields.String{Context: updateMetadataEndpointListener}
	fieldSize, err = listenerField.SizeInBytes(version, o.listener)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"listener\" field")
	}
	size += fieldSize

	securityProtocolField := fields.Int16{Context: updateMetadataEndpointSecurityProtocol}
	fieldSize, err = securityProtocolField.SizeInBytes(version, o.securityProtocol)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"securityProtocol\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < UpdateMetadataEndpointLowestSupportedFlexVersion() || version > UpdateMetadataEndpointHighestSupportedFlexVersion() {
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
func (o *UpdateMetadataEndpoint) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.host.Release()
	o.listener.Release()
}

func (o *UpdateMetadataEndpoint) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *UpdateMetadataEndpoint) validateNonIgnorableFields(version int16) error {
	if !updateMetadataEndpointPort.IsSupportedVersion(version) {
		if o.port != 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"port\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !updateMetadataEndpointHost.IsSupportedVersion(version) {
		if o.host.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"host\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !updateMetadataEndpointSecurityProtocol.IsSupportedVersion(version) {
		if o.securityProtocol != 0 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"securityProtocol\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func UpdateMetadataEndpointLowestSupportedVersion() int16 {
	return 1
}

func UpdateMetadataEndpointHighestSupportedVersion() int16 {
	return 32767
}

func UpdateMetadataEndpointLowestSupportedFlexVersion() int16 {
	return 6
}

func UpdateMetadataEndpointHighestSupportedFlexVersion() int16 {
	return 32767
}

func UpdateMetadataEndpointDefault() UpdateMetadataEndpoint {
	var d UpdateMetadataEndpoint
	d.SetDefault()

	return d
}