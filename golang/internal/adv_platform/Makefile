# Media Platform Makefile
.PHONY: build test clean deps run

# Configuration
SERVICE_NAME := ad_platform
BINARY_NAME := ad_platform-server
MAIN_PATH := cmd/main.go
OUTPUT_DIR := bin
OUTPUT_PATH := $(OUTPUT_DIR)/$(BINARY_NAME)

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Default target
all: deps build

# Build targets
build: ## Build Media Platform binary
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) -o $(OUTPUT_PATH) $(MAIN_PATH)
	@echo "Built $(OUTPUT_PATH)"

build-local: build

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Testing
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Running
run: build-local ## Build and run locally
	@echo "Running $(SERVICE_NAME)..."
	./$(OUTPUT_PATH)

run-dev: ## Run in development mode
	@echo "Running $(SERVICE_NAME) in development mode..."
	$(GOCMD) run $(MAIN_PATH)

# Cleaning
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(OUTPUT_DIR)

# Health check
health: ## Check service health
	@echo "Checking service health..."
	@curl -f http://localhost:8082/health || echo "Service is not running or unhealthy"

# Help
help: ## Show this help message
	@echo "$(SERVICE_NAME) Makefile"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
