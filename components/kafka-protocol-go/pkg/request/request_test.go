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

package request_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"emperror.dev/errors"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/assets/test"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/request"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/response"
)

func TestParse(t *testing.T) {
	t.Parallel()

	msg, err := test.KafkaProtocolPayloads.ReadFile("protocol_payload/api_versions_req.dmp")
	if err != nil {
		t.Error(err)
		return
	}

	r, err := request.Parse(msg)
	if err != nil {
		t.Error(err)
		return
	}
	r.Release()

	msg, err = test.KafkaProtocolPayloads.ReadFile("protocol_payload/metadata_req.dmp")
	if err != nil {
		t.Error(err)
		return
	}

	r, err = request.Parse(msg)
	if err != nil {
		t.Error(err)
		return
	}
	r.Release()
}

//nolint:funlen,gocognit
func TestZippedRequestResponseParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName      string
		requestsFile  string
		responsesFile string
	}{
		{
			testName:      "test-1",
			requestsFile:  "protocol_payload/requests-0.dmp",
			responsesFile: "protocol_payload/responses-0.dmp",
		},
		{
			testName:      "test-2",
			requestsFile:  "protocol_payload/requests-1.dmp",
			responsesFile: "protocol_payload/responses-1.dmp",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			requestMessages, err := test.KafkaProtocolPayloads.ReadFile(tc.requestsFile)
			if err != nil {
				t.Error(err)
				return
			}

			responseMessages, err := test.KafkaProtocolPayloads.ReadFile(tc.responsesFile)
			if err != nil {
				t.Error(err)
				return
			}

			var length int32

			r := bytes.NewReader(requestMessages)
			p := bytes.NewReader(responseMessages)

			for {
				if r.Len() < 4 {
					break
				}
				err = binary.Read(r, binary.BigEndian, &length)
				if err != nil {
					t.Error(err)
					return
				}
				if r.Len() < int(length) {
					break
				}

				requestMessage := make([]byte, length)
				_, err = r.Read(requestMessage)
				if err != nil {
					t.Error(err)
					return
				}

				// parse request
				req, err := request.Parse(requestMessage)
				if err != nil {
					t.Error(err)
					return
				}

				// serialize request
				serializedMessage, err := req.Serialize(req.HeaderData().RequestApiVersion())
				req.Release()
				if err != nil {
					t.Error(err)
					return
				}

				req, err = request.Parse(serializedMessage)
				if err != nil {
					t.Error(err)
					return
				}
				req.Release()

				if r.Len() < 4 {
					break
				}
				err = binary.Read(p, binary.BigEndian, &length)
				if err != nil {
					t.Error(err)
					return
				}
				if r.Len() < int(length) {
					break
				}

				responseMessage := make([]byte, length)
				_, err = p.Read(responseMessage)
				if err != nil {
					t.Error(err)
					return
				}

				// parse response
				resp, err := response.Parse(
					responseMessage,
					req.HeaderData().RequestApiKey(),
					req.HeaderData().RequestApiVersion(),
					req.HeaderData().CorrelationId())
				if err != nil {
					t.Error(err)
					return
				}

				// serialize request
				serializedMessage, err = resp.Serialize(req.HeaderData().RequestApiVersion())
				resp.Release()
				if err != nil {
					t.Error(err)
					return
				}

				resp, err = response.Parse(
					serializedMessage,
					req.HeaderData().RequestApiKey(),
					req.HeaderData().RequestApiVersion(),
					req.HeaderData().CorrelationId())
				if err != nil {
					t.Error(err)
					return
				}
				resp.Release()
			}
		})
	}
}

type testCase struct {
	TestName   string
	RawMessage []byte
	ApiKey     int16
	ApiVersion int16
	MsgType    byte
}

