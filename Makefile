# Zplus SaaS Base - Makefile
# Author: Zplus Team
# Description: Development and deployment automation

.PHONY: help dev-setup dev-up dev-down dev-logs test build deploy clean db-setup db-start db-stop db-status db-logs db-backup db-reset db-tools db-monitor

# Default target
.DEFAULT_GOAL := help

# Variables
PROJECT_NAME := zplus-saas-base
DOCKER_COMPOSE_DEV := docker-compose.dev.yml
DOCKER_COMPOSE_DB := docker-compose.database.yml
DOCKER_COMPOSE_PROD := docker-compose.prod.yml
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DOCS_DIR := docs
DB_SCRIPT := scripts/db-manager.sh

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

## Help
help: ## Show this help message
	@echo "$(BLUE)Zplus SaaS Base - Development Commands$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(BLUE)<target>$(RESET)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## Development Setup
dev-setup: ## Initial development environment setup
	@echo "$(GREEN)Setting up development environment...$(RESET)"
	@if [ ! -f .env ]; then cp .env.example .env; echo "$(YELLOW)Created .env file$(RESET)"; fi
	@if [ ! -f $(BACKEND_DIR)/.env ]; then cp $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env; echo "$(YELLOW)Created backend .env file$(RESET)"; fi
	@if [ ! -f $(FRONTEND_DIR)/.env.local ]; then cp $(FRONTEND_DIR)/.env.example $(FRONTEND_DIR)/.env.local; echo "$(YELLOW)Created frontend .env.local file$(RESET)"; fi
	@echo "$(GREEN)Development environment setup complete!$(RESET)"
	@echo "$(YELLOW)Please review and update the .env files before running 'make dev-up'$(RESET)"
	@echo ""
	@echo "$(BLUE)Environment Files Created:$(RESET)"
	@echo "  - .env (main configuration)"
	@echo "  - backend/.env (backend-specific)"
	@echo "  - frontend/.env.local (frontend-specific)"

env-setup-staging: ## Setup staging environment configuration
	@echo "$(GREEN)Setting up staging environment...$(RESET)"
	@if [ ! -f .env ]; then cp .env.staging .env; echo "$(YELLOW)Created .env from staging template$(RESET)"; fi
	@echo "$(RED)IMPORTANT: Update the secrets and database URLs in .env$(RESET)"
	@echo "$(YELLOW)Generate secrets using: make generate-secrets$(RESET)"

env-setup-production: ## Setup production environment configuration
	@echo "$(GREEN)Setting up production environment...$(RESET)"
	@if [ ! -f .env ]; then cp .env.production .env; echo "$(YELLOW)Created .env from production template$(RESET)"; fi
	@echo "$(RED)CRITICAL: Update ALL secrets and production URLs in .env$(RESET)"
	@echo "$(YELLOW)Generate secrets using: make generate-secrets$(RESET)"

generate-secrets: ## Generate secure secrets for production
	@echo "$(GREEN)Generating secure secrets...$(RESET)"
	@echo ""
	@echo "$(BLUE)Copy these secrets to your .env file:$(RESET)"
	@echo "$(YELLOW)JWT_SECRET=$(RESET)$(shell openssl rand -base64 32)"
	@echo "$(YELLOW)NEXTAUTH_SECRET=$(RESET)$(shell openssl rand -base64 32)"
	@echo "$(YELLOW)ENCRYPTION_KEY=$(RESET)$(shell openssl rand -hex 16)"
	@echo "$(YELLOW)SESSION_SECRET=$(RESET)$(shell openssl rand -base64 32)"
	@echo ""
	@echo "$(RED)⚠️  Store these secrets securely and never commit them to git!$(RESET)"

dev-up: ## Start development environment
	@echo "$(GREEN)Starting development environment...$(RESET)"
	@docker-compose -f $(DOCKER_COMPOSE_DEV) up -d
	@echo "$(GREEN)Development environment started!$(RESET)"
	@echo "$(BLUE)Services:$(RESET)"
	@echo "  - System Admin: http://admin.localhost"
	@echo "  - Tenant Example: http://tenant1.localhost"
	@echo "  - Tenant Admin: http://tenant1.localhost/admin"
	@echo "  - API Gateway: http://localhost:8080"
	@echo "  - GraphQL Playground: http://localhost:8080/graphql"
	@echo "  - Keycloak: http://localhost:8081"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - MongoDB: localhost:27017"
	@echo "  - Redis: localhost:6379"
	@echo ""
	@echo "$(YELLOW)Note: Add these entries to /etc/hosts for local development:$(RESET)"
	@echo "127.0.0.1 admin.localhost"
	@echo "127.0.0.1 tenant1.localhost"

