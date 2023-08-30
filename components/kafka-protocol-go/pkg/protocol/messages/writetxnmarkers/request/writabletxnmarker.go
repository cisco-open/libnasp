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
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/writetxnmarkers/request/writabletxnmarker"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var writableTxnMarkerProducerId = fields.Context{
	SpecName:                    "ProducerId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  1,
	HighestSupportedFlexVersion: 32767,
}
var writableTxnMarkerProducerEpoch = fields.Context{
	SpecName:                    "ProducerEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  1,
	HighestSupportedFlexVersion: 32767,
}
var writableTxnMarkerTransactionResult = fields.Context{
	SpecName:                    "TransactionResult",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  1,
	HighestSupportedFlexVersion: 32767,
}
var writableTxnMarkerTopics = fields.Context{
	SpecName:                    "Topics",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  1,
	HighestSupportedFlexVersion: 32767,
}
var writableTxnMarkerCoordinatorEpoch = fields.Context{
	SpecName:                    "CoordinatorEpoch",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  1,
	HighestSupportedFlexVersion: 32767,
}

type WritableTxnMarker struct {
	topics              []writabletxnmarker.WritableTxnMarkerTopic
	unknownTaggedFields []fields.RawTaggedField
	producerId          int64
	coordinatorEpoch    int32
	producerEpoch       int16
	transactionResult   bool
	isNil               bool
}

func (o *WritableTxnMarker) ProducerId() int64 {
	return o.producerId
}

func (o *WritableTxnMarker) SetProducerId(val int64) {
	o.isNil = false
	o.producerId = val
}

func (o *WritableTxnMarker) ProducerEpoch() int16 {
	return o.producerEpoch
}

func (o *WritableTxnMarker) SetProducerEpoch(val int16) {
	o.isNil = false
	o.producerEpoch = val
}

func (o *WritableTxnMarker) TransactionResult() bool {
	return o.transactionResult
}

func (o *WritableTxnMarker) SetTransactionResult(val bool) {
	o.isNil = false
	o.transactionResult = val
}

func (o *WritableTxnMarker) Topics() []writabletxnmarker.WritableTxnMarkerTopic {
	return o.topics
}

func (o *WritableTxnMarker) SetTopics(val []writabletxnmarker.WritableTxnMarkerTopic) {
	o.isNil = false
	o.topics = val
}

func (o *WritableTxnMarker) CoordinatorEpoch() int32 {
	return o.coordinatorEpoch
}

func (o *WritableTxnMarker) SetCoordinatorEpoch(val int32) {
	o.isNil = false
	o.coordinatorEpoch = val
}

func (o *WritableTxnMarker) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *WritableTxnMarker) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *WritableTxnMarker) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	producerIdField := fields.Int64{Context: writableTxnMarkerProducerId}
	if err := producerIdField.Read(buf, version, &o.producerId); err != nil {
		return errors.WrapIf(err, "couldn't set \"producerId\" field")
	}

	producerEpochField := fields.Int16{Context: writableTxnMarkerProducerEpoch}
	if err := producerEpochField.Read(buf, version, &o.producerEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"producerEpoch\" field")
	}

	transactionResultField := fields.Bool{Context: writableTxnMarkerTransactionResult}
	if err := transactionResultField.Read(buf, version, &o.transactionResult); err != nil {
		return errors.WrapIf(err, "couldn't set \"transactionResult\" field")
	}

	topicsField := fields.ArrayOfStruct[writabletxnmarker.WritableTxnMarkerTopic, *writabletxnmarker.WritableTxnMarkerTopic]{Context: writableTxnMarkerTopics}
	topics, err := topicsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"topics\" field")
	}
	o.topics = topics

	coordinatorEpochField := fields.Int32{Context: writableTxnMarkerCoordinatorEpoch}
	if err := coordinatorEpochField.Read(buf, version, &o.coordinatorEpoch); err != nil {
		return errors.WrapIf(err, "couldn't set \"coordinatorEpoch\" field")
	}

	// process tagged fields

	if version < WritableTxnMarkerLowestSupportedFlexVersion() || version > WritableTxnMarkerHighestSupportedFlexVersion() {
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

func (o *WritableTxnMarker) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	producerIdField := fields.Int64{Context: writableTxnMarkerProducerId}
	if err := producerIdField.Write(buf, version, o.producerId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"producerId\" field")
	}
	producerEpochField := fields.Int16{Context: writableTxnMarkerProducerEpoch}
	if err := producerEpochField.Write(buf, version, o.producerEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"producerEpoch\" field")
	}
	transactionResultField := fields.Bool{Context: writableTxnMarkerTransactionResult}
	if err := transactionResultField.Write(buf, version, o.transactionResult); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"transactionResult\" field")
	}

	topicsField := fields.ArrayOfStruct[writabletxnmarker.WritableTxnMarkerTopic, *writabletxnmarker.WritableTxnMarkerTopic]{Context: writableTxnMarkerTopics}
	if err := topicsField.Write(buf, version, o.topics); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topics\" field")
	}

	coordinatorEpochField := fields.Int32{Context: writableTxnMarkerCoordinatorEpoch}
	if err := coordinatorEpochField.Write(buf, version, o.coordinatorEpoch); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"coordinatorEpoch\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < WritableTxnMarkerLowestSupportedFlexVersion() || version > WritableTxnMarkerHighestSupportedFlexVersion() {
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

