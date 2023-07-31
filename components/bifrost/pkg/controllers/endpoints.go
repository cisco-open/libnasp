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
	"fmt"
	"net"
	"os"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"

	"github.com/cisco-open/nasp/components/bifrost/pkg/k8s"
	"github.com/cisco-open/nasp/components/heimdall/pkg/predicates"
	"github.com/cisco-open/operator-tools/pkg/reconciler"
)

var (
	managedByBifrostLabelSelector = map[string]string{
		k8s.ManagedByLabel: k8s.BifrostControllerName,
	}
)

type EndpointsReconciler struct {
	client.Client

	Cache  cache.Cache
	Scheme *runtime.Scheme
	Logger logr.Logger

	builder *ctrlBuilder.Builder
	ctrl    controller.Controller
}

type PortConfigs []PortConfig
type PortConfigsBySessionID map[string]PortConfigs
type PortsByServiceAndSessionID map[string]PortConfigsBySessionID

func (r *EndpointsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger.Info("endpoints reconciler triggered")

	result := ctrl.Result{}
	managedEndpoints := map[string]struct{}{}
	var multiErr error

	services, err := r.collectPortsBySessionAndService(ctx)
	if err != nil {
		return result, err
	}

	for serviceID, portsBySessionID := range services {
		svc := &corev1.Service{}
		if err := r.Get(ctx, ParsePortConfigService(serviceID), svc); err != nil {
			r.Logger.Error(err, "could not get service")
			continue
		}

		reso := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.GetName(),
				Namespace: svc.GetNamespace(),
				Labels:    managedByBifrostLabelSelector,
			},
		}

		if err := controllerutil.SetOwnerReference(svc, reso, r.Scheme); err != nil {
			r.Logger.Error(err, "could not get owner reference")
			continue
		}

		reso.Subsets = make([]corev1.EndpointSubset, 0)
		for _, ports := range portsBySessionID {
			reso.Subsets = append(reso.Subsets, r.getSubset(ports, svc.Spec.Ports))
		}

		rec := reconciler.NewGenericReconciler(r.Client, r.Logger, reconciler.ReconcilerOpts{})
		_, err := rec.ReconcileResource(reso, reconciler.StatePresent)
		if err != nil {
			multiErr = errors.Append(multiErr, errors.WrapIf(err, "could not reconcile endpoints"))
			continue
		}

		managedEndpoints[fmt.Sprintf("%s.%s", reso.GetNamespace(), reso.GetName())] = struct{}{}
	}

	endpoints := &corev1.EndpointsList{}
	if err := r.List(ctx, endpoints, client.MatchingLabels(managedByBifrostLabelSelector)); err != nil {
		return result, errors.WrapIf(err, "could not list endpoints")
	}

	for _, ep := range endpoints.Items {
		ep := ep
		if _, ok := managedEndpoints[ep.GetNamespace()+"."+ep.GetName()]; ok {
			continue
		}
		if err := r.Delete(ctx, &ep); err != nil {
			r.Logger.Error(err, "could not delete endpoints")
		}
	}

	return result, multiErr
}

func (r *EndpointsReconciler) SetupWithManager(mgr ctrl.Manager) (err error) {
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
		}, ctrlBuilder.WithPredicates(lsp, predicates.NamespacePredicate(os.Getenv("POD_NAMESPACE")))).
		Build(r)
	if err != nil {
		return err
	}

	if err := r.ctrl.Watch(source.Kind(r.Cache, &corev1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
	},
	), &handler.EnqueueRequestForObject{}, lsp); err != nil {
		return err
	}

	return nil
}

func (r *EndpointsReconciler) getSubset(ports PortConfigs, svcPorts []corev1.ServicePort) corev1.EndpointSubset {
	subset := corev1.EndpointSubset{
		Addresses: []corev1.EndpointAddress{},
		Ports:     []corev1.EndpointPort{},
	}

	isSvcPortExists := func(port PortConfig, svcPorts []corev1.ServicePort) (PortConfig, bool) {
		if port.Name == "http-metrics" {
			return port, true
		}

		for _, svcPort := range svcPorts {
			if svcPort.TargetPort.StrVal == port.Name || svcPort.TargetPort.IntValue() == port.TargetPort {
				port.Name = svcPort.Name

				return port, true
			}
		}

		return port, false
	}

	for _, port := range ports {
		var found bool
		port, found := isSvcPortExists(port, svcPorts)
		if !found {
			continue
		}

		addr, err := net.ResolveTCPAddr("tcp", port.Address)
		if err != nil {
			continue
		}

		found = false
		for _, v := range subset.Addresses {
			if v.IP == addr.IP.String() {
				found = true
				break
			}
		}
		if !found {
			subset.Addresses = append(subset.Addresses, corev1.EndpointAddress{IP: addr.IP.String()})
		}
		subset.Ports = append(subset.Ports, corev1.EndpointPort{
			Name:     port.Name,
			Protocol: corev1.ProtocolTCP,
			Port:     int32(port.Port),
		})
	}

	return subset
}

func (r *EndpointsReconciler) collectPortsBySessionAndService(ctx context.Context) (PortsByServiceAndSessionID, error) {
	services := PortsByServiceAndSessionID{}

	configmaps := &corev1.ConfigMapList{}
	if err := r.List(ctx, configmaps, client.InNamespace(os.Getenv("POD_NAMESPACE")), client.MatchingLabels{
		k8s.ManagedByLabel: k8s.BifrostControllerName,
	}); err != nil {
		return services, err
	}

	for _, cm := range configmaps.Items {
		ip := cm.Data["ip"]

		sessions := PortConfigsBySessionID{}
		err := yaml.Unmarshal([]byte(cm.Data["sessions"]), &sessions)
		if err != nil {
			return services, err
		}

		for sessionID, ports := range sessions {
			for _, port := range ports {
				if port.Service.Name == "" || port.Service.Namespace == "" {
					continue
				}

				var service PortConfigsBySessionID
				var ok bool
				if service, ok = services[port.Service.String()]; !ok {
					service = PortConfigsBySessionID{}
					services[port.Service.String()] = service
				}

				if _, ok := service[sessionID]; !ok {
					service[sessionID] = PortConfigs{}
				}

				addr, err := net.ResolveTCPAddr("tcp", port.Address)
				if err != nil {
					continue
				}
				addr.IP = net.ParseIP(ip)
				port.Address = addr.String()

				service[sessionID] = append(service[sessionID], port)
			}
		}
	}

	return services, nil
}
