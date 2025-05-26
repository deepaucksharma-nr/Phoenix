# loadsim-operator Makefile
include ../../build/makefiles/common.mk
include ../../build/makefiles/go.mk
include ../../build/makefiles/docker.mk

PROJECT_NAME := loadsim-operator
BINARY_NAME := loadsim-operator
MAIN_PATH := ./cmd/loadsim-operator

# Default target
.DEFAULT_GOAL := help

# Build targets
build: go-build
test: go-test
lint: go-lint
fmt: go-fmt
clean: go-clean clean-dirs

# Docker targets
docker: docker-build docker-simulator
docker-push: docker-push docker-push-simulator

# Load simulator Docker image
docker-simulator:
	@echo "Building load simulator Docker image..."
	@docker build -f build/Dockerfile -t phoenix/load-simulator:latest .

docker-push-simulator:
	@echo "Pushing load simulator Docker image..."
	@docker push phoenix/load-simulator:latest

# Development targets
dev:
	@air -c .air.toml

run: build
	@./bin/$(BINARY_NAME)

# Generate targets
generate: go-generate go-mocks

.PHONY: all build test lint fmt clean docker docker-push docker-simulator docker-push-simulator dev run generate
