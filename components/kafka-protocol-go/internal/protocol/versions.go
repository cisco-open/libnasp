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

import (
	"math"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

// VersionRange parses the provided version string and returns a tuple
// that corresponds to the version range described in version.
// e.g. for "0-2" returns (0,2)
//
//	for "3+" returns (3,  MaxInt16)
func VersionRange(versions string) (int16, int16, error) {
	boundaries := strings.Split(versions, "-")
	if len(boundaries) == 2 {
		lower, err := strconv.ParseInt(boundaries[0], 10, 16)
		if err != nil {
			return -1, -1, err
		}
		upper, err := strconv.ParseInt(boundaries[1], 10, 16)
		if err != nil {
			return -1, -1, err
		}
		return int16(lower), int16(upper), nil
	}

	boundaries = strings.Split(versions, "+")
	if len(boundaries) == 2 {
		lower, err := strconv.ParseInt(boundaries[0], 10, 16)
		if err != nil {
			return -1, -1, err
		}
		return int16(lower), math.MaxInt16, nil
	}

	v, err := strconv.ParseInt(versions, 10, 16)
	if err == nil {
		return int16(v), int16(v), nil
	}

	return -1, -1, errors.Errorf("unsupported versions string format: %q", versions)
}
