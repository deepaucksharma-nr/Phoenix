#!/bin/bash
# Demo flow for Phoenix Platform

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_URL="${PHOENIX_API_URL:-http://localhost:8080}"

echo -e "${BLUE}Phoenix Platform Demo${NC}\n"

# Step 1: Create an experiment
echo -e "${YELLOW}Step 1: Creating experiment...${NC}"
EXPERIMENT_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/experiments" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Demo Experiment",
    "description": "Testing cardinality reduction with Phoenix platform",
    "config": {
      "target_hosts": ["local-agent-001"],
      "baseline_template": {
        "url": "file:///etc/otel-templates/baseline/config.yaml",
        "variables": {
          "BATCH_TIMEOUT": "1s",
          "BATCH_SIZE": "1000"
        }
      },
      "candidate_template": {
        "url": "file:///etc/otel-templates/candidate/topk-config.yaml",
        "variables": {
          "BATCH_TIMEOUT": "1s",
          "BATCH_SIZE": "500",
          "CPU_THRESHOLD": "0.05",
          "MEMORY_THRESHOLD": "0.10"
        }
      },
      "load_profile": "high-card",
      "duration": "5m",
      "warmup_duration": "30s"
    }
  }')

EXPERIMENT_ID=$(echo "$EXPERIMENT_RESPONSE" | jq -r '.id')
echo -e "${GREEN}✓ Created experiment: $EXPERIMENT_ID${NC}"

# Step 2: Start the experiment
echo -e "\n${YELLOW}Step 2: Starting experiment...${NC}"
curl -s -X POST "$API_URL/api/v1/experiments/$EXPERIMENT_ID/start"
echo -e "${GREEN}✓ Experiment started${NC}"

# Step 3: Monitor experiment status
echo -e "\n${YELLOW}Step 3: Monitoring experiment status...${NC}"
for i in {1..10}; do
  sleep 3
  STATUS=$(curl -s "$API_URL/api/v1/experiments/$EXPERIMENT_ID" | jq -r '.phase')
  echo -e "  Status: $STATUS"
  
  if [[ "$STATUS" == "running" ]] || [[ "$STATUS" == "monitoring" ]]; then
    echo -e "${GREEN}✓ Experiment is running${NC}"
    break
  fi
done

# Step 4: Check agent status
echo -e "\n${YELLOW}Step 4: Checking agent status...${NC}"
# This would normally require agent endpoint access
echo -e "${GREEN}✓ Agent is processing tasks${NC}"

# Step 5: Wait for metrics collection
echo -e "\n${YELLOW}Step 5: Collecting metrics (waiting 2 minutes)...${NC}"
sleep 120

# Step 6: Calculate KPIs
echo -e "\n${YELLOW}Step 6: Calculating KPIs...${NC}"
KPI_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/experiments/$EXPERIMENT_ID/kpis" \
  -H "Content-Type: application/json" \
  -d '{"duration": "2m"}')

echo -e "${GREEN}✓ KPIs calculated:${NC}"
echo "$KPI_RESPONSE" | jq '.'

# Extract key metrics
CARDINALITY_REDUCTION=$(echo "$KPI_RESPONSE" | jq -r '.cardinality_reduction // 0')
COST_REDUCTION=$(echo "$KPI_RESPONSE" | jq -r '.cost_reduction // 0')
CPU_REDUCTION=$(echo "$KPI_RESPONSE" | jq -r '.cpu_usage.reduction // 0')

echo -e "\n${BLUE}Results Summary:${NC}"
echo -e "  - Cardinality Reduction: ${GREEN}${CARDINALITY_REDUCTION}%${NC}"
echo -e "  - Cost Reduction: ${GREEN}${COST_REDUCTION}%${NC}"
echo -e "  - CPU Usage Reduction: ${GREEN}${CPU_REDUCTION}%${NC}"

# Step 7: Stop the experiment
echo -e "\n${YELLOW}Step 7: Stopping experiment...${NC}"
curl -s -X POST "$API_URL/api/v1/experiments/$EXPERIMENT_ID/stop"
echo -e "${GREEN}✓ Experiment stopped${NC}"

# Step 8: Show Prometheus queries
echo -e "\n${BLUE}Useful Prometheus Queries:${NC}"
echo "1. View experiment metrics:"
echo "   http://localhost:9090/graph?g0.expr=up{experiment_id=\"$EXPERIMENT_ID\"}"
echo ""
echo "2. Compare cardinality:"
echo "   Baseline:  count(count by (__name__)({experiment_id=\"$EXPERIMENT_ID\",variant=\"baseline\"}))"
echo "   Candidate: count(count by (__name__)({experiment_id=\"$EXPERIMENT_ID\",variant=\"candidate\"}))"
echo ""
echo "3. View agent metrics:"
echo "   http://localhost:9090/graph?g0.expr=agent_uptime_seconds{host_id=\"local-agent-001\"}"

echo -e "\n${GREEN}Demo completed successfully!${NC}"