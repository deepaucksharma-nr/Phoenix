#!/bin/bash
# End-to-end test script for Phoenix Platform

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_URL="http://localhost:8080"
WS_URL="ws://localhost:8081"
DASHBOARD_URL="http://localhost:3000"

echo -e "${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        Phoenix Platform E2E Test Suite                ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name=$1
    local test_command=$2
    
    echo -ne "\n${YELLOW}Testing:${NC} $test_name... "
    
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Wait for services
echo -e "\n${YELLOW}Waiting for services to be ready...${NC}"
sleep 5

# 1. Test API Health
run_test "API Health Check" "curl -f -s $API_URL/health"

# 2. Test WebSocket Endpoint
run_test "WebSocket Endpoint" "curl -f -s -I $API_URL/api/v1/ws | grep -E '(400|426)'"

# 3. Test Dashboard
run_test "Dashboard UI" "curl -f -s $DASHBOARD_URL | grep -q 'Phoenix'"

# 4. Test Fleet Status
run_test "Fleet Status API" "curl -f -s $API_URL/api/v1/fleet/status | jq '.total_agents' > /dev/null"

# 5. Test Pipeline Templates
run_test "Pipeline Templates" "curl -f -s $API_URL/api/v1/pipelines/templates | jq '.[0].id' | grep -q 'top-k-20'"

# 6. Test Metric Cost Flow
run_test "Metric Cost Flow" "curl -f -s $API_URL/api/v1/metrics/cost-flow | jq '.total_cost_rate' > /dev/null"

# 7. Create Test Experiment
echo -e "\n${YELLOW}Creating test experiment...${NC}"
EXPERIMENT_JSON='{
  "name": "E2E Test Experiment",
  "description": "Automated test",
  "host_selector": ["group:demo"],
  "pipeline_type": "top-k-20",
  "duration_hours": 1
}'

if EXPERIMENT_RESPONSE=$(curl -s -X POST $API_URL/api/v1/experiments/wizard \
  -H "Content-Type: application/json" \
  -d "$EXPERIMENT_JSON"); then
  
  EXPERIMENT_ID=$(echo $EXPERIMENT_RESPONSE | jq -r '.id')
  if [ "$EXPERIMENT_ID" != "null" ] && [ -n "$EXPERIMENT_ID" ]; then
    echo -e "${GREEN}✓ Created experiment: $EXPERIMENT_ID${NC}"
    ((TESTS_PASSED++))
  else
    echo -e "${RED}✗ Failed to create experiment${NC}"
    ((TESTS_FAILED++))
  fi
else
  echo -e "${RED}✗ Failed to create experiment${NC}"
  ((TESTS_FAILED++))
fi

# 8. Test Pipeline Preview
PREVIEW_JSON='{
  "pipeline_config": {
    "processors": [{"type": "top_k", "config": {"k": 20}}]
  },
  "target_hosts": ["demo-host"]
}'

run_test "Pipeline Preview" "curl -f -s -X POST $API_URL/api/v1/pipelines/preview -H 'Content-Type: application/json' -d '$PREVIEW_JSON' | jq '.estimated_cost_reduction' > /dev/null"

# 9. Test Task Queue
run_test "Task Queue Status" "curl -f -s $API_URL/api/v1/tasks/queue | jq '.pending_tasks' > /dev/null"

# 10. Test Cost Analytics
run_test "Cost Analytics" "curl -f -s '$API_URL/api/v1/cost-analytics?period=1d' | jq '.total_cost' > /dev/null"

# 11. Test Agent Endpoints (with proper header)
run_test "Agent Task Polling" "curl -f -s -H 'X-Agent-Host-ID: test-agent' $API_URL/api/v1/agent/tasks"

# 12. Test Database Connection
echo -ne "\n${YELLOW}Testing:${NC} Database Connection... "
if docker-compose exec -T postgres pg_isready -U phoenix > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAILED${NC}"
    ((TESTS_FAILED++))
fi

# 13. Test Redis Connection
echo -ne "${YELLOW}Testing:${NC} Redis Connection... "
if docker-compose exec -T redis redis-cli ping | grep -q PONG > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAILED${NC}"
    ((TESTS_FAILED++))
fi

# 14. Test Prometheus
run_test "Prometheus Health" "curl -f -s http://localhost:9090/-/healthy"

# 15. Test Pushgateway
run_test "Pushgateway Health" "curl -f -s http://localhost:9091/metrics | grep -q pushgateway"

# Results Summary
echo -e "\n${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                  Test Results                         ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"
echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
echo -e "${RED}Failed:${NC} $TESTS_FAILED"
TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo -e "${YELLOW}Total:${NC}  $TOTAL_TESTS"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed! Phoenix Platform is working correctly.${NC}"
    echo -e "\n${YELLOW}Next steps:${NC}"
    echo -e "1. Open dashboard: ${BLUE}http://localhost:3000${NC}"
    echo -e "2. Try the demo: ${GREEN}./scripts/demo-ui-flow.sh${NC}"
    echo -e "3. Create experiments via UI or CLI"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please check the logs:${NC}"
    echo -e "  docker-compose logs -f"
    exit 1
fi