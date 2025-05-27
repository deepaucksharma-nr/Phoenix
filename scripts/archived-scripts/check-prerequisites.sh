#!/bin/bash
# Check prerequisites for running Phoenix Platform

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Checking prerequisites for Phoenix Platform...${NC}\n"

READY=true

# Check Docker
echo -n "Docker: "
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    echo -e "${GREEN}✓${NC} (version $DOCKER_VERSION)"
else
    echo -e "${RED}✗ Not installed${NC}"
    echo "  Please install Docker: https://docs.docker.com/get-docker/"
    READY=false
fi

# Check Docker is running
echo -n "Docker daemon: "
if docker info &> /dev/null; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    echo "  Please start Docker Desktop"
    READY=false
fi

# Check Docker Compose
echo -n "Docker Compose: "
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $4}' | sed 's/,//')
    echo -e "${GREEN}✓${NC} (version $COMPOSE_VERSION)"
elif docker compose version &> /dev/null; then
    COMPOSE_VERSION=$(docker compose version | awk '{print $4}')
    echo -e "${GREEN}✓${NC} (Docker Compose V2: $COMPOSE_VERSION)"
else
    echo -e "${RED}✗ Not installed${NC}"
    echo "  Please install Docker Compose"
    READY=false
fi

# Check Go (optional)
echo -n "Go (optional): "
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} ($GO_VERSION)"
else
    echo -e "${YELLOW}⚠ Not installed${NC} (only needed for development)"
fi

# Check Node.js (optional)
echo -n "Node.js (optional): "
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}✓${NC} ($NODE_VERSION)"
else
    echo -e "${YELLOW}⚠ Not installed${NC} (only needed for dashboard development)"
fi

# Check available memory
echo -n "Available memory: "
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    TOTAL_MEM=$(sysctl -n hw.memsize | awk '{print $1/1024/1024/1024}')
    FREE_MEM=$(vm_stat | grep "Pages free" | awk '{print $3}' | sed 's/\.//' | awk '{print $1*4096/1024/1024/1024}')
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    TOTAL_MEM=$(free -g | awk '/^Mem:/{print $2}')
    FREE_MEM=$(free -g | awk '/^Mem:/{print $7}')
fi

if [ -n "$FREE_MEM" ]; then
    echo -e "${GREEN}${FREE_MEM}GB free${NC}"
    if (( $(echo "$FREE_MEM < 4" | bc -l) )); then
        echo -e "  ${YELLOW}⚠ Warning: Less than 4GB free. Phoenix needs ~4GB to run smoothly${NC}"
    fi
else
    echo -e "${YELLOW}Unable to check${NC}"
fi

# Check required ports
echo -e "\nChecking required ports..."
PORTS=(3000 5432 6379 8080 8081 9090 9091 3001)
PORT_ISSUES=false

for PORT in "${PORTS[@]}"; do
    echo -n "Port $PORT: "
    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        PROCESS=$(lsof -Pi :$PORT -sTCP:LISTEN | grep LISTEN | head -1 | awk '{print $1}')
        echo -e "${RED}✗ In use${NC} (by $PROCESS)"
        PORT_ISSUES=true
    else
        echo -e "${GREEN}✓ Available${NC}"
    fi
done

if [ "$PORT_ISSUES" = true ]; then
    echo -e "\n${YELLOW}Warning: Some ports are in use. You may need to:${NC}"
    echo "1. Stop the conflicting services, or"
    echo "2. Modify docker-compose.yml to use different ports"
fi

# Check disk space
echo -e "\nDisk space:"
if [[ "$OSTYPE" == "darwin"* ]]; then
    DISK_AVAIL=$(df -h / | awk 'NR==2 {print $4}')
else
    DISK_AVAIL=$(df -h / | awk 'NR==2 {print $4}')
fi
echo -e "Available: ${GREEN}$DISK_AVAIL${NC}"

# Summary
echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if [ "$READY" = true ] && [ "$PORT_ISSUES" = false ]; then
    echo -e "${GREEN}✅ All prerequisites met! You're ready to run Phoenix.${NC}"
    echo -e "\nNext step: ${GREEN}./scripts/start-phoenix-ui.sh${NC}"
else
    echo -e "${RED}❌ Some prerequisites are missing.${NC}"
    echo -e "Please address the issues above before continuing."
fi
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"