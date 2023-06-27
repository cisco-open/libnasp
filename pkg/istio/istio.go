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

package istio

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/openzipkin/zipkin-go"

	"github.com/cisco-open/nasp/pkg/istio/tracing"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/pborman/uuid"
	"github.com/prometheus/client_golang/prometheus"
	prometheus_push "github.com/prometheus/client_golang/prometheus/push"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"

	"github.com/banzaicloud/proxy-wasm-go-host/runtime/wazero"
	"github.com/cisco-open/nasp/pkg/ca"
	istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
	"github.com/cisco-open/nasp/pkg/environment"
	"github.com/cisco-open/nasp/pkg/istio/discovery"
	"github.com/cisco-open/nasp/pkg/istio/filters"
	itcp "github.com/cisco-open/nasp/pkg/istio/tcp"
	k8slabels "github.com/cisco-open/nasp/pkg/k8s/labels"
	"github.com/cisco-open/nasp/pkg/network"
	"github.com/cisco-open/nasp/pkg/network/listener"
	"github.com/cisco-open/nasp/pkg/proxywasm"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	pwgrpc "github.com/cisco-open/nasp/pkg/proxywasm/grpc"
	pwhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
	"github.com/cisco-open/nasp/pkg/proxywasm/middleware"
	"github.com/cisco-open/nasp/pkg/proxywasm/tcp"
)

var DefaultIstioIntegrationHandlerConfig = IstioIntegrationHandlerConfig{
	Enabled:        true,
	MetricsPath:    "/stats/prometheus",
	MetricsAddress: ":16090",
	UseTLS:         true,

	IstioCAConfigGetter: IstioCAConfigGetterAuto,
}

type IstioCAConfigGetterFunc func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error)

var (
	IstioCAConfigGetterRemote = func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
		return istio_ca.GetIstioCAClientConfig(e.ClusterID, e.IstioRevision)
	}
	IstioCAConfigGetterLocal = func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
		return istio_ca.GetIstioCAClientConfigFromLocal(e.ClusterID, e.IstioCAAddress)
	}
	IstioCAConfigGetterHeimdall = func(ctx context.Context, heimdallURL, authorizationToken string, version string) IstioCAConfigGetterFunc {
		return func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
			c, err := istio_ca.GetIstioCAClientConfigFromHeimdall(ctx, heimdallURL, authorizationToken, version)
			if err != nil {
				return istio_ca.IstioCAClientConfig{}, err
			}

			e.Override(c.Environment)

			e.ClusterID = c.CAClientConfig.ClusterID
			e.IstioRevision = c.CAClientConfig.Revision
			e.IstioCAAddress = c.CAClientConfig.CAEndpointSAN
			e.IstioVersion = c.Environment.IstioVersion

			return c.CAClientConfig, nil
		}
	}
	IstioCAConfigGetterAuto = func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
		if heimdallURL := os.Getenv("HEIMDALL_URL"); heimdallURL != "" {
			c := IstioCAConfigGetterHeimdall(context.Background(), heimdallURL, os.Getenv("HEIMDALL_AUTH_TOKEN"), os.Getenv("NASP_APP_VERSION"))

			return c(e)
		}

		fe := func(filename string) bool {
			info, err := os.Stat(filename)
			if os.IsNotExist(err) {
				return false
			}
			return !info.IsDir()
		}

		if fe(istio_ca.K8sSATrustworthyJWTFileName) {
			return IstioCAConfigGetterLocal(e)
		}

		return IstioCAConfigGetterRemote(e)
	}
)

type IstioIntegrationHandlerConfig struct {
	Enabled             bool
	MetricsPath         string
	MetricsAddress      string
	PushgatewayConfig   *PushgatewayConfig
	UseTLS              bool
	ServerFilters       []api.WasmPluginConfig
	ClientFilters       []api.WasmPluginConfig
	IstioCAConfigGetter IstioCAConfigGetterFunc
	DefaultWASMRuntime  string
}

type PushgatewayConfig struct {
	Address           string
	PushInterval      time.Duration
	UseUniqueIDLabel  bool
	UniqueIDLabelName string
}

const (
	defaultPushGatewayUniqueLabelName = "nasp_instance_id"
	defaultWASMRuntime                = "wazero"
)

