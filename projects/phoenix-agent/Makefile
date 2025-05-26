# Phoenix Agent Makefile

.PHONY: all build test clean run docker install

# Variables
BINARY_NAME=phoenix-agent
DOCKER_IMAGE=phoenix/agent
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"
INSTALL_PATH=/usr/local/bin

# Default target
all: test build

# Build the binary
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/phoenix-agent/main.go

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 cmd/phoenix-agent/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 cmd/phoenix-agent/main.go
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 cmd/phoenix-agent/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 cmd/phoenix-agent/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run the agent locally
run: build
	@echo "Running ${BINARY_NAME}..."
	./bin/${BINARY_NAME} -api-url=http://localhost:8080

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t ${DOCKER_IMAGE}:${VERSION} -t ${DOCKER_IMAGE}:latest .

# Push Docker image
docker-push: docker
	@echo "Pushing Docker image..."
	docker push ${DOCKER_IMAGE}:${VERSION}
	docker push ${DOCKER_IMAGE}:latest

# Install agent on the system
install: build
	@echo "Installing ${BINARY_NAME} to ${INSTALL_PATH}..."
	sudo cp bin/${BINARY_NAME} ${INSTALL_PATH}/
	@echo "Creating systemd service..."
	sudo cp deployments/systemd/phoenix-agent.service /etc/systemd/system/
	sudo systemctl daemon-reload
	@echo "Phoenix Agent installed. To start: sudo systemctl start phoenix-agent"

# Uninstall agent
uninstall:
	@echo "Uninstalling ${BINARY_NAME}..."
	sudo systemctl stop phoenix-agent || true
	sudo systemctl disable phoenix-agent || true
	sudo rm -f ${INSTALL_PATH}/${BINARY_NAME}
	sudo rm -f /etc/systemd/system/phoenix-agent.service
	sudo systemctl daemon-reload
	@echo "Phoenix Agent uninstalled."

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

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Create release artifacts
release: clean build-all
	@echo "Creating release artifacts..."
	mkdir -p dist
	tar -czf dist/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz -C bin ${BINARY_NAME}-linux-amd64
	tar -czf dist/${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz -C bin ${BINARY_NAME}-linux-arm64
	tar -czf dist/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz -C bin ${BINARY_NAME}-darwin-amd64
	tar -czf dist/${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz -C bin ${BINARY_NAME}-darwin-arm64
	@echo "Release artifacts created in dist/"

# Run with debug logging
debug: build
	@echo "Running ${BINARY_NAME} with debug logging..."
	./bin/${BINARY_NAME} -log-level=debug

# Performance profiling
profile:
	@echo "Running with profiling enabled..."
	go run -race cmd/phoenix-agent/main.go -cpuprofile=cpu.prof -memprofile=mem.prof

# Show agent status (when installed)
status:
	@sudo systemctl status phoenix-agent || echo "Phoenix Agent is not installed as a service"

# Show logs (when installed)
logs:
	@sudo journalctl -u phoenix-agent -f

# Show help
help:
	@echo "Available targets:"
	@echo "  all              - Run tests and build"
	@echo "  build            - Build the binary"
	@echo "  build-all        - Build for multiple platforms"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  run              - Build and run locally"
	@echo "  docker           - Build Docker image"
	@echo "  docker-push      - Push Docker image"
	@echo "  install          - Install agent as system service"
	@echo "  uninstall        - Uninstall agent service"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  security         - Run security scan"
	@echo "  release          - Create release artifacts"
	@echo "  debug            - Run with debug logging"
	@echo "  status           - Show agent service status"
	@echo "  logs             - Show agent logs"