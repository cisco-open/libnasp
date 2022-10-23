# Copyright (c) 2021, and 2022 Cisco and/or its affiliates. All rights reserved.

FROM golang:1.18.0 as build

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
COPY experimental/heimdall/go.mod experimental/heimdall/go.mod
COPY experimental/heimdall/go.sum experimental/heimdall/go.sum

RUN if [ -n "${GITHUB_ACCESS_TOKEN}" ]; then \
      git config --global url."https://${GITHUB_ACCESS_TOKEN}:@github.com/".insteadOf "https://github.com/"; \
    fi

# hadolint ignore=DL3059
RUN go mod download

COPY . /go/src/app/
WORKDIR /go/src/app/experimental/heimdall

RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /go/bin/app


FROM gcr.io/distroless/static:nonroot

COPY --from=build /go/bin/app /
#ENV GIN_MODE=release
CMD ["/app"]
