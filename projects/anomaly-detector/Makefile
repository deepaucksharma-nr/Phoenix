# Anomaly Detector Service Makefile

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Directories
ROOT_DIR := $(shell pwd)
BUILD_DIR := $(ROOT_DIR)/build
CMD_DIR := $(ROOT_DIR)/cmd
API_DIR := $(ROOT_DIR)/api
SCRIPTS_DIR := $(ROOT_DIR)/scripts

# Version
VERSION ?= $(shell cat ../../VERSION 2>/dev/null || echo "0.0.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Colors
CYAN := \033[0;36m
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all help clean build test lint fmt proto

##@ General

all: build test ## Run build and test (proto skipped if protoc not available)

help: ## Display this help message
	@echo -e "$(CYAN)Anomaly Detector Service Makefile$(NC)"
	@echo -e "$(CYAN)===========================$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(CYAN)<target>$(NC)\n\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(CYAN)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

clean: ## Clean build artifacts
	@echo -e "$(CYAN)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo -e "$(GREEN)✓ Cleaned$(NC)"

##@ Building

build: ## Build the service
	@echo -e "$(CYAN)Building...$(NC)"
	@if command -v protoc >/dev/null 2>&1; then \
		$(MAKE) proto; \
	else \
		echo -e "$(YELLOW)Skipping proto generation (protoc not found)$(NC)"; \
	fi
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/anomaly-detector $(CMD_DIR)/anomaly-detector
	@echo -e "$(GREEN)✓ Built$(NC)"

##@ Testing

test: ## Run tests
	@echo -e "$(CYAN)Testing...$(NC)"
	@go test -v ./...
	@echo -e "$(GREEN)✓ Tests passed$(NC)"

test-coverage: ## Run tests with coverage
	@echo -e "$(CYAN)Running tests with coverage...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo -e "$(GREEN)✓ Coverage report: coverage.html$(NC)"

##@ Code Quality

lint: ## Run linters
	@echo -e "$(CYAN)Linting...$(NC)"
	@golangci-lint run
	@echo -e "$(GREEN)✓ Linted$(NC)"

fmt: ## Format code
	@echo -e "$(CYAN)Formatting...$(NC)"
	@go fmt ./...
	@echo -e "$(GREEN)✓ Formatted$(NC)"

##@ Proto

proto: ## Generate protobuf code
	@echo -e "$(CYAN)Generating protobuf code...$(NC)"
	@if command -v protoc >/dev/null 2>&1; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			api/*.proto; \
	else \
		echo -e "$(RED)protoc not found. Please install protobuf compiler$(NC)"; \
		exit 1; \
	fi
	@echo -e "$(GREEN)✓ Generated$(NC)" 