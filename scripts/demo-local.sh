#!/bin/bash

# Demo script to run Phoenix Platform locally (without Docker)

set -e

echo "=== Phoenix Platform Local Demo ==="
echo "==================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
print_step() {
    echo -e "${GREEN}[STEP]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
if ! command -v go &> /dev/null; then
    print_error "Go is required but not installed."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    print_error "jq is required but not installed."
    exit 1
fi

# 1. Start PostgreSQL with Docker (if not running)
print_step "Checking PostgreSQL..."
if ! docker ps | grep -q postgres; then
    print_info "Starting PostgreSQL with Docker..."
    docker run -d --name postgres-phoenix \
        -e POSTGRES_USER=phoenix \
        -e POSTGRES_PASSWORD=phoenix \
        -e POSTGRES_DB=phoenix \
        -p 5432:5432 \
        postgres:15-alpine
    
    print_info "Waiting for PostgreSQL to be ready..."
    sleep 10
else
    print_info "PostgreSQL is already running"
fi

# 2. Run migrations
print_step "Running database migrations"
cd projects/phoenix-api
if command -v migrate &> /dev/null; then
    migrate -path ./migrations -database "postgresql://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable" up
else
    print_info "migrate tool not found, skipping migrations"
fi

# 3. Start Phoenix API
print_step "Starting Phoenix API"
export DATABASE_URL="postgresql://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"
export PORT=8080
export JWT_SECRET=development-secret
export ENABLE_WEBSOCKET=true

go run cmd/api/main.go &
API_PID=$!
print_info "Phoenix API started with PID: $API_PID"

# Wait for API to be ready
print_info "Waiting for API to be ready..."
sleep 5
until curl -s -f http://localhost:8080/health > /dev/null 2>&1; do
    sleep 2
done
print_info "API is ready!"

# 4. Create an experiment
print_step "Creating a cost optimization experiment"
EXPERIMENT=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Local Demo Experiment",
    "description": "Testing 70% cardinality reduction locally",
    "config": {
      "target_hosts": ["local-host"],
      "baseline_template": {
        "name": "baseline",
        "config_url": "file:///tmp/baseline.yaml"
      },
      "candidate_template": {
        "name": "adaptive-filter",
        "config_url": "file:///tmp/adaptive.yaml",
        "variables": {
          "threshold": "0.7"
        }
      },
      "duration": "2m",
      "warmup_duration": "30s"
    }
  }')

EXPERIMENT_ID=$(echo $EXPERIMENT | jq -r '.id' 2>/dev/null || echo "unknown")
print_info "Created experiment: $EXPERIMENT_ID"
echo "$EXPERIMENT" | jq '.' 2>/dev/null || echo "$EXPERIMENT"

# 5. List experiments
print_step "Listing all experiments"
curl -s "http://localhost:8080/api/v1/experiments" | jq '.' 2>/dev/null

# 6. Check cost flow
print_step "Checking cost flow"
COST_FLOW=$(curl -s "http://localhost:8080/api/v1/cost-flow")
echo "Current cost flow:"
echo "$COST_FLOW" | jq '.' 2>/dev/null || echo "$COST_FLOW"

# 7. Check fleet status
print_step "Checking fleet status"
FLEET_STATUS=$(curl -s "http://localhost:8080/api/v1/fleet/status")
echo "Fleet status:"
echo "$FLEET_STATUS" | jq '.' 2>/dev/null || echo "$FLEET_STATUS"

# 8. Create a pipeline deployment
print_step "Creating a pipeline deployment"
DEPLOYMENT=$(curl -s -X POST "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/pipelines" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_name": "adaptive-filter",
    "config_url": "file:///tmp/adaptive.yaml",
    "target_hosts": ["local-host"],
    "variant": "A"
  }')
echo "$DEPLOYMENT" | jq '.' 2>/dev/null || echo "$DEPLOYMENT"

echo ""
echo "=== Demo Complete ==="
echo "===================="
echo ""
echo "Services running:"
echo "- Phoenix API: http://localhost:8080 (PID: $API_PID)"
echo "- WebSocket: ws://localhost:8080/ws"
echo ""
echo "To stop the API: kill $API_PID"
echo "To stop PostgreSQL: docker stop postgres-phoenix && docker rm postgres-phoenix"
echo ""

# Keep the script running
print_info "Press Ctrl+C to stop the demo..."
wait $API_PID