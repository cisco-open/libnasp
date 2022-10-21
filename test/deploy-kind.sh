#!/bin/bash

set -euo pipefail

DIRECTORY=`dirname $(readlink -f $0)`

kind create cluster --wait 5m --config ${DIRECTORY}/kind.yaml

helm repo add metallb https://metallb.github.io/metallb
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm repo update

helm install -n metallb-system --create-namespace metallb metallb/metallb --wait
kubectl apply -f ${DIRECTORY}/metallb-config.yaml

helm install --create-namespace --namespace=istio-system istio-operator banzaicloud-stable/istio-operator --wait

kubectl apply --namespace istio-system -f ${DIRECTORY}/istio-controlplane.yaml

while [ "$(kubectl get icp -n istio-system icp-v115x -o jsonpath='{.status.status}')" != "Available" ];
do
    sleep 2
done

kubectl create namespace heimdall
kubectl label namespace heimdall istio.io/rev=icp-v115x.istio-system
helm install -n heimdall heimdall ${DIRECTORY}/../experimental/heimdall/charts/heimdall --wait
