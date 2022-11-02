// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxywasm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cisco-open/nasp/pkg/proxywasm"
)

func TestStringify(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	for _, tc := range []struct {
		m        map[string]string
		expected string
	}{
		{
			map[string]string{"key": "key", "value": "value"},
			"\x02\x00\x00\x00\x03\x00\x00\x00\x03\x00\x00\x00\x05\x00\x00\x00\x05\x00\x00\x00key\x00key\x00value\x00value\x00",
		},
		{
			map[string]string{"key": "key", "value": "value", "type": "example"},
			"\x03\x00\x00\x00\x03\x00\x00\x00\x03\x00\x00\x00\x04\x00\x00\x00\a\x00\x00\x00\x05\x00\x00\x00\x05\x00\x00\x00key\x00key\x00type\x00example\x00value\x00value\x00",
		},
	} {
		// run several times to check if output is deterministic
		for i := 0; i < 5; i++ {
			assert.Equal(tc.expected, proxywasm.Stringify(tc.m))
		}
	}
}
