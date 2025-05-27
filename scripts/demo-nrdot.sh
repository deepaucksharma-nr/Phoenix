#!/bin/bash
# Demo script for NRDOT integration

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Phoenix Platform - NRDOT Integration Demo${NC}"
echo "=========================================="
echo

# Check for New Relic license key
if [ -z "$NEW_RELIC_LICENSE_KEY" ]; then
    echo -e "${RED}Error: NEW_RELIC_LICENSE_KEY environment variable not set${NC}"
    echo "Please set your New Relic license key:"
    echo "  export NEW_RELIC_LICENSE_KEY=your-license-key"
    exit 1
fi

# Set NRDOT endpoint (with default)
NEW_RELIC_OTLP_ENDPOINT=${NEW_RELIC_OTLP_ENDPOINT:-"otlp.nr-data.net:4317"}

echo -e "${YELLOW}Configuration:${NC}"
echo "  License Key: ****${NEW_RELIC_LICENSE_KEY: -4}"
echo "  OTLP Endpoint: $NEW_RELIC_OTLP_ENDPOINT"
echo

# Start services
echo -e "${YELLOW}Starting services...${NC}"
docker-compose up -d postgres redis prometheus pushgateway

# Wait for services
echo "Waiting for services to be ready..."
sleep 10

# Start API
echo -e "${YELLOW}Starting Phoenix API...${NC}"
cd projects/phoenix-api
make run &
API_PID=$!
cd ../..

# Wait for API
echo "Waiting for API to be ready..."
sleep 5

# Create NRDOT experiment
echo -e "${YELLOW}Creating NRDOT cardinality reduction experiment...${NC}"
phoenix-cli experiment create \
    --name "NRDOT Cardinality Demo" \
    --description "Demonstrate 70% cardinality reduction with NRDOT" \
    --baseline-pipeline "baseline" \
    --candidate-pipeline "nrdot-cardinality" \
    --use-nrdot \
    --nr-license-key "$NEW_RELIC_LICENSE_KEY" \
    --nr-otlp-endpoint "$NEW_RELIC_OTLP_ENDPOINT" \
    --max-cardinality 5000 \
    --reduction-percentage 70 \
    --duration 5m

# Get experiment ID
EXPERIMENT_ID=$(phoenix-cli experiment list --format json | jq -r '.[0].id')
echo "Experiment created: $EXPERIMENT_ID"

# Start agents with NRDOT
echo -e "${YELLOW}Starting agents with NRDOT collectors...${NC}"

# Start baseline agent
PHOENIX_HOST_ID=demo-agent-baseline \
USE_NRDOT=false \
phoenix-agent \
    --api-url http://localhost:8080 \
    --poll-interval 10s &
BASELINE_AGENT_PID=$!

# Start NRDOT agent
PHOENIX_HOST_ID=demo-agent-nrdot \
USE_NRDOT=true \
NEW_RELIC_LICENSE_KEY="$NEW_RELIC_LICENSE_KEY" \
NEW_RELIC_OTLP_ENDPOINT="$NEW_RELIC_OTLP_ENDPOINT" \
phoenix-agent \
    --api-url http://localhost:8080 \
    --poll-interval 10s &
NRDOT_AGENT_PID=$!

# Wait for agents
sleep 5

# Start experiment
echo -e "${YELLOW}Starting experiment...${NC}"
phoenix-cli experiment start --id "$EXPERIMENT_ID"

# Monitor experiment
echo -e "${YELLOW}Monitoring experiment progress...${NC}"
echo "You can view metrics in New Relic UI at: https://one.newrelic.com"
echo

# Show real-time status
for i in {1..30}; do
    clear
    echo -e "${GREEN}NRDOT Integration Demo - Live Status${NC}"
    echo "====================================="
    
    # Get experiment status
    phoenix-cli experiment status --id "$EXPERIMENT_ID"
    
    # Get metrics
    echo -e "\n${YELLOW}Metrics Summary:${NC}"
    phoenix-cli experiment metrics --id "$EXPERIMENT_ID"
    
    echo -e "\n${YELLOW}Agent Status:${NC}"
    curl -s http://localhost:8080/api/v1/fleet/status | jq -r '.agents[] | "\(.host_id): \(.status)"'
    
    sleep 10
done

# Stop experiment
echo -e "${YELLOW}Stopping experiment...${NC}"
phoenix-cli experiment stop --id "$EXPERIMENT_ID"

# Show final results
echo -e "${GREEN}Final Results:${NC}"
phoenix-cli experiment status --id "$EXPERIMENT_ID" --verbose

# Cleanup
echo -e "${YELLOW}Cleaning up...${NC}"
kill $API_PID $BASELINE_AGENT_PID $NRDOT_AGENT_PID 2>/dev/null || true

echo -e "${GREEN}Demo completed!${NC}"
echo
echo "Key takeaways:"
echo "1. NRDOT collector successfully integrated with Phoenix"
echo "2. Cardinality reduction achieved while maintaining visibility"
echo "3. Metrics exported to New Relic for analysis"
echo "4. A/B testing allowed safe validation of configuration"