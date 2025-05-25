#!/bin/bash
# phase-0-foundation.sh - Foundation setup phase implementation
# This phase creates the base directory structure and workspace configuration

set -euo pipefail

# Source libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/common.sh"
source "$SCRIPT_DIR/../lib/state-tracker.sh"

PHASE_ID="phase-0-foundation"

log_phase "Phase 0: Foundation Setup"

# Ensure we have the lock for this phase
if ! acquire_lock "$PHASE_ID" "$AGENT_ID"; then
    log_error "Failed to acquire lock for phase"
    exit 1
fi

# Track phase components
COMPONENTS=(
    "directory-structure"
    "workspace-config"
    "build-infrastructure"
    "git-configuration"
    "environment-setup"
)

# Initialize component tracking
for component in "${COMPONENTS[@]}"; do
    track_component "$PHASE_ID" "$component" "pending"
done

# Component 1: Create directory structure
log_info "Creating directory structure..."
track_component "$PHASE_ID" "directory-structure" "in_progress"

DIRECTORIES=(
    "services"
    "packages/common"
    "packages/contracts/openapi"
    "packages/contracts/proto"
    "packages/contracts/schemas"
    "packages/config"
    "packages/go-common"
    "packages/ui-components"
    "infrastructure/docker/compose"
    "infrastructure/kubernetes/base"
    "infrastructure/kubernetes/overlays/dev"
    "infrastructure/kubernetes/overlays/staging"
    "infrastructure/kubernetes/overlays/prod"
    "infrastructure/helm/charts"
    "infrastructure/terraform/modules"
    "infrastructure/terraform/environments"
    "monitoring/grafana/dashboards"
    "monitoring/grafana/provisioning"
    "monitoring/prometheus/rules"
    "monitoring/prometheus/alerts"
    "config/environments/dev"
    "config/environments/staging"
    "config/environments/prod"
    "tools/scripts"
    "tools/generators"
    "tests/integration/scenarios"
    "tests/e2e/flows"
    "tests/performance/load"
    "tests/contracts"
    "docs/architecture/decisions"
    "docs/api/rest"
    "docs/api/grpc"
    "docs/guides/developer"
    "docs/guides/operator"
    "docs/runbooks"
)

for dir in "${DIRECTORIES[@]}"; do
    create_directory "$dir"
done

track_component "$PHASE_ID" "directory-structure" "completed"
log_success "Directory structure created"

# Component 2: Setup workspace configuration
log_info "Setting up workspace configuration..."
track_component "$PHASE_ID" "workspace-config" "in_progress"

# Create package.json if it doesn't exist
if [[ ! -f package.json ]]; then
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
    "build:changed": "turbo run build --filter=[HEAD^1]",
    "test": "turbo run test",
    "test:unit": "turbo run test:unit",
    "test:integration": "turbo run test:integration",
    "lint": "turbo run lint",
    "dev": "turbo run dev --parallel",
    "clean": "turbo run clean && rm -rf node_modules",
    "prepare": "husky install"
  },
  "devDependencies": {
    "turbo": "^1.11.0",
    "@types/node": "^20.10.0",
    "typescript": "^5.3.0",
    "husky": "^8.0.0",
    "commitlint": "^18.4.0"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0"
  }
}
EOF
    log_success "Created package.json"
else
    log_info "package.json already exists"
fi

