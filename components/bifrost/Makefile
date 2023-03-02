include ../../common.mk

BUILD_DIR ?= build
BINARY_NAME ?= $(shell basename $$PWD)
BUILD_PACKAGE ?= ./cmd/bifrost
VERSION ?= $(shell (git symbolic-ref -q --short HEAD || git describe --tags --exact-match) | tr "/" "-")

DOCKER_IMAGE ?= cisco-open/bifrost
DOCKER_TAG ?= ${VERSION}

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build a binary
ifeq (${VERBOSE}, 1)
	go env
endif
	@$(eval GENERATED_BINARY_NAME = ${BINARY_NAME})
	@$(if $(strip ${BINARY_NAME_SUFFIX}),$(eval GENERATED_BINARY_NAME = ${BINARY_NAME}-$(subst $(eval) ,-,$(strip ${BINARY_NAME_SUFFIX}))),)
	go build ${GOARGS} -tags "${GOTAGS}" -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${GENERATED_BINARY_NAME} ${BUILD_PACKAGE}

.PHONY: build-release
build-release: LDFLAGS += -w
build-release: build ## Build a binary without debug information

.PHONY: build-debug
build-debug: GOARGS += -gcflags "all=-N -l"
build-debug: BINARY_NAME_SUFFIX += debug
build-debug: build ## Build a binary with remote debugging capabilities

.PHONY: docker-build
docker: export GOOS = linux
docker-build: ## Build docker image.
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} -f Dockerfile --progress=plain ../..

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}
