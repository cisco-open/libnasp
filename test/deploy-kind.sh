#!/bin/bash

set -euo pipefail

DIRECTORY=`dirname $(readlink -f $0)`

echo "creating kind cluster"
kind create cluster --wait 5m --config ${DIRECTORY}/kind.yaml

echo "setup and update helm repositories"
helm repo add metallb https://metallb.github.io/metallb
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm repo update

echo "install metallb"
helm install -n metallb-system --create-namespace metallb metallb/metallb --wait
kubectl apply -f ${DIRECTORY}/metallb-config.yaml

echo "install istio"
helm install --create-namespace --namespace=istio-system istio-operator banzaicloud-stable/istio-operator --wait
kubectl apply --namespace istio-system -f ${DIRECTORY}/istio-controlplane.yaml

echo "waiting for istio controlplane to be available"
kubectl wait --namespace istio-system \
                --for=jsonpath='{.status.status}'="Available" \
                --timeout=120s \
                icp/icp-v115x

echo "install heimdall"
kubectl create namespace heimdall
kubectl label namespace heimdall istio.io/rev=icp-v115x.istio-system
helm install -n heimdall heimdall ${DIRECTORY}/../experimental/heimdall/charts/heimdall --wait

echo "install echo service for testing"
kubectl create namespace testing
kubectl label namespace testing istio.io/rev=icp-v115x.istio-system
kubectl apply --namespace testing -f ${DIRECTORY}/echo-service.yaml