func (c *PushgatewayConfig) SetDefaults() {
	if c.PushInterval == 0 {
		c.PushInterval = 5 * time.Second
	}
	if c.UniqueIDLabelName == "" {
		c.UniqueIDLabelName = defaultPushGatewayUniqueLabelName
	}
}

func (c *IstioIntegrationHandlerConfig) SetDefaults() {
	if c.MetricsPath == "" {
		c.MetricsPath = DefaultIstioIntegrationHandlerConfig.MetricsPath
	}
	if c.MetricsAddress == "" {
		c.MetricsAddress = DefaultIstioIntegrationHandlerConfig.MetricsAddress
	}
	if c.IstioCAConfigGetter == nil {
		c.IstioCAConfigGetter = IstioCAConfigGetterAuto
	}
	if c.PushgatewayConfig != nil {
		c.PushgatewayConfig.SetDefaults()
	}
	if c.DefaultWASMRuntime == "" {
		c.DefaultWASMRuntime = defaultWASMRuntime
	}
}

type istioIntegrationHandler struct {
	config IstioIntegrationHandlerConfig
	logger logr.Logger

	caClient      ca.Client
	metricHandler api.MetricHandler
	metricsPusher *prometheus_push.Pusher
	pluginManager api.WasmPluginManager
	environment   *environment.IstioEnvironment

	zipkinTracer *zipkin.Tracer

	discoveryClient discovery.DiscoveryClient
}

type IstioIntegrationHandler interface {
	Run(context.Context) error
	GetGRPCDialOptions() ([]grpc.DialOption, error)
	ListenAndServe(ctx context.Context, listenAddress string, handler http.Handler) error
	GetHTTPTransport(transport http.RoundTripper) (http.RoundTripper, error)
	GetTCPListener(l net.Listener) (net.Listener, error)
	GetTCPDialer() (itcp.Dialer, error)
	GetDiscoveryClient() discovery.DiscoveryClient
}

