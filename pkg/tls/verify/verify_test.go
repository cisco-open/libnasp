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
	GHWebPublicCert = `-----BEGIN CERTIFICATE-----
MIIFajCCBPGgAwIBAgIQDNCovsYyz+ZF7KCpsIT7HDAKBggqhkjOPQQDAzBWMQsw
CQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp
Q2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjMwMjE0MDAw
MDAwWhcNMjQwMzE0MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs
aWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi
LCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAEo6QDRgPfRlFWy8k5qyLN52xZlnqToPu5QByQMog2xgl2nFD1Vfd2Xmgg
nO4i7YMMFTAQQUReMqyQodWq8uVDs6OCA48wggOLMB8GA1UdIwQYMBaAFAq8CCkX
jKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBTHByd4hfKdM8lMXlZ9XNaOcmfr3jAl
BgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E
BAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw
gZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5
YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp
Z2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j
cmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3
dy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG
GGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh
Y2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy
MENBMS0xLmNydDAJBgNVHRMEAjAAMIIBgAYKKwYBBAHWeQIEAgSCAXAEggFsAWoA
dwDuzdBk1dsazsVct520zROiModGfLzs3sNRSFlGcR+1mwAAAYZQ3Rv6AAAEAwBI
MEYCIQDkFq7T4iy6gp+pefJLxpRS7U3gh8xQymmxtI8FdzqU6wIhALWfw/nLD63Q
YPIwG3EFchINvWUfB6mcU0t2lRIEpr8uAHYASLDja9qmRzQP5WoC+p0w6xxSActW
3SyB2bu/qznYhHMAAAGGUN0cKwAABAMARzBFAiAePGAyfiBR9dbhr31N9ZfESC5G
V2uGBTcyTyUENrH3twIhAPwJfsB8A4MmNr2nW+sdE1n2YiCObW+3DTHr2/UR7lvU
AHcAO1N3dT4tuYBOizBbBv5AO2fYT8P0x70ADS1yb+H61BcAAAGGUN0cOgAABAMA
SDBGAiEAzOBr9OZ0+6OSZyFTiywN64PysN0FLeLRyL5jmEsYrDYCIQDu0jtgWiMI
KU6CM0dKcqUWLkaFE23c2iWAhYAHqrFRRzAKBggqhkjOPQQDAwNnADBkAjAE3A3U
3jSZCpwfqOHBdlxi9ASgKTU+wg0qw3FqtfQ31OwLYFdxh0MlNk/HwkjRSWgCMFbQ
vMkXEPvNvv4t30K6xtpG26qmZ+6OiISBIIXMljWnsiYR1gyZnTzIg3AQSw4Vmw==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEFzCCAv+gAwIBAgIQB/LzXIeod6967+lHmTUlvTANBgkqhkiG9w0BAQwFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0yMTA0MTQwMDAwMDBaFw0zMTA0MTMyMzU5NTlaMFYxCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxMDAuBgNVBAMTJ0RpZ2lDZXJ0IFRMUyBI
eWJyaWQgRUNDIFNIQTM4NCAyMDIwIENBMTB2MBAGByqGSM49AgEGBSuBBAAiA2IA
BMEbxppbmNmkKaDp1AS12+umsmxVwP/tmMZJLwYnUcu/cMEFesOxnYeJuq20ExfJ
qLSDyLiQ0cx0NTY8g3KwtdD3ImnI8YDEe0CPz2iHJlw5ifFNkU3aiYvkA8ND5b8v
c6OCAYIwggF+MBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFAq8CCkXjKU5
bXoOzjPHLrPt+8N6MB8GA1UdIwQYMBaAFAPeUDVW0Uy7ZvCj4hsbw5eyPdFVMA4G
A1UdDwEB/wQEAwIBhjAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwdgYI
KwYBBQUHAQEEajBoMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdpY2VydC5j
b20wQAYIKwYBBQUHMAKGNGh0dHA6Ly9jYWNlcnRzLmRpZ2ljZXJ0LmNvbS9EaWdp
Q2VydEdsb2JhbFJvb3RDQS5jcnQwQgYDVR0fBDswOTA3oDWgM4YxaHR0cDovL2Ny
bDMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0R2xvYmFsUm9vdENBLmNybDA9BgNVHSAE
NjA0MAsGCWCGSAGG/WwCATAHBgVngQwBATAIBgZngQwBAgEwCAYGZ4EMAQICMAgG
BmeBDAECAzANBgkqhkiG9w0BAQwFAAOCAQEAR1mBf9QbH7Bx9phdGLqYR5iwfnYr
6v8ai6wms0KNMeZK6BnQ79oU59cUkqGS8qcuLa/7Hfb7U7CKP/zYFgrpsC62pQsY
kDUmotr2qLcy/JUjS8ZFucTP5Hzu5sn4kL1y45nDHQsFfGqXbbKrAjbYwrwsAZI/
BKOLdRHHuSm8EdCGupK8JvllyDfNJvaGEwwEqonleLHBTnm8dqMLUeTF0J5q/hos
Vq4GNiejcxwIfZMy0MJEGdqN9A57HSgDKwmKdsp33Id6rHtSJlWncg+d0ohP/rEh
xRqhqjn1VtvChMQ1H3Dau0bwhr9kAMQ+959GG50jBbl9s08PqUU643QwmA==
-----END CERTIFICATE-----		
`
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

	chain := [][]byte{}
	rest := []byte(GHWebPublicCert)
	var decodedPEM *pem.Block
	for {
		decodedPEM, rest = pem.Decode(rest)
		if decodedPEM == nil || rest == nil {
			break
		}
		chain = append(chain, decodedPEM.Bytes)
	}

	config := &verify.CertVerifierConfig{
		TrustSystemRoots: true,
		Time: func() time.Time {
			currentTime, _ := time.Parse(time.DateOnly, "2023-07-14")

			return currentTime
		},
	}

	// it should pass with trusted system roots
	_, err := verify.NewCertVerifier(config).VerifyCertificate(chain)
	if err != nil {
		t.Fatal(err.Error(), errors.GetDetails(err))
	}

	// it should fail without trusted system roots
	config.TrustSystemRoots = false
	if _, err = verify.NewCertVerifier(config).VerifyCertificate(chain); err == nil {
		t.Fatal(errors.New("cert should have been invalid"))
	}
}