dev-down: ## Stop development environment
	@echo "$(YELLOW)Stopping development environment...$(RESET)"
	@docker-compose -f $(DOCKER_COMPOSE_DEV) down
	@echo "$(GREEN)Development environment stopped!$(RESET)"

dev-logs: ## Show development environment logs
	@docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f

dev-restart: ## Restart development environment
	@make dev-down
	@make dev-up

## Database Management
migrate-up: ## Run database migrations
	@echo "$(GREEN)Running database migrations...$(RESET)"
	@cd $(BACKEND_DIR) && go run cmd/migrate/main.go up
	@echo "$(GREEN)Migrations completed!$(RESET)"

## Database Management
db-setup: ## Setup databases with migrations and seeders
	@echo "$(GREEN)Setting up databases...$(RESET)"
	@$(DB_SCRIPT) setup
	@echo "$(GREEN)Database setup completed!$(RESET)"

db-start: ## Start database services
	@echo "$(GREEN)Starting database services...$(RESET)"
	@$(DB_SCRIPT) start

db-stop: ## Stop database services
	@echo "$(YELLOW)Stopping database services...$(RESET)"
	@$(DB_SCRIPT) stop

db-restart: ## Restart database services
	@echo "$(YELLOW)Restarting database services...$(RESET)"
	@$(DB_SCRIPT) restart

db-status: ## Show database services status
	@$(DB_SCRIPT) status

db-logs: ## Show database logs (usage: make db-logs service=postgres)
	@$(DB_SCRIPT) logs $(service)

db-backup: ## Create database backup
	@echo "$(GREEN)Creating database backup...$(RESET)"
	@$(DB_SCRIPT) backup

db-reset: ## Reset databases (WARNING: Deletes all data)
	@echo "$(RED)WARNING: This will delete all database data!$(RESET)"
	@$(DB_SCRIPT) reset

db-tools: ## Show database management tools URLs
	@$(DB_SCRIPT) tools

db-monitor: ## Real-time database performance monitoring
	@$(DB_SCRIPT) monitor

db-migrate: ## Run database migrations
	@echo "$(GREEN)Running database migrations...$(RESET)"
	@cd $(BACKEND_DIR)/database && ./db-migrate.sh migrate

db-seed: ## Run database seeders
	@echo "$(GREEN)Running database seeders...$(RESET)"
	@cd $(BACKEND_DIR)/database && ./db-migrate.sh seed

## Database Quick Commands
db-dev: db-setup ## Quick database setup for development
	@echo "$(BLUE)Database development environment ready!$(RESET)"
	@echo "$(GREEN)Management Tools:$(RESET)"
	@echo "  - pgAdmin (PostgreSQL): http://localhost:8080"
	@echo "  - Mongo Express: http://localhost:8081"
	@echo "  - Redis Commander: http://localhost:8082"
	@echo "  - Adminer: http://localhost:8083"

migrate-rollback: ## Rollback last migration
	@echo "$(YELLOW)Rolling back database migrations...$(RESET)"
	@cd $(BACKEND_DIR) && go run cmd/migrate/main.go down 1
	@echo "$(GREEN)Rollback completed!$(RESET)"

migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then echo "$(RED)Error: Please provide migration name. Usage: make migrate-create name=migration_name$(RESET)"; exit 1; fi
	@echo "$(GREEN)Creating migration: $(name)$(RESET)"
	@cd $(BACKEND_DIR) && go run cmd/migrate/main.go create $(name)
	@echo "$(GREEN)Migration created!$(RESET)"

seed-data: ## Seed database with sample data
	@echo "$(GREEN)Seeding database with sample data...$(RESET)"
	@cd $(BACKEND_DIR) && go run cmd/seed/main.go
	@echo "$(GREEN)Database seeded!$(RESET)"

