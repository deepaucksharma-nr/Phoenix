#!/bin/bash

# Phoenix Platform Service Standardization Script
# Ensures all services follow the standard structure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🏗️  Phoenix Platform Service Standardization"
echo "=========================================="
echo ""

# Standard directories that should exist in each service
STANDARD_DIRS=(
    "cmd"
    "internal"
    "internal/config"
    "internal/handlers"
    "internal/services"
    "internal/store"
    "api"
)

# Standard files that should exist
STANDARD_FILES=(
    "Dockerfile"
    "Makefile"
    "README.md"
    "go.mod"
)

# Function to check and create standard structure
standardize_service() {
    local service=$1
    local service_path=$2
    
    echo -e "${BLUE}Checking service:${NC} $service"
    
    # Check/create standard directories
    for dir in "${STANDARD_DIRS[@]}"; do
        if [ ! -d "$service_path/$dir" ]; then
            echo -e "  ${YELLOW}Creating:${NC} $dir/"
            mkdir -p "$service_path/$dir"
            
            # Add .gitkeep to empty directories
            touch "$service_path/$dir/.gitkeep"
        fi
    done
    
    # Check for standard files
    for file in "${STANDARD_FILES[@]}"; do
        if [ ! -f "$service_path/$file" ]; then
            echo -e "  ${YELLOW}Missing:${NC} $file"
            
            # Create basic templates
            case $file in
                "README.md")
                    cat > "$service_path/$file" << EOF
# $service

## Overview
Service description goes here.

## Development

\`\`\`bash
# Build
make build

# Test
make test

# Run
make run
\`\`\`

## API Documentation
See [api/](./api/) directory for API specifications.
EOF
                    ;;
                "Makefile")
                    cat > "$service_path/$file" << EOF
.PHONY: build test run clean

SERVICE_NAME := $service
DOCKER_IMAGE := phoenix/\$(SERVICE_NAME):latest

build:
	go build -o bin/\$(SERVICE_NAME) ./cmd

test:
	go test -v ./...

run: build
	./bin/\$(SERVICE_NAME)

docker:
	docker build -t \$(DOCKER_IMAGE) .

clean:
	rm -rf bin/
EOF
                    ;;
            esac
        fi
    done
    
    # Check for main.go
    if [ ! -f "$service_path/cmd/main.go" ]; then
        echo -e "  ${YELLOW}Creating:${NC} cmd/main.go template"
        cat > "$service_path/cmd/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown gracefully
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutting down...")
        cancel()
    }()

    // Start service
    if err := run(ctx); err != nil {
        log.Fatal(err)
    }
}

func run(ctx context.Context) error {
    log.Printf("Starting $service...")
    
    // TODO: Implement service logic
    <-ctx.Done()
    
    return nil
}
EOF
    fi
    
    echo -e "  ${GREEN}✓${NC} Standardization complete"
    echo ""
}

# Check all services in projects/
echo "🔍 Scanning services in projects/ directory..."
echo ""

for service_dir in projects/*/; do
    if [ -d "$service_dir" ]; then
        service=$(basename "$service_dir")
        
        # Skip non-service directories
        if [[ "$service" == "scripts" ]] || [[ "$service" == "docs" ]]; then
            continue
        fi
        
        standardize_service "$service" "$service_dir"
    fi
done

# Create shared interfaces if they don't exist
echo "📦 Creating shared interfaces..."
mkdir -p pkg/interfaces

if [ ! -f "pkg/interfaces/service.go" ]; then
    cat > "pkg/interfaces/service.go" << 'EOF'
package interfaces

import "context"

// Service defines the standard interface for all Phoenix services
type Service interface {
    // Start begins the service operation
    Start(ctx context.Context) error
    
    // Stop gracefully shuts down the service
    Stop(ctx context.Context) error
    
    // Health returns nil if the service is healthy
    Health(ctx context.Context) error
    
    // Name returns the service name
    Name() string
}

// Store defines the standard interface for data stores
type Store interface {
    // Connect establishes connection to the store
    Connect(ctx context.Context) error
    
    // Close closes the store connection
    Close() error
    
    // Ping verifies the store is accessible
    Ping(ctx context.Context) error
}
EOF
    echo -e "${GREEN}✓${NC} Created pkg/interfaces/service.go"
fi

# Create a service template generator
cat > scripts/new-service.sh << 'EOF'
#!/bin/bash
# Creates a new Phoenix service with standard structure

if [ -z "$1" ]; then
    echo "Usage: $0 <service-name>"
    exit 1
fi

SERVICE_NAME=$1
SERVICE_PATH="projects/$SERVICE_NAME"

if [ -d "$SERVICE_PATH" ]; then
    echo "Service $SERVICE_NAME already exists!"
    exit 1
fi

echo "Creating new service: $SERVICE_NAME"

# Create service with standard structure
mkdir -p "$SERVICE_PATH"/{cmd,internal/{config,handlers,services,store},api}

# Copy templates and customize
echo "✓ Service structure created at $SERVICE_PATH"
echo "Next steps:"
echo "  1. cd $SERVICE_PATH"
echo "  2. go mod init github.com/phoenix/platform/projects/$SERVICE_NAME"
echo "  3. Implement service logic"
EOF

chmod +x scripts/new-service.sh
echo -e "${GREEN}✓${NC} Created service generator: scripts/new-service.sh"

echo ""
echo "📊 Standardization Summary"
echo "========================="
echo ""

# Count services
total_services=$(find projects -maxdepth 1 -type d | grep -v "^projects$" | wc -l)
echo "Total services: $total_services"

# Show services with missing standard components
echo ""
echo "⚠️  Services needing attention:"
for service_dir in projects/*/; do
    if [ -d "$service_dir" ]; then
        service=$(basename "$service_dir")
        missing=""
        
        # Check for critical files
        [ ! -f "$service_dir/Dockerfile" ] && missing="$missing Dockerfile"
        [ ! -f "$service_dir/Makefile" ] && missing="$missing Makefile"
        [ ! -f "$service_dir/cmd/main.go" ] && missing="$missing main.go"
        
        if [ -n "$missing" ]; then
            echo "  - $service: Missing$missing"
        fi
    fi
done

echo ""
echo -e "${GREEN}✅ Standardization complete!${NC}"
echo ""
echo "📋 Next steps:"
echo "  1. Review generated templates and customize"
echo "  2. Move service-specific code to internal/"
echo "  3. Update imports to use shared interfaces"
echo "  4. Run 'make build' in each service"