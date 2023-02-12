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

//nolint:goerr113
package tcp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	"golang.org/x/net/nettest"

	"github.com/banzaicloud/proxy-wasm-go-host/runtime/wazero"
	"github.com/cisco-open/nasp/pkg/environment"
	"github.com/cisco-open/nasp/pkg/istio/filters"
	"github.com/cisco-open/nasp/pkg/proxywasm"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	"github.com/cisco-open/nasp/pkg/proxywasm/tcp"
)

func TestNetConn(t *testing.T) {
	t.Parallel()

	nettest.TestConn(t, pipemaker())
}

func TestWrappedNetConn(t *testing.T) {
	t.Parallel()

	nettest.TestConn(t, wrappedpipemaker())
}

func BenchmarkNetConn(b *testing.B) {
	b.Run("plain", func(b *testing.B) {
		benchmark(b, pipemaker())
	})
	b.Run("wrapped", func(b *testing.B) {
		benchmark(b, wrappedpipemaker())
	})
}

func benchmark(b *testing.B, piper nettest.MakePipe) {
	b.Helper()

	c1, c2, stop, _ := piper()
	defer stop()

	content := bytes.Repeat([]byte("."), 1024)
	readbuff := make([]byte, len(content))

	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errs := make(chan error, 1)

		wg.Add(1)
		go func() {
			defer wg.Done()

			_, _ = c1.Write(content)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			l, err := c2.Read(readbuff)
			if err != nil {
				errs <- err
				return
			}

			if !bytes.Equal(content, readbuff[:l]) {
				err = errors.New("transmitted content differs")
			}

			errs <- err
		}()

		wg.Wait()

		err := <-errs
		if err != nil {
			b.Fatal(err)
		}
	}
}

func pipemaker() nettest.MakePipe {
	return func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2 = net.Pipe()
		stop = func() {
			c1.Close()
			c2.Close()
		}

		return
	}
}

var clientSH = getstream("wazero", "clientStream")
var serverSH = getstream("wazero", "serverStream")

func wrappedpipemaker() nettest.MakePipe {
	return func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2 = net.Pipe()

		serverStream, err := serverSH.NewStream(api.ListenerDirectionInbound)
		if err != nil {
			return nil, nil, nil, err
		}

		clientStream, err := clientSH.NewStream(api.ListenerDirectionOutbound)
		if err != nil {
			return nil, nil, nil, err
		}

		clientStream.Set("upstream.negotiated_protocol", "istio-peer-exchange")
		serverStream.Set("upstream.negotiated_protocol", "istio-peer-exchange")

		c1 = tcp.NewWrappedConn(c1, clientStream)
		c2 = tcp.NewWrappedConn(c2, serverStream)

		stop = func() {
			c1.Close()
			c2.Close()
		}

		return
	}
}

func getstream(runtime string, prefix string) api.StreamHandler {
	for _, env := range []string{
		prefix + "_NASP_TYPE=sidecar",
		fmt.Sprintf("%s_NASP_POD_NAME=%s-alpine-efefefef-f5wwf", prefix, prefix),
		prefix + "_NASP_POD_NAMESPACE=default",
		prefix + "_NASP_WORKLOAD_NAME=alpine",
		prefix + "_NASP_INSTANCE_IP=10.20.4.75",
		prefix + "_NASP_ISTIO_VERSION=1.13.5",
	} {
		p := strings.Split(env, "=")
		if len(p) != 2 {
			continue
		}
		os.Setenv(p[0], p[1])
	}

	e, err := environment.GetIstioEnvironment(prefix + "_NASP_")
	if err != nil {
		panic(err)
	}

	logger := logr.Discard()

	runtimeCreators := proxywasm.NewRuntimeCreatorStore()
	runtimeCreators.Set("wazero", func() api.WasmRuntime {
		return wazero.NewVM(context.Background(), wazero.VMWithLogger(logger))
	})
	baseContext := proxywasm.GetBaseContext(prefix)
	baseContext.Set("node", e.GetNodePropertiesFromEnvironment())

	vms := proxywasm.NewVMStore(runtimeCreators, logger)
	pm := proxywasm.NewWasmPluginManager(vms, baseContext, logger)
	sh, err := proxywasm.NewStreamHandler(pm,
		[]api.WasmPluginConfig{
			{
				Name:   "tcp-metadata-exchange",
				RootID: "tcp-metadata-exchange",
				VMConfig: api.WasmVMConfig{
					Runtime: runtime,
					ID:      "",
					Code:    proxywasm.NewFileDataSource(filters.Filters, "tcp-metadata-exchange-filter.wasm"),
				},
				Configuration: api.JsonnableMap{},
				InstanceCount: 1,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	return sh
}
