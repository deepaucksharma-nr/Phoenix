#!/bin/bash
# Phoenix MVP Validation Script
# Validates all critical flows for MVP readiness

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
API_URL="${PHOENIX_API_URL:-http://localhost:8080}"
PHOENIX_CLI="${PHOENIX_CLI:-./projects/phoenix-cli/phoenix-cli}"
EXPERIMENT_DURATION="${EXPERIMENT_DURATION:-60s}"

echo "🔍 Phoenix MVP Validation Suite"
echo "================================"
echo "API URL: $API_URL"
echo "CLI: $PHOENIX_CLI"
echo ""

# Track failures
FAILURES=0

# Helper functions
check_command() {
    local cmd=$1
    local expected=$2
    local description=$3
    
    echo -n "Testing: $description... "
    
    if output=$($cmd 2>&1); then
        if [[ -n "$expected" && ! "$output" =~ $expected ]]; then
            echo -e "${RED}FAIL${NC} - Output missing expected content: $expected"
            echo "Output: $output"
            ((FAILURES++))
        else
            echo -e "${GREEN}PASS${NC}"
        fi
    else
        echo -e "${RED}FAIL${NC} - Command failed with exit code $?"
        echo "Error: $output"
        ((FAILURES++))
    fi
}

check_api() {
    local method=$1
    local endpoint=$2
    local expected_code=$3
    local description=$4
    local data=$5
    
    echo -n "Testing API: $description... "
    
    if [[ -n "$data" ]]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [[ "$http_code" == "$expected_code" ]]; then
        echo -e "${GREEN}PASS${NC}"
        echo "$body" > /tmp/last_api_response.json
    else
        echo -e "${RED}FAIL${NC} - Expected $expected_code, got $http_code"
        echo "Response: $body"
        ((FAILURES++))
    fi
}

# Phase 1: Basic Connectivity
echo -e "\n${YELLOW}Phase 1: Basic Connectivity${NC}"
echo "=============================="

check_api "GET" "/health" "200" "API Health Check"
check_command "$PHOENIX_CLI version" "Phoenix CLI" "CLI Version"

# Phase 2: Pipeline Operations
echo -e "\n${YELLOW}Phase 2: Pipeline Operations${NC}"
echo "=============================="

# List available templates
check_command "$PHOENIX_CLI pipeline list" "baseline\|adaptive\|topk" "List Pipeline Templates"

# Deploy a baseline pipeline
DEPLOYMENT_ID=$($PHOENIX_CLI pipeline deploy baseline --output json 2>/dev/null | jq -r '.id' || echo "")
if [[ -n "$DEPLOYMENT_ID" ]]; then
    echo -e "${GREEN}✓${NC} Pipeline deployed: $DEPLOYMENT_ID"
    
    # Check deployment status
    check_command "$PHOENIX_CLI pipeline status $DEPLOYMENT_ID" "running\|active" "Pipeline Status Check"
    
    # Test rollback
    check_command "$PHOENIX_CLI pipeline rollback $DEPLOYMENT_ID" "rolled back\|stopped" "Pipeline Rollback"
else
    echo -e "${RED}✗${NC} Failed to deploy pipeline"
    ((FAILURES++))
fi

# Phase 3: Experiment Lifecycle
echo -e "\n${YELLOW}Phase 3: Experiment Lifecycle${NC}"
echo "==============================="

# Create experiment
EXPERIMENT_JSON='{
  "name": "MVP Validation Test",
  "description": "Automated validation of experiment flow",
  "baseline_template": "baseline",
  "candidate_template": "topk",
  "target_hosts": ["localhost"],
  "load_profile": "normal",
  "duration": "'$EXPERIMENT_DURATION'"
}'

check_api "POST" "/api/v1/experiments" "201" "Create Experiment" "$EXPERIMENT_JSON"

# Extract experiment ID
EXPERIMENT_ID=$(cat /tmp/last_api_response.json | jq -r '.id' 2>/dev/null || echo "")

if [[ -n "$EXPERIMENT_ID" ]]; then
    echo -e "${GREEN}✓${NC} Experiment created: $EXPERIMENT_ID"
    
    # Start experiment
    check_api "POST" "/api/v1/experiments/$EXPERIMENT_ID/start" "200" "Start Experiment"
    
    # Wait a bit for it to be running
    sleep 5
    
    # Check status
    check_api "GET" "/api/v1/experiments/$EXPERIMENT_ID" "200" "Get Experiment Status"
    
    # Check metrics endpoint
    check_command "$PHOENIX_CLI experiment metrics $EXPERIMENT_ID" "baseline\|candidate" "Get Experiment Metrics"
    
    # Stop experiment
    check_api "POST" "/api/v1/experiments/$EXPERIMENT_ID/stop" "200" "Stop Experiment"
    
    # Wait for completion
    sleep 3
    
    # Check final status
    check_api "GET" "/api/v1/experiments/$EXPERIMENT_ID" "200" "Get Final Status"
    
    # Check KPIs
    check_api "GET" "/api/v1/experiments/$EXPERIMENT_ID/kpis" "200" "Get Experiment KPIs"
else
    echo -e "${RED}✗${NC} Failed to create experiment"
    ((FAILURES++))
fi

# Phase 4: WebSocket Events
echo -e "\n${YELLOW}Phase 4: WebSocket Events${NC}"
echo "=========================="

# Test WebSocket connection (simplified check)
echo -n "Testing: WebSocket connectivity... "
if curl -s -I -N \
    -H "Connection: Upgrade" \
    -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" \
    -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
    "$API_URL/api/v1/ws" | grep -q "101"; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC} - WebSocket upgrade failed"
    ((FAILURES++))
fi

# Phase 5: Agent Operations
echo -e "\n${YELLOW}Phase 5: Agent Operations${NC}"
echo "========================="

# Check agent status
check_api "GET" "/api/v1/agents" "200" "List Agents"

# Phase 6: UI Endpoints
echo -e "\n${YELLOW}Phase 6: UI Endpoints${NC}"
echo "======================"

check_api "GET" "/api/v1/metrics/cost-flow" "200" "Metrics Cost Flow"
check_api "GET" "/api/v1/metrics/cardinality" "200" "Cardinality Breakdown"
check_api "GET" "/api/v1/pipeline-templates" "200" "Pipeline Templates"

# Phase 7: Error Handling
echo -e "\n${YELLOW}Phase 7: Error Handling${NC}"
echo "======================="

# Test invalid operations
check_api "GET" "/api/v1/experiments/invalid-id" "404" "Invalid Experiment ID"
check_api "POST" "/api/v1/experiments/invalid-id/start" "404" "Start Invalid Experiment"

# Invalid experiment creation
INVALID_JSON='{
  "name": "",
  "baseline_template": "non-existent"
}'
check_api "POST" "/api/v1/experiments" "400" "Invalid Experiment Data" "$INVALID_JSON"

# Summary
echo -e "\n${YELLOW}Validation Summary${NC}"
echo "=================="
if [[ $FAILURES -eq 0 ]]; then
    echo -e "${GREEN}✅ All tests passed! Phoenix MVP is ready.${NC}"
    exit 0
else
    echo -e "${RED}❌ $FAILURES tests failed. Please check the output above.${NC}"
    exit 1
fi