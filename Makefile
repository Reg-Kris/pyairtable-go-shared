# Variables
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=pyairtable-go-shared
BINARY_UNIX=$(BINARY_NAME)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run tests
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage
	$(GOCMD) tool cover -html=coverage.out

.PHONY: test-short
test-short: ## Run only short tests
	$(GOTEST) -short -v ./...

.PHONY: bench
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: vet
vet: ## Run go vet
	$(GOVET) ./...

.PHONY: fmt
fmt: ## Run go fmt
	$(GOCMD) fmt ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	staticcheck ./...

.PHONY: security
security: ## Run gosec security scanner
	gosec ./...

.PHONY: vuln-check
vuln-check: ## Check for vulnerabilities
	govulncheck ./...

.PHONY: deps
deps: ## Download dependencies
	$(GOCMD) mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	$(GOCMD) get -u ./...
	$(GOCMD) mod tidy

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	$(GOCMD) mod verify

.PHONY: clean-testcache
clean-testcache: ## Clean test cache
	$(GOCMD) clean -testcache

.PHONY: clean
clean: clean-testcache ## Clean build cache
	$(GOCMD) clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out

.PHONY: install-tools
install-tools: ## Install development tools
	$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2

.PHONY: check
check: fmt vet lint staticcheck security ## Run all checks

.PHONY: ci
ci: deps-verify check test ## Run CI pipeline

.PHONY: pre-commit
pre-commit: fmt vet lint test-short ## Run pre-commit checks

.PHONY: docker-test
docker-test: ## Run tests in Docker
	docker run --rm -v "$$(pwd)":/usr/src/app -w /usr/src/app golang:1.21 make ci

.PHONY: generate-mocks
generate-mocks: ## Generate mocks
	$(GOCMD) generate ./...

.PHONY: docs
docs: ## Generate documentation
	$(GOCMD) doc -all ./...

.PHONY: serve-docs
serve-docs: ## Serve documentation locally
	godoc -http=:6060

.PHONY: release-patch
release-patch: ## Create a patch release
	@echo "Creating patch release..."
	git tag -a "v$$(git describe --tags --abbrev=0 | awk -F. '{print $$1"."$$2"."$$3+1}')" -m "Patch release"

.PHONY: release-minor
release-minor: ## Create a minor release
	@echo "Creating minor release..."
	git tag -a "v$$(git describe --tags --abbrev=0 | awk -F. '{print $$1"."$$2+1".0"}')" -m "Minor release"

.PHONY: release-major
release-major: ## Create a major release
	@echo "Creating major release..."
	git tag -a "v$$(git describe --tags --abbrev=0 | awk -F. '{print $$1+1".0.0"}')" -m "Major release"

.PHONY: mod-graph
mod-graph: ## Show module dependency graph
	$(GOCMD) mod graph

.PHONY: mod-why
mod-why: ## Show why modules are needed
	$(GOCMD) mod why -m all

.PHONY: licenses
licenses: ## Show licenses of dependencies
	$(GOCMD) list -m -json all | jq -r '.Path + " " + (.Version // "N/A")'

# Default target
.DEFAULT_GOAL := help