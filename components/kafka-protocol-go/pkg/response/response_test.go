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

package response_test

import (
	"testing"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/assets/test"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/response"
)

func TestParse(t *testing.T) {
	t.Parallel()

	msg, err := test.KafkaProtocolPayloads.ReadFile("protocol_payload/api_versions_resp.dmp")
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := response.Parse(msg, 18, 3, 2)
	if err != nil {
		t.Error(err)
		return
	}
	resp.Release()

	msg, err = test.KafkaProtocolPayloads.ReadFile("protocol_payload/sync_group_resp.dmp")
	if err != nil {
		t.Error(err)
		return
	}

	resp, err = response.Parse(msg, 14, 5, 2)
	if err != nil {
		t.Error(err)
		return
	}
	resp.Release()

	msg, err = test.KafkaProtocolPayloads.ReadFile("protocol_payload/fetch_recordbatchv2_resp.dmp")
	if err != nil {
		t.Error(err)
		return
	}

	resp, err = response.Parse(msg, 1, 13, 2)
	if err != nil {
		t.Error(err)
		return
	}
	resp.Release()

	msg, err = test.KafkaProtocolPayloads.ReadFile("protocol_payload/fetch_control_recordbatch_resp.dmp")
	if err != nil {
		t.Error(err)
		return
	}

	resp, err = response.Parse(msg, 1, 13, 2)
	if err != nil {
		t.Error(err)
		return
	}
	resp.Release()
}
