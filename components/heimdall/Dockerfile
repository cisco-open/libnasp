# Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.

FROM golang:1.19.6 as builder

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
RUN go mod download

COPY . /go/src/app/

RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 make webhook-binary server-binary


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
