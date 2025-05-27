#!/bin/bash
# Simple Phoenix Demo - Minimal manual test

echo "🚀 Phoenix Platform Simple Demo"
echo "=============================="
echo ""

# Check if postgres is running
if ! pg_isready -h localhost >/dev/null 2>&1; then
    echo "❌ PostgreSQL is not running. Please start it first."
    echo "   On macOS: brew services start postgresql"
    echo "   On Linux: sudo systemctl start postgresql"
    exit 1
fi

# Set environment
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/phoenix_demo?sslmode=disable"
export JWT_SECRET="demo-secret"
export PORT="8080"
export WEBSOCKET_PORT="8081"

# Create database
echo "Creating database..."
createdb phoenix_demo 2>/dev/null || echo "Database already exists"

# Run migrations
echo "Running migrations..."
psql $DATABASE_URL < projects/phoenix-api/migrations/001_core_tables.up.sql 2>/dev/null || true
psql $DATABASE_URL < projects/phoenix-api/migrations/002_ui_enhancements.up.sql 2>/dev/null || true
psql $DATABASE_URL < projects/phoenix-api/migrations/003_agent_tasks.up.sql 2>/dev/null || true

# Start API
echo ""
echo "Starting Phoenix API..."
./projects/phoenix-api/bin/phoenix-api &
API_PID=$!

# Cleanup on exit
trap "kill $API_PID 2>/dev/null; dropdb phoenix_demo 2>/dev/null" EXIT

# Wait for API
sleep 3

echo ""
echo "Testing API endpoints..."
echo "======================="

# 1. Health check
echo -e "\n1. Health Check:"
curl -s http://localhost:8080/health | jq '.'

# 2. Create experiment
echo -e "\n2. Creating Experiment:"
EXPERIMENT=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cost Optimization Test",
    "description": "Testing 70% cardinality reduction",
    "config": {
      "target_hosts": ["demo-host-1", "demo-host-2"],
      "baseline_template": {
        "name": "baseline",
        "config_url": "file:///configs/baseline.yaml"
      },
      "candidate_template": {
        "name": "adaptive",
        "config_url": "file:///configs/adaptive.yaml",
        "variables": {
          "threshold": "0.7"
        }
      },
      "duration": "5m",
      "warmup_duration": "1m"
    }
  }')

echo "$EXPERIMENT" | jq '.'
EXPERIMENT_ID=$(echo "$EXPERIMENT" | jq -r '.id')

# 3. List experiments
echo -e "\n3. List Experiments:"
curl -s http://localhost:8080/api/v1/experiments | jq '.experiments[] | {id, name, phase}'

# 4. Start experiment
echo -e "\n4. Starting Experiment:"
curl -s -X POST http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/start
echo "✓ Experiment started"

# 5. Simulate agent polling
echo -e "\n5. Agent Task Polling:"
curl -s -H "X-Agent-Host-ID: demo-host-1" \
  http://localhost:8080/api/v1/agent/tasks | jq '.'

# 6. Cost analysis
echo -e "\n6. Cost Analysis:"
curl -s http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/cost-analysis | \
  jq '.cost_analysis | {monthly_savings, yearly_savings, savings_percentage}'

# 7. Metric cost flow
echo -e "\n7. Real-time Metric Costs:"
curl -s http://localhost:8080/api/v1/metrics/cost-flow | \
  jq '{total_cost_per_minute, top_metrics: .top_metrics[:2]}'

# 8. Fleet status
echo -e "\n8. Fleet Status:"
curl -s http://localhost:8080/api/v1/fleet/status | jq '.'

echo ""
echo "================================"
echo "✅ Demo Complete!"
echo ""
echo "API is running at http://localhost:8080"
echo "WebSocket at ws://localhost:8081/api/v1/ws"
echo ""
echo "Try these CLI commands:"
echo "  ./projects/phoenix-cli/bin/phoenix-cli experiment list"
echo "  ./projects/phoenix-cli/bin/phoenix-cli experiment status $EXPERIMENT_ID"
echo ""
echo "Press Ctrl+C to stop"

# Keep running
while true; do sleep 1; done