# ADX Core Library Makefile
.PHONY: build test clean deps

# Configuration
MODULE_NAME := github.com/echoface/admux/shared/adxcore

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Default target
all: deps test

# Build targets
build: ## Build library (for validation)
	@echo "Building $(MODULE_NAME) for validation..."
	$(GOBUILD) ./...

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Testing
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

# Linting
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Formatting
format: ## Format Go code
	@echo "Formatting Go code..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed, using gofmt..."; \
		gofmt -w .; \
	fi

# Cleaning
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCMD) clean
	@rm -f coverage.out coverage.html

# Help
help: ## Show this help message
	@echo "ADX Core Library Makefile"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
