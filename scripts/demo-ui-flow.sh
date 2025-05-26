#!/bin/bash
# Demo flow for Phoenix UI Revolution

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

API_URL=${PHOENIX_API_URL:-http://localhost:8080}

echo -e "${BLUE}=== Phoenix UI Revolution Demo ===${NC}"
echo -e "${YELLOW}This demo showcases the revolutionary UI features${NC}\n"

# Step 1: Check services
echo -e "${GREEN}Step 1: Checking services...${NC}"
if ! curl -s $API_URL/health > /dev/null; then
    echo -e "${RED}Phoenix API is not running. Start it with: ./scripts/start-phoenix-ui.sh${NC}"
    exit 1
fi
echo -e "✓ Phoenix API is running"

# Step 2: Get fleet status
echo -e "\n${GREEN}Step 2: Checking fleet status...${NC}"
FLEET_STATUS=$(curl -s $API_URL/api/v1/fleet/status)
TOTAL_AGENTS=$(echo $FLEET_STATUS | jq -r '.total_agents')
HEALTHY_AGENTS=$(echo $FLEET_STATUS | jq -r '.healthy_agents')
echo -e "Fleet Status: ${HEALTHY_AGENTS}/${TOTAL_AGENTS} agents healthy"

# Step 3: Show current cost flow
echo -e "\n${GREEN}Step 3: Current metric cost flow...${NC}"
COST_FLOW=$(curl -s $API_URL/api/v1/metrics/cost-flow)
TOTAL_COST=$(echo $COST_FLOW | jq -r '.total_cost_rate')
echo -e "Current cost rate: ₹${TOTAL_COST}/minute"
echo -e "Top cost drivers:"
echo $COST_FLOW | jq -r '.top_metrics[:3] | .[] | "  - \(.metric_name): ₹\(.cost_per_minute)/min (\(.percentage)%)"'

# Step 4: Show available pipeline templates
echo -e "\n${GREEN}Step 4: Available optimization templates...${NC}"
TEMPLATES=$(curl -s $API_URL/api/v1/pipelines/templates)
echo $TEMPLATES | jq -r '.[] | "  - \(.name): \(.estimated_savings_percent)% savings"'

# Step 5: Create experiment using wizard
echo -e "\n${GREEN}Step 5: Creating experiment via wizard...${NC}"
EXPERIMENT_DATA='{
  "name": "Demo Cost Optimization",
  "description": "Reduce metrics cost using Top-K filter",
  "host_selector": ["env=demo"],
  "pipeline_type": "top-k-20",
  "duration_hours": 1
}'

EXPERIMENT=$(curl -s -X POST $API_URL/api/v1/experiments/wizard \
  -H "Content-Type: application/json" \
  -d "$EXPERIMENT_DATA")

EXPERIMENT_ID=$(echo $EXPERIMENT | jq -r '.id')
echo -e "✓ Created experiment: $EXPERIMENT_ID"

# Step 6: Preview pipeline impact
echo -e "\n${GREEN}Step 6: Previewing pipeline impact...${NC}"
PREVIEW_DATA='{
  "pipeline_config": {
    "processors": [{"type": "top_k", "config": {"k": 20}}]
  },
  "target_hosts": ["demo-host"]
}'

PREVIEW=$(curl -s -X POST $API_URL/api/v1/pipelines/preview \
  -H "Content-Type: application/json" \
  -d "$PREVIEW_DATA")

echo -e "Estimated impact:"
echo -e "  - Cost reduction: $(echo $PREVIEW | jq -r '.estimated_cost_reduction')%"
echo -e "  - CPU impact: +$(echo $PREVIEW | jq -r '.estimated_cpu_impact')%"
echo -e "  - Memory impact: +$(echo $PREVIEW | jq -r '.estimated_memory_impact')MB"

# Step 7: Quick deploy demo
echo -e "\n${GREEN}Step 7: Quick deploy pipeline...${NC}"
DEPLOY_DATA='{
  "pipeline_template": "priority-sli-slo",
  "target_hosts": ["group:demo"],
  "auto_rollback": true
}'

DEPLOYMENT=$(curl -s -X POST $API_URL/api/v1/pipelines/quick-deploy \
  -H "Content-Type: application/json" \
  -d "$DEPLOY_DATA")

DEPLOYMENT_ID=$(echo $DEPLOYMENT | jq -r '.deployment_id')
echo -e "✓ Deployment started: $DEPLOYMENT_ID"

# Step 8: Check task queue
echo -e "\n${GREEN}Step 8: Checking task queue...${NC}"
QUEUE_STATUS=$(curl -s $API_URL/api/v1/tasks/queue)
echo -e "Task queue status:"
echo -e "  - Pending: $(echo $QUEUE_STATUS | jq -r '.pending_tasks')"
echo -e "  - Running: $(echo $QUEUE_STATUS | jq -r '.running_tasks')"
echo -e "  - Completed: $(echo $QUEUE_STATUS | jq -r '.completed_tasks')"

# Step 9: Show cost analytics
echo -e "\n${GREEN}Step 9: Cost analytics summary...${NC}"
ANALYTICS=$(curl -s $API_URL/api/v1/cost-analytics?period=7d)
echo -e "Weekly cost summary:"
echo -e "  - Total cost: ₹$(echo $ANALYTICS | jq -r '.total_cost')"
echo -e "  - Total savings: ₹$(echo $ANALYTICS | jq -r '.total_savings')"
echo -e "  - Savings percent: $(echo $ANALYTICS | jq -r '.savings_percent')%"

# Step 10: WebSocket demo (if available)
echo -e "\n${GREEN}Step 10: Real-time updates via WebSocket...${NC}"
echo -e "${YELLOW}To see real-time updates, open the Phoenix Dashboard at:${NC}"
echo -e "${BLUE}http://localhost:3000${NC}"
echo -e "\nOr connect to WebSocket at: ws://localhost:8081/api/v1/ws"

echo -e "\n${BLUE}=== Demo Complete ===${NC}"
echo -e "\n${GREEN}Key UI Features Demonstrated:${NC}"
echo -e "✓ Fleet status visualization"
echo -e "✓ Real-time cost flow monitoring"
echo -e "✓ Experiment wizard (no YAML)"
echo -e "✓ Pipeline impact preview"
echo -e "✓ One-click deployment"
echo -e "✓ Task queue visibility"
echo -e "✓ Executive analytics"

echo -e "\n${YELLOW}Next steps:${NC}"
echo -e "1. Open the dashboard: ${BLUE}http://localhost:3000${NC}"
echo -e "2. Try the visual pipeline builder"
echo -e "3. Explore the cardinality explorer"
echo -e "4. Test instant rollback with time machine"
echo -e "\n${GREEN}Happy optimizing! 🚀${NC}"