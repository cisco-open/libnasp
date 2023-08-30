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

package protocol

type MessageSpec struct {
	APIKey           int                `json:"apiKey"`
	Type             string             `json:"type"`
	Listeners        []string           `json:"listeners,omitempty"`
	Name             string             `json:"name"`
	ValidVersions    string             `json:"validVersions"`
	FlexibleVersions string             `json:"flexibleVersions"`
	Fields           []MessageFieldSpec `json:"fields"`
	CommonStructs    []StructSpec       `json:"commonStructs,omitempty"`
}

type MessageFieldSpec struct {
	Name             string             `json:"name"`
	Type             string             `json:"type"`
	MapKey           bool               `json:"mapKey,omitempty"`
	Default          any                `json:"default,omitempty"`
	EntityType       string             `json:"entityType,omitempty"`
	ZeroCopy         *bool              `json:"zeroCopy,omitempty"`
	Versions         string             `json:"versions"`
	FlexibleVersions string             `json:"flexibleVersions,omitempty"`
	NullableVersions string             `json:"nullableVersions,omitempty"`
	TaggedVersions   string             `json:"taggedVersions,omitempty"`
	Ignorable        bool               `json:"ignorable,omitempty"`
	Tag              *uint32            `json:"tag,omitempty"`
	About            string             `json:"about"`
	Fields           []MessageFieldSpec `json:"fields,omitempty"`
}

type StructSpec struct {
	Name     string             `json:"name"`
	Versions string             `json:"versions"`
	Fields   []MessageFieldSpec `json:"fields"`
}
