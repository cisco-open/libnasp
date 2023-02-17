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
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
)

// AfterCheckFunctionType is the function signature for functions run if the
// check was triggered from outside or after the certificate was renewed.
type AfterCheckFunctionType func(certificate *Certificate, needsUpdate bool) error

// Renewer handles the automatic renewal of certificates.
type Renewer struct {
	// logger is the interface containing log helper functions.
	logger logr.Logger

	// dnsNames are holding the DNS names for the server certificate generation.
	dnsNames []string

	// certificateDirectoryPath is the path of the directory where the generated
	// certificates should be written.
	certificateDirectoryPath string

	// afterCheckFunctions are optional callbacks for triggered checks and
	// renewals.
	afterCheckFunctions []AfterCheckFunctionType
}

// NewRenewer returns a certificate renewer configured to the specified values.
func NewRenewer(
	logger logr.Logger,
	dnsNames []string,
	certificateDirectoryPath string,
	shouldCheckCertificate bool,
	afterCheckFunctions ...AfterCheckFunctionType,
) (*Renewer, error) {
	if dnsNames == nil {
		dnsNames = []string{}
	}

	renewer := &Renewer{
		logger:                   logger,
		dnsNames:                 dnsNames,
		certificateDirectoryPath: certificateDirectoryPath,
		afterCheckFunctions:      afterCheckFunctions,
	}

	if shouldCheckCertificate {
		err := renewer.checkCertificate(true)
		if err != nil {
			return nil, errors.Wrap(err, "checking certificate failed")
		}
	}

	return renewer, nil
}

func (renewer *Renewer) Start(ctx context.Context, triggers <-chan struct{}) error {
	if renewer == nil {
		return errors.Errorf("invalid nil renewer")
	}

	err := renewer.checkCertificate(true)
	if err != nil {
		return errors.Wrap(err, "checking certificate failed")
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-triggers:
			_ = renewer.checkCertificate(true)
		case <-ticker.C:
			_ = renewer.checkCertificate(false)
		case <-ctx.Done():
			return nil
		}
	}
}

// CheckCertificate checks the certificate to be valid, renews it if it is
// not and runs the after check functions.
func (renewer *Renewer) checkCertificate(wasTriggered bool) error {
	if renewer == nil {
		return errors.Errorf("invalid nil renewer")
	}

	renewer.logger.Info("checking certificate")

	isValid, certificate, err := renewer.verifyCertificate()
	if err != nil {
		renewer.logger.Error(err, "verifying certificate failed")
	}

	var wasRenewed bool

	if isValid {
		renewer.logger.Info("certificate is valid")
	} else {
		renewer.logger.Info("certificate is invalid")

		certificate, err = renewer.renewCertificate()
		if err != nil {
			renewer.logger.Error(err, "renewing certificate failed")
		}

		wasRenewed = true
	}

	for _, afterCheckFunction := range renewer.afterCheckFunctions {
		err = afterCheckFunction(certificate, wasRenewed || wasTriggered)
		if err != nil {
			renewer.logger.Error(err, "running after check function failed")
		}
	}

	return nil
}

// renewCertificate generates and returns a new certificate for the stored DNS
// names and writes them to the stored directory path.
func (renewer *Renewer) renewCertificate() (*Certificate, error) {
	if renewer == nil {
		return nil, errors.Errorf("invalid nil renewer")
	}

	renewer.logger.Info("renewing certificate")

	certificate, err := NewCertificate(renewer.dnsNames)
	if err != nil {
		return nil, errors.WrapWithDetails(
			err,
			"creating new certificate for DNS names failed",
			"dnsNames", renewer.dnsNames,
		)
	}

	err = certificate.Write(renewer.certificateDirectoryPath)
	if err != nil {
		return nil, errors.WrapWithDetails(
			err,
			"writing renewed certificate to directory failed",
			"directoryPath", renewer.certificateDirectoryPath,
		)
	}

	return certificate, nil
}

// verifyCertificate verifies whether the underlying certificate is still valid.
func (renewer *Renewer) verifyCertificate() (isValid bool, certificate *Certificate, err error) {
	if renewer == nil {
		return false, nil, errors.Errorf("invalid nil renewer")
	}

	renewer.logger.Info("verifying certificate")

	certificate, err = NewCertificateFromDirectory(renewer.certificateDirectoryPath)
	if err != nil {
		return false, nil, errors.WrapWithDetails(
			err,
			"loading certificate from directory failed",
			"directoryPath", renewer.certificateDirectoryPath,
		)
	}

	renewer.logger.Info("verifying certificate")

	dnsNames := renewer.dnsNames
	if len(dnsNames) == 0 {
		dnsNames = append(dnsNames, "")
	}

	isValid = true

	for _, dnsName := range dnsNames {
		isValid = isValid && certificate.Verify(dnsName, time.Now().AddDate(0, 6, 0))
		if !isValid {
			break
		}
	}

	return isValid, certificate, nil
}

// WithAfterCheckFunctions adds appends the specified functions to the existing
// after check function chain.
func (renewer *Renewer) WithAfterCheckFunctions(afterCheckFunctions ...AfterCheckFunctionType) {
	if renewer == nil {
		return
	}

	renewer.afterCheckFunctions = append(renewer.afterCheckFunctions, afterCheckFunctions...)
}

// WithDNSNames sets the DNS names used by the renewer in the certificate.
func (renewer *Renewer) WithDNSNames(dnsNames ...string) {
	if renewer == nil {
		return
	}

	renewer.dnsNames = dnsNames
}
