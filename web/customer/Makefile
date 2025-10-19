# Publisher Portal Makefile
.PHONY: build dev test lint clean help

# Configuration
SERVICE_NAME := publisher_portal
BUILD_DIR := dist
PORT := 3001

# Node settings
NPM := npm
NPX := npx

# Default target
.DEFAULT_GOAL := help

# Build targets
build: ## Build publisher portal for production
	@echo "Building $(SERVICE_NAME)..."
	@$(NPM) run build
	@echo "Built $(BUILD_DIR) directory"

dev: ## Start development server
	@echo "Starting $(SERVICE_NAME) development server on port $(PORT)..."
	@$(NPM) run dev

preview: ## Preview production build
	@echo "Previewing $(SERVICE_NAME)..."
	@$(NPM) run preview

# Testing
test: ## Run tests
	@echo "Running tests..."
	@$(NPM) run test

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(NPM) run test:coverage

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	@$(NPM) run lint

type-check: ## Run TypeScript type checking
	@echo "Running type check..."
	@$(NPM) run type-check

# Dependency management
install: ## Install dependencies
	@echo "Installing dependencies..."
	@$(NPM) install

update: ## Update dependencies
	@echo "Updating dependencies..."
	@$(NPM) update
	@$(NPM) audit fix

# Cleaning
clean: ## Clean build artifacts and dependencies
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf node_modules
	@rm -f package-lock.json

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t admux/$(SERVICE_NAME):latest .

# Help
help: ## Show help message
	@echo "$(SERVICE_NAME) Makefile"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
