#!/bin/bash
# Local Phoenix Platform Demo (No Docker Required)
# This script runs the services locally for development/testing

set -e

echo "🚀 Phoenix Platform Local Demo"
echo "=============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Set up environment
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/phoenix_dev?sslmode=disable"
export JWT_SECRET="dev-secret-key"
export PROMETHEUS_URL="http://localhost:9090"
export PORT="8080"
export WEBSOCKET_PORT="8081"
export ENVIRONMENT="development"

# Check if PostgreSQL is running
echo "📋 Checking prerequisites..."
if ! pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  PostgreSQL is not running. Starting it...${NC}"
    # Try to start PostgreSQL (macOS)
    if command -v brew >/dev/null 2>&1; then
        brew services start postgresql@14 || brew services start postgresql
    else
        echo -e "${RED}Please start PostgreSQL manually${NC}"
        exit 1
    fi
    sleep 3
fi
echo -e "${GREEN}✓ PostgreSQL is running${NC}"

# Create database if it doesn't exist
echo ""
echo "🗄️  Setting up database..."
createdb phoenix_dev 2>/dev/null || echo "Database already exists"

# Run migrations
echo "Running migrations..."
for migration in projects/phoenix-api/migrations/*.up.sql; do
    echo "Applying $(basename $migration)..."
    psql $DATABASE_URL < $migration 2>/dev/null || echo "Migration may already be applied"
done
echo -e "${GREEN}✓ Database ready${NC}"

# Build services
echo ""
echo "🔨 Building services..."
make build
echo -e "${GREEN}✓ All services built${NC}"

# Start API in background
echo ""
echo "🚀 Starting Phoenix API..."
./projects/phoenix-api/bin/phoenix-api &
API_PID=$!
echo "API PID: $API_PID"

# Wait for API to be ready
echo "Waiting for API to start..."
sleep 5
until curl -f http://localhost:8080/health >/dev/null 2>&1; do
    echo "Waiting for API..."
    sleep 2
done
echo -e "${GREEN}✓ API is running on http://localhost:8080${NC}"

# Start Agent 1 in background
echo ""
echo "🤖 Starting Phoenix Agent 1..."
PHOENIX_HOST_ID="demo-agent-1" PHOENIX_API_URL="http://localhost:8080" \
    ./projects/phoenix-agent/bin/phoenix-agent &
AGENT1_PID=$!
echo "Agent 1 PID: $AGENT1_PID"

# Start Agent 2 in background
echo ""
echo "🤖 Starting Phoenix Agent 2..."
PHOENIX_HOST_ID="demo-agent-2" PHOENIX_API_URL="http://localhost:8080" \
    ./projects/phoenix-agent/bin/phoenix-agent &
AGENT2_PID=$!
echo "Agent 2 PID: $AGENT2_PID"

echo -e "${GREEN}✓ Agents started${NC}"
sleep 3

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    kill $API_PID $AGENT1_PID $AGENT2_PID 2>/dev/null || true
    echo "Services stopped"
}
trap cleanup EXIT

# Run demo workflow
echo ""
echo "📟 Running Demo Workflow"
echo "======================="

# 1. Create experiment via CLI
echo ""
echo -e "${BLUE}1. Creating experiment using CLI...${NC}"
cat > /tmp/exp-config.json <<EOF
{
  "name": "Local Demo Experiment",
  "description": "Testing Phoenix locally",
  "config": {
    "target_hosts": ["demo-agent-1", "demo-agent-2"],
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
}
EOF

EXPERIMENT_ID=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
    -H "Content-Type: application/json" \
    -d @/tmp/exp-config.json | jq -r '.id')

echo -e "${GREEN}✓ Created experiment: $EXPERIMENT_ID${NC}"

# 2. List experiments
echo ""
echo -e "${BLUE}2. Listing experiments...${NC}"
./projects/phoenix-cli/bin/phoenix-cli experiment list

# 3. Start experiment
echo ""
echo -e "${BLUE}3. Starting experiment...${NC}"
./projects/phoenix-cli/bin/phoenix-cli experiment start $EXPERIMENT_ID
echo -e "${GREEN}✓ Experiment started${NC}"

# 4. Check experiment status
echo ""
echo -e "${BLUE}4. Checking experiment status...${NC}"
./projects/phoenix-cli/bin/phoenix-cli experiment status $EXPERIMENT_ID

# 5. Check agent tasks
echo ""
echo -e "${BLUE}5. Verifying agent task distribution...${NC}"
echo "Checking tasks for agent 1..."
curl -s -H "X-Agent-Host-ID: demo-agent-1" http://localhost:8080/api/v1/agent/tasks | jq '.[]'

# 6. Simulate metrics push
echo ""
echo -e "${BLUE}6. Simulating agent metrics...${NC}"
curl -s -X POST http://localhost:8080/api/v1/agent/metrics \
    -H "X-Agent-Host-ID: demo-agent-1" \
    -H "Content-Type: application/json" \
    -d '[
        {"name": "cardinality_total", "value": 50000, "type": "gauge", "labels": {"variant": "baseline"}},
        {"name": "cardinality_total", "value": 15000, "type": "gauge", "labels": {"variant": "candidate"}},
        {"name": "cpu_percent", "value": 25.5, "type": "gauge", "labels": {"variant": "baseline"}},
        {"name": "cpu_percent", "value": 18.2, "type": "gauge", "labels": {"variant": "candidate"}}
    ]'
echo -e "${GREEN}✓ Metrics pushed${NC}"

# 7. Get cost analysis
echo ""
echo -e "${BLUE}7. Getting cost analysis...${NC}"
curl -s http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/cost-analysis | \
    jq '.cost_analysis | {monthly_savings, yearly_savings, savings_percentage}'

# 8. Check real-time metrics
echo ""
echo -e "${BLUE}8. Checking real-time metric costs...${NC}"
curl -s http://localhost:8080/api/v1/metrics/cost-flow | jq '.'

# 9. Test rollback
echo ""
echo -e "${BLUE}9. Testing experiment rollback...${NC}"
./projects/phoenix-cli/bin/phoenix-cli experiment rollback $EXPERIMENT_ID --reason "Demo complete"
echo -e "${GREEN}✓ Experiment rolled back${NC}"

# 10. Open dashboard
echo ""
echo -e "${BLUE}10. Dashboard Access${NC}"
echo "To open the dashboard, run in a new terminal:"
echo "  ./projects/phoenix-cli/bin/phoenix-cli ui"
echo ""
echo "Or access directly at: http://localhost:3000"

# WebSocket info
echo ""
echo -e "${BLUE}11. WebSocket Endpoint${NC}"
echo "WebSocket available at: ws://localhost:8081/api/v1/ws"
echo "Real-time updates will be broadcast here"

# Keep running
echo ""
echo "=================================="
echo -e "${GREEN}🎉 Demo Running Successfully!${NC}"
echo ""
echo "Services:"
echo "- API: http://localhost:8080"
echo "- WebSocket: ws://localhost:8081/api/v1/ws"
echo "- Health: http://localhost:8080/health"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Wait for user to stop
while true; do
    sleep 1
done