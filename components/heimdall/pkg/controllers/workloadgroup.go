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

	"github.com/go-logr/logr"
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

	"github.com/banzaicloud/operator-tools/pkg/utils"
	"github.com/cisco-open/nasp/components/heimdall/pkg/predicates"
	"github.com/cisco-open/nasp/components/heimdall/pkg/server"
	istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
	k8slabels "github.com/cisco-open/nasp/pkg/k8s/labels"
)

type IstioWorkloadGroupReconciler struct {
	client.Client

	Config *rest.Config
	Scheme *runtime.Scheme
	Logger logr.Logger

	builder *ctrlBuilder.Builder
	ctrl    controller.Controller
}

const (
	//#nosec G101
	naspSecretType corev1.SecretType = "nasp.k8s.cisco.com/access-token"
)

var (
	naspLabels = labels.Set(map[string]string{
		k8slabels.NASPMonitoringLabel: "true",
	})
	tokenExpirationSeconds = 60 * 60 * 24
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

	hasNaspLabels := labels.SelectorFromSet(naspLabels).Matches(labels.Set(wg.Labels))

	if hasNaspLabels {
		err := r.createToken(ctx, wg)
		if err != nil {
			r.Logger.Error(err, "could not create token")
		}

		return ctrl.Result{}, err
	}

	err = r.deleteToken(ctx, wg)
	if err != nil {
		r.Logger.Error(err, "could not delete token")
	}

	return ctrl.Result{}, err
}

func (r *IstioWorkloadGroupReconciler) deleteToken(ctx context.Context, wg *istionetworkingv1beta1.WorkloadGroup) error {
	name := fmt.Sprintf("wg-%s-nasp-token", wg.GetName())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wg.GetNamespace(),
		},
	}

	err := r.Get(ctx, client.ObjectKey{
		Name:      secret.GetName(),
		Namespace: secret.GetNamespace(),
	}, secret)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	r.Logger.Info("delete token secret", "wgName", wg.GetName(), "wgNamespace", wg.GetNamespace())

	return r.Delete(ctx, secret)
}

func (r *IstioWorkloadGroupReconciler) createToken(ctx context.Context, wg *istionetworkingv1beta1.WorkloadGroup) error {
	name := fmt.Sprintf("%s-nasp-token", wg.GetName())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				k8slabels.NASPWorkloadgroupLabel: wg.GetName(),
			},
			Name:      name,
			Namespace: wg.GetNamespace(),
		},
		Type:      naspSecretType,
		Immutable: utils.BoolPointer(true),
		Data:      map[string][]byte{},
	}

	if err := controllerutil.SetControllerReference(wg, secret, r.Scheme); err != nil {
		return err
	}

	err := r.Get(ctx, client.ObjectKey{
		Name:      secret.GetName(),
		Namespace: secret.GetNamespace(),
	}, secret)
	if !k8serrors.IsNotFound(err) {
		return err
	}

	aud := server.NASPTokenAudience("").Generate(client.ObjectKeyFromObject(wg))

	token, err := istio_ca.CreateK8SToken(ctx, r.Config, wg.Spec.Template.ServiceAccount, wg.GetNamespace(), []string{aud}, tokenExpirationSeconds)
	if err != nil {
		return err
	}

	secret.Data["token"] = token

	r.Logger.Info("create token secret", "wgName", wg.GetName(), "wgNamespace", wg.GetNamespace())

	return r.Create(ctx, secret)
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