# Create turbo.json if it doesn't exist
if [[ ! -f turbo.json ]]; then
    cat > turbo.json << 'EOF'
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": ["**/.env.*local"],
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", "build/**", "bin/**"],
      "env": ["NODE_ENV"]
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": ["coverage/**"]
    },
    "lint": {
      "outputs": []
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "clean": {
      "cache": false
    }
  }
}
EOF
    log_success "Created turbo.json"
else
    log_info "turbo.json already exists"
fi

track_component "$PHASE_ID" "workspace-config" "completed"

# Component 3: Setup build infrastructure
log_info "Setting up build infrastructure..."
track_component "$PHASE_ID" "build-infrastructure" "in_progress"

# Create Makefile.common
if [[ ! -f Makefile.common ]]; then
    cp "$SCRIPT_DIR/../../Makefile.common" . 2>/dev/null || cat > Makefile.common << 'EOF'
# Common Makefile configuration
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Common variables
ROOT_DIR := $(shell git rev-parse --show-toplevel 2>/dev/null || pwd)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")
EOF
    log_success "Created Makefile.common"
fi

# Create root Makefile
if [[ ! -f Makefile ]]; then
    cat > Makefile << 'EOF'
# Phoenix Platform - Root Makefile
include Makefile.common

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Phoenix Platform - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: setup
setup: ## Initial setup
	npm install
	@echo "Setup complete!"

.PHONY: build
build: ## Build all services
	npm run build

.PHONY: test
test: ## Run all tests
	npm run test

.PHONY: dev
dev: ## Start development environment
	docker-compose up -d
	npm run dev

.PHONY: clean
clean: ## Clean all build artifacts
	npm run clean
EOF
    log_success "Created root Makefile"
fi

track_component "$PHASE_ID" "build-infrastructure" "completed"

# Component 4: Git configuration
log_info "Setting up git configuration..."
track_component "$PHASE_ID" "git-configuration" "in_progress"

# Create .gitignore
if [[ ! -f .gitignore ]]; then
    cat > .gitignore << 'EOF'
# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
bin/
*.exe

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local

# Logs
*.log
logs/

# Temporary
tmp/
.turbo/

# Migration
.migration/
migration-archive-*.tar.gz
EOF
    log_success "Created .gitignore"
fi

# Create .gitattributes
if [[ ! -f .gitattributes ]]; then
    cat > .gitattributes << 'EOF'
# Auto detect text files and perform LF normalization
* text=auto

# Force LF for these files
*.sh text eol=lf
*.yaml text eol=lf
*.yml text eol=lf
Makefile text eol=lf

# Binary files
*.png binary
*.jpg binary
*.gif binary
*.ico binary
*.exe binary
EOF
    log_success "Created .gitattributes"
fi

track_component "$PHASE_ID" "git-configuration" "completed"

# Component 5: Environment setup
log_info "Setting up environment files..."
track_component "$PHASE_ID" "environment-setup" "in_progress"

# Create .env.template
if [[ ! -f .env.template ]]; then
    cat > .env.template << 'EOF'
# Phoenix Platform Environment Configuration

# Environment
NODE_ENV=development
LOG_LEVEL=info

# Database
DATABASE_URL=postgres://phoenix:phoenix@localhost:5432/phoenix
REDIS_URL=redis://localhost:6379

# Services
API_GATEWAY_URL=http://localhost:8080
CONTROL_SERVICE_URL=http://localhost:8081
DASHBOARD_URL=http://localhost:3000

# Monitoring
PROMETHEUS_URL=http://localhost:9090
GRAFANA_URL=http://localhost:3000

# Security
JWT_SECRET=change-me-in-production
API_KEY=change-me-in-production
EOF
    log_success "Created .env.template"
fi

# Create VERSION file
if [[ ! -f VERSION ]]; then
    echo "0.1.0" > VERSION
    log_success "Created VERSION file"
fi

track_component "$PHASE_ID" "environment-setup" "completed"

# Run validations
log_info "Running phase validations..."

VALIDATIONS_PASSED=true

# Validation 1: Directory structure
if [[ -d services && -d packages && -d infrastructure && -d monitoring ]]; then
    record_validation "$PHASE_ID" "directory_structure" "passed" "All required directories exist"
else
    record_validation "$PHASE_ID" "directory_structure" "failed" "Missing required directories"
    VALIDATIONS_PASSED=false
fi

# Validation 2: Workspace files
if [[ -f package.json && -f turbo.json && -f Makefile ]]; then
    record_validation "$PHASE_ID" "workspace_files" "passed" "All workspace files exist"
else
    record_validation "$PHASE_ID" "workspace_files" "failed" "Missing workspace files"
    VALIDATIONS_PASSED=false
fi

# Validation 3: Git configuration
if [[ -f .gitignore && -f .gitattributes ]]; then
    record_validation "$PHASE_ID" "git_config" "passed" "Git configuration files exist"
else
    record_validation "$PHASE_ID" "git_config" "failed" "Missing git configuration"
    VALIDATIONS_PASSED=false
fi

# Release lock
release_lock "$PHASE_ID" "$AGENT_ID"

# Final status
if [[ "$VALIDATIONS_PASSED" == "true" ]]; then
    log_success "Phase 0: Foundation setup completed successfully!"
    exit 0
else
    log_error "Phase 0: Foundation setup completed with validation failures"
    exit 1
fi