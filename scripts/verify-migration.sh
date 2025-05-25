#!/bin/bash
# verify-migration.sh - Verify the Phoenix Platform migration is complete

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform Migration Verification ===${NC}"
echo ""

# Track verification results
ERRORS=0
WARNINGS=0

# Function to check item
check() {
    local description="$1"
    local command="$2"
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $description"
    else
        echo -e "${RED}✗${NC} $description"
        ((ERRORS++))
    fi
}

# Function to warn
warn() {
    local description="$1"
    local command="$2"
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $description"
    else
        echo -e "${YELLOW}⚠${NC} $description"
        ((WARNINGS++))
    fi
}

echo -e "${YELLOW}1. Directory Structure${NC}"
check "Packages directory exists" "[[ -d packages ]]"
check "Projects directory exists" "[[ -d projects ]]"
check "Scripts directory exists" "[[ -d scripts ]]"
check "Deployments directory exists" "[[ -d deployments ]]"
check "OLD_IMPLEMENTATION archived" "[[ ! -d OLD_IMPLEMENTATION ]]"
check "Archive exists" "ls archives/OLD_IMPLEMENTATION-*.tar.gz >/dev/null 2>&1"

echo ""
echo -e "${YELLOW}2. Go Workspace${NC}"
check "go.work exists" "[[ -f go.work ]]"
check "Go workspace is valid" "go work sync"

echo ""
echo -e "${YELLOW}3. Shared Packages${NC}"
check "go-common package exists" "[[ -f packages/go-common/go.mod ]]"
check "contracts package exists" "[[ -f packages/contracts/go.mod ]]"

echo ""
echo -e "${YELLOW}4. Services Migration${NC}"
EXPECTED_SERVICES=(
    "analytics"
    "anomaly-detector"
    "api"
    "benchmark"
    "collector"
    "control-actuator-go"
    "controller"
    "dashboard"
    "generator"
    "loadsim-operator"
    "pipeline-operator"
    "platform-api"
    "phoenix-cli"
)

for service in "${EXPECTED_SERVICES[@]}"; do
    check "Service $service migrated" "[[ -d projects/$service ]]"
done

echo ""
echo -e "${YELLOW}5. Validation Scripts${NC}"
check "Boundary check script exists" "[[ -x scripts/validate-boundaries.sh ]]"
check "Structure validation script exists" "[[ -x scripts/validate-structure.sh ]]"
check "Import validation script exists" "[[ -x scripts/update-imports.sh ]]"

echo ""
echo -e "${YELLOW}6. Development Tools${NC}"
check "Deploy script exists" "[[ -f scripts/deploy-dev.sh ]]"
check "Setup script exists" "[[ -f scripts/setup-dev-env.sh ]]"
check "Archive script exists" "[[ -f scripts/archive-old-implementation.sh ]]"

echo ""
echo -e "${YELLOW}7. Documentation${NC}"
check "README.md updated" "grep -q 'Monorepo Structure' README.md"
check "CLAUDE.md exists" "[[ -f CLAUDE.md ]]"
check "Migration summary exists" "[[ -f MIGRATION_SUMMARY.md ]]"
warn "E2E demo guide exists" "[[ -f E2E_DEMO_GUIDE.md ]]"

echo ""
echo -e "${YELLOW}8. Git Status${NC}"
# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${YELLOW}⚠${NC} Uncommitted changes found"
    ((WARNINGS++))
else
    echo -e "${GREEN}✓${NC} All changes committed"
fi

echo ""
echo -e "${YELLOW}9. Import Boundaries${NC}"
echo "Running boundary validation..."
if ./scripts/validate-boundaries.sh > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} No cross-project imports found"
else
    echo -e "${RED}✗${NC} Cross-project import violations detected"
    ((ERRORS++))
fi

echo ""
echo -e "${YELLOW}10. Go Module Health${NC}"
# Check if all Go modules can be built
echo "Checking Go module health..."
SUCCESS=0
TOTAL=0
for project in projects/*/; do
    if [[ -f "$project/go.mod" ]]; then
        ((TOTAL++))
        if (cd "$project" && go mod tidy > /dev/null 2>&1); then
            ((SUCCESS++))
        fi
    fi
done
if [[ $SUCCESS -eq $TOTAL ]]; then
    echo -e "${GREEN}✓${NC} All $TOTAL Go modules are healthy"
else
    echo -e "${YELLOW}⚠${NC} $SUCCESS/$TOTAL Go modules are healthy"
    ((WARNINGS++))
fi

echo ""
echo -e "${BLUE}=== Verification Summary ===${NC}"
if [[ $ERRORS -eq 0 ]]; then
    if [[ $WARNINGS -eq 0 ]]; then
        echo -e "${GREEN}✅ Migration verification passed with no issues!${NC}"
    else
        echo -e "${GREEN}✅ Migration verification passed with $WARNINGS warnings${NC}"
    fi
else
    echo -e "${RED}❌ Migration verification failed with $ERRORS errors and $WARNINGS warnings${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Deploy to development: cd scripts && ./deploy-dev.sh"
echo "2. Set up local dev environment: ./scripts/setup-dev-env.sh"
echo "3. Run tests: go test ./..."
echo "4. Update CI/CD pipelines for the new structure"