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
package metadata

import (
	"bytes"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/metadata/response"
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var responseThrottleTimeMs = fields.Context{
	SpecName:                    "ThrottleTimeMs",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var responseBrokers = fields.Context{
	SpecName:                    "Brokers",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var responseClusterId = fields.Context{
	SpecName:           "ClusterId",
	CustomDefaultValue: "null",

	LowestSupportedVersion:          2,
	HighestSupportedVersion:         32767,
	LowestSupportedFlexVersion:      9,
	HighestSupportedFlexVersion:     32767,
	LowestSupportedNullableVersion:  2,
	HighestSupportedNullableVersion: 32767,
}
var responseControllerId = fields.Context{
	SpecName:                    "ControllerId",
	CustomDefaultValue:          int32(-1),
	LowestSupportedVersion:      1,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var responseTopics = fields.Context{
	SpecName:                    "Topics",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}
var responseClusterAuthorizedOperations = fields.Context{
	SpecName:                    "ClusterAuthorizedOperations",
	CustomDefaultValue:          int32(-2147483648),
	LowestSupportedVersion:      8,
	HighestSupportedVersion:     10,
	LowestSupportedFlexVersion:  9,
	HighestSupportedFlexVersion: 32767,
}

type Response struct {
	brokers                     []response.MetadataResponseBroker
	topics                      []response.MetadataResponseTopic
	unknownTaggedFields         []fields.RawTaggedField
	clusterId                   fields.NullableString
	throttleTimeMs              int32
	controllerId                int32
	clusterAuthorizedOperations int32
	isNil                       bool
}

func (o *Response) ThrottleTimeMs() int32 {
	return o.throttleTimeMs
}

func (o *Response) SetThrottleTimeMs(val int32) {
	o.isNil = false
	o.throttleTimeMs = val
}

func (o *Response) Brokers() []response.MetadataResponseBroker {
	return o.brokers
}

func (o *Response) SetBrokers(val []response.MetadataResponseBroker) {
	o.isNil = false
	o.brokers = val
}

func (o *Response) ClusterId() fields.NullableString {
	return o.clusterId
}

func (o *Response) SetClusterId(val fields.NullableString) {
	o.isNil = false
	o.clusterId = val
}

func (o *Response) ControllerId() int32 {
	return o.controllerId
}

func (o *Response) SetControllerId(val int32) {
	o.isNil = false
	o.controllerId = val
}

func (o *Response) Topics() []response.MetadataResponseTopic {
	return o.topics
}

func (o *Response) SetTopics(val []response.MetadataResponseTopic) {
	o.isNil = false
	o.topics = val
}

func (o *Response) ClusterAuthorizedOperations() int32 {
	return o.clusterAuthorizedOperations
}

func (o *Response) SetClusterAuthorizedOperations(val int32) {
	o.isNil = false
	o.clusterAuthorizedOperations = val
}

func (o *Response) ApiKey() int16 {
	return 3
}

func (o *Response) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *Response) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *Response) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	throttleTimeMsField := fields.Int32{Context: responseThrottleTimeMs}
	if err := throttleTimeMsField.Read(buf, version, &o.throttleTimeMs); err != nil {
		return errors.WrapIf(err, "couldn't set \"throttleTimeMs\" field")
	}

	brokersField := fields.ArrayOfStruct[response.MetadataResponseBroker, *response.MetadataResponseBroker]{Context: responseBrokers}
	brokers, err := brokersField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"brokers\" field")
	}
	o.brokers = brokers

	clusterIdField := fields.String{Context: responseClusterId}
	if err := clusterIdField.Read(buf, version, &o.clusterId); err != nil {
		return errors.WrapIf(err, "couldn't set \"clusterId\" field")
	}

	controllerIdField := fields.Int32{Context: responseControllerId}
	if err := controllerIdField.Read(buf, version, &o.controllerId); err != nil {
		return errors.WrapIf(err, "couldn't set \"controllerId\" field")
	}

	topicsField := fields.ArrayOfStruct[response.MetadataResponseTopic, *response.MetadataResponseTopic]{Context: responseTopics}
	topics, err := topicsField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"topics\" field")
	}
	o.topics = topics

	clusterAuthorizedOperationsField := fields.Int32{Context: responseClusterAuthorizedOperations}
	if err := clusterAuthorizedOperationsField.Read(buf, version, &o.clusterAuthorizedOperations); err != nil {
		return errors.WrapIf(err, "couldn't set \"clusterAuthorizedOperations\" field")
	}

	// process tagged fields

	if version < ResponseLowestSupportedFlexVersion() || version > ResponseHighestSupportedFlexVersion() {
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

func (o *Response) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	throttleTimeMsField := fields.Int32{Context: responseThrottleTimeMs}
	if err := throttleTimeMsField.Write(buf, version, o.throttleTimeMs); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"throttleTimeMs\" field")
	}

	brokersField := fields.ArrayOfStruct[response.MetadataResponseBroker, *response.MetadataResponseBroker]{Context: responseBrokers}
	if err := brokersField.Write(buf, version, o.brokers); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"brokers\" field")
	}

	clusterIdField := fields.String{Context: responseClusterId}
	if err := clusterIdField.Write(buf, version, o.clusterId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"clusterId\" field")
	}
	controllerIdField := fields.Int32{Context: responseControllerId}
	if err := controllerIdField.Write(buf, version, o.controllerId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"controllerId\" field")
	}

	topicsField := fields.ArrayOfStruct[response.MetadataResponseTopic, *response.MetadataResponseTopic]{Context: responseTopics}
	if err := topicsField.Write(buf, version, o.topics); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"topics\" field")
	}

	clusterAuthorizedOperationsField := fields.Int32{Context: responseClusterAuthorizedOperations}
	if err := clusterAuthorizedOperationsField.Write(buf, version, o.clusterAuthorizedOperations); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"clusterAuthorizedOperations\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < ResponseLowestSupportedFlexVersion() || version > ResponseHighestSupportedFlexVersion() {
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

