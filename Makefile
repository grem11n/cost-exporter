.PHONY: help

APP = cost-explorer
DOCKERFILE_CI = Dockerfile.ci

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

check_deps: ## Check if dependencies installed
	@for dep in "richgo" "docker" ; do \
		which $$dep &>/dev/null  || echo "$$dep is not installed"; \
	done

install_deps: ## Install test dependencies
	@go get -u github.com/kyoh86/richgo
	@go install github.com/goreleaser/goreleaser/v2@latest

docker-build-test-image: ## Build an image to run tests
	@docker build -f $(DOCKERFILE_CI) -t $(APP)-test:latest .

docker-run-tests: ## Run tests in a Docker container
	@docker run --rm \
		-v $(PWD):/src:z \
		--workdir=/src \
		$(APP)-test:latest \
			sh -c \
			"go mod tidy && richgo test -v ./..."
