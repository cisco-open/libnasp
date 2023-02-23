NASP_AUTH_TOKEN ?= $(shell kubectl -n external get secret -l nasp.k8s.cisco.com/workloadgroup=test-grpc -o jsonpath='{@.items[0].data.token}' | base64 -D)

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: proto
proto: ## Generate hello protobuf
	protoc --go_out=./pb --go_opt=paths=source_relative \
	--go-grpc_out=./pb --go-grpc_opt=paths=source_relative \
	hello.proto

.PHONY: run-server
run-server: ## Run gRPC server
	NASP_AUTH_TOKEN=$(NASP_AUTH_TOKEN) go run .

.PHONY: run-client
run-client: ## Run gRPC client
	NASP_AUTH_TOKEN=$(NASP_AUTH_TOKEN) go run ./client
