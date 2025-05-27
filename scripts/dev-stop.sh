#!/bin/bash
# Stop Phoenix development services

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "🛑 Stopping Phoenix Development Services"
echo "========================================"
echo ""

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"

# Function to stop process by PID file
stop_by_pid_file() {
    local pid_file=$1
    local service_name=$2
    
    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        if ps -p $PID > /dev/null 2>&1; then
            echo -n "Stopping $service_name (PID: $PID)... "
            kill $PID 2>/dev/null || true
            
            # Wait for process to stop
            for i in {1..10}; do
                if ! ps -p $PID > /dev/null 2>&1; then
                    echo -e "${GREEN}Stopped${NC}"
                    rm -f "$pid_file"
                    return 0
                fi
                sleep 0.5
            done
            
            # Force kill if still running
            echo -n "force killing... "
            kill -9 $PID 2>/dev/null || true
            sleep 1
            echo -e "${GREEN}Stopped${NC}"
        else
            echo -e "$service_name PID file exists but process not running"
        fi
        rm -f "$pid_file"
    fi
}

# Stop Phoenix services using PID files
echo -e "${YELLOW}Stopping Phoenix services...${NC}"
stop_by_pid_file "$LOG_DIR/phoenix-api.pid" "Phoenix API"
stop_by_pid_file "$LOG_DIR/phoenix-agent-1.pid" "Phoenix Agent 1"
stop_by_pid_file "$LOG_DIR/phoenix-agent-2.pid" "Phoenix Agent 2"

# Also try to stop by process name (fallback)
echo -e "\n${YELLOW}Checking for remaining Phoenix processes...${NC}"

# Stop Phoenix API
if pgrep -f "phoenix-api" > /dev/null; then
    echo -n "Stopping Phoenix API processes... "
    pkill -f "phoenix-api" || true
    sleep 1
    echo -e "${GREEN}Done${NC}"
fi

# Stop Phoenix Agents
if pgrep -f "phoenix-agent" > /dev/null; then
    echo -n "Stopping Phoenix Agent processes... "
    pkill -f "phoenix-agent" || true
    sleep 1
    echo -e "${GREEN}Done${NC}"
fi

# Check if infrastructure should be stopped
if [ "$1" = "--all" ]; then
    echo -e "\n${YELLOW}Stopping infrastructure services...${NC}"
    cd "$PROJECT_ROOT"
    
    if [ -f "docker-compose.yml" ]; then
        docker-compose down
    elif [ -f "docker-compose.dev.yml" ]; then
        docker-compose -f docker-compose.dev.yml down
    fi
    
    echo -e "${GREEN}Infrastructure services stopped${NC}"
fi

# Clean up stale PID files
echo -e "\n${YELLOW}Cleaning up...${NC}"
rm -f "$LOG_DIR"/*.pid

# Show final status
echo -e "\n${YELLOW}Checking remaining processes...${NC}"
REMAINING=false

if pgrep -f "phoenix-api" > /dev/null; then
    echo -e "${RED}✗ Phoenix API processes still running${NC}"
    REMAINING=true
else
    echo -e "${GREEN}✓ No Phoenix API processes${NC}"
fi

if pgrep -f "phoenix-agent" > /dev/null; then
    echo -e "${RED}✗ Phoenix Agent processes still running${NC}"
    REMAINING=true
else
    echo -e "${GREEN}✓ No Phoenix Agent processes${NC}"
fi

if [ "$REMAINING" = true ]; then
    echo -e "\n${RED}Warning: Some processes are still running${NC}"
    echo "You may need to manually kill them:"
    echo "  ps aux | grep phoenix"
    echo "  kill -9 <PID>"
else
    echo -e "\n${GREEN}✅ All Phoenix services stopped successfully!${NC}"
fi

# Show infrastructure status if not stopped
if [ "$1" != "--all" ]; then
    echo -e "\n${YELLOW}Infrastructure services still running:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(phoenix|NAME)" || echo "None"
    echo ""
    echo "To stop infrastructure: $0 --all"
fi