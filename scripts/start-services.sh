#!/bin/bash

# Start Phoenix services individually

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

# Kill any existing services
pkill -f phoenix- 2>/dev/null || true

# Start API service
if [ -f "projects/platform-api/bin/api" ]; then
    log "Starting API service..."
    export DB_HOST=localhost DB_PORT=5432 DB_USER=phoenix DB_PASSWORD=phoenix DB_NAME=phoenix_db
    export REDIS_HOST=localhost REDIS_PORT=6379 REDIS_PASSWORD=phoenix
    export PORT=8080
    nohup projects/platform-api/bin/api > logs/api.log 2>&1 &
    echo $! > pids/api.pid
    success "API service started (PID: $!)"
else
    error "API binary not found"
fi

# Start Controller service
if [ -f "projects/controller/bin/controller" ]; then
    log "Starting Controller service..."
    export DB_HOST=localhost DB_PORT=5432 DB_USER=phoenix DB_PASSWORD=phoenix DB_NAME=experiments_db
    export PORT=8081
    nohup projects/controller/bin/controller > logs/controller.log 2>&1 &
    echo $! > pids/controller.pid
    success "Controller service started (PID: $!)"
else
    error "Controller binary not found"
fi

# Start Generator service
if [ -f "projects/generator/bin/generator" ]; then
    log "Starting Generator service..."
    export PORT=8082
    nohup projects/generator/bin/generator > logs/generator.log 2>&1 &
    echo $! > pids/generator.pid
    success "Generator service started (PID: $!)"
else
    error "Generator binary not found"
fi

# Wait and check status
sleep 3

log "Checking service status..."
for port in 8080 8081 8082; do
    if curl -s -f http://localhost:$port/health >/dev/null 2>&1; then
        success "Service on port $port is healthy"
    else
        error "Service on port $port is not responding"
    fi
done

log "Check logs in logs/ directory for any issues"