#!/bin/bash
# Start Phoenix Platform

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Phoenix Platform${NC}"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose is not installed. Please install Docker Compose first.${NC}"
    exit 1
fi

# Check if postgres is already running
if lsof -Pi :5432 -sTCP:LISTEN -t >/dev/null; then
    echo -e "${YELLOW}Port 5432 is already in use. Is PostgreSQL already running?${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build images
echo -e "${YELLOW}Building Phoenix API and Agent images...${NC}"
cd "$PROJECT_ROOT"

# Build Phoenix API
echo "Building Phoenix API..."
docker build -t phoenix/api:latest -f projects/phoenix-api/Dockerfile projects/phoenix-api/

# Build Phoenix Agent
echo "Building Phoenix Agent..."
docker build -t phoenix/agent:latest -f projects/phoenix-agent/Dockerfile projects/phoenix-agent/

# Start services
echo -e "${YELLOW}Starting services...${NC}"
docker-compose up -d

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 5

# Check PostgreSQL
echo -n "Checking PostgreSQL... "
if docker-compose exec -T postgres pg_isready -U phoenix > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "PostgreSQL is not ready. Check logs with: docker-compose logs postgres"
fi

# Run migrations
echo -e "${YELLOW}Running database migrations...${NC}"
docker-compose exec -T phoenix-api /app/phoenix-api migrate || true

# Check Phoenix API
echo -n "Checking Phoenix API... "
if curl -f -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "Phoenix API is not ready. Check logs with: docker-compose logs phoenix-api"
fi

# Check Prometheus
echo -n "Checking Prometheus... "
if curl -f -s http://localhost:9090/-/healthy > /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

# Check Pushgateway
echo -n "Checking Pushgateway... "
if curl -f -s http://localhost:9091/metrics > /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

# Show status
echo -e "\n${GREEN}Phoenix Platform is running!${NC}"
echo -e "\nServices:"
echo -e "  - Phoenix API:    http://localhost:8080"
echo -e "  - WebSocket Hub:  ws://localhost:8081"
echo -e "  - Prometheus:     http://localhost:9090"
echo -e "  - Pushgateway:    http://localhost:9091"
echo -e "  - Grafana:        http://localhost:3001 (admin/admin)"
echo -e "  - PostgreSQL:     localhost:5432 (phoenix/phoenix)"
echo -e "  - Redis:          localhost:6379"

echo -e "\n${YELLOW}Want the full UI experience?${NC}"
echo -e "Run: ${GREEN}./scripts/start-phoenix-ui.sh${NC}"

echo -e "\nUseful commands:"
echo -e "  - View logs:          docker-compose logs -f"
echo -e "  - Stop services:      docker-compose down"
echo -e "  - Run tests:          ./scripts/test-integration.sh"
echo -e "  - Create experiment:  ./scripts/demo-flow.sh"

echo -e "\n${YELLOW}Quick start with API:${NC}"
echo 'curl -X POST http://localhost:8080/api/v1/experiments/wizard \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"Quick test\", \"host_selector\": [\"all\"], \"pipeline_type\": \"top-k-20\", \"duration_hours\": 1}"'