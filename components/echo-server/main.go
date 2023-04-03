// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"

	"github.com/cisco-open/nasp/pkg/istio"
)

var (
	httpServerListenAddress = ":8080"
	tcpServerListenAddress  = ":9000"
	tcpServerResponsePrefix = "hello"
)

var httpResponseTemplate = `Hostname: {{ getenv "HOSTNAME" "N/A" }}

Pod Information:
{{- $podName := getenv "POD_NAME" "" }}
{{- if ne $podName "" }}
	node name:	{{ getenv "NODE_NAME" "N/A" }}
	pod name:	{{ getenv "POD_NAME" "N/A" }}
	pod namespace:	{{ getenv "POD_NAMESPACE" "N/A" }}
	pod IP:	{{ getenv "POD_IP" "N/A" }}
{{- else }}
	-no pod information available-
{{- end }}

Request Information:
	client_address={{ .RemoteAddress }}
	method={{ .Method }}
	real path={{ .RequestURI }}
	query={{ .Query }}
	request_version={{ .HTTPVersion }}
	request_scheme={{ .Scheme }}
	request_url={{ .RequestURL }}

Request Headers:
{{- range $k, $v := .Headers }}
	{{ $k }}={{ $v }}
{{- end }}

Request Body:
{{ .Body }}
`

type HTTPRequest struct {
	RemoteAddress string
	Method        string
	RequestURI    string
	RequestURL    string
	Query         string
	HTTPVersion   string
	Scheme        string
	Headers       map[string]string
	Body          string
}

func init() {
	klog.InitFlags(nil)
	flag.Parse()
}

func main() {
	ctx := context.Background()

	logger := klog.Background()
	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig
	istioHandlerConfig.Enabled = false
	iih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, logger)
	if err != nil {
		panic(err)
	}

	if err := iih.Run(ctx); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer(ctx, iih, logger); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := tcpServer(ctx, iih, logger); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func httpServer(ctx context.Context, iih istio.IstioIntegrationHandler, logger logr.Logger) error {
	tpl, err := template.New("response").Funcs(template.FuncMap{
		"getenv": func(name, def string) string {
			if v := os.Getenv(name); v != "" {
				return v
			}

			return def
		},
	}).Parse(httpResponseTemplate)
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Logger())
	r.Any("/*path", func(c *gin.Context) {
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			headers[strings.ToLower(k)] = strings.Join(v, ", ")
		}

		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
		}

		c.Request.URL.Host = c.Request.Host
		c.Request.URL.Scheme = scheme

		req := HTTPRequest{
			RemoteAddress: c.Request.RemoteAddr,
			Method:        c.Request.Method,
			RequestURI:    c.Request.RequestURI,
			RequestURL:    c.Request.URL.String(),
			Query:         c.Request.URL.RawQuery,
			HTTPVersion:   fmt.Sprintf("%d.%d", c.Request.ProtoMajor, c.Request.ProtoMinor),
			Scheme:        scheme,
			Headers:       headers,
			Body:          string(body),
		}

		if err := tpl.Execute(c.Writer, req); err != nil {
			_ = c.Error(err)
			c.Abort()
		}
	})

	logger.Info("HTTP server listening", "address", httpServerListenAddress)

	return iih.ListenAndServe(ctx, httpServerListenAddress, r)
}

func tcpServer(ctx context.Context, iih istio.IstioIntegrationHandler, logger logr.Logger) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", tcpServerListenAddress)
	if err != nil {
		return err
	}

	listener, err = iih.GetTCPListener(listener)
	if err != nil {
		return err
	}

	logger.Info("TCP server listening", "address", listener.Addr(), "prefix", tcpServerResponsePrefix)

	handleConnection := func(conn net.Conn, prefix string) {
		defer conn.Close()

		reader := bufio.NewReader(conn)
		for {
			// read client request data
			bytes, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				if err != io.EOF {
					logger.Error(err, "could not read data")
				}

				return
			}

			logger.V(1).Info("request", "content", string(bytes))

			// prepend prefix and send as response
			line := fmt.Sprintf("%s %s", prefix, bytes)

			logger.V(1).Info("response", "content", line)

			if _, err := conn.Write([]byte(line)); err != nil {
				logger.Error(err, "could not send response")
			}
		}
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err, "could not accept connection")

			continue
		}

		// pass an accepted connection to a handler
		go handleConnection(conn, tcpServerResponsePrefix)
	}
}
