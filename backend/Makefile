# Backend Makefile for Zplus SaaS Base

.PHONY: help install build run test clean dev gen-gql fmt lint vet check deps-update deps-clean docker-build docker-run

# Variables
APP_NAME=zplus-saas-base
BINARY_NAME=api
MAIN_PATH=./cmd/api
BUILD_DIR=./build
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -path "*/generated/*")

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Installation
install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

deps-clean: ## Clean module cache
	@echo "Cleaning module cache..."
	go clean -modcache

# Code generation
gen-gql: ## Generate GraphQL code with gqlgen
	@echo "Generating GraphQL code..."
	go run github.com/99designs/gqlgen generate

# Build
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

build-local: ## Build for local OS
	@echo "Building $(APP_NAME) for local OS..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run
run: ## Run the application
	@echo "Running $(APP_NAME)..."
	go run $(MAIN_PATH)

dev: ## Run in development mode with hot reload (requires air)
	@echo "Running in development mode..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Testing
test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -race -tags=integration ./...

test-auth: ## Test authentication system
	@echo "Testing authentication system..."
	./scripts/test-login-interfaces.sh

test-login: ## Test login interfaces
	@echo "Testing login interfaces..."
	./scripts/test-login-interfaces.sh

# Code quality
fmt: ## Format Go code
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)

lint: ## Run golangci-lint
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint ## Run all code quality checks

# Database
db-migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	migrate -path database/migrations -database "postgres://postgres:postgres@localhost:5432/zplus_saas?sslmode=disable" up

db-migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	migrate -path database/migrations -database "postgres://postgres:postgres@localhost:5432/zplus_saas?sslmode=disable" down

db-migrate-create: ## Create a new migration file (usage: make db-migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir database/migrations -seq $(NAME)

db-seed: ## Seed the database
	@echo "Seeding database..."
	go run ./scripts/seed/main.go

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-run: ## Run application in Docker
	@echo "Running $(APP_NAME) in Docker..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

# Development environment
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please update .env with your configuration"

dev-db-up: ## Start development databases with Docker Compose
	@echo "Starting development databases..."
	docker-compose -f ../docker-compose.database.yml up -d

dev-db-down: ## Stop development databases
	@echo "Stopping development databases..."
	docker-compose -f ../docker-compose.database.yml down

dev-db-reset: ## Reset development databases
	@echo "Resetting development databases..."
	docker-compose -f ../docker-compose.database.yml down -v
	docker-compose -f ../docker-compose.database.yml up -d
	sleep 5
	$(MAKE) db-migrate-up
	$(MAKE) db-seed

# Utilities
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache
	go clean -testcache

graph-schema: ## Generate GraphQL schema documentation
	@echo "Generating GraphQL schema documentation..."
	@which graphql-markdown > /dev/null || npm install -g graphql-markdown
	graphql-markdown ./graph/schema.graphqls > ./docs/api/graphql-schema.md

generate: gen-gql ## Run all code generation

full-setup: install dev-setup dev-db-up db-migrate-up db-seed ## Full development setup

# Benchmarks
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Security
security-scan: ## Run security scan with gosec
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Profiling
profile-cpu: ## Generate CPU profile
	@echo "Generating CPU profile..."
	go test -cpuprofile=cpu.prof -bench=. ./...

profile-mem: ## Generate memory profile
	@echo "Generating memory profile..."
	go test -memprofile=mem.prof -bench=. ./...

# Release
version: ## Show current version
	@echo "Current version: $(shell git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')"

tag: ## Create a new tag (usage: make tag VERSION=v1.0.0)
	@echo "Creating tag: $(VERSION)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Database inspection
db-status: ## Show database migration status
	@echo "Database migration status:"
	migrate -path database/migrations -database "postgres://postgres:postgres@localhost:5432/zplus_saas?sslmode=disable" version
