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

package main

import (
	"math"
	"strconv"
	"strings"

	"emperror.dev/errors"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/internal/protocol"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
)

const (
	Bool    = "bool"
	Int8    = "int8"
	Int16   = "int16"
	Uint16  = "uint16"
	Int32   = "int32"
	Int64   = "int64"
	Float64 = "float64"
	String  = "string"
	Uuid    = "uuid"
	Bytes   = "bytes"
	Records = "records"

	null = "null"
	none = "none"
)

type structConstraints struct {
	LowestSupportedVersion      int16
	HighestSupportedVersion     int16
	LowestSupportedFlexVersion  int16
	HighestSupportedFlexVersion int16
	HasKnownTaggedFields        bool
}

func structConstraintsFromMessageSpec(spec protocol.MessageSpec) (structConstraints, error) {
	lowestSupportedVersion, highestSupportedVersion, err := protocol.VersionRange(spec.ValidVersions)
	if err != nil {
		return structConstraints{}, errors.WrapIff(err, "can not determine supported versions for message %q", spec.Name)
	}

	lowestSupportedFlexVersion := int16(math.MaxInt16)
	highestSupportedFlexVersion := int16(math.MaxInt16)
	if len(spec.FlexibleVersions) > 0 && spec.FlexibleVersions != none {
		lowestSupportedFlexVersion, highestSupportedFlexVersion, err = protocol.VersionRange(spec.FlexibleVersions)
		if err != nil {
			return structConstraints{}, errors.WrapIff(err, "can not determine supported flexible versions for message %q", spec.Name)
		}
	}

	constraints := structConstraints{
		LowestSupportedVersion:      lowestSupportedVersion,
		HighestSupportedVersion:     highestSupportedVersion,
		LowestSupportedFlexVersion:  lowestSupportedFlexVersion,
		HighestSupportedFlexVersion: highestSupportedFlexVersion,
		HasKnownTaggedFields:        hasKnownTaggedFields(spec.Fields),
	}

	return constraints, nil
}

func structConstraintsFromStructSpec(spec protocol.StructSpec) (structConstraints, error) {
	lowestSupportedVersion, highestSupportedVersion, err := protocol.VersionRange(spec.Versions)
	if err != nil {
		return structConstraints{}, errors.WrapIff(err, "can not determine supported versions for %q", spec.Name)
	}

	constraints := structConstraints{
		LowestSupportedVersion:      lowestSupportedVersion,
		HighestSupportedVersion:     highestSupportedVersion,
		LowestSupportedFlexVersion:  -1,
		HighestSupportedFlexVersion: -1,
		HasKnownTaggedFields:        hasKnownTaggedFields(spec.Fields),
	}

	return constraints, nil
}

func structConstraintsFromFieldSpec(parentConstraints structConstraints, spec protocol.MessageFieldSpec) (structConstraints, error) {
	lowestSupportedVersion, highestSupportedVersion, err := protocol.VersionRange(spec.Versions)
	if err != nil {
		return structConstraints{}, errors.WrapIff(err, "can not determine supported versions for %q", spec.Name)
	}

	var lowestSupportedFlexVersion int16
	var highestSupportedFlexVersion int16

	switch spec.FlexibleVersions {
	case "":
		lowestSupportedFlexVersion = parentConstraints.LowestSupportedFlexVersion
		highestSupportedFlexVersion = parentConstraints.HighestSupportedFlexVersion
	case "none":
		lowestSupportedFlexVersion = int16(math.MaxInt16)
		highestSupportedFlexVersion = int16(math.MaxInt16)
	default:
		lowestSupportedFlexVersion, highestSupportedFlexVersion, err = protocol.VersionRange(spec.FlexibleVersions)
		if err != nil {
			return structConstraints{}, errors.WrapIff(err, "can not determine supported flexible versions for %q", spec.Name)
		}
	}

	constraints := structConstraints{
		LowestSupportedVersion:      lowestSupportedVersion,
		HighestSupportedVersion:     highestSupportedVersion,
		LowestSupportedFlexVersion:  lowestSupportedFlexVersion,
		HighestSupportedFlexVersion: highestSupportedFlexVersion,
		HasKnownTaggedFields:        hasKnownTaggedFields(spec.Fields),
	}

	return constraints, nil
}