func (o *Response) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *Response) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 7)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.throttleTimeMs); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"throttleTimeMs\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("brokers", o.brokers); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.clusterId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"clusterId\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.controllerId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"controllerId\""), b}, []byte(": ")))
	}
	if b, err := fields.ArrayOfStructMarshalJSON("topics", o.topics); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.clusterAuthorizedOperations); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"clusterAuthorizedOperations\""), b}, []byte(": ")))
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

func (o *Response) IsNil() bool {
	return o.isNil
}

func (o *Response) Clear() {
	o.Release()
	o.isNil = true

	o.brokers = nil
	o.topics = nil
	o.unknownTaggedFields = nil
}

func (o *Response) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.throttleTimeMs = 0
	for i := range o.brokers {
		o.brokers[i].Release()
	}
	o.brokers = nil
	o.clusterId.SetValue("null")
	o.controllerId = -1
	for i := range o.topics {
		o.topics[i].Release()
	}
	o.topics = nil
	o.clusterAuthorizedOperations = -2147483648

	o.isNil = false
}

func (o *Response) Equal(that *Response) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if o.throttleTimeMs != that.throttleTimeMs {
		return false
	}
	if len(o.brokers) != len(that.brokers) {
		return false
	}
	for i := range o.brokers {
		if !o.brokers[i].Equal(&that.brokers[i]) {
			return false
		}
	}
	if !o.clusterId.Equal(&that.clusterId) {
		return false
	}
	if o.controllerId != that.controllerId {
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
	if o.clusterAuthorizedOperations != that.clusterAuthorizedOperations {
		return false
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *Response) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	throttleTimeMsField := fields.Int32{Context: responseThrottleTimeMs}
	fieldSize, err = throttleTimeMsField.SizeInBytes(version, o.throttleTimeMs)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"throttleTimeMs\" field")
	}
	size += fieldSize

	brokersField := fields.ArrayOfStruct[response.MetadataResponseBroker, *response.MetadataResponseBroker]{Context: responseBrokers}
	fieldSize, err = brokersField.SizeInBytes(version, o.brokers)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"brokers\" field")
	}
	size += fieldSize

	clusterIdField := fields.String{Context: responseClusterId}
	fieldSize, err = clusterIdField.SizeInBytes(version, o.clusterId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"clusterId\" field")
	}
	size += fieldSize

	controllerIdField := fields.Int32{Context: responseControllerId}
	fieldSize, err = controllerIdField.SizeInBytes(version, o.controllerId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"controllerId\" field")
	}
	size += fieldSize

	topicsField := fields.ArrayOfStruct[response.MetadataResponseTopic, *response.MetadataResponseTopic]{Context: responseTopics}
	fieldSize, err = topicsField.SizeInBytes(version, o.topics)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"topics\" field")
	}
	size += fieldSize

	clusterAuthorizedOperationsField := fields.Int32{Context: responseClusterAuthorizedOperations}
	fieldSize, err = clusterAuthorizedOperationsField.SizeInBytes(version, o.clusterAuthorizedOperations)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"clusterAuthorizedOperations\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < ResponseLowestSupportedFlexVersion() || version > ResponseHighestSupportedFlexVersion() {
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
func (o *Response) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	for i := range o.brokers {
		o.brokers[i].Release()
	}
	o.brokers = nil
	o.clusterId.Release()
	for i := range o.topics {
		o.topics[i].Release()
	}
	o.topics = nil
}

func (o *Response) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *Response) validateNonIgnorableFields(version int16) error {
	if !responseClusterAuthorizedOperations.IsSupportedVersion(version) {
		if o.clusterAuthorizedOperations != -2147483648 {
			return errors.New(strings.Join([]string{"attempted to write non-default \"clusterAuthorizedOperations\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func ResponseLowestSupportedVersion() int16 {
	return 0
}

func ResponseHighestSupportedVersion() int16 {
	return 12
}

func ResponseLowestSupportedFlexVersion() int16 {
	return 9
}

func ResponseHighestSupportedFlexVersion() int16 {
	return 32767
}

func ResponseDefault() Response {
	var d Response
	d.SetDefault()

	return d
}
