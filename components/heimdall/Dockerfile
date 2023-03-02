# Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
ARG GO_VERSION=1.19.6

FROM golang:${GO_VERSION}-alpine3.17 AS builder

RUN apk add --update --no-cache bash=~5.2 ca-certificates~=20220614 make~=4.3 git~=2.38 curl~=7.88

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

# hadolint ignore=DL3059
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod go mod download

COPY . /go/src/app/

RUN CGO_ENABLED=0 make webhook-binary server-binary


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

# Heimdall server
COPY --from=builder /go/src/app/components/heimdall/bin/server  /heimdall-server

# Heimdall Webhook
COPY --from=builder /go/src/app/components/heimdall/bin/webhook  /heimdall-webhook

WORKDIR /
USER 65532:65532

ENTRYPOINT ["/heimdall-webhook"]
