# Go Web Starter Makefile

# Variables
BINARY_NAME=go-web-starter
MAIN_PATH=./cmd/server
BUILD_DIR=./build
DOCKER_IMAGE=go-web-starter:latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: help build clean test coverage run deps docker-build docker-run

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run the application
run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)/main.go

# Download dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build Docker image
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# Development setup
dev-setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	$(GOGET) -u github.com/cosmtrek/air@latest

# Hot reload for development
dev: ## Run with hot reload
	@echo "Starting development server with hot reload..."
	air

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
lint: ## Lint Go code
	@echo "Linting code..."
	golangci-lint run

# Generate mocks
mocks: ## Generate mocks for testing
	@echo "Generating mocks..."
	mockgen -source=internal/repository/interfaces.go -destination=test/mocks/repository_mock.go

# Database migration
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	./scripts/migrate.sh up

migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	./scripts/migrate.sh down

# Docker compose operations
docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-compose-down: ## Stop all services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose -f deployments/docker/docker-compose.yml down

# Security scan
security: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

# All-in-one development setup
setup: deps dev-setup ## Complete development setup
	@echo "Development environment setup complete!"