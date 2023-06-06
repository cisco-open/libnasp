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

package network

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidCertificate = errors.New("invalid tls certificate")

type Certificate interface {
	GetCertificate() *x509.Certificate
	GetSHA256Digest() string
	GetSubject() string
	GetFirstURI() string
	GetFirstDNSName() string
	GetFirstIP() string
	String() string
}

type certificate struct {
	*x509.Certificate
}

func NewCertificate(cert *x509.Certificate) Certificate {
	return &certificate{
		Certificate: cert,
	}
}

func ParseTLSCertificate(cert *tls.Certificate) (Certificate, error) {
	if cert == nil || len(cert.Certificate) < 1 {
		return nil, ErrInvalidCertificate
	}

	x509cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	return NewCertificate(x509cert), nil
}

func (c *certificate) GetCertificate() *x509.Certificate {
	return c.Certificate
}

func (c *certificate) GetSHA256Digest() string {
	return fmt.Sprintf("%x", sha256.Sum256(c.Raw))
}

func (c *certificate) GetSubject() string {
	return c.Certificate.Subject.String()
}

func (c *certificate) GetFirstURI() string {
	if len(c.URIs) < 1 {
		return ""
	}

	return c.URIs[0].String()
}

func (c *certificate) GetFirstDNSName() string {
	if len(c.DNSNames) < 1 {
		return ""
	}

	return c.DNSNames[0]
}

func (c *certificate) GetFirstIP() string {
	if len(c.IPAddresses) < 1 {
		return ""
	}

	return c.IPAddresses[0].String()
}

func (c *certificate) String() string {
	p := make([]string, 0)

	if s := c.GetSubject(); s != "" {
		p = append(p, fmt.Sprintf("Subject=%s", s))
	}

	if u := c.URIs; len(u) > 0 {
		uris := []string{}
		for _, uri := range u {
			uris = append(uris, uri.String())
		}
		p = append(p, fmt.Sprintf("URIs=%s", strings.Join(uris, ",")))
	}

	if d := c.DNSNames; len(d) > 0 {
		p = append(p, fmt.Sprintf("DNSNames=%s", strings.Join(d, ",")))
	}

	if i := c.IPAddresses; len(i) > 0 {
		ips := []string{}
		for _, ip := range i {
			ips = append(ips, ip.String())
		}
		p = append(p, fmt.Sprintf("IPs=%s", strings.Join(ips, ",")))
	}

	return strings.Join(p, " ")
}
