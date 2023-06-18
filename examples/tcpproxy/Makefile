NASP_AUTH_TOKEN ?= $(shell kubectl -n external get secret -l nasp.k8s.cisco.com/workloadgroup=test-tcp -o jsonpath='{@.items[0].data.token}' | base64 -d)
HEIMDALL_URL ?= https://localhost:16443/config
BUFFER_SIZE ?= 65536
CLIENT_ADDRESS ?= :5001
SERVER_ADDRESS ?= :5002

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run-proxy
run-proxy: ## Run TCP proxy
	HEIMDALL_AUTH_TOKEN=$(NASP_AUTH_TOKEN) HEIMDALL_URL=$(HEIMDALL_URL) go run . -buffer-size $(BUFFER_SIZE) -client-address $(CLIENT_ADDRESS) -server-address $(SERVER_ADDRESS)
