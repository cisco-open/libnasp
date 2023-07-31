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

package util

import (
	"fmt"
	"regexp"
	"strings"

	"emperror.dev/errors"
	"istio.io/istio/pkg/config/host"

	v3matcherpb "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

type StringMatchType string

const (
	ExactStringMatchType    StringMatchType = "EXACT"
	PrefixStringMatchType   StringMatchType = "PREFIX"
	SuffixStringMatchType   StringMatchType = "SUFFIX"
	ContainsStringMatchType StringMatchType = "CONTAINS"
	RegexStringMatchType    StringMatchType = "REGEX"
	HostnameStringMatchType StringMatchType = "HOSTNAME"
)

type StringMatcher interface {
	Match(input string) bool
	String() string
}

type stringMatcher struct {
	exact      *string
	prefix     *string
	suffix     *string
	contains   *string
	regex      *regexp.Regexp
	dns        *string
	ignoreCase bool
}

func NewStringMatcher(matchType StringMatchType, value string, ignoreCase bool) (StringMatcher, error) {
	m := &stringMatcher{
		ignoreCase: ignoreCase,
	}

	if err := m.setMatcher(matchType, value); err != nil {
		return nil, err
	}

	return m, nil
}

func NewStringMatcherFromString(matcher string) (StringMatcher, error) {
	p := strings.SplitN(matcher, ":", 3)
	if len(p) < 3 {
		return nil, errors.NewWithDetails("invalid matcher string", "matcher", matcher)
	}

	ignoreCase := false
	if p[1] != "" && p[1] != "0" {
		ignoreCase = true
	}

	return NewStringMatcher(StringMatchType(p[0]), p[2], ignoreCase)
}

func NewStringMatcherFromXDSProto(matcherProto *v3matcherpb.StringMatcher) (StringMatcher, error) {
	if matcherProto == nil {
		return nil, errors.New("input StringMatcher proto is nil")
	}

	matcher := &stringMatcher{ignoreCase: matcherProto.GetIgnoreCase()}
	switch mt := matcherProto.GetMatchPattern().(type) {
	case *v3matcherpb.StringMatcher_Exact:
		if err := matcher.setMatcher(ExactStringMatchType, mt.Exact); err != nil {
			return nil, err
		}
	case *v3matcherpb.StringMatcher_Prefix:
		if err := matcher.setMatcher(PrefixStringMatchType, mt.Prefix); err != nil {
			return nil, err
		}
	case *v3matcherpb.StringMatcher_Suffix:
		if err := matcher.setMatcher(SuffixStringMatchType, mt.Suffix); err != nil {
			return nil, err
		}
	case *v3matcherpb.StringMatcher_Contains:
		if err := matcher.setMatcher(ContainsStringMatchType, mt.Contains); err != nil {
			return nil, err
		}
	case *v3matcherpb.StringMatcher_SafeRegex:
		if err := matcher.setMatcher(RegexStringMatchType, matcherProto.GetSafeRegex().GetRegex()); err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("invalid matcher: %v", matcherProto)
	}

	return matcher, nil
}

func (m *stringMatcher) String() string {
	switch {
	case m.exact != nil:
		return fmt.Sprintf("%s:%s", ExactStringMatchType, *m.exact)
	case m.prefix != nil:
		return fmt.Sprintf("%s:%s", PrefixStringMatchType, *m.prefix)
	case m.suffix != nil:
		return fmt.Sprintf("%s:%s", SuffixStringMatchType, *m.suffix)
	case m.contains != nil:
		return fmt.Sprintf("%s:%s", ContainsStringMatchType, *m.contains)
	case m.dns != nil:
		return fmt.Sprintf("%s:%s", HostnameStringMatchType, *m.dns)
	case m.regex != nil:
		r := *m.regex
		return fmt.Sprintf("%s:%s", RegexStringMatchType, r.String())
	}

	return ""
}

func (m *stringMatcher) Match(input string) bool {
	if m.ignoreCase {
		input = strings.ToLower(input)
	}

	switch {
	case m.exact != nil:
		return input == *m.exact
	case m.prefix != nil:
		return strings.HasPrefix(input, *m.prefix)
	case m.suffix != nil:
		return strings.HasSuffix(input, *m.suffix)
	case m.contains != nil:
		return strings.Contains(input, *m.contains)
	case m.dns != nil:
		return host.Name(*m.dns).Matches(host.Name(input))
	case m.regex != nil:
		if len(input) == 0 {
			return m.regex.MatchString(input)
		}

		m.regex.Longest()
		rem := m.regex.FindString(input)

		return len(rem) == len(input)
	}

	return false
}

func (m *stringMatcher) setMatcher(matchType StringMatchType, value string) error {
	if m.ignoreCase {
		value = strings.ToLower(value)
	}

	switch matchType {
	case ExactStringMatchType:
		m.exact = &value
	case PrefixStringMatchType:
		m.prefix = &value
	case SuffixStringMatchType:
		m.suffix = &value
	case ContainsStringMatchType:
		m.contains = &value
	case HostnameStringMatchType:
		m.dns = &value
	case RegexStringMatchType:
		re, err := regexp.Compile(value)
		if err != nil {
			return errors.Errorf("invalid regex: %q", value)
		}
		m.regex = re
	default:
		return errors.Errorf("invalid match type: %q", matchType)
	}

	return nil
}
