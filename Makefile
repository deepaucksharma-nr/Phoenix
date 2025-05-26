# Phoenix Platform - Root Makefile
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Directories
ROOT_DIR := $(shell pwd)
BUILD_DIR := $(ROOT_DIR)/build
PROJECTS_DIR := $(ROOT_DIR)/projects
PKG_DIR := $(ROOT_DIR)/pkg
TOOLS_DIR := $(ROOT_DIR)/tools

# Version
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "0.0.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Docker
DOCKER_REGISTRY ?= ghcr.io/phoenix
DOCKER_BUILD_ARGS := \
	--build-arg VERSION=$(VERSION) \
	--build-arg GIT_COMMIT=$(GIT_COMMIT) \
	--build-arg BUILD_DATE=$(BUILD_DATE)

# Colors
CYAN := \033[0;36m
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Projects
ALL_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec basename {} \; 2>/dev/null)
GO_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec test -f {}/go.mod \; -print 2>/dev/null | xargs -n1 basename)
NODE_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec test -f {}/package.json \; -print 2>/dev/null | xargs -n1 basename)

# Core Projects
CORE_PROJECTS := phoenix-api phoenix-agent phoenix-cli dashboard

# Include shared makefiles
-include $(BUILD_DIR)/makefiles/*.mk

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all help clean build test lint fmt security docker release

## General Targets

all: validate build test ## Run validate, build, and test

help: ## Display this help message
	@echo -e "$(CYAN)Phoenix Platform - Monorepo Makefile$(NC)"
	@echo -e "$(CYAN)=====================================$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(CYAN)<target>$(NC)\n\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(CYAN)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo -e "$(GREEN)Project-specific targets:$(NC)"
	@echo -e "  $(CYAN)build-<project>$(NC)  Build specific project"
	@echo -e "  $(CYAN)test-<project>$(NC)   Test specific project"
	@echo -e "  $(CYAN)lint-<project>$(NC)   Lint specific project"
	@echo ""
	@echo -e "$(GREEN)Available projects:$(NC)"
	@for project in $(ALL_PROJECTS); do echo "  - $$project"; done

clean: $(ALL_PROJECTS:%=clean-%) ## Clean all build artifacts
	@echo -e "$(GREEN)✓ All projects cleaned$(NC)"

##@ Development

setup: ## Setup development environment
	@echo -e "$(CYAN)Setting up development environment...$(NC)"
	@$(TOOLS_DIR)/dev-env/setup.sh
	@echo -e "$(GREEN)✓ Development environment ready$(NC)"

dev-up: ## Start development services
	@echo -e "$(CYAN)Starting development services...$(NC)"
	@docker-compose up -d
	@echo -e "$(GREEN)✓ Services started$(NC)"
	@echo -e "$(YELLOW)Services:$(NC)"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000"

dev-down: ## Stop development services
	@echo -e "$(CYAN)Stopping development services...$(NC)"
	@docker-compose down
	@echo -e "$(GREEN)✓ Services stopped$(NC)"

dev-logs: ## Show development service logs
	@docker-compose logs -f

dev-reset: dev-down ## Reset development environment
	@echo -e "$(YELLOW)Removing volumes...$(NC)"
	@docker-compose down -v
	@echo -e "$(GREEN)✓ Development environment reset$(NC)"

##@ Building

build: $(GO_PROJECTS:%=build-%) $(NODE_PROJECTS:%=build-node-%) ## Build all projects
	@echo -e "$(GREEN)✓ All projects built$(NC)"

build-%: ## Build specific project
	@echo -e "$(CYAN)Building $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* build
	@echo -e "$(GREEN)✓ $* built$(NC)"

build-node-%: ## Build Node.js project
	@echo -e "$(CYAN)Building $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* build
	@echo -e "$(GREEN)✓ $* built$(NC)"

build-changed: ## Build only changed projects
	@echo -e "$(CYAN)Building changed projects...$(NC)"
	@$(BUILD_DIR)/scripts/ci/build-changed.sh
	@echo -e "$(GREEN)✓ Changed projects built$(NC)"

##@ Testing

test: $(ALL_PROJECTS:%=test-%) ## Run all tests
	@echo -e "$(GREEN)✓ All tests passed$(NC)"

test-%: ## Test specific project
	@echo -e "$(CYAN)Testing $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* test
	@echo -e "$(GREEN)✓ $* tests passed$(NC)"

test-integration: ## Run integration tests
	@echo -e "$(CYAN)Running integration tests...$(NC)"
	@$(MAKE) -C $(ROOT_DIR)/tests/integration test
	@echo -e "$(GREEN)✓ Integration tests passed$(NC)"

test-e2e: ## Run end-to-end tests
	@echo -e "$(CYAN)Running e2e tests...$(NC)"
	@$(MAKE) -C $(ROOT_DIR)/tests/e2e test
	@echo -e "$(GREEN)✓ E2E tests passed$(NC)"

test-coverage: ## Generate test coverage report
	@echo -e "$(CYAN)Generating coverage report...$(NC)"
	@$(BUILD_DIR)/scripts/ci/coverage.sh
	@echo -e "$(GREEN)✓ Coverage report generated$(NC)"

##@ Code Quality

lint: $(ALL_PROJECTS:%=lint-%) ## Lint all projects
	@echo -e "$(GREEN)✓ All projects linted$(NC)"

lint-%: ## Lint specific project
	@echo -e "$(CYAN)Linting $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* lint
	@echo -e "$(GREEN)✓ $* linted$(NC)"

fmt: $(ALL_PROJECTS:%=fmt-%) ## Format all code
	@echo -e "$(GREEN)✓ All code formatted$(NC)"

fmt-%: ## Format specific project
	@echo -e "$(CYAN)Formatting $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* fmt
	@echo -e "$(GREEN)✓ $* formatted$(NC)"

validate: ## Validate repository structure
	@echo -e "$(CYAN)Validating repository structure...$(NC)"
	@$(BUILD_DIR)/scripts/utils/validate-structure.sh
	@echo -e "$(GREEN)✓ Repository structure valid$(NC)"

##@ Security

security: $(ALL_PROJECTS:%=security-%) ## Run security scans
	@echo -e "$(GREEN)✓ Security scans completed$(NC)"

security-%: ## Security scan specific project
	@echo -e "$(CYAN)Scanning $* for vulnerabilities...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* security
	@echo -e "$(GREEN)✓ $* security scan completed$(NC)"

audit: ## Audit dependencies
	@echo -e "$(CYAN)Auditing dependencies...$(NC)"
	@$(TOOLS_DIR)/analyzers/dependency-check.sh
	@echo -e "$(GREEN)✓ Dependency audit completed$(NC)"

##@ Docker

docker: $(ALL_PROJECTS:%=docker-%) ## Build all Docker images
	@echo -e "$(GREEN)✓ All Docker images built$(NC)"

docker-%: ## Build Docker image for specific project
	@echo -e "$(CYAN)Building Docker image for $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* docker-build
	@echo -e "$(GREEN)✓ $* Docker image built$(NC)"

docker-push: $(ALL_PROJECTS:%=docker-push-%) ## Push all Docker images
	@echo -e "$(GREEN)✓ All Docker images pushed$(NC)"

docker-push-%: ## Push Docker image for specific project
	@echo -e "$(CYAN)Pushing Docker image for $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* docker-push
	@echo -e "$(GREEN)✓ $* Docker image pushed$(NC)"

##@ Kubernetes

k8s-generate: ## Generate Kubernetes manifests
	@echo -e "$(CYAN)Generating Kubernetes manifests...$(NC)"
	@$(BUILD_DIR)/scripts/k8s/generate-manifests.sh
	@echo -e "$(GREEN)✓ Kubernetes manifests generated$(NC)"

k8s-validate: ## Validate Kubernetes manifests
	@echo -e "$(CYAN)Validating Kubernetes manifests...$(NC)"
	@$(BUILD_DIR)/scripts/k8s/validate-manifests.sh
	@echo -e "$(GREEN)✓ Kubernetes manifests valid$(NC)"

k8s-deploy-dev: ## Deploy to development cluster
	@echo -e "$(CYAN)Deploying to development...$(NC)"
	@$(BUILD_DIR)/scripts/k8s/deploy.sh development
	@echo -e "$(GREEN)✓ Deployed to development$(NC)"

##@ Release

version: ## Display current version
	@echo $(VERSION)

changelog: ## Generate changelog
	@echo -e "$(CYAN)Generating changelog...$(NC)"
	@$(BUILD_DIR)/scripts/release/generate-changelog.sh
	@echo -e "$(GREEN)✓ Changelog generated$(NC)"

release: ## Create a new release
	@echo -e "$(CYAN)Creating release...$(NC)"
	@$(BUILD_DIR)/scripts/release/create-release.sh
	@echo -e "$(GREEN)✓ Release created$(NC)"

release-notes: ## Generate release notes
	@echo -e "$(CYAN)Generating release notes...$(NC)"
	@$(BUILD_DIR)/scripts/release/generate-notes.sh
	@echo -e "$(GREEN)✓ Release notes generated$(NC)"

##@ Utilities

generate: ## Run code generation
	@echo -e "$(CYAN)Running code generation...$(NC)"
	@$(MAKE) -C $(PKG_DIR) generate
	@for project in $(GO_PROJECTS); do \
		$(MAKE) -C $(PROJECTS_DIR)/$$project generate 2>/dev/null || true; \
	done
	@echo -e "$(GREEN)✓ Code generation completed$(NC)"

deps: ## Update dependencies
	@echo -e "$(CYAN)Updating dependencies...$(NC)"
	@go work sync
	@for project in $(GO_PROJECTS); do \
		echo -e "$(CYAN)Updating $$project dependencies...$(NC)"; \
		cd $(PROJECTS_DIR)/$$project && go mod tidy; \
	done
	@echo -e "$(GREEN)✓ Dependencies updated$(NC)"

tools: ## Install development tools
	@echo -e "$(CYAN)Installing development tools...$(NC)"
	@$(TOOLS_DIR)/install-tools.sh
	@echo -e "$(GREEN)✓ Development tools installed$(NC)"

##@ Phoenix UI (Revolutionary Experience)

ui-up: dev-up ## Start Phoenix with full UI experience
	@echo -e "$(CYAN)Starting Phoenix UI Experience...$(NC)"
	@./scripts/start-phoenix-ui.sh

ui-dev: ## Start UI development environment
	@echo -e "$(CYAN)Starting UI development mode...$(NC)"
	@docker-compose up -d postgres redis phoenix-api
	@cd projects/dashboard && npm install && npm run dev

ui-build: build-phoenix-api build-dashboard ## Build UI components
	@echo -e "$(GREEN)✓ UI components built$(NC)"

ui-test: test-phoenix-api test-dashboard ## Test UI components
	@echo -e "$(GREEN)✓ UI tests passed$(NC)"

##@ Dashboard

build-dashboard: ## Build dashboard
	@echo -e "$(CYAN)Building dashboard...$(NC)"
	@cd projects/dashboard && npm install && npm run build
	@echo -e "$(GREEN)✓ Dashboard built$(NC)"

run-dashboard: ## Run dashboard in development mode
	@echo -e "$(CYAN)Starting dashboard...$(NC)"
	@cd projects/dashboard && npm run dev

test-dashboard: ## Test dashboard
	@echo -e "$(CYAN)Testing dashboard...$(NC)"
	@cd projects/dashboard && npm test

##@ Phoenix Core Services

run-phoenix-api: ## Run Phoenix API with WebSocket support
	@echo -e "$(CYAN)Starting Phoenix API...$(NC)"
	@cd projects/phoenix-api && go run cmd/api/main.go

run-phoenix-agent: ## Run Phoenix Agent
	@echo -e "$(CYAN)Starting Phoenix Agent...$(NC)"
	@cd projects/phoenix-agent && go run cmd/phoenix-agent/main.go

run-phoenix: dev-up run-phoenix-api ## Run Phoenix platform (API + dependencies)
	@echo -e "$(GREEN)✓ Phoenix platform running$(NC)"

# Project-specific targets
$(foreach project,$(ALL_PROJECTS),$(eval $(call PROJECT_TARGET,$(project))))