# Phoenix API Makefile

.PHONY: all build test clean run docker migrate

# Variables
BINARY_NAME=phoenix-api
DOCKER_IMAGE=phoenix/api
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Default target
all: test build

# Build the binary
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run the application locally
run: build
	@echo "Running ${BINARY_NAME}..."
	./bin/${BINARY_NAME}

# Run database migrations
migrate:
	@echo "Running database migrations..."
	migrate -path migrations -database "${DATABASE_URL}" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "${DATABASE_URL}" down

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t ${DOCKER_IMAGE}:${VERSION} -t ${DOCKER_IMAGE}:latest .

# Push Docker image
docker-push: docker
	@echo "Pushing Docker image..."
	docker push ${DOCKER_IMAGE}:${VERSION}
	docker push ${DOCKER_IMAGE}:latest

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Generate code (if needed)
generate:
	@echo "Generating code..."
	go generate ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

# Run E2E tests
test-e2e:
	@echo "Running E2E tests..."
	go test -v -tags=e2e ./tests/e2e/...

# Development mode with hot reload
dev:
	@echo "Starting in development mode..."
	air -c .air.toml

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Check database connection
db-check:
	@echo "Checking database connection..."
	@go run cmd/api/main.go -check-db

# Performance profiling
profile:
	@echo "Running with profiling enabled..."
	go run -race cmd/api/main.go -cpuprofile=cpu.prof -memprofile=mem.prof

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...
	
# API documentation
docs:
	@echo "Generating API documentation..."
	swag init -g cmd/api/main.go -o docs

# Health check
health:
	@curl -f http://localhost:8080/health || exit 1

# Show help
help:
	@echo "Available targets:"
	@echo "  all              - Run tests and build"
	@echo "  build            - Build the binary"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e         - Run E2E tests"
	@echo "  run              - Build and run locally"
	@echo "  docker           - Build Docker image"
	@echo "  docker-push      - Push Docker image"
	@echo "  migrate          - Run database migrations"
	@echo "  migrate-down     - Rollback database migrations"
	@echo "  migrate-create   - Create a new migration"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  dev              - Run in development mode"
	@echo "  security         - Run security scan"
	@echo "  docs             - Generate API documentation"