func hasKnownTaggedFields(fields []protocol.MessageFieldSpec) bool {
	for _, f := range fields {
		if f.Tag != nil {
			return true
		}
	}

	return false
}

func contextFromMessageSpec(spec *protocol.MessageSpec) (*fields.Context, error) {
	lowestSupportedVersion, highestSupportedVersion, err := protocol.VersionRange(spec.ValidVersions)
	if err != nil {
		return nil, errors.WrapIff(err, "can not determine supported versions for message %q", spec.Name)
	}

	lowestSupportedFlexVersion := int16(math.MaxInt16)
	highestSupportedFlexVersion := int16(math.MaxInt16)
	if len(spec.FlexibleVersions) > 0 && spec.FlexibleVersions != none {
		lowestSupportedFlexVersion, highestSupportedFlexVersion, err = protocol.VersionRange(spec.FlexibleVersions)
		if err != nil {
			return nil, errors.WrapIff(err, "can not determine supported flexible versions for message %q", spec.Name)
		}
	}

	context := fields.Context{
		SpecName:                    spec.Name,
		LowestSupportedVersion:      lowestSupportedVersion,
		HighestSupportedVersion:     highestSupportedVersion,
		LowestSupportedFlexVersion:  lowestSupportedFlexVersion,
		HighestSupportedFlexVersion: highestSupportedFlexVersion,
	}

	return &context, nil
}

func contextFromFieldSpec(parentContext *fields.Context, spec *protocol.MessageFieldSpec) (*fields.Context, error) {
	lowestSupportedVersion, highestSupportedVersion, err := protocol.VersionRange(spec.Versions)
	if err != nil {
		return nil, errors.WrapIff(err, "can not determine supported versions for field %q", spec.Name)
	}

	lowestSupportedNullableVersion := int16(math.MaxInt16)
	highestSupportedNullableVersion := int16(math.MaxInt16)
	if parentContext != nil {
		lowestSupportedNullableVersion = parentContext.LowestSupportedNullableVersion
		highestSupportedNullableVersion = parentContext.HighestSupportedNullableVersion
	}
	if len(spec.NullableVersions) > 0 {
		lowestSupportedNullableVersion, highestSupportedNullableVersion, err = protocol.VersionRange(spec.NullableVersions)
		if err != nil {
			return nil, errors.WrapIff(err, "can not determine supported nullable versions for field %q", spec.Name)
		}
	}

	lowestSupportedFlexVersion := int16(math.MaxInt16)
	highestSupportedFlexVersion := int16(math.MaxInt16)
	if parentContext != nil {
		lowestSupportedFlexVersion = parentContext.LowestSupportedFlexVersion
		highestSupportedFlexVersion = parentContext.HighestSupportedFlexVersion
	}
	if len(spec.FlexibleVersions) > 0 {
		if spec.FlexibleVersions == none {
			lowestSupportedFlexVersion = int16(math.MaxInt16)
			highestSupportedFlexVersion = int16(math.MaxInt16)
		} else {
			lowestSupportedFlexVersion, highestSupportedFlexVersion, err = protocol.VersionRange(spec.FlexibleVersions)
			if err != nil {
				return nil, errors.WrapIff(err, "can not determine supported flexible versions for field %q", spec.Name)
			}
		}
	}

	lowestSupportedTaggedVersion := int16(-1)
	highestSupportedTaggedVersion := int16(-1)
	if len(spec.TaggedVersions) > 0 {
		lowestSupportedTaggedVersion, highestSupportedTaggedVersion, err = protocol.VersionRange(spec.TaggedVersions)
		if err != nil {
			return nil, errors.WrapIff(err, "can not determine supported tagged versions for field %q", spec.Name)
		}
	}

	var customDefaultValue any
	if spec.Default != nil {
		customDefaultValue, err = cast(spec.Default, spec.Type)
		if err != nil {
			return nil, errors.WrapIf(err, "couldn't process field's custom default value")
		}
	}

	var specTag fields.OptionalTag
	if spec.Tag != nil {
		specTag.Set(*spec.Tag)
	}

	context := fields.Context{
		SpecName:                        spec.Name,
		SpecTag:                         specTag,
		CustomDefaultValue:              customDefaultValue,
		LowestSupportedVersion:          lowestSupportedVersion,
		HighestSupportedVersion:         highestSupportedVersion,
		LowestSupportedFlexVersion:      lowestSupportedFlexVersion,
		HighestSupportedFlexVersion:     highestSupportedFlexVersion,
		LowestSupportedNullableVersion:  lowestSupportedNullableVersion,
		HighestSupportedNullableVersion: highestSupportedNullableVersion,
		LowestSupportedTaggedVersion:    lowestSupportedTaggedVersion,
		HighestSupportedTaggedVersion:   highestSupportedTaggedVersion,
	}

	return &context, nil
}

