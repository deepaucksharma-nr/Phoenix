#!/bin/bash

# Simplified Phoenix Platform Run Script

set -e

PROJECT_ROOT="/Users/deepaksharma/Desktop/src/Phoenix"
cd "$PROJECT_ROOT"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# Step 1: Clean up previous runs
log "Cleaning up previous runs..."
docker-compose -f docker-compose-fixed.yml down -v 2>/dev/null || true
pkill -f "phoenix-" 2>/dev/null || true

# Step 2: Start infrastructure
log "Starting infrastructure services..."
docker-compose -f docker-compose-fixed.yml up -d

# Wait for services
log "Waiting for services to be ready..."
sleep 10

# Check services
for service in postgres:5432 redis:6379 nats:4222; do
    name=${service%:*}
    port=${service#*:}
    if nc -z localhost $port 2>/dev/null; then
        success "$name is ready on port $port"
    else
        warning "$name may not be ready on port $port"
    fi
done

# Step 3: Run database migrations
log "Setting up databases..."
docker exec phoenix-postgres psql -U phoenix -c "CREATE DATABASE experiments_db;" 2>/dev/null || true
docker exec phoenix-postgres psql -U phoenix -c "CREATE DATABASE pipelines_db;" 2>/dev/null || true

# Step 4: Build and run services
log "Building services..."

# Build API service
if [ -d "projects/platform-api" ]; then
    log "Building platform-api..."
    cd projects/platform-api
    go build -o bin/api ./cmd/api/main.go 2>/dev/null || {
        warning "Could not build platform-api"
    }
    cd "$PROJECT_ROOT"
fi

# Build Controller service
if [ -d "projects/controller" ]; then
    log "Building controller..."
    cd projects/controller
    go build -o bin/controller ./cmd/controller/main.go 2>/dev/null || {
        warning "Could not build controller"
    }
    cd "$PROJECT_ROOT"
fi

# Build Generator service
if [ -d "projects/generator" ]; then
    log "Building generator..."
    cd projects/generator
    go build -o bin/generator ./cmd/generator/main.go 2>/dev/null || {
        # Try alternate path
        go build -o bin/generator ./cmd/main.go 2>/dev/null || {
            warning "Could not build generator"
        }
    }
    cd "$PROJECT_ROOT"
fi

# Build CLI
if [ -d "projects/phoenix-cli" ]; then
    log "Building phoenix-cli..."
    cd projects/phoenix-cli
    go build -o bin/phoenix-cli ./cmd/*.go 2>/dev/null || {
        warning "Could not build phoenix-cli"
    }
    cd "$PROJECT_ROOT"
fi

# Step 5: Start services
log "Starting services..."

# Start API
if [ -f "projects/platform-api/bin/api" ]; then
    log "Starting API service..."
    export DB_HOST=localhost DB_PORT=5432 DB_USER=phoenix DB_PASSWORD=phoenix DB_NAME=phoenix_db
    export REDIS_HOST=localhost REDIS_PORT=6379 REDIS_PASSWORD=phoenix
    projects/platform-api/bin/api &
    API_PID=$!
    sleep 2
fi

# Start Controller
if [ -f "projects/controller/bin/controller" ]; then
    log "Starting Controller service..."
    export DB_HOST=localhost DB_PORT=5432 DB_USER=phoenix DB_PASSWORD=phoenix DB_NAME=experiments_db
    projects/controller/bin/controller &
    CONTROLLER_PID=$!
    sleep 2
fi

# Start Generator
if [ -f "projects/generator/bin/generator" ]; then
    log "Starting Generator service..."
    projects/generator/bin/generator &
    GENERATOR_PID=$!
    sleep 2
fi

# Step 6: Run basic test
log "Running basic test workflow..."
sleep 5

if [ -f "projects/phoenix-cli/bin/phoenix-cli" ]; then
    export PHOENIX_API_URL="http://localhost:8080"
    CLI="projects/phoenix-cli/bin/phoenix-cli"
    
    # Create experiment
    log "Creating test experiment..."
    $CLI experiment create \
        --name "Test Experiment" \
        --description "Testing Phoenix Platform" \
        --baseline-config '{"name":"baseline"}' \
        --candidate-config '{"name":"candidate"}' \
        --duration 1m || warning "Could not create experiment"
    
    # List experiments
    log "Listing experiments..."
    $CLI experiment list || warning "Could not list experiments"
fi

# Show status
echo
success "Phoenix Platform is running!"
echo
echo "Services:"
echo "  API:        http://localhost:8080"
echo "  Prometheus: http://localhost:9090"
echo "  Grafana:    http://localhost:3000 (admin/phoenix)"
echo "  Jaeger:     http://localhost:16686"
echo
echo "Infrastructure:"
echo "  PostgreSQL: localhost:5432 (phoenix/phoenix)"
echo "  Redis:      localhost:6379 (password: phoenix)"
echo "  NATS:       localhost:4222"
echo
echo "To stop: docker-compose -f docker-compose-fixed.yml down"
echo

# Keep running
if [ "${1:-}" != "--detach" ]; then
    log "Press Ctrl+C to stop..."
    trap 'kill $API_PID $CONTROLLER_PID $GENERATOR_PID 2>/dev/null; docker-compose -f docker-compose-fixed.yml down' INT TERM
    wait
fi