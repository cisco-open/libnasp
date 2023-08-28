# Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
ARG BASE_IMAGE=istio/proxyv2
ARG BASE_IMAGE_TAG=1.15.3

FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

COPY kafka_msg_trace_filter.wasm /etc/istio/extensions/