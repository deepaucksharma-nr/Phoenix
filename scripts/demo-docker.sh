#!/bin/bash

# Demo script to run Phoenix Platform end-to-end with Docker Compose

set -e

echo "=== Phoenix Platform End-to-End Demo ==="
echo "======================================="
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

wait_for_service() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=0
    
    print_info "Waiting for $service_name to be ready..."
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            print_info "$service_name is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    print_error "$service_name failed to start"
    return 1
}

# Check prerequisites
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is required but not installed."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    print_error "jq is required but not installed."
    exit 1
fi

# 1. Clean up and start services
print_step "Starting Phoenix Platform with Docker Compose"
docker-compose down -v 2>/dev/null || true
docker-compose up -d --build

# 2. Wait for services to be ready
print_step "Waiting for services to be ready"
wait_for_service "Phoenix API" "http://localhost:8080/health"
wait_for_service "Prometheus" "http://localhost:9090/-/ready"
wait_for_service "Pushgateway" "http://localhost:9091/metrics"

# Give database time to run migrations
print_info "Waiting for database migrations..."
sleep 10

# 3. Create an experiment
print_step "Creating a cost optimization experiment"
EXPERIMENT=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cost Optimization Demo",
    "description": "Testing 70% cardinality reduction with adaptive filtering",
    "config": {
      "target_hosts": ["local-agent-001"],
      "baseline_template": {
        "name": "baseline",
        "config_url": "file:///etc/otel-templates/baseline/config.yaml"
      },
      "candidate_template": {
        "name": "adaptive-filter",
        "config_url": "file:///etc/otel-templates/candidate/adaptive-filter-config.yaml",
        "variables": {
          "threshold": "0.7",
          "sampling_rate": "0.3"
        }
      },
      "duration": "5m",
      "warmup_duration": "1m"
    }
  }')

EXPERIMENT_ID=$(echo $EXPERIMENT | jq -r '.id' 2>/dev/null || echo "unknown")
print_info "Created experiment: $EXPERIMENT_ID"

# 4. Start the experiment
print_step "Starting the experiment"
curl -s -X POST "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/start" \
  -H "Content-Type: application/json"

# 5. Monitor experiment progress
print_step "Monitoring experiment progress"
for i in {1..10}; do
    STATUS=$(curl -s "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID" | jq -r '.phase' 2>/dev/null || echo "unknown")
    print_info "Experiment status: $STATUS"
    
    if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
        break
    fi
    
    sleep 10
done

# 6. Get experiment results
print_step "Fetching experiment results"
RESULTS=$(curl -s "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/analysis")
echo "$RESULTS" | jq '.' 2>/dev/null || echo "$RESULTS"

# 7. Check cost savings
print_step "Checking cost savings"
COST_FLOW=$(curl -s "http://localhost:8080/api/v1/cost-flow")
echo "Current cost flow:"
echo "$COST_FLOW" | jq '.' 2>/dev/null || echo "$COST_FLOW"

# 8. Check agent status
print_step "Checking agent fleet status"
FLEET_STATUS=$(curl -s "http://localhost:8080/api/v1/fleet/status")
echo "Fleet status:"
echo "$FLEET_STATUS" | jq '.' 2>/dev/null || echo "$FLEET_STATUS"

# 9. List all experiments
print_step "Listing all experiments"
curl -s "http://localhost:8080/api/v1/experiments" | jq '.' 2>/dev/null

# 10. Show WebSocket endpoint info
print_step "WebSocket endpoint information"
print_info "WebSocket available at: ws://localhost:8080/ws"
print_info "Dashboard available at: http://localhost:3000"

echo ""
echo "=== Demo Complete ==="
echo "===================="
echo ""
echo "Services running:"
echo "- Phoenix API: http://localhost:8080"
echo "- Dashboard: http://localhost:3000"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3001 (admin/admin)"
echo ""
echo "To view logs: docker-compose logs -f"
echo "To stop: docker-compose down"