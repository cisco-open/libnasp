#!/bin/bash

# Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

DIRECTORY=`dirname $(readlink -f $0)`
BUILD_IMAGE=${BUILD_IMAGE:-true}

function log() {
    echo -e "\n>>> ${1}\n"
}

function create_and_label_namespace() {
    if ! kubectl get namespace ${1} >/dev/null 2>&1; then
    kubectl create namespace ${1}
    fi
    kubectl label namespace ${1} istio.io/rev=${2} --overwrite
}

function create_sa() {
    if ! kubectl -n ${1} get sa ${2} >/dev/null 2>&1; then
        kubectl -n ${1} create sa ${2}
    fi
}

if ! kind get kubeconfig --name nasp-test-cluster &> /dev/null; then
    log "creating kind cluster"
    kind create cluster --wait 5m --config ${DIRECTORY}/kind.yaml
else
    log "kind cluster already exists"
fi

log "setup and update helm repositories"
helm repo add metallb https://metallb.github.io/metallb
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm repo update

log "install metallb"
helm upgrade --install -n metallb-system --create-namespace metallb metallb/metallb --wait
kubectl apply -f ${DIRECTORY}/metallb-config.yaml

log "install istio"
helm upgrade --install --create-namespace --namespace=istio-system istio-operator banzaicloud-stable/istio-operator --wait
kubectl apply --namespace istio-system -f ${DIRECTORY}/istio-controlplane.yaml

log "waiting for istio controlplane to be available"
while [ "$(kubectl get icp -n istio-system icp-v115x -o jsonpath='{.status.status}')" != "Available" ];
do
    sleep 2
done

if [ ${BUILD_IMAGE} == "true" ]; then
    log "build and load heimdall image"
    ${DIRECTORY}/../scripts/heimdall-image-build.sh
    ${DIRECTORY}/../scripts/heimdall-image-load.sh
fi

log "install heimdall"
create_and_label_namespace heimdall icp-v115x.istio-system
helm upgrade --install -n heimdall heimdall ${DIRECTORY}/../experimental/heimdall/charts/heimdall --wait --values ${DIRECTORY}/heimdall-values.yaml

log "install echo service for testing"
create_and_label_namespace testing icp-v115x.istio-system
kubectl apply --namespace testing -f ${DIRECTORY}/echo-service.yaml

log "waiting for echo service to be available"
kubectl wait -n testing deployment/echo --for condition=Available=True --timeout=90s

log "create external namespace"
create_and_label_namespace external icp-v115x.istio-system

log "create service accounts in namespace external"
for saName in ios-mobile android-mobile test-http test-tcp test-grpc; do
    create_sa external ${saName}
done
