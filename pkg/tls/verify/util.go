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
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"net"
	"regexp"
	"strings"
)

var dnRegexp *regexp.Regexp

func init() {
	dnRegexp = regexp.MustCompile(`[/;,]?([^= ]+) ?= ?([^/;,]+)`)
}

func ParseDN(dn string) pkix.Name {
	name := pkix.Name{}

	matches := dnRegexp.FindAllStringSubmatch(dn, -1)

	for _, match := range matches {
		val := match[2]
		if val == "" {
			continue
		}

		switch match[1] {
		case "C":
			name.Country = append(name.Country, val)
		case "O":
			name.Organization = append(name.Organization, val)
		case "OU":
			name.OrganizationalUnit = append(name.OrganizationalUnit, val)
		case "L":
			name.Locality = append(name.Locality, val)
		case "ST":
			name.Province = append(name.Province, val)
		case "SN":
			name.SerialNumber = val
		case "CN":
			name.CommonName = val
		}
	}

	return name
}

func GetSPKIHash(cert *x509.Certificate) (string, error) {
	pkBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", err
	}

	shaSum := sha256.Sum256(pkBytes)

	return base64.RawStdEncoding.EncodeToString(shaSum[:]), nil
}

func GetSANTypeFromSAN(san string) SANType {
	if strings.Contains(san, "://") {
		return URISANType
	}

	if strings.Contains(san, "@") {
		return EmailSANType
	}

	if ip := net.ParseIP(san); ip != nil {
		return IPSANType
	}

	return DNSSANType
}

func SetCertVerifierToTLSConfig(getter CertVerifierConfigGetter, tlsConfig *tls.Config) {
	tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		config := getter.GetCertVerifierConfig()
		if config.Roots == nil {
			config.Roots = tlsConfig.RootCAs
		}

		_, err := NewCertVerifier(config).VerifyCertificate(rawCerts)

		return err
	}
}
