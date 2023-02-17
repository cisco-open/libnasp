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
	"fmt"
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

// WebhookCertifier handles the generation and renewal of webhook
// certificates.
type WebhookCertifier struct {
	// logger is the log handler of the certifier.
	logger logr.Logger

	// webhookName is the name of the webhook the certifier is related to.
	webhookName string

	// webhookManager is the manager of the webhook.
	webhookManager manager.Manager

	// mutatingWebhookConfiguration holds the configuration of a mutating
	// webhook.
	mutatingWebhookConfiguration *admissionregistrationv1.MutatingWebhookConfiguration

	// validatingWebhookConfiguration holds the configuration of a validating
	// webhook.
	validatingWebhookConfiguration *admissionregistrationv1.ValidatingWebhookConfiguration

	// certificateRenewer is the renewer handling the webhook's certificates.
	certificateRenewer *Renewer

	// triggers is the channel to notify the certificate renewer on changes.
	triggers chan struct{}
}

// NewWebhookCertifier returns an object which can manage the generation and
// renewal of a certificate based on a webhook configuration.
func NewWebhookCertifier(
	logger logr.Logger,
	webhookName string,
	webhookManager manager.Manager,
	certificateRenewer *Renewer,
) *WebhookCertifier {
	return &WebhookCertifier{
		logger: logger.WithValues("key", client.ObjectKey{
			Name: webhookName,
		}),
		webhookName:        webhookName,
		webhookManager:     webhookManager,
		triggers:           make(chan struct{}),
		certificateRenewer: certificateRenewer,
	}
}

func (certifier *WebhookCertifier) Start(ctx context.Context) error {
	err := certifier.loadWebhookConfiguration()
	if err != nil {
		return errors.Wrap(err, "loading webhook configuration failed")
	}

	err = certifier.startWebhookConfigurationInformer()
	if err != nil {
		return errors.Wrap(err, "starting webhook configuration informer failed")
	}

	certifier.certificateRenewer.WithDNSNames(certifier.dnsNames()...)

	certifier.certificateRenewer.WithAfterCheckFunctions(func(c *Certificate, needsUpdate bool) error {
		failPolicy := admissionregistrationv1.Fail
		switch {
		case certifier.mutatingWebhookConfiguration != nil:
			for index, webhook := range certifier.mutatingWebhookConfiguration.Webhooks {
				webhook.FailurePolicy = &failPolicy
				webhook.ClientConfig.CABundle = c.CACertificate
				certifier.mutatingWebhookConfiguration.Webhooks[index] = webhook
			}
		case certifier.validatingWebhookConfiguration != nil:
			for index, webhook := range certifier.validatingWebhookConfiguration.Webhooks {
				webhook.FailurePolicy = &failPolicy
				webhook.ClientConfig.CABundle = c.CACertificate
				certifier.validatingWebhookConfiguration.Webhooks[index] = webhook
			}
		default:
			return errors.New("invalid certifier, all known webhook configurations are nil")
		}

		if needsUpdate {
			err = certifier.updateCertificate()
			if err != nil {
				return errors.Wrap(err, "updating webhook certificate failed")
			}
		}

		return nil
	})

	defer close(certifier.triggers)

	return certifier.certificateRenewer.Start(ctx, certifier.triggers)
}

// dnsNames returns the DNS names of the certifier's webhook configuration.
func (certifier *WebhookCertifier) dnsNames() []string {
	if certifier == nil {
		return nil
	}

	switch {
	case certifier.mutatingWebhookConfiguration != nil:
		dnsNames := make([]string, 0, len(certifier.mutatingWebhookConfiguration.Webhooks))

		for _, webhook := range certifier.mutatingWebhookConfiguration.Webhooks {
			dnsNames = append(
				dnsNames,
				fmt.Sprintf("%s.%s.svc", webhook.ClientConfig.Service.Name, webhook.ClientConfig.Service.Namespace),
			)
		}

		return dnsNames
	case certifier.validatingWebhookConfiguration != nil:
		dnsNames := make([]string, 0, len(certifier.validatingWebhookConfiguration.Webhooks))

		for _, webhook := range certifier.validatingWebhookConfiguration.Webhooks {
			dnsNames = append(
				dnsNames,
				fmt.Sprintf("%s.%s.svc", webhook.ClientConfig.Service.Name, webhook.ClientConfig.Service.Namespace),
			)
		}

		return dnsNames
	default:
		return nil
	}
}

// loadWebhookConfiguration retrieves the configuration of the corresponding
// webhook and stores it in the certifier.
func (certifier *WebhookCertifier) loadWebhookConfiguration() error {
	if certifier == nil {
		return errors.New("invalid nil webhook certifier")
	}

	mgrClient := certifier.webhookManager.GetClient()

	certifier.mutatingWebhookConfiguration = &admissionregistrationv1.MutatingWebhookConfiguration{}

	certifier.logger.Info("trying to retrieve mutating webhook configuration")

	mutatingErr := mgrClient.Get(context.Background(), client.ObjectKey{
		Name: certifier.webhookName,
	}, certifier.mutatingWebhookConfiguration)
	if mutatingErr == nil {
		certifier.logger.Info("retrieved mutating webhook configuration successfully")

		return nil
	}

	certifier.mutatingWebhookConfiguration = nil

	certifier.logger.Info("trying to retrieve validating webhook configuration")

	certifier.validatingWebhookConfiguration = &admissionregistrationv1.ValidatingWebhookConfiguration{}

	validatingErr := mgrClient.Get(context.Background(), client.ObjectKey{
		Name: certifier.webhookName,
	}, certifier.validatingWebhookConfiguration)
	if validatingErr == nil {
		certifier.logger.Info("retrieved validating webhook configuration successfully")

		return nil
	}

	err := errors.Wrap(
		errors.Combine(mutatingErr, validatingErr),
		"retrieving webhook configuration failed for each known webhook type",
	)

	certifier.logger.Error(err, "loading webhook configuration failed")

	return err
}

