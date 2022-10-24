#!/bin/bash

set -euo pipefail

DIRECTORY=`dirname $(readlink -f $0)`
IMAGE="${IMAGE:-heimdall:test}"
KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-nasp-test-cluster}

docker build "${DIRECTORY}/.." -f "${DIRECTORY}/../experimental/heimdall/Dockerfile" -t "${IMAGE}"

kind load docker-image -n ${KIND_CLUSTER_NAME} "${IMAGE}"
