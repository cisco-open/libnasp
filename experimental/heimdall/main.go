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

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
	"github.com/cisco-open/nasp/pkg/environment"
)

func init() {
	cluster_registry.AddToScheme(scheme.Scheme)
}

var clusterID = os.Getenv("NASP_CLUSTER_ID")
var istioVersion = os.Getenv("NASP_ISTIO_VERSION")
var istioRevision = os.Getenv("NASP_ISTIO_REVISION")

var ErrClientNotFound = errors.New("client not found in database")
var ErrClusterIDNotFound = errors.New("clusterID not found")
var ClientDatabaseConfigMap = types.NamespacedName{Namespace: "heimdall", Name: "client-database"}

type ConfigRequest struct {
	ClientID     string `binding:"required"`
	ClientSecret string `binding:"required"`
}

type Client struct {
	ClientID     string
	ClientSecret string
	WorkloadName string
	PodNamespace string
	Network      string
	MeshID       string
}

type ClientDatabase interface {
	Lookup(ClientID string) (*Client, error)
}

type ConfigMapClientDatabase struct {
	c client.Client
}

func NewConfigMapClientDatabase() (ClientDatabase, error) {
	kubeconfig := config.GetConfigOrDie()
	c, err := client.New(kubeconfig, client.Options{})
	if err != nil {
		return nil, err
	}

	var configMap corev1.ConfigMap
	err = c.Get(context.Background(), ClientDatabaseConfigMap, &configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			configMap.Name = ClientDatabaseConfigMap.Name
			configMap.Namespace = ClientDatabaseConfigMap.Namespace
			err = c.Create(context.Background(), &configMap)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, err
		}
	}

	return &ConfigMapClientDatabase{c: c}, nil
}

func (db *ConfigMapClientDatabase) Lookup(ClientID string) (*Client, error) {
	var configMap corev1.ConfigMap
	err := db.c.Get(context.Background(), ClientDatabaseConfigMap, &configMap)
	if err != nil {
		return nil, err
	}

	if clientData, ok := configMap.Data[ClientID]; ok {
		var client Client
		err = json.Unmarshal([]byte(clientData), &client)
		if err != nil {
			return nil, err
		}
		return &client, nil
	}

	return nil, ErrClientNotFound
}

func main() {
	clientDb, err := NewConfigMapClientDatabase()
	if err != nil {
		panic(err)
	}

	if clusterID == "" {
		clusterID, err = getClusterIDFromRegistry()
		if err != nil {
			panic(err)
		}
	}

	r := gin.Default()
	r.POST("/config", func(c *gin.Context) {
		var configRequest ConfigRequest
		if err := c.ShouldBindJSON(&configRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("client with id %s is requesting config", configRequest.ClientID)

		client, err := clientDb.Lookup(configRequest.ClientID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if client.ClientSecret != configRequest.ClientSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// TODO generate a unique config for a given client
		var e IstioCAClientConfigAndEnvironment

		e.CAClientConfig, err = istio_ca.GetIstioCAClientConfig(clusterID, istioRevision)
		if err != nil {
			c.AbortWithError(500, err)
		}

		e.Environment.Type = "sidecar"
		e.Environment.InstanceIPs = []string{c.ClientIP()}
		e.Environment.WorkloadName = client.WorkloadName
		e.Environment.PodName = client.WorkloadName + "-" + configRequest.ClientID
		e.Environment.PodNamespace = client.PodNamespace
		e.Environment.Network = client.Network
		e.Environment.MeshID = client.MeshID
		e.Environment.IstioVersion = istioVersion
		e.Environment.SearchDomains = []string{"svc.cluster.local", "cluster.local"}
		e.Environment.Labels = map[string]string{
			"security.istio.io/tlsMode":           "istio",
			"pod-template-hash":                   "efefefef",
			"service.istio.io/canonical-revision": "latest",
			"istio.io/rev":                        e.CAClientConfig.Revision,
			"topology.istio.io/network":           e.Environment.Network,
			"k8s-app":                             client.WorkloadName,
			"service.istio.io/canonical-name":     client.WorkloadName,
			"app":                                 client.WorkloadName,
			"version":                             "v1",
		}

		c.JSON(http.StatusOK, e)
	})

	r.Run()
}

func getClusterIDFromRegistry() (string, error) {
	config, err := clientconfig.GetConfig()
	if err != nil {
		return "", err
	}

	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return "", err
	}

	var clusters cluster_registry.ClusterList
	err = k8sClient.List(context.Background(), &clusters)
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

type IstioCAClientConfigAndEnvironment struct {
	CAClientConfig istio_ca.IstioCAClientConfig
	Environment    environment.IstioEnvironment
}
