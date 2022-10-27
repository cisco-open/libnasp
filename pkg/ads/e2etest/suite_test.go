// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2etest

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"

	upda_type_v1 "github.com/cncf/xds/go/udpa/type/v1"
	envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	grpcv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	corsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	faultv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	grpc_statsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	httpwasmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	http_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/http_inspector/v3"
	original_dstv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_dst/v3"
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	http_connection_managerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcp_proxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	uuidv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/request_id/uuid/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test"
	. "github.com/onsi/ginkgo"
	ginkoconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/cisco-open/nasp/pkg/ads"
	adsconfig "github.com/cisco-open/nasp/pkg/ads/config"
	"github.com/cisco-open/nasp/pkg/ads/internal/endpoint"
	"github.com/cisco-open/nasp/pkg/ads/internal/listener"

	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"istio.io/api/envoy/config/filter/http/alpn/v2alpha1"
	"istio.io/api/envoy/config/filter/network/metadata_exchange"
)

var (
	testSuiteContext context.Context
)

const (
	testPollInterval = 1 * time.Second
	testPollDuration = 20 * time.Second
)

func TestEndToEnd(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	ginkoconfig.DefaultReporterConfig.SlowSpecThreshold = 360 // slow test threshold in seconds

	testSuiteContext = logr.NewContext(context.Background(), testr.NewWithOptions(t, testr.Options{LogTimestamp: true}))

	t.Parallel()

	RunSpecs(t, "End-to-End Test Suite")
}