func NewIstioIntegrationHandler(config *IstioIntegrationHandlerConfig, logger logr.Logger) (IstioIntegrationHandler, error) {
	defaultConfig := DefaultIstioIntegrationHandlerConfig
	if config == nil {
		config = &defaultConfig
	}
	config.SetDefaults()
	s := &istioIntegrationHandler{
		config: *config,
		logger: logger,
	}

	e, err := environment.GetIstioEnvironment("NASP_")
	if err != nil {
		return nil, err
	}

	if !e.Enabled && !config.Enabled {
		logger.Info("NASP_ENABLED env var is not set to true, istio integration handler is disabled")
		return &passthroughIstioIntegrationHandler{}, nil
	}

	s.environment = e

	registry := prometheus.NewRegistry()
	s.metricHandler = proxywasm.NewPrometheusMetricHandler(registry, logger)
	baseContext := proxywasm.GetBaseContext("root")
	baseContext.Set("metric.handler", s.metricHandler)
	if config.UseTLS {
		if istioCAConfig, err := config.IstioCAConfigGetter(e); err != nil {
			return nil, err
		} else {
			s.caClient = istio_ca.NewIstioCAClient(istioCAConfig, logger)
		}
	}

	if config.PushgatewayConfig == nil {
		if e.Labels == nil {
			e.Labels = make(map[string]string)
		}
		e.Labels[k8slabels.NASPMonitoringLabel] = "true"
		e.Labels[k8slabels.NASPWorkloadUID] = uuid.New()
		if strings.Contains(config.MetricsAddress, ":") {
			if p := strings.Split(config.MetricsAddress, ":"); len(p) > 1 {
				e.Labels[k8slabels.NASPMonitoringPortLabel] = p[len(p)-1]
			}
		}

		e.Labels[k8slabels.NASPMonitoringPathLabel] = base64.RawURLEncoding.EncodeToString([]byte(config.MetricsPath))
	}

	baseContext.Set("node", e.GetNodePropertiesFromEnvironment())

	runtimeCreators := proxywasm.NewRuntimeCreatorStore()
	runtimeCreators.Set("wazero", func() api.WasmRuntime {
		return wazero.NewVM(context.Background(), wazero.VMWithLogger(logger))
	})
	if getWasmtimeRuntime != nil {
		runtimeCreators.Set("wasmtime", func() api.WasmRuntime {
			return getWasmtimeRuntime(context.Background(), logger)
		})
	}
	if getWasmerRuntime != nil {
		runtimeCreators.Set("wasmer", func() api.WasmRuntime {
			return getWasmerRuntime
		})
	}
	vms := proxywasm.NewVMStore(runtimeCreators, logger)

	s.pluginManager = proxywasm.NewWasmPluginManager(vms, baseContext, logger)

	s.discoveryClient = discovery.NewXDSDiscoveryClient(e, s.caClient, s.logger.WithName("xds-discovery"))

	if config.PushgatewayConfig != nil {
		if config.PushgatewayConfig.Address == "" {
			return nil, errors.New("if Pushgateway is enabled then the PushgatewayConfig.Address config is required")
		}
		// use node id as job name to ensure that each instance of the app has its metrics stored under different key
		// in pushgateway to avoid instances of the same app overriding each other's metrics
		r := regexp.MustCompile(`\.|/|%`) // need to replace . / and % as job name is a URL segment
		jobName := r.ReplaceAllString(e.GetNodeID(), "_")

		transport, err := s.GetHTTPTransport(http.DefaultTransport)
		if err != nil {
			return nil, err
		}

		httpClient := &http.Client{
			Transport: transport,
		}

		s.metricsPusher = createMetricsPusher(config.PushgatewayConfig, jobName, httpClient, registry)
	}

	if s.environment.ZipkinAddress != "" {
		transport, err := s.GetHTTPTransport(http.DefaultTransport)
		if err != nil {
			return nil, err
		}

		s.zipkinTracer, err = tracing.SetupZipkinTracing(s.environment, &http.Client{
			Transport: transport,
		})
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (h *istioIntegrationHandler) GetDiscoveryClient() discovery.DiscoveryClient {
	return h.discoveryClient
}

func (h *istioIntegrationHandler) NewStreamHandler(filters ...api.WasmPluginConfig) (api.StreamHandler, error) {
	return proxywasm.NewStreamHandler(h.pluginManager, filters)
}

func (h *istioIntegrationHandler) GetHTTPTransport(transport http.RoundTripper) (http.RoundTripper, error) {
	streamHandler, err := h.NewStreamHandler(h.defaultClientFilters()...)
	if err != nil {
		return nil, errors.Wrap(err, "could not get stream handler")
	}

	logger := h.logger.WithName("http-transport")
	tp := NewIstioHTTPRequestTransport(transport, h.caClient, h.discoveryClient, logger, h.zipkinTracer)

	if h.zipkinTracer != nil {
		tp, err = zipkinhttp.NewTransport(h.zipkinTracer, zipkinhttp.RoundTripper(tp))
		if err != nil {
			return nil, errors.Wrap(err, "could not get zipkin tracer")
		}
	}

	httpTransport := pwhttp.NewHTTPTransport(tp, streamHandler, logger)

	if h.zipkinTracer != nil {
		httpTransport.AddMiddleware(tracing.NewZipkinHTTPClientTracingMiddleware(h.zipkinTracer, h.environment))
	}

	httpTransport.AddMiddleware(middleware.NewEnvoyHTTPHandlerMiddleware())
	httpTransport.AddMiddleware(NewIstioHTTPHandlerMiddleware())

	return httpTransport, nil
}

func (h *istioIntegrationHandler) GetGRPCDialOptions() ([]grpc.DialOption, error) {
	streamHandler, err := h.NewStreamHandler(h.defaultClientFilters()...)
	if err != nil {
		return nil, errors.Wrap(err, "could not get stream handler")
	}

	grpcDialer := pwgrpc.NewGRPCDialer(h.caClient, streamHandler, h.discoveryClient, h.logger)

	if h.zipkinTracer != nil {
		grpcDialer.AddMiddleware(tracing.NewZipkinGRPCClientTracingMiddleware(h.zipkinTracer, h.environment))
	}

	grpcDialer.AddMiddleware(middleware.NewEnvoyHTTPHandlerMiddleware())
	grpcDialer.AddMiddleware(NewIstioHTTPHandlerMiddleware())

	return []grpc.DialOption{
		grpc.WithContextDialer(grpcDialer.Dial),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // tls credentials will be set at dial
		grpc.WithUnaryInterceptor(grpcDialer.RequestInterceptor),
	}, nil
}

func (h *istioIntegrationHandler) ListenAndServe(ctx context.Context, listenAddress string, handler http.Handler) error {
	filters := h.config.ServerFilters
	if len(filters) == 0 {
		filters = h.defaultServerFilters()
	}

	streamHandler, err := h.NewStreamHandler(filters...)
	if err != nil {
		return errors.Wrap(err, "could not get stream handler")
	}

	httpHandler := pwhttp.NewHandler(handler, streamHandler, api.ListenerDirectionInbound)

	if h.zipkinTracer != nil {
		httpHandler.AddMiddleware(tracing.NewZipkinHTTPTracingMiddleware(h.zipkinTracer, h.environment))
	}

	httpHandler.AddMiddleware(middleware.NewEnvoyHTTPHandlerMiddleware())
	httpHandler.AddMiddleware(NewIstioHTTPHandlerMiddleware())

	server := network.WrapHTTPServer(&http.Server{
		Handler: httpHandler,
	})

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(h.caClient.GetCAPem())

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequestClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := h.caClient.GetCertificate("", time.Hour*24)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs: certPool,
		NextProtos: []string{
			"h2",
			"http/1.1",
			"http/1.0",
		},
	}

	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	go func() {
		_, err = h.discoveryClient.GetListenerProperties(ctx, listenAddress, func(lp discovery.ListenerProperties) {
			h.logger.Info("got new listener properties", "address", listenAddress, "properties", lp)
			ul := server.GetUnifiedListener()
			if ul == nil {
				return
			}

			if lp.UseTLS() {
				if lp.Permissive() {
					ul.SetTLSMode(listener.TLSModePermissive)
				} else {
					ul.SetTLSMode(listener.TLSModeStrict)
				}
			} else {
				ul.SetTLSMode(listener.TLSModeDisabled)
			}
			if lp.IsClientCertificateRequired() {
				ul.SetTLSClientAuthMode(tls.RequireAnyClientCert)
			} else {
				ul.SetTLSClientAuthMode(tls.RequestClientCert)
			}
		})
		if err != nil {
			h.logger.Error(err, "could not find listener properties", "address", listenAddress)
		}
	}()

	return server.ServeWithTLSConfig(ln, tlsConfig)
}

func (h *istioIntegrationHandler) Run(ctx context.Context) error {
	if err := h.discoveryClient.Connect(ctx); err != nil {
		return err
	}

	if h.config.PushgatewayConfig == nil {
		go h.RunMetricsServer(ctx)
	} else {
		// TODO(@waynz0r): check this
		go func() {
			_ = h.RunMetricsPusher(ctx)
		}()
	}

	return nil
}

func (h *istioIntegrationHandler) RunMetricsServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc(h.config.MetricsPath, h.metricHandler.HTTPHandler().ServeHTTP)
	s := http.Server{Addr: h.config.MetricsAddress, Handler: mux, ReadHeaderTimeout: time.Second * 60}
	go func() {
		_ = s.ListenAndServe()
	}()

	<-ctx.Done()

	_ = s.Shutdown(ctx)
}

