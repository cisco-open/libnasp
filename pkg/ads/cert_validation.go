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

package ads

import (
	"crypto/x509"
	"os"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/cisco-open/libnasp/pkg/tls/verify"
	"github.com/cisco-open/libnasp/pkg/util"
)

func GetCertVerifierConfigFromValidationContext(validationContext *extensions_transport_sockets_tls_v3.CertificateValidationContext) *verify.CertVerifierConfig {
	config := &verify.CertVerifierConfig{}

	// trusted CA
	if trustedCA := validationContext.GetTrustedCa(); trustedCA != nil {
		var trustedCAPem []byte

		switch specifier := trustedCA.GetSpecifier().(type) {
		case *corev3.DataSource_EnvironmentVariable:
			trustedCAPem = []byte(os.Getenv(specifier.EnvironmentVariable))
		case *corev3.DataSource_Filename:
			if content, err := os.ReadFile(specifier.Filename); err == nil {
				trustedCAPem = content
			}
		case *corev3.DataSource_InlineBytes:
			trustedCAPem = specifier.InlineBytes
		case *corev3.DataSource_InlineString:
			trustedCAPem = []byte(specifier.InlineString)
		}

		if len(trustedCAPem) > 0 {
			config.Roots = x509.NewCertPool()
			config.Roots.AppendCertsFromPEM(trustedCAPem)
		}
	}

	// SPKIs
	config.MatchSPKIHash = validationContext.GetVerifyCertificateSpki()

	// Certificate hash
	config.MatchCertificateHash = validationContext.GetVerifyCertificateHash()

	// SAN matchers
	if len(validationContext.GetMatchTypedSubjectAltNames()) > 0 {
		config.MatchTypedSAN = []verify.SANMatcher{}
	}

	for _, matcher := range validationContext.GetMatchSubjectAltNames() {
		validationContext.MatchTypedSubjectAltNames = append(validationContext.MatchTypedSubjectAltNames, &extensions_transport_sockets_tls_v3.SubjectAltNameMatcher{
			SanType: extensions_transport_sockets_tls_v3.SubjectAltNameMatcher_SAN_TYPE_UNSPECIFIED,
			Matcher: matcher,
		})
	}

	for _, matcher := range validationContext.GetMatchTypedSubjectAltNames() {
		var sanType verify.SANType
		switch matcher.GetSanType() {
		case extensions_transport_sockets_tls_v3.SubjectAltNameMatcher_DNS:
			sanType = verify.DNSSANType
		case extensions_transport_sockets_tls_v3.SubjectAltNameMatcher_EMAIL:
			sanType = verify.EmailSANType
		case extensions_transport_sockets_tls_v3.SubjectAltNameMatcher_IP_ADDRESS:
			sanType = verify.IPSANType
		case extensions_transport_sockets_tls_v3.SubjectAltNameMatcher_URI:
			sanType = verify.URISANType
		case extensions_transport_sockets_tls_v3.SubjectAltNameMatcher_SAN_TYPE_UNSPECIFIED:
		default:
			continue
		}

		if sanType == "" {
			for _, st := range []verify.SANType{verify.DNSSANType, verify.EmailSANType, verify.IPSANType, verify.URISANType} {
				if matcher, err := util.NewStringMatcherFromXDSProto(matcher.GetMatcher()); err == nil {
					config.MatchTypedSAN = append(config.MatchTypedSAN, verify.SANMatcher{
						Type:    st,
						Matcher: matcher,
					})
				}
			}
		} else {
			if matcher, err := util.NewStringMatcherFromXDSProto(matcher.GetMatcher()); err == nil {
				config.MatchTypedSAN = append(config.MatchTypedSAN, verify.SANMatcher{
					Type:    sanType,
					Matcher: matcher,
				})
			}
		}
	}

	return config
}
