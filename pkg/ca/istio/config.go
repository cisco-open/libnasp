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

package istio_ca

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/cisco-open/nasp/pkg/environment"
)

var (
	istioNamespace = "istio-system"
)

const (
	// K8sSATrustworthyJWTFileName is the token volume mount file name for k8s trustworthy jwt token.
	K8sSATrustworthyJWTFileName = "/var/run/secrets/tokens/istio-token"

	// K8sSAJWTFileName is the token volume mount file name for k8s jwt token.
	K8sSAJWTFileName = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	// The data name in the ConfigMap of each namespace storing the root cert of non-Kube CA.
	CACertPEMFileName = "/var/run/secrets/istio/root-cert.pem"
)

func GetIstioCAClientConfigFromLocal(clusterID string, endpointAddress string) (config IstioCAClientConfig, err error) {
	config.Token, err = os.ReadFile(K8sSATrustworthyJWTFileName)
	if err != nil {
		return config, err
	}

	config.CApem, err = os.ReadFile(CACertPEMFileName)
	if err != nil {
		return config, err
	}

	config.ClusterID = clusterID
	config.CAEndpoint = endpointAddress

	return config, nil
}

func GetIstioCAClientConfig(clusterID string, istioRevision string) (IstioCAClientConfig, error) {
	return GetIstioCAClientConfigWithKubeConfig(clusterID, istioRevision, nil)
}

func GetIstioCAClientConfigWithKubeConfig(clusterID string, istioRevision string, kubeConfig []byte) (IstioCAClientConfig, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return IstioCAClientConfig{}, err
	}

	var config *rest.Config
	var err error

	if kubeConfig != nil {
		config, err = clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	} else {
		config, err = clientconfig.GetConfig()
	}
	if err != nil {
		return IstioCAClientConfig{}, errors.WrapIf(err, "could not get k8s config")
	}

	cl, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return IstioCAClientConfig{}, errors.WrapIf(err, "could not create client")
	}

	svc, err := GetIstiodService(cl, istioRevision)
	if err != nil {
		return IstioCAClientConfig{}, errors.WrapIf(err, "could not get istiod service")
	}

	token, address, err := GetIMGWData(cl, config, scheme, istioRevision)
	if err != nil {
		return IstioCAClientConfig{}, errors.WrapIf(err, "could not get imgw data")
	}

	pem, err := GetIstioRootCAPEM(cl, istioRevision)
	if err != nil {
		return IstioCAClientConfig{}, errors.WrapIf(err, "could not get istio root ca pem")
	}

	return IstioCAClientConfig{
		CAEndpoint:    address + ":15012",
		CAEndpointSAN: fmt.Sprintf("%s.%s.svc", svc.GetName(), svc.GetNamespace()),
		Token:         token,
		CApem:         pem,
		ClusterID:     clusterID,
		Revision:      istioRevision,
	}, nil
}

func GetIstiodService(cl client.Client, istioRevision string) (*corev1.Service, error) {
	labels := map[string]string{
		"istio":        "istiod",
		"istio.io/rev": istioRevision,
	}

	services := &corev1.ServiceList{}
	err := cl.List(context.Background(), services, client.InNamespace(istioNamespace), client.MatchingLabels(labels))
	if err != nil {
		return nil, errors.WrapIf(err, "could not get mexp pods")
	}

	if len(services.Items) < 1 {
		return nil, errors.New("istiod service is not found")
	}

	return &services.Items[0], nil
}

