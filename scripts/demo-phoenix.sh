#!/bin/bash
# demo-phoenix.sh - Demonstrate Phoenix Platform capabilities

set -euo pipefail

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          🔥 Phoenix Platform Demonstration 🔥              ║${NC}"
echo -e "${BLUE}║     Reduce Observability Costs by 90% with AI/ML          ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}\n"

# Check if services are running
echo -e "${YELLOW}1. Checking Infrastructure Services...${NC}"
services=(
    "phoenix-postgres:5432:PostgreSQL"
    "phoenix-redis:6379:Redis"
    "phoenix-nats:4222:NATS"
    "phoenix-jaeger:16686:Jaeger"
)

for service in "${services[@]}"; do
    IFS=':' read -r container port name <<< "$service"
    if docker ps | grep -q "$container"; then
        echo -e "${GREEN}✓ $name is running on port $port${NC}"
    else
        echo -e "${RED}✗ $name is not running${NC}"
    fi
done

# Check hello-phoenix service
echo -e "\n${YELLOW}2. Phoenix API Service Status...${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    health=$(curl -s http://localhost:8080/health | jq -r '.status')
    uptime=$(curl -s http://localhost:8080/health | jq -r '.uptime')
    echo -e "${GREEN}✓ Phoenix API is $health (uptime: $uptime)${NC}"
else
    echo -e "${RED}✗ Phoenix API is not responding${NC}"
    echo -e "${YELLOW}Starting Phoenix API...${NC}"
    cd projects/hello-phoenix && ./hello-phoenix &
    sleep 2
fi

# Display service info
echo -e "\n${YELLOW}3. Service Information...${NC}"
curl -s http://localhost:8080/info | jq '{
    service: .service,
    version: .version,
    description: .description
}'

# Show active experiments
echo -e "\n${YELLOW}4. Active Cost Optimization Experiments...${NC}"
experiments=$(curl -s http://localhost:8080/api/v1/experiments)
echo "$experiments" | jq '.experiments[] | {
    id: .id,
    name: .name,
    status: .status,
    savings: (.cost_saving_percent | tostring + "%")
}'

# Show cost savings metrics
echo -e "\n${YELLOW}5. Cost Optimization Metrics...${NC}"
metrics=$(curl -s http://localhost:8080/api/v1/metrics)
echo "$metrics" | jq '{
    "Monthly Savings": ("$" + (.monthly_savings_usd | tostring)),
    "Average Cost Reduction": ((.average_cost_saving | tostring) + "%"),
    "Cardinality Reduction": .cardinality_reduction,
    "Metrics Processed": .metrics_processed
}'

# Architecture components
echo -e "\n${YELLOW}6. Phoenix Platform Components...${NC}"
echo -e "${GREEN}Core Services:${NC}"
echo "  • Platform API - RESTful API gateway"
echo "  • Experiment Controller - K8s operator for A/B tests"
echo "  • Pipeline Operator - Manages telemetry pipelines"
echo "  • Web Dashboard - React-based UI"
echo "  • Phoenix CLI - Command-line interface"

echo -e "\n${GREEN}Infrastructure:${NC}"
echo "  • PostgreSQL - Experiment metadata"
echo "  • Redis - Caching & real-time data"
echo "  • NATS/Kafka - Event streaming"
echo "  • Jaeger - Distributed tracing"
echo "  • Prometheus - Metrics collection"
echo "  • Grafana - Visualization"

# URLs
echo -e "\n${YELLOW}7. Access Points...${NC}"
echo -e "${BLUE}Phoenix API:${NC} http://localhost:8080/info"
echo -e "${BLUE}Jaeger UI:${NC} http://localhost:16686"
echo -e "${BLUE}Prometheus:${NC} http://localhost:9090"
echo -e "${BLUE}Grafana:${NC} http://localhost:3000 (admin/phoenix)"

# Next steps
echo -e "\n${YELLOW}8. Try These Commands...${NC}"
echo -e "${GREEN}# Get experiment details:${NC}"
echo "curl http://localhost:8080/api/v1/experiments/exp-001 | jq ."
echo -e "\n${GREEN}# Check system health:${NC}"
echo "curl http://localhost:8080/health | jq ."
echo -e "\n${GREEN}# View all metrics:${NC}"
echo "curl http://localhost:8080/api/v1/metrics | jq ."

echo -e "\n${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Phoenix Platform is reducing observability costs!      ║${NC}"
echo -e "${BLUE}║          Visit http://localhost:8080/info                  ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"