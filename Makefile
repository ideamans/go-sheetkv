.PHONY: all build test test-unit test-integration test-e2e coverage lint fmt clean install help

# Variables
GO=go
GOTEST=$(GO) test
GOCOVER=$(GO) tool cover
GOLINT=golangci-lint
GOFMT=gofmt

# Default target
all: lint test

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v -race ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -race -short ./ ./adapters/...

# Run only root package tests
test-core:
	@echo "Running core package tests..."
	$(GOTEST) -v -race ./

# Run only adapter tests
test-adapters:
	@echo "Running adapter tests..."
	$(GOTEST) -v -race ./adapters/...

# Run Excel adapter tests
test-excel:
	@echo "Running Excel adapter tests..."
	$(GOTEST) -v -race ./adapters/excel/...

# Run Google Sheets adapter tests
test-googlesheets:
	@echo "Running Google Sheets adapter tests..."
	$(GOTEST) -v -race ./adapters/googlesheets/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race -timeout 10m ./tests/integration

# Run API tests
test-api:
	@echo "Running API tests..."
	@echo "Note: If Google Sheets credentials are not configured, tests will automatically use Excel adapter"
	$(GOTEST) -v -race -timeout 10m ./tests/api

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCOVER) -func=coverage.txt
	$(GOCOVER) -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Running go vet instead..."; \
		$(GO) vet ./...; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -l -w .
	$(GO) mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GO) clean
	rm -f coverage.txt coverage.html
	rm -rf dist/

# Run go mod tidy
tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

# Check if code needs formatting
check-fmt:
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code needs formatting. Run 'make fmt'" && exit 1)

# Run static analysis
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run specific test
test-pkg:
	@echo "Running tests for specific package..."
	@echo "Usage: make test-pkg PKG=./"
	$(GOTEST) -v -race $(PKG)

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# CI/CD specific targets
ci-test:
	@echo "Running CI tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

ci-lint:
	@echo "Running CI lint..."
	$(GO) vet ./...
	$(GOFMT) -l .
	@test -z "$$($(GOFMT) -l .)" || (echo "Code needs formatting" && exit 1)

# Help target
help:
	@echo "Available targets:"
	@echo "  make all              - Run lint and test"
	@echo "  make test             - Run all tests"
	@echo "  make test-unit        - Run unit tests only"
	@echo "  make test-core        - Run core package tests"
	@echo "  make test-adapters    - Run adapter tests"
	@echo "  make test-excel       - Run Excel adapter tests"
	@echo "  make test-googlesheets - Run Google Sheets adapter tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-api         - Run API tests (auto-detects adapter)"
	@echo "  make coverage         - Run tests with coverage report"
	@echo "  make lint             - Run linter"
	@echo "  make fmt              - Format code"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make install          - Install dependencies"
	@echo "  make tidy             - Run go mod tidy"
	@echo "  make check-fmt        - Check if code needs formatting"
	@echo "  make vet              - Run go vet"
	@echo "  make bench            - Run benchmarks"
	@echo "  make dev-tools        - Install development tools"
	@echo "  make help             - Show this help message"