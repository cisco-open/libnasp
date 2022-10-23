.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run-server
run-server: ## Run HTTP server
	go run . -mode=server

.PHONY: run-client
run-client: ## Run HTTP client
	 go run . -mode=client -dump-client-response=1 -request-url http://localhost:8080