func setup() []testCase {
	testProtocolPayloadRootDir := "protocol_payload"
	dirs, err := test.KafkaProtocolPayloads.ReadDir(testProtocolPayloadRootDir)
	if err != nil {
		panic(err)
	}

	var testCases []testCase

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		i := strings.LastIndex(d.Name(), "-")
		if i == -1 {
			continue
		}

		apiKey, err := strconv.Atoi(d.Name()[i+1:])
		if err != nil {
			panic(err)
		}
		apiKeyName := d.Name()[:i]

		apiPayLoadRootDir := path.Join(testProtocolPayloadRootDir, d.Name())
		apiPayloadVersionDirs, err := test.KafkaProtocolPayloads.ReadDir(apiPayLoadRootDir)
		if err != nil {
			panic(err)
		}

		for _, versionDir := range apiPayloadVersionDirs {
			if !versionDir.IsDir() {
				panic(fmt.Sprintf("%q is not a directory", versionDir.Name()))
			}

			apiVer, err := strconv.Atoi(strings.TrimPrefix(versionDir.Name(), "v"))
			if err != nil {
				panic(err)
			}

			currentDir := path.Join(apiPayLoadRootDir, versionDir.Name())
			protocolPayloadFiles, err := test.KafkaProtocolPayloads.ReadDir(currentDir)
			if err != nil {
				panic(err)
			}

			for _, protocolPayloadFile := range protocolPayloadFiles {
				if protocolPayloadFile.IsDir() {
					panic(fmt.Sprintf("%q is not a file", protocolPayloadFile.Name()))
				}

				filePath := path.Join(currentDir, protocolPayloadFile.Name())
				testName := ""
				var msgType byte

				if strings.HasSuffix(filePath, "produce-0/v2/request.dmp") {
					continue
				}

				switch protocolPayloadFile.Name() {
				case "request.dmp":
					testName = fmt.Sprintf("%s request version %d", apiKeyName, apiVer)
					msgType = 0
				case "response.dmp":
					testName = fmt.Sprintf("%s response version %d", apiKeyName, apiVer)
					msgType = 1
				case "response_err.dmp":
					testName = fmt.Sprintf("%s error response version %d", apiKeyName, apiVer)
					msgType = 1
				default:
					panic(fmt.Sprintf("unexpected %q protocol test payload file", versionDir.Name()))
				}

				rawMsg, err := test.KafkaProtocolPayloads.ReadFile(filePath)
				if err != nil {
					panic(err)
				}

				testCases = append(testCases, testCase{
					TestName: testName,
					MsgType:  msgType,
					//nolint:gosec
					ApiKey: int16(apiKey),
					//nolint:gosec
					ApiVersion: int16(apiVer),
					RawMessage: rawMsg,
				})
			}
		}
	}

	return testCases
}

func TestRequestResponseParse(t *testing.T) {
	t.Parallel()

	testCases := setup()

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			switch tc.MsgType {
			case 0:
				req, err := request.Parse(tc.RawMessage)
				//nolint:goconst
				if err != nil &&
					errors.Cause(err).Error() != "message sets version 0 (pre Kafka 0.10) are not supported" &&
					errors.Cause(err).Error() != "message sets version 1 (Kafka 0.10) are not supported" {
					t.Error(err)
					return
				}
				req.Release()
			case 1:
				resp, err := response.Parse(tc.RawMessage, tc.ApiKey, tc.ApiVersion, 2)
				if err != nil {
					t.Error(err)
					return
				}
				resp.Release()
			default:
				t.Errorf("unexpected message type %d", tc.MsgType)
			}
		})
	}
}

func TestMessageSizeInBytes(t *testing.T) {
	t.Parallel()

	testCases := setup()

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			switch tc.MsgType {
			case 0:
				req, err := request.Parse(tc.RawMessage)
				if err != nil &&
					errors.Cause(err).Error() != "message sets version 0 (pre Kafka 0.10) are not supported" &&
					errors.Cause(err).Error() != "message sets version 1 (Kafka 0.10) are not supported" {
					t.Error(err)
					return
				}
				actual, err := req.SizeInBytes(tc.ApiVersion)
				req.Release()
				if err != nil {
					t.Error(err)
					return
				}
				expected := len(tc.RawMessage)
				if actual != expected {
					t.Errorf("expected request size: %d, got: %d", expected, actual)
					return
				}
			case 1:
				resp, err := response.Parse(tc.RawMessage, tc.ApiKey, tc.ApiVersion, 2)
				if err != nil {
					t.Error(err)
					return
				}
				actual, err := resp.SizeInBytes(tc.ApiVersion)
				resp.Release()
				if err != nil {
					t.Error(err)
					return
				}
				expected := len(tc.RawMessage)
				if actual != expected {
					t.Errorf("expected request size: %d, got: %d", expected, actual)
					return
				}
			default:
				t.Errorf("unexpected message type %d", tc.MsgType)
			}
		})
	}
}

func TestRequestResponseSerialize(t *testing.T) {
	t.Parallel()

	testCases := setup()

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			switch tc.MsgType {
			case 0:
				req, err := request.Parse(tc.RawMessage)
				if err != nil &&
					errors.Cause(err).Error() != "message sets version 0 (pre Kafka 0.10) are not supported" &&
					errors.Cause(err).Error() != "message sets version 1 (Kafka 0.10) are not supported" {
					t.Error(err)
					return
				}

				serializedMsg, err := req.Serialize(tc.ApiVersion)
				req.Release()
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, tc.RawMessage, serializedMsg, "serialized request message doesn't match original message")
			case 1:
				resp, err := response.Parse(tc.RawMessage, tc.ApiKey, tc.ApiVersion, 2)
				if err != nil {
					t.Error(err)
					return
				}

				serializedMsg, err := resp.Serialize(tc.ApiVersion)
				resp.Release()
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, tc.RawMessage, serializedMsg, "serialized response message doesn't match original message")
			default:
				t.Errorf("unexpected message type %d", tc.MsgType)
			}
		})
	}
}

var (
	g_req  request.Request
	g_resp response.Response
)

