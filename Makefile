PKG            				:= gitlab.panel.loc/cryptobecky/cmc/currency
PKG_LIST       				:= $(shell go list ${PKG}/... | grep -v /vendor/)
CONFIG         				:= $(wildcard local.yml)
NAMESPACE	   				:= "default"

.PHONY: test

.PHONY: all
all: setup test

.PHONY: setup
setup: ## Installing all service dependencies.
	@echo "Setup..."
	vgo mod vendor

.PHONY: test
test: ## Run tests for all packages.
	@echo "Testing..."
	go test -race ${PKG_LIST}

.PHONY: coverage
coverage: ## Calculating code test coverage.
	@echo "Calculating coverage..."
	PKG=$(PKG) ./tools/coverage.sh

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-\:]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ": .*?## "}; {gsub(/[\\]*/,""); printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'