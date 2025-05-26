#!/bin/bash
# experiment-simulation.sh - Simulate Phoenix Platform experiment workflow

set -euo pipefail

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

API_URL="http://localhost:8080/api/v1"

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}    Phoenix Platform - Experiment Simulation Demo${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"

# Function to display step
step() {
    echo -e "\n${CYAN}▶ $1${NC}"
    sleep 1
}

# Function to run command and display result
run_cmd() {
    echo -e "${YELLOW}$ $1${NC}"
    eval "$1"
    echo
    sleep 2
}

# 1. Check current experiments
step "Step 1: Viewing current optimization experiments"
run_cmd "curl -s $API_URL/experiments | jq '.experiments[] | {id, name, status, savings: .cost_saving_percent}'"

# 2. Get detailed metrics
step "Step 2: Checking platform-wide cost optimization metrics"
run_cmd "curl -s $API_URL/metrics | jq ."

# 3. Simulate experiment details
step "Step 3: Getting details of a specific experiment"
run_cmd "curl -s $API_URL/experiments/exp-001 | jq ."

# 4. Simulate real-time monitoring
step "Step 4: Monitoring real-time optimization performance"
echo -e "${GREEN}Simulating real-time metrics updates...${NC}"
for i in {1..3}; do
    savings=$((45 + i * 5))
    processed=$((1234567 + i * 100000))
    echo -e "  [$(date +%H:%M:%S)] Metrics processed: $processed | Cost reduction: $savings%"
    sleep 1
done

# 5. Cost projection
step "Step 5: Calculating projected annual savings"
monthly_savings=$(curl -s $API_URL/metrics | jq -r '.monthly_savings_usd')
annual_savings=$((monthly_savings * 12))
echo -e "${GREEN}Monthly Savings: \$$monthly_savings${NC}"
echo -e "${GREEN}Projected Annual Savings: \$$annual_savings${NC}"

# 6. Optimization recommendations
step "Step 6: AI-powered optimization recommendations"
echo -e "${BLUE}Based on current metrics, Phoenix recommends:${NC}"
echo "  • Increase cardinality reduction threshold to 90%"
echo "  • Enable adaptive sampling for high-volume metrics"
echo "  • Implement tag consolidation for Kubernetes labels"
echo "  • Activate intelligent metric aggregation"

# Summary
echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Simulation Complete!${NC}"
echo -e "\nPhoenix Platform is actively optimizing your observability costs."
echo -e "Current efficiency: ${GREEN}87% cardinality reduction${NC}"
echo -e "Projected annual savings: ${GREEN}\$$annual_savings${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"