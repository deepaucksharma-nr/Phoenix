#!/bin/bash
# final-pre-push-check.sh - Final validation before pushing to remote

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform - Final Pre-Push Validation ===${NC}"
echo ""

# Initialize counters
CHECKS_PASSED=0
CHECKS_FAILED=0
WARNINGS=0

# Function to run a check
check() {
    local description="$1"
    local command="$2"
    local critical="${3:-true}"
    
    echo -ne "Checking: $description... "
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        ((CHECKS_PASSED++))
        return 0
    else
        if [[ "$critical" == "true" ]]; then
            echo -e "${RED}✗${NC}"
            ((CHECKS_FAILED++))
            return 1
        else
            echo -e "${YELLOW}⚠${NC}"
            ((WARNINGS++))
            return 0
        fi
    fi
}

# 1. Git Status
echo -e "${CYAN}1. Repository Status${NC}"
check "Clean working directory" "[[ -z \$(git status --porcelain) ]]"
check "On main branch" "[[ \$(git branch --show-current) == 'main' ]]"
check "Commits to push" "[[ \$(git rev-list --count origin/main..HEAD) -gt 0 ]]"

# 2. Migration Validation
echo ""
echo -e "${CYAN}2. Migration Validation${NC}"
check "Projects directory exists" "[[ -d projects ]]"
check "Packages directory exists" "[[ -d packages ]]"
check "OLD_IMPLEMENTATION archived" "[[ ! -d OLD_IMPLEMENTATION ]]"
check "Archive exists" "ls archives/OLD_IMPLEMENTATION-*.tar.gz >/dev/null 2>&1"

# 3. Go Workspace
echo ""
echo -e "${CYAN}3. Go Workspace Health${NC}"
check "go.work exists" "[[ -f go.work ]]"
check "Go workspace valid" "go work sync"
check "All modules build" "go build ./..." "false"

# 4. Import Boundaries
echo ""
echo -e "${CYAN}4. Architectural Boundaries${NC}"
if ./scripts/validate-boundaries.sh > /dev/null 2>&1; then
    echo -e "Checking: Import boundaries... ${GREEN}✓${NC}"
    ((CHECKS_PASSED++))
else
    echo -e "Checking: Import boundaries... ${RED}✗${NC}"
    echo -e "${RED}  Run ./scripts/validate-boundaries.sh for details${NC}"
    ((CHECKS_FAILED++))
fi

# 5. Documentation
echo ""
echo -e "${CYAN}5. Documentation Completeness${NC}"
check "README.md updated" "grep -q 'Monorepo Structure' README.md"
check "CLAUDE.md exists" "[[ -f CLAUDE.md ]]"
check "Team onboarding guide" "[[ -f TEAM_ONBOARDING.md ]]"
check "Migration summary" "[[ -f MIGRATION_SUMMARY.md ]]"
check "Handoff checklist" "[[ -f HANDOFF_CHECKLIST.md ]]"

# 6. Development Tools
echo ""
echo -e "${CYAN}6. Developer Tools${NC}"
check "Quick-start script" "[[ -x scripts/quick-start.sh ]]"
check "Deploy script" "[[ -f scripts/deploy-dev.sh ]]"
check "Validation scripts" "[[ -f scripts/validate-boundaries.sh ]]"
check "Pre-commit hook" "[[ -x .git/hooks/pre-commit ]]"

# 7. Service Health Check
echo ""
echo -e "${CYAN}7. Service Migration Status${NC}"
EXPECTED_SERVICES=(
    "analytics" "anomaly-detector" "api" "benchmark" 
    "collector" "control-actuator-go" "controller" 
    "dashboard" "generator" "loadsim-operator" 
    "pipeline-operator" "platform-api" "phoenix-cli"
)

for service in "${EXPECTED_SERVICES[@]}"; do
    check "Service $service" "[[ -d projects/$service ]]"
done

# Summary
echo ""
echo -e "${BLUE}=== Validation Summary ===${NC}"
echo -e "Checks Passed: ${GREEN}$CHECKS_PASSED${NC}"
echo -e "Checks Failed: ${RED}$CHECKS_FAILED${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"

# Git information
echo ""
echo -e "${BLUE}=== Git Information ===${NC}"
echo "Branch: $(git branch --show-current)"
echo "Commits to push: $(git rev-list --count origin/main..HEAD)"
echo "Last commit: $(git log -1 --oneline)"

# Final recommendation
echo ""
if [[ $CHECKS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}✅ All critical checks passed!${NC}"
    echo ""
    echo -e "${GREEN}Ready to push with:${NC}"
    echo -e "${CYAN}  git push origin main${NC}"
    echo ""
    echo -e "${YELLOW}Don't forget to:${NC}"
    echo "1. Notify the team (see HANDOFF_CHECKLIST.md)"
    echo "2. Monitor CI/CD pipelines"
    echo "3. Be available for questions"
    exit 0
else
    echo -e "${RED}❌ Critical checks failed!${NC}"
    echo ""
    echo "Please fix the issues above before pushing."
    exit 1
fi