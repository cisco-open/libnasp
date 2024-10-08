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
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var listenerName = fields.Context{
	SpecName:                    "Name",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var listenerHost = fields.Context{
	SpecName:                    "Host",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var listenerPort = fields.Context{
	SpecName:                    "Port",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var listenerSecurityProtocol = fields.Context{
	SpecName:                    "SecurityProtocol",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}

type Listener struct {
	unknownTaggedFields []fields.RawTaggedField
	name                fields.NullableString
	host                fields.NullableString
	port                uint16
	securityProtocol    int16
	isNil               bool
}

func (o *Listener) Name() fields.NullableString {
	return o.name
}

func (o *Listener) SetName(val fields.NullableString) {
	o.isNil = false
	o.name = val
}

func (o *Listener) Host() fields.NullableString {
	return o.host
}

func (o *Listener) SetHost(val fields.NullableString) {
	o.isNil = false
	o.host = val
}

func (o *Listener) Port() uint16 {
	return o.port
}

func (o *Listener) SetPort(val uint16) {
	o.isNil = false
	o.port = val
}

func (o *Listener) SecurityProtocol() int16 {
	return o.securityProtocol
}

func (o *Listener) SetSecurityProtocol(val int16) {
	o.isNil = false
	o.securityProtocol = val
}

func (o *Listener) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Listener) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Listener) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	nameField := fields.String{Context: listenerName}
	if err := nameField.Read(buf, version, &o.name); err != nil {
		return errors.WrapIf(err, "couldn't set \"name\" field")
	}

	hostField := fields.String{Context: listenerHost}
	if err := hostField.Read(buf, version, &o.host); err != nil {
		return errors.WrapIf(err, "couldn't set \"host\" field")
	}

	portField := fields.Uint16{Context: listenerPort}
	if err := portField.Read(buf, version, &o.port); err != nil {
		return errors.WrapIf(err, "couldn't set \"port\" field")
	}

	securityProtocolField := fields.Int16{Context: listenerSecurityProtocol}
	if err := securityProtocolField.Read(buf, version, &o.securityProtocol); err != nil {
		return errors.WrapIf(err, "couldn't set \"securityProtocol\" field")
	}

	// process tagged fields

	if version < ListenerLowestSupportedFlexVersion() || version > ListenerHighestSupportedFlexVersion() {
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

func (o *Listener) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	nameField := fields.String{Context: listenerName}
	if err := nameField.Write(buf, version, o.name); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"name\" field")
	}
	hostField := fields.String{Context: listenerHost}
	if err := hostField.Write(buf, version, o.host); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"host\" field")
	}
	portField := fields.Uint16{Context: listenerPort}
	if err := portField.Write(buf, version, o.port); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"port\" field")
	}
	securityProtocolField := fields.Int16{Context: listenerSecurityProtocol}
	if err := securityProtocolField.Write(buf, version, o.securityProtocol); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"securityProtocol\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ListenerLowestSupportedFlexVersion() || version > ListenerHighestSupportedFlexVersion() {
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

func (o *Listener) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *Listener) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 5)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.name); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"name\""), b}, []byte(": ")))
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

func (o *Listener) IsNil() bool {
	return o.isNil
}

func (o *Listener) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *Listener) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.name.SetValue("")
	o.host.SetValue("")
	o.port = 0
	o.securityProtocol = 0

	o.isNil = false
}

func (o *Listener) Equal(that *Listener) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.name.Equal(&that.name) {
		return false
	}
	if !o.host.Equal(&that.host) {
		return false
	}
	if o.port != that.port {
		return false
	}
	if o.securityProtocol != that.securityProtocol {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *Listener) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	nameField := fields.String{Context: listenerName}
	fieldSize, err = nameField.SizeInBytes(version, o.name)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"name\" field")
	}
	size += fieldSize

	hostField := fields.String{Context: listenerHost}
	fieldSize, err = hostField.SizeInBytes(version, o.host)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"host\" field")
	}
	size += fieldSize

	portField := fields.Uint16{Context: listenerPort}
	fieldSize, err = portField.SizeInBytes(version, o.port)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"port\" field")
	}
	size += fieldSize

	securityProtocolField := fields.Int16{Context: listenerSecurityProtocol}
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
	if version < ListenerLowestSupportedFlexVersion() || version > ListenerHighestSupportedFlexVersion() {
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
func (o *Listener) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.name.Release()
	o.host.Release()
}

func (o *Listener) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Listener) validateNonIgnorableFields(version int16) error {
	return nil
}

func ListenerLowestSupportedVersion() int16 {
	return 0
}

func ListenerHighestSupportedVersion() int16 {
	return 32767
}

func ListenerLowestSupportedFlexVersion() int16 {
	return 0
}

func ListenerHighestSupportedFlexVersion() int16 {
	return 32767
}

func ListenerDefault() Listener {
	var d Listener
	d.SetDefault()

	return d
}
