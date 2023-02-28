# Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
ARG GO_VERSION=1.19.6

FROM golang:${GO_VERSION}-alpine3.17 AS builder

RUN apk add --update --no-cache ca-certificates~=20220614 make~=4.3 git~=2.38 curl~=7.87

ARG BUILD_DIR /go/src/app
RUN mkdir -p /${BUILD_DIR}

# Copy the go module manifests
COPY ./go.mod /${BUILD_DIR}/
COPY ./go.sum /${BUILD_DIR}/

# Copy the Bifrost go module manifests
COPY components/bifrost/go.mod /${BUILD_DIR}/components/bifrost/go.mod
COPY components/bifrost/go.sum /${BUILD_DIR}/components/bifrost/go.sum

WORKDIR /${BUILD_DIR}/components/bifrost

# hadolint ignore=DL3059
RUN go mod download

COPY . /${BUILD_DIR}

RUN CGO_ENABLED=0 BUILD_DIR='' BINARY_NAME=app make build-release

# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest
COPY --from=builder /app /app
USER nobody:nobody
ENTRYPOINT ["/app"]
