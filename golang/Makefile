.DEFAULT_GOAL := help

.PHONY: help
test: ## Execute tests
	go test -race -shuffle=on ./...

.PHONY: lint
lint: ## Run Linter
	golangci-lint run --config=./.golangci.yaml ./...

.PHONY: fmt
fmt: ## Run formatter
	go fmt ./...

.PHONY: init
init: ## Initialize project
	cobra-cli init

.PHONY: test
test: ## Run tests
	go test -v ./...


.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