## Keycloak & Authentication
keycloak-setup: ## Setup Keycloak realm and configuration
	@echo "$(GREEN)Setting up Keycloak...$(RESET)"
	@chmod +x keycloak/setup-keycloak.sh
	@./keycloak/setup-keycloak.sh
	@echo "$(GREEN)Keycloak setup completed!$(RESET)"

keycloak-logs: ## View Keycloak logs
	@docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f keycloak

keycloak-restart: ## Restart Keycloak service
	@echo "$(YELLOW)Restarting Keycloak...$(RESET)"
	@docker-compose -f $(DOCKER_COMPOSE_DEV) restart keycloak
	@echo "$(GREEN)Keycloak restarted!$(RESET)"

keycloak-admin: ## Open Keycloak Admin Console
	@echo "$(BLUE)Opening Keycloak Admin Console...$(RESET)"
	@echo "$(GREEN)URL: http://localhost:8081$(RESET)"
	@echo "$(YELLOW)Username: admin$(RESET)"
	@echo "$(YELLOW)Password: admin123$(RESET)"
	@open http://localhost:8081 2>/dev/null || echo "$(YELLOW)Open http://localhost:8081 in your browser$(RESET)"

auth-test: ## Test authentication endpoints
	@echo "$(GREEN)Testing authentication...$(RESET)"
	@chmod +x keycloak/test-auth.sh
	@./keycloak/test-auth.sh

## Testing
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "$(GREEN)Running backend tests...$(RESET)"
	@cd $(BACKEND_DIR) && go test -v ./...
	@echo "$(GREEN)Backend tests completed!$(RESET)"

test-frontend: ## Run frontend tests
	@echo "$(GREEN)Running frontend tests...$(RESET)"
	@cd $(FRONTEND_DIR) && npm test
	@echo "$(GREEN)Frontend tests completed!$(RESET)"

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	@cd $(BACKEND_DIR) && go test -coverprofile=coverage.out ./...
	@cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@cd $(FRONTEND_DIR) && npm run test:coverage
	@echo "$(GREEN)Coverage reports generated!$(RESET)"

test-e2e: ## Run end-to-end tests
	@echo "$(GREEN)Running end-to-end tests...$(RESET)"
	@cd $(FRONTEND_DIR) && npm run test:e2e
	@echo "$(GREEN)E2E tests completed!$(RESET)"

## Code Quality
lint: lint-backend lint-frontend ## Run linting for all code

lint-backend: ## Run backend linting
	@echo "$(GREEN)Running backend linting...$(RESET)"
	@cd $(BACKEND_DIR) && golangci-lint run
	@echo "$(GREEN)Backend linting completed!$(RESET)"

lint-frontend: ## Run frontend linting
	@echo "$(GREEN)Running frontend linting...$(RESET)"
	@cd $(FRONTEND_DIR) && npm run lint
	@echo "$(GREEN)Frontend linting completed!$(RESET)"

lint-fix: ## Fix linting issues
	@echo "$(GREEN)Fixing linting issues...$(RESET)"
	@cd $(BACKEND_DIR) && golangci-lint run --fix
	@cd $(FRONTEND_DIR) && npm run lint:fix
	@echo "$(GREEN)Linting issues fixed!$(RESET)"

format: ## Format code
	@echo "$(GREEN)Formatting code...$(RESET)"
	@cd $(BACKEND_DIR) && gofmt -w .
	@cd $(FRONTEND_DIR) && npm run format
	@echo "$(GREEN)Code formatted!$(RESET)"

## Building
build: build-backend build-frontend ## Build all applications

build-backend: ## Build backend application
	@echo "$(GREEN)Building backend...$(RESET)"
	@cd $(BACKEND_DIR) && go build -o bin/api cmd/api/main.go
	@echo "$(GREEN)Backend built successfully!$(RESET)"

build-frontend: ## Build frontend application
	@echo "$(GREEN)Building frontend...$(RESET)"
	@cd $(FRONTEND_DIR) && npm run build
	@echo "$(GREEN)Frontend built successfully!$(RESET)"

## Docker Operations
docker-build: ## Build Docker images
	@echo "$(GREEN)Building Docker images...$(RESET)"
	@docker build -t $(PROJECT_NAME)-api:latest ./$(BACKEND_DIR)
	@docker build -t $(PROJECT_NAME)-frontend:latest ./$(FRONTEND_DIR)
	@echo "$(GREEN)Docker images built!$(RESET)"

