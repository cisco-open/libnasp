# Copyright (c) 2021 Cisco and/or its affiliates. All rights reserved.
ARG BASE_IMAGE=033498657557.dkr.ecr.us-east-2.amazonaws.com/banzaicloud/istio-proxyv2
ARG BASE_IMAGE_TAG=v1.17.0-bzc.0

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

COPY kafka-message-parser.wasm /etc/istio/extensions/