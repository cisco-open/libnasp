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
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/ca/selfsigned"
	"github.com/cisco-open/nasp/pkg/tls/verify"
)

const (
	CiscoWebPublicCert = `-----BEGIN CERTIFICATE-----
MIIH+TCCBuGgAwIBAgIQQAGGnuzFNavVFtEVb3iGMDANBgkqhkiG9w0BAQsFADBy
MQswCQYDVQQGEwJVUzESMBAGA1UEChMJSWRlblRydXN0MS4wLAYDVQQLEyVIeWRy
YW50SUQgVHJ1c3RlZCBDZXJ0aWZpY2F0ZSBTZXJ2aWNlMR8wHQYDVQQDExZIeWRy
YW50SUQgU2VydmVyIENBIE8xMB4XDTIzMDMwMTIwNDYwMloXDTI0MDIyOTIwNDUw
MlowajEWMBQGA1UEAxMNd3d3LmNpc2NvLmNvbTEbMBkGA1UEChMSQ2lzY28gU3lz
dGVtcyBJbmMuMREwDwYDVQQHEwhTYW4gSm9zZTETMBEGA1UECBMKQ2FsaWZvcm5p
YTELMAkGA1UEBhMCVVMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDY
go1yYvHWGGMGMgq/muSYkXLLxhD92ZVpYD5ZDxwjBNiUwQO37hjy7y1ovmykw9CA
nXvsYPmof7ovH08qzFePkXK9k2zgjOEFsCVkHJ8MGUn0o9p4RgdEjIAhSUwgh1co
rKfm4o5JhOLEBROBz6vPkqCKcsBMFe79oA2+TD5b8axcEOVbWcYiGAMx9pF7YKZy
hSCHt75d7EGGaQWAwP6cYw3ICnvk4e5bAJvJwnEwuU4m8z+FokSRraUhIW4xZiq+
msWnf79AVMxpg4CYD06YFX7VSZ7IuH9l68FIScKLapMu8XX4AVqY0BWBfBgTb5XJ
ZxZnoJ6IcqIsyDPh1M8XAgMBAAGjggSRMIIEjTAOBgNVHQ8BAf8EBAMCBaAwgYUG
CCsGAQUFBwEBBHkwdzAwBggrBgEFBQcwAYYkaHR0cDovL2NvbW1lcmNpYWwub2Nz
cC5pZGVudHJ1c3QuY29tMEMGCCsGAQUFBzAChjdodHRwOi8vdmFsaWRhdGlvbi5p
ZGVudHJ1c3QuY29tL2NlcnRzL2h5ZHJhbnRpZGNhTzEucDdjMB8GA1UdIwQYMBaA
FIm4m7ae7fuwxr0N7GdOPKOSnS35MIIBJgYDVR0gBIIBHTCCARkwDAYKYIZIAYb5
LwAGAzCCAQcGBmeBDAECAjCB/DBABggrBgEFBQcCARY0aHR0cHM6Ly9zZWN1cmUu
aWRlbnRydXN0LmNvbS9jZXJ0aWZpY2F0ZXMvcG9saWN5L3RzLzCBtwYIKwYBBQUH
AgIwgaoMgadUaGlzIFRydXN0SUQgU2VydmVyIENlcnRpZmljYXRlIGhhcyBiZWVu
IGlzc3VlZCBpbiBhY2NvcmRhbmNlIHdpdGggSWRlblRydXN0J3MgVHJ1c3RJRCBD
ZXJ0aWZpY2F0ZSBQb2xpY3kgZm91bmQgYXQgaHR0cHM6Ly9zZWN1cmUuaWRlbnRy
dXN0LmNvbS9jZXJ0aWZpY2F0ZXMvcG9saWN5L3RzLzBGBgNVHR8EPzA9MDugOaA3
hjVodHRwOi8vdmFsaWRhdGlvbi5pZGVudHJ1c3QuY29tL2NybC9oeWRyYW50aWRj
YW8xLmNybDCBnwYDVR0RBIGXMIGUgg13d3cuY2lzY28uY29tghR3d3cuc3RhdGlj
LWNpc2NvLmNvbYIXd3d3LWNsb3VkLWNkbi5jaXNjby5jb22CGHd3dy5tZWRpYWZp
bGVzLWNpc2NvLmNvbYIbd3d3LWRldi1jbG91ZC1jZG4uY2lzY28uY29tgh13d3ct
c3RhZ2UtY2xvdWQtY2RuLmNpc2NvLmNvbTAdBgNVHQ4EFgQUWNN+ZnIToLXY07Oi
2v1ba3enULIwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIIBfgYKKwYB
BAHWeQIEAgSCAW4EggFqAWgAdwB2/4g/Crb7lVHCYcz1h7o0tKTNuyncaEIKn+Zn
TFo6dAAAAYae7MjfAAAEAwBIMEYCIQD/+6LSna0aTdLd8BAEgYxfCpEDeqXvZML+
IYvJbztF3gIhAKDJGElBu+Jw5U9jwfpcO8YjpeFLFADa3CcLFfO7A3eCAHYA7s3Q
ZNXbGs7FXLedtM0TojKHRny87N7DUUhZRnEftZsAAAGGnuzG6wAABAMARzBFAiBP
lZxu0Un31OYO9W2nRdwuowrQnK99LZIbg4lKAnRtngIhALtIuWGh+KmezeKcI7Pt
DVxYxpq4MBwvxv8pyQFBZuTEAHUAc9meiRtMlnigIH1HneayxhzQUV5xGSqMa4AQ
esF3crUAAAGGnuzGwQAABAMARjBEAiA9id+RGmTXIjxm4/9jYOMznM4IPwXbKhF8
TmAz4akUwAIgEmFm+ijbs1dh+BEjvJ3R/odjD0Jp4cVEw1mjSI9vA5IwDQYJKoZI
hvcNAQELBQADggEBANyUUX18WwStIxF3O/fInhbCDe8xZQv3+mrmdDc92thrKWWU
KQzFmwylNdsu1ShV9Z2sVs7FP475eh7lFVIHgIH5igLGcMvHRKhJR9CIiG6v5pLX
jv8cEhTweYQAvRZC0psvfrUWoV6vLweB0aZRqgG2hS8piSpqzLhrcNqVHx6hyWo2
KT/+UEVq47pXpRwbNcGsTyXrltv5N7muHw1sVzg891T0tZnqcEkcUO89qeneER4H
JqkK9n7C1D6ExbZMqWZlZKl5ejYVGcHwJM2kVdJfuXdeAYi4Rf66U80G0OSJXHA/
d2dWJ4ByEyWE8la3GIMxUCtBik0UND/In5JlgxM=
-----END CERTIFICATE-----`
)

