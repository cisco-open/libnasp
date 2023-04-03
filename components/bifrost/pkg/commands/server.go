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

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network/tunnel/server"
)

func NewServerCommand() *cobra.Command {
	var serverAddress string
	var healthcheckAddress string
	var naspSupportEnabled bool

	logger := klog.Background()

	cmd := &cobra.Command{
		Use:   "server",
		Short: "tcp tunneling server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			options := []server.ServerOption{server.ServerWithLogger(logger)}

			if naspSupportEnabled {
				istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
				ih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, logger)
				if err != nil {
					return errors.WrapIf(err, "could not instantiate NASP istio integration handler")
				}
				options = append(options, server.ServerWithListenerWrapperFunc(ih.GetTCPListener))
			}

			srv, err := server.NewServer(serverAddress, options...)
			if err != nil {
				return err
			}

			go simpleHealthCheck(healthcheckAddress)

			err = srv.Start(context.Background())
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverAddress, "server-address", "0.0.0.0:8001", "Control server address.")
	cmd.Flags().StringVar(&healthcheckAddress, "healthcheck-address", "0.0.0.0:8002", "HTTP healthcheck address.")
	cmd.Flags().BoolVar(&naspSupportEnabled, "with-nasp-support", false, "Enables NASP support")

	return cmd
}

func simpleHealthCheck(address string) {
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

	if err := srv.Serve(l); err != nil {
		panic(err)
	}
}
