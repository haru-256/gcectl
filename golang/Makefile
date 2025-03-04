ifndef GOARCH
	GOARCH=$(shell go env GOARCH)
endif

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif

.DEFAULT_GOAL := help

.PHONY: test
test: ## Execute tests
	go test -race -shuffle=on ./...

.PHONY: lint
lint: ## Run Linter
	golangci-lint run --config=./.golangci.yml ./...

.PHONY: fmt
fmt: ## Run formatter
	go fmt ./...

.PHONY: init
init: ## Initialize project
	cobra-cli init

.PHONY: build
build: ## Build project
	CGO_ENABLED=0 GO111MODULE=on GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/main

.PHONY: install
install: ## Install project
	$(MAKE) build
	ln -sf $(PWD)/bin/main $${HOME}/.local/bin/gcectl
	$(MAKE) completion

.PHONY: completion
completion: ./bin/main ## Generate completion
	./bin/main completion fish > $${HOME}/.config/fish/completions/gcectl.fish

.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
