NASP_AUTH_TOKEN ?= $(shell kubectl -n external get secret -l nasp.k8s.cisco.com/workloadgroup=test-tcp -o jsonpath='{@.items[0].data.token}' | base64 -d)

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run-server
run-server: ## Run TCP server
	NASP_AUTH_TOKEN=$(NASP_AUTH_TOKEN) go run . -mode=server

.PHONY: run-client
run-client: ## Run TCP client
	NASP_AUTH_TOKEN=$(NASP_AUTH_TOKEN) go run . -mode=client
