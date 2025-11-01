.DEFAULT_GOAL := help

.PHONY: install
install: ## Install project
	curl -sSL "https://raw.githubusercontent.com/haru-256/gcectl/main/scripts/install.sh" | sh
	gcectl completion fish > $${HOME}/.config/fish/completions/gcectl.fish

.PHONY: help
help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