docker-push: ## Push Docker images to registry
	@echo "$(GREEN)Pushing Docker images...$(RESET)"
	@docker push $(PROJECT_NAME)-api:latest
	@docker push $(PROJECT_NAME)-frontend:latest
	@echo "$(GREEN)Docker images pushed!$(RESET)"

docker-clean: ## Clean Docker images and containers
	@echo "$(YELLOW)Cleaning Docker images and containers...$(RESET)"
	@docker system prune -f
	@echo "$(GREEN)Docker cleanup completed!$(RESET)"

## Dependencies
deps-update: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(RESET)"
	@cd $(BACKEND_DIR) && go mod tidy && go mod download
	@cd $(FRONTEND_DIR) && npm update
	@echo "$(GREEN)Dependencies updated!$(RESET)"

deps-install: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(RESET)"
	@cd $(BACKEND_DIR) && go mod download
	@cd $(FRONTEND_DIR) && npm install
	@echo "$(GREEN)Dependencies installed!$(RESET)"

## Code Generation
generate: ## Generate code (GraphQL, mocks, etc.)
	@echo "$(GREEN)Generating code...$(RESET)"
	@cd $(BACKEND_DIR) && go generate ./...
	@cd $(FRONTEND_DIR) && npm run codegen
	@echo "$(GREEN)Code generation completed!$(RESET)"

## Kubernetes & Deployment
k8s-apply: ## Apply Kubernetes manifests
	@echo "$(GREEN)Applying Kubernetes manifests...$(RESET)"
	@kubectl apply -f k8s/
	@echo "$(GREEN)Kubernetes manifests applied!$(RESET)"

k8s-delete: ## Delete Kubernetes resources
	@echo "$(YELLOW)Deleting Kubernetes resources...$(RESET)"
	@kubectl delete -f k8s/
	@echo "$(GREEN)Kubernetes resources deleted!$(RESET)"

helm-install: ## Install Helm charts
	@echo "$(GREEN)Installing Helm charts...$(RESET)"
	@helm upgrade --install zplus-api ./helm/zplus-api -f ./helm/zplus-api/values.dev.yaml
	@helm upgrade --install zplus-frontend ./helm/zplus-frontend -f ./helm/zplus-frontend/values.dev.yaml
	@echo "$(GREEN)Helm charts installed!$(RESET)"

helm-uninstall: ## Uninstall Helm charts
	@echo "$(YELLOW)Uninstalling Helm charts...$(RESET)"
	@helm uninstall zplus-api
	@helm uninstall zplus-frontend
	@echo "$(GREEN)Helm charts uninstalled!$(RESET)"

## Deployment
deploy-staging: ## Deploy to staging environment
	@echo "$(GREEN)Deploying to staging...$(RESET)"
	@./scripts/deploy-staging.sh
	@echo "$(GREEN)Staging deployment completed!$(RESET)"

deploy-prod: ## Deploy to production environment
	@echo "$(GREEN)Deploying to production...$(RESET)"
	@./scripts/deploy-production.sh
	@echo "$(GREEN)Production deployment completed!$(RESET)"

## Monitoring & Health
health-check: ## Check health of all services
	@echo "$(GREEN)Checking service health...$(RESET)"
	@curl -f http://localhost:8080/health || echo "$(RED)API health check failed$(RESET)"
	@curl -f http://localhost:3000/api/health || echo "$(RED)Frontend health check failed$(RESET)"
	@echo "$(GREEN)Health checks completed!$(RESET)"

logs-api: ## Show API logs
	@docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f api

logs-frontend: ## Show frontend logs
	@docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f frontend

logs-db: ## Show database logs
	@docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f postgres mongodb redis

logs-all: ## Show all service logs
	@docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f

## Database Operations
db-console: ## Connect to PostgreSQL database
	@docker-compose -f $(DOCKER_COMPOSE_DEV) exec postgres psql -U postgres -d zplus

mongo-console: ## Connect to MongoDB database
	@docker-compose -f $(DOCKER_COMPOSE_DEV) exec mongodb mongosh

redis-console: ## Connect to Redis
	@docker-compose -f $(DOCKER_COMPOSE_DEV) exec redis redis-cli

