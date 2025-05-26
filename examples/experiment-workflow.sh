#!/bin/bash
# Phoenix Platform - Experiment Workflow Example
# This script demonstrates a complete experiment lifecycle

set -e

API_URL="${API_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080/ws}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Phoenix Platform - Experiment Workflow Demo${NC}"
echo "============================================"

# Function to create an experiment
create_experiment() {
    local name=$1
    local description=$2
    local baseline=$3
    local candidate=$4
    
    echo -e "\n${GREEN}Creating experiment: $name${NC}"
    
    response=$(curl -s -X POST "$API_URL/api/v1/experiments" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "'"$name"'",
            "description": "'"$description"'",
            "baseline_pipeline": "'"$baseline"'",
            "candidate_pipeline": "'"$candidate"'",
            "target_nodes": {
                "prometheus": "prometheus-0",
                "collector": "otel-collector-0"
            }
        }')
    
    experiment_id=$(echo "$response" | jq -r '.id')
    echo "Created experiment with ID: $experiment_id"
    echo "$experiment_id"
}

# Function to start an experiment
start_experiment() {
    local experiment_id=$1
    
    echo -e "\n${GREEN}Starting experiment: $experiment_id${NC}"
    
    curl -s -X PUT "$API_URL/api/v1/experiments/$experiment_id/status" \
        -H "Content-Type: application/json" \
        -d '{"status": "running"}' > /dev/null
    
    echo "Experiment started successfully"
}

# Function to simulate metrics
simulate_metrics() {
    local experiment_id=$1
    local duration=$2
    
    echo -e "\n${YELLOW}Simulating metrics for $duration seconds...${NC}"
    
    for i in $(seq 1 $duration); do
        # Simulate baseline metrics
        baseline_cpu=$((50 + RANDOM % 20))
        baseline_memory=$((60 + RANDOM % 15))
        baseline_cardinality=$((10000 + RANDOM % 2000))
        
        # Simulate candidate metrics (improved)
        candidate_cpu=$((40 + RANDOM % 15))
        candidate_memory=$((50 + RANDOM % 10))
        candidate_cardinality=$((3000 + RANDOM % 1000))
        
        echo -ne "\rProgress: $i/$duration seconds - "
        echo -ne "CPU: ${baseline_cpu}% → ${candidate_cpu}% | "
        echo -ne "Memory: ${baseline_memory}% → ${candidate_memory}% | "
        echo -ne "Cardinality: ${baseline_cardinality} → ${candidate_cardinality}"
        
        sleep 1
    done
    echo ""
}

# Function to analyze results
analyze_experiment() {
    local experiment_id=$1
    
    echo -e "\n${GREEN}Analyzing experiment results...${NC}"
    
    # In a real scenario, this would fetch actual metrics
    cost_reduction=$((60 + RANDOM % 20))
    cardinality_reduction=$((70 + RANDOM % 15))
    
    echo "Analysis complete:"
    echo "- Cost Reduction: ${cost_reduction}%"
    echo "- Cardinality Reduction: ${cardinality_reduction}%"
    echo "- Performance Impact: < 1%"
    echo "- Recommendation: PROMOTE"
}

# Function to complete an experiment
complete_experiment() {
    local experiment_id=$1
    
    echo -e "\n${GREEN}Completing experiment: $experiment_id${NC}"
    
    curl -s -X PUT "$API_URL/api/v1/experiments/$experiment_id/status" \
        -H "Content-Type: application/json" \
        -d '{"status": "completed"}' > /dev/null
    
    echo "Experiment completed successfully"
}

# Function to list all experiments
list_experiments() {
    echo -e "\n${BLUE}Listing all experiments:${NC}"
    
    response=$(curl -s "$API_URL/api/v1/experiments")
    echo "$response" | jq -r '.[] | "- \(.name) [\(.status)] - Cost Saving: \(.cost_saving_percent // 0)%"'
}

# Main workflow
main() {
    echo -e "\n${BLUE}Step 1: Create a new experiment${NC}"
    experiment_id=$(create_experiment \
        "prometheus-optimization-$(date +%s)" \
        "Optimize Prometheus metric collection using intelligent sampling" \
        "prometheus-baseline" \
        "prometheus-optimized")
    
    echo -e "\n${BLUE}Step 2: Start the experiment${NC}"
    start_experiment "$experiment_id"
    
    echo -e "\n${BLUE}Step 3: Monitor metrics (simulated)${NC}"
    simulate_metrics "$experiment_id" 10
    
    echo -e "\n${BLUE}Step 4: Analyze results${NC}"
    analyze_experiment "$experiment_id"
    
    echo -e "\n${BLUE}Step 5: Complete the experiment${NC}"
    complete_experiment "$experiment_id"
    
    echo -e "\n${BLUE}Step 6: View all experiments${NC}"
    list_experiments
    
    echo -e "\n${GREEN}✓ Experiment workflow completed successfully!${NC}"
}

# WebSocket monitoring example
monitor_websocket() {
    echo -e "\n${BLUE}Bonus: WebSocket Real-time Monitoring${NC}"
    echo "To monitor experiments in real-time, use:"
    echo ""
    echo "  # Using websocat (install with: brew install websocat)"
    echo "  websocat $WS_URL"
    echo ""
    echo "  # Then subscribe to an experiment:"
    echo '  {"type":"subscribe","data":{"topic":"experiment:EXPERIMENT_ID"}}'
    echo ""
    echo "  # Or monitor all metrics:"
    echo '  {"type":"subscribe","data":{"topic":"metrics:EXPERIMENT_ID"}}'
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed. Install it with: brew install jq${NC}"
    exit 1
fi

# Check if API is reachable
if ! curl -s "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}Error: Platform API is not reachable at $API_URL${NC}"
    echo "Please ensure the platform-api service is running."
    exit 1
fi

# Run main workflow
main

# Show WebSocket monitoring info
monitor_websocket

echo -e "\n${BLUE}Demo complete! 🚀${NC}"