# Phoenix CLI Makefile

.PHONY: build test lint fmt clean install

# Build variables
BINARY_NAME=phoenix
BUILD_DIR=./bin
GO=go
GOFLAGS=-mod=mod

# Version information
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build the CLI binary
build:
	@echo "Building Phoenix CLI..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Install the CLI to system path
install: build
	@echo "Installing Phoenix CLI..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Phoenix CLI installed to /usr/local/bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Please install it from https://golangci-lint.run/usage/install/"; \
	fi

# Format the code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Run the CLI
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

# Generate completion scripts
completions: build
	@echo "Generating completion scripts..."
	@mkdir -p $(BUILD_DIR)/completions
	@$(BUILD_DIR)/$(BINARY_NAME) completion bash > $(BUILD_DIR)/completions/$(BINARY_NAME).bash
	@$(BUILD_DIR)/$(BINARY_NAME) completion zsh > $(BUILD_DIR)/completions/$(BINARY_NAME).zsh
	@$(BUILD_DIR)/$(BINARY_NAME) completion fish > $(BUILD_DIR)/completions/$(BINARY_NAME).fish

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the CLI binary"
	@echo "  install     - Install the CLI to /usr/local/bin"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  run         - Build and run the CLI"
	@echo "  completions - Generate shell completion scripts"
	@echo "  help        - Show this help message"