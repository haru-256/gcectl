.DEFAULT_GOAL := help
include ../../scripts/common.mk

.PHONY: plan
plan: # plan terraform
	terraform plan

.PHONY: apply
apply: # apply terraform
	terraform apply

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done
