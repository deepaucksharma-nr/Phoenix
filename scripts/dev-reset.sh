#!/bin/bash
# Reset Phoenix development environment

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "🔄 Resetting Phoenix Development Environment"
echo "==========================================="
echo ""
echo -e "${RED}WARNING: This will delete all data and logs!${NC}"
echo -n "Are you sure? (y/N): "
read -r response

if [[ ! "$response" =~ ^[Yy]$ ]]; then
    echo "Reset cancelled"
    exit 0
fi

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"

# Stop all services
echo -e "\n${YELLOW}Stopping all services...${NC}"
"$SCRIPT_DIR/dev-stop.sh" --all

# Remove Docker volumes
echo -e "\n${YELLOW}Removing Docker volumes...${NC}"
cd "$PROJECT_ROOT"
if [ -f "docker-compose.yml" ]; then
    docker-compose down -v
fi
if [ -f "docker-compose.dev.yml" ]; then
    docker-compose -f docker-compose.dev.yml down -v
fi

# Clean up logs
echo -e "\n${YELLOW}Cleaning up logs...${NC}"
rm -rf "$LOG_DIR"/*
echo "Logs cleaned"

# Clean up build artifacts
echo -e "\n${YELLOW}Cleaning build artifacts...${NC}"
find "$PROJECT_ROOT/projects" -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
find "$PROJECT_ROOT/projects" -name "build" -type d -exec rm -rf {} + 2>/dev/null || true
find "$PROJECT_ROOT/projects" -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true
echo "Build artifacts cleaned"

# Reset git-ignored files
if [ "$1" = "--deep" ]; then
    echo -e "\n${YELLOW}Deep clean: Removing git-ignored files...${NC}"
    cd "$PROJECT_ROOT"
    git clean -fdx -e ".env" -e "*.secret"
fi

# Recreate directories
echo -e "\n${YELLOW}Recreating directories...${NC}"
mkdir -p "$LOG_DIR"
mkdir -p "$PROJECT_ROOT"/{bin,build}

# Summary
echo -e "\n${GREEN}✅ Development environment reset complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Start infrastructure: docker-compose up -d"
echo "2. Build services: make build-all"
echo "3. Start Phoenix: ./scripts/dev-start.sh"
echo ""
echo "Note: You may need to run database migrations again."