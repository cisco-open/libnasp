#!/bin/bash

set -euo pipefail

function log() {
    echo -e "\n>>> ${1}\n"
}

DIRECTORY=`dirname $(readlink -f $0)`

log "creating kind cluster"
kind create cluster --wait 5m --config ${DIRECTORY}/kind.yaml

log "setup and update helm repositories"
helm repo add metallb https://metallb.github.io/metallb
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm repo update

log "install metallb"
helm install -n metallb-system --create-namespace metallb metallb/metallb --wait
kubectl apply -f ${DIRECTORY}/metallb-config.yaml

log "install istio"
helm install --create-namespace --namespace=istio-system istio-operator banzaicloud-stable/istio-operator --wait
kubectl apply --namespace istio-system -f ${DIRECTORY}/istio-controlplane.yaml

log "waiting for istio controlplane to be available"
kubectl wait --namespace istio-system \
                --for=jsonpath='{.status.status}'="Available" \
                --timeout=120s \
                icp/icp-v115x

log "install heimdall"
kubectl create namespace heimdall
kubectl label namespace heimdall istio.io/rev=icp-v115x.istio-system
helm install -n heimdall heimdall ${DIRECTORY}/../experimental/heimdall/charts/heimdall --wait

log "install echo service for testing"
kubectl create namespace testing
kubectl label namespace testing istio.io/rev=icp-v115x.istio-system
kubectl apply --namespace testing -f ${DIRECTORY}/echo-service.yaml