func (h *istioIntegrationHandler) defaultClientFilters() []api.WasmPluginConfig {
	return []api.WasmPluginConfig{
		{
			Name:   "istio-metadata-exchange",
			RootID: "",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "metadata-exchange-filter.wasm"),
			},
			Configuration: api.JsonnableMap{
				"max_peer_cache_size": 10000,
			},
			InstanceCount: 1,
		},
		{
			Name:   "istio-stats-outbound",
			RootID: "stats_outbound",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "stats-filter.wasm"),
			},
			Configuration: api.JsonnableMap{},
			InstanceCount: 1,
		},
	}
}

func (h *istioIntegrationHandler) defaultServerFilters() []api.WasmPluginConfig {
	return []api.WasmPluginConfig{
		{
			Name:   "istio-metadata-exchange",
			RootID: "",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "metadata-exchange-filter.wasm"),
			},
			Configuration: api.JsonnableMap{
				"max_peer_cache_size": 10000,
			},
			InstanceCount: 1,
		},
		{
			Name:   "istio-stats-inbound",
			RootID: "stats_inbound",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "stats-filter.wasm"),
			},
			Configuration: api.JsonnableMap{},
			InstanceCount: 1,
		},
	}
}

func createMetricsPusher(config *PushgatewayConfig, jobName string, httpClient prometheus_push.HTTPDoer, metricsGatherers ...prometheus.Gatherer) *prometheus_push.Pusher {
	pusher := prometheus_push.New(config.Address, jobName)
	if config.UseUniqueIDLabel {
		pusher.Grouping(config.UniqueIDLabelName, uuid.New())
	}
	pusher.Client(httpClient)

	for _, g := range metricsGatherers {
		pusher.Gatherer(g)
	}

	return pusher
}
func (h *istioIntegrationHandler) RunMetricsPusher(ctx context.Context) error {
	if h.metricsPusher == nil {
		return nil
	}

	ticker := time.NewTicker(h.config.PushgatewayConfig.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// delete the metrics group from Pushgateway upon termination otherwise the
			// earlier published metrics as Pushgateway doesn't have a TTL for metrics
			err := h.metricsPusher.Delete()
			if err != nil {
				h.logger.Error(err, "deleting metric group from Pushgateway failed")
			}
			return ctx.Err()
		case <-ticker.C:
			err := h.metricsPusher.PushContext(ctx)
			if err != nil {
				h.logger.Error(err, "pushing metrics to Pushgateway failed")
			}
		}
	}
}

