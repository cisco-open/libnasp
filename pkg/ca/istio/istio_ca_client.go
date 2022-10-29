// Copyright 2017 Istio Authors
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

//nolint:all
package istio_ca

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/klog/v2"

	pb "istio.io/api/security/v1alpha1"
	"istio.io/istio/pkg/security"
	"istio.io/istio/security/pkg/nodeagent/caclient"
)

const (
	bearerTokenPrefix = "Bearer "
)

type CitadelClient struct {
	// It means enable tls connection to Citadel if this is not nil.
	tlsOpts   *TLSOptions
	client    pb.IstioCertificateServiceClient
	conn      *grpc.ClientConn
	provider  *caclient.TokenProvider
	opts      *security.Options
	usingMtls *atomic.Bool
	logger    logr.Logger
}

type TLSOptions struct {
	RootCertPEM []byte
	KeyPEM      []byte
	CertPEM     []byte
}

func init() {
	// clear zaplogger set by istio log package
	klog.ClearLogger()
}

// NewCitadelClient create a CA client for Citadel.
func NewCitadelClient(opts *security.Options, tlsOpts *TLSOptions, logger logr.Logger) (*CitadelClient, error) {
	c := &CitadelClient{
		tlsOpts:   tlsOpts,
		opts:      opts,
		provider:  caclient.NewCATokenProvider(opts),
		usingMtls: atomic.NewBool(false),
		logger:    logger.WithName("citadel-client"),
	}

	conn, err := c.buildConnection()
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to connect to endpoint", "endpoint", opts.CAEndpoint)
	}
	c.conn = conn
	c.client = pb.NewIstioCertificateServiceClient(conn)
	return c, nil
}

func (c *CitadelClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// CSRSign calls Citadel to sign a CSR.
func (c *CitadelClient) CSRSign(csrPEM []byte, certValidTTLInSec int64) ([]string, error) {
	crMetaStruct := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			security.CertSigner: {
				Kind: &structpb.Value_StringValue{StringValue: c.opts.CertSigner},
			},
		},
	}
	req := &pb.IstioCertificateRequest{
		Csr:              string(csrPEM),
		ValidityDuration: certValidTTLInSec,
		Metadata:         crMetaStruct,
	}

	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("ClusterID", c.opts.ClusterID))
	resp, err := c.client.CreateCertificate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %v", err)
	}

	if len(resp.CertChain) <= 1 {
		return nil, errors.New("invalid empty CertChain")
	}

	return resp.CertChain, nil
}

func (c *CitadelClient) getTLSDialOption() (grpc.DialOption, error) {
	certPool, err := c.getRootCertificate(c.tlsOpts.RootCertPEM)
	if err != nil {
		return nil, err
	}
	config := tls.Config{
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			var certificate tls.Certificate
			key, cert := c.tlsOpts.KeyPEM, c.tlsOpts.CertPEM
			if cert != nil {
				var isExpired bool
				isExpired, err = c.isCertExpired(cert)
				if err != nil {
					err = errors.WrapIf(err, "cannot parse the cert chain, using token instead")
					c.logger.V(1).Info(err.Error())
					return &certificate, nil
				}
				if isExpired {
					c.logger.V(1).Info("cert expired, using token instead")
					return &certificate, nil
				}

				// Load the certificate from disk
				certificate, err = tls.X509KeyPair(cert, key)
				if err != nil {
					return nil, err
				}
				c.usingMtls.Store(true)
			}
			return &certificate, nil
		},
		RootCAs: certPool,
	}

	// For debugging on localhost (with port forward)
	// TODO: remove once istiod is stable and we have a way to validate JWTs locally
	if strings.Contains(c.opts.CAEndpoint, "localhost") {
		config.ServerName = "istiod.istio-system.svc"
	}
	if c.opts.CAEndpointSAN != "" {
		config.ServerName = c.opts.CAEndpointSAN
	}

	transportCreds := credentials.NewTLS(&config)
	return grpc.WithTransportCredentials(transportCreds), nil
}

func (c *CitadelClient) getRootCertificate(rootCertPEM []byte) (*x509.CertPool, error) {
	if rootCertPEM == nil {
		// No explicit certificate - assume the citadel-compatible server uses a public cert
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		c.logger.Info("citadel client using system cert")
		return certPool, nil
	}

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(rootCertPEM))
	if !ok {
		return nil, fmt.Errorf("failed to append certificates")
	}
	c.logger.Info("citadel client using custom root certs")
	return certPool, nil
}

func (c *CitadelClient) isCertExpired(certPEMBlock []byte) (bool, error) {
	var err error
	var certDERBlock *pem.Block
	certDERBlock, _ = pem.Decode(certPEMBlock)
	if certDERBlock == nil {
		return true, fmt.Errorf("failed to decode certificate")
	}
	x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return true, fmt.Errorf("failed to parse the cert, err is %v", err)
	}
	return x509Cert.NotAfter.Before(time.Now()), nil
}

func (c *CitadelClient) buildConnection() (*grpc.ClientConn, error) {
	var opts grpc.DialOption
	var err error
	// CA tls disabled
	if c.tlsOpts == nil {
		opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		opts, err = c.getTLSDialOption()
		if err != nil {
			return nil, err
		}
	}

	conn, err := grpc.Dial(c.opts.CAEndpoint,
		opts,
		grpc.WithPerRPCCredentials(c.provider),
		security.CARetryInterceptor())
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to connect to endpoint", "endpoint", c.opts.CAEndpoint)
	}

	return conn, nil
}

func (c *CitadelClient) reconnectIfNeeded() error {
	if c.opts.ProvCert == "" || c.usingMtls.Load() {
		// No need to reconnect, already using mTLS or never will use it
		return nil
	}
	_, err := tls.X509KeyPair(c.tlsOpts.CertPEM, c.tlsOpts.KeyPEM)
	if err != nil {
		// Cannot load the certificates yet, don't both reconnecting
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection")
	}

	conn, err := c.buildConnection()
	if err != nil {
		return err
	}
	c.conn = conn
	c.client = pb.NewIstioCertificateServiceClient(conn)
	c.logger.V(1).Info("recreated connection")
	return nil
}

// GetRootCertBundle: Citadel (Istiod) CA doesn't publish any endpoint to retrieve CA certs
func (c *CitadelClient) GetRootCertBundle() ([]string, error) {
	return []string{}, nil
}
