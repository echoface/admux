# Admin Dashboard Makefile
.PHONY: build dev test lint clean help

# Configuration
SERVICE_NAME := admin_dashboard
BUILD_DIR := dist
PORT := 3000

# Node settings
NPM := npm
NPX := npx

# Default target
.DEFAULT_GOAL := help

# Build targets
build: ## Build admin dashboard for production
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

lint:fix: ## Fix linting issues
	@echo "Fixing linting issues..."
	@$(NPM) run lint:fix

type-check: ## Run TypeScript type checking
	@echo "Running type check..."
	@$(NPM) run type-check

# Dependency management
install: ## Install dependencies
	@echo "Installing dependencies..."
	@$(NPM) install

install-dev: ## Install development dependencies
	@echo "Installing development dependencies..."
	@$(NPM) install --include=dev

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

clean-build: ## Clean build artifacts only
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t admux/$(SERVICE_NAME):latest .

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p $(PORT):80 --name $(SERVICE_NAME) admux/$(SERVICE_NAME):latest

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	@docker stop $(SERVICE_NAME) || true
	@docker rm $(SERVICE_NAME) || true

# Storybook
storybook: ## Start Storybook development server
	@echo "Starting Storybook..."
	@$(NPM) run storybook

build-storybook: ## Build Storybook
	@echo "Building Storybook..."
	@$(NPM) run build-storybook

# Help
help: ## Show help message
	@echo "$(SERVICE_NAME) Makefile"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