//nolint:funlen,gocognit,maintidx
func createSnapshot(version string, config *envoy_admin_v3.ConfigDump) (*cache.Snapshot, error) {
	var (
		clusters []*envoy_config_cluster_v3.Cluster

		clusterLoadAssignments    []*envoy_config_endpoint_v3.ClusterLoadAssignment
		listeners                 []*envoy_config_listener_v3.Listener
		routeConfigurations       []*envoy_config_route_v3.RouteConfiguration
		scopedRouteConfigurations []*envoy_config_route_v3.ScopedRouteConfiguration
		secrets                   []*tlsv3.Secret
		virtualHosts              []*envoy_config_route_v3.VirtualHost
	)

	for _, config := range config.GetConfigs() {
		proto, err := config.UnmarshalNew()
		Expect(err).NotTo(HaveOccurred())

		switch configDump := proto.(type) {
		case *envoy_admin_v3.ClustersConfigDump:
			for _, activeCluster := range configDump.GetDynamicActiveClusters() {
				var cluster envoy_config_cluster_v3.Cluster
				err = activeCluster.GetCluster().UnmarshalTo(&cluster)
				Expect(err).NotTo(HaveOccurred())

				clusters = append(clusters, &cluster)
			}
			for _, warmingCluster := range configDump.GetDynamicWarmingClusters() {
				var cluster envoy_config_cluster_v3.Cluster
				err = warmingCluster.GetCluster().UnmarshalTo(&cluster)
				Expect(err).NotTo(HaveOccurred())

				clusters = append(clusters, &cluster)
			}
			for _, staticCluster := range configDump.GetStaticClusters() {
				var cluster envoy_config_cluster_v3.Cluster
				err = staticCluster.GetCluster().UnmarshalTo(&cluster)
				Expect(err).NotTo(HaveOccurred())

				clusters = append(clusters, &cluster)
			}
		case *envoy_admin_v3.EndpointsConfigDump:
			for _, endpointConfig := range configDump.GetDynamicEndpointConfigs() {
				var cla envoy_config_endpoint_v3.ClusterLoadAssignment
				err := endpointConfig.GetEndpointConfig().UnmarshalTo(&cla)
				Expect(err).NotTo(HaveOccurred())

				clusterLoadAssignments = append(clusterLoadAssignments, &cla)
			}
			for _, endpointConfig := range configDump.GetStaticEndpointConfigs() {
				var cla envoy_config_endpoint_v3.ClusterLoadAssignment
				err := endpointConfig.GetEndpointConfig().UnmarshalTo(&cla)
				Expect(err).NotTo(HaveOccurred())

				clusterLoadAssignments = append(clusterLoadAssignments, &cla)
			}
		case *envoy_admin_v3.ListenersConfigDump:
			for _, listenerConfig := range configDump.GetDynamicListeners() {
				var listener envoy_config_listener_v3.Listener
				if listenerConfig.GetActiveState() != nil {
					err := listenerConfig.GetActiveState().GetListener().UnmarshalTo(&listener)
					Expect(err).NotTo(HaveOccurred())

					listeners = append(listeners, &listener)
				}
				if listenerConfig.GetDrainingState() != nil {
					err := listenerConfig.GetDrainingState().GetListener().UnmarshalTo(&listener)
					Expect(err).NotTo(HaveOccurred())

					listeners = append(listeners, &listener)
				}
				if listenerConfig.GetWarmingState() != nil {
					err := listenerConfig.GetWarmingState().GetListener().UnmarshalTo(&listener)
					Expect(err).NotTo(HaveOccurred())

					listeners = append(listeners, &listener)
				}
			}
			for _, listenerConfig := range configDump.GetStaticListeners() {
				var listener envoy_config_listener_v3.Listener
				err := listenerConfig.GetListener().UnmarshalTo(&listener)
				Expect(err).NotTo(HaveOccurred())

				listeners = append(listeners, &listener)
			}
		case *envoy_admin_v3.ScopedRoutesConfigDump:
			for _, rConfig := range configDump.GetDynamicScopedRouteConfigs() {
				for _, scopedRouteConfig := range rConfig.GetScopedRouteConfigs() {
					var routeConfig envoy_config_route_v3.ScopedRouteConfiguration
					err := scopedRouteConfig.UnmarshalTo(&routeConfig)
					Expect(err).NotTo(HaveOccurred())

					scopedRouteConfigurations = append(scopedRouteConfigurations, &routeConfig)
					virtualHosts = append(virtualHosts, routeConfig.GetRouteConfiguration().GetVirtualHosts()...)
				}
			}
			for _, rConfig := range configDump.GetInlineScopedRouteConfigs() {
				for _, scopedRouteConfig := range rConfig.GetScopedRouteConfigs() {
					var routeConfig envoy_config_route_v3.ScopedRouteConfiguration
					err := scopedRouteConfig.UnmarshalTo(&routeConfig)
					Expect(err).NotTo(HaveOccurred())

					scopedRouteConfigurations = append(scopedRouteConfigurations, &routeConfig)
					virtualHosts = append(virtualHosts, routeConfig.GetRouteConfiguration().GetVirtualHosts()...)
				}
			}
		case *envoy_admin_v3.RoutesConfigDump:
			for _, rConfig := range configDump.GetStaticRouteConfigs() {
				var routeConfig envoy_config_route_v3.RouteConfiguration
				err := rConfig.GetRouteConfig().UnmarshalTo(&routeConfig)
				Expect(err).NotTo(HaveOccurred())

				routeConfigurations = append(routeConfigurations, &routeConfig)
				virtualHosts = append(virtualHosts, routeConfig.GetVirtualHosts()...)
			}
			for _, rConfig := range configDump.GetDynamicRouteConfigs() {
				var routeConfig envoy_config_route_v3.RouteConfiguration
				err := rConfig.GetRouteConfig().UnmarshalTo(&routeConfig)
				Expect(err).NotTo(HaveOccurred())

				routeConfigurations = append(routeConfigurations, &routeConfig)
				virtualHosts = append(virtualHosts, routeConfig.GetVirtualHosts()...)
			}
		case *envoy_admin_v3.SecretsConfigDump:
			for _, secretConfig := range configDump.GetDynamicActiveSecrets() {
				var secret tlsv3.Secret
				err := secretConfig.GetSecret().UnmarshalTo(&secret)
				Expect(err).NotTo(HaveOccurred())

				secrets = append(secrets, &secret)
			}
			for _, secretConfig := range configDump.GetDynamicWarmingSecrets() {
				var secret tlsv3.Secret
				err := secretConfig.GetSecret().UnmarshalTo(&secret)
				Expect(err).NotTo(HaveOccurred())

				secrets = append(secrets, &secret)
			}
			for _, secretConfig := range configDump.GetStaticSecrets() {
				var secret tlsv3.Secret
				err := secretConfig.GetSecret().UnmarshalTo(&secret)
				Expect(err).NotTo(HaveOccurred())

				secrets = append(secrets, &secret)
			}
		}
	}

	var clasReferencingClusters []types.Resource
	for _, clusterName := range endpoint.GetClusterLoadAssignmentReferences(clusters) {
		for _, cla := range clusterLoadAssignments {
			if cla.GetClusterName() == clusterName {
				clasReferencingClusters = append(clasReferencingClusters, cla)
			}
		}
	}

	var routeConfigResources, scopedRouteConfigResources []types.Resource

	for _, routeConfigName := range listener.GetRouteReferences(listeners) {
		for _, rc := range routeConfigurations {
			if rc.GetName() == routeConfigName {
				routeConfigResources = append(routeConfigResources, rc)
			}
		}
		for _, rc := range scopedRouteConfigurations {
			if rc.GetName() == routeConfigName {
				scopedRouteConfigResources = append(scopedRouteConfigResources, rc)
			}
		}
	}

	clusterResources := make([]types.Resource, 0, len(clusters))
	for _, cluster := range clusters {
		clusterResources = append(clusterResources, cluster)
	}

	vhResources := make([]types.Resource, 0, len(virtualHosts))
	for _, vh := range virtualHosts {
		vhResources = append(vhResources, vh)
	}
	listenerResources := make([]types.Resource, 0, len(listeners))
	for _, listener := range listeners {
		listenerResources = append(listenerResources, listener)
	}
	secretResources := make([]types.Resource, 0, len(secrets))
	for _, secret := range secrets {
		secretResources = append(secretResources, secret)
	}

	snapshot, err := cache.NewSnapshot(version, map[resource.Type][]types.Resource{
		resource.EndpointType:    clasReferencingClusters,
		resource.ClusterType:     clusterResources,
		resource.RouteType:       routeConfigResources,
		resource.ScopedRouteType: scopedRouteConfigResources,
		resource.VirtualHostType: vhResources,
		resource.ListenerType:    listenerResources,
		resource.SecretType:      secretResources,
	})

	return snapshot, err
}

