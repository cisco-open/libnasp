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
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/describedelegationtoken/response/describeddelegationtoken"
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/fields"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/varint"
)

var describedDelegationTokenPrincipalType = fields.Context{
	SpecName:                    "PrincipalType",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenPrincipalName = fields.Context{
	SpecName:                    "PrincipalName",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenTokenRequesterPrincipalType = fields.Context{
	SpecName:                    "TokenRequesterPrincipalType",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenTokenRequesterPrincipalName = fields.Context{
	SpecName:                    "TokenRequesterPrincipalName",
	LowestSupportedVersion:      3,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenIssueTimestamp = fields.Context{
	SpecName:                    "IssueTimestamp",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenExpiryTimestamp = fields.Context{
	SpecName:                    "ExpiryTimestamp",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenMaxTimestamp = fields.Context{
	SpecName:                    "MaxTimestamp",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenTokenId = fields.Context{
	SpecName:                    "TokenId",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenHmac = fields.Context{
	SpecName:                    "Hmac",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}
var describedDelegationTokenRenewers = fields.Context{
	SpecName:                    "Renewers",
	LowestSupportedVersion:      0,
	HighestSupportedVersion:     32767,
	LowestSupportedFlexVersion:  2,
	HighestSupportedFlexVersion: 32767,
}

type DescribedDelegationToken struct {
	hmac                        []byte
	renewers                    []describeddelegationtoken.DescribedDelegationTokenRenewer
	unknownTaggedFields         []fields.RawTaggedField
	principalType               fields.NullableString
	principalName               fields.NullableString
	tokenRequesterPrincipalType fields.NullableString
	tokenRequesterPrincipalName fields.NullableString
	tokenId                     fields.NullableString
	issueTimestamp              int64
	expiryTimestamp             int64
	maxTimestamp                int64
	isNil                       bool
}

func (o *DescribedDelegationToken) PrincipalType() fields.NullableString {
	return o.principalType
}

func (o *DescribedDelegationToken) SetPrincipalType(val fields.NullableString) {
	o.isNil = false
	o.principalType = val
}

func (o *DescribedDelegationToken) PrincipalName() fields.NullableString {
	return o.principalName
}

func (o *DescribedDelegationToken) SetPrincipalName(val fields.NullableString) {
	o.isNil = false
	o.principalName = val
}

func (o *DescribedDelegationToken) TokenRequesterPrincipalType() fields.NullableString {
	return o.tokenRequesterPrincipalType
}

func (o *DescribedDelegationToken) SetTokenRequesterPrincipalType(val fields.NullableString) {
	o.isNil = false
	o.tokenRequesterPrincipalType = val
}

func (o *DescribedDelegationToken) TokenRequesterPrincipalName() fields.NullableString {
	return o.tokenRequesterPrincipalName
}

func (o *DescribedDelegationToken) SetTokenRequesterPrincipalName(val fields.NullableString) {
	o.isNil = false
	o.tokenRequesterPrincipalName = val
}

func (o *DescribedDelegationToken) IssueTimestamp() int64 {
	return o.issueTimestamp
}

func (o *DescribedDelegationToken) SetIssueTimestamp(val int64) {
	o.isNil = false
	o.issueTimestamp = val
}

func (o *DescribedDelegationToken) ExpiryTimestamp() int64 {
	return o.expiryTimestamp
}

func (o *DescribedDelegationToken) SetExpiryTimestamp(val int64) {
	o.isNil = false
	o.expiryTimestamp = val
}

func (o *DescribedDelegationToken) MaxTimestamp() int64 {
	return o.maxTimestamp
}

func (o *DescribedDelegationToken) SetMaxTimestamp(val int64) {
	o.isNil = false
	o.maxTimestamp = val
}

func (o *DescribedDelegationToken) TokenId() fields.NullableString {
	return o.tokenId
}

func (o *DescribedDelegationToken) SetTokenId(val fields.NullableString) {
	o.isNil = false
	o.tokenId = val
}

func (o *DescribedDelegationToken) Hmac() []byte {
	return o.hmac
}

func (o *DescribedDelegationToken) SetHmac(val []byte) {
	o.isNil = false
	o.hmac = val
}

func (o *DescribedDelegationToken) Renewers() []describeddelegationtoken.DescribedDelegationTokenRenewer {
	return o.renewers
}

func (o *DescribedDelegationToken) SetRenewers(val []describeddelegationtoken.DescribedDelegationTokenRenewer) {
	o.isNil = false
	o.renewers = val
}

func (o *DescribedDelegationToken) UnknownTaggedFields() []fields.RawTaggedField {
	return o.unknownTaggedFields
}

func (o *DescribedDelegationToken) SetUnknownTaggedFields(val []fields.RawTaggedField) {
	o.unknownTaggedFields = val
}

func (o *DescribedDelegationToken) Read(buf *bytes.Reader, version int16) error {
	o.SetDefault()

	principalTypeField := fields.String{Context: describedDelegationTokenPrincipalType}
	if err := principalTypeField.Read(buf, version, &o.principalType); err != nil {
		return errors.WrapIf(err, "couldn't set \"principalType\" field")
	}

	principalNameField := fields.String{Context: describedDelegationTokenPrincipalName}
	if err := principalNameField.Read(buf, version, &o.principalName); err != nil {
		return errors.WrapIf(err, "couldn't set \"principalName\" field")
	}

	tokenRequesterPrincipalTypeField := fields.String{Context: describedDelegationTokenTokenRequesterPrincipalType}
	if err := tokenRequesterPrincipalTypeField.Read(buf, version, &o.tokenRequesterPrincipalType); err != nil {
		return errors.WrapIf(err, "couldn't set \"tokenRequesterPrincipalType\" field")
	}

	tokenRequesterPrincipalNameField := fields.String{Context: describedDelegationTokenTokenRequesterPrincipalName}
	if err := tokenRequesterPrincipalNameField.Read(buf, version, &o.tokenRequesterPrincipalName); err != nil {
		return errors.WrapIf(err, "couldn't set \"tokenRequesterPrincipalName\" field")
	}

	issueTimestampField := fields.Int64{Context: describedDelegationTokenIssueTimestamp}
	if err := issueTimestampField.Read(buf, version, &o.issueTimestamp); err != nil {
		return errors.WrapIf(err, "couldn't set \"issueTimestamp\" field")
	}

	expiryTimestampField := fields.Int64{Context: describedDelegationTokenExpiryTimestamp}
	if err := expiryTimestampField.Read(buf, version, &o.expiryTimestamp); err != nil {
		return errors.WrapIf(err, "couldn't set \"expiryTimestamp\" field")
	}

	maxTimestampField := fields.Int64{Context: describedDelegationTokenMaxTimestamp}
	if err := maxTimestampField.Read(buf, version, &o.maxTimestamp); err != nil {
		return errors.WrapIf(err, "couldn't set \"maxTimestamp\" field")
	}

	tokenIdField := fields.String{Context: describedDelegationTokenTokenId}
	if err := tokenIdField.Read(buf, version, &o.tokenId); err != nil {
		return errors.WrapIf(err, "couldn't set \"tokenId\" field")
	}

	hmacField := fields.Bytes{Context: describedDelegationTokenHmac}
	if err := hmacField.Read(buf, version, &o.hmac); err != nil {
		return errors.WrapIf(err, "couldn't set \"hmac\" field")
	}

	renewersField := fields.ArrayOfStruct[describeddelegationtoken.DescribedDelegationTokenRenewer, *describeddelegationtoken.DescribedDelegationTokenRenewer]{Context: describedDelegationTokenRenewers}
	renewers, err := renewersField.Read(buf, version)
	if err != nil {
		return errors.WrapIf(err, "couldn't set \"renewers\" field")
	}
	o.renewers = renewers

	// process tagged fields

	if version < DescribedDelegationTokenLowestSupportedFlexVersion() || version > DescribedDelegationTokenHighestSupportedFlexVersion() {
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

func (o *DescribedDelegationToken) Write(buf *typesbytes.SliceWriter, version int16) error {
	if o.IsNil() {
		return nil
	}
	if err := o.validateNonIgnorableFields(version); err != nil {
		return err
	}

	principalTypeField := fields.String{Context: describedDelegationTokenPrincipalType}
	if err := principalTypeField.Write(buf, version, o.principalType); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"principalType\" field")
	}
	principalNameField := fields.String{Context: describedDelegationTokenPrincipalName}
	if err := principalNameField.Write(buf, version, o.principalName); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"principalName\" field")
	}
	tokenRequesterPrincipalTypeField := fields.String{Context: describedDelegationTokenTokenRequesterPrincipalType}
	if err := tokenRequesterPrincipalTypeField.Write(buf, version, o.tokenRequesterPrincipalType); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"tokenRequesterPrincipalType\" field")
	}
	tokenRequesterPrincipalNameField := fields.String{Context: describedDelegationTokenTokenRequesterPrincipalName}
	if err := tokenRequesterPrincipalNameField.Write(buf, version, o.tokenRequesterPrincipalName); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"tokenRequesterPrincipalName\" field")
	}
	issueTimestampField := fields.Int64{Context: describedDelegationTokenIssueTimestamp}
	if err := issueTimestampField.Write(buf, version, o.issueTimestamp); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"issueTimestamp\" field")
	}
	expiryTimestampField := fields.Int64{Context: describedDelegationTokenExpiryTimestamp}
	if err := expiryTimestampField.Write(buf, version, o.expiryTimestamp); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"expiryTimestamp\" field")
	}
	maxTimestampField := fields.Int64{Context: describedDelegationTokenMaxTimestamp}
	if err := maxTimestampField.Write(buf, version, o.maxTimestamp); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"maxTimestamp\" field")
	}
	tokenIdField := fields.String{Context: describedDelegationTokenTokenId}
	if err := tokenIdField.Write(buf, version, o.tokenId); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"tokenId\" field")
	}
	hmacField := fields.Bytes{Context: describedDelegationTokenHmac}
	if err := hmacField.Write(buf, version, o.hmac); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"hmac\" field")
	}

	renewersField := fields.ArrayOfStruct[describeddelegationtoken.DescribedDelegationTokenRenewer, *describeddelegationtoken.DescribedDelegationTokenRenewer]{Context: describedDelegationTokenRenewers}
	if err := renewersField.Write(buf, version, o.renewers); err != nil {
		return errors.WrapIf(err, "couldn't serialize \"renewers\" field")
	}

	// serialize tagged fields
	numTaggedFields := o.getTaggedFieldsCount(version)
	if version < DescribedDelegationTokenLowestSupportedFlexVersion() || version > DescribedDelegationTokenHighestSupportedFlexVersion() {
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

func (o *DescribedDelegationToken) String() string {
	s, err := o.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (o *DescribedDelegationToken) MarshalJSON() ([]byte, error) {
	if o == nil || o.IsNil() {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, 11)
	if b, err := fields.MarshalPrimitiveTypeJSON(o.principalType); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"principalType\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.principalName); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"principalName\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.tokenRequesterPrincipalType); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"tokenRequesterPrincipalType\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.tokenRequesterPrincipalName); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"tokenRequesterPrincipalName\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.issueTimestamp); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"issueTimestamp\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.expiryTimestamp); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"expiryTimestamp\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.maxTimestamp); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"maxTimestamp\""), b}, []byte(": ")))
	}
	if b, err := fields.MarshalPrimitiveTypeJSON(o.tokenId); err != nil {
		return nil, err
	} else {
		s = append(s, bytes.Join([][]byte{[]byte("\"tokenId\""), b}, []byte(": ")))
	}
	if b, err := fields.BytesMarshalJSON("hmac", o.hmac); err != nil {
		return nil, err
	} else {
		s = append(s, b)
	}
	if b, err := fields.ArrayOfStructMarshalJSON("renewers", o.renewers); err != nil {
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

func (o *DescribedDelegationToken) IsNil() bool {
	return o.isNil
}

func (o *DescribedDelegationToken) Clear() {
	o.Release()
	o.isNil = true

	o.renewers = nil
	o.unknownTaggedFields = nil
}

func (o *DescribedDelegationToken) SetDefault() {
	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil
	o.principalType.SetValue("")
	o.principalName.SetValue("")
	o.tokenRequesterPrincipalType.SetValue("")
	o.tokenRequesterPrincipalName.SetValue("")
	o.issueTimestamp = 0
	o.expiryTimestamp = 0
	o.maxTimestamp = 0
	o.tokenId.SetValue("")
	if o.hmac != nil {
		pools.ReleaseByteSlice(o.hmac)
	}
	o.hmac = nil
	for i := range o.renewers {
		o.renewers[i].Release()
	}
	o.renewers = nil

	o.isNil = false
}

func (o *DescribedDelegationToken) Equal(that *DescribedDelegationToken) bool {
	if !fields.RawTaggedFieldsEqual(o.unknownTaggedFields, that.unknownTaggedFields) {
		return false
	}

	if !o.principalType.Equal(&that.principalType) {
		return false
	}
	if !o.principalName.Equal(&that.principalName) {
		return false
	}
	if !o.tokenRequesterPrincipalType.Equal(&that.tokenRequesterPrincipalType) {
		return false
	}
	if !o.tokenRequesterPrincipalName.Equal(&that.tokenRequesterPrincipalName) {
		return false
	}
	if o.issueTimestamp != that.issueTimestamp {
		return false
	}
	if o.expiryTimestamp != that.expiryTimestamp {
		return false
	}
	if o.maxTimestamp != that.maxTimestamp {
		return false
	}
	if !o.tokenId.Equal(&that.tokenId) {
		return false
	}
	if !bytes.Equal(o.hmac, that.hmac) {
		return false
	}
	if len(o.renewers) != len(that.renewers) {
		return false
	}
	for i := range o.renewers {
		if !o.renewers[i].Equal(&that.renewers[i]) {
			return false
		}
	}

	return true
}

// SizeInBytes returns the size of this data structure in bytes when it's serialized
func (o *DescribedDelegationToken) SizeInBytes(version int16) (int, error) {
	if o.IsNil() {
		return 0, nil
	}

	if err := o.validateNonIgnorableFields(version); err != nil {
		return 0, err
	}

	size := 0
	fieldSize := 0
	var err error

	principalTypeField := fields.String{Context: describedDelegationTokenPrincipalType}
	fieldSize, err = principalTypeField.SizeInBytes(version, o.principalType)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"principalType\" field")
	}
	size += fieldSize

	principalNameField := fields.String{Context: describedDelegationTokenPrincipalName}
	fieldSize, err = principalNameField.SizeInBytes(version, o.principalName)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"principalName\" field")
	}
	size += fieldSize

	tokenRequesterPrincipalTypeField := fields.String{Context: describedDelegationTokenTokenRequesterPrincipalType}
	fieldSize, err = tokenRequesterPrincipalTypeField.SizeInBytes(version, o.tokenRequesterPrincipalType)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"tokenRequesterPrincipalType\" field")
	}
	size += fieldSize

	tokenRequesterPrincipalNameField := fields.String{Context: describedDelegationTokenTokenRequesterPrincipalName}
	fieldSize, err = tokenRequesterPrincipalNameField.SizeInBytes(version, o.tokenRequesterPrincipalName)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"tokenRequesterPrincipalName\" field")
	}
	size += fieldSize

	issueTimestampField := fields.Int64{Context: describedDelegationTokenIssueTimestamp}
	fieldSize, err = issueTimestampField.SizeInBytes(version, o.issueTimestamp)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"issueTimestamp\" field")
	}
	size += fieldSize

	expiryTimestampField := fields.Int64{Context: describedDelegationTokenExpiryTimestamp}
	fieldSize, err = expiryTimestampField.SizeInBytes(version, o.expiryTimestamp)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"expiryTimestamp\" field")
	}
	size += fieldSize

	maxTimestampField := fields.Int64{Context: describedDelegationTokenMaxTimestamp}
	fieldSize, err = maxTimestampField.SizeInBytes(version, o.maxTimestamp)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"maxTimestamp\" field")
	}
	size += fieldSize

	tokenIdField := fields.String{Context: describedDelegationTokenTokenId}
	fieldSize, err = tokenIdField.SizeInBytes(version, o.tokenId)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"tokenId\" field")
	}
	size += fieldSize

	hmacField := fields.Bytes{Context: describedDelegationTokenHmac}
	fieldSize, err = hmacField.SizeInBytes(version, o.hmac)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"hmac\" field")
	}
	size += fieldSize

	renewersField := fields.ArrayOfStruct[describeddelegationtoken.DescribedDelegationTokenRenewer, *describeddelegationtoken.DescribedDelegationTokenRenewer]{Context: describedDelegationTokenRenewers}
	fieldSize, err = renewersField.SizeInBytes(version, o.renewers)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute size of \"renewers\" field")
	}
	size += fieldSize

	// tagged fields
	numTaggedFields := int64(o.getTaggedFieldsCount(version))
	if numTaggedFields > 0xffffffff {
		return 0, errors.New(strings.Join([]string{"invalid tagged fields count:", strconv.Itoa(int(numTaggedFields))}, " "))
	}
	if version < DescribedDelegationTokenLowestSupportedFlexVersion() || version > DescribedDelegationTokenHighestSupportedFlexVersion() {
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
func (o *DescribedDelegationToken) Release() {
	if o.IsNil() {
		return
	}

	for i := range o.unknownTaggedFields {
		o.unknownTaggedFields[i].Release()
	}
	o.unknownTaggedFields = nil

	o.principalType.Release()
	o.principalName.Release()
	o.tokenRequesterPrincipalType.Release()
	o.tokenRequesterPrincipalName.Release()
	o.tokenId.Release()
	if o.hmac != nil {
		pools.ReleaseByteSlice(o.hmac)
	}
	for i := range o.renewers {
		o.renewers[i].Release()
	}
	o.renewers = nil
}

func (o *DescribedDelegationToken) getTaggedFieldsCount(version int16) int {
	numTaggedFields := len(o.unknownTaggedFields)

	return numTaggedFields
}

// validateNonIgnorableFields throws an error if any non-ignorable field not supported by current version is set to
// non-default value
func (o *DescribedDelegationToken) validateNonIgnorableFields(version int16) error {
	if !describedDelegationTokenTokenRequesterPrincipalType.IsSupportedVersion(version) {
		if o.tokenRequesterPrincipalType.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"tokenRequesterPrincipalType\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	if !describedDelegationTokenTokenRequesterPrincipalName.IsSupportedVersion(version) {
		if o.tokenRequesterPrincipalName.Bytes() != nil {
			return errors.New(strings.Join([]string{"attempted to write non-default \"tokenRequesterPrincipalName\" at version", strconv.Itoa(int(version))}, " "))
		}
	}
	return nil
}

func DescribedDelegationTokenLowestSupportedVersion() int16 {
	return 0
}

func DescribedDelegationTokenHighestSupportedVersion() int16 {
	return 32767
}

func DescribedDelegationTokenLowestSupportedFlexVersion() int16 {
	return 2
}

func DescribedDelegationTokenHighestSupportedFlexVersion() int16 {
	return 32767
}

func DescribedDelegationTokenDefault() DescribedDelegationToken {
	var d DescribedDelegationToken
	d.SetDefault()

	return d
}
