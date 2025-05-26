#!/bin/bash
# Start Phoenix Platform with UI-first experience

set -e

echo "🚀 Starting Phoenix Platform with Visual UI..."
echo "================================================"

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Aborting." >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed. Aborting." >&2; exit 1; }

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Start backend services
echo -e "${BLUE}Starting backend services...${NC}"
cd "$PROJECT_ROOT"
docker-compose up -d postgres redis

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}Waiting for PostgreSQL...${NC}"
until docker-compose exec -T postgres pg_isready -U phoenix > /dev/null 2>&1; do
  echo -n "."
  sleep 1
done
echo -e "${GREEN}✓ PostgreSQL ready${NC}"

# Start other required services
echo -e "${BLUE}Starting Prometheus and Pushgateway...${NC}"
docker-compose up -d prometheus pushgateway

# Start Phoenix API with WebSocket support
echo -e "${BLUE}Starting Phoenix API...${NC}"
docker-compose up -d phoenix-api

# Wait for API to be ready
echo -e "${YELLOW}Waiting for Phoenix API...${NC}"
until curl -s http://localhost:8080/health > /dev/null 2>&1; do
  echo -n "."
  sleep 1
done
echo -e "${GREEN}✓ Phoenix API ready${NC}"

# Run database migrations
echo -e "${BLUE}Running database migrations...${NC}"
docker-compose exec -T phoenix-api sh -c "cd /app && ./phoenix-api migrate up" || true

# Start Phoenix agents
echo -e "${BLUE}Starting Phoenix agents...${NC}"
docker-compose up -d phoenix-agent

# Check if dashboard directory exists
if [ ! -d "$PROJECT_ROOT/projects/dashboard" ]; then
  echo -e "${RED}❌ Dashboard directory not found at $PROJECT_ROOT/projects/dashboard${NC}"
  exit 1
fi

# Build and start dashboard
echo -e "${BLUE}Starting Phoenix Dashboard...${NC}"
cd "$PROJECT_ROOT/projects/dashboard"

# Check if node_modules exists, install if not
if [ ! -d "node_modules" ]; then
  echo -e "${YELLOW}Installing dashboard dependencies...${NC}"
  npm install
fi

# Start dashboard using docker-compose if available, otherwise use npm
if grep -q "phoenix-dashboard" "$PROJECT_ROOT/docker-compose.yml"; then
  cd "$PROJECT_ROOT"
  docker-compose up -d phoenix-dashboard
  USING_DOCKER=true
else
  # Start dashboard in development mode
  npm run dev &
  DASHBOARD_PID=$!
  USING_DOCKER=false
fi

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 5

# Check if everything is running
if curl -s http://localhost:8080/health > /dev/null; then
  echo -e "${GREEN}✓ Phoenix API is running${NC}"
else
  echo -e "❌ Phoenix API failed to start"
  exit 1
fi

# Open browser
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✨ Phoenix Platform is ready!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo -e "📊 Dashboard: ${BLUE}http://localhost:3000${NC}"
echo -e "🔌 API: ${BLUE}http://localhost:8080${NC}"
echo -e "📡 WebSocket: ${BLUE}ws://localhost:8081${NC}"
echo ""
echo -e "${YELLOW}Opening dashboard in your browser...${NC}"

# Open browser based on OS
if [[ "$OSTYPE" == "darwin"* ]]; then
  open http://localhost:3000
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
  xdg-open http://localhost:3000 2>/dev/null || echo "Please open http://localhost:3000 in your browser"
else
  echo "Please open http://localhost:3000 in your browser"
fi

echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"

# Keep script running and handle cleanup
cleanup() {
  echo -e "\n${YELLOW}Shutting down Phoenix Platform...${NC}"
  if [ "$USING_DOCKER" = "false" ] && [ ! -z "$DASHBOARD_PID" ]; then
    kill $DASHBOARD_PID 2>/dev/null || true
  fi
  cd "$PROJECT_ROOT"
  docker-compose down
  echo -e "${GREEN}✓ Phoenix Platform stopped${NC}"
}

trap cleanup EXIT

# Change back to project root
cd "$PROJECT_ROOT"

# Wait for user to press Ctrl+C
if [ "$USING_DOCKER" = "false" ] && [ ! -z "$DASHBOARD_PID" ]; then
  wait $DASHBOARD_PID
else
  while true; do sleep 1; done
fi