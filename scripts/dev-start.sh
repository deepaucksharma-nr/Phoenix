#!/bin/bash
# Start Phoenix services for local development

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Starting Phoenix Development Services"
echo "========================================"
echo ""

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"

# Create log directory
mkdir -p "$LOG_DIR"

# Function to check if service is running
check_service() {
    local name=$1
    local check_cmd=$2
    
    if eval "$check_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $name is running${NC}"
        return 0
    else
        echo -e "${RED}✗ $name is not running${NC}"
        return 1
    fi
}

# Check infrastructure services
echo -e "${YELLOW}Checking infrastructure services...${NC}"
INFRA_OK=true

if ! check_service "PostgreSQL" "nc -z localhost 5432"; then
    INFRA_OK=false
fi
if ! check_service "Redis" "nc -z localhost 6379"; then
    INFRA_OK=false
fi
if ! check_service "Prometheus" "curl -s http://localhost:9090/-/healthy"; then
    INFRA_OK=false
fi

if [ "$INFRA_OK" = false ]; then
    echo -e "\n${YELLOW}Starting infrastructure services...${NC}"
    docker-compose -f "$PROJECT_ROOT/docker-compose.yml" up -d postgres redis prometheus
    
    echo "Waiting for services to be ready..."
    sleep 10
fi

# Kill any existing Phoenix processes
echo -e "\n${YELLOW}Stopping any existing Phoenix services...${NC}"
pkill -f "phoenix-api" || true
pkill -f "phoenix-agent" || true
sleep 2

# Build services if needed
echo -e "\n${YELLOW}Building Phoenix services...${NC}"

# Build Phoenix API
if [ ! -f "$PROJECT_ROOT/projects/phoenix-api/bin/phoenix-api" ] || [ "$1" = "--rebuild" ]; then
    echo "Building Phoenix API..."
    cd "$PROJECT_ROOT/projects/phoenix-api"
    make build
fi

# Build Phoenix Agent  
if [ ! -f "$PROJECT_ROOT/projects/phoenix-agent/bin/phoenix-agent" ] || [ "$1" = "--rebuild" ]; then
    echo "Building Phoenix Agent..."
    cd "$PROJECT_ROOT/projects/phoenix-agent"
    make build
fi

# Build Phoenix CLI
if [ ! -f "$PROJECT_ROOT/projects/phoenix-cli/phoenix-cli" ] || [ "$1" = "--rebuild" ]; then
    echo "Building Phoenix CLI..."
    cd "$PROJECT_ROOT/projects/phoenix-cli"
    make build
fi

# Start Phoenix API
echo -e "\n${YELLOW}Starting Phoenix API...${NC}"
cd "$PROJECT_ROOT/projects/phoenix-api"
DATABASE_PASSWORD=phoenix-dev SKIP_MIGRATIONS=true nohup ./bin/phoenix-api > "$LOG_DIR/phoenix-api.log" 2>&1 &
API_PID=$!
echo "Phoenix API started (PID: $API_PID)"

# Wait for API to be ready
echo -n "Waiting for API to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e " ${GREEN}Ready!${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

# Start Phoenix Agent(s)
echo -e "\n${YELLOW}Starting Phoenix Agent(s)...${NC}"

# Start local agent 1
cd "$PROJECT_ROOT/projects/phoenix-agent"
nohup ./bin/phoenix-agent \
    -host-id=local-agent-1 \
    -api-url=http://localhost:8080 \
    -poll-interval=15s \
    > "$LOG_DIR/phoenix-agent-1.log" 2>&1 &
AGENT1_PID=$!
echo "Phoenix Agent 1 started (PID: $AGENT1_PID)"

# Optionally start additional agents
if [ "$2" = "--multi-agent" ]; then
    nohup ./bin/phoenix-agent \
        -host-id=local-agent-2 \
        -api-url=http://localhost:8080 \
        -poll-interval=15s \
        > "$LOG_DIR/phoenix-agent-2.log" 2>&1 &
    AGENT2_PID=$!
    echo "Phoenix Agent 2 started (PID: $AGENT2_PID)"
fi

# Save PIDs for later
echo "$API_PID" > "$LOG_DIR/phoenix-api.pid"
echo "$AGENT1_PID" > "$LOG_DIR/phoenix-agent-1.pid"
[ -n "$AGENT2_PID" ] && echo "$AGENT2_PID" > "$LOG_DIR/phoenix-agent-2.pid"

# Wait for services to stabilize
sleep 3

# Check final status
echo -e "\n${YELLOW}Verifying services...${NC}"
"$SCRIPT_DIR/dev-status.sh"

# Start log monitoring
echo -e "\n${YELLOW}Log files:${NC}"
echo "  Phoenix API: $LOG_DIR/phoenix-api.log"
echo "  Phoenix Agent 1: $LOG_DIR/phoenix-agent-1.log"
[ -n "$AGENT2_PID" ] && echo "  Phoenix Agent 2: $LOG_DIR/phoenix-agent-2.log"

echo -e "\n${GREEN}✅ Phoenix development services started!${NC}"
echo ""
echo "Next steps:"
echo "  1. Check status: ./scripts/dev-status.sh"
echo "  2. View logs: tail -f logs/*.log"
echo "  3. Start dashboard: cd projects/dashboard && npm run dev"
echo "  4. Use CLI: ./projects/phoenix-cli/phoenix-cli --help"
echo ""
echo "To stop services: ./scripts/dev-stop.sh"