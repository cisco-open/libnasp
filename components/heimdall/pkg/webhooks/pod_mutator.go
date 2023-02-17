// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
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

package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"wwwin-github.cisco.com/eti/nasp-webhook/pkg/k8sutil"

	"github.com/go-logr/logr"
)

type PodMutator struct {
	// config contains the configuration for the pod mutator
	config PodMutatorConfig

	// logger is the log interface to use inside the validator.
	logger logr.Logger

	// manager is responsible for handling the communication between the
	// validator and the Kubernetes API server.
	manager ctrl.Manager

	// decoder is responsible for decoding the webhook request into structured
	// data.
	decoder *admission.Decoder
}

type PodMutatorConfig struct {
	ClusterName    string
	MeshID         string
	IstioVersion   string
	IstioCAAddress string
	IstioNetwork   string
	IstioRevision  string
}

func NewPodMutator(config PodMutatorConfig, logger logr.Logger, manager ctrl.Manager) *PodMutator {
	return &PodMutator{
		config:  config,
		logger:  logger,
		manager: manager,
		decoder: nil,
	}
}

// Handle handles the mutating admission requests.
func (m *PodMutator) Handle(ctx context.Context, request admission.Request) admission.Response {
	m.logger.Info("mutating pod", "request", request)

	pod := &corev1.Pod{}

	err := m.decoder.Decode(request, pod)
	if err != nil {
		err = errors.Wrap(err, "decoding admission request as pod failed")

		m.logger.Error(err, "could not mutate pod", "request", request)

		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Namespace == "" {
		pod.Namespace = request.Namespace
	}

	cms := &corev1.ConfigMapList{}
	cmLabels := map[string]string{
		"istio.io/config": "true",
		"istio.io/rev":    m.config.IstioRevision,
	}
	err = m.manager.GetClient().List(ctx, cms, client.InNamespace(pod.Namespace), client.MatchingLabels(cmLabels))
	if err != nil {
		m.logger.Error(err, "could not get configmaps", "namespace", pod.Namespace, "labels", cmLabels)
		return admission.Errored(http.StatusBadRequest, err)
	}
	if len(cms.Items) == 0 {
		err = errors.NewWithDetails("istio ca configmap is not found")
		m.logger.Error(err, "", "namespace", pod.Namespace, "labels", cmLabels)
		return admission.Errored(http.StatusBadRequest, err)
	}
	istioCAConfigmap := cms.Items[0]

	m.setAnnotations(pod)
	m.setLabels(pod)
	m.setEnvs(pod, request)
	m.setVolumes(pod, istioCAConfigmap)
	m.setVolumeMounts(pod)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	m.logger.Info("pod mutation successful", "request", request)

	return admission.PatchResponseFromRaw(request.Object.Raw, marshaledPod)
}

func (m *PodMutator) InjectDecoder(decoder *admission.Decoder) error {
	m.decoder = decoder

	return nil
}

func (m *PodMutator) setVolumeMounts(pod *corev1.Pod) {
	for i, container := range pod.Spec.Containers {
		if container.VolumeMounts == nil {
			container.VolumeMounts = []corev1.VolumeMount{}
		}

		container.VolumeMounts = append(container.VolumeMounts, []corev1.VolumeMount{
			{
				Name:      "istiod-ca-cert",
				MountPath: "/var/run/secrets/istio",
			},
			{
				Name:      "istio-token",
				MountPath: "/var/run/secrets/tokens",
			},
		}...)

		pod.Spec.Containers[i] = container
	}
}

func (m *PodMutator) setVolumes(pod *corev1.Pod, istioCAConfigmap corev1.ConfigMap) {
	if pod.Spec.Volumes == nil {
		pod.Spec.Volumes = []corev1.Volume{}
	}

	defaultMode := int32(420)
	expirationSecond := int64(43200)

	pod.Spec.Volumes = append(pod.Spec.Volumes, []corev1.Volume{
		{
			Name: "istio-token",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					DefaultMode: &defaultMode,
					Sources: []corev1.VolumeProjection{
						{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								Audience:          "istio-ca",
								ExpirationSeconds: &expirationSecond,
								Path:              "istio-token",
							},
						},
					},
				},
			},
		},
		{
			Name: "istiod-ca-cert",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: istioCAConfigmap.GetName(),
					},
				},
			},
		},
	}...)
}

