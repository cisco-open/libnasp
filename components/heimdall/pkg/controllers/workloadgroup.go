// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	     https://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

package controllers

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt/v4"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/cisco-open/libnasp/components/heimdall/pkg/predicates"
	"github.com/cisco-open/libnasp/components/heimdall/pkg/server"
	istio_ca "github.com/cisco-open/libnasp/pkg/ca/istio"
	k8slabels "github.com/cisco-open/libnasp/pkg/k8s/labels"
	"github.com/cisco-open/operator-tools/pkg/reconciler"
)

type IstioWorkloadGroupReconciler struct {
	client.Client

	Config *rest.Config
	Scheme *runtime.Scheme
	Logger logr.Logger

	builder *ctrlBuilder.Builder
	ctrl    controller.Controller
}

// #nosec G101
const (
	naspSecretType  corev1.SecretType = "nasp.k8s.cisco.com/access-token"
	tokenSecretKey  string            = "token"
	tokenNameFormat string            = "%s-nasp-token"
)

var (
	naspLabels = labels.Set(map[string]string{
		k8slabels.NASPMonitoringLabel: "true",
	})
	defaultTokenExpirationDurationString = "24h"
)

func (r *IstioWorkloadGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	wg := &istionetworkingv1beta1.WorkloadGroup{}

	err := r.Get(ctx, client.ObjectKey{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, wg)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Logger.V(1).Info("reconcile", "req", req)

	hasNaspLabels := labels.SelectorFromSet(naspLabels).Matches(labels.Set(wg.Labels))

	var res *reconcile.Result
	if hasNaspLabels {
		res, err = r.createToken(ctx, wg)
		if err != nil {
			r.Logger.Error(err, "could not create token")
		}
	} else {
		res, err = r.deleteToken(wg)
		if err != nil {
			r.Logger.Error(err, "could not delete token")
		}
	}

	if res != nil {
		return *res, err
	}

	return ctrl.Result{
		RequeueAfter: time.Minute * 5,
	}, err
}

func (r *IstioWorkloadGroupReconciler) deleteToken(wg *istionetworkingv1beta1.WorkloadGroup) (*ctrl.Result, error) {
	name := fmt.Sprintf(tokenNameFormat, wg.GetName())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wg.GetNamespace(),
		},
	}

	rec := reconciler.NewGenericReconciler(r.Client, r.Logger, reconciler.ReconcilerOpts{})

	r.Logger.Info("delete token secret", "wgName", wg.GetName(), "wgNamespace", wg.GetNamespace())

	return rec.ReconcileResource(secret, reconciler.StateAbsent)
}

func (r *IstioWorkloadGroupReconciler) createToken(ctx context.Context, wg *istionetworkingv1beta1.WorkloadGroup) (*ctrl.Result, error) {
	name := fmt.Sprintf(tokenNameFormat, wg.GetName())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				k8slabels.NASPWorkloadgroupLabel: wg.GetName(),
			},
			Name:      name,
			Namespace: wg.GetNamespace(),
		},
		Type: naspSecretType,
		Data: map[string][]byte{},
	}

	if err := controllerutil.SetControllerReference(wg, secret, r.Scheme); err != nil {
		return nil, err
	}

	err := r.Get(ctx, client.ObjectKey{
		Name:      secret.GetName(),
		Namespace: secret.GetNamespace(),
	}, secret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	var createToken bool

	if k8serrors.IsNotFound(err) || secret.Data["token"] == nil {
		createToken = true
	}

	if !createToken && secret.Data[tokenSecretKey] != nil {
		// parse JWT token to extract expiration
		claims := jwt.RegisteredClaims{}
		_, _, err := jwt.NewParser().ParseUnverified(string(secret.Data["token"]), &claims)
		if err != nil {
			return nil, errors.WrapIf(err, "could not parse JWT token")
		}
		if time.Now().Add(time.Minute * 10).After(claims.ExpiresAt.Time) {
			r.Logger.Info("token expires within 10 minutes", "expireAt", claims.ExpiresAt)
			createToken = true
		}
	}

	if createToken {
		tokenExpirationDurationString := defaultTokenExpirationDurationString
		if duration, ok := wg.Annotations[k8slabels.NASPHeimdallTokenExpirationDuration]; ok {
			tokenExpirationDurationString = duration
		}

		tokenExpirationDuration, err := time.ParseDuration(tokenExpirationDurationString)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "could not parse token expiration duration", "duration", tokenExpirationDurationString)
		}

		r.Logger.Info("create token secret", "wgName", wg.GetName(), "wgNamespace", wg.GetNamespace(), "expireAt", time.Now().Add(tokenExpirationDuration))

		aud := server.NASPTokenAudience("").Generate(client.ObjectKeyFromObject(wg))
		secret.Data[tokenSecretKey], err = istio_ca.CreateK8SToken(ctx, r.Config, wg.Spec.Template.ServiceAccount, wg.GetNamespace(), []string{aud}, int(tokenExpirationDuration.Seconds()))
		if err != nil {
			return nil, err
		}
	}

	rec := reconciler.NewGenericReconciler(r.Client, r.Logger, reconciler.ReconcilerOpts{})

	return rec.ReconcileResource(secret, reconciler.StatePresent)
}

func (r *IstioWorkloadGroupReconciler) SetupWithManager(mgr ctrl.Manager) (err error) {
	r.builder = ctrl.NewControllerManagedBy(mgr)

	lsp, err := predicates.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: naspLabels,
	})
	if err != nil {
		return err
	}

	r.ctrl, err = r.builder.
		For(&istionetworkingv1beta1.WorkloadGroup{
			TypeMeta: metav1.TypeMeta{
				Kind:       "WorkloadGroup",
				APIVersion: istionetworkingv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(lsp)).
		Owns(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}).
		Build(r)

	return err
}
