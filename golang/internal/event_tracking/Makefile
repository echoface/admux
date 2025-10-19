# Event Tracking System Makefile
.PHONY: build test clean deps run

# Configuration
SERVICE_NAME := event-tracking
BINARY_NAME := tracking-server
MAIN_PATH := cmd/main.go
OUTPUT_DIR := bin
OUTPUT_PATH := $(OUTPUT_DIR)/$(BINARY_NAME)

# Build settings
BUILD_FLAGS := -ldflags="-s -w"
CGO_ENABLED := 0
GOOS := linux
GOARCH := amd64

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Default target
all: deps build

# Build targets
build: ## Build Event Tracking system binary
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(BUILD_FLAGS) -o $(OUTPUT_PATH) $(MAIN_PATH)
	@echo "Built $(OUTPUT_PATH)"

build-local: ## Build for local platform
	@echo "Building $(SERVICE_NAME) for local platform..."
	@mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) -o $(OUTPUT_PATH) $(MAIN_PATH)
	@echo "Built $(OUTPUT_PATH)"

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

# Running
run: build-local ## Build and run locally
	@echo "Running $(SERVICE_NAME)..."
	./$(OUTPUT_PATH)

run-dev: ## Run in development mode
	@echo "Running $(SERVICE_NAME) in development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "air not installed, using go run..."; \
		$(GOCMD) run $(MAIN_PATH); \
	fi

# Cleaning
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(OUTPUT_DIR)
	@rm -f coverage.out coverage.html

# Health check
health: ## Check service health
	@echo "Checking service health..."
	@curl -f http://localhost:8081/health || echo "Service is not running or unhealthy"

# Help
help: ## Show this help message
	@echo "$(SERVICE_NAME) Makefile"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
