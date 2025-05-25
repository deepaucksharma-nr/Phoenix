#!/bin/bash
# Phoenix Platform E2E Demo Runner

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

cd "$(dirname "$0")/.."

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Phoenix Platform E2E Demo          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"

# Function to cleanup
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    pkill -f "go run ./cmd" || true
    docker-compose -f docker-compose.e2e.yml down 2>/dev/null || true
}

# Set trap for cleanup
trap cleanup EXIT

# Option 1: Local Go execution (faster for development)
run_local() {
    echo -e "\n${BLUE}Starting services locally...${NC}"
    
    # Build services
    echo "Building services..."
    (cd services/api && go build -o bin/api ./cmd/api) &
    (cd services/controller && go build -o bin/controller ./cmd/controller) &
    (cd services/generator && go build -o bin/generator ./cmd/generator) &
    wait
    
    # Start services
    echo -e "\n${GREEN}Starting API service...${NC}"
    (cd services/api && ./bin/api) &
    API_PID=$!
    
    echo -e "${GREEN}Starting Controller service...${NC}"
    (cd services/controller && ./bin/controller) &
    CTRL_PID=$!
    
    echo -e "${GREEN}Starting Generator service...${NC}"
    (cd services/generator && ./bin/generator) &
    GEN_PID=$!
    
    sleep 3
}

# Option 2: Docker Compose execution
run_docker() {
    echo -e "\n${BLUE}Starting services with Docker...${NC}"
    docker-compose -f docker-compose.e2e.yml up -d --build
    
    echo "Waiting for services to start..."
    sleep 15
}

# Test the services
test_services() {
    echo -e "\n${BLUE}Testing Phoenix Platform...${NC}"
    
    # Test health endpoints
    echo -e "\n${YELLOW}1. Health Check${NC}"
    echo -n "API Health: "
    curl -s http://localhost:8080/health | jq -r '.status' || echo "Failed"
    echo -n "Controller Health: "
    curl -s http://localhost:8082/health | jq -r '.status' || echo "Failed"
    echo -n "Generator Health: "
    curl -s http://localhost:8083/health | jq -r '.status' || echo "Failed"
    
    # Create an experiment
    echo -e "\n${YELLOW}2. Create Experiment${NC}"
    RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
        -H "Content-Type: application/json" \
        -d '{
            "name": "demo-experiment",
            "baseline_pipeline": "baseline-v1",
            "candidate_pipeline": "optimized-v1",
            "target_selector": {"app": "demo-app"},
            "duration": "30m"
        }')
    
    echo "Response: "
    echo "$RESPONSE" | jq . || echo "$RESPONSE"
    
    # Extract experiment ID
    EXP_ID=$(echo "$RESPONSE" | jq -r '.id' 2>/dev/null || echo "exp-demo")
    echo -e "${GREEN}Created experiment: $EXP_ID${NC}"
    
    # Check experiment status
    echo -e "\n${YELLOW}3. Check Experiment Status${NC}"
    curl -s http://localhost:8080/api/v1/experiments/$EXP_ID | jq . || echo "Failed"
    
    # List pipelines
    echo -e "\n${YELLOW}4. List Available Pipelines${NC}"
    curl -s http://localhost:8080/api/v1/pipelines | jq . || echo "Failed"
    
    # Generate config
    echo -e "\n${YELLOW}5. Generate Pipeline Config${NC}"
    curl -s -X POST http://localhost:8083/generate \
        -H "Content-Type: application/json" \
        -d '{"experiment_id": "'$EXP_ID'", "type": "baseline"}' | jq . || echo "Failed"
    
    # Show metrics endpoint
    echo -e "\n${YELLOW}6. Prometheus Metrics${NC}"
    echo "API Metrics: http://localhost:8080/metrics"
    echo "Controller Metrics: http://localhost:8082/metrics"
    echo "Generator Metrics: http://localhost:8083/metrics"
    
    echo -e "\n${GREEN}✅ E2E Demo Complete!${NC}"
    echo -e "\nThe Phoenix Platform is running. You can:"
    echo "- Access the API at http://localhost:8080"
    echo "- View logs with: docker-compose -f docker-compose.e2e.yml logs -f"
    echo "- Stop services with: docker-compose -f docker-compose.e2e.yml down"
}

# Main execution
main() {
    echo -e "\n${YELLOW}Choose execution mode:${NC}"
    echo "1) Local Go execution (faster)"
    echo "2) Docker Compose (isolated)"
    echo -n "Enter choice [1-2]: "
    read -r choice
    
    case $choice in
        1)
            run_local
            test_services
            echo -e "\n${YELLOW}Press Ctrl+C to stop services${NC}"
            wait
            ;;
        2)
            run_docker
            test_services
            echo -e "\n${YELLOW}Services are running in background${NC}"
            echo "Stop with: docker-compose -f docker-compose.e2e.yml down"
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            exit 1
            ;;
    esac
}

# Run main
main