#!/bin/bash
# Check status of Phoenix development services

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "📊 Phoenix Development Status"
echo "============================="
echo ""

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"

# Function to check service
check_service() {
    local name=$1
    local check_cmd=$2
    local port=$3
    
    echo -n "$name "
    if [ -n "$port" ]; then
        echo -n "(port $port): "
    else
        echo -n ": "
    fi
    
    if eval "$check_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Running${NC}"
        return 0
    else
        echo -e "${RED}✗ Not running${NC}"
        return 1
    fi
}

# Function to check process
check_process() {
    local name=$1
    local pid_file=$2
    
    echo -n "$name: "
    
    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        if ps -p $PID > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Running (PID: $PID)${NC}"
            return 0
        else
            echo -e "${RED}✗ Not running (stale PID file)${NC}"
            return 1
        fi
    else
        # Try to find by process name
        if pgrep -f "$3" > /dev/null 2>&1; then
            PID=$(pgrep -f "$3" | head -1)
            echo -e "${GREEN}✓ Running (PID: $PID, no PID file)${NC}"
            return 0
        else
            echo -e "${RED}✗ Not running${NC}"
            return 1
        fi
    fi
}

# Infrastructure Services
echo -e "${BLUE}Infrastructure Services:${NC}"
echo "------------------------"
check_service "PostgreSQL" "nc -z localhost 5432" "5432"
check_service "Redis" "nc -z localhost 6379" "6379"
check_service "Prometheus" "curl -s http://localhost:9090/-/healthy" "9090"
check_service "Grafana" "curl -s http://localhost:3000/api/health" "3000"
echo ""

# Phoenix Services
echo -e "${BLUE}Phoenix Services:${NC}"
echo "-----------------"
check_process "Phoenix API" "$LOG_DIR/phoenix-api.pid" "phoenix-api"
check_process "Phoenix Agent 1" "$LOG_DIR/phoenix-agent-1.pid" "phoenix-agent.*local-agent-1"
check_process "Phoenix Agent 2" "$LOG_DIR/phoenix-agent-2.pid" "phoenix-agent.*local-agent-2"
echo ""

# API Endpoints
echo -e "${BLUE}API Health Checks:${NC}"
echo "------------------"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "Health endpoint: ${GREEN}✓ OK${NC}"
    
    # Additional endpoint checks
    echo -n "Pipelines endpoint: "
    status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/pipelines)
    if [ "$status" = "200" ]; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ Error (HTTP $status)${NC}"
    fi
    
    echo -n "Experiments endpoint: "
    status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/experiments)
    if [ "$status" = "200" ]; then
        echo -e "${GREEN}✓ OK${NC}"
    else
        echo -e "${RED}✗ Error (HTTP $status)${NC}"
    fi
else
    echo -e "API endpoints: ${RED}✗ Not accessible${NC}"
fi
echo ""

# Database Status
if nc -z localhost 5432 > /dev/null 2>&1; then
    echo -e "${BLUE}Database Status:${NC}"
    echo "----------------"
    
    # Check agent registration
    agent_count=$(docker exec phoenix-postgres psql -U phoenix -d phoenix -t -c "SELECT COUNT(*) FROM agents;" 2>/dev/null | tr -d ' ' || echo "0")
    echo "Registered agents: $agent_count"
    
    # Check experiments
    exp_count=$(docker exec phoenix-postgres psql -U phoenix -d phoenix -t -c "SELECT COUNT(*) FROM experiments;" 2>/dev/null | tr -d ' ' || echo "0")
    echo "Total experiments: $exp_count"
    
    # Check metrics
    metric_count=$(docker exec phoenix-postgres psql -U phoenix -d phoenix -t -c "SELECT COUNT(*) FROM metric_cache;" 2>/dev/null | tr -d ' ' || echo "0")
    echo "Cached metrics: $metric_count"
    echo ""
fi

# Recent Logs
echo -e "${BLUE}Recent Activity:${NC}"
echo "----------------"
if [ -f "$LOG_DIR/phoenix-api.log" ]; then
    echo "Phoenix API (last 5 lines):"
    tail -5 "$LOG_DIR/phoenix-api.log" | sed 's/^/  /'
    echo ""
fi

if [ -f "$LOG_DIR/phoenix-agent-1.log" ]; then
    echo "Phoenix Agent 1 (last 3 lines):"
    tail -3 "$LOG_DIR/phoenix-agent-1.log" | sed 's/^/  /'
    echo ""
fi

# Docker containers
echo -e "${BLUE}Docker Containers:${NC}"
echo "------------------"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(phoenix|NAME)" || echo "No Phoenix containers running"
echo ""

# Summary
echo -e "${BLUE}Summary:${NC}"
echo "--------"
ISSUES=0

# Check critical services
if ! nc -z localhost 5432 > /dev/null 2>&1; then
    echo -e "${RED}⚠ PostgreSQL is not running${NC}"
    ((ISSUES++))
fi

if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}⚠ Phoenix API is not accessible${NC}"
    ((ISSUES++))
fi

if ! pgrep -f "phoenix-agent" > /dev/null 2>&1; then
    echo -e "${RED}⚠ No Phoenix agents are running${NC}"
    ((ISSUES++))
fi

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✅ All systems operational${NC}"
else
    echo -e "${RED}❌ $ISSUES issue(s) detected${NC}"
    echo ""
    echo "To start services: ./scripts/dev-start.sh"
fi