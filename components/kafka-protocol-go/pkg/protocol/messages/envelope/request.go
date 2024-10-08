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
package envelope

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/pools"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var requestRequestData = fields.Context{
	SpecName:                    "RequestData",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var requestRequestPrincipal = fields.Context{
	SpecName:                        "RequestPrincipal",
	LowestSupportedVersion:          0,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      0,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  0,
	HighestSupportedNullableVersion: 32767,
}
var requestClientHostAddress = fields.Context{
	SpecName:                    "ClientHostAddress",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}

type Request struct {

	// The embedded request header and data.
	requestData []byte
	// Value of the initial client principal when the request is redirected by a broker.
	requestPrincipal []byte
	// The original client's address in bytes.
	clientHostAddress   []byte
	unknownTaggedFields []fields.RawTaggedField

	isNil bool
}

func (o *Request) RequestData() []byte {
	return o.requestData
}

func (o *Request) SetRequestData(val []byte) {
	o.isNil = false
	o.requestData = val
}

func (o *Request) RequestPrincipal() []byte {
	return o.requestPrincipal
}

func (o *Request) SetRequestPrincipal(val []byte) {
	o.isNil = false
	o.requestPrincipal = val
}

func (o *Request) ClientHostAddress() []byte {
	return o.clientHostAddress
}

func (o *Request) SetClientHostAddress(val []byte) {
	o.isNil = false
	o.clientHostAddress = val
}

func (o *Request) ApiKey() int16 {
	return 58
}

func (o *Request) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Request) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Request) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	requestDataField := fields.Bytes{Context: requestRequestData}
	if err := requestDataField.Read(buf, version, &o.requestData); err != nil {
		return errors.WrapIf(err, "couldn't set \"requestData\" field")
	}

	requestPrincipalField := fields.Bytes{Context: requestRequestPrincipal}
	if err := requestPrincipalField.Read(buf, version, &o.requestPrincipal); err != nil {
		return errors.WrapIf(err, "couldn't set \"requestPrincipal\" field")
	}

	clientHostAddressField := fields.Bytes{Context: requestClientHostAddress}
	if err := clientHostAddressField.Read(buf, version, &o.clientHostAddress); err != nil {
		return errors.WrapIf(err, "couldn't set \"clientHostAddress\" field")
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

	requestDataField := fields.Bytes{Context: requestRequestData}
	if err := requestDataField.Write(buf, version, o.requestData); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"requestData\" field")
	}
	requestPrincipalField := fields.Bytes{Context: requestRequestPrincipal}
	if err := requestPrincipalField.Write(buf, version, o.requestPrincipal); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"requestPrincipal\" field")
	}
	clientHostAddressField := fields.Bytes{Context: requestClientHostAddress}
	if err := clientHostAddressField.Write(buf, version, o.clientHostAddress); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"clientHostAddress\" field")
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

	s := make([][]byte, 0, 4)
	if b, err := fields.BytesMarshalJSON("requestData", o.requestData); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.BytesMarshalJSON("requestPrincipal", o.requestPrincipal); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.BytesMarshalJSON("clientHostAddress", o.clientHostAddress); err != nil {
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
	if o.requestData != nil {
		pools.ReleaseByteSlice(o.requestData)
	}
	o.requestData = nil
	if o.requestPrincipal != nil {
		pools.ReleaseByteSlice(o.requestPrincipal)
	}
	o.requestPrincipal = nil
	if o.clientHostAddress != nil {
		pools.ReleaseByteSlice(o.clientHostAddress)
	}
	o.clientHostAddress = nil

	o.isNil = false
}

func (o *Request) Equal(that *Request) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !bytes.Equal(o.requestData, that.requestData) {
		return false
	}
	if !bytes.Equal(o.requestPrincipal, that.requestPrincipal) {
		return false
	}
	if !bytes.Equal(o.clientHostAddress, that.clientHostAddress) {
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

	requestDataField := fields.Bytes{Context: requestRequestData}
	fieldSize, err = requestDataField.SizeInBytes(version, o.requestData)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"requestData\" field")
	}
	size += fieldSize

	requestPrincipalField := fields.Bytes{Context: requestRequestPrincipal}
	fieldSize, err = requestPrincipalField.SizeInBytes(version, o.requestPrincipal)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"requestPrincipal\" field")
	}
	size += fieldSize

	clientHostAddressField := fields.Bytes{Context: requestClientHostAddress}
	fieldSize, err = clientHostAddressField.SizeInBytes(version, o.clientHostAddress)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"clientHostAddress\" field")
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

	if o.requestData != nil {
		pools.ReleaseByteSlice(o.requestData)
	}
	if o.requestPrincipal != nil {
		pools.ReleaseByteSlice(o.requestPrincipal)
	}
	if o.clientHostAddress != nil {
		pools.ReleaseByteSlice(o.clientHostAddress)
	}
}

func (o *Request) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Request) validateNonIgnorableFields(version int16) error {
	return nil
}

func RequestLowestSupportedVersion() int16 {
	return 0
}

func RequestHighestSupportedVersion() int16 {
	return 0
}

func RequestLowestSupportedFlexVersion() int16 {
	return 0
}

func RequestHighestSupportedFlexVersion() int16 {
	return 32767
}

func RequestDefault() Request {
	var d Request
	d.SetDefault()

	return d
}
