HEIMDALL_IMAGE ?= heimdall

include common.mk

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

heimdall-docker: ## Build heimdall docker container
	docker build -t $(HEIMDALL_IMAGE) -f experimental/heimdall/Dockerfile --platform linux/amd64 .

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
license-cache-all: bin/licensei
	./scripts/for_all_go_modules.sh --with-file Makefile -- make license-cache

.PHONY: license-check-all
license-check-all: bin/licensei
	./scripts/for_all_go_modules.sh --with-file Makefile -- make license-check

.PHONY: fmt-all
fmt-all:	## go fmt all go modules
	./scripts/for_all_go_modules.sh --with-file Makefile -- make fmt

.PHONY: vet-all
vet-all:	## go vet all go modules
	./scripts/for_all_go_modules.sh --with-file Makefile -- make vet

.PHONY: lint-all
lint-all: bin/golangci-lint ## lint the whole repo
	./scripts/for_all_go_modules.sh --parallel 1 -- make lint

.PHONY: lint-fix-all
lint-fix-all: bin/golangci-lint ## lint --fix the whole repo
	./scripts/for_all_go_modules.sh --parallel 1 -- make lint-fix

.PHONY: mod-download-all
mod-download-all:	## go mod download all go modules
	./scripts/for_all_go_modules.sh -- go mod download all
