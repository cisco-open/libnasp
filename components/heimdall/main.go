/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/cisco-open/heimdall-webhook/pkg/cert"
	"github.com/cisco-open/heimdall-webhook/pkg/webhooks"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	clusterregistryv1alpha1 "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"

	"github.com/cisco-open/heimdall-webhook/pkg/k8sutil"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	_ = clusterregistryv1alpha1.AddToScheme(scheme)

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var clusterName string
	var istioMeshID string
	var istioVersion string
	var istioCAAddress string
	var istioNetwork string
	var istioRevision string
	var webhookPort int
	var webhookName string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8081", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8082", "The address the probe endpoint binds to.")
	flag.StringVar(&clusterName, "cluster-name", "", "The name of the cluster.")
	flag.StringVar(&istioMeshID, "istio-mesh-id", "", "The mesh id of the istio mesh.")
	flag.StringVar(&istioVersion, "istio-version", "", "The version of the istio mesh.")
	flag.StringVar(&istioCAAddress, "istio-ca-address", "", "The ca address of the istio mesh.")
	flag.StringVar(&istioNetwork, "istio-network", "", "The network of the istio mesh.")
	flag.StringVar(&istioRevision, "istio-revision", "", "The used revision of the istio mesh.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.IntVar(&webhookPort, "webhook-port", 9443, "The port of the webhook server.")
	flag.StringVar(&webhookName, "webhook-name", "nasp-config-injector", "Name of the mutating webhook resource.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgrOptions := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   webhookPort,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "webhook.nasp.k8s.cisco.com",
		CertDir:                "/tmp/cert",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		LeaderElectionReleaseOnCancel: true,
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if clusterName == "" {
		cluster, err := k8sutil.GetLocalCluster(context.Background(), mgr.GetAPIReader())
		if err != nil {
			setupLog.Error(err, "could not get local cluster")
		}
		clusterName = cluster.GetName()
		setupLog.Info("local cluster detected", "name", cluster.GetName())
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	podMutatorLogger := ctrl.Log.WithName("pod-mutator")
	mgr.GetWebhookServer().Register(
		"/mutate-pod",
		&webhook.Admission{
			Handler: webhooks.NewPodMutator(webhooks.PodMutatorConfig{
				ClusterName:    clusterName,
				MeshID:         istioMeshID,
				IstioVersion:   istioVersion,
				IstioCAAddress: istioCAAddress,
				IstioNetwork:   istioNetwork,
				IstioRevision:  istioRevision,
			}, podMutatorLogger, mgr),
		},
	)

	podMutatorCertRenewer, err := cert.NewRenewer(
		podMutatorLogger,
		nil,
		mgrOptions.CertDir,
		true,
	)
	if err != nil {
		setupLog.Error(err, "initializing certificate renewer failed")

		os.Exit(1)
	}

	err = mgr.Add(
		cert.NewWebhookCertifier(
			podMutatorLogger,
			webhookName,
			mgr,
			podMutatorCertRenewer,
		),
	)
	if err != nil {
		setupLog.Error(err, "adding certificate provisioner to manager failed")

		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
