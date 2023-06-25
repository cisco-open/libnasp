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

package controllers

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/operator-tools/pkg/reconciler"

	"github.com/cisco-open/nasp/components/heimdall/pkg/predicates"
)

type ConfigMapReconciler struct {
	client.Client

	Scheme *runtime.Scheme
	Logger logr.Logger

	builder *ctrlBuilder.Builder
	ctrl    controller.Controller

	Ports *sync.Map

	PodInfo PodInfo

	events chan event.GenericEvent
}

type PodInfo struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	IP        string `json:"ip,omitempty"`
}

type PortConfig struct {
	Name       string            `json:"name,omitempty"`
	Address    string            `json:"address,omitempty"`
	Port       int               `json:"port,omitempty"`
	TargetPort int               `json:"targetPort,omitempty"`
	Service    PortConfigService `json:"service,omitempty"`
	SessionID  string            `json:"sessionID,omitempty"`
}

type PortConfigService struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func (s PortConfigService) String() string {
	return s.Name + "." + s.Namespace
}

func ParsePortConfigService(str string) client.ObjectKey {
	key := strings.SplitN(str, ".", 2)

	return client.ObjectKey{
		Name:      key[0],
		Namespace: key[1],
	}
}

func (r *ConfigMapReconciler) TriggerReconcile() {
	r.events <- event.GenericEvent{
		Object: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      os.Getenv("POD_NAME"),
				Namespace: r.PodInfo.Namespace,
			},
		},
	}
}

func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	podName := os.Getenv("POD_NAME")
	result := ctrl.Result{}

	if req.Name != podName {
		return result, nil
	}

	r.Logger.Info("reconcile", "resource", req)

	pod := &corev1.Pod{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      podName,
		Namespace: r.PodInfo.Namespace,
	}, pod); err != nil {
		r.Logger.Error(err, "could not get self pod")
		result.RequeueAfter = time.Second * 5

		return result, nil
	}

	sessions := map[string][]PortConfig{}

	r.Ports.Range(func(key, val any) bool {
		if pc, ok := val.(PortConfig); ok {
			if _, ok := sessions[pc.SessionID]; !ok {
				sessions[pc.SessionID] = make([]PortConfig, 0)
			}
			sessions[pc.SessionID] = append(sessions[pc.SessionID], pc)
		}

		return true
	})

	sessionsYaml, err := yaml.Marshal(sessions)
	if err != nil {
		return result, err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: r.PodInfo.Namespace,
			Labels:    managedByBifrostLabelSelector,
		},
		Data: map[string]string{
			"ip":       r.PodInfo.IP,
			"sessions": string(sessionsYaml),
		},
	}

	if err := controllerutil.SetOwnerReference(pod, cm, r.Scheme); err != nil {
		return result, err
	}

	rec := reconciler.NewGenericReconciler(r.Client, r.Logger, reconciler.ReconcilerOpts{})
	res, err := rec.ReconcileResource(cm, reconciler.StatePresent)
	if err != nil {
		return result, err
	}

	if res != nil {
		result = *res
	}

	return result, nil
}

func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) (err error) {
	r.events = make(chan event.GenericEvent)

	r.builder = ctrl.NewControllerManagedBy(mgr)

	lsp, err := predicates.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: labels.Set(managedByBifrostLabelSelector),
	})
	if err != nil {
		return err
	}

	r.ctrl, err = r.builder.
		For(&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(lsp, predicates.NamespacePredicate(r.PodInfo.Namespace))).
		Build(r)
	if err != nil {
		return err
	}

	if err := r.ctrl.Watch(&source.Channel{
		Source: r.events,
	}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	return nil
}
