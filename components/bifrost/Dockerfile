# Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
ARG GO_VERSION=1.19.6

FROM golang:${GO_VERSION}-alpine3.17 AS builder

# hadolint ignore=DL3018
RUN apk add --update --no-cache ca-certificates make git curl

ARG GITHUB_ACCESS_TOKEN

ARG GOPROXY="https://proxy.golang.org,direct"
ENV GOPROXY="${GOPROXY}"
ENV GOPRIVATE='github.com/cisco-open,github.com/banzaicloud'
ENV GONOPROXY='gopkg.in,go.uber.org'
ENV GOFLAGS="-mod=readonly"

WORKDIR /go/src/app/components/bifrost

# Copy the go module manifests
COPY ./go.mod /go/src/app/
COPY ./go.sum /go/src/app/

# Copy the Bifrost go module manifests
COPY components/bifrost/go.mod /go/src/app/components/bifrost
COPY components/bifrost/go.sum /go/src/app/components/bifrost

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod go mod download

COPY . /go/src/app

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 BUILD_DIR='' BINARY_NAME=app make build-release


# Use distroless as minimal base image
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app /bifrost

USER 65532:65532

ENTRYPOINT ["/bifrost"]
