.DEFAULT_GOAL := help

.PHONY: test
test: ## Execute tests
	cargo test

.PHONY: lint
lint: ## Run Linter
	cargo check
	cargo clippy

.PHONY: fmt
fmt: ## Run formatter
	cargo fmt

.PHONY: build
build: ## Build project
	cargo build --release

.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
