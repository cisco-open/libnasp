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
	"crypto/tls"
	"net"
	"strings"
)

func CreateTLSConn(conn net.Conn, tlsConfig *tls.Config, f func(net.Conn, *tls.Config) *tls.Conn) *tls.Conn {
	tlsConn := f(conn, tlsConfig)

	if co, ok := conn.(interface {
		SetTLSConn(*tls.Conn)
	}); ok {
		co.SetTLSConn(tlsConn)
	}

	return tlsConn
}

func CreateTLSServerConn(conn net.Conn, tlsConfig *tls.Config) *tls.Conn {
	return CreateTLSConn(conn, tlsConfig, tls.Server)
}

func CreateTLSClientConn(conn net.Conn, tlsConfig *tls.Config) *tls.Conn {
	return CreateTLSConn(conn, tlsConfig, tls.Client)
}

func WrapTLSConfig(config *tls.Config) *tls.Config { //nolint:gocognit,funlen
	if config == nil {
		return nil
	}

	tlsConfig := config.Clone()

	// server
	if len(config.Certificates) > 0 || config.GetCertificate != nil { //nolint:nestif
		tlsConfig.Certificates = nil
		tlsConfig.SessionTicketsDisabled = true
		tlsConfig.GetCertificate = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			getCert := func(config *tls.Config) (*tls.Certificate, error) {
				var err error
				cert := new(tls.Certificate)

				if config.GetCertificate != nil {
					cert, err = config.GetCertificate(chi)
					if err != nil {
						return nil, err
					}

					return cert, nil
				}

				if len(config.Certificates) == 1 {
					cert = &config.Certificates[0]

					return cert, nil
				}

				if config.NameToCertificate != nil {
					name := strings.ToLower(chi.ServerName)
					if cert, ok := config.NameToCertificate[name]; ok {
						return cert, nil
					}
					if len(name) > 0 {
						labels := strings.Split(name, ".")
						labels[0] = "*"
						wildcardName := strings.Join(labels, ".")
						if cert, ok := config.NameToCertificate[wildcardName]; ok {
							return cert, nil
						}
					}
				}

				for _, chain := range config.Certificates {
					chain := chain
					if err := chi.SupportsCertificate(&chain); err == nil {
						cert = &chain

						return cert, nil
					}
				}

				return cert, nil
			}

			cert, err := getCert(config)
			if err != nil {
				return nil, err
			}

			if cert != nil {
				if s, ok := ConnectionStateFromContext(chi.Context()); ok {
					s.SetLocalCertificate(*cert)
				}
				if s, ok := ConnectionStateFromNetConn(chi.Conn); ok {
					s.SetLocalCertificate(*cert)
				}
			}

			return cert, nil
		}
	}

	// client
	if len(config.Certificates) > 0 || config.GetClientCertificate != nil { //nolint:nestif
		tlsConfig.Certificates = nil
		tlsConfig.GetClientCertificate = func(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			getCert := func(config *tls.Config) (*tls.Certificate, error) {
				var err error
				cert := new(tls.Certificate)

				if config.GetClientCertificate != nil {
					cert, err = config.GetClientCertificate(cri)
					if err != nil {
						return nil, err
					}
				}

				for _, chain := range config.Certificates {
					chain := chain
					if err := cri.SupportsCertificate(&chain); err == nil {
						cert = &chain

						return cert, nil
					}
				}

				return cert, nil
			}

			cert, err := getCert(config)
			if err != nil {
				return nil, err
			}

			if cert != nil {
				if s, ok := ConnectionStateHolderFromContext(cri.Context()); ok {
					s.Get().SetLocalCertificate(*cert)
				}
			}

			return cert, nil
		}
	}

	return tlsConfig
}
