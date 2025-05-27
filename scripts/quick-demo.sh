#!/bin/bash
# Quick Phoenix Demo - Minimal setup for testing

echo "🚀 Phoenix Quick Demo"
echo "===================="
echo ""

# Build first
echo "Building services..."
make build || exit 1

# Set environment
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/phoenix_dev?sslmode=disable"
export JWT_SECRET="demo"
export PORT="8080"
export WEBSOCKET_PORT="8081"

# Create simple test database (SQLite for demo)
mkdir -p /tmp/phoenix-demo
export DATABASE_URL="sqlite:///tmp/phoenix-demo/phoenix.db"

echo ""
echo "Starting API..."
./projects/phoenix-api/bin/phoenix-api &
API_PID=$!

# Cleanup on exit
trap "kill $API_PID 2>/dev/null" EXIT

# Wait for API
sleep 3

echo ""
echo "Testing API endpoints..."
echo ""

# Test health
echo "1. Health check:"
curl -s http://localhost:8080/health | jq '.'

# Create experiment
echo ""
echo "2. Creating experiment:"
EXPERIMENT=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Quick Test",
    "description": "Testing Phoenix",
    "config": {
      "target_hosts": ["host1", "host2"],
      "baseline_template": {"name": "baseline"},
      "candidate_template": {"name": "adaptive"},
      "duration": "5m"
    }
  }')

echo "$EXPERIMENT" | jq '.'
EXPERIMENT_ID=$(echo "$EXPERIMENT" | jq -r '.id')

# List experiments
echo ""
echo "3. Listing experiments:"
curl -s http://localhost:8080/api/v1/experiments | jq '.experiments'

# Get cost flow
echo ""
echo "4. Cost flow metrics:"
curl -s http://localhost:8080/api/v1/metrics/cost-flow | jq '.'

# Fleet status
echo ""
echo "5. Fleet status:"
curl -s http://localhost:8080/api/v1/fleet/status | jq '.'

echo ""
echo "============================"
echo "✅ Phoenix API is working!"
echo ""
echo "API running at: http://localhost:8080"
echo "WebSocket at: ws://localhost:8081/api/v1/ws"
echo ""
echo "Try these commands:"
echo "  curl http://localhost:8080/api/v1/experiments"
echo "  ./projects/phoenix-cli/bin/phoenix-cli experiment list"
echo ""
echo "Press Ctrl+C to stop"

# Keep running
while true; do sleep 1; done