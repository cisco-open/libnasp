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

package selfsigned

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"strings"
	"time"

	"emperror.dev/errors"

	"wwwin-github.cisco.com/eti/nasp/pkg/ca"
)

type selfsigned struct {
	keySize int

	ca           *x509.Certificate
	caPrivKey    *rsa.PrivateKey
	caPrivKeyPEM []byte
	caPEM        []byte
}

type ClientOption func(c *selfsigned)

func WithKeySize(keySize int) ClientOption {
	return func(c *selfsigned) {
		c.keySize = keySize
	}
}

func NewSelfSignedCAClient(opts ...ClientOption) (ca.Client, error) {
	c := &selfsigned{
		keySize: 2048,
	}
	if err := c.createCACert(); err != nil {
		return nil, errors.WrapIf(err, "could not create client")
	}

	for _, o := range opts {
		o(c)
	}

	return c, nil
}

func (c *selfsigned) createCACert() (err error) {
	c.ca = &x509.Certificate{
		SerialNumber:          big.NewInt(2022),
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	c.caPrivKey, err = rsa.GenerateKey(rand.Reader, c.keySize)
	if err != nil {
		return errors.WrapIf(err, "could not generate RSA key for the CA")
	}

	c.caPrivKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(c.caPrivKey),
	})

	caBytes, err := x509.CreateCertificate(rand.Reader, c.ca, c.ca, &c.caPrivKey.PublicKey, c.caPrivKey)
	if err != nil {
		return errors.WrapIf(err, "could not create CA certificate")
	}

	c.caPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	return nil
}

func (c *selfsigned) GetCAEndpoint() string {
	return ""
}

func (c *selfsigned) GetCAPem() []byte {
	return c.caPEM
}

func (c *selfsigned) GetCertificate(hostname string, ttl time.Duration) (ca.Certificate, error) {
	serialNum, err := c.genSerialNum()
	if err != nil {
		return nil, errors.WrapIf(err, "could not generate serial number")
	}

	x509cert := &x509.Certificate{
		SerialNumber: serialNum,
		Subject:      pkix.Name{},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(ttl),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		IsCA:         false,
	}

	if strings.Contains(hostname, "://") {
		u, err := url.Parse(hostname)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "could not parse URL", "url", hostname)
		}
		x509cert.URIs = []*url.URL{u}
	} else {
		x509cert.DNSNames = []string{hostname}
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, c.keySize)
	if err != nil {
		return nil, errors.WrapIf(err, "could not generate RSA key for the certificate")
	}

	certPrivKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	certBytes, err := x509.CreateCertificate(rand.Reader, x509cert, c.ca, &certPrivKey.PublicKey, c.caPrivKey)
	if err != nil {
		return nil, errors.WrapIf(err, "could not create certificate")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	tlsCertificate, err := tls.X509KeyPair(certPEM, certPrivKeyPEM)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get tls certiciate from PEMs")
	}

	cert := &ca.CertificateContainer{
		PrivateKeyPEM:  certPrivKeyPEM,
		CertificatePEM: certPEM,
		TLSCertificate: &tlsCertificate,
	}

	return cert, nil
}

func (c *selfsigned) genSerialNum() (*big.Int, error) {
	serialNumLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNum, err := rand.Int(rand.Reader, serialNumLimit)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return serialNum, nil
}
