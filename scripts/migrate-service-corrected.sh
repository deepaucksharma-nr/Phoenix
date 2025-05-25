#!/bin/bash
# migrate-service-corrected.sh - Migrate a service to the correct monorepo structure
# Usage: ./migrate-service-corrected.sh <service-name> <old-path> <service-type>
# Example: ./migrate-service-corrected.sh anomaly-detector apps/anomaly-detector go

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validate arguments
if [ $# -lt 3 ]; then
    echo -e "${RED}Usage: $0 <service-name> <old-path> <service-type>${NC}"
    echo "Example: $0 anomaly-detector apps/anomaly-detector go"
    exit 1
fi

SERVICE_NAME=$1
OLD_PATH="OLD_IMPLEMENTATION/$2"
SERVICE_TYPE=$3
NEW_PATH="services/$SERVICE_NAME"  # Correct path: services/, not projects/

echo -e "${YELLOW}Migrating $SERVICE_NAME from $OLD_PATH to $NEW_PATH${NC}"

# Check if old path exists
if [ ! -d "$OLD_PATH" ]; then
    echo -e "${RED}Error: $OLD_PATH does not exist${NC}"
    exit 1
fi

# Check if new path already exists
if [ -d "$NEW_PATH" ]; then
    echo -e "${RED}Error: $NEW_PATH already exists${NC}"
    exit 1
fi

# Create directory structure based on service type
echo "Creating directory structure..."
mkdir -p "$NEW_PATH"

if [ "$SERVICE_TYPE" = "go" ]; then
    # Go service structure
    mkdir -p "$NEW_PATH"/{cmd,internal,api,build,deployments,tests,docs,scripts,configs}
    mkdir -p "$NEW_PATH"/internal/{handlers,services,repositories,models,middleware}
elif [ "$SERVICE_TYPE" = "node" ] || [ "$SERVICE_TYPE" = "react" ]; then
    # Node/React service structure
    mkdir -p "$NEW_PATH"/{src,public,build,tests,docs,scripts,configs}
    if [ "$SERVICE_TYPE" = "react" ]; then
        mkdir -p "$NEW_PATH"/src/{components,pages,hooks,services,store,types}
    fi
fi

# Copy source code preserving structure
echo "Copying source code..."
if [ -d "$OLD_PATH/cmd" ]; then
    cp -r "$OLD_PATH/cmd"/* "$NEW_PATH/cmd/" 2>/dev/null || true
fi

if [ -d "$OLD_PATH/internal" ]; then
    cp -r "$OLD_PATH/internal"/* "$NEW_PATH/internal/" 2>/dev/null || true
fi

if [ -d "$OLD_PATH/pkg" ]; then
    # Note: pkg should go to packages/go-common, not service-specific
    echo -e "${YELLOW}Note: Found pkg/ directory - shared code should be migrated to packages/go-common${NC}"
fi

if [ -d "$OLD_PATH/src" ]; then
    cp -r "$OLD_PATH/src"/* "$NEW_PATH/src/" 2>/dev/null || true
fi

# Copy API definitions
if [ -d "$OLD_PATH/api" ]; then
    cp -r "$OLD_PATH/api"/* "$NEW_PATH/api/" 2>/dev/null || true
fi

# Copy configuration files
echo "Copying configuration files..."
if [ -f "$OLD_PATH/go.mod" ]; then
    cp "$OLD_PATH/go.mod" "$NEW_PATH/"
    cp "$OLD_PATH/go.sum" "$NEW_PATH/" 2>/dev/null || true
fi

if [ -f "$OLD_PATH/package.json" ]; then
    cp "$OLD_PATH/package.json" "$NEW_PATH/"
    # Update package name to include workspace prefix
    sed -i "s/\"name\": \".*\"/\"name\": \"@phoenix\/$SERVICE_NAME\"/" "$NEW_PATH/package.json"
fi

# Copy lock files
for lockfile in package-lock.json pnpm-lock.yaml yarn.lock; do
    if [ -f "$OLD_PATH/$lockfile" ]; then
        cp "$OLD_PATH/$lockfile" "$NEW_PATH/"
    fi
done

# Copy Docker files
echo "Copying Docker configuration..."
if [ -f "$OLD_PATH/Dockerfile" ]; then
    mkdir -p "$NEW_PATH/docker"
    cp "$OLD_PATH/Dockerfile" "$NEW_PATH/docker/Dockerfile"
fi

# Copy test files
echo "Copying test files..."
if [ -d "$OLD_PATH/test" ] || [ -d "$OLD_PATH/tests" ]; then
    cp -r "$OLD_PATH/test"/* "$NEW_PATH/tests/" 2>/dev/null || true
    cp -r "$OLD_PATH/tests"/* "$NEW_PATH/tests/" 2>/dev/null || true
fi

# Update import paths for Go services
if [ "$SERVICE_TYPE" = "go" ]; then
    echo "Updating Go import paths..."
    find "$NEW_PATH" -type f -name "*.go" -exec sed -i \
        -e "s|github.com/phoenix/|github.com/phoenix-vnext/|g" \
        -e "s|phoenix-platform/pkg/|packages/go-common/|g" \
        -e "s|OLD_IMPLEMENTATION/||g" \
        {} +
    
    # Update go.mod
    if [ -f "$NEW_PATH/go.mod" ]; then
        sed -i "s|module .*|module github.com/phoenix-vnext/services/$SERVICE_NAME|" "$NEW_PATH/go.mod"
        
        # Add replace directive for local packages
        echo "" >> "$NEW_PATH/go.mod"
        echo "replace github.com/phoenix-vnext/packages/go-common => ../../packages/go-common" >> "$NEW_PATH/go.mod"
    fi
fi

# Create Makefile based on service type
echo "Creating Makefile..."
if [ "$SERVICE_TYPE" = "go" ]; then
    cat > "$NEW_PATH/Makefile" << 'EOF'
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
EOF

elif [ "$SERVICE_TYPE" = "node" ] || [ "$SERVICE_TYPE" = "react" ]; then
    cat > "$NEW_PATH/Makefile" << 'EOF'
# Service Makefile
SERVICE_NAME := $(notdir $(CURDIR))

# Include common makefiles if they exist
-include ../../Makefile.common

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  install  - Install dependencies"
	@echo "  build    - Build the service"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linters"
	@echo "  docker   - Build Docker image"
	@echo "  dev      - Run development server"
	@echo "  clean    - Clean build artifacts"

.PHONY: install
install:
	@echo "Installing dependencies..."
	@npm install

.PHONY: build
build: install
	@echo "Building $(SERVICE_NAME)..."
	@npm run build

.PHONY: test
test:
	@echo "Running tests..."
	@npm test

.PHONY: lint
lint:
	@echo "Running linters..."
	@npm run lint

.PHONY: docker
docker:
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t $(SERVICE_NAME):latest .

.PHONY: dev
dev:
	@echo "Starting development server..."
	@npm run dev

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf node_modules/ dist/ build/ coverage/
EOF
fi

# Create README.md
echo "Creating README.md..."
cat > "$NEW_PATH/README.md" << EOF
# $SERVICE_NAME

## Overview

Service migrated from: \`$2\`

## Development

### Prerequisites

- Go 1.21+ (for Go services)
- Node.js 18+ (for Node services)
- Docker
- Make

### Quick Start

\`\`\`bash
# Install dependencies
make install

# Run tests
make test

# Build the service
make build

# Run locally
make run   # or 'make dev' for Node services
\`\`\`

## Docker

\`\`\`bash
# Build Docker image
make docker

# Run with docker-compose (from root)
docker-compose up $SERVICE_NAME
\`\`\`

## Configuration

Configuration is managed through environment variables and config files.
See \`configs/\` directory for examples.

## API Documentation

[TODO: Add API documentation]

## Testing

\`\`\`bash
# Unit tests
make test

# With coverage
make test-coverage
\`\`\`
EOF

# Create .gitignore
echo "Creating .gitignore..."
cat > "$NEW_PATH/.gitignore" << 'EOF'
# Dependencies
node_modules/
vendor/

# Build outputs
bin/
dist/
build/
*.exe

# Test artifacts
coverage/
*.test
*.out

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
temp/
EOF

# Create basic Kubernetes deployment
echo "Creating Kubernetes deployment..."
mkdir -p "$NEW_PATH/deployments/k8s"
cat > "$NEW_PATH/deployments/k8s/deployment.yaml" << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $SERVICE_NAME
  labels:
    app: $SERVICE_NAME
spec:
  replicas: 3
  selector:
    matchLabels:
      app: $SERVICE_NAME
  template:
    metadata:
      labels:
        app: $SERVICE_NAME
    spec:
      containers:
      - name: $SERVICE_NAME
        image: $SERVICE_NAME:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVICE_NAME
          value: $SERVICE_NAME
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: $SERVICE_NAME
spec:
  selector:
    app: $SERVICE_NAME
  ports:
  - port: 80
    targetPort: 8080
EOF

# Create package.json for Go services (for workspace integration)
if [ "$SERVICE_TYPE" = "go" ] && [ ! -f "$NEW_PATH/package.json" ]; then
    cat > "$NEW_PATH/package.json" << EOF
{
  "name": "@phoenix/$SERVICE_NAME",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "build": "make build",
    "test": "make test",
    "lint": "make lint",
    "clean": "make clean"
  }
}
EOF
fi

# Summary
echo -e "${GREEN}✓ Migration completed successfully!${NC}"
echo ""
echo "Summary:"
echo "- Source: $OLD_PATH"
echo "- Destination: $NEW_PATH"
echo "- Type: $SERVICE_TYPE"
echo "- Files migrated: $(find $NEW_PATH -type f | wc -l)"
echo ""
echo "Next steps:"
echo "1. Review the migrated code in $NEW_PATH"
echo "2. Update service-specific configurations"
echo "3. Run 'cd $NEW_PATH && make test' to verify"
echo "4. Update any cross-service dependencies"
echo ""
echo -e "${YELLOW}Note: Remember to update the root package.json workspaces if needed${NC}"