func TestCertificateVerify(t *testing.T) { //nolint:funlen
	t.Parallel()

	cac, err := selfsigned.NewSelfSignedCAClient(selfsigned.WithCASubject(pkix.Name{
		CommonName: "test ca",
	}))
	if err != nil {
		t.Fatal(err.Error(), errors.GetDetails(err))
	}

	cc, err := cac.CreateIntermediateCA(pkix.Name{
		CommonName: "test intermediate ca",
	})
	if err != nil {
		t.Fatal(err.Error(), errors.GetDetails(err))
	}

	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM(cac.GetCAPem())

	type Test struct {
		SANMatchers   []string
		Roots         *x509.CertPool
		Intermediates [][]byte
		Valid         bool
	}

	for i, certTest := range []struct {
		CertSANs []string
		Tests    []Test
	}{
		{
			CertSANs: []string{
				"*.acme.corp",
				"spiffe://cluster.local/ns/default/sa/default",
				"127.0.0.1",
				"ops@acme.corp",
			},
			Tests: []Test{
				{
					SANMatchers: []string{
						"URI:PREFIX:1:spiffe://cluster.local",
					},
					Intermediates: [][]byte{
						cc.CertificateDER,
					},
					Valid: true,
				},
				{
					SANMatchers: []string{
						"URI:PREFIX:1:spiffe://acme.cluster.local",
					},
					Intermediates: [][]byte{
						cc.CertificateDER,
					},
					Valid: false,
				},
				{
					Valid: false,
				},
				{
					Intermediates: [][]byte{
						cc.CertificateDER,
					},
					Valid: true,
				},
				{
					SANMatchers: []string{
						"URI:PREFIX:1:spiffe://acme.cluster.local",
						"DNS:CONTAINS:1:acme",
					},
					Intermediates: [][]byte{
						cc.CertificateDER,
					},
					Valid: true,
				},
				{
					SANMatchers: []string{
						"DNS:HOSTNAME:1:coyote.acme.corp",
					},
					Intermediates: [][]byte{
						cc.CertificateDER,
					},
					Valid: true,
				},
			},
		},
	} {
		certTest := certTest

		cert, err := cac.GetCertificateWithSigner(pkix.Name{
			CommonName: fmt.Sprintf("cert-%d", i),
		}, certTest.CertSANs, time.Hour*24, cc)
		if err != nil {
			t.Fatal(err.Error(), errors.GetDetails(err))
		}

		spkiHash, err := verify.GetSPKIHash(cert.Certificate)
		if err != nil {
			t.Fatal(err.Error(), errors.GetDetails(err))
		}
		certHash := fmt.Sprintf("%x", sha256.Sum256(cert.Certificate.Raw))

		t.Run(cert.Certificate.Subject.CommonName, func(t *testing.T) {
			t.Parallel()

			for i, test := range certTest.Tests {
				test := test

				t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
					t.Parallel()

					sanMatchers := []verify.SANMatcher{}
					for _, sanMatcherRaw := range test.SANMatchers {
						if matcher, err := verify.NewSANMatcherFromString(sanMatcherRaw); err != nil {
							t.Fatal(err.Error(), errors.GetDetails(err))
						} else {
							sanMatchers = append(sanMatchers, matcher)
						}
					}

					v := verify.NewCertVerifier(&verify.CertVerifierConfig{
						Roots: cp,

						MatchCertificateHash: []string{certHash, "test"}, // any
						MatchSPKIHash:        []string{spkiHash, "test"}, // any
						MatchTypedSAN:        sanMatchers,                // any
					})

					_, err = v.VerifyCertificate(append([][]byte{cert.CertificateDER}, test.Intermediates...))
					if test.Valid && err != nil {
						t.Fatal(err.Error(), errors.GetDetails(err))
					}
					if !test.Valid && err == nil {
						t.Fatal(errors.New("cert should have been invalid"))
					}
				})
			}
		})
	}
}

func TestCertificateVerifySystemRoots(t *testing.T) {
	t.Parallel()

	decodedPEM, _ := pem.Decode([]byte(CiscoWebPublicCert))

	// it should pass with trusted system roots
	v := verify.NewCertVerifier(&verify.CertVerifierConfig{
		TrustSystemRoots: true,
	})

	_, err := v.VerifyCertificate([][]byte{decodedPEM.Bytes})
	if err != nil {
		t.Fatal(err.Error(), errors.GetDetails(err))
	}

	// it should fail without trusted system roots
	v = verify.NewCertVerifier(&verify.CertVerifierConfig{
		TrustSystemRoots: false,
	})

	if _, err = v.VerifyCertificate([][]byte{decodedPEM.Bytes}); err == nil {
		t.Fatal(errors.New("cert should have been invalid"))
	}
}
