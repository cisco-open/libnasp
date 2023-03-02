include ../../common.mk

.DEFAULT_GOAL := help

BUILD_DIR ?= build
BINARY_NAME ?= heimdall-server
BUILD_PACKAGE ?= ./cmd/server
VERSION ?= $(shell (git symbolic-ref -q --short HEAD || git describe --tags --exact-match) | tr "/" "-")

DOCKER_IMAGE ?= cisco-open/heimdall
DOCKER_TAG ?= ${VERSION}

CONTROLLER_TOOLS_VERSION ?= v0.9.2
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CGO_ENABLED ?= 0

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

##@ Build

.PHONY: build
build:
ifeq (${VERBOSE}, 1)
	go env
endif
	@$(eval GENERATED_BINARY_NAME = ${BINARY_NAME})
	@$(if $(strip ${BINARY_NAME_SUFFIX}),$(eval GENERATED_BINARY_NAME = ${BINARY_NAME}-$(subst $(eval) ,-,$(strip ${BINARY_NAME_SUFFIX}))),)
	go build ${GOARGS} -tags "${GOTAGS}" -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${GENERATED_BINARY_NAME} ${BUILD_PACKAGE}


.PHONY: build-webhook
build-webhook: generate fmt vet webhook-binary ## Build webhook binary with generate.

.PHONY: build-server
build-server: fmt vet server-binary ## Build server binary with generate.

.PHONY: webhook-binary
webhook-binary: LDFLAGS += -w
webhook-binary: BINARY_NAME = heimdall-webhook
webhook-binary: BUILD_PACKAGE = ./cmd/webhook
webhook-binary: build ## Build webhook binary.

.PHONY: server-binary
server-binary: LDFLAGS += -w
server-binary: BINARY_NAME = heimdall-server
server-binary: BUILD_PACKAGE = ./cmd/server
server-binary: build ## Build server binary.

docker: export GOOS = linux
docker-build: ## Build docker image.
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} -f Dockerfile --progress=plain ../..

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
