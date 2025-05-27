#!/bin/bash

# Working demo script for Phoenix Platform

set -e

echo "=== Phoenix Platform Working Demo ==="
echo "===================================="
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

if ! command -v curl &> /dev/null; then
    print_error "curl is required but not installed."
    exit 1
fi

# Kill any existing Phoenix API
pkill -f "phoenix-api" 2>/dev/null || true

# 1. Start Phoenix API
print_step "Starting Phoenix API"
cd projects/phoenix-api

# Use SQLite for simplicity
export DATABASE_URL="sqlite://phoenix.db"
export PORT=8080
export JWT_SECRET=development-secret
export ENABLE_WEBSOCKET=true

go run cmd/api/main.go &
API_PID=$!
print_info "Phoenix API started with PID: $API_PID"

# Wait for API to be ready
print_info "Waiting for API to be ready..."
sleep 3
until curl -s -f http://localhost:8080/health > /dev/null 2>&1; do
    sleep 1
done
print_info "API is ready!"

cd ../..

# 2. Test health endpoint
print_step "Testing health endpoint"
curl -s http://localhost:8080/health | jq '.' || echo "OK"

# 3. List experiments (should be empty or have existing ones)
print_step "Listing experiments"
curl -s http://localhost:8080/api/v1/experiments | jq '.'

# 4. Create a properly formatted experiment
print_step "Creating a new experiment"
EXPERIMENT=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Phoenix Working Demo",
    "description": "Demonstrating Phoenix Platform capabilities",
    "config": {
      "target_hosts": ["demo-agent-1", "demo-agent-2"],
      "baseline_template": {
        "name": "baseline",
        "config_url": "file:///tmp/baseline.yaml"
      },
      "candidate_template": {
        "name": "adaptive",
        "config_url": "file:///tmp/adaptive.yaml"
      },
      "duration": "5m",
      "warmup_duration": "1m"
    }
  }')

if echo "$EXPERIMENT" | grep -q '"id"'; then
    EXPERIMENT_ID=$(echo $EXPERIMENT | jq -r '.id')
    print_info "Created experiment: $EXPERIMENT_ID"
    echo "$EXPERIMENT" | jq '.'
else
    print_error "Failed to create experiment"
    echo "$EXPERIMENT"
fi

# 5. Get experiment details
if [ ! -z "$EXPERIMENT_ID" ] && [ "$EXPERIMENT_ID" != "null" ]; then
    print_step "Getting experiment details"
    curl -s "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID" | jq '.'
    
    # 6. Get experiment metrics
    print_step "Getting experiment metrics"
    curl -s "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/metrics" | jq '.'
fi

# 7. Test agent endpoints
print_step "Testing agent registration"
AGENT_REG=$(curl -s -X POST http://localhost:8080/api/v1/agents/register \
  -H "Content-Type: application/json" \
  -d '{
    "host_id": "demo-agent-1",
    "hostname": "demo-host-1",
    "ip_address": "10.0.0.1",
    "version": "0.1.0",
    "capabilities": ["metrics", "logs"],
    "tags": {"env": "demo", "region": "us-east"}
  }')
echo "$AGENT_REG" | jq '.' || echo "$AGENT_REG"

# 8. Check fleet status
print_step "Checking fleet status"
curl -s "http://localhost:8080/api/v1/fleet/status" | jq '.'

# 9. List pipelines
print_step "Listing available pipelines"
curl -s "http://localhost:8080/api/v1/pipelines" | jq '.'

# 10. Validate a pipeline configuration
print_step "Validating a pipeline configuration"
VALIDATION=$(curl -s -X POST http://localhost:8080/api/v1/pipelines/validate \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "receivers": {
        "otlp": {
          "protocols": {
            "grpc": {
              "endpoint": "0.0.0.0:4317"
            }
          }
        }
      },
      "processors": {
        "batch": {
          "timeout": "1s"
        }
      },
      "exporters": {
        "prometheus": {
          "endpoint": "0.0.0.0:8889"
        }
      },
      "service": {
        "pipelines": {
          "metrics": {
            "receivers": ["otlp"],
            "processors": ["batch"],
            "exporters": ["prometheus"]
          }
        }
      }
    }
  }')
echo "$VALIDATION" | jq '.'

# 11. Test WebSocket endpoint
print_step "Testing WebSocket endpoint"
print_info "WebSocket available at: ws://localhost:8080/ws"

# 12. Show available endpoints
print_step "Available API Endpoints"
echo "
Health:          GET  http://localhost:8080/health
Experiments:     GET  http://localhost:8080/api/v1/experiments
                 POST http://localhost:8080/api/v1/experiments
                 GET  http://localhost:8080/api/v1/experiments/{id}
                 POST http://localhost:8080/api/v1/experiments/{id}/start
                 POST http://localhost:8080/api/v1/experiments/{id}/stop
                 GET  http://localhost:8080/api/v1/experiments/{id}/metrics

Agents:          POST http://localhost:8080/api/v1/agents/register
                 GET  http://localhost:8080/api/v1/agents/{id}/tasks
                 POST http://localhost:8080/api/v1/agents/{id}/heartbeat

Fleet:           GET  http://localhost:8080/api/v1/fleet/status

Pipelines:       GET  http://localhost:8080/api/v1/pipelines
                 POST http://localhost:8080/api/v1/pipelines/validate
                 POST http://localhost:8080/api/v1/pipelines/render

Deployments:     POST http://localhost:8080/api/v1/deployments
                 GET  http://localhost:8080/api/v1/deployments
                 GET  http://localhost:8080/api/v1/deployments/{id}

WebSocket:       ws://localhost:8080/ws
"

echo ""
echo "=== Demo Running Successfully ==="
echo "================================="
echo ""
echo "Phoenix API is running at: http://localhost:8080 (PID: $API_PID)"
echo ""
echo "To stop the API: kill $API_PID"
echo ""

# Keep the script running
print_info "Press Ctrl+C to stop the demo..."
wait $API_PID