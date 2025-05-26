#!/bin/bash
# Simplified Phoenix Platform startup for local development

set -e

echo "🚀 Starting Phoenix Platform (Simplified)..."
echo "=========================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Start PostgreSQL and Redis only
echo -e "${BLUE}Starting PostgreSQL and Redis...${NC}"
cd "$PROJECT_ROOT"
docker-compose up -d postgres redis

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}Waiting for PostgreSQL...${NC}"
until docker-compose exec -T postgres pg_isready -U phoenix > /dev/null 2>&1; do
  echo -n "."
  sleep 1
done
echo -e "${GREEN}✓ PostgreSQL ready${NC}"

# Start dashboard in development mode
echo -e "${BLUE}Starting Phoenix Dashboard...${NC}"
cd "$PROJECT_ROOT/projects/dashboard"

# Check if node_modules exists, install if not
if [ ! -d "node_modules" ]; then
  echo -e "${YELLOW}Installing dashboard dependencies...${NC}"
  npm install
fi

# Start dashboard in background
npm run dev &
DASHBOARD_PID=$!

# Give dashboard time to start
sleep 3

echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✨ Phoenix Platform is ready!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo -e "📊 Dashboard: ${BLUE}http://localhost:3000${NC}"
echo -e "🐘 PostgreSQL: ${BLUE}localhost:5432${NC} (user: phoenix, pass: phoenix)"
echo -e "🔴 Redis: ${BLUE}localhost:6379${NC}"
echo ""
echo -e "${YELLOW}Note: Phoenix API is not running. The dashboard will show mock data.${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"

# Keep script running and handle cleanup
cleanup() {
  echo -e "\n${YELLOW}Shutting down Phoenix Platform...${NC}"
  kill $DASHBOARD_PID 2>/dev/null || true
  cd "$PROJECT_ROOT"
  docker-compose down
  echo -e "${GREEN}✓ Phoenix Platform stopped${NC}"
}

trap cleanup EXIT

# Wait for user to press Ctrl+C
wait $DASHBOARD_PID