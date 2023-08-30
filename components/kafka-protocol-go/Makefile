REPO_ROOT=$(shell git rev-parse --show-toplevel)

LICENSEI_VERSION = 0.7.0
GOLANGCI_VERSION = 1.51.2


.PHONY: test
test: ## Run tests
	@find . -iname "go.mod" | xargs -L1 sh -c 'cd $$(dirname $$0); go test ./... -coverpkg=./... -coverprofile cover.out -v -failfast -test.v -test.paniconexit0 -timeout 1h'

.PHONY: escape-analysis
escape-analysis: ## Run escape analysis
	@cd pkg && go build -gcflags="-m -m -l" ./... 2>escape.out

.PHONY: benchmark
benchmark: ## Run benchmarks
	@find . -iname "go.mod" | xargs -L1 sh -c 'cd $$(dirname $$0); go test ./... -bench=. -count=6 -benchmem  -run=^$ -failfast -test.v -test.paniconexit0 -timeout 4h | tee benchmark_stats.out'

.PHONY: fmt
fmt: ## Run go fmt against code
	@go fmt ./...
	@cd pkg && go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	@find . -iname "go.mod" | xargs -L1 sh -c 'cd $$(dirname $$0); go vet ./...'

.PHONY: tidy
tidy: ## Run `go mod tidy` against all modules
	find . -iname "go.mod" | xargs -L1 sh -c 'cd $$(dirname $$0); go mod tidy && go mod download all'

generate:   ## Generate code
	@rm -rf ${REPO_ROOT}/components/kafka-protocol-go/pkg/protocol/messages
	@cd cmd/gen && go run . ${REPO_ROOT}/components/kafka-protocol-go

${REPO_ROOT}/bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p ${REPO_ROOT}/bin
	@mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

${REPO_ROOT}/bin/golangci-lint: ${REPO_ROOT}/bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} ${REPO_ROOT}/bin/golangci-lint

.PHONY: lint
lint: ${REPO_ROOT}/bin/golangci-lint ## Run linter
# "unused" linter is a memory hog, but running it separately keeps it contained (probably because of caching)
	@${REPO_ROOT}/bin/golangci-lint run --disable=unused -c ${REPO_ROOT}/.golangci.yml --timeout 2m
	@${REPO_ROOT}/bin/golangci-lint run -c ${REPO_ROOT}/.golangci.yml --timeout 2m

	@cd cmd && ${REPO_ROOT}/bin/golangci-lint run -c ${REPO_ROOT}/.golangci.yml --timeout 2m
	@cd pkg && ${REPO_ROOT}/bin/golangci-lint run -c ${REPO_ROOT}/.golangci.yml --timeout 2m

.PHONY: lint-fix
lint-fix: ${REPO_ROOT}/bin/golangci-lint ## Run linter
	@${REPO_ROOT}/bin/golangci-lint run -c ${REPO_ROOT}/.golangci.yml --fix --timeout 2m

	@cd cmd && ${REPO_ROOT}/bin/golangci-lint run -c ${REPO_ROOT}/.golangci.yml --fix --timeout 2m
	@cd pkg && ${REPO_ROOT}/bin/golangci-lint run -c ${REPO_ROOT}/.golangci.yml --fix --timeout 2m

${REPO_ROOT}/bin/licensei: ${REPO_ROOT}/bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} ${REPO_ROOT}/bin/licensei

${REPO_ROOT}/bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p ${REPO_ROOT}/bin
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	mv bin/licensei $@

.PHONY: license-check
license-check: ${REPO_ROOT}/bin/licensei ## Run license check
	@${REPO_ROOT}/bin/licensei --config ${REPO_ROOT}/.licensei.toml check
	@${REPO_ROOT}/bin/licensei --config ${REPO_ROOT}/.licensei.toml header

.PHONY: license-cache
license-cache: ${REPO_ROOT}/bin/licensei ## Generate license cache
	${REPO_ROOT}/bin/licensei --config ${REPO_ROOT}/.licensei.toml cache

${REPO_ROOT}/bin/benchstat:
	@mkdir -p ${REPO_ROOT}/bin
	@mkdir -p bin
	@GOBIN=${REPO_ROOT}/bin/ go install golang.org/x/perf/cmd/benchstat@latest

.PHONY:
show-benchmark-stat: ${REPO_ROOT}/bin/benchstat
	${REPO_ROOT}/bin/benchstat pkg/benchmark_stats.out

${REPO_ROOT}/bin/fieldalignment:
	@mkdir -p ${REPO_ROOT}/bin
	@mkdir -p bin
	@GOBIN=${REPO_ROOT}/bin/ go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

.PHONY:
fieldalignment: ${REPO_ROOT}/bin/fieldalignment
	cd pkg && ${REPO_ROOT}/bin/fieldalignment -fix ./...