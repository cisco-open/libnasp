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
package brokerheartbeat

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var requestBrokerId = fields.Context{
	SpecName:                    "BrokerId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var requestBrokerEpoch = fields.Context{
	SpecName:                    "BrokerEpoch",
	CustomDefaultValue:          int64(-1),
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var requestCurrentMetadataOffset = fields.Context{
	SpecName:                    "CurrentMetadataOffset",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var requestWantFence = fields.Context{
	SpecName:                    "WantFence",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}
var requestWantShutDown = fields.Context{
	SpecName:                    "WantShutDown",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  0,
	HighestSupportedFlexVersion: 32767,
}

type Request struct {
	unknownTaggedFields   []fields.RawTaggedField
	brokerEpoch           int64
	currentMetadataOffset int64
	brokerId              int32
	wantFence             bool
	wantShutDown          bool
	isNil                 bool
}

func (o *Request) BrokerId() int32 {
	return o.brokerId
}

func (o *Request) SetBrokerId(val int32) {
	o.isNil = false
	o.brokerId = val
}

func (o *Request) BrokerEpoch() int64 {
	return o.brokerEpoch
}

func (o *Request) SetBrokerEpoch(val int64) {
	o.isNil = false
	o.brokerEpoch = val
}

func (o *Request) CurrentMetadataOffset() int64 {
	return o.currentMetadataOffset
}

func (o *Request) SetCurrentMetadataOffset(val int64) {
	o.isNil = false
	o.currentMetadataOffset = val
}

func (o *Request) WantFence() bool {
	return o.wantFence
}

func (o *Request) SetWantFence(val bool) {
	o.isNil = false
	o.wantFence = val
}

func (o *Request) WantShutDown() bool {
	return o.wantShutDown
}

func (o *Request) SetWantShutDown(val bool) {
	o.isNil = false
	o.wantShutDown = val
}

func (o *Request) ApiKey() int16 {
	return 63
}

func (o *Request) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Request) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Request) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	brokerIdField := fields.Int32{Context: requestBrokerId}
	if err := brokerIdField.Read(buf, version, &o.brokerId); err != nil {
		return errors.WrapIf(err, "couldn't set \"brokerId\" field")
	}

	brokerEpochField := fields.Int64{Context: requestBrokerEpoch}
	if err := brokerEpochField.Read(buf, version, &o.brokerEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"brokerEpoch\" field")
	}

	currentMetadataOffsetField := fields.Int64{Context: requestCurrentMetadataOffset}
	if err := currentMetadataOffsetField.Read(buf, version, &o.currentMetadataOffset); err != nil {
		return errors.WrapIf(err, "couldn't set \"currentMetadataOffset\" field")
	}

	wantFenceField := fields.Bool{Context: requestWantFence}
	if err := wantFenceField.Read(buf, version, &o.wantFence); err != nil {
		return errors.WrapIf(err, "couldn't set \"wantFence\" field")
	}

	wantShutDownField := fields.Bool{Context: requestWantShutDown}
	if err := wantShutDownField.Read(buf, version, &o.wantShutDown); err != nil {
		return errors.WrapIf(err, "couldn't set \"wantShutDown\" field")
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

	brokerIdField := fields.Int32{Context: requestBrokerId}
	if err := brokerIdField.Write(buf, version, o.brokerId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"brokerId\" field")
	}
	brokerEpochField := fields.Int64{Context: requestBrokerEpoch}
	if err := brokerEpochField.Write(buf, version, o.brokerEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"brokerEpoch\" field")
	}
	currentMetadataOffsetField := fields.Int64{Context: requestCurrentMetadataOffset}
	if err := currentMetadataOffsetField.Write(buf, version, o.currentMetadataOffset); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"currentMetadataOffset\" field")
	}
	wantFenceField := fields.Bool{Context: requestWantFence}
	if err := wantFenceField.Write(buf, version, o.wantFence); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"wantFence\" field")
	}
	wantShutDownField := fields.Bool{Context: requestWantShutDown}
	if err := wantShutDownField.Write(buf, version, o.wantShutDown); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"wantShutDown\" field")
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

	s := make([][]byte, 0, 6)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.brokerId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"brokerId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.brokerEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"brokerEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.currentMetadataOffset); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"currentMetadataOffset\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.wantFence); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"wantFence\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.wantShutDown); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"wantShutDown\""), b}, []byte(": ")))
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
	o.brokerId = 0
	o.brokerEpoch = -1
	o.currentMetadataOffset = 0
	o.wantFence = false
	o.wantShutDown = false

	o.isNil = false
}

func (o *Request) Equal(that *Request) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.brokerId != that.brokerId {
		return false
	}
	if o.brokerEpoch != that.brokerEpoch {
		return false
	}
	if o.currentMetadataOffset != that.currentMetadataOffset {
		return false
	}
	if o.wantFence != that.wantFence {
		return false
	}
	if o.wantShutDown != that.wantShutDown {
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

	brokerIdField := fields.Int32{Context: requestBrokerId}
	fieldSize, err = brokerIdField.SizeInBytes(version, o.brokerId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"brokerId\" field")
	}
	size += fieldSize

	brokerEpochField := fields.Int64{Context: requestBrokerEpoch}
	fieldSize, err = brokerEpochField.SizeInBytes(version, o.brokerEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"brokerEpoch\" field")
	}
	size += fieldSize

	currentMetadataOffsetField := fields.Int64{Context: requestCurrentMetadataOffset}
	fieldSize, err = currentMetadataOffsetField.SizeInBytes(version, o.currentMetadataOffset)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"currentMetadataOffset\" field")
	}
	size += fieldSize

	wantFenceField := fields.Bool{Context: requestWantFence}
	fieldSize, err = wantFenceField.SizeInBytes(version, o.wantFence)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"wantFence\" field")
	}
	size += fieldSize

	wantShutDownField := fields.Bool{Context: requestWantShutDown}
	fieldSize, err = wantShutDownField.SizeInBytes(version, o.wantShutDown)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"wantShutDown\" field")
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
