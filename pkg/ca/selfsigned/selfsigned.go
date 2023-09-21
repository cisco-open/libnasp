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
	"net"
	"net/url"
	"time"

	"emperror.dev/errors"
	"go.uber.org/atomic"

	"github.com/cisco-open/libnasp/pkg/ca"
	"github.com/cisco-open/libnasp/pkg/tls/verify"
)

type SelfsignedCAClient interface {
	ca.Client

	GetCertificateWithSigner(subject pkix.Name, SANs []string, ttl time.Duration, signer *Certificate) (*Certificate, error)
	CreateIntermediateCA(subject pkix.Name) (*Certificate, error)
	CreateCertificate(template *x509.Certificate, parent *Certificate) (*Certificate, error)
}

type selfsigned struct {
	caSubject pkix.Name
	keySize   int

	serial *atomic.Int64

	ca *Certificate
}

type Certificate struct {
	Certificate    *x509.Certificate
	CertificateDER []byte
	CertificatePEM []byte
	PrivateKey     *rsa.PrivateKey
	PrivateKeyPEM  []byte
	PublicKey      *rsa.PublicKey
	PublicKeyPEM   []byte
}

type ClientOption func(c *selfsigned)

func WithKeySize(keySize int) ClientOption {
	return func(c *selfsigned) {
		c.keySize = keySize
	}
}

func WithCASubject(subject pkix.Name) ClientOption {
	return func(c *selfsigned) {
		c.caSubject = subject
	}
}

func NewSelfSignedCAClient(opts ...ClientOption) (SelfsignedCAClient, error) {
	c := &selfsigned{
		caSubject: pkix.Name{},
		serial:    atomic.NewInt64(0),
		keySize:   2048,
	}

	if err := c.createRootCA(); err != nil {
		return nil, errors.WrapIf(err, "could not create client")
	}

	for _, o := range opts {
		o(c)
	}

	return c, nil
}

func (c *selfsigned) GetCAEndpoint() string {
	return ""
}

func (c *selfsigned) GetCAPem() []byte {
	return c.ca.CertificatePEM
}

func (c *selfsigned) GetCertificateWithSigner(subject pkix.Name, sans []string, ttl time.Duration, signer *Certificate) (*Certificate, error) {
	template := &x509.Certificate{
		SerialNumber: c.getNextSerial(),
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(ttl),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		IsCA:         false,
	}

	for _, san := range sans {
		sanType := verify.GetSANTypeFromSAN(san)
		switch sanType {
		case verify.URISANType:
			u, err := url.Parse(san)
			if err != nil {
				return nil, errors.WrapIfWithDetails(err, "could not parse URL", "url", san)
			}
			template.URIs = []*url.URL{u}
		case verify.EmailSANType:
			template.EmailAddresses = append(template.EmailAddresses, san)
			continue
		case verify.IPSANType:
			if ip := net.ParseIP(san); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			}
		case verify.DNSSANType:
			template.DNSNames = append(template.DNSNames, san)
		}
	}

	if signer == nil {
		signer = c.ca
	}

	return c.CreateCertificate(template, signer)
}

func (c *selfsigned) GetCertificate(hostname string, ttl time.Duration) (ca.Certificate, error) {
	cert, err := c.GetCertificateWithSigner(pkix.Name{}, []string{hostname}, ttl, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	tlsCertificate, err := tls.X509KeyPair(cert.CertificatePEM, cert.PrivateKeyPEM)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get tls certiciate from PEMs")
	}

	return &ca.CertificateContainer{
		PrivateKeyPEM:  cert.PrivateKeyPEM,
		CertificatePEM: cert.CertificatePEM,
		TLSCertificate: &tlsCertificate,
	}, nil
}

func (c *selfsigned) CreateIntermediateCA(subject pkix.Name) (*Certificate, error) {
	template := &x509.Certificate{
		SerialNumber:          c.getNextSerial(),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	return c.CreateCertificate(template, c.ca)
}

func (c *selfsigned) createRootCA() (err error) {
	template := &x509.Certificate{
		SerialNumber: c.getNextSerial(),
		Subject: pkix.Name{
			CommonName: "ACME CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	c.ca, err = c.CreateCertificate(template, nil)

	return err
}

func (c *selfsigned) CreateCertificate(template *x509.Certificate, parent *Certificate) (*Certificate, error) {
	cert := &Certificate{}

	var err error
	cert.PrivateKey, err = rsa.GenerateKey(rand.Reader, c.keySize)
	if err != nil {
		return nil, errors.WrapIf(err, "could not generate RSA key for the CA")
	}
	cert.PublicKey = &cert.PrivateKey.PublicKey

	if parent == nil {
		parent = cert
		parent.Certificate = template
	}

	cert.CertificateDER, err = x509.CreateCertificate(rand.Reader, template, parent.Certificate, cert.PublicKey, parent.PrivateKey)
	if err != nil {
		return nil, errors.WrapIf(err, "could not create CA certificate")
	}

	cert.Certificate, err = x509.ParseCertificate(cert.CertificateDER)
	if err != nil {
		return nil, errors.WrapIf(err, "could not parse created certificate")
	}

	cert.PrivateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(cert.PrivateKey),
	})

	cert.PublicKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(cert.PublicKey),
	})

	cert.CertificatePEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.CertificateDER,
	})

	return cert, nil
}

func (c *selfsigned) getNextSerial() *big.Int {
	c.serial.Add(1)

	return big.NewInt(c.serial.Load())
}
