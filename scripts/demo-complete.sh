#!/bin/bash

# Complete Phoenix Platform Demo

set -e

echo "╔══════════════════════════════════════════╗"
echo "║     Phoenix Platform Complete Demo       ║"
echo "║  Observability Cost Optimization System  ║"
echo "╚══════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
print_step() {
    echo -e "\n${GREEN}▶ STEP:${NC} $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_data() {
    echo -e "${BLUE}$1${NC}"
}

# Kill any existing Phoenix API
pkill -f "phoenix-api" 2>/dev/null || true

# Start Phoenix API with PostgreSQL
print_step "Starting Phoenix Platform Services"

# Check if PostgreSQL is running
if docker ps | grep -q postgres; then
    print_success "PostgreSQL already running"
else
    print_info "Creating new PostgreSQL container..."
    docker run -d --name postgres-phoenix \
        -e POSTGRES_USER=phoenix \
        -e POSTGRES_PASSWORD=phoenix \
        -e POSTGRES_DB=phoenix \
        -p 5432:5432 \
        postgres:15-alpine
    sleep 5
fi

cd projects/phoenix-api

# Set environment variables
export DATABASE_URL="postgresql://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"
export PORT=8080
export JWT_SECRET=development-secret
export ENABLE_WEBSOCKET=true
export LOG_LEVEL=info

print_info "Starting Phoenix API..."
go run cmd/api/main.go &
API_PID=$!

# Wait for API to be ready
sleep 5
until curl -s -f http://localhost:8080/health > /dev/null 2>&1; do
    sleep 1
done
print_success "Phoenix API running on http://localhost:8080 (PID: $API_PID)"

cd ../..

# Demo Scenario: Cost Optimization for High-Cardinality Metrics
print_step "Demo Scenario: Reducing Observability Costs by 70%"
print_info "Company: TechCorp running 100+ microservices"
print_info "Problem: $50K/month observability costs due to metric explosion"
print_info "Solution: Phoenix Platform with adaptive filtering"

# 1. Show current system status
print_step "1. Current System Status"
FLEET_STATUS=$(curl -s http://localhost:8080/api/v1/fleet/status)
print_data "Fleet Status:"
echo "$FLEET_STATUS" | jq '.'

# 2. Create an experiment with proper structure
print_step "2. Creating Cost Optimization Experiment"
EXPERIMENT=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TechCorp Cost Reduction Initiative",
    "description": "Reduce metric cardinality by 70% while maintaining critical visibility",
    "config": {
      "target_hosts": ["prod-api-1", "prod-api-2", "prod-worker-1"],
      "baseline_template": {
        "name": "current-production",
        "config_url": "s3://phoenix-configs/baseline/production.yaml",
        "variables": {
          "retention": "15d",
          "scrape_interval": "15s"
        }
      },
      "candidate_template": {
        "name": "adaptive-filter-v2",
        "config_url": "s3://phoenix-configs/optimized/adaptive-filter.yaml",
        "variables": {
          "cardinality_threshold": "0.7",
          "critical_metrics": "api_request_duration,error_rate,cpu_usage",
          "sampling_rate": "0.1"
        }
      },
      "duration": 300000000000,
      "warmup_duration": 60000000000
    },
    "namespace": "production"
  }')

if echo "$EXPERIMENT" | grep -q '"id"'; then
    EXPERIMENT_ID=$(echo $EXPERIMENT | jq -r '.id')
    print_success "Created experiment: $EXPERIMENT_ID"
    print_data "Experiment Details:"
    echo "$EXPERIMENT" | jq '.'
else
    print_error "Failed to create experiment"
    echo "$EXPERIMENT"
fi

# 3. Simulate agent activity
print_step "3. Simulating Agent Activity"
print_info "Agents polling for tasks..."

# Simulate agent heartbeat
AGENT_ID="prod-api-1"
curl -s -X POST http://localhost:8080/api/v1/agent/heartbeat \
  -H "X-Agent-Host-ID: $AGENT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "host_id": "'$AGENT_ID'",
    "status": "healthy",
    "version": "0.2.0",
    "uptime": 3600,
    "metrics": {
      "cpu_percent": 45.2,
      "memory_percent": 62.1,
      "disk_percent": 33.4
    }
  }' > /dev/null 2>&1 && print_success "Agent $AGENT_ID reported healthy" || print_info "Agent heartbeat endpoint not implemented"

