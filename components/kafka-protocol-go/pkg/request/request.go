//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package request

import (
	"strconv"
	"strings"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/messages/request"
)

type Request struct {
	data       protocol.MessageData
	headerData request.Header
}

func (r *Request) HeaderData() *request.Header {
	return &r.headerData
}
func (r *Request) Data() protocol.MessageData {
	return r.data
}

func (r *Request) String() string {
	var s strings.Builder

	s.WriteString("{\"headerData\": ")
	s.WriteString(r.headerData.String())
	s.WriteString(", \"data\": ")
	s.WriteString(r.data.String())
	s.WriteRune('}')

	return s.String()
}

// Serialize serializes this Kafka request according to the given Kafka protocol version
func (r *Request) Serialize(apiVersion int16) ([]byte, error) {
	apiKey := r.headerData.RequestApiKey()

	hrdVersion := headerVersion(apiKey, apiVersion)
	if hrdVersion == -1 {
		return nil, errors.New(strings.Join([]string{"unsupported api key", strconv.Itoa(int(apiKey))}, " "))
	}
	if hrdVersion < request.HeaderLowestSupportedVersion() ||
		hrdVersion > request.HeaderHighestSupportedVersion() {
		return nil, errors.New(strings.Join([]string{"unsupported request header version", strconv.Itoa(int(hrdVersion))}, " "))
	}

	size, err := r.SizeInBytes(apiVersion)
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't compute request size")
	}
	buf := make([]byte, 0, size)
	w := typesbytes.NewSliceWriter(buf)
	err = r.headerData.Write(&w, hrdVersion)
	if err != nil {
		return nil, errors.WrapIf(err, strings.Join([]string{"couldn't serialize request header for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion)), ", header version", strconv.Itoa(int(hrdVersion))}, " "))
	}

	err = r.data.Write(&w, apiVersion)
	if err != nil {
		return nil, errors.WrapIf(err, strings.Join([]string{"couldn't serialize request message for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion))}, " "))
	}

	return w.Bytes(), nil
}

// SizeInBytes returns the max size of this Kafka request in bytes when it's serialized
func (r *Request) SizeInBytes(apiVersion int16) (int, error) {
	apiKey := r.headerData.RequestApiKey()

	hrdVersion := headerVersion(apiKey, apiVersion)
	if hrdVersion == -1 {
		return 0, errors.New(strings.Join([]string{"unsupported api key", strconv.Itoa(int(apiKey))}, " "))
	}
	if hrdVersion < request.HeaderLowestSupportedVersion() ||
		hrdVersion > request.HeaderHighestSupportedVersion() {
		return 0, errors.New(strings.Join([]string{"unsupported request header version", strconv.Itoa(int(hrdVersion))}, " "))
	}
	hdrSize, err := r.headerData.SizeInBytes(hrdVersion)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute request header size")
	}

	dataSize, err := r.data.SizeInBytes(apiVersion)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute request data size")
	}

	return hdrSize + dataSize, nil
}

// Release releases the dynamically allocated fields of this object by returning then to object pools
func (r *Request) Release() {
	if r == nil {
		return
	}

	r.headerData.Release()
	if r.data != nil {
		r.data.Release()
	}
}

// Parse deserializes this Kafka request
func Parse(msg []byte) (Request, error) {
	// We derive the header version from the request api version, so we read that first.
	// The request api version is part of the request header data, so we create a copy of
	// the original buffer to keep the read position
	if len(msg) < 4 {
		return Request{}, errors.New(strings.Join([]string{"expected 4 bytes to hold api key and api version but got", strconv.Itoa(len(msg)), "bytes"}, " "))
	}

	r := pools.GetBytesReader(msg[:4])
	defer pools.ReleaseBytesReader(r)

	var apiKey int16
	err := types.ReadInt16(r, &apiKey)
	if err != nil {
		return Request{}, errors.WrapIf(err, "couldn't read api key")
	}

	var apiVersion int16
	err = types.ReadInt16(r, &apiVersion)
	if err != nil {
		return Request{}, errors.WrapIf(err, "couldn't read api version")
	}

	hrdVersion := headerVersion(apiKey, apiVersion)
	if hrdVersion == -1 {
		return Request{}, errors.New(strings.Join([]string{"unsupported api key", strconv.Itoa(int(apiKey))}, " "))
	}

	if hrdVersion < request.HeaderLowestSupportedVersion() ||
		hrdVersion > request.HeaderHighestSupportedVersion() {
		return Request{}, errors.New(strings.Join([]string{"unsupported request header version", strconv.Itoa(int(hrdVersion))}, " "))
	}

	r.Reset(msg)

	var hdr request.Header
	err = hdr.Read(r, hrdVersion)
	if err != nil {
		hdr.Release()
		return Request{}, errors.WrapIf(err, strings.Join([]string{"couldn't parse request header for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion)), ", header version", strconv.Itoa(int(hrdVersion))}, " "))
	}

	// create typed request data struct
	req, err := messages.NewRequest(apiKey)
	if err != nil {
		hdr.Release()
		return Request{}, errors.WrapIf(err, strings.Join([]string{"couldn't create new request object for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion))}, " "))
	}

	err = req.Read(r, apiVersion)
	if err != nil {
		hdr.Release()
		return Request{}, errors.WrapIf(err, strings.Join([]string{"couldn't parse request message for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion))}, " "))
	}

	return Request{
		headerData: hdr,
		data:       req,
	}, nil
}
