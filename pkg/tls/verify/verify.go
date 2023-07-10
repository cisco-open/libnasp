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
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"emperror.dev/errors"

	"github.com/cisco-open/nasp/pkg/util"
)

type CertVerifier interface {
	VerifyCertificate(chain [][]byte) ([]*x509.Certificate, error)
}

type CertVerifierConfig struct {
	// Time returns the current time as the number of seconds since the epoch.
	// If Time is nil, TLS uses time.Now.
	Time func() time.Time

	// RootCAs defines the set of root certificate authorities
	// that clients use when verifying server certificates.
	// If RootCAs is nil, TLS uses the host's root CA set.
	Roots *x509.CertPool

	// TrustSystemRoots whether the verifier trust system roots.
	TrustSystemRoots bool

	// An optional list of base64-encoded SHA-256 hashes. If specified, the verifier will verify that the
	// SHA-256 of the DER-encoded Subject Public Key Information (SPKI) of the presented certificate
	// matches one of the specified values.
	//
	// A base64-encoded SHA-256 of the Subject Public Key Information (SPKI) of the certificate
	// can be generated with the following command:
	//
	// .. code-block:: bash
	//
	//   $ openssl x509 -in path/to/client.crt -noout -pubkey
	//     | openssl pkey -pubin -outform DER
	//     | openssl dgst -sha256 -binary
	//     | openssl enc -base64
	//   NvqYIYSbgK2vCJpQhObf77vv+bQWtc5ek5RIOwPiC9A=
	//
	// This is the format used in HTTP Public Key Pinning.
	MatchSPKIHash []string

	// An optional list of hex-encoded SHA-256 hashes. If specified, the verifier will verify that
	// the SHA-256 of the DER-encoded presented certificate matches one of the specified values.
	//
	// A hex-encoded SHA-256 of the certificate can be generated with the following command:
	//
	// .. code-block:: bash
	//
	//   $ openssl x509 -in path/to/client.crt -outform DER | openssl dgst -sha256 | cut -d" " -f2
	//   df6ff72fe9116521268f6f2dd4966f51df479883fe7037b39f75916ac3049d1a
	MatchCertificateHash []string

	// An optional list of Subject Alternative name matchers. If specified, the verifier will verify that the
	// Subject Alternative Name of the presented certificate matches one of the specified matchers.
	// The matching uses "any" semantics, that is to say, the SAN is verified if at least one matcher is
	// matched.
	MatchTypedSAN []SANMatcher
}

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

type certVerifier struct {
	config *CertVerifierConfig
}

func NewCertVerifier(config *CertVerifierConfig) CertVerifier {
	return &certVerifier{
		config: config,
	}
}

// Verify parses and verifies the provided chain.
func (v *certVerifier) VerifyCertificate(chain [][]byte) ([]*x509.Certificate, error) {
	chains, err := v.verify(chain)
	if err != nil {
		return nil, err
	}

	if len(chains) < 1 || len(chains[0]) < 1 {
		return nil, errors.New("cert verification failed")
	}

	for _, f := range []func(*x509.Certificate) error{
		v.verifyCertificateHash,
		v.verifySPKIHash,
		v.verifyTypedSANs,
	} {
		if err := f(chains[0][0]); err != nil {
			return nil, err
		}
	}

	return chains[0], nil
}

func (v *certVerifier) verify(chain [][]byte) ([][]*x509.Certificate, error) {
	certs := make([]*x509.Certificate, len(chain))
	for i, asn1Data := range chain {
		cert, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			return nil, errors.WrapIf(err, "failed to parse certificate")
		}
		certs[i] = cert
	}

	verifiedChains := [][]*x509.Certificate{}

	opts := x509.VerifyOptions{
		CurrentTime:   v.time(),
		Intermediates: x509.NewCertPool(),
	}

	if !v.config.TrustSystemRoots && v.config.Roots == nil {
		v.config.Roots = x509.NewCertPool()
	}

	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	var err error

	if v.config.Roots != nil {
		copts := opts
		copts.Roots = v.config.Roots
		chains, err := certs[0].Verify(copts)
		if !v.config.TrustSystemRoots {
			return chains, errors.WrapIf(err, "cert verification failed")
		}

		verifiedChains = append(verifiedChains, chains...)
	}

	if len(verifiedChains) == 0 {
		verifiedChains, err = certs[0].Verify(opts)
		if err != nil {
			return nil, errors.WrapIf(err, "cert verification failed")
		}
	}

	return verifiedChains, nil
}

func (v *certVerifier) verifySPKIHash(cert *x509.Certificate) error {
	if len(v.config.MatchSPKIHash) < 1 {
		return nil
	}

	spkiHash, err := GetSPKIHash(cert)
	if err != nil {
		return errors.WrapIf(err, "could not get SPKI hash")
	}

	for _, hash := range v.config.MatchSPKIHash {
		if hash == spkiHash {
			return nil
		}
	}

	return errors.NewWithDetails("invalid SPKI hash", "hash", spkiHash)
}

func (v *certVerifier) verifyCertificateHash(cert *x509.Certificate) error {
	if len(v.config.MatchCertificateHash) < 1 {
		return nil
	}

	certHash := fmt.Sprintf("%x", sha256.Sum256(cert.Raw))

	for _, hash := range v.config.MatchCertificateHash {
		if hash == certHash {
			return nil
		}
	}

	return errors.NewWithDetails("invalid certificate hash", "hash", certHash)
}

func (v *certVerifier) verifyTypedSANs(cert *x509.Certificate) error {
	if len(v.config.MatchTypedSAN) < 1 {
		return nil
	}

	checkedSANs := []string{}

	for _, match := range v.config.MatchTypedSAN {
		switch match.Type {
		case EmailSANType:
			for _, v := range cert.EmailAddresses {
				if match.Matcher.Match(v) {
					return nil
				}

				checkedSANs = append(checkedSANs, v)
			}
		case URISANType:
			for _, v := range cert.URIs {
				if match.Matcher.Match(v.String()) {
					return nil
				}

				checkedSANs = append(checkedSANs, v.String())
			}
		case DNSSANType:
			for _, v := range cert.DNSNames {
				if match.Matcher.Match(v) {
					return nil
				}

				checkedSANs = append(checkedSANs, v)
			}
		case IPSANType:
			for _, v := range cert.IPAddresses {
				if match.Matcher.Match(v.String()) {
					return nil
				}

				checkedSANs = append(checkedSANs, v.String())
			}
		}
	}

	return errors.NewWithDetails("invalid SAN", "matches", v.config.MatchTypedSAN, "checkedSANs", checkedSANs)
}

func (v *certVerifier) time() time.Time {
	t := v.config.Time
	if t == nil {
		t = time.Now
	}
	return t()
}