func GetIMGWData(cl client.Client, config *rest.Config, scheme *runtime.Scheme, istioRevision string) (token []byte, address string, err error) {
	labels := map[string]string{
		"app":          "istio-meshexpansion-gateway",
		"istio.io/rev": istioRevision,
	}

	pods := &corev1.PodList{}
	err = cl.List(context.Background(), pods, client.InNamespace(istioNamespace), client.MatchingLabels(labels))
	if err != nil {
		err = errors.WrapIf(err, "could not get mexp pods")
		return
	}

	if len(pods.Items) < 1 {
		err = errors.New("mexp pod is not found")
		return
	}

	token, err = GetIstioTokenFromPod(config, scheme, pods.Items[0].Name, pods.Items[0].Namespace)
	if err != nil {
		err = errors.WrapIf(err, "could not get token from mexp pod")
		return
	}

	imgws := &unstructured.UnstructuredList{}
	imgws.SetAPIVersion("servicemesh.cisco.com/v1alpha1")
	imgws.SetKind("IstioMeshGatewayList")
	err = cl.List(context.Background(), imgws, client.InNamespace(istioNamespace), client.MatchingLabels(labels))
	if err != nil {
		err = errors.WrapIf(err, "could not list imgws")
		return
	}

	if len(imgws.Items) < 1 {
		err = errors.New("imgw is not found")
		return
	}

	status := imgws.Items[0].UnstructuredContent()["status"]
	if s, ok := status.(map[string]interface{}); ok {
		if val, ok := s["GatewayAddress"].([]interface{}); ok && len(val) > 0 {
			if addr, ok := val[0].(string); ok {
				address = addr
			}
		}
	}

	if address == "" {
		err = errors.New("imgw address is not found")
	}

	return token, address, err
}

func GetIstioRootCAPEM(cl client.Client, istioRevision string) ([]byte, error) {
	cms := &corev1.ConfigMapList{}
	err := cl.List(context.Background(), cms, client.InNamespace(istioNamespace), client.MatchingLabels(map[string]string{
		"istio.io/rev":    istioRevision,
		"istio.io/config": "true",
	}))
	if err != nil {
		return nil, errors.WrapIf(err, "could not get configmaps")
	}

	configmap := &corev1.ConfigMap{}
	err = cl.Get(context.Background(), client.ObjectKeyFromObject(&cms.Items[0]), configmap)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get istio root ca configmap")
	}

	return []byte(configmap.Data["root-cert.pem"]), nil
}

func GetIstioTokenFromPod(config *rest.Config, scheme *runtime.Scheme, name, namespace string) ([]byte, error) {
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}

	restClient, err := apiutil.RESTClientForGVK(gvk, false, config, serializer.NewCodecFactory(scheme))
	if err != nil {
		return nil, err
	}

	execReq := restClient.
		Post().
		Name(name).
		Namespace(namespace).
		Resource("pods").
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"cat", K8sSATrustworthyJWTFileName},
			Stdout:  true,
		}, runtime.NewParameterCodec(scheme))

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execReq.URL())
	if err != nil {
		return nil, fmt.Errorf("error while creating remote command executor: %w", err)
	}

	stdout := bytes.Buffer{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Tty:    false,
	})
	if err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

func GetIstioCAClientConfigFromHeimdall(heimdallURL, clientID, clientSecret string) (config IstioCAClientConfigAndEnvironment, err error) {
	body, err := json.Marshal(map[string]string{
		"ClientID":     clientID,
		"ClientSecret": clientSecret,
	})
	if err != nil {
		return config, err
	}

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	// TODO(@nandork): use ctx
	response, err := client.Post(heimdallURL, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return config, errors.WrapIf(err, "failed to communicate with Heimdall")
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		err = json.NewDecoder(response.Body).Decode(&config)
		if err != nil {
			return config, errors.WrapIf(err, "failed to decode Heimdall response")
		}

		return
	}

	return config, ConfigRetrievalError{Status: response.Status}
}

type IstioCAClientConfigAndEnvironment struct {
	CAClientConfig IstioCAClientConfig
	Environment    environment.IstioEnvironment
}

type ConfigRetrievalError struct {
	Status string
}

func (e ConfigRetrievalError) Error() string {
	return fmt.Sprintf("failed to get Istio CA config: %s", e.Status)
}
