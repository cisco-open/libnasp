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

package fields

import (
	"bytes"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
)

type OptionalTag struct {
	tag       uint32
	isDefined bool
}

func (t *OptionalTag) Set(tag uint32) {
	t.tag = tag
	t.isDefined = true
}

func (t *OptionalTag) Get() uint32 {
	return t.tag
}

func (t *OptionalTag) IsDefined() bool {
	return t.isDefined
}
func Tag(tag uint32) OptionalTag {
	return OptionalTag{tag: tag, isDefined: true}
}

type Context struct {
	CustomDefaultValue              any
	SpecName                        string
	SpecTag                         OptionalTag
	LowestSupportedVersion          int16
	HighestSupportedVersion         int16
	LowestSupportedFlexVersion      int16
	HighestSupportedFlexVersion     int16
	LowestSupportedNullableVersion  int16
	HighestSupportedNullableVersion int16
	LowestSupportedTaggedVersion    int16
	HighestSupportedTaggedVersion   int16
}

func (f *Context) Name() string {
	return f.SpecName
}

func (f *Context) Tag() OptionalTag {
	return f.SpecTag
}

func (f *Context) IsTaggedVersion(version int16) bool {
	return f.SpecTag.isDefined && version >= f.LowestSupportedTaggedVersion && version <= f.HighestSupportedTaggedVersion
}

func (f *Context) IsSupportedVersion(version int16) bool {
	return version >= f.LowestSupportedVersion && version <= f.HighestSupportedVersion
}

func (f *Context) IsFlexibleVersion(version int16) bool {
	return version >= f.LowestSupportedFlexVersion && version <= f.HighestSupportedFlexVersion
}

func (f *Context) IsNullableVersion(version int16) bool {
	return version >= f.LowestSupportedNullableVersion && version <= f.HighestSupportedNullableVersion
}

func (f *Context) OnlyTaggedVersionsSupported() bool {
	return f.LowestSupportedTaggedVersion == f.LowestSupportedVersion &&
		f.HighestSupportedTaggedVersion == f.HighestSupportedVersion
}

func (f *Context) NonTaggedVersionsSupported() bool {
	return (f.LowestSupportedTaggedVersion == -1 && f.HighestSupportedTaggedVersion == -1) ||
		!(f.LowestSupportedTaggedVersion == f.LowestSupportedVersion &&
			f.HighestSupportedTaggedVersion == f.HighestSupportedVersion)
}

type PrimitiveTypeProcessor[T bool | float64 | int8 | int16 | uint16 | int32 | int64 | NullableString] interface {
	Read(buf *bytes.Reader, messageVersion int16, out *T) error
	Write(w *typesbytes.SliceWriter, messageVersion int16, data T) error
	SizeInBytes(version int16, data T) (int, error)
}

type StructType interface {
	Read(r *bytes.Reader, messageVersion int16) error
	Write(w *typesbytes.SliceWriter, messageVersion int16) error
	SizeInBytes(version int16) (int, error)
	MarshalJSON() ([]byte, error)
}
