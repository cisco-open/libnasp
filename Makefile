REPO_ROOT=$(shell git rev-parse --show-toplevel)
HEIMDALL_IMAGE ?= heimdall

include common.mk

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

heimdall-docker: ## Build heimdall docker container
	scripts/heimdall-image-build.sh

.PHONY: test
test: ## Run tests
	go test ./pkg/... \
    	-coverprofile cover.out \
    	-v \
    	-failfast \
    	-test.v \
    	-test.paniconexit0 \
    	-timeout 1h

.PHONY: tidy-all
tidy-all:	## go mod tidy all go modules
	./scripts/for_all_go_modules.sh --with-file Makefile -- make tidy

.PHONY: license-cache-all
license-cache-all: ${REPO_ROOT}/bin/licensei
	./scripts/for_all_go_modules.sh --with-file Makefile -- make license-cache

.PHONY: license-check-all
license-check-all: ${REPO_ROOT}/bin/licensei
	./scripts/for_all_go_modules.sh --with-file Makefile -- make license-check

.PHONY: fmt-all
fmt-all:	## go fmt all go modules
	./scripts/for_all_go_modules.sh --with-file Makefile -- make fmt

.PHONY: vet-all
vet-all:	## go vet all go modules
	./scripts/for_all_go_modules.sh --with-file Makefile -- make vet

.PHONY: lint-all
lint-all: ${REPO_ROOT}/bin/golangci-lint ## lint the whole repo
	./scripts/for_all_go_modules.sh --parallel 1 --with-file Makefile -- make lint

.PHONY: lint-fix-all
lint-fix-all: ${REPO_ROOT}/bin/golangci-lint ## lint --fix the whole repo
	./scripts/for_all_go_modules.sh --parallel 1 --with-file Makefile -- make lint-fix

.PHONY: mod-download-all
mod-download-all:	## go mod download all go modules
	./scripts/for_all_go_modules.sh -- go mod download all

.PHONY: tcp-metadata-exchange-filter
tcp-metadata-exchange-filter:	## build the tcp-metadata-exchange-filter
	rustup target add wasm32-unknown-unknown
	cargo build --target wasm32-unknown-unknown --release
	cp target/wasm32-unknown-unknown/release/wasm_tcp_metadata.wasm pkg/istio/filters/tcp-metadata-exchange-filter.wasm

.PHONY: tcp-metadata-exchange-filter-shrinked
tcp-metadata-exchange-filter-shrinked: ## build the smaller version of tcp-metadata-exchange-filter 
	rustup target add wasm32-unknown-unknown
	cargo +nightly build -Z build-std=std,panic_abort -Z build-std-features=panic_immediate_abort -j1 --release --target wasm32-unknown-unknown
	cp target/wasm32-unknown-unknown/release/wasm_tcp_metadata.wasm pkg/istio/filters/tcp-metadata-exchange-filter.wasm
