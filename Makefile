LICENSEI_VERSION = 0.5.0
GOLANGCI_VERSION = 1.49.0
HEIMDALL_IMAGE ?= heimdall

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

heimdall-docker: ## Build heimdall docker container
	docker build -t $(HEIMDALL_IMAGE) -f experimental/heimdall/Dockerfile --platform linux/amd64 .

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b ./bin/ v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
# "unused" linter is a memory hog, but running it separately keeps it contained (probably because of caching)
	bin/golangci-lint run --disable=unused -c .golangci.yml --timeout 2m
	bin/golangci-lint run -c .golangci.yml --timeout 2m

.PHONY: lint-fix
lint-fix: bin/golangci-lint ## Run linter
	@bin/golangci-lint run --fix

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-check
license-check: bin/licensei ## Run license check
	bin/licensei check
	bin/licensei header

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	bin/licensei cache

.PHONY: test
test: ## Run tests
	go test ./pkg/... \
    	-coverprofile cover.out \
    	-v \
    	-failfast \
    	-test.v \
    	-test.paniconexit0 \
    	-timeout 1h
