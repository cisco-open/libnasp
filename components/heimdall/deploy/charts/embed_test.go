// Copyright (c) 2021, and 2022 Cisco and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package charts_test

import (
	"io"
	"os"
	"testing"

	"github.com/cisco-open/libnasp/components/heimdall/deploy/charts"
)

func TestEmbed(t *testing.T) {
	t.Parallel()

	file, err := charts.Heimdall.Open("Chart.yaml")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	embeddedContent, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	localContent, err := os.ReadFile("heimdall/Chart.yaml")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if string(embeddedContent) != string(localContent) {
		t.Fatalf("embedded content %s does not equal local content %s", string(embeddedContent), string(localContent))
	}
}
