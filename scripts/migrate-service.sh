#!/bin/bash
# migrate-service.sh - Migrate a service from OLD_IMPLEMENTATION to new structure
# Usage: ./migrate-service.sh <service-name> <old-path> <service-type>
# Example: ./migrate-service.sh anomaly-detector apps/anomaly-detector go

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
NEW_PATH="projects/$SERVICE_NAME"

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

# Create new directory structure based on service type
echo "Creating directory structure..."
mkdir -p "$NEW_PATH"/{cmd,internal,api,build,deployments,tests,docs,scripts,configs}

if [ "$SERVICE_TYPE" = "go" ]; then
    mkdir -p "$NEW_PATH"/{pkg,migrations}
elif [ "$SERVICE_TYPE" = "node" ] || [ "$SERVICE_TYPE" = "react" ]; then
    mkdir -p "$NEW_PATH"/{src,public}
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
    cp -r "$OLD_PATH/pkg"/* "$NEW_PATH/pkg/" 2>/dev/null || true
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
    cp "$OLD_PATH/package-lock.json" "$NEW_PATH/" 2>/dev/null || true
    cp "$OLD_PATH/pnpm-lock.yaml" "$NEW_PATH/" 2>/dev/null || true
    cp "$OLD_PATH/yarn.lock" "$NEW_PATH/" 2>/dev/null || true
fi

# Copy Docker files
echo "Copying Docker configuration..."
if [ -f "$OLD_PATH/Dockerfile" ]; then
    cp "$OLD_PATH/Dockerfile" "$NEW_PATH/build/"
fi

if [ -f "$OLD_PATH/docker-compose.yml" ]; then
    cp "$OLD_PATH/docker-compose.yml" "$NEW_PATH/build/"
fi

# Copy test files
echo "Copying test files..."
if [ -d "$OLD_PATH/test" ]; then
    cp -r "$OLD_PATH/test"/* "$NEW_PATH/tests/" 2>/dev/null || true
fi

if [ -d "$OLD_PATH/tests" ]; then
    cp -r "$OLD_PATH/tests"/* "$NEW_PATH/tests/" 2>/dev/null || true
fi

# Update import paths for Go services
if [ "$SERVICE_TYPE" = "go" ]; then
    echo "Updating Go import paths..."
    find "$NEW_PATH" -type f -name "*.go" -exec sed -i \
        -e "s|github.com/phoenix/|github.com/phoenix-vnext/|g" \
        -e "s|OLD_IMPLEMENTATION/||g" \
        {} +
    
    # Update go.mod
    if [ -f "$NEW_PATH/go.mod" ]; then
        sed -i "s|module .*|module github.com/phoenix-vnext/projects/$SERVICE_NAME|" "$NEW_PATH/go.mod"
    fi
fi

# Create Makefile based on service type
echo "Creating Makefile..."
if [ "$SERVICE_TYPE" = "go" ]; then
    cat > "$NEW_PATH/Makefile" << EOF
# $SERVICE_NAME Makefile
include ../../build/makefiles/common.mk
include ../../build/makefiles/go.mk
include ../../build/makefiles/docker.mk

PROJECT_NAME := $SERVICE_NAME
BINARY_NAME := $SERVICE_NAME
MAIN_PATH := ./cmd/$SERVICE_NAME

# Default target
.DEFAULT_GOAL := help

# Build targets
build: go-build
test: go-test
lint: go-lint
fmt: go-fmt
clean: go-clean clean-dirs

# Docker targets
docker: docker-build
docker-push: docker-push

# Development targets
dev:
	@air -c .air.toml

run: build
	@./bin/\$(BINARY_NAME)

# Generate targets
generate: go-generate go-mocks

.PHONY: all build test lint fmt clean docker docker-push dev run generate
EOF

elif [ "$SERVICE_TYPE" = "node" ] || [ "$SERVICE_TYPE" = "react" ]; then
    cat > "$NEW_PATH/Makefile" << EOF
# $SERVICE_NAME Makefile
include ../../build/makefiles/common.mk
include ../../build/makefiles/node.mk
include ../../build/makefiles/docker.mk

PROJECT_NAME := $SERVICE_NAME

# Default target
.DEFAULT_GOAL := help

# Build targets
build: node-build
test: node-test
lint: node-lint
fmt: node-fmt
clean: node-clean

# Docker targets
docker: docker-build
docker-push: docker-push

# Development targets
dev: node-dev
preview: node-preview

# Type checking
typecheck: node-typecheck

.PHONY: all build test lint fmt clean docker docker-push dev preview typecheck
EOF
fi

# Create README.md
echo "Creating README.md..."
cat > "$NEW_PATH/README.md" << EOF
# $SERVICE_NAME

## Overview

[Brief description of the service]

## Architecture

[Service architecture and design decisions]

## Development

### Prerequisites

- Go 1.21+ (for Go services)
- Node.js 18+ (for Node services)
- Docker
- Make

### Setup

\`\`\`bash
# Install dependencies
make install

# Run tests
make test

# Build the service
make build
\`\`\`

### Running Locally

\`\`\`bash
# Start development server
make dev

# Or run the built binary
make run
\`\`\`

## Configuration

Configuration is managed through environment variables and config files.

See \`configs/\` directory for configuration examples.

## API Documentation

[Link to API documentation]

## Testing

\`\`\`bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run with coverage
make test-coverage
\`\`\`

## Deployment

\`\`\`bash
# Build Docker image
make docker

# Push to registry
make docker-push
\`\`\`

## Monitoring

- Metrics: Available at \`/metrics\`
- Health: Available at \`/health\`
- Ready: Available at \`/ready\`

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md)
EOF

# Create .gitignore
echo "Creating .gitignore..."
cat > "$NEW_PATH/.gitignore" << EOF
# Binaries
bin/
dist/
*.exe
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/
node_modules/

# Build directories
build/
.next/
.cache/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
.env.*.local

# Logs
*.log
logs/

# Temporary files
tmp/
temp/
EOF

# Create VERSION file
echo "0.1.0" > "$NEW_PATH/VERSION"

# Copy and update deployment files
if [ -d "$OLD_PATH/deployments" ]; then
    cp -r "$OLD_PATH/deployments"/* "$NEW_PATH/deployments/" 2>/dev/null || true
fi

if [ -d "$OLD_PATH/k8s" ]; then
    mkdir -p "$NEW_PATH/deployments/k8s"
    cp -r "$OLD_PATH/k8s"/* "$NEW_PATH/deployments/k8s/" 2>/dev/null || true
fi

# Create basic deployment files if they don't exist
if [ ! -f "$NEW_PATH/deployments/k8s/deployment.yaml" ]; then
    mkdir -p "$NEW_PATH/deployments/k8s"
    cat > "$NEW_PATH/deployments/k8s/deployment.yaml" << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $SERVICE_NAME
  labels:
    app.kubernetes.io/name: $SERVICE_NAME
    app.kubernetes.io/part-of: phoenix-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: $SERVICE_NAME
  template:
    metadata:
      labels:
        app.kubernetes.io/name: $SERVICE_NAME
    spec:
      containers:
      - name: $SERVICE_NAME
        image: ghcr.io/phoenix/$SERVICE_NAME:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: ENVIRONMENT
          value: production
        livenessProbe:
          httpGet:
            path: /health
            port: http
        readinessProbe:
          httpGet:
            path: /ready
            port: http
EOF
fi

# Log migration details
echo -e "${GREEN}✓ Migration completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Review the migrated code in $NEW_PATH"
echo "2. Update any service-specific configurations"
echo "3. Run 'make test' to verify the migration"
echo "4. Update CI/CD pipelines if needed"
echo ""
echo "Migration summary:"
echo "- Source: $OLD_PATH"
echo "- Destination: $NEW_PATH"
echo "- Type: $SERVICE_TYPE"
echo "- Files migrated: $(find $NEW_PATH -type f | wc -l)"