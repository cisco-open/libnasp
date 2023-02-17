// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path"
	"time"

	"emperror.dev/errors"
)

const (
	// caCertificateFileName is the name of the PEM encoded certificate
	// authority  certificate file.
	caCertificateFileName = "ca.crt"

	// caPrivateKeyFileName is the name of the PEM encoded certificate authority
	// private key file.
	caPrivateKeyFileName = "ca.key"

	// serverCertificateFileName is the name of the PEM encoded server
	// certificate file.
	serverCertificateFileName = "tls.crt"

	// serverPrivateKeyFileName is the name of the PEM encoded server private
	// key file.
	serverPrivateKeyFileName = "tls.key"
)

// Certificate describes an x509 certificate.
type Certificate struct {
	// CACertificate is the PEM encoded certificate authority certificate.
	CACertificate []byte

	// CAPrivateKey is the PEM encoded certificate authority private key.
	CAPrivateKey []byte

	// ServerCertificate is the PEM encoded server certificate.
	ServerCertificate []byte

	// ServerPrivateKey is the PEM encoded server private key.
	ServerPrivateKey []byte
}

// NewCertificate returns a self signed certificate generated for the specified
// DNS names or alternatively an error.
func NewCertificate(dnsNames []string) (*Certificate, error) {
	certificate := new(Certificate)

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "generating CA private key failed")
	}

	certificate.CAPrivateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})
	if certificate.CAPrivateKey == nil {
		return nil, errors.Wrap(err, "encoding CA private key failed")
	}

	commonName := "unspecified"
	if len(dnsNames) > 0 {
		commonName = dnsNames[0]
	}

	caCertificateConfiguration := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caCertificate, err := x509.CreateCertificate(
		rand.Reader,
		caCertificateConfiguration,
		caCertificateConfiguration,
		&caPrivateKey.PublicKey,
		caPrivateKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "creating CA certificate failed")
	}

	certificate.CACertificate = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertificate,
	})
	if certificate.CACertificate == nil {
		return nil, errors.Wrap(err, "encoding CA certificate failed")
	}

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, errors.Wrap(err, "generating server private key failed")
	}

	certificate.ServerPrivateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})
	if certificate.ServerPrivateKey == nil {
		return nil, errors.Wrap(err, "encoding server private key failed")
	}

	serverCertificateConfiguration := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// sign the server cert
	serverCertificate, err := x509.CreateCertificate(
		rand.Reader,
		serverCertificateConfiguration,
		caCertificateConfiguration,
		&serverPrivateKey.PublicKey,
		caPrivateKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "creating server certificate failed")
	}

	certificate.ServerCertificate = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertificate,
	})
	if certificate.ServerCertificate == nil {
		return nil, errors.Wrap(err, "encoding server certificate failed")
	}

	return certificate, nil
}

// NewCertificateFromDirectory returns a certificate by reading the
// corresponding certificate files from the specified directory path  or
// alternatively an error.
func NewCertificateFromDirectory(directoryPath string) (*Certificate, error) {
	fileContents := map[string][]byte{
		caCertificateFileName:     nil,
		caPrivateKeyFileName:      nil,
		serverCertificateFileName: nil,
		serverPrivateKeyFileName:  nil,
	}

	errs := make([]error, 0, len(fileContents))
	var err error

	for fileName := range fileContents {
		filePath := path.Join(directoryPath, fileName)

		fileContents[fileName], err = os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, errors.WrapWithDetails(err, "reading certificate file failed", "filePath", filePath))

			continue
		}
	}

	if len(errs) != 0 {
		return nil, errors.Combine(errs...)
	}

	return NewCertificateFromFileContents(fileContents), nil
}

// NewCertificateFromFileContents returns a certificate by mapping the specified
// certificate file name keys and content values to the certificate fields.
func NewCertificateFromFileContents(fileContents map[string][]byte) *Certificate {
	return &Certificate{
		CACertificate:     fileContents[caCertificateFileName],
		CAPrivateKey:      fileContents[caPrivateKeyFileName],
		ServerCertificate: fileContents[serverCertificateFileName],
		ServerPrivateKey:  fileContents[serverPrivateKeyFileName],
	}
}

// Verify returns true if
//
// 1. the server key pair is valid,
//
// and the server certificate
//
// 2. is good for the specified DNS name,
//
// 3. is signed by the certificate authority certificate, and
//
// 4. is valid for the desired period of time,
//
// otherwise returns false.
func (certificate *Certificate) Verify(dnsName string, checkTime time.Time) bool {
	if certificate == nil ||
		certificate.CACertificate == nil ||
		certificate.ServerPrivateKey == nil ||
		certificate.ServerCertificate == nil {
		return false
	}

	_, err := tls.X509KeyPair(certificate.ServerCertificate, certificate.ServerPrivateKey)
	if err != nil {
		return false
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certificate.CACertificate) {
		return false
	}

	block, _ := pem.Decode(certificate.ServerCertificate)
	if block == nil {
		return false
	}

	x509Certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	_, err = x509Certificate.Verify(x509.VerifyOptions{
		DNSName:     dnsName,
		Roots:       pool,
		CurrentTime: checkTime,
	})

	return err == nil
}

// Write writes the content of the certificate files held by the certificate to
// files under the specified directory path using the standard certificate file
// names or returns an error.
func (certificate *Certificate) Write(directoryPath string) error {
	if certificate == nil {
		return errors.Errorf("invalid nil certificate")
	}

	err := os.MkdirAll(directoryPath, 0o755)
	if err != nil {
		return errors.WrapWithDetails(err, "creating directory path failed", "directoryPath", directoryPath)
	}

	for file, content := range map[string][]byte{
		caCertificateFileName:     certificate.CACertificate,
		caPrivateKeyFileName:      certificate.CAPrivateKey,
		serverCertificateFileName: certificate.ServerCertificate,
		serverPrivateKeyFileName:  certificate.ServerPrivateKey,
	} {
		filePath := path.Join(directoryPath, file)

		err = os.WriteFile(filePath, content, 0o600)
		if err != nil {
			return errors.WrapWithDetails(err, "writing certificate file failed", "filePath", filePath)
		}
	}

	return nil
}
