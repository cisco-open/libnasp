# Copyright (c) 2021, and 2022 Cisco and/or its affiliates. All rights reserved.

FROM golang:1.20.0 as build-heimdall

ARG GITHUB_ACCESS_TOKEN

ARG GOPROXY="https://proxy.golang.org,direct"
ENV GOPROXY="${GOPROXY}"
ENV GOPRIVATE='github.com/cisco-open,github.com/banzaicloud'
ENV GONOPROXY='gopkg.in,go.uber.org'
ENV GOFLAGS="-mod=readonly"

WORKDIR /go/src/app

# Copy the Go Modules manifests
COPY ./go.mod /go/src/app/
COPY ./go.sum /go/src/app/
# Copy the Heimdall Go Modules manifests
COPY components/heimdall/heimdall/go.mod components/heimdall/heimdall/go.mod
COPY components/heimdall/heimdall/go.sum components/heimdall/heimdall/go.sum
# Copy the charts Go Modules manifests
COPY components/heimdall/deploy/charts/go.mod components/heimdall/deploy/charts/go.mod

# hadolint ignore=DL3059
RUN go mod download

COPY . /go/src/app/
WORKDIR /go/src/app/components/heimdall/heimdall

RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /go/bin/app


# Build the manager binary
FROM golang:1.18 as build-webhook

WORKDIR /workspace
# Copy the Go Modules manifests
COPY components/heimdall/go.mod go.mod
COPY components/heimdall/go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY components/heimdall/pkg pkg
COPY components/heimdall/main.go main.go

# Build
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

# Heimdall
COPY --from=build-heimdall /go/bin/app /
#ENV GIN_MODE=release
CMD ["/app"]

# Webhook
WORKDIR /
COPY --from=build-webhook /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
