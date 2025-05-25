#!/bin/bash
# End-to-End Test Script for Phoenix Platform

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform E2E Test ===${NC}"

# Configuration
API_URL="http://localhost:8080"
MAX_RETRIES=30
RETRY_DELAY=2

# Function to wait for service
wait_for_service() {
    local service_name=$1
    local health_url=$2
    local retries=0
    
    echo -n "Waiting for $service_name..."
    while [ $retries -lt $MAX_RETRIES ]; do
        if curl -s -f "$health_url" > /dev/null 2>&1; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        echo -n "."
        sleep $RETRY_DELAY
        ((retries++))
    done
    echo -e " ${RED}✗${NC}"
    return 1
}

# Function to create experiment
create_experiment() {
    echo -e "\n${YELLOW}Creating test experiment...${NC}"
    
    local response=$(curl -s -X POST "$API_URL/api/v1/experiments" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "e2e-test-experiment",
            "description": "End-to-end test experiment",
            "baseline_pipeline": "baseline-v1",
            "candidate_pipeline": "optimized-v1",
            "target_selector": {
                "app": "test-app"
            },
            "duration": "1h",
            "traffic_split": {
                "baseline": 50,
                "candidate": 50
            }
        }' 2>/dev/null || echo '{"error": "Failed to create experiment"}')
    
    echo "Response: $response"
    
    # Extract experiment ID
    local exp_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    if [ -z "$exp_id" ]; then
        echo -e "${RED}Failed to create experiment${NC}"
        return 1
    fi
    
    echo -e "${GREEN}Created experiment: $exp_id${NC}"
    echo "$exp_id"
}

# Function to check experiment status
check_experiment_status() {
    local exp_id=$1
    echo -e "\n${YELLOW}Checking experiment status...${NC}"
    
    local status=$(curl -s "$API_URL/api/v1/experiments/$exp_id" | \
        grep -o '"status":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$status" ]; then
        echo -e "${RED}Failed to get experiment status${NC}"
        return 1
    fi
    
    echo -e "Experiment status: ${GREEN}$status${NC}"
}

# Main test flow
main() {
    echo -e "\n${BLUE}1. Starting services...${NC}"
    docker-compose -f docker-compose.e2e.yml up -d
    
    echo -e "\n${BLUE}2. Waiting for services to be ready...${NC}"
    wait_for_service "PostgreSQL" "localhost:5432" || exit 1
    wait_for_service "API Service" "$API_URL/health" || exit 1
    wait_for_service "Controller" "http://localhost:8082/health" || exit 1
    wait_for_service "Generator" "http://localhost:8083/health" || exit 1
    
    echo -e "\n${BLUE}3. Running database migrations...${NC}"
    # This would run migrations, for now we'll skip
    echo -e "${YELLOW}Skipping migrations (implement later)${NC}"
    
    echo -e "\n${BLUE}4. Creating test experiment...${NC}"
    EXP_ID=$(create_experiment)
    if [ $? -ne 0 ]; then
        echo -e "${RED}Test failed!${NC}"
        docker-compose -f docker-compose.e2e.yml logs
        exit 1
    fi
    
    echo -e "\n${BLUE}5. Waiting for controller to process...${NC}"
    sleep 5
    
    echo -e "\n${BLUE}6. Checking experiment status...${NC}"
    check_experiment_status "$EXP_ID"
    
    echo -e "\n${BLUE}7. Verifying generated configs...${NC}"
    # Check if configs were generated
    echo -e "${YELLOW}Config verification not implemented yet${NC}"
    
    echo -e "\n${GREEN}✅ E2E Test Completed Successfully!${NC}"
    
    # Cleanup
    echo -e "\n${BLUE}8. Cleaning up...${NC}"
    read -p "Stop services? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f docker-compose.e2e.yml down
    fi
}

# Handle errors
trap 'echo -e "\n${RED}Test failed! Check logs:${NC}"; docker-compose -f docker-compose.e2e.yml logs' ERR

# Run main
main "$@"