## Utilities
clean: ## Clean build artifacts and dependencies
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BACKEND_DIR)/bin
	@rm -rf $(FRONTEND_DIR)/.next
	@rm -rf $(FRONTEND_DIR)/out
	@rm -rf $(BACKEND_DIR)/coverage.out
	@rm -rf $(BACKEND_DIR)/coverage.html
	@echo "$(GREEN)Cleanup completed!$(RESET)"

reset: clean dev-down ## Reset entire development environment
	@echo "$(YELLOW)Resetting development environment...$(RESET)"
	@docker-compose -f $(DOCKER_COMPOSE_DEV) down -v
	@docker system prune -f
	@echo "$(GREEN)Environment reset completed!$(RESET)"

docs-serve: ## Serve documentation locally
	@echo "$(GREEN)Serving documentation...$(RESET)"
	@cd $(DOCS_DIR) && python3 -m http.server 8000
	@echo "$(BLUE)Documentation available at: http://localhost:8000$(RESET)"

## Security
security-scan: ## Run security scans
	@echo "$(GREEN)Running security scans...$(RESET)"
	@cd $(BACKEND_DIR) && gosec ./...
	@cd $(FRONTEND_DIR) && npm audit
	@echo "$(GREEN)Security scans completed!$(RESET)"

## Performance
benchmark: ## Run performance benchmarks
	@echo "$(GREEN)Running performance benchmarks...$(RESET)"
	@cd $(BACKEND_DIR) && go test -bench=. ./...
	@echo "$(GREEN)Benchmarks completed!$(RESET)"

