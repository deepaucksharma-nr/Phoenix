#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Removing Duplicate Services${NC}"
echo "==========================="
echo ""

# Services to remove from services/ directory (keeping projects/)
TO_REMOVE=(
    "services/analytics"
    "services/benchmark"
    "services/loadsim-operator"
    "services/phoenix-cli"
    "services/pipeline-operator"
)

# Services that need code merged from services/ to projects/
TO_MERGE=(
    "anomaly-detector"
    "api"
    "controller"
    "generator"
)

# Remove duplicates where projects/ is clearly better
echo -e "${BLUE}Removing services that exist in projects/ with better structure:${NC}"
for service in "${TO_REMOVE[@]}"; do
    if [ -d "$service" ]; then
        echo -e "${YELLOW}Removing: $service${NC}"
        rm -rf "$service"
        echo -e "${GREEN}✓ Removed${NC}"
    fi
done

# Merge services where services/ has implementation
echo -e "\n${BLUE}Services that need manual merge:${NC}"
for service in "${TO_MERGE[@]}"; do
    if [ -d "services/$service" ] && [ -d "projects/$service" ]; then
        echo -e "${YELLOW}$service:${NC}"
        echo "  - Check services/$service for implementation"
        echo "  - Merge into projects/$service"
        echo "  - Update module name to github.com/phoenix-vnext/platform"
    fi
done

# Handle special cases
echo -e "\n${BLUE}Special cases:${NC}"

# Collector - Node.js service
if [ -d "services/collector" ] && [ -d "projects/collector" ]; then
    echo -e "${YELLOW}collector:${NC} Node.js service, keep projects/ version"
    rm -rf "services/collector"
    echo -e "${GREEN}✓ Removed services/collector${NC}"
fi

# Dashboard - React app
if [ -d "services/dashboard" ] && [ -d "projects/dashboard" ]; then
    echo -e "${YELLOW}dashboard:${NC} React app, keep projects/ version"
    rm -rf "services/dashboard"
    echo -e "${GREEN}✓ Removed services/dashboard${NC}"
fi

# control-actuator-go
if [ -d "services/control-actuator-go" ] && [ -d "projects/control-actuator-go" ]; then
    echo -e "${YELLOW}control-actuator-go:${NC} Keep projects/ version"
    rm -rf "services/control-actuator-go"
    echo -e "${GREEN}✓ Removed services/control-actuator-go${NC}"
fi

# Update go.work
echo -e "\n${BLUE}Updating go.work...${NC}"
cp go.work go.work.backup
echo "Created backup: go.work.backup"

# Remove deleted paths from go.work
for service in "${TO_REMOVE[@]}"; do
    service_path=$(echo "$service" | sed 's/services\///')
    sed -i '' "/\.\/services\/$service_path/d" go.work 2>/dev/null || true
done

# Also remove the merged ones
sed -i '' "/\.\/services\/collector/d" go.work 2>/dev/null || true
sed -i '' "/\.\/services\/dashboard/d" go.work 2>/dev/null || true
sed -i '' "/\.\/services\/control-actuator-go/d" go.work 2>/dev/null || true

echo -e "${GREEN}✓ Updated go.work${NC}"

# Summary
echo -e "\n${BLUE}Summary:${NC}"
echo "========="
echo -e "${GREEN}Removed:${NC}"
echo "- services/analytics"
echo "- services/benchmark"
echo "- services/collector"
echo "- services/control-actuator-go"
echo "- services/dashboard"
echo "- services/loadsim-operator"
echo "- services/phoenix-cli"
echo "- services/pipeline-operator"

echo -e "\n${YELLOW}Need manual merge:${NC}"
echo "- anomaly-detector"
echo "- api"
echo "- controller"
echo "- generator"

echo -e "\n${BLUE}Next steps:${NC}"
echo "1. Manually merge the 4 services listed above"
echo "2. Run: go work sync"
echo "3. Commit changes"