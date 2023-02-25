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

package istio_ca

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"istio.io/istio/pkg/security"
	pkiutil "istio.io/istio/security/pkg/pki/util"

	"github.com/cisco-open/nasp/pkg/ca"
)

type IstioCAClient struct {
	config       IstioCAClientConfig
	certificates map[string]ca.Certificate
	logger       logr.Logger

	mu sync.Mutex
}

type IstioCAClientConfig struct {
	CAEndpoint    string
	CAEndpointSAN string
	ClusterID     string
	Token         []byte
	CApem         []byte
	Revision      string
	MeshID        string
	TrustDomain   string
}

func NewIstioCAClient(config IstioCAClientConfig, logger logr.Logger) ca.Client {
	return &IstioCAClient{
		config:       config,
		logger:       logger,
		certificates: map[string]ca.Certificate{},
	}
}

func (c *IstioCAClient) GetCAEndpoint() string {
	return c.config.CAEndpoint
}

func (c *IstioCAClient) GetCAPem() []byte {
	return c.config.CApem
}

func (c *IstioCAClient) GetConfig() IstioCAClientConfig {
	return c.config
}

func (c *IstioCAClient) GetCertificate(hostname string, ttl time.Duration) (ca.Certificate, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.certificates[hostname] != nil {
		certificate, err := x509.ParseCertificate(c.certificates[hostname].GetTLSCertificate().Certificate[0])
		if err != nil {
			return nil, err
		}

		c.logger.V(1).Info("certificate exists", "expiry", certificate.NotAfter)

		if time.Since(certificate.NotAfter).Seconds() < 0 {
			return c.certificates[hostname], nil
		}
	}

	endpointSAN := c.config.CAEndpointSAN
	if endpointSAN == "" {
		if host, _, err := net.SplitHostPort(c.config.CAEndpoint); err != nil {
			endpointSAN = host
		}
	}

	caClient, err := NewCitadelClient(&security.Options{
		CAEndpoint:    c.config.CAEndpoint,
		CAEndpointSAN: endpointSAN,
		ClusterID:     c.config.ClusterID,
		CredFetcher: CredFetcher{
			Token: string(c.config.Token),
		},
	}, &TLSOptions{
		RootCertPEM: c.config.CApem,
	}, c.logger)
	if err != nil {
		return nil, err
	}

	options := pkiutil.CertOptions{
		Host:       hostname,
		RSAKeySize: 2048,
	}

	csr, key, err := pkiutil.GenCSR(options)
	if err != nil {
		return nil, err
	}

	cert := new(bytes.Buffer)
	certChainPEM, err := caClient.CSRSign(csr, int64(ttl.Seconds()))
	if err != nil {
		return nil, err
	}
	for _, pem := range certChainPEM {
		cert.WriteString(pem)
	}

	certificate, err := tls.X509KeyPair(cert.Bytes(), key)
	if err != nil {
		return nil, err
	}

	c.certificates[hostname] = &ca.CertificateContainer{
		PrivateKeyPEM:  key,
		CertificatePEM: cert.Bytes(),
		TLSCertificate: &certificate,
	}

	return c.certificates[hostname], err
}

type CredFetcher struct {
	Token string
}

func (f CredFetcher) GetPlatformCredential() (string, error) {
	return f.Token, nil
}

func (f CredFetcher) GetType() string {
	return ""
}

func (f CredFetcher) GetIdentityProvider() string {
	return "static"
}

func (f CredFetcher) Stop() {}
