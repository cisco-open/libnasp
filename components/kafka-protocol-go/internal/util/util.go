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

package util

import "unicode"

func CamelCase(s string) string {
	r := []rune(s)
	if len(r) > 0 && unicode.IsUpper(r[0]) {
		r[0] = unicode.ToLower(r[0])
		s = string(r)
	}

	return s
}

func PascalCase(s string) string {
	r := []rune(s)
	if len(r) > 0 && unicode.IsLower(r[0]) {
		r[0] = unicode.ToUpper(r[0])
		s = string(r)
	}

	return s
}

func EscapeGoReservedKeyword(s string) string {
	switch s {
	case "type":
		return "_type"
	default:
		return s
	}
}
