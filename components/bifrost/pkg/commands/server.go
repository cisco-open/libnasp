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

package commands

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/cisco-open/nasp/components/bifrost/pkg/controllers"
	"github.com/cisco-open/nasp/components/bifrost/pkg/k8s"
	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
	"github.com/cisco-open/nasp/pkg/network/tunnel/auth"
	"github.com/cisco-open/nasp/pkg/network/tunnel/server"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func NewServerCommand() *cobra.Command {
	logger := klog.Background()

	cmd := &cobra.Command{
		Use:   "server",
		Short: "tcp tunneling server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			serverAddress := viper.GetString("server-address")
			healthCheckAddress := viper.GetString("healthcheck-address")
			naspEnabled := viper.GetBool("with-nasp")

			options := []server.ServerOption{server.ServerWithLogger(logger)}

			ctx := ctrl.SetupSignalHandler()

			if naspEnabled {
				istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
				ih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, logger)
				if err != nil {
					return errors.WrapIf(err, "could not instantiate NASP istio integration handler")
				}
				if err := ih.Run(ctx); err != nil {
					return err
				}

				options = append(options, server.ServerWithListenerWrapperFunc(ih.GetTCPListener))
			}

			if os.Getenv("POD_NAMESPACE") != "" {
				kubeconfig, err := config.GetConfig()
				if err != nil {
					return err
				}

				ports := &sync.Map{}

				mgr, configMapReconciler, err := startConfigMapManager(ctx, kubeconfig, ports, logger)
				if err != nil {
					return err
				}

				if err := startEndpointsManager(ctx, kubeconfig, logger); err != nil {
					return err
				}

				options = append(options, server.ServerWithAuthenticator(auth.NewK8sAuthenticator(
					mgr.GetClient(),
					auth.K8sAuthenticatorWithAud(auth.DefaultK8sAuthenticatorAud, "istio-ca"),
				)))

				options = append(options, server.ServerWithEventSubscribe(string(api.PortListenEventName), func(topic string, msg api.PortListenEvent) {
					ports.Store(msg.Port, controllers.PortConfig{
						Name:       msg.Name,
						Address:    msg.Address,
						Port:       msg.Port,
						TargetPort: msg.TargetPort,
						Service: controllers.PortConfigService{
							Name:      msg.Metadata[k8s.ClientServiceNameLabel],
							Namespace: msg.User.Metadata()["namespace"],
						},
						SessionID: msg.SessionID,
					})
					configMapReconciler.TriggerReconcile()
				}))

				options = append(options, server.ServerWithEventSubscribe(string(api.PortReleaseEventName), func(topic string, msg api.PortReleaseEvent) {
					ports.Delete(msg.Port)
					configMapReconciler.TriggerReconcile()
				}))
			}

			srv, err := server.NewServer(serverAddress, options...)
			if err != nil {
				return err
			}

			go simpleHealthCheck(healthCheckAddress, logger)

			err = srv.Start(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("server-address", "0.0.0.0:8001", "Control server address.")
	cmd.Flags().String("healthcheck-address", "0.0.0.0:8002", "HTTP healthcheck address.")
	cmd.Flags().Bool("with-nasp", false, "Whether to use Nasp")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}

	return cmd
}

func startConfigMapManager(ctx context.Context, kubeconfig *rest.Config, ports *sync.Map, logger logr.Logger) (manager.Manager, *controllers.ConfigMapReconciler, error) {
	logger.Info("start configmap manager")

	mgr, err := manager.New(kubeconfig, manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		LeaderElection:     false,
		Namespace:          os.Getenv("POD_NAMESPACE"),
		Logger:             logger,
	})
	if err != nil {
		return nil, nil, err
	}

	rec := (&controllers.ConfigMapReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger,
		Ports:  ports,

		PodInfo: controllers.PodInfo{
			Name:      os.Getenv("POD_NAME"),
			Namespace: os.Getenv("POD_NAMESPACE"),
			IP:        os.Getenv("POD_IP"),
		},
	})

	if err := rec.SetupWithManager(mgr); err != nil {
		return nil, nil, err
	}

	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic(err)
		}
	}()

	mgr.GetCache().WaitForCacheSync(ctx)

	return mgr, rec, nil
}

func startEndpointsManager(ctx context.Context, kubeconfig *rest.Config, logger logr.Logger) error {
	logger.Info("start endpoints manager")

	mgr, err := manager.New(kubeconfig, manager.Options{
		Scheme:                        scheme,
		MetricsBindAddress:            "0",
		LeaderElection:                true,
		LeaderElectionNamespace:       os.Getenv("POD_NAMESPACE"),
		LeaderElectionID:              "endpoints.bifrost.k8s.cisco.com",
		LeaderElectionReleaseOnCancel: true,
		Logger:                        logger,
	})
	if err != nil {
		return err
	}

	rec := (&controllers.EndpointsReconciler{
		Cache:  mgr.GetCache(),
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Logger: logger,
	})

	if err := rec.SetupWithManager(mgr); err != nil {
		return err
	}

	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic(err)
		}
	}()

	mgr.GetCache().WaitForCacheSync(ctx)

	return nil
}

func simpleHealthCheck(address string, logger logr.Logger) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	m := http.DefaultServeMux
	m.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, _ = rw.Write([]byte("ok"))
	}))

	srv := http.Server{
		Handler: m,
	}

	logger.Info("start healthcheck server", "address", address)

	if err := srv.Serve(l); err != nil {
		panic(err)
	}
}
