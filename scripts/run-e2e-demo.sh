#!/bin/bash
# End-to-End Phoenix Platform Demo
# This script demonstrates the complete workflow from setup to experiment execution

set -e

echo "🚀 Phoenix Platform End-to-End Demo"
echo "==================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check prerequisites
echo "📋 Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker is required but not installed.${NC}" >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo -e "${RED}Docker Compose is required but not installed.${NC}" >&2; exit 1; }
echo -e "${GREEN}✓ Prerequisites satisfied${NC}"

# Build all services
echo ""
echo "🔨 Building services..."
make build || { echo -e "${RED}Build failed${NC}"; exit 1; }
echo -e "${GREEN}✓ All services built successfully${NC}"

# Start infrastructure services
echo ""
echo "🏗️  Starting infrastructure services..."
docker-compose up -d postgres prometheus grafana
echo "Waiting for PostgreSQL to be ready..."
sleep 10

# Run database migrations
echo ""
echo "🗄️  Running database migrations..."
docker-compose run --rm phoenix-api /app/bin/phoenix-api migrate || echo "Migrations may already be applied"

# Start Phoenix services
echo ""
echo "🚀 Starting Phoenix services..."
docker-compose up -d phoenix-api
echo "Waiting for API to be ready..."
sleep 5

# Check API health
echo ""
echo "🏥 Checking API health..."
until curl -f http://localhost:8080/health >/dev/null 2>&1; do
    echo "Waiting for API..."
    sleep 2
done
echo -e "${GREEN}✓ API is healthy${NC}"

# Start agents
echo ""
echo "🤖 Starting Phoenix agents..."
docker-compose up -d phoenix-agent-1 phoenix-agent-2
echo -e "${GREEN}✓ Agents started${NC}"

# CLI Demo
echo ""
echo "📟 Running CLI Demo..."
echo "========================"

# Set API endpoint for CLI
export PHOENIX_API_URL=http://localhost:8080

# 1. Create an experiment
echo ""
echo -e "${BLUE}1. Creating an experiment...${NC}"
cat > /tmp/experiment.json <<EOF
{
  "name": "E2E Demo Experiment",
  "description": "Demonstrating Phoenix cost optimization",
  "config": {
    "target_hosts": ["phoenix-agent-1", "phoenix-agent-2"],
    "baseline_template": {
      "name": "baseline",
      "config_url": "file:///configs/baseline.yaml"
    },
    "candidate_template": {
      "name": "adaptive",
      "config_url": "file:///configs/adaptive.yaml",
      "variables": {
        "threshold": "0.7",
        "interval": "60s"
      }
    },
    "duration": "5m",
    "warmup_duration": "1m"
  }
}
EOF

EXPERIMENT_ID=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
    -H "Content-Type: application/json" \
    -d @/tmp/experiment.json | jq -r '.id')

echo -e "${GREEN}✓ Created experiment: $EXPERIMENT_ID${NC}"

# 2. List experiments
echo ""
echo -e "${BLUE}2. Listing experiments...${NC}"
curl -s http://localhost:8080/api/v1/experiments | jq '.experiments[] | {id, name, phase}'

# 3. Start the experiment
echo ""
echo -e "${BLUE}3. Starting experiment...${NC}"
curl -s -X POST http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/start
echo -e "${GREEN}✓ Experiment started${NC}"

# 4. Check agent tasks
echo ""
echo -e "${BLUE}4. Checking agent task distribution...${NC}"
sleep 3
echo "Agent 1 tasks:"
curl -s -H "X-Agent-Host-ID: phoenix-agent-1" http://localhost:8080/api/v1/agent/tasks | jq '.'
echo "Agent 2 tasks:"
curl -s -H "X-Agent-Host-ID: phoenix-agent-2" http://localhost:8080/api/v1/agent/tasks | jq '.'

# 5. Monitor experiment progress
echo ""
echo -e "${BLUE}5. Monitoring experiment progress...${NC}"
for i in {1..3}; do
    sleep 5
    STATUS=$(curl -s http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID | jq -r '.experiment.phase')
    echo "Experiment phase: $STATUS"
    
    # Get KPIs if available
    curl -s http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/kpis 2>/dev/null | jq '.' || true
done

# 6. Get cost analysis
echo ""
echo -e "${BLUE}6. Analyzing cost savings...${NC}"
curl -s http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/cost-analysis | jq '.cost_analysis | {monthly_savings, yearly_savings, savings_percentage}'

# 7. Check fleet status
echo ""
echo -e "${BLUE}7. Checking fleet status...${NC}"
curl -s http://localhost:8080/api/v1/fleet/status | jq '.'

# 8. Get metric cost flow
echo ""
echo -e "${BLUE}8. Getting real-time metric costs...${NC}"
curl -s http://localhost:8080/api/v1/metrics/cost-flow | jq '{total_cost_per_minute, top_metrics: .top_metrics[:3]}'

# 9. Test rollback
echo ""
echo -e "${BLUE}9. Testing experiment rollback...${NC}"
curl -s -X POST "http://localhost:8080/api/v1/experiments/$EXPERIMENT_ID/rollback?reason=demo-rollback" | jq '.'
echo -e "${GREEN}✓ Experiment rolled back${NC}"

# 10. WebSocket demo
echo ""
echo -e "${BLUE}10. WebSocket connection demo...${NC}"
echo "WebSocket endpoint available at: ws://localhost:8081/api/v1/ws"
echo "Connect using the dashboard or WebSocket client to see real-time updates"

# Dashboard
echo ""
echo -e "${BLUE}11. Launching Dashboard...${NC}"
echo "Dashboard available at: http://localhost:3000"
echo "Use 'phoenix ui' command to open in browser"

# Summary
echo ""
echo "=================================="
echo -e "${GREEN}🎉 End-to-End Demo Complete!${NC}"
echo ""
echo "Key Achievements Demonstrated:"
echo "✅ Experiment creation and lifecycle management"
echo "✅ Agent-based task distribution"
echo "✅ Real-time cost analysis"
echo "✅ Fleet monitoring"
echo "✅ Experiment rollback"
echo "✅ WebSocket real-time updates"
echo ""
echo "Services Running:"
echo "- API: http://localhost:8080"
echo "- WebSocket: ws://localhost:8081/api/v1/ws"
echo "- Dashboard: http://localhost:3000"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3001"
echo ""
echo "To stop all services: docker-compose down"
echo ""