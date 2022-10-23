#!/bin/bash

set -euo pipefail

function log() {
    echo -e "\n>>> ${1}\n"
}

DIRECTORY=`dirname $(readlink -f $0)`

log "creating kind cluster"
if ! $(kind create cluster --wait 5m --config ${DIRECTORY}/kind.yaml); then
    echo ""
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
# until https://github.com/kubernetes/kubectl/issues/1236 is fixed,
# the icp needs to updated first, then the embedded status field appears
kubectl wait --namespace istio-system \
                --for=jsonpath='{.metadata.generation}'=2 \
                --timeout=120s \
                icp/icp-v115x

kubectl wait --namespace istio-system \
                --for=jsonpath='{.status.status}'="Available" \
                --timeout=120s \
                icp/icp-v115x

log "build and load heimdall image"
${DIRECTORY}/build-heimdall-image.sh

log "install heimdall"
if ! kubectl get namespace heimdall >/dev/null 2>&1; then
kubectl create namespace heimdall
fi
kubectl label namespace heimdall istio.io/rev=icp-v115x.istio-system --overwrite
helm upgrade --install -n heimdall heimdall ${DIRECTORY}/../experimental/heimdall/charts/heimdall --wait --values ${DIRECTORY}/heimdall-values.yaml

log "install echo service for testing"
if ! kubectl get namespace testing >/dev/null 2>&1; then
kubectl create namespace testing
fi
kubectl label namespace testing istio.io/rev=icp-v115x.istio-system
kubectl apply --namespace testing -f ${DIRECTORY}/echo-service.yaml
