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
	"strconv"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
)

func PrimitiveTypeSliceEqual[T bool | int8 | int16 | uint16 | int32 | int64 | float64, S []T](x, y S) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}

	return true
}

func RawTaggedFieldsEqual(x, y []RawTaggedField) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if !x[i].Equal(&y[i]) {
			return false
		}
	}

	return true
}

func NullableStringSliceEqual(x, y []NullableString) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if !x[i].Equal(&y[i]) {
			return false
		}
	}

	return true
}

func MarshalPrimitiveTypeJSON[T bool | float64 | int8 | int16 | uint16 | int32 | int64 | NullableString](val T) ([]byte, error) {
	switch p := any(val).(type) {
	case bool:
		if p {
			return bytes.TrueLiteral, nil
		}
		return bytes.FalseLiteral, nil
	case float64:
		return []byte(strconv.FormatFloat(p, 'f', -1, 64)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(p), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(p), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(p), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(p), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(p, 10)), nil
	case NullableString:
		j, err := p.MarshalJSON()
		if err != nil {
			return nil, err
		}
		return j, nil
	}

	return nil, nil
}
