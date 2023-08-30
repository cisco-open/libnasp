include ../../../common.mk

.PHONY: escape-analysis
escape-analysis: ## Run escape analysis
	@go build -gcflags="-m -m -l" ./... 2>escape.out

.PHONY: benchmark
benchmark: ## Run benchmarks
	@find . -iname "go.mod" | xargs -L1 sh -c 'cd $$(dirname $$0); go test ./... -bench=. -count=6 -benchmem  -run=^$ -failfast -test.v -test.paniconexit0 -timeout 4h | tee benchmark_stats.out'

${REPO_ROOT}/bin/benchstat:
	@mkdir -p ${REPO_ROOT}/bin
	@mkdir -p bin
	@GOBIN=${REPO_ROOT}/bin/ go install golang.org/x/perf/cmd/benchstat@latest

.PHONY:
show-benchmark-stat: ${REPO_ROOT}/bin/benchstat
	${REPO_ROOT}/bin/benchstat benchmark_stats.out
