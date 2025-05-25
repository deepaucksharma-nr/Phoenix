# Service Makefile
SERVICE_NAME := $(notdir $(CURDIR))
BINARY_NAME := $(SERVICE_NAME)
MAIN_PATH := ./cmd

# Include common makefiles if they exist
-include ../../Makefile.common

# Default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build    - Build the service"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linters"
	@echo "  docker   - Build Docker image"
	@echo "  run      - Run the service locally"
	@echo "  clean    - Clean build artifacts"

.PHONY: build
build:
	@echo "Building $(SERVICE_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: lint
lint:
	@echo "Running linters..."
	@golangci-lint run

.PHONY: docker
docker:
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t $(SERVICE_NAME):latest .

.PHONY: run
run: build
	@echo "Running $(SERVICE_NAME)..."
	@./bin/$(BINARY_NAME)

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage/
