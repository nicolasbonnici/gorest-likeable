.PHONY: help test lint lint-fix quality-test build clean install coverage

# Default target
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
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "  Installing golangci-lint..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "  Installing quality check tools..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/gordonklaus/ineffassign@latest
	@go install github.com/client9/misspell/cmd/misspell@latest
	@go install github.com/kisielk/errcheck@latest
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "✓ Development tools installed"
	@echo ""
	@echo "[3/3] Installing git hooks..."
	@bash .githooks/install.sh
	@echo ""
	@echo "✅ Installation complete! Ready to develop."
	@echo ""
	@echo "Next steps:"
	@echo "  • Run 'make test' to verify your setup"
	@echo "  • Run 'make quality-test' to run all quality checks"
	@echo "  • Run 'make lint' to check code quality"
	@echo "  • See 'make help' for all available commands"
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race ./...

lint: ## Run linter
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
	@echo "[1/7] Checking gofmt formatting..."
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
	@echo "[2/7] Running go vet..."
	@go vet ./...
	@echo "✓ go vet passed"
	@echo ""
	@echo "[3/7] Running staticcheck..."
	@$$(go env GOPATH)/bin/staticcheck ./...
	@echo "✓ staticcheck passed"
	@echo ""
	@echo "[4/7] Checking for ineffectual assignments..."
	@$$(go env GOPATH)/bin/ineffassign ./...
	@echo "✓ ineffassign passed"
	@echo ""
	@echo "[5/7] Checking for misspellings..."
	@$$(go env GOPATH)/bin/misspell -error .
	@echo "✓ misspell passed"
	@echo ""
	@echo "[6/7] Checking for unchecked errors..."
	@$$(go env GOPATH)/bin/errcheck ./...
	@echo "✓ errcheck passed"
	@echo ""
	@echo "[7/7] Checking cyclomatic complexity (threshold: 15)..."
	@COMPLEX=$$($$( go env GOPATH))/bin/gocyclo -over 15 . 2>&1); \
	if [ -n "$$COMPLEX" ]; then \
		echo "❌ The following functions have cyclomatic complexity > 15:"; \
		echo "$$COMPLEX"; \
		echo ""; \
		echo "Consider refactoring these functions to reduce complexity"; \
		exit 1; \
	fi
	@echo "✓ gocyclo passed"
	@echo ""
	@echo "Running golangci-lint..."
	@$$(go env GOPATH)/bin/golangci-lint run ./...
	@echo "✓ golangci-lint passed"
	@echo ""
	@echo "======================================================="
	@echo "✓ All quality checks passed!"
	@echo "======================================================="
	@echo "Go Report Card Checks:"
	@echo "  ✓ gofmt -s (formatting)"
	@echo "  ✓ go vet (correctness)"
	@echo "  ✓ staticcheck (static analysis)"
	@echo "  ✓ ineffassign (ineffectual assignments)"
	@echo "  ✓ misspell (spelling)"
	@echo "  ✓ errcheck (error handling)"
	@echo "  ✓ gocyclo (complexity ≤ 15)"
	@echo ""
	@echo "Additional Checks:"
	@echo "  ✓ golangci-lint (comprehensive linting)"
	@echo "======================================================="

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
