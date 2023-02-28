// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
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

package nasp

import (
	"testing"
)

func TestSelectedKey(t *testing.T) {
	sk := NewSelectedKey(OP_ACCEPT, 12345)

	if sk.Operation() != OP_ACCEPT {
		t.Fatalf("sk.Operation() -> %d != operation -> %d", sk.Operation(), OP_ACCEPT)
	}

	if sk.SelectedKeyId() != 12345 {
		t.Fatalf("sk.SelectedKeyId() -> %d != selectedKeyId -> %d", sk.SelectedKeyId(), 12345)
	}
}
