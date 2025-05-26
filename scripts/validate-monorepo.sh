#!/bin/bash
# Phoenix Platform Mono-repo Validation Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Phoenix Platform - Mono-repo Validation${NC}"
echo "========================================"
echo ""

# Initialize counters
ERRORS=0
WARNINGS=0
PASS=0

# Function to check condition
check() {
    local condition=$1
    local description=$2
    local severity=${3:-error}
    
    if eval "$condition"; then
        echo -e "${GREEN}✓${NC} $description"
        ((PASS++))
    else
        if [ "$severity" = "warning" ]; then
            echo -e "${YELLOW}⚠${NC} $description"
            ((WARNINGS++))
        else
            echo -e "${RED}✗${NC} $description"
            ((ERRORS++))
        fi
    fi
}

echo -e "${BLUE}1. Checking module consistency...${NC}"
# Check for module path inconsistencies
MODULE_PATHS=$(find . -name "go.mod" -type f -exec grep "^module" {} \; | sort -u)
VNEXT_COUNT=$(echo "$MODULE_PATHS" | grep -c "phoenix-vnext" || true)
check "[ $VNEXT_COUNT -eq 0 ]" "Module paths are consistent (no phoenix-vnext references)"

echo ""
echo -e "${BLUE}2. Checking go.work configuration...${NC}"
# Check go.work exists
check "[ -f go.work ]" "go.work file exists"

# Check all Go projects are in go.work
for dir in projects/*/; do
    if [ -f "$dir/go.mod" ]; then
        project=$(basename "$dir")
        check "grep -q \"./projects/$project\" go.work" "Project $project is in go.work"
    fi
done

echo ""
echo -e "${BLUE}3. Checking documentation files...${NC}"
# Check for referenced documentation
check "[ -f README.md ]" "README.md exists"
check "[ -f PLATFORM_STATUS.md ]" "PLATFORM_STATUS.md exists" "warning"
check "[ -f PRD_STATUS.md ]" "PRD_STATUS.md exists" "warning"
check "[ -f CONTRIBUTING.md ]" "CONTRIBUTING.md exists"
check "[ -f VERSION ]" "VERSION file exists"

echo ""
echo -e "${BLUE}4. Checking version consistency...${NC}"
# Check VERSION file content
VERSION_FILE=$(cat VERSION 2>/dev/null || echo "missing")
README_VERSION=$(grep -oP 'version:?\s*v?\K[0-9]+\.[0-9]+\.[0-9]+' README.md | head -1 || echo "not found")
check "[ '$VERSION_FILE' != 'missing' ]" "VERSION file is readable"
check "[ '$VERSION_FILE' = '$README_VERSION' ]" "Version in README ($README_VERSION) matches VERSION file ($VERSION_FILE)" "warning"

echo ""
echo -e "${BLUE}5. Checking service implementations...${NC}"
# Check for services mentioned in README
EXPECTED_SERVICES=("phoenix-api" "phoenix-agent" "phoenix-cli" "dashboard")
for service in "${EXPECTED_SERVICES[@]}"; do
    check "[ -d projects/$service ]" "Service $service exists"
done

# Check for core services
CORE_SERVICES=("phoenix-api" "phoenix-agent" "phoenix-cli")
for service in "${CORE_SERVICES[@]}"; do
    check "[ -d projects/$service ]" "Core service $service exists"
done

echo ""
echo -e "${BLUE}6. Checking dashboard integration...${NC}"
check "[ -d projects/dashboard ]" "Dashboard project exists"
check "[ -f projects/dashboard/package.json ]" "Dashboard has package.json"
check "grep -q dashboard Makefile" "Dashboard integrated in root Makefile" "warning"

echo ""
echo -e "${BLUE}7. Checking test structure...${NC}"
# Check for test directories
check "[ -d tests ]" "Root tests directory exists"
check "[ -d tests/integration ]" "Integration tests directory exists"
check "[ -d tests/e2e ]" "E2E tests directory exists"

# Check for unit tests in services
for project in projects/*/; do
    if [ -f "$project/go.mod" ]; then
        project_name=$(basename "$project")
        TEST_FILES=$(find "$project" -name "*_test.go" 2>/dev/null | wc -l)
        check "[ $TEST_FILES -gt 0 ]" "Project $project_name has test files" "warning"
    fi
done

echo ""
echo -e "${BLUE}8. Checking CI/CD configuration...${NC}"
check "[ -d .github/workflows ]" "GitHub workflows directory exists"
check "[ -f docker-compose.yml ]" "docker-compose.yml exists"
check "[ -f Makefile ]" "Root Makefile exists"

echo ""
echo -e "${BLUE}9. Checking security configuration...${NC}"
check "[ -f .gitignore ]" ".gitignore exists"
check "[ -f CODEOWNERS ]" "CODEOWNERS file exists"
check "[ -f .pre-commit-config.yaml ]" "Pre-commit hooks configured" "warning"

echo ""
echo -e "${BLUE}10. Checking build system...${NC}"
# Try to run go work sync
if go work sync 2>/dev/null; then
    echo -e "${GREEN}✓${NC} go work sync successful"
    ((PASS++))
else
    echo -e "${RED}✗${NC} go work sync failed"
    ((ERRORS++))
fi

echo ""
echo "========================================"
echo -e "${BLUE}Validation Summary:${NC}"
echo -e "  ${GREEN}Passed:${NC} $PASS"
echo -e "  ${YELLOW}Warnings:${NC} $WARNINGS"
echo -e "  ${RED}Errors:${NC} $ERRORS"
echo ""

if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}✅ Mono-repo validation passed with no issues!${NC}"
        exit 0
    else
        echo -e "${YELLOW}⚠️  Mono-repo validation passed with warnings.${NC}"
        exit 0
    fi
else
    echo -e "${RED}❌ Mono-repo validation failed with $ERRORS errors.${NC}"
    echo ""
    echo "Run the remediation script to fix critical issues:"
    echo "  ./scripts/fix-critical.sh"
    exit 1
fi