func (h *istioIntegrationHandler) GetTCPListener(l net.Listener) (net.Listener, error) {
	streamHandler, err := h.NewStreamHandler(h.defaultTCPServerFilters()...)
	if err != nil {
		return nil, errors.Wrap(err, "could not get stream handler")
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(h.caClient.GetCAPem())

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequestClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := h.caClient.GetCertificate(h.environment.GetSpiffeID(), time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
		NextProtos: []string{
			"istio-peer-exchange",
			"istio",
		},
	}

	l = listener.NewUnifiedListener(l, network.WrapTLSConfig(tlsConfig), listener.TLSModePermissive,
		listener.UnifiedListenerWithTLSConnectionCreator(network.CreateTLSServerConn),
		listener.UnifiedListenerWithConnectionWrapper(func(c net.Conn) net.Conn {
			return network.WrapConnection(c)
		}))

	return tcp.WrapListener(l, streamHandler), nil
}

func (h *istioIntegrationHandler) GetTCPDialer() (itcp.Dialer, error) {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(h.caClient.GetCAPem())

	tlsConfig := &tls.Config{
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := h.caClient.GetCertificate(h.environment.GetSpiffeID(), time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
		NextProtos: []string{
			"istio-peer-exchange",
			"istio",
		},
	}

	streamHandler, err := h.NewStreamHandler(h.defaultTCPClientFilters()...)
	if err != nil {
		return nil, errors.Wrap(err, "could not get stream handler")
	}

	return itcp.NewTCPDialer(streamHandler, tlsConfig, h.discoveryClient), nil
}

func (h *istioIntegrationHandler) defaultTCPClientFilters() []api.WasmPluginConfig {
	return []api.WasmPluginConfig{
		{
			Name:   "tcp-metadata-exchange",
			RootID: "tcp-metadata-exchange",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "tcp-metadata-exchange-filter.wasm"),
			},
			Configuration: api.JsonnableMap{},
			InstanceCount: 1,
		},
		{
			Name:   "istio-stats-outbound",
			RootID: "stats_outbound",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "stats-filter.wasm"),
			},
			Configuration: api.JsonnableMap{},
			InstanceCount: 1,
		},
	}
}

func (h *istioIntegrationHandler) defaultTCPServerFilters() []api.WasmPluginConfig {
	return []api.WasmPluginConfig{
		{
			Name:   "tcp-metadata-exchange",
			RootID: "tcp-metadata-exchange",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "tcp-metadata-exchange-filter.wasm"),
			},
			Configuration: api.JsonnableMap{},
			InstanceCount: 1,
		},
		{
			Name:   "istio-stats-inbound",
			RootID: "stats_inbound",
			VMConfig: api.WasmVMConfig{
				Runtime: h.config.DefaultWASMRuntime,
				ID:      "",
				Code:    proxywasm.NewFileDataSource(filters.Filters, "stats-filter.wasm"),
			},
			Configuration: api.JsonnableMap{},
			InstanceCount: 1,
		},
	}
}
