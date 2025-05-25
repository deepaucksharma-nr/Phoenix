#!/bin/bash
# setup-workspace.sh - Initialize the monorepo workspace configuration
# Usage: ./setup-workspace.sh

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Monorepo Workspace Setup ===${NC}"
echo ""

# Create root package.json
echo -e "${YELLOW}Creating root package.json...${NC}"
cat > package.json << 'EOF'
{
  "name": "phoenix-vnext",
  "version": "1.0.0",
  "private": true,
  "description": "Phoenix Platform - Observability Cost Optimization System",
  "workspaces": [
    "packages/*",
    "services/*",
    "tools/*"
  ],
  "scripts": {
    "build": "turbo run build",
    "build:services": "turbo run build --filter='./services/*'",
    "build:packages": "turbo run build --filter='./packages/*'",
    "test": "turbo run test",
    "test:unit": "turbo run test:unit",
    "test:integration": "turbo run test:integration",
    "lint": "turbo run lint",
    "lint:fix": "turbo run lint:fix",
    "format": "turbo run format",
    "dev": "turbo run dev --parallel",
    "clean": "turbo run clean",
    "docker:build": "turbo run docker",
    "deps:check": "npm audit",
    "deps:update": "npm update",
    "workspace:info": "npm ls --depth=0"
  },
  "devDependencies": {
    "turbo": "^1.11.0",
    "@types/node": "^20.10.0",
    "typescript": "^5.3.0",
    "eslint": "^8.54.0",
    "prettier": "^3.1.0",
    "@commitlint/cli": "^18.4.0",
    "@commitlint/config-conventional": "^18.4.0",
    "husky": "^8.0.0",
    "lint-staged": "^15.1.0"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0"
  },
  "repository": {
    "type": "git",
    "url": "https://github.com/phoenix/phoenix-vnext.git"
  },
  "license": "Apache-2.0"
}
EOF

