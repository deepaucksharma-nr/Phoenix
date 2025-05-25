#!/bin/bash
# Validate that all services can build

set -euo pipefail

echo "=== Phoenix Build Validation ==="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Track results
PASSED=0
FAILED=0
WARNINGS=0

# Function to test Go service build
test_go_build() {
    local service=$1
    local path=$2
    
    echo -n "Testing $service... "
    
    if [ ! -f "$path/go.mod" ]; then
        echo -e "${RED}✗ No go.mod${NC}"
        ((FAILED++))
        return
    fi
    
    if [ ! -d "$path/cmd" ] && [ ! -f "$path/main.go" ]; then
        echo -e "${YELLOW}⚠ No cmd directory or main.go${NC}"
        ((WARNINGS++))
        return
    fi
    
    # Test if it would build (dry run)
    if cd "$path" && go list ./... >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Valid${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ Build errors${NC}"
        ((FAILED++))
    fi
}

# Test shared packages
echo "=== Testing Shared Packages ==="
test_go_build "go-common" "packages/go-common"
test_go_build "contracts" "packages/contracts"

# Test core services
echo -e "\n=== Testing Core Services ==="
test_go_build "api" "services/api"
test_go_build "controller" "services/controller"
test_go_build "generator" "services/generator"

# Test analytics services
echo -e "\n=== Testing Analytics Services ==="
for svc in analytics anomaly-detector benchmark validator; do
    if [ -d "services/$svc" ]; then
        test_go_build "$svc" "services/$svc"
    fi
done

# Test operators
echo -e "\n=== Testing Operators ==="
test_go_build "loadsim-operator" "operators/loadsim"
test_go_build "pipeline-operator" "operators/pipeline"

# Test Node.js services
echo -e "\n=== Testing Node.js Services ==="
if [ -f "services/dashboard/package.json" ]; then
    echo -n "Testing dashboard... "
    if [ -f "services/dashboard/Dockerfile" ]; then
        echo -e "${GREEN}✓ Valid${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ No Dockerfile${NC}"
        ((FAILED++))
    fi
else
    echo "Dashboard: No package.json found"
fi

# Summary
echo -e "\n=== Build Validation Summary ==="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✅ All critical validations passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some validations failed!${NC}"
    exit 1
fi