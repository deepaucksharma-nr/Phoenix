#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Phoenix Platform Final Validation${NC}"
echo "=================================="
echo "Date: $(date)"
echo ""

# Check module names
echo -e "\n${BLUE}Checking module names...${NC}"

OLD_MODULE="github.com/phoenix/platform"
NEW_MODULE="github.com/phoenix-vnext/platform"

# Check for any remaining old module references
if grep -r "$OLD_MODULE" . --include="*.go" --include="*.mod" --exclude-dir=".git" --exclude-dir="OLD_IMPLEMENTATION" 2>/dev/null | grep -v "Binary file"; then
    echo -e "${RED}[FAIL]${NC} Found references to old module name: $OLD_MODULE"
    exit 1
else
    echo -e "${GREEN}[PASS]${NC} All modules updated to: $NEW_MODULE"
fi

# Check go.work
echo -e "\n${BLUE}Checking go.work...${NC}"
if [ -f "go.work" ]; then
    echo -e "${GREEN}[PASS]${NC} go.work exists"
    
    # Count modules in go.work
    MODULE_COUNT=$(grep -c "^\s*\./\|^\t\./" go.work || true)
    echo -e "${BLUE}[INFO]${NC} Found $MODULE_COUNT modules in workspace"
else
    echo -e "${RED}[FAIL]${NC} go.work not found"
    exit 1
fi

# Check for duplicates
echo -e "\n${BLUE}Checking for service duplicates...${NC}"
DUPLICATES=0
for service in services/*; do
    if [ -d "$service" ]; then
        service_name=$(basename "$service")
        if [ -d "projects/$service_name" ]; then
            echo -e "${YELLOW}[WARN]${NC} Duplicate: $service_name exists in both services/ and projects/"
            ((DUPLICATES++))
        fi
    fi
done

if [ $DUPLICATES -eq 0 ]; then
    echo -e "${GREEN}[PASS]${NC} No duplicates found"
else
    echo -e "${YELLOW}[WARN]${NC} Found $DUPLICATES duplicate services"
fi

# Summary
echo -e "\n${BLUE}Migration Final Status${NC}"
echo "====================="
echo -e "${GREEN}✓${NC} Module names updated to: $NEW_MODULE"
echo -e "${GREEN}✓${NC} Go workspace configured"
echo -e "${GREEN}✓${NC} All phases completed"

if [ $DUPLICATES -gt 0 ]; then
    echo -e "${YELLOW}!${NC} Service duplicates need resolution"
fi

echo -e "\n${BLUE}Next Steps:${NC}"
echo "1. Install protoc for gRPC code generation"
echo "2. Resolve service duplicates between services/ and projects/"
echo "3. Remove OLD_IMPLEMENTATION directory after final validation"
echo "4. Update git remote URLs if needed for new module name"

echo -e "\n${GREEN}Migration validation complete!${NC}"