#!/bin/bash

# Phoenix Platform Final Migration Check
# This script performs a comprehensive check of the migration

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Phoenix Platform Final Migration Check               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Change to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/.."

# Initialize counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Check function
check() {
    local description="$1"
    local command="$2"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $description"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo -e "${RED}✗${NC} $description"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 1. Check Go workspace
echo -e "${YELLOW}Checking Go Workspace...${NC}"
check "go.work exists" "test -f go.work"
check "go.work is valid" "go work sync"

# 2. Check core services
echo -e "\n${YELLOW}Checking Core Services...${NC}"
check "API service module exists" "test -f services/api/go.mod"
check "Controller service module exists" "test -f services/controller/go.mod"
check "Generator service module exists" "test -f services/generator/go.mod"
check "Phoenix CLI module exists" "test -f services/phoenix-cli/go.mod"

# 3. Check packages
echo -e "\n${YELLOW}Checking Shared Packages...${NC}"
check "go-common package exists" "test -d packages/go-common"
check "contracts package exists" "test -d packages/contracts"
check "go-common has auth package" "test -d packages/go-common/auth"
check "go-common has store package" "test -d packages/go-common/store"

# 4. Check for old module names
echo -e "\n${YELLOW}Checking for Old Module Names...${NC}"
OLD_COUNT=$(find . -name "go.mod" -type f -exec grep -l "phoenix-vnext" {} \; 2>/dev/null | wc -l)
if [ "$OLD_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No old module names found"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "${RED}✗${NC} Found $OLD_COUNT files with old module names"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# 5. Check for old imports
echo -e "\n${YELLOW}Checking for Old Imports...${NC}"
OLD_IMPORTS=$(find . -name "*.go" -type f -exec grep -l "phoenix-vnext/platform" {} \; 2>/dev/null | wc -l)
if [ "$OLD_IMPORTS" -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No old imports found"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "${RED}✗${NC} Found $OLD_IMPORTS files with old imports"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# 6. Check documentation
echo -e "\n${YELLOW}Checking Documentation...${NC}"
check "MIGRATION_SUMMARY.md exists" "test -f MIGRATION_SUMMARY.md"
check "NEXT_STEPS.md exists" "test -f NEXT_STEPS.md"
check "QUICK_START.md exists" "test -f QUICK_START.md"
check "DEVELOPMENT_GUIDE.md exists" "test -f DEVELOPMENT_GUIDE.md"

# 7. Check Phoenix CLI build
echo -e "\n${YELLOW}Checking Phoenix CLI...${NC}"
if [ -f "services/phoenix-cli/bin/phoenix" ]; then
    echo -e "${GREEN}✓${NC} Phoenix CLI binary exists"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "${YELLOW}⚠${NC} Phoenix CLI binary not built (run: cd services/phoenix-cli && go build -o bin/phoenix .)"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# 8. Summary
echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                        Check Summary                           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Total Checks: ${TOTAL_CHECKS}"
echo -e "Passed: ${GREEN}${PASSED_CHECKS}${NC}"
echo -e "Failed: ${RED}${FAILED_CHECKS}${NC}"
echo ""

if [ "$FAILED_CHECKS" -eq 0 ]; then
    echo -e "${GREEN}🎉 All checks passed! The migration is complete.${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "1. Install protoc: bash scripts/install-protoc.sh"
    echo "2. Generate protos: cd packages/contracts && bash generate.sh"
    echo "3. Build services: go work sync && make build-all"
    echo "4. Run tests: go test ./..."
    exit 0
else
    echo -e "${RED}⚠️  Some checks failed. Please review the issues above.${NC}"
    exit 1
fi