func cast(value any, toType string) (any, error) {
	switch toType {
	case Bool:
		return castBool(value)
	case Int8:
		return castInt[int8](value, toType)
	case Int16:
		return castInt[int16](value, toType)
	case Uint16:
		return castInt[uint16](value, toType)
	case Int32:
		return castInt[int32](value, toType)
	case Int64:
		return castInt[int64](value, toType)
	case Float64:
		return castFloat64(value)
	case String:
		return castNullableString(value)
	case Bytes:
		return castBytes(value)
	default:
		if strings.HasPrefix(toType, "[]") {
			// array type
			return castArray(value)
		}

		return nil, errors.Errorf("type %q not supported", toType)
	}
}

func castInt[T int8 | int16 | uint16 | int32 | int64](value any, toType string) (T, error) {
	var zero T

	sValue, isString := value.(string)
	bitSize := 0

	switch toType {
	case Int8:
		bitSize = 8
	case Int16, Uint16:
		bitSize = 16
	case Int32:
		bitSize = 32
	case Int64:
		bitSize = 64
	default:
		return zero, errors.Errorf("%q is not a numeric type", toType)
	}

	if isString {
		i, err := strconv.ParseInt(sValue, 0, bitSize)
		if err != nil {
			return zero, errors.WrapIff(err, "couldn't parse value %V as type %s", value, toType)
		}
		return T(i), nil
	}

	nValue, isNumeric := value.(float64)
	if !isNumeric {
		return zero, errors.Errorf("value %V can not be converted to %s", value, toType)
	}

	return T(nValue), nil
}

func castBool(value any) (bool, error) {
	sValue, isString := value.(string)
	if isString {
		b, err := strconv.ParseBool(sValue)
		if err != nil {
			return false, errors.WrapIff(err, "couldn't parse value %V as type bool", value)
		}
		return b, nil
	}
	if b, ok := value.(bool); ok {
		return b, nil
	}
	return false, errors.Errorf("value %V can not be converted to bool", value)
}

func castFloat64(value any) (float64, error) {
	sValue, isString := value.(string)

	if isString {
		f, err := strconv.ParseFloat(sValue, 64)
		if err != nil {
			return 0, errors.WrapIff(err, "couldn't parse value %V as type \"float64\"", value)
		}
		return f, nil
	}

	nValue, isNumeric := value.(float64)
	if !isNumeric {
		return 0, errors.Errorf("value %V can not be converted to \"float64\"", value)
	}

	return nValue, nil
}

func castBytes(value any) ([]byte, error) {
	sValue, isString := value.(string)

	if isString {
		if sValue == null {
			return []byte(nil), nil
		}
		return nil, errors.Errorf("string value %q not allowed for \"bytes\" type", sValue)
	}
	if b, ok := value.([]byte); ok {
		return b, nil
	}
	return nil, errors.Errorf("value %V can not be converted to bytes", value)
}

func castNullableString(value any) (string, error) {
	sValue, isString := value.(string)
	if !isString {
		return "", errors.Errorf("value %V can not be converted to string", value)
	}

	return sValue, nil
}

func castArray(value any) ([]any, error) {
	sValue, isString := value.(string)

	if isString {
		if sValue == null {
			return []any(nil), nil
		}
		return nil, errors.Errorf("string value %q not allowed for \"array\" type", sValue)
	}
	if b, ok := value.([]any); ok {
		return b, nil
	}
	return nil, errors.Errorf("value %V can not be converted to array", value)
}
