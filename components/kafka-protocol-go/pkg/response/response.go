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

package response

import (
	"io"
	"strconv"
	"strings"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/pools"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/request"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/messages/response"

	"emperror.dev/errors"
)

type Response struct {
	data       protocol.MessageData
	headerData response.Header
}

func (r *Response) HeaderData() *response.Header {
	return &r.headerData
}
func (r *Response) Data() protocol.MessageData {
	return r.data
}
func (r *Response) String() string {
	var s strings.Builder

	s.WriteString("{\"headerData\": ")
	s.WriteString(r.headerData.String())
	s.WriteString(", \"data\": ")
	s.WriteString(r.data.String())
	s.WriteRune('}')

	return s.String()
}

// Serialize serializes this Kafka response according to the given Kafka protocol version
func (r *Response) Serialize(apiVersion int16) ([]byte, error) {
	apiKey := r.data.ApiKey()
	hdrVersion := headerVersion(apiKey, apiVersion)
	if hdrVersion == -1 {
		return nil, errors.New(strings.Join([]string{"unsupported api key", strconv.Itoa(int(apiKey))}, " "))
	}

	if hdrVersion < response.HeaderLowestSupportedVersion() ||
		hdrVersion > response.HeaderHighestSupportedVersion() {
		return nil, errors.New(strings.Join([]string{"unsupported response header version", strconv.Itoa(int(hdrVersion))}, " "))
	}

	size, err := r.SizeInBytes(apiVersion)
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't compute response size")
	}
	buf := make([]byte, 0, size)
	w := typesbytes.NewSliceWriter(buf)
	err = r.headerData.Write(&w, hdrVersion)
	if err != nil {
		return nil, errors.WrapIf(err, strings.Join([]string{"couldn't serialize response header for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion)), ", header version", strconv.Itoa(int(hdrVersion))}, " "))
	}

	err = r.data.Write(&w, apiVersion)
	if err != nil {
		return nil, errors.WrapIf(err, strings.Join([]string{"couldn't serialize response message for api key", strconv.Itoa(int(apiKey)), ", api version", strconv.Itoa(int(apiVersion))}, " "))
	}

	return w.Bytes(), nil
}

// SizeInBytes returns the max size of this Kafka response in bytes when it's serialized
func (r *Response) SizeInBytes(apiVersion int16) (int, error) {
	apiKey := r.data.ApiKey()

	hrdVersion := headerVersion(apiKey, apiVersion)
	if hrdVersion == -1 {
		return 0, errors.New(strings.Join([]string{"unsupported api key", strconv.Itoa(int(apiKey))}, " "))
	}
	if hrdVersion < request.HeaderLowestSupportedVersion() ||
		hrdVersion > request.HeaderHighestSupportedVersion() {
		return 0, errors.New(strings.Join([]string{"unsupported response header version", strconv.Itoa(int(hrdVersion))}, " "))
	}
	hdrSize, err := r.headerData.SizeInBytes(hrdVersion)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute response header size")
	}

	dataSize, err := r.data.SizeInBytes(apiVersion)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't compute response data size")
	}

	return hdrSize + dataSize, nil
}

// Release releases the dynamically allocated fields of this object by returning then to object pools
func (r *Response) Release() {
	if r == nil {
		return
	}

	r.headerData.Release()
	if r.data != nil {
		r.data.Release()
	}
}

// Parse deserializes this Kafka response
func Parse(msg []byte, requestApiKey, requestApiVersion int16, requestCorrelationId int32) (Response, error) {
	hdrVersion := headerVersion(requestApiKey, requestApiVersion)
	if hdrVersion == -1 {
		return Response{}, errors.New(strings.Join([]string{"unsupported api key", strconv.Itoa(int(requestApiKey))}, " "))
	}

	r := pools.GetBytesReader(msg)
	defer pools.ReleaseBytesReader(r)

	var hdr response.Header
	err := hdr.Read(r, hdrVersion)
	if err != nil {
		hdr.Release()
		return Response{}, errors.WrapIf(err, strings.Join([]string{"couldn't parse response header for api key", strconv.Itoa(int(requestApiKey)), ", api version", strconv.Itoa(int(requestApiVersion)), ", header version", strconv.Itoa(int(hdrVersion))}, " "))
	}

	if hdr.CorrelationId() != requestCorrelationId {
		hdr.Release()
		return Response{}, errors.New(strings.Join([]string{"response correlation id", strconv.Itoa(int(hdr.CorrelationId())), "doesn't match request correlation id", strconv.Itoa(int(requestCorrelationId))}, " "))
	}

	bufCurrentReadingPos := r.Size() - int64(r.Len())

	// create typed request data struct
	resp, err := messages.NewResponse(requestApiKey)
	if err != nil {
		hdr.Release()
		return Response{}, errors.WrapIf(err, strings.Join([]string{"couldn't create new response object for api key", strconv.Itoa(int(requestApiKey)), ", api version", strconv.Itoa(int(requestApiVersion))}, " "))
	}

	err = resp.Read(r, requestApiVersion)
	if err != nil {
		// ApiVersions needs special handling. If the broker doesn't support the api version specified in the ApiVersions
		// request than the broker will respond with an ApiVersions response version 0. Thus, if we can not parse
		// the ApiVersions response with request's api version than fall back to version 0
		if requestApiVersion == 18 && requestApiVersion != 0 {
			if _, err = r.Seek(bufCurrentReadingPos, io.SeekStart); err != nil {
				hdr.Release()
				return Response{}, errors.WrapIf(err, "couldn't reset buffer position to retry parsing ApiVersions response with api version 0")
			}

			err = resp.Read(r, 0)
		}

		if err != nil {
			hdr.Release()
			return Response{}, errors.WrapIf(err, strings.Join([]string{"couldn't parse response message for api key", strconv.Itoa(int(requestApiKey)), ", api version", strconv.Itoa(int(requestApiVersion)), ", request correlation id", strconv.Itoa(int(requestCorrelationId))}, " "))
		}
	}

	return Response{
		headerData: hdr,
		data:       resp,
	}, nil
}
