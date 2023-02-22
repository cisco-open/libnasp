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

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
	istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
	"github.com/cisco-open/nasp/pkg/environment"
)

var (
	clusterID     = os.Getenv("NASP_CLUSTER_ID")
	istioVersion  = os.Getenv("NASP_ISTIO_VERSION")
	istioRevision = os.Getenv("NASP_ISTIO_REVISION")
)

var ErrClusterIDNotFound = errors.New("clusterID not found")

type server struct {
	mgr    manager.Manager
	logger logr.Logger
}

func New(mgr manager.Manager, logger logr.Logger) *server {
	return &server{
		mgr:    mgr,
		logger: logger,
	}
}

func (s *server) Run(ctx context.Context) error {
	if clusterID == "" {
		if _, err := s.getClusterIDFromRegistry(); err != nil {
			return err
		}
	}

	srv := s.run()

	<-ctx.Done()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (s *server) run() *http.Server {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {})

	r.POST("/config", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			err := errors.NewPlain("authorization header is missing")
			s.logger.Error(err, "authorization failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		jwtToken := strings.Split(authHeader, " ")
		if len(jwtToken) != 2 {
			err := errors.NewPlain("incorrectly formatted authorization header")
			s.logger.Error(err, "authorization failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		user, err := AuthenticateToken(c, s.mgr.GetClient(), jwtToken[1])
		if err != nil {
			s.logger.Error(err, "authorization failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		log := s.logger.WithValues("user", user.Name)
		log.Info("client requesting config")

		if user.ServiceAccountRef.Name == "" {
			user.ServiceAccountRef.Name = "default"
		}

		workloadGroup := &istionetworkingv1beta1.WorkloadGroup{}
		err = s.mgr.GetClient().Get(c, user.WorkloadGroupRef, workloadGroup)
		if err != nil {
			s.logger.Error(err, "could not get workload group")
			err = errors.WrapIf(err, "could not get workload group")
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		type IstioCAClientConfigAndEnvironment struct {
			CAClientConfig istio_ca.IstioCAClientConfig
			Environment    environment.IstioEnvironment
		}

		var e IstioCAClientConfigAndEnvironment

		e.CAClientConfig, err = istio_ca.GetIstioCAClientConfigWithKubeConfig(clusterID, istioRevision, nil, &user.ServiceAccountRef)
		if err != nil {
			s.logger.Error(err, "could not get ca client config")
			err = errors.WrapIf(err, "could not get ca client config")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		clientIP := c.ClientIP()
		if clientIP == "::1" {
			clientIP = "127.0.0.1"
		}

		e.Environment = environment.IstioEnvironment{
			Type:              "sidecar",
			PodNamespace:      workloadGroup.GetNamespace(),
			PodOwner:          fmt.Sprintf("kubernetes://apis/v1/namespaces/%s/workloadgroups/%s", user.WorkloadGroupRef.Namespace, workloadGroup.GetName()),
			PodServiceAccount: user.ServiceAccountRef.Name,
			WorkloadName:      workloadGroup.GetName(),
			AppContainers:     nil,
			InstanceIPs:       []string{clientIP},
			Labels: map[string]string{
				"security.istio.io/tlsMode":           "istio",
				"service.istio.io/canonical-revision": getLabelValueWithDefault(workloadGroup.Labels, []string{"service.istio.io/canonical-revision", "app.kubernetes.io/version", "version"}, "latest"),
				"istio.io/rev":                        e.CAClientConfig.Revision,
				"topology.istio.io/network":           workloadGroup.Spec.Template.GetNetwork(),
				"k8s-app":                             workloadGroup.GetName(),
				"service.istio.io/canonical-name":     getLabelValueWithDefault(workloadGroup.Labels, []string{"service.istio.io/canonical-name", "app.kubernetes.io/name", "app"}, workloadGroup.GetName()),
				"app":                                 workloadGroup.GetName(),
			},
			PlatformMetadata: nil,
			Network:          workloadGroup.Spec.Template.GetNetwork(),
			SearchDomains:    []string{"svc.cluster.local", "cluster.local"},
			ClusterID:        e.CAClientConfig.ClusterID,
			DNSDomain:        "cluster.local",
			MeshID:           e.CAClientConfig.MeshID,
			IstioCAAddress:   e.CAClientConfig.CAEndpoint,
			IstioVersion:     istioVersion,
			IstioRevision:    istioRevision,
		}

		c.JSON(http.StatusOK, e)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server: %s\n", err)
		}
	}()

	return srv
}

func (s *server) getClusterIDFromRegistry() (string, error) {
	var clusters cluster_registry.ClusterList
	err := s.mgr.GetClient().List(context.Background(), &clusters)
	if k8serrors.IsInvalid(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	for _, cluster := range clusters.Items {
		if cluster.Status.Type == cluster_registry.ClusterTypeLocal {
			return cluster.Name, nil
		}
	}

	return "", ErrClusterIDNotFound
}

func getLabelValueWithDefault(labels map[string]string, precedence []string, def string) string {
	for _, k := range precedence {
		if v, ok := labels[k]; ok {
			return v
		}
	}

	return def
}
