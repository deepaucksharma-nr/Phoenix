# Platform API Service Makefile

# Project specific variables
PROJECT_NAME := platform-api
BINARY_NAME := platform-api
MAIN_PATH := ./cmd/api

# Include shared makefiles
ROOT_DIR := $(shell git rev-parse --show-toplevel)
include $(ROOT_DIR)/build/makefiles/common.mk
include $(ROOT_DIR)/build/makefiles/go.mk
include $(ROOT_DIR)/build/makefiles/docker.mk

# Default target
.DEFAULT_GOAL := build

# Phony targets
.PHONY: all build test lint fmt clean run docker

## Build targets

all: clean lint test build ## Clean, lint, test and build

build: go-build ## Build the service binary

build-all: go-build-all ## Build for all platforms

## Test targets

test: go-test ## Run tests

test-unit: go-test-unit ## Run unit tests only

test-integration: go-test-integration ## Run integration tests

test-coverage: go-test-coverage ## Generate test coverage report

## Code quality targets

lint: go-lint ## Run linter

fmt: go-fmt ## Format code

vet: go-vet ## Run go vet

## Development targets

run: build ## Build and run the service
	$(call log_info,"Starting $(PROJECT_NAME)...")
	@$(BIN_DIR)/$(BINARY_NAME)

dev: ## Run with hot reload
	$(call log_info,"Starting $(PROJECT_NAME) in development mode...")
	@air -c .air.toml

generate: go-generate go-swagger ## Run code generation

## Docker targets

docker: docker-build ## Build Docker image

docker-push: docker-build-push ## Build and push Docker image

## Database targets

migrate-up: ## Run database migrations up
	$(call log_info,"Running migrations up...")
	@migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down
	$(call log_info,"Running migrations down...")
	@migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a new migration
	$(call log_info,"Creating migration: $(name)")
	@migrate create -ext sql -dir ./migrations -seq $(name)

## Utility targets

deps: go-mod-tidy ## Update dependencies

clean: clean-dirs go-clean ## Clean build artifacts

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)