func (m *PodMutator) getPodOwnerAndWorkloadName(pod *corev1.Pod, _ admission.Request) (owner string, workload string) {
	deploy, typeMeta := k8sutil.GetDeployMetaFromPod(pod)

	owner = fmt.Sprintf("kubernetes://apis/%s/namespaces/%s/%ss/%s", typeMeta.APIVersion, deploy.GetNamespace(), strings.ToLower(typeMeta.Kind), deploy.GetName())

	workload = deploy.GetName()

	return
}

func (m *PodMutator) setAnnotations(pod *corev1.Pod) {
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations["sidecar.istio.io/inject"] = "false"
	pod.Annotations["prometheus.io/path"] = "/stats/prometheus"
	pod.Annotations["prometheus.io/port"] = "15090"
	pod.Annotations["prometheus.io/scrape"] = "true"
}

func (m *PodMutator) setLabels(pod *corev1.Pod) {
	firstValue := func(values ...string) string {
		for _, v := range values {
			if v != "" {
				return v
			}
		}

		return ""
	}

	deploy, _ := k8sutil.GetDeployMetaFromPod(pod)
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	if _, ok := pod.Labels["security.istio.io/tlsMode"]; !ok {
		pod.Labels["security.istio.io/tlsMode"] = "istio"
	}
	if _, ok := pod.Labels["service.istio.io/canonical-name"]; !ok {
		pod.Labels["service.istio.io/canonical-name"] = firstValue(pod.Labels["app.kubernetes.io/name"], pod.Labels["app"], deploy.GetName())
	}
	if _, ok := pod.Labels["service.istio.io/canonical-revision"]; !ok {
		pod.Labels["service.istio.io/canonical-revision"] = firstValue(pod.Labels["app.kubernetes.io/version"], pod.Labels["version"], "latest")
	}

	pod.Labels["istio.io/rev"] = m.config.IstioRevision
	pod.Labels["topology.istio.io/network"] = m.config.IstioNetwork
}

func (m *PodMutator) setEnvs(pod *corev1.Pod, request admission.Request) {
	owner, workload := m.getPodOwnerAndWorkloadName(pod, request)

	containerNames := []string{}
	for _, container := range pod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}

	podLabels := []string{}
	for k, v := range pod.GetLabels() {
		podLabels = append(podLabels, fmt.Sprintf("%s:%s", k, v))
	}

	podIPs := []string{}
	for _, ip := range pod.Status.PodIPs {
		podIPs = append(podIPs, ip.IP)
	}

	for i, container := range pod.Spec.Containers {
		if container.Ports == nil {
			container.Ports = []corev1.ContainerPort{}
		}
		container.Ports = append(container.Ports, corev1.ContainerPort{
			ContainerPort: 15090,
			Name:          "http-envoy-prom",
			Protocol:      corev1.ProtocolTCP,
		})
		if container.Env == nil {
			container.Env = []corev1.EnvVar{}
		}
		container.Env = append(container.Env, []corev1.EnvVar{
			{
				Name:  "NASP_TYPE",
				Value: "sidecar",
			},
			{
				Name: "NASP_POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "NASP_POD_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name:  "NASP_POD_OWNER",
				Value: owner,
			},
			{
				Name: "NASP_POD_SERVICE_ACCOUNT",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "spec.serviceAccountName",
					},
				},
			},
			{
				Name:  "NASP_WORKLOAD_NAME",
				Value: workload,
			},
			{
				Name:  "NASP_APP_CONTAINERS",
				Value: strings.Join(containerNames, ","),
			},
			{
				Name:  "NASP_INSTANCE_IP",
				Value: strings.Join(podIPs, ","),
			},
			{
				Name:  "NASP_LABELS",
				Value: strings.Join(podLabels, ","),
			},
			{
				Name:  "NASP_NETWORK",
				Value: m.config.IstioNetwork,
			},
			{
				Name:  "NASP_SEARCH_DOMAINS",
				Value: fmt.Sprintf("%s.svc.cluster.local,svc.cluster.local", pod.GetNamespace()),
			},
			{
				Name:  "NASP_CLUSTER_ID",
				Value: m.config.ClusterName,
			},
			{
				Name:  "NASP_DNS_DOMAIN",
				Value: "cluster.local",
			},
			{
				Name:  "NASP_MESH_ID",
				Value: m.config.MeshID,
			},
			{
				Name:  "NASP_ISTIO_CA_ADDR",
				Value: m.config.IstioCAAddress,
			},
			{
				Name:  "NASP_ISTIO_VERSION",
				Value: m.config.IstioVersion,
			},
			{
				Name:  "NASP_ISTIO_REVISION",
				Value: m.config.IstioRevision,
			},
		}...)
		pod.Spec.Containers[i] = container
	}
}
