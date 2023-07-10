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

package verify_test

import (
	"crypto/x509/pkix"
	"fmt"
	"reflect"
	"testing"

	"github.com/cisco-open/nasp/pkg/tls/verify"
)

func TestParseDN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dn   string
		name pkix.Name
	}{
		{
			dn: "OU=foo",
			name: pkix.Name{
				OrganizationalUnit: []string{"foo"},
			},
		},
		{
			dn: "/OU=foo",
			name: pkix.Name{
				OrganizationalUnit: []string{"foo"},
			},
		},
		{
			dn: "/OU=foo/OU=bar",
			name: pkix.Name{
				OrganizationalUnit: []string{"foo", "bar"},
			},
		},
		{
			dn: "CN=ACME CA",
			name: pkix.Name{
				CommonName: "ACME CA",
			},
		},
		{
			dn: "CN=Sample Cert, OU = R&D;O=Company Ltd./L=Dublin,C=IE, S=bogus",
			name: pkix.Name{
				Country:            []string{"IE"},
				Organization:       []string{"Company Ltd."},
				Locality:           []string{"Dublin"},
				CommonName:         "Sample Cert",
				OrganizationalUnit: []string{"R&D"},
			},
		},
	}

	for k, test := range tests {
		test := test

		t.Run(fmt.Sprintf("test-%d", k), func(t *testing.T) {
			t.Parallel()

			gotName := verify.ParseDN(test.dn)
			if !reflect.DeepEqual(gotName, test.name) {
				t.Fatalf("ParseDN returned mismatch: %v != %v", test.name, gotName)
			}
		})
	}
}