## Git Operations
git-hooks: ## Install Git hooks
	@echo "$(GREEN)Installing Git hooks...$(RESET)"
	@cp scripts/git-hooks/* .git/hooks/
	@chmod +x .git/hooks/*
	@echo "$(GREEN)Git hooks installed!$(RESET)"

## Development Tools
dev-backend: ## Start backend development server
	@echo "$(GREEN)Starting backend development server...$(RESET)"
	@cd $(BACKEND_DIR) && air

dev-frontend: ## Start frontend development server
	@echo "$(GREEN)Starting frontend development server...$(RESET)"
	@cd $(FRONTEND_DIR) && npm run dev

## Tenant Management
create-tenant: ## Create new tenant (usage: make create-tenant tenant=tenant_name)
	@if [ -z "$(tenant)" ]; then echo "$(RED)Error: Please provide tenant name. Usage: make create-tenant tenant=tenant_name$(RESET)"; exit 1; fi
	@echo "$(GREEN)Creating tenant: $(tenant)$(RESET)"
	@./scripts/create-tenant.sh $(tenant)
	@echo "$(GREEN)Tenant created successfully!$(RESET)"

delete-tenant: ## Delete tenant (usage: make delete-tenant tenant=tenant_name)
	@if [ -z "$(tenant)" ]; then echo "$(RED)Error: Please provide tenant name. Usage: make delete-tenant tenant=tenant_name$(RESET)"; exit 1; fi
	@echo "$(YELLOW)Deleting tenant: $(tenant)$(RESET)"
	@./scripts/delete-tenant.sh $(tenant)
	@echo "$(GREEN)Tenant deleted successfully!$(RESET)"

## Environment Information
info: ## Show environment information
	@echo "$(BLUE)Zplus SaaS Base - Environment Information$(RESET)"
	@echo ""
	@echo "$(YELLOW)Project:$(RESET) $(PROJECT_NAME)"
	@echo "$(YELLOW)Go Version:$(RESET) $(shell go version 2>/dev/null || echo 'Not installed')"
	@echo "$(YELLOW)Node Version:$(RESET) $(shell node --version 2>/dev/null || echo 'Not installed')"
	@echo "$(YELLOW)Docker Version:$(RESET) $(shell docker --version 2>/dev/null || echo 'Not installed')"
	@echo "$(YELLOW)Kubectl Version:$(RESET) $(shell kubectl version --client --short 2>/dev/null || echo 'Not installed')"
	@echo "$(YELLOW)Helm Version:$(RESET) $(shell helm version --short 2>/dev/null || echo 'Not installed')"
	@echo ""
	@echo "$(BLUE)Environment Files:$(RESET)"
	@if [ -f .env ]; then echo "  ✅ .env (main)"; else echo "  ❌ .env (missing - run 'make dev-setup')"; fi
	@if [ -f $(BACKEND_DIR)/.env ]; then echo "  ✅ backend/.env"; else echo "  ❌ backend/.env (missing)"; fi
	@if [ -f $(FRONTEND_DIR)/.env.local ]; then echo "  ✅ frontend/.env.local"; else echo "  ❌ frontend/.env.local (missing)"; fi
	@echo ""
	@echo "$(BLUE)Services Status:$(RESET)"
	@docker-compose -f $(DOCKER_COMPOSE_DEV) ps 2>/dev/null || echo "$(YELLOW)Development environment not running$(RESET)"

env-validate: ## Validate environment configuration
	@echo "$(GREEN)Validating environment configuration...$(RESET)"
	@echo ""
	@echo "$(BLUE)Checking required environment files:$(RESET)"
	@if [ -f .env ]; then echo "  ✅ .env exists"; else echo "  ❌ .env missing (run 'make dev-setup')"; exit 1; fi
	@if [ -f $(BACKEND_DIR)/.env ]; then echo "  ✅ backend/.env exists"; else echo "  ❌ backend/.env missing"; exit 1; fi
	@if [ -f $(FRONTEND_DIR)/.env.local ]; then echo "  ✅ frontend/.env.local exists"; else echo "  ❌ frontend/.env.local missing"; exit 1; fi
	@echo ""
	@echo "$(BLUE)Checking critical variables:$(RESET)"
	@if grep -q "^DATABASE_URL=" .env 2>/dev/null; then echo "  ✅ DATABASE_URL configured"; else echo "  ❌ DATABASE_URL missing"; fi
	@if grep -q "^MONGODB_URL=" .env 2>/dev/null; then echo "  ✅ MONGODB_URL configured"; else echo "  ❌ MONGODB_URL missing"; fi
	@if grep -q "^REDIS_URL=" .env 2>/dev/null; then echo "  ✅ REDIS_URL configured"; else echo "  ❌ REDIS_URL missing"; fi
	@if grep -q "^JWT_SECRET=" .env 2>/dev/null; then echo "  ✅ JWT_SECRET configured"; else echo "  ⚠️  JWT_SECRET missing or using default"; fi
	@echo ""
	@echo "$(GREEN)Environment validation complete!$(RESET)"

env-check-secrets: ## Check if default/weak secrets are being used
	@echo "$(YELLOW)Checking for default/weak secrets...$(RESET)"
	@echo ""
	@if grep -q "your-.*-secret\|dev-.*-secret\|not-for-production" .env 2>/dev/null; then \
		echo "$(RED)⚠️  WARNING: Default or weak secrets detected in .env$(RESET)"; \
		echo "$(YELLOW)Generate new secrets using: make generate-secrets$(RESET)"; \
	else \
		echo "$(GREEN)✅ No default secrets detected$(RESET)"; \
	fi

env-diff: ## Show differences between environment templates and current config
	@echo "$(BLUE)Environment Configuration Differences:$(RESET)"
	@echo ""
	@if [ -f .env ] && [ -f .env.example ]; then \
		echo "$(YELLOW)Main .env vs .env.example:$(RESET)"; \
		diff .env.example .env || true; \
		echo ""; \
	fi
	@if [ -f $(BACKEND_DIR)/.env ] && [ -f $(BACKEND_DIR)/.env.example ]; then \
		echo "$(YELLOW)Backend .env vs .env.example:$(RESET)"; \
		diff $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env || true; \
		echo ""; \
	fi

env-list: ## List all environment variables from .env files
	@echo "$(BLUE)Current Environment Variables:$(RESET)"
	@echo ""
	@if [ -f .env ]; then \
		echo "$(YELLOW)Main .env:$(RESET)"; \
		grep -E "^[A-Z_]+=.*" .env | sed 's/=.*/=***/' | head -20; \
		echo ""; \
	fi
	@if [ -f $(BACKEND_DIR)/.env ]; then \
		echo "$(YELLOW)Backend .env:$(RESET)"; \
		grep -E "^[A-Z_]+=.*" $(BACKEND_DIR)/.env | sed 's/=.*/=***/' | head -10; \
		echo ""; \
	fi
