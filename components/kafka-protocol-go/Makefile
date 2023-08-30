include ../../common.mk

.PHONY: test
test: ## Run tests
	@find . -iname "go.mod" | xargs -L1 sh -c 'cd $$(dirname $$0); go test ./... -coverpkg=./... -coverprofile cover.out -v -failfast -test.v -test.paniconexit0 -timeout 1h'

generate:   ## Generate code
	@rm -rf ${REPO_ROOT}/components/kafka-protocol-go/pkg/protocol/messages
	@cd cmd/gen && go run . ${REPO_ROOT}/components/kafka-protocol-go

${REPO_ROOT}/bin/fieldalignment:
	@mkdir -p ${REPO_ROOT}/bin
	@mkdir -p bin
	@GOBIN=${REPO_ROOT}/bin/ go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

.PHONY:
fieldalignment: ${REPO_ROOT}/bin/fieldalignment
	@cd pkg && ${REPO_ROOT}/bin/fieldalignment -fix ./...