func BenchmarkRequestResponseParse(b *testing.B) {
	testCases := setup()

	// https://teivah.medium.com/how-to-write-accurate-benchmarks-in-go-4266d7dd1a95

	var req request.Request
	var resp response.Response
	var err error

	for _, tc := range testCases {
		tc := tc

		b.Run(tc.TestName, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				switch tc.MsgType {
				case 0:
					req, err = request.Parse(tc.RawMessage)
					if err != nil &&
						errors.Cause(err).Error() != "message sets version 0 (pre Kafka 0.10) are not supported" &&
						errors.Cause(err).Error() != "message sets version 1 (Kafka 0.10) are not supported" {
						b.Error(err)
						return
					}
					req.Release()
				case 1:
					resp, err = response.Parse(tc.RawMessage, tc.ApiKey, tc.ApiVersion, 2)
					if err != nil {
						b.Error(err)
						return
					}
					resp.Release()

				default:
					b.Errorf("unexpected message type %d", tc.MsgType)
				}
			}
			g_req = req
			g_resp = resp
		})
	}
}

var g_serializedMsg []byte

func BenchmarkRequestResponseSerialize(b *testing.B) {
	var err error

	var serializedMsg []byte
	for _, tc := range setup() {
		tc := tc

		var req *request.Request
		var resp *response.Response

		switch tc.MsgType {
		case 0:
			r, err := request.Parse(tc.RawMessage)
			if err != nil {
				if errors.Cause(err).Error() != "message sets version 0 (pre Kafka 0.10) are not supported" &&
					errors.Cause(err).Error() != "message sets version 1 (Kafka 0.10) are not supported" {
					b.Error(err)
					return
				}
				continue
			}
			req = &r
		case 1:
			r, err := response.Parse(tc.RawMessage, tc.ApiKey, tc.ApiVersion, 2)
			if err != nil {
				b.Error(err)
				return
			}
			resp = &r
		}

		b.Run(tc.TestName, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				switch tc.MsgType {
				case 0:
					serializedMsg, err = req.Serialize(tc.ApiVersion)
					if err != nil {
						b.Error(err)
						return
					}
				case 1:
					serializedMsg, err = resp.Serialize(tc.ApiVersion)
					if err != nil {
						b.Error(err)
						return
					}

				default:
					b.Errorf("unexpected message type %d", tc.MsgType)
					return
				}
			}
			g_serializedMsg = serializedMsg
		})
		b.StopTimer()
		if req != nil {
			req.Release()
		}
		if resp != nil {
			resp.Release()
		}
		b.StartTimer()
	}
}

//nolint:gocognit
func BenchmarkZippedRequestResponseParse(b *testing.B) {
	testCases := []struct {
		testName      string
		requestsFile  string
		responsesFile string
	}{
		{
			testName:      "test-1",
			requestsFile:  "protocol_payload/requests-0.dmp",
			responsesFile: "protocol_payload/responses-0.dmp",
		},
		{
			testName:      "test-2",
			requestsFile:  "protocol_payload/requests-1.dmp",
			responsesFile: "protocol_payload/responses-1.dmp",
		},
	}

	for _, tc := range testCases {
		tc := tc

		b.Run(tc.testName, func(b *testing.B) {
			requestMessages, err := test.KafkaProtocolPayloads.ReadFile(tc.requestsFile)
			if err != nil {
				b.Error(err)
				return
			}

			responseMessages, err := test.KafkaProtocolPayloads.ReadFile(tc.responsesFile)
			if err != nil {
				b.Error(err)
				return
			}

			var length int32

			r := bytes.NewReader(requestMessages)
			p := bytes.NewReader(responseMessages)

			for {
				if r.Len() < 4 {
					break
				}
				err = binary.Read(r, binary.BigEndian, &length)
				if err != nil {
					b.Error(err)
					return
				}
				if r.Len() < int(length) {
					break
				}

				requestMessage := make([]byte, length)
				_, err = r.Read(requestMessage)
				if err != nil {
					b.Error(err)
					return
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// parse request
					req, err := request.Parse(requestMessage)
					if err != nil {
						b.Error(err)
						return
					}
					req.Release()
					g_req = req
				}

				if r.Len() < 4 {
					break
				}
				err = binary.Read(p, binary.BigEndian, &length)
				if err != nil {
					b.Error(err)
					return
				}
				if r.Len() < int(length) {
					break
				}

				responseMessage := make([]byte, length)
				_, err = p.Read(responseMessage)
				if err != nil {
					b.Error(err)
					return
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// parse response
					resp, err := response.Parse(
						responseMessage,
						g_req.HeaderData().RequestApiKey(),
						g_req.HeaderData().RequestApiVersion(),
						g_req.HeaderData().CorrelationId())
					if err != nil {
						b.Error(err)
						return
					}
					resp.Release()
					g_resp = resp
				}
			}
		})
	}
}
