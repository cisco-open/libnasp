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

package ca

import (
	"crypto/tls"
	"time"
)

type Certificate interface {
	GetTLSCertificate() *tls.Certificate
	GetCertificatePEM() []byte
	GetPrivateKeyPEM() []byte
}

type Client interface {
	GetCAEndpoint() string
	GetCAPem() []byte
	GetCertificate(hostname string, ttl time.Duration) (Certificate, error)
}

type CertificateContainer struct {
	PrivateKeyPEM  []byte
	CertificatePEM []byte
	TLSCertificate *tls.Certificate
}

func (c *CertificateContainer) GetTLSCertificate() *tls.Certificate {
	return c.TLSCertificate
}

func (c *CertificateContainer) GetPrivateKeyPEM() []byte {
	return c.PrivateKeyPEM
}

func (c *CertificateContainer) GetCertificatePEM() []byte {
	return c.CertificatePEM
}
