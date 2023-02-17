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
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/network/proxy"
	"github.com/cisco-open/nasp/pkg/network/tunnel/api"
	"github.com/cisco-open/nasp/pkg/network/tunnel/client"
)

var ErrLocalAddressNotSpecified = errors.New("at least one local address must be specified")

var addrRegex = regexp.MustCompile(`^((\d+):)?(.*):(\d+)$`)

func NewClientCommand() *cobra.Command {
	var serverAddress string
	localAddresses := []string{}

	logger := klog.Background()

	cmd := &cobra.Command{
		Use:   "client",
		Short: "tcp tunneling client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if len(localAddresses) == 0 {
				return ErrLocalAddressNotSpecified
			}
			c := client.NewClient(serverAddress, client.ClientWithLogger(logger))
			go func() {
				if err := c.Connect(context.Background()); err != nil {
					panic(err)
				}
			}()

			wg := sync.WaitGroup{}

			for k, addr := range localAddresses {
				var requestedPort int
				v := addrRegex.FindStringSubmatch(addr)
				if len(v) < 5 {
					logger.Info("invalid local address", "address", addr)
					continue
				}
				if v, err := strconv.Atoi(v[2]); err == nil {
					requestedPort = v
				}
				addr = fmt.Sprintf("%s:%s", v[3], v[4])
				tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
				if err != nil {
					return err
				}

				name := fmt.Sprintf("port-%d", k+1)

				l, err := c.AddTCPPort(name, requestedPort)
				if err != nil {
					return err
				}

				wg.Add(1)
				go func(logger logr.Logger) {
					logger.Info("start proxying", "destination", tcpAddr.String())
					defer func() {
						logger.Info("proxying stopped")
						wg.Done()
					}()
					(&proxyclient{
						Listener: l,

						logger:  logger,
						tcpAddr: tcpAddr,
					}).Run()
				}(logger.WithName(name))
			}

			wg.Wait()

			return nil
		},
	}

	cmd.Flags().StringVarP(&serverAddress, "server-address", "s", "127.0.0.1:8001", "Server address")
	cmd.Flags().StringSliceVarP(&localAddresses, "local-address", "l", []string{}, "Local address to expose")

	return cmd
}

type proxyclient struct {
	net.Listener

	logger  logr.Logger
	tcpAddr *net.TCPAddr
}

func (p *proxyclient) Run() {
	for {
		rconn, err := p.Listener.Accept()
		if err != nil {
			p.logger.Error(err, "accept error")
			if errors.Is(err, api.ErrListenerStopped) {
				break
			}
			continue
		}

		lconn, err := net.DialTCP("tcp", nil, p.tcpAddr)
		if err != nil {
			p.logger.Error(err, "could not dial")
			rconn.Close()
			continue
		}

		p.logger.V(3).Info("start proxying", "client", rconn.RemoteAddr(), "server", lconn.RemoteAddr())

		go proxy.New(rconn, lconn).Start()
	}
}
