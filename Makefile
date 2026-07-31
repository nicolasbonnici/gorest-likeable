.PHONY: help test lint lint-fix quality-test build clean install coverage

# Add Go bin to PATH for all targets
GOPATH ?= $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

# Default target
GOLANGCI_LINT_VERSION := v2.12.2

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies, dev tools, and git hooks
	@echo "[INFO] Installing development environment..."
	@echo ""
	@echo "[1/3] Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"
	@echo ""
	@echo "[2/3] Installing development tools..."
	@if ! golangci-lint --version 2>/dev/null | grep -qE 'version v?2\.'; then \
		echo "  Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		GOWORK=off go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@echo "✓ Development tools installed"
	@echo ""
	@echo "[3/3] Installing git hooks..."
	@bash .githooks/install.sh
	@echo ""
	@echo "✅ Installation complete! Ready to develop."
	@echo ""
	@echo "Next steps:"
	@echo "  • Run 'make test' to verify your setup"
	@echo "  • Run 'make lint' to check code quality"
	@echo "  • See 'make help' for all available commands"
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race ./...

lint: ## Run golangci-lint (bundles staticcheck, errcheck, govet, gocyclo, misspell)
	@echo "Running golangci-lint..."
	@$$(go env GOPATH)/bin/golangci-lint run ./...

lint-fix: ## Run linter with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@$$(go env GOPATH)/bin/golangci-lint run --fix ./...

quality-test: ## Run all Go Report Card quality checks locally
	@echo "======================================================="
	@echo "Running Go Report Card Quality Checks..."
	@echo "======================================================="
	@echo ""
	@echo "[1/3] Checking gofmt formatting..."
	@UNFORMATTED=$$(gofmt -s -l . 2>&1); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "❌ The following files are not properly formatted:"; \
		echo "$$UNFORMATTED"; \
		echo ""; \
		echo "Run 'gofmt -s -w .' to fix formatting issues"; \
		exit 1; \
	fi
	@echo "✓ gofmt passed"
	@echo ""
	@echo "[2/3] Running go vet..."
	@go vet ./...
	@echo "✓ go vet passed"
	@echo ""
	@echo "[3/3] Running golangci-lint..."
	@$$(go env GOPATH)/bin/golangci-lint run ./...
	@echo "✓ golangci-lint passed"
	@echo ""
	@echo "======================================================="
	@echo "✓ All quality checks passed!"
	@echo "======================================================="
	@echo "golangci-lint covers staticcheck, ineffassign, misspell,"
	@echo "errcheck and gocyclo; running them separately duplicated it."

build: ## Build verification
	@echo "Building plugin..."
	@go build -v ./...
	@echo "✓ Build successful"

clean: ## Clean build artifacts and caches
	@echo "Cleaning..."
	@go clean -cache -testcache -modcache
	@rm -f coverage.out coverage.html
	@echo "✓ Cleaned"

coverage: ## Generate and display coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out
	@echo ""
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report saved to coverage.html"
