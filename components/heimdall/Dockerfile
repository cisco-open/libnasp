# Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
ARG GO_VERSION=1.19.6

FROM golang:${GO_VERSION}-alpine3.17 AS builder

# hadolint ignore=DL3018
RUN apk add --update --no-cache bash ca-certificates make git curl

ARG GITHUB_ACCESS_TOKEN

ARG GOPROXY="https://proxy.golang.org,direct"
ENV GOPROXY="${GOPROXY}"
ENV GOPRIVATE='github.com/cisco-open,github.com/banzaicloud'
ENV GONOPROXY='gopkg.in,go.uber.org'
ENV GOFLAGS="-mod=readonly"

WORKDIR /go/src/app/components/heimdall

# Copy the go modules manifests
COPY ./go.mod /go/src/app
COPY ./go.sum /go/src/app

# Copy the heimdall go modules manifests
COPY ./components/heimdall/go.mod /go/src/app/components/heimdall
COPY ./components/heimdall/go.sum /go/src/app/components/heimdall

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod go mod download

COPY . /go/src/app/

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 make webhook-binary \
    && CGO_ENABLED=0 make server-binary


# Use distroless as minimal base image
FROM gcr.io/distroless/static:nonroot

# Heimdall server
COPY --from=builder /go/src/app/components/heimdall/build/heimdall-server  /heimdall-server

# Heimdall Webhook
COPY --from=builder /go/src/app/components/heimdall/build/heimdall-webhook  /heimdall-webhook

WORKDIR /
USER 65532:65532

ENTRYPOINT ["/heimdall-webhook"]
