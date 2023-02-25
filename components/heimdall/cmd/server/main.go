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

package main

import (
	"net"

	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
	"github.com/cisco-open/nasp/components/heimdall/pkg/controllers"
	"github.com/cisco-open/nasp/components/heimdall/pkg/server"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = cluster_registry.AddToScheme(scheme)
	_ = istionetworkingv1beta1.AddToScheme(scheme)
}

func main() {
	kubeconfig, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	logger := klog.TODO()

	mgr, err := manager.New(kubeconfig, manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
	})
	if err != nil {
		panic(err)
	}

	if err := (&controllers.IstioWorkloadGroupReconciler{
		Client: mgr.GetClient(),
		Config: mgr.GetConfig(),
		Scheme: mgr.GetScheme(),
		Logger: logger,
	}).SetupWithManager(mgr); err != nil {
		panic(err)
	}

	ctx := ctrl.SetupSignalHandler()

	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic(err)
		}
	}()

	//#nosec G102
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	if err := server.New(listener, mgr, logger).Run(ctx); err != nil {
		panic(err)
	}
}