# Create Turborepo configuration
echo -e "${YELLOW}Creating turbo.json...${NC}"
cat > turbo.json << 'EOF'
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": ["**/.env.*local"],
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", "build/**", "bin/**", ".next/**", "!.next/cache/**"],
      "env": ["NODE_ENV", "CI"]
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": ["coverage/**"],
      "env": ["NODE_ENV", "CI"]
    },
    "test:unit": {
      "outputs": ["coverage/**"],
      "env": ["NODE_ENV"]
    },
    "test:integration": {
      "dependsOn": ["build"],
      "env": ["NODE_ENV", "DATABASE_URL", "REDIS_URL"]
    },
    "lint": {
      "outputs": [],
      "cache": false
    },
    "lint:fix": {
      "outputs": [],
      "cache": false
    },
    "format": {
      "outputs": [],
      "cache": false
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "docker": {
      "dependsOn": ["build"],
      "cache": false
    },
    "clean": {
      "cache": false
    }
  }
}
EOF

# Create common Makefile
echo -e "${YELLOW}Creating Makefile.common...${NC}"
cat > Makefile.common << 'EOF'
# Common Makefile variables and targets
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Colors
CYAN := \033[0;36m
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
NC := \033[0m

# Common variables
ROOT_DIR := $(shell git rev-parse --show-toplevel)
SERVICE_NAME := $(notdir $(CURDIR))
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")

# Docker
DOCKER_REGISTRY ?= ghcr.io/phoenix
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(SERVICE_NAME):$(VERSION)

# Helper functions
define log_info
	@echo -e "$(CYAN)[INFO]$(NC) $(1)"
endef

define log_success
	@echo -e "$(GREEN)[SUCCESS]$(NC) $(1)"
endef

define log_error
	@echo -e "$(RED)[ERROR]$(NC) $(1)"
endef

# Common targets
.PHONY: version
version:
	@echo $(VERSION)

.PHONY: help
help:
	@echo -e "$(CYAN)$(SERVICE_NAME) - Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-15s$(NC) %s\n", $$1, $$2}'
EOF

# Create root Makefile
echo -e "${YELLOW}Creating root Makefile...${NC}"
cat > Makefile << 'EOF'
# Phoenix Platform - Root Makefile
include Makefile.common

.DEFAULT_GOAL := help

## General Commands

.PHONY: setup
setup: ## Initial setup of the monorepo
	$(call log_info,"Setting up Phoenix monorepo...")
	@npm install
	@./scripts/setup-dev-env.sh || true
	$(call log_success,"Setup complete!")

.PHONY: build
build: ## Build all services and packages
	$(call log_info,"Building all components...")
	@npm run build
	$(call log_success,"Build complete!")

.PHONY: test
test: ## Run all tests
	$(call log_info,"Running all tests...")
	@npm run test
	$(call log_success,"All tests passed!")

.PHONY: lint
lint: ## Lint all code
	$(call log_info,"Linting all code...")
	@npm run lint

.PHONY: dev
dev: ## Start development environment
	$(call log_info,"Starting development environment...")
	@docker-compose up -d
	@npm run dev

.PHONY: clean
clean: ## Clean all build artifacts
	$(call log_info,"Cleaning build artifacts...")
	@npm run clean
	@rm -rf node_modules/
	$(call log_success,"Clean complete!")

## Docker Commands

.PHONY: docker-build
docker-build: ## Build all Docker images
	$(call log_info,"Building Docker images...")
	@npm run docker:build

.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	$(call log_info,"Starting services...")
	@docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all services
	$(call log_info,"Stopping services...")
	@docker-compose down

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	@docker-compose logs -f

## Migration Commands

.PHONY: migrate
migrate: ## Run the complete migration
	$(call log_info,"Running migration...")
	@./scripts/run-migration.sh

.PHONY: migrate-validate
migrate-validate: ## Validate the migration
	$(call log_info,"Validating migration...")
	@./scripts/validate-migration.sh

## Utility Commands

.PHONY: deps-check
deps-check: ## Check for security vulnerabilities
	@npm audit

.PHONY: deps-update
deps-update: ## Update dependencies
	@npm update

.PHONY: workspace-info
workspace-info: ## Show workspace information
	@npm run workspace:info

.PHONY: help
help: ## Show this help message
	@echo -e "$(CYAN)Phoenix Platform - Monorepo Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2}'
EOF

# Create .gitignore
echo -e "${YELLOW}Creating .gitignore...${NC}"
cat > .gitignore << 'EOF'
# Dependencies
node_modules/
vendor/
.pnp
.pnp.js

# Build outputs
dist/
build/
bin/
out/
.next/
*.exe
*.dll
*.so
*.dylib

# Testing
coverage/
.nyc_output/
*.test
*.out
test-results/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~
.project
.classpath
.settings/

# OS
.DS_Store
Thumbs.db
Desktop.ini

# Environment
.env
.env.local
.env.*.local
.env.development
.env.test
.env.production

# Logs
logs/
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*
lerna-debug.log*

# Temporary
tmp/
temp/
.tmp/
.cache/

# Turborepo
.turbo/

# Docker
.dockerignore.local

# Terraform
*.tfstate
*.tfstate.*
.terraform/

# Secrets
*.pem
*.key
*.crt
*.p12
secrets/
EOF

# Create .npmrc
echo -e "${YELLOW}Creating .npmrc...${NC}"
cat > .npmrc << 'EOF'
# Phoenix Platform npm configuration
engine-strict=true
save-exact=true
package-lock=true
EOF

# Create workspace directories
echo -e "${YELLOW}Creating workspace directories...${NC}"
mkdir -p packages/{common,contracts,config,go-common,ui-components}
mkdir -p services
mkdir -p tools
mkdir -p infrastructure/{docker,kubernetes,helm,terraform}
mkdir -p monitoring/{grafana/dashboards,prometheus/rules}
mkdir -p config/environments/{dev,staging,prod}
mkdir -p tests/{integration,e2e,performance,contracts}
mkdir -p scripts
mkdir -p docs

# Create package READMEs
echo -e "${YELLOW}Creating package documentation...${NC}"

# Common package
cat > packages/common/package.json << 'EOF'
{
  "name": "@phoenix/common",
  "version": "1.0.0",
  "description": "Common utilities for Phoenix platform",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "lint": "eslint src --ext .ts,.tsx"
  },
  "dependencies": {},
  "devDependencies": {
    "typescript": "^5.3.0",
    "jest": "^29.7.0",
    "@types/jest": "^29.5.0"
  }
}
EOF

# Contracts package
cat > packages/contracts/package.json << 'EOF'
{
  "name": "@phoenix/contracts",
  "version": "1.0.0",
  "description": "API contracts and schemas for Phoenix platform",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "generate": "npm run generate:openapi && npm run generate:proto",
    "generate:openapi": "echo 'TODO: OpenAPI generation'",
    "generate:proto": "echo 'TODO: Proto generation'"
  }
}
EOF

# Go common package
cat > packages/go-common/go.mod << 'EOF'
module github.com/phoenix-vnext/packages/go-common

go 1.21

require (
    go.uber.org/zap v1.26.0
    github.com/pkg/errors v0.9.1
)
EOF

echo -e "${GREEN}✓ Workspace setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run 'npm install' to install dependencies"
echo "2. Run 'make setup' to complete the setup"
echo "3. Start migrating services with './scripts/migrate-service-corrected.sh'"
echo ""
echo -e "${YELLOW}Note: The workspace is now properly configured for the monorepo structure${NC}"