// startWebhookConfigurationInformer initializes an informer on webhook
// configuration updates.
func (certifier *WebhookCertifier) startWebhookConfigurationInformer() error {
	var config runtimeclient.Object

	switch {
	case certifier.mutatingWebhookConfiguration != nil:
		config = certifier.mutatingWebhookConfiguration
	case certifier.validatingWebhookConfiguration != nil:
		config = certifier.validatingWebhookConfiguration
	default:
		return errors.New("invalid certifier, all known webhook configurations are nil")
	}

	kind := strings.Split(reflect.TypeOf(config).String(), ".")[1] // Note: removing *v1. pointer and package prefix.
	logger := certifier.logger.WithValues("kind", kind)

	logger.Info("starting webhook configuration informer")

	informer, err := certifier.webhookManager.GetCache().GetInformerForKind(
		context.Background(),
		admissionregistrationv1.SchemeGroupVersion.WithKind(kind),
	)
	if err != nil {
		err = errors.Wrap(err, "retrieving informer for webhook failed")

		logger.Error(err, "starting webhook configuration informer failed")

		return err
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldConfig, newConfig interface{}) {
			var config runtimeclient.Object
			var ok bool

			switch {
			case certifier.mutatingWebhookConfiguration != nil:
				config, ok = newConfig.(*admissionregistrationv1.MutatingWebhookConfiguration)
			case certifier.validatingWebhookConfiguration != nil:
				config, ok = newConfig.(*admissionregistrationv1.ValidatingWebhookConfiguration)
			default:
				logger.Error(
					errors.New("invalid certifier, all known webhook configurations are nil"),
					"webhook configuration informer update function failed",
				)
			}

			if ok &&
				config.GetName() == certifier.webhookName {
				err := certifier.webhookManager.GetClient().Get(context.Background(), client.ObjectKey{
					Name: certifier.webhookName,
				}, config)
				if err != nil {
					logger.Error(err, "retrieving webhook configuration failed")

					return
				}

				logger.Info(
					"triggering certificate renewer on webhook configuration informer update",
				)

				certifier.triggers <- struct{}{}
			}
		},
	})

	logger.Info("started webhook configuration informer successfully")

	return nil
}

// updateCertificate updates the webhook's certificate with the renewer's
// current one.
func (certifier *WebhookCertifier) updateCertificate() error {
	logger := certifier.logger

	logger.Info("updating webhook certificate")

	if certifier == nil {
		err := errors.New("invalid nil certifier")

		logger.Error(err, "updating webhook certificate failed")

		return err
	}

	var desiredConfig runtimeclient.Object
	var currentConfig runtimeclient.Object

	switch {
	case certifier.mutatingWebhookConfiguration != nil:
		desiredConfig = certifier.mutatingWebhookConfiguration
		currentConfig = &admissionregistrationv1.MutatingWebhookConfiguration{}
	case certifier.validatingWebhookConfiguration != nil:
		desiredConfig = certifier.validatingWebhookConfiguration
		currentConfig = &admissionregistrationv1.ValidatingWebhookConfiguration{}
	default:
		err := errors.NewWithDetails("invalid certifier, all webhook configurations are nil")

		logger.Error(err, "choosing configuration for webhook certificate update failed")

		return err
	}

	key := runtimeclient.ObjectKey{
		Name: certifier.webhookName,
	}
	logger = logger.WithValues("kind", strings.Split(reflect.TypeOf(desiredConfig).String(), ".")[1])

	client := certifier.webhookManager.GetClient()

	err := client.Get(context.Background(), key, currentConfig)
	if err != nil &&
		!apierrors.IsNotFound(err) {
		err = errors.Wrap(err, "retrieving webhook configuration failed")

		logger.Error(err, "updating webhook certificate failed")

		return err
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(currentConfig, desiredConfig, patch.IgnoreStatusFields())
	if err != nil {
		err = errors.WrapWithDetails(err, "calculating config patch failed")

		logger.Error(err, "updating webhook certificate failed")

		return err
	} else if patchResult.IsEmpty() {
		logger.Info("current config is up to date with the desired state, there is nothing to update")

		return nil
	} else {
		logger.Info("updating webhook configuration with new certificate")
	}

	// Need to set this before resourceversion is set, as it would constantly
	// change otherwise.
	err = patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredConfig)
	if err != nil {
		err = errors.Wrap(err, "setting last applied annotation on webhook configuration failed")

		logger.Error(err, "updating webhook certificate failed")

		return err
	}

	metaAccessor := meta.NewAccessor()

	currentResourceVersion, err := metaAccessor.ResourceVersion(currentConfig)
	if err != nil {
		err = errors.Wrap(err, "retrieving current webhook configuration resource version failed")

		logger.Error(err, "updating webhook certificate failed")

		return err
	}

	err = metaAccessor.SetResourceVersion(desiredConfig, currentResourceVersion)
	if err != nil {
		err = errors.Wrap(err, "setting webhook configuration resource version failed")

		logger.Error(err, "updating webhook certificate failed")

		return err
	}

	err = client.Update(context.Background(), desiredConfig)
	if err != nil {
		err = errors.Wrap(err, "updating webhook configuration failed")

		logger.Error(err, "updating webhook certificate failed")

		return err
	}

	logger.Info("webhook configuration certificate updated successfully")

	return nil
}