var _ = BeforeSuite(func() {
	// initialize used packages by referencing a type from them
	_ = http_connection_managerv3.HttpConnectionManager{}
	_ = tls_inspectorv3.TlsInspector{}
	_ = http_inspectorv3.HttpInspector{}
	_ = original_dstv3.OriginalDst{}
	_ = grpc_statsv3.FilterConfig{}
	_ = tcp_proxyv3.TcpProxy{}
	_ = routerv3.Router{}
	_ = httpv3.HttpProtocolOptions{}
	_ = metadata_exchange.MetadataExchange{}
	_ = v2alpha1.FilterConfig{}
	_ = upda_type_v1.TypedStruct{}
	_ = grpcv3.TcpGrpcAccessLogConfig{}
	_ = httpwasmv3.Wasm{}
	_ = faultv3.HTTPFault{}
	_ = corsv3.Cors{}
	_ = uuidv3.UuidRequestIdConfig{}

	ctx, cancel := context.WithCancel(testSuiteContext)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	testSuiteContext = ctx

})

var _ = AfterSuite(func() {

})

var _ = Describe("The management server is running", func() {
	var (
		resourceConfigCache  cache.SnapshotCache
		managementServerPort uint
		done                 context.CancelFunc

		count     uint64 = 0
		adsClient ads.Client
	)

	BeforeEach(func() {
		atomic.AddUint64(&count, 1)
		// use unique port for management server per each test case in order to avoid port conflict
		managementServerPort = 18000 + uint(count)
		resourceConfigCache = cache.NewSnapshotCache(true, cache.IDHash{}, nil)
	})

	JustBeforeEach(func() {
		var config envoy_admin_v3.ConfigDump
		err := protojson.Unmarshal(configV0, &config)
		Expect(err).NotTo(HaveOccurred())

		snapshot, err := createSnapshot("v0", &config)
		Expect(err).NotTo(HaveOccurred())
		err = snapshot.Consistent()
		Expect(err).NotTo(HaveOccurred())

		var ctx context.Context
		ctx, done = context.WithCancel(testSuiteContext)

		srv := server.NewServer(ctx, resourceConfigCache, nil)
		go test.RunManagementServer(ctx, srv, managementServerPort)

		err = resourceConfigCache.SetSnapshot(ctx, "test-node-id", snapshot)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		done()
	})

	AssertADSClientAPI := func() {
		It("the ADS client API should work fine", func() {
			ctx, cancel := context.WithCancel(testSuiteContext)
			defer cancel()

			By("verify that correct listener properties are returned for :8080")
			resp, err := adsClient.GetListenerProperties(ctx, ":8080")
			Expect(err).NotTo(HaveOccurred())

			var listenerProps ads.ListenerProperties
			Eventually(func() error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case r := <-resp:
						listenerProps = r.ListenerProperties()
						return r.Error()
					}
				}
			}, testPollDuration, testPollInterval).Should(Succeed())

			Expect(listenerProps.UseTLS()).To(BeTrue())
			Expect(listenerProps.IsClientCertificateRequired()).To(BeTrue())
			Expect(listenerProps.Permissive()).To(BeTrue())

			By("verify that correct tcp client properties are returned for 10.10.42.110:12050")
			tcpResp, err := adsClient.GetTCPClientPropertiesByHost(ctx, "10.10.42.110:12050")
			Expect(err).NotTo(HaveOccurred())

			var tcpClientProps ads.ClientProperties
			Eventually(func() error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case r := <-tcpResp:
						tcpClientProps = r.ClientProperties()
						return r.Error()
					}
				}
			}, testPollDuration, testPollInterval).Should(Succeed())

			Expect(tcpClientProps.Permissive()).To(BeTrue())
			Expect(tcpClientProps.UseTLS()).To(BeTrue())
			Expect(tcpClientProps.ServerName()).To(Equal("outbound_.12050_._.echo.demo.svc.cluster.local"))
			Expect(tcpClientProps.Address()).To(Equal(&net.TCPAddr{IP: net.ParseIP("10.20.160.131"), Port: 8080}))

			By("verify that correct http client properties are returned for 10.10.42.110:80")
			httpResp, err := adsClient.GetHTTPClientPropertiesByHost(ctx, "10.10.42.110:80")
			Expect(err).NotTo(HaveOccurred())

			var httpClientProps ads.ClientProperties
			Eventually(func() error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case r := <-httpResp:
						httpClientProps = r.ClientProperties()
						return r.Error()
					}
				}
			}, testPollDuration, testPollInterval).Should(Succeed())

			Expect(httpClientProps.Permissive()).To(BeTrue())
			Expect(httpClientProps.UseTLS()).To(BeTrue())
			Expect(httpClientProps.ServerName()).To(Equal("outbound_.80_._.echo.demo.svc.cluster.local"))
			Expect(httpClientProps.Address()).To(Equal(&net.TCPAddr{IP: net.ParseIP("10.20.160.131"), Port: 8080}))

			// --- load config_v1.json which contains three echo endpoints in the demo namespace as the 'echo' deployment has been scaled up to three
			// replicas into Management server
			var config envoy_admin_v3.ConfigDump
			err = protojson.Unmarshal(configV1, &config)
			Expect(err).NotTo(HaveOccurred())

			snapshot, err := createSnapshot("v1", &config)
			Expect(err).NotTo(HaveOccurred())
			err = snapshot.Consistent()
			Expect(err).NotTo(HaveOccurred())

			err = resourceConfigCache.SetSnapshot(ctx, "test-node-id", snapshot)
			Expect(err).NotTo(HaveOccurred())

			By("verify that 3 endpoints are returned by tcp client properties for 10.10.42.110:12050")
			endpoints := make([]net.Addr, 0, 3)
			Eventually(func() []net.Addr {
				tcpResp, err = adsClient.GetTCPClientPropertiesByHost(ctx, "10.10.42.110:12050")
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() error {
					for {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case r := <-tcpResp:
							tcpClientProps = r.ClientProperties()
							return r.Error()
						}
					}
				}, testPollDuration, testPollInterval).Should(Succeed())

				Expect(tcpClientProps.Permissive()).To(BeTrue())
				Expect(tcpClientProps.UseTLS()).To(BeTrue())
				Expect(tcpClientProps.ServerName()).To(Equal("outbound_.12050_._.echo.demo.svc.cluster.local"))

				ep, err := tcpClientProps.Address()
				Expect(err).NotTo(HaveOccurred())

				endpoints = append(endpoints, ep)

				return endpoints
			}, 2*testPollDuration, testPollInterval).Should(ContainElements(
				&net.TCPAddr{IP: net.ParseIP("10.20.164.172"), Port: 8080},
				&net.TCPAddr{IP: net.ParseIP("10.20.166.13"), Port: 8080},
				&net.TCPAddr{IP: net.ParseIP("10.20.166.2"), Port: 8080},
			))

			By("verify that 3 endpoints are returned by http client properties for 10.10.42.110:80")
			endpoints = make([]net.Addr, 0, 3)
			Eventually(func() []net.Addr {
				httpResp, err = adsClient.GetHTTPClientPropertiesByHost(ctx, "10.10.42.110:80")
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() error {
					for {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case r := <-httpResp:
							httpClientProps = r.ClientProperties()
							return r.Error()
						}
					}
				}, testPollDuration, testPollInterval).Should(Succeed())

				Expect(httpClientProps.Permissive()).To(BeTrue())
				Expect(httpClientProps.UseTLS()).To(BeTrue())
				Expect(httpClientProps.ServerName()).To(Equal("outbound_.80_._.echo.demo.svc.cluster.local"))

				ep, err := httpClientProps.Address()
				Expect(err).NotTo(HaveOccurred())

				endpoints = append(endpoints, ep)
				return endpoints
			}, 2*testPollDuration, testPollInterval).Should(ContainElements(
				&net.TCPAddr{IP: net.ParseIP("10.20.164.172"), Port: 8080},
				&net.TCPAddr{IP: net.ParseIP("10.20.166.13"), Port: 8080},
				&net.TCPAddr{IP: net.ParseIP("10.20.166.2"), Port: 8080},
			))

			// --- load config_v2.json which contains a newly added workload named 'ngnix' in the demo namespace compared to config_v1
			err = protojson.Unmarshal(configV2, &config)
			Expect(err).NotTo(HaveOccurred())

			snapshot, err = createSnapshot("v2", &config)
			Expect(err).NotTo(HaveOccurred())
			err = snapshot.Consistent()
			Expect(err).NotTo(HaveOccurred())

			err = resourceConfigCache.SetSnapshot(ctx, "test-node-id", snapshot)
			Expect(err).NotTo(HaveOccurred())

			By("verify that correct http client properties are returned for the newly added HTTP workload listening at 10.10.104.42:80")
			httpResp, err = adsClient.GetHTTPClientPropertiesByHost(ctx, "10.10.104.42:80")
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case r := <-httpResp:
						httpClientProps = r.ClientProperties()
						return r.Error()
					}
				}
			}, testPollDuration, testPollInterval).Should(Succeed())

			Expect(httpClientProps.Permissive()).To(BeTrue())
			Expect(httpClientProps.UseTLS()).To(BeTrue())
			Expect(httpClientProps.ServerName()).To(Equal("outbound_.80_._.nginx-service.demo.svc.cluster.local"))
			Expect(httpClientProps.Address()).To(Equal(&net.TCPAddr{IP: net.ParseIP("10.20.177.35"), Port: 80}))

		})

	}

	Describe("the ADS client interacts with the management server using SotW protocol", func() {
		var (
			adsClientConfig *adsconfig.ClientConfig
		)
		BeforeEach(func() {
			clusterID := "waynz0r-0626-01"
			md, _ := structpb.NewStruct(map[string]interface{}{
				"NAMESPACE":     "demo",
				"MESH_ID":       "mesh1",
				"CLUSTER_ID":    clusterID,
				"NETWORK":       "network1",
				"ISTIO_VERSION": "1.13.5",
			})

			adsClientConfig = &adsconfig.ClientConfig{
				ManagementServerAddress: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: int(managementServerPort)},
				ClusterID:               clusterID,
				NodeInfo: &adsconfig.NodeInfo{
					Id:          "test-node-id",
					ClusterName: "echo.demo",
					Metadata:    md,
				},
				SearchDomains:         []string{"demo.svc.cluster.local", "svc.cluster.local", "cluster.local"},
				IncrementalXDSEnabled: false,
				ManagementServerCapabilities: &adsconfig.ManagementServerCapabilities{
					SotWWildcardSubscriptionSupported:          false,
					DeltaWildcardSubscriptionSupported:         false,
					ClusterDiscoveryServiceProvided:            true,
					EndpointDiscoveryServiceProvided:           true,
					ListenerDiscoveryServiceProvided:           true,
					RouteConfigurationDiscoveryServiceProvided: true,
					NameTableDiscoveryServiceProvided:          false,
				},
			}
		})
		JustBeforeEach(func() {
			var ctx context.Context
			ctx, done = context.WithCancel(testSuiteContext)

			var err error
			adsClient, err = ads.Connect(ctx, adsClientConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		AssertADSClientAPI()

		AfterEach(func() {
			done()
		})
	})

	Describe("the ADS client interacts with the management server using incremental xDS protocol", func() {
		var (
			adsClientConfig *adsconfig.ClientConfig
		)

		BeforeEach(func() {
			clusterID := "waynz0r-0626-01"
			md, _ := structpb.NewStruct(map[string]interface{}{
				"NAMESPACE":     "demo",
				"MESH_ID":       "mesh1",
				"CLUSTER_ID":    clusterID,
				"NETWORK":       "network1",
				"ISTIO_VERSION": "1.13.5",
			})

			adsClientConfig = &adsconfig.ClientConfig{
				ManagementServerAddress: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: int(managementServerPort)},
				ClusterID:               clusterID,
				NodeInfo: &adsconfig.NodeInfo{
					Id:          "test-node-id",
					ClusterName: "echo.demo",
					Metadata:    md,
				},
				SearchDomains:         []string{"demo.svc.cluster.local", "svc.cluster.local", "cluster.local"},
				IncrementalXDSEnabled: true,
				ManagementServerCapabilities: &adsconfig.ManagementServerCapabilities{
					SotWWildcardSubscriptionSupported:          false,
					DeltaWildcardSubscriptionSupported:         false,
					ClusterDiscoveryServiceProvided:            true,
					EndpointDiscoveryServiceProvided:           true,
					ListenerDiscoveryServiceProvided:           true,
					RouteConfigurationDiscoveryServiceProvided: true,
					NameTableDiscoveryServiceProvided:          false,
				},
			}
		})
		JustBeforeEach(func() {
			var ctx context.Context
			ctx, done = context.WithCancel(testSuiteContext)

			var err error
			adsClient, err = ads.Connect(ctx, adsClientConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			done()
		})

		AssertADSClientAPI()

	})
})