func (o *WritableTxnMarker) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *WritableTxnMarker) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 6)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.producerId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"producerId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.producerEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"producerEpoch\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.transactionResult); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"transactionResult\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("topics", o.topics); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.coordinatorEpoch); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"coordinatorEpoch\""), b}, []byte(": ")))
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

func (o *WritableTxnMarker) IsNil() bool {
	return o.isNil
}

func (o *WritableTxnMarker) Clear() {
	o.Release()
	o.isNil = true

	o.topics = nil
	o.unknownTaggedFields = nil
}

func (o *WritableTxnMarker) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.producerId = 0
	o.producerEpoch = 0
	o.transactionResult = false
	for i := range o.topics {
		o.topics[i].Release()
	}
	o.topics = nil
	o.coordinatorEpoch = 0

	o.isNil = false
}

func (o *WritableTxnMarker) Equal(that *WritableTxnMarker) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.producerId != that.producerId {
		return false
	}
	if o.producerEpoch != that.producerEpoch {
		return false
	}
	if o.transactionResult != that.transactionResult {
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
	if o.coordinatorEpoch != that.coordinatorEpoch {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *WritableTxnMarker) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	producerIdField := fields.Int64{Context: writableTxnMarkerProducerId}
	fieldSize, err = producerIdField.SizeInBytes(version, o.producerId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"producerId\" field")
	}
	size += fieldSize

	producerEpochField := fields.Int16{Context: writableTxnMarkerProducerEpoch}
	fieldSize, err = producerEpochField.SizeInBytes(version, o.producerEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"producerEpoch\" field")
	}
	size += fieldSize

	transactionResultField := fields.Bool{Context: writableTxnMarkerTransactionResult}
	fieldSize, err = transactionResultField.SizeInBytes(version, o.transactionResult)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"transactionResult\" field")
	}
	size += fieldSize

	topicsField := fields.ArrayOfStruct[writabletxnmarker.WritableTxnMarkerTopic, *writabletxnmarker.WritableTxnMarkerTopic]{Context: writableTxnMarkerTopics}
	fieldSize, err = topicsField.SizeInBytes(version, o.topics)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topics\" field")
	}
	size += fieldSize

	coordinatorEpochField := fields.Int32{Context: writableTxnMarkerCoordinatorEpoch}
	fieldSize, err = coordinatorEpochField.SizeInBytes(version, o.coordinatorEpoch)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"coordinatorEpoch\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < WritableTxnMarkerLowestSupportedFlexVersion() || version > WritableTxnMarkerHighestSupportedFlexVersion() {
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
func (o *WritableTxnMarker) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	for i := range o.topics {
		o.topics[i].Release()
	}
	o.topics = nil
}

func (o *WritableTxnMarker) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *WritableTxnMarker) validateNonIgnorableFields(version int16) error {
	return nil
}

func WritableTxnMarkerLowestSupportedVersion() int16 {
	return 0
}

func WritableTxnMarkerHighestSupportedVersion() int16 {
	return 32767
}

func WritableTxnMarkerLowestSupportedFlexVersion() int16 {
	return 1
}

func WritableTxnMarkerHighestSupportedFlexVersion() int16 {
	return 32767
}

func WritableTxnMarkerDefault() WritableTxnMarker {
	var d WritableTxnMarker
	d.SetDefault()

	return d
}
