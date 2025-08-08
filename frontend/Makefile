# Zplus SaaS Frontend Makefile

.PHONY: help install dev build start lint test clean

# Default target
help:
	@echo "Zplus SaaS Frontend - Available Commands:"
	@echo ""
	@echo "  install     Install dependencies"
	@echo "  dev         Start development server"
	@echo "  build       Build for production"
	@echo "  start       Start production server"
	@echo "  lint        Run ESLint"
	@echo "  lint-fix    Fix ESLint issues"
	@echo "  test        Run tests"
	@echo "  test-watch  Run tests in watch mode"
	@echo "  test-cov    Run tests with coverage"
	@echo "  type-check  Run TypeScript type checking"
	@echo "  format      Format code with Prettier"
	@echo "  storybook   Start Storybook"
	@echo "  clean       Clean build artifacts"
	@echo "  docker-dev  Start with Docker Compose"
	@echo "  docker-prod Build production Docker image"

# Dependencies
install:
	npm install

# Development
dev:
	npm run dev

# Build
build:
	npm run build

# Production
start:
	npm run start

# Code Quality
lint:
	npm run lint

lint-fix:
	npm run lint:fix

format:
	npm run format

format-check:
	npm run format:check

type-check:
	npm run type-check

# Testing
test:
	npm run test

test-watch:
	npm run test:watch

test-cov:
	npm run test:coverage

# Documentation
storybook:
	npm run storybook

build-storybook:
	npm run build-storybook

# Clean
clean:
	rm -rf .next
	rm -rf out
	rm -rf dist
	rm -rf node_modules/.cache
	rm -rf coverage
	rm -rf storybook-static

# Docker
docker-dev:
	docker-compose up --build

docker-prod:
	docker build -t zplus-frontend:latest .

# Environment
setup-env:
	cp .env.example .env.local
	@echo "Please edit .env.local with your configuration"

# Full setup for new developers
setup: install setup-env
	@echo "Frontend setup complete!"
	@echo "Run 'make dev' to start development server"

# CI/CD
ci-install:
	npm ci

ci-build: ci-install
	npm run build

ci-test: ci-install
	npm run test:coverage
	npm run type-check
	npm run lint

# Health check
health:
	@curl -f http://localhost:3000/api/health || echo "Frontend not running"
