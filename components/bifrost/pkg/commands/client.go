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
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/components/bifrost/pkg/k8s"
	"github.com/cisco-open/nasp/pkg/istio"
	"github.com/cisco-open/nasp/pkg/network/proxy"
	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
	"github.com/cisco-open/nasp/pkg/network/tunnel/client"
)

var ErrLocalAddressNotSpecified = errors.New("at least one local address must be specified")

var addrRegex = regexp.MustCompile(`^((\d+):)?(.*):(\d+)$`)

func NewClientCommand() *cobra.Command {
	logger := klog.Background()

	cmd := &cobra.Command{
		Use:   "client",
		Short: "tcp tunneling client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			localAddresses := viper.GetStringSlice("local-address")
			serverAddress := viper.GetString("server-address")
			naspEnabled := viper.GetBool("with-nasp")
			directClient := !viper.GetBool("with-nasp-bifrost")
			authToken := viper.GetString("auth-token")

			if len(localAddresses) == 0 {
				return ErrLocalAddressNotSpecified
			}

			ctx := context.Background()

			var tclient TunnelClient
			var dialer api.Dialer
			dialer = &net.Dialer{}

			if naspEnabled {
				istioConfig := istio.DefaultIstioIntegrationHandlerConfig
				if !directClient {
					istioConfig.BifrostAddress = serverAddress
				}
				ih, err := istio.NewIstioIntegrationHandler(&istioConfig, logger)
				if err != nil {
					return errors.WrapIf(err, "could not instantiate NASP istio integration handler")
				}
				tclient = ih

				if err := ih.Run(ctx); err != nil {
					return errors.WrapIf(err, "could not run istio integration handler")
				}

				dialer, err = ih.GetTCPDialer()
				if err != nil {
					return errors.WrapIf(err, "could not get NASP tcp dialer")
				}
			}

			if !naspEnabled || directClient {
				opts := []client.ClientOption{
					client.ClientWithLogger(logger),
					client.ClientWithDialer(dialer),
				}
				if authToken != "" {
					opts = append(opts, client.ClientWithBearerToken(authToken))
				}

				opts = append(opts, client.ClientWithMetadata(map[string]string{
					k8s.ClientServiceNameLabel: "demo",
				}))
				c := client.NewClient(serverAddress, opts...)
				go func() {
					if err := c.Connect(context.Background()); err != nil {
						panic(err)
					}
				}()

				tclient = &tunnelClient{
					client: c,
				}
			}

			return runProxies(ctx, localAddresses, dialer, tclient, logger)
		},
	}

	cmd.Flags().StringP("server-address", "s", "127.0.0.1:8001", "Server address")
	cmd.Flags().StringSliceP("local-address", "l", []string{}, "Local address to expose")
	cmd.Flags().Bool("with-nasp", false, "Whether to use Nasp")
	cmd.Flags().Bool("with-nasp-bifrost", true, "Whether to use Nasp internal bifrost client or direct client")
	cmd.Flags().String("auth-token", "", "Bifrost authentication token")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}

	return cmd
}

func runProxies(ctx context.Context, localAddresses []string, dialer api.Dialer, tclient TunnelClient, logger logr.Logger) error {
	wg := sync.WaitGroup{}

	for k, addr := range localAddresses {
		var requestedPort int
		var targetPort int
		v := addrRegex.FindStringSubmatch(addr)
		if len(v) < 5 {
			logger.Info("invalid local address", "address", addr)
			continue
		}
		if v, err := strconv.Atoi(v[2]); err == nil {
			requestedPort = v
		}
		if v, err := strconv.Atoi(v[4]); err == nil {
			targetPort = v
		}
		addr = fmt.Sprintf("%s:%s", v[3], v[4])
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return err
		}

		name := fmt.Sprintf("port-%d", k+1)

		l, err := tclient.GetVirtualTCPListener(requestedPort, targetPort, name)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(logger logr.Logger) {
			logger.Info("start proxy", "destination", tcpAddr.String())
			defer func() {
				logger.Info("proxying stopped")
				wg.Done()
			}()
			(&proxyclient{
				Listener: l,

				dialer:  dialer,
				logger:  logger,
				tcpAddr: tcpAddr,
			}).Run(ctx)
		}(logger.WithName(name))
	}

	wg.Wait()

	return nil
}

type TunnelClient interface {
	GetVirtualTCPListener(requestedPort int, targetPort int, name string) (net.Listener, error)
}

type tunnelClient struct {
	client api.Client
}

func (c *tunnelClient) GetVirtualTCPListener(requestedPort int, targetPort int, name string) (net.Listener, error) {
	return c.client.GetTCPListener((&client.ManagedPortOptions{}).SetRequestedPort(requestedPort).SetTargetPort(targetPort).SetName(name))
}

type proxyclient struct {
	net.Listener

	dialer  api.Dialer
	logger  logr.Logger
	tcpAddr *net.TCPAddr
}

func (p *proxyclient) Run(ctx context.Context) {
	for {
		rconn, err := p.Listener.Accept()
		if err != nil {
			p.logger.Error(err, "accept error")
			if errors.Is(err, api.ErrListenerStopped) {
				break
			}
			continue
		}

		lconn, err := p.dialer.DialContext(ctx, "tcp", p.tcpAddr.String())
		if err != nil {
			p.logger.Error(err, "could not dial")
			rconn.Close()
			continue
		}

		p.logger.V(3).Info("start proxying", "client", rconn.RemoteAddr(), "server", lconn.RemoteAddr())

		go proxy.New(rconn, lconn).Start()
	}
}
