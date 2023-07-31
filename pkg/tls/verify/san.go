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

package verify

import (
	"fmt"
	"strings"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/util"
)

type SANType string

const (
	EmailSANType SANType = "EMAIL"
	DNSSANType   SANType = "DNS"
	URISANType   SANType = "URI"
	IPSANType    SANType = "IP"
)

type SANMatcher struct {
	Type    SANType
	Matcher util.StringMatcher
}

func (m SANMatcher) String() string {
	return fmt.Sprintf("%s:%s", m.Type, m.Matcher)
}

func NewSANMatcherFromString(match string) (SANMatcher, error) {
	p := strings.SplitN(match, ":", 2)
	sanType := SANType(p[0])

	switch sanType {
	case EmailSANType, DNSSANType, URISANType, IPSANType:
		if matcher, err := util.NewStringMatcherFromString(p[1]); err != nil {
			return SANMatcher{}, err
		} else {
			return SANMatcher{
				Type:    sanType,
				Matcher: matcher,
			}, nil
		}
	}

	return SANMatcher{}, errors.NewWithDetails("invalid SAN matcher type", "type", p[0])
}
