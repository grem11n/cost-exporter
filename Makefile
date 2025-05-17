PHONY: help

GOLANGCI_LINT_VERSION := v2.1.6

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

check_deps: ## Check if dependencies installed
	@for dep in "richgo" "docker" ; do \
		which $$dep &>/dev/null && echo "$$dep is installed"  || echo "$$dep is not installed"; \
	done

install_deps: ## Install test dependencies
	@go get -u github.com/kyoh86/richgo
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

lint: ## Run golangci-lint
	@golangci-lint run -E asciicheck -E bidichk -E bodyclose -E canonicalheader -E copyloopvar -E cyclop -E err113 -E errname -E funcorder -E funlen -E forcetypeassert -E gocognit -E goconst -E gocritic -E gosec -E iface -E lll -E loggercheck -E misspell -E tagalign -E whitespace #-E revive

test: ## Run tests with Richgo
	@richgo test -v ./...

run: ## Run the app locally
	@go run . -c config.example.yaml

run-debug: ## Run the app locally with DEBUG logs
	@LOG_LEVEL=DEBUG go run . -c config.example.yaml
