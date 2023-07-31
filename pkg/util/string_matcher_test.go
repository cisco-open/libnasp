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

package util_test

import (
	"fmt"
	"testing"

	v3matcherpb "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/cisco-open/nasp/pkg/util"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	var (
		exactMatcher, _    = util.NewStringMatcher(util.ExactStringMatchType, "exact", false)
		prefixMatcher, _   = util.NewStringMatcher(util.PrefixStringMatchType, "prefix", false)
		suffixMatcher, _   = util.NewStringMatcher(util.SuffixStringMatchType, "suffix", false)
		containsMatcher, _ = util.NewStringMatcher(util.ContainsStringMatchType, "contains", false)
		regexMatcher, _    = util.NewStringMatcher(util.RegexStringMatchType, "^spiffe://.*$", false)

		exactMatcherIgnoreCase, _    = util.NewStringMatcher(util.ExactStringMatchType, "exact", true)
		prefixMatcherIgnoreCase, _   = util.NewStringMatcher(util.PrefixStringMatchType, "prefix", true)
		suffixMatcherIgnoreCase, _   = util.NewStringMatcher(util.SuffixStringMatchType, "suffix", true)
		containsMatcherIgnoreCase, _ = util.NewStringMatcher(util.ContainsStringMatchType, "contains", true)
	)

	tests := []struct {
		matcher util.StringMatcher
		input   string
		match   bool
	}{
		{
			matcher: exactMatcher,
			input:   "exact",
			match:   true,
		},
		{
			matcher: exactMatcher,
			input:   "Exact",
		},
		{
			matcher: exactMatcherIgnoreCase,
			input:   "Exact",
			match:   true,
		},
		{
			matcher: exactMatcherIgnoreCase,
			input:   "not-exact",
		},
		{
			matcher: prefixMatcher,
			input:   "prefix://cluster.local",
			match:   true,
		},
		{
			matcher: prefixMatcher,
			input:   "Prefix://CLUSTER.LOCAL",
		},
		{
			matcher: prefixMatcherIgnoreCase,
			input:   "Prefix://CLUSTER.LOCAL",
			match:   true,
		},
		{
			matcher: prefixMatcherIgnoreCase,
			input:   "not-prefix",
		},
		{
			matcher: suffixMatcher,
			input:   "spiffe://cluster.local/ns/suffix",
			match:   true,
		},
		{
			matcher: suffixMatcher,
			input:   "spiffe://cluster.local/ns/default",
		},
		{
			matcher: suffixMatcherIgnoreCase,
			input:   "spiffe://cluster.local/ns/SUFFIX",
			match:   true,
		},
		{
			matcher: suffixMatcherIgnoreCase,
			input:   "spiffe://cluster.local/ns/default",
		},
		{
			matcher: regexMatcher,
			input:   "spiffe://cluster.local/ns/suffix",
			match:   true,
		},
		{
			matcher: regexMatcher,
			input:   "http://cluster.local/ns/suffix",
		},
		{
			matcher: containsMatcher,
			input:   "it-contains-this",
			match:   true,
		},
		{
			matcher: containsMatcher,
			input:   "its-not-here",
		},
		{
			matcher: containsMatcherIgnoreCase,
			input:   "it-CONTAINS-this",
			match:   true,
		},
		{
			matcher: containsMatcherIgnoreCase,
			input:   "its-NOT-here",
		},
	}

	for k, test := range tests {
		test := test

		t.Run(fmt.Sprintf("test-%d", k), func(t *testing.T) {
			t.Parallel()

			if gotMatch := test.matcher.Match(test.input); gotMatch != test.match {
				t.Errorf("StringMatcher.Match(%s) returned %v, want %v", test.input, gotMatch, test.match)
			}
		})
	}
}

func TestStringMatcherFromXDSProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputProto *v3matcherpb.StringMatcher
		input      string
		err        bool
	}{
		{
			err: true,
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_SafeRegex{
					SafeRegex: &v3matcherpb.RegexMatcher{Regex: "\\"},
				},
			},
			err: true,
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Exact{Exact: "exact"},
			},
			input: "exact",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Exact{Exact: "EXACT"},
				IgnoreCase:   true,
			},
			input: "Exact",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Prefix{Prefix: "prefix"},
			},
			input: "prefix://cluster.local",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Prefix{Prefix: "PREFIX"},
				IgnoreCase:   true,
			},
			input: "Prefix://CLUSTER.LOCAL",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Suffix{Suffix: "suffix"},
			},
			input: "spiffe://cluster.local/ns/suffix",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Suffix{Suffix: "SUFFIX"},
				IgnoreCase:   true,
			},
			input: "spiffe://cluster.local/ns/SuffiX",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Contains{Contains: "contains"},
			},
			input: "it-contains-this",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_Contains{Contains: "CONTAINS"},
				IgnoreCase:   true,
			},
			input: "it-CONTAINS-this",
		},
		{
			inputProto: &v3matcherpb.StringMatcher{
				MatchPattern: &v3matcherpb.StringMatcher_SafeRegex{
					SafeRegex: &v3matcherpb.RegexMatcher{Regex: "^spiffe://.*$"},
				},
			},
			input: "spiffe://cluster.local/ns/default/sa/default",
		},
	}

	for k, test := range tests {
		test := test

		t.Run(fmt.Sprintf("test-%d", k), func(t *testing.T) {
			t.Parallel()

			matcher, err := util.NewStringMatcherFromXDSProto(test.inputProto)
			if (err != nil) != test.err {
				t.Fatalf("StringMatcherFromXDSProto(%+v) returned unexpected err: %v", test.inputProto, err)
			}
			if matcher != nil && !matcher.Match(test.input) {
				t.Fatalf("StringMatcherFromXDSProto(%+v) unexpected mismatch: %s", test.inputProto, test.input)
			}
		})
	}
}
