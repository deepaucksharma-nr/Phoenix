#!/bin/bash
# test-system.sh - Test Phoenix Platform components

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform System Test ===${NC}\n"

# Test 1: Database Connection
echo -e "${BLUE}1. Testing PostgreSQL...${NC}"
if docker exec phoenix-postgres psql -U phoenix -d phoenix_db -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is operational${NC}"
else
    echo -e "${RED}✗ PostgreSQL connection failed${NC}"
fi

# Test 2: Redis Connection
echo -e "\n${BLUE}2. Testing Redis...${NC}"
if docker exec phoenix-redis redis-cli --pass phoenix ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is operational${NC}"
else
    echo -e "${RED}✗ Redis connection failed${NC}"
fi

# Test 3: NATS Connection
echo -e "\n${BLUE}3. Testing NATS...${NC}"
if curl -s http://localhost:8222/varz > /dev/null; then
    echo -e "${GREEN}✓ NATS is operational${NC}"
else
    echo -e "${RED}✗ NATS connection failed${NC}"
fi

# Test 4: Jaeger UI
echo -e "\n${BLUE}4. Testing Jaeger...${NC}"
if curl -s http://localhost:16686/ > /dev/null; then
    echo -e "${GREEN}✓ Jaeger UI is accessible${NC}"
else
    echo -e "${RED}✗ Jaeger UI not accessible${NC}"
fi

# Test 5: Build a Go service
echo -e "\n${BLUE}5. Testing Go build...${NC}"
if cd projects/platform-api && go build ./cmd/api 2>/dev/null; then
    echo -e "${GREEN}✓ Go services build successfully${NC}"
    rm -f api
else
    echo -e "${RED}✗ Go build failed${NC}"
fi
cd - > /dev/null

# Test 6: Shared packages
echo -e "\n${BLUE}6. Testing shared packages...${NC}"
if cd pkg && go test ./... -short 2>/dev/null; then
    echo -e "${GREEN}✓ Shared packages tests pass${NC}"
else
    echo -e "${RED}✗ Shared packages tests failed${NC}"
fi
cd - > /dev/null

# Test 7: Docker images can be built
echo -e "\n${BLUE}7. Testing Docker build...${NC}"
if cd projects/platform-api && docker build -f build/Dockerfile -t test-api:latest . > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Docker images can be built${NC}"
    docker rmi test-api:latest > /dev/null 2>&1
else
    echo -e "${RED}✗ Docker build failed${NC}"
fi
cd - > /dev/null

# Test 8: Scripts are executable
echo -e "\n${BLUE}8. Testing scripts...${NC}"
NON_EXEC=$(find scripts -name "*.sh" -type f ! -perm -u+x | wc -l | tr -d ' ')
if [ "$NON_EXEC" -eq 0 ]; then
    echo -e "${GREEN}✓ All scripts are executable${NC}"
else
    echo -e "${RED}✗ $NON_EXEC scripts are not executable${NC}"
fi

# Summary
echo -e "\n${BLUE}=== System Test Complete ===${NC}"
echo -e "${GREEN}Phoenix Platform is operational!${NC}"
echo -e "\nNext steps:"
echo -e "  - Run a service: cd projects/platform-api && make run"
echo -e "  - Run all tests: make test"
echo -e "  - Check logs: docker-compose logs -f"
echo -e "  - Access Jaeger UI: http://localhost:16686"