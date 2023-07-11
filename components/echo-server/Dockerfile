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

WORKDIR /go/src/components/echo-server

# Copy the go modules manifests
COPY ./components/echo-server/go.mod /go/src/components/echo-server/go.mod
COPY ./components/echo-server/go.sum /go/src/components/echo-server/go.sum
COPY ./components/bifrost/go.mod /go/src/components/bifrost/go.mod
COPY ./components/bifrost/go.sum /go/src/components/bifrost/go.sum
COPY ./go.mod /go/src/go.mod
COPY ./go.sum /go/src/go.sum

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod go mod download

COPY . /go/src/

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 BUILD_DIR='' BINARY_NAME=app make build-release


# Use distroless as minimal base image
FROM gcr.io/distroless/static:nonroot

# server binary
COPY --from=builder /app  /server

WORKDIR /
USER 65532:65532

ENTRYPOINT ["/server"]
