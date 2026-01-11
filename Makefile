# Connect 4 Multiplayer Makefile

# Variables
BINARY_NAME=connect4-server
ANALYTICS_BINARY=connect4-analytics
MIGRATE_BINARY=connect4-migrate
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
.PHONY: all
all: build

# Build targets
.PHONY: build
build: build-server build-analytics build-migrate

.PHONY: build-server
build-server:
	@echo "Building server..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go

.PHONY: build-analytics
build-analytics:
	@echo "Building analytics service..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(ANALYTICS_BINARY) cmd/analytics/main.go

.PHONY: build-migrate
build-migrate:
	@echo "Building migration tool..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(MIGRATE_BINARY) cmd/migrate/main.go

# Development targets
.PHONY: dev
dev:
	@echo "Starting development server with hot reload..."
	@air

.PHONY: run-server
run-server:
	@echo "Running server..."
	@go run cmd/server/main.go

.PHONY: run-analytics
run-analytics:
	@echo "Running analytics service..."
	@go run cmd/analytics/main.go

# Credential setup targets
.PHONY: setup-credentials
setup-credentials:
	@echo "Setting up cloud service credentials..."
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/setup-credentials.ps1
else
	@bash scripts/setup-credentials.sh
endif

.PHONY: create-env
create-env:
	@echo "Creating .env file from template..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "Created .env file. Please edit it with your credentials."; else echo ".env file already exists."; fi

.PHONY: validate-env
validate-env:
	@echo "Validating environment configuration..."
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/setup-credentials.ps1 -Action validate
else
	@bash scripts/setup-credentials.sh validate
endif

.PHONY: test-db
test-db:
	@echo "Testing database connection..."
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/setup-credentials.ps1 -Action test-db
else
	@bash scripts/setup-credentials.sh test-db
endif

.PHONY: test-kafka
test-kafka:
	@echo "Testing Kafka connection..."
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/setup-credentials.ps1 -Action test-kafka
else
	@bash scripts/setup-credentials.sh test-kafka
endif

# Database targets
.PHONY: migrate-up
migrate-up: build-migrate
	@echo "Running database migrations..."
	@./$(BUILD_DIR)/$(MIGRATE_BINARY) -direction=up

.PHONY: migrate-down
migrate-down: build-migrate
	@echo "Rolling back database migrations..."
	@./$(BUILD_DIR)/$(MIGRATE_BINARY) -direction=down

.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go

# Testing targets
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: test-property
test-property:
	@echo "Running property-based tests..."
	@go test -tags=property -v ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -v ./...

# Code quality targets
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Documentation targets
.PHONY: docs
docs:
	@echo "Generating API documentation..."
	@swag init -g cmd/server/main.go -o docs/swagger

.PHONY: docs-serve
docs-serve:
	@echo "Serving documentation at http://localhost:8080/swagger/index.html"
	@make run-server

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

.PHONY: docker-up
docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

.PHONY: docker-down
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

.PHONY: docker-logs
docker-logs:
	@docker-compose logs -f

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean
	@echo "Cleaning all generated files..."
	@go clean -cache
	@go clean -modcache

# Development setup
.PHONY: setup
setup:
	@echo "Setting up development environment..."
	@go mod download
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest

# Production deployment
.PHONY: build-prod
build-prod:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(ANALYTICS_BINARY) cmd/analytics/main.go

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build all binaries"
	@echo "  build-server   - Build server binary"
	@echo "  build-analytics - Build analytics service binary"
	@echo "  build-migrate  - Build migration tool binary"
	@echo "  build-prod     - Build for production"
	@echo ""
	@echo "Development targets:"
	@echo "  dev            - Start development server with hot reload"
	@echo "  run-server     - Run server directly"
	@echo "  run-analytics  - Run analytics service directly"
	@echo ""
	@echo "Credential setup targets:"
	@echo "  setup-credentials - Interactive credential setup"
	@echo "  create-env     - Create .env file from template"
	@echo "  validate-env   - Validate environment configuration"
	@echo "  test-db        - Test database connection"
	@echo "  test-kafka     - Test Kafka connection"
	@echo ""
	@echo "Database targets:"
	@echo "  migrate        - Run database migrations"
	@echo "  migrate-up     - Run database migrations (with binary)"
	@echo "  migrate-down   - Rollback database migrations"
	@echo ""
	@echo "Testing targets:"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-property  - Run property-based tests"
	@echo "  test-integration - Run integration tests"
	@echo ""
	@echo "Code quality targets:"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo ""
	@echo "Documentation targets:"
	@echo "  docs           - Generate API documentation"
	@echo "  docs-serve     - Serve documentation"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start services with Docker"
	@echo "  docker-down    - Stop Docker services"
	@echo "  docker-logs    - Show Docker logs"
	@echo ""
	@echo "Utility targets:"
	@echo "  deps           - Download dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  setup          - Setup development environment"
	@echo "  help           - Show this help message"