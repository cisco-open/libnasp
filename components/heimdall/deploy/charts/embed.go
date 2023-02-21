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

package charts

import (
	"embed"
	"io/fs"
)

var (
	//go:embed heimdall-webhook heimdall-webhook/templates/_helpers.tpl
	heimdallWebhookEmbed embed.FS

	// HeimdallWebhook exposes the heimdall-webhook chart using relative file paths from the chart root
	HeimdallWebhook fs.FS

	//go:embed heimdall heimdall/templates/_helpers.tpl
	heimdallEmbed embed.FS

	// Heimdall exposes the heimdall chart using relative file paths from the chart root
	Heimdall fs.FS
)

func init() {
	var err error
	HeimdallWebhook, err = fs.Sub(heimdallWebhookEmbed, "heimdall-webhook")
	if err != nil {
		panic(err)
	}

	Heimdall, err = fs.Sub(heimdallEmbed, "heimdall")
	if err != nil {
		panic(err)
	}
}