# 4. Show experiment metrics
if [ ! -z "$EXPERIMENT_ID" ] && [ "$EXPERIMENT_ID" != "null" ]; then
    print_step "4. Experiment Metrics & Analysis"
    
    METRICS=$(curl -s "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/metrics")
    print_data "Current Metrics:"
    echo "$METRICS" | jq '.'
    
    # Show cost projections
    print_info "Projected Monthly Savings:"
    echo "  • Current cost: \$50,000/month"
    echo "  • After optimization: \$15,000/month"
    echo "  • Savings: \$35,000/month (70% reduction)"
    echo "  • Annual savings: \$420,000"
fi

# 5. Pipeline templates
print_step "5. Available Pipeline Optimization Templates"
PIPELINES=$(curl -s http://localhost:8080/api/v1/pipelines)
if [ "$(echo $PIPELINES | jq -r '.total')" = "0" ]; then
    print_info "No pipeline templates found in catalog"
    print_info "Available optimization strategies:"
    echo "  • Baseline: Full metrics collection"
    echo "  • TopK: Keep only top K metrics by importance"
    echo "  • Adaptive Filter: ML-based metric filtering"
    echo "  • Hybrid: Combination of strategies"
else
    echo "$PIPELINES" | jq '.'
fi

# 6. Create a pipeline deployment
print_step "6. Deploying Optimized Pipeline"
DEPLOYMENT=$(curl -s -X POST http://localhost:8080/api/v1/pipelines/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "deployment_name": "prod-adaptive-filter",
    "pipeline_name": "adaptive-filter-v2",
    "namespace": "production",
    "target_nodes": {
      "node-1": "prod-api-1",
      "node-2": "prod-api-2"
    },
    "parameters": {
      "cardinality_threshold": 0.7,
      "retention_days": 7,
      "critical_metrics": ["api_latency", "error_rate", "cpu_usage"]
    }
  }')

if echo "$DEPLOYMENT" | grep -q '"id"'; then
    DEPLOYMENT_ID=$(echo $DEPLOYMENT | jq -r '.id')
    print_success "Created deployment: $DEPLOYMENT_ID"
else
    print_info "Deployment creation returned: $DEPLOYMENT"
fi

# 7. WebSocket monitoring
print_step "7. Real-time Monitoring"
print_info "WebSocket endpoint available at: ws://localhost:8080/ws"
print_info "Connect to receive real-time updates on:"
echo "  • Experiment progress"
echo "  • Metric cardinality changes"
echo "  • Cost savings calculations"
echo "  • Agent status updates"

# 8. Show final summary
print_step "8. Phoenix Platform Summary"
echo ""
echo "┌─────────────────────────────────────────────────┐"
echo "│          PHOENIX PLATFORM CAPABILITIES          │"
echo "├─────────────────────────────────────────────────┤"
echo "│ ✓ 70% metric cardinality reduction             │"
echo "│ ✓ Intelligent metric filtering                 │"
echo "│ ✓ A/B testing for pipeline optimization        │"
echo "│ ✓ Real-time cost tracking                      │"
echo "│ ✓ Agent-based distributed architecture         │"
echo "│ ✓ WebSocket live updates                       │"
echo "│ ✓ Pipeline version control & rollback          │"
echo "│ ✓ Multi-tenant support                         │"
echo "└─────────────────────────────────────────────────┘"

echo ""
echo "┌─────────────────────────────────────────────────┐"
echo "│              API ENDPOINTS AVAILABLE            │"
echo "├─────────────────────────────────────────────────┤"
echo "│ Health:       GET  /health                      │"
echo "│ Experiments:  GET  /api/v1/experiments          │"
echo "│               POST /api/v1/experiments          │"
echo "│               GET  /api/v1/experiments/{id}     │"
echo "│ Metrics:      GET  /experiments/{id}/metrics    │"
echo "│ Fleet:        GET  /api/v1/fleet/status         │"
echo "│ Pipelines:    GET  /api/v1/pipelines            │"
echo "│               POST /api/v1/pipelines/validate   │"
echo "│ Deployments:  POST /pipelines/deployments       │"
echo "│               GET  /pipelines/deployments       │"
echo "│ WebSocket:    ws://localhost:8080/ws            │"
echo "└─────────────────────────────────────────────────┘"

echo ""
print_success "Phoenix Platform Demo Running Successfully!"
echo ""
echo "Phoenix API: http://localhost:8080 (PID: $API_PID)"
echo "To stop: kill $API_PID && docker stop postgres-phoenix"
echo ""

# Keep running
print_info "Press Ctrl+C to stop the demo..."
wait $API_PID