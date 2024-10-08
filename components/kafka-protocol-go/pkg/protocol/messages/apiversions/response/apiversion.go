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
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var apiVersionApiKey = fields.Context{
	SpecName:                    "ApiKey",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var apiVersionMinVersion = fields.Context{
	SpecName:                    "MinVersion",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}
var apiVersionMaxVersion = fields.Context{
	SpecName:                    "MaxVersion",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  3,
	HighestSupportedFlexVersion: 32767,
}

type ApiVersion struct {
	unknownTaggedFields []fields.RawTaggedField
	apiKey              int16
	minVersion          int16
	maxVersion          int16
	isNil               bool
}

func (o *ApiVersion) ApiKey() int16 {
	return o.apiKey
}

func (o *ApiVersion) SetApiKey(val int16) {
	o.isNil = false
	o.apiKey = val
}

func (o *ApiVersion) MinVersion() int16 {
	return o.minVersion
}

func (o *ApiVersion) SetMinVersion(val int16) {
	o.isNil = false
	o.minVersion = val
}

func (o *ApiVersion) MaxVersion() int16 {
	return o.maxVersion
}

func (o *ApiVersion) SetMaxVersion(val int16) {
	o.isNil = false
	o.maxVersion = val
}

func (o *ApiVersion) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *ApiVersion) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *ApiVersion) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	apiKeyField := fields.Int16{Context: apiVersionApiKey}
	if err := apiKeyField.Read(buf, version, &o.apiKey); err != nil {
		return errors.WrapIf(err, "couldn't set \"apiKey\" field")
	}

	minVersionField := fields.Int16{Context: apiVersionMinVersion}
	if err := minVersionField.Read(buf, version, &o.minVersion); err != nil {
		return errors.WrapIf(err, "couldn't set \"minVersion\" field")
	}

	maxVersionField := fields.Int16{Context: apiVersionMaxVersion}
	if err := maxVersionField.Read(buf, version, &o.maxVersion); err != nil {
		return errors.WrapIf(err, "couldn't set \"maxVersion\" field")
	}

	// process tagged fields

	if version < ApiVersionLowestSupportedFlexVersion() || version > ApiVersionHighestSupportedFlexVersion() {
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

func (o *ApiVersion) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	apiKeyField := fields.Int16{Context: apiVersionApiKey}
	if err := apiKeyField.Write(buf, version, o.apiKey); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"apiKey\" field")
	}
	minVersionField := fields.Int16{Context: apiVersionMinVersion}
	if err := minVersionField.Write(buf, version, o.minVersion); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"minVersion\" field")
	}
	maxVersionField := fields.Int16{Context: apiVersionMaxVersion}
	if err := maxVersionField.Write(buf, version, o.maxVersion); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"maxVersion\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ApiVersionLowestSupportedFlexVersion() || version > ApiVersionHighestSupportedFlexVersion() {
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

func (o *ApiVersion) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *ApiVersion) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 4)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.apiKey); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"apiKey\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.minVersion); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"minVersion\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.maxVersion); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"maxVersion\""), b}, []byte(": ")))
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

func (o *ApiVersion) IsNil() bool {
	return o.isNil
}

func (o *ApiVersion) Clear() {
	o.Release()
	o.isNil = true

	o.unknownTaggedFields = nil
}

func (o *ApiVersion) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.apiKey = 0
	o.minVersion = 0
	o.maxVersion = 0

	o.isNil = false
}

func (o *ApiVersion) Equal(that *ApiVersion) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.apiKey != that.apiKey {
		return false
	}
	if o.minVersion != that.minVersion {
		return false
	}
	if o.maxVersion != that.maxVersion {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *ApiVersion) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	apiKeyField := fields.Int16{Context: apiVersionApiKey}
	fieldSize, err = apiKeyField.SizeInBytes(version, o.apiKey)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"apiKey\" field")
	}
	size += fieldSize

	minVersionField := fields.Int16{Context: apiVersionMinVersion}
	fieldSize, err = minVersionField.SizeInBytes(version, o.minVersion)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"minVersion\" field")
	}
	size += fieldSize

	maxVersionField := fields.Int16{Context: apiVersionMaxVersion}
	fieldSize, err = maxVersionField.SizeInBytes(version, o.maxVersion)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"maxVersion\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < ApiVersionLowestSupportedFlexVersion() || version > ApiVersionHighestSupportedFlexVersion() {
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
func (o *ApiVersion) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

}

func (o *ApiVersion) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *ApiVersion) validateNonIgnorableFields(version int16) error {
	return nil
}

func ApiVersionLowestSupportedVersion() int16 {
	return 0
}

func ApiVersionHighestSupportedVersion() int16 {
	return 32767
}

func ApiVersionLowestSupportedFlexVersion() int16 {
	return 3
}

func ApiVersionHighestSupportedFlexVersion() int16 {
	return 32767
}

func ApiVersionDefault() ApiVersion {
	var d ApiVersion
	d.SetDefault()

	return d
}
