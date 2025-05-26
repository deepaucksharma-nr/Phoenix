#!/bin/bash
# push-to-remote.sh - Final push script with safety checks

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform - Push to Remote ===${NC}"
echo ""

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo -e "${RED}Error: Not on main branch (currently on $CURRENT_BRANCH)${NC}"
    exit 1
fi

# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${RED}Error: Uncommitted changes found${NC}"
    git status --short
    exit 1
fi

# Count commits to push
COMMITS_TO_PUSH=$(git rev-list --count origin/main..HEAD)
if [[ $COMMITS_TO_PUSH -eq 0 ]]; then
    echo -e "${YELLOW}No commits to push${NC}"
    exit 0
fi

# Show summary
echo -e "${CYAN}Ready to push $COMMITS_TO_PUSH commits:${NC}"
echo ""
git log --oneline origin/main..HEAD | head -10
if [[ $COMMITS_TO_PUSH -gt 10 ]]; then
    echo "... and $((COMMITS_TO_PUSH - 10)) more"
fi

echo ""
echo -e "${YELLOW}This will push the complete Phoenix Platform migration including:${NC}"
echo "- Monorepo structure transformation"
echo "- Service consolidation to projects/"
echo "- Removal of duplicate services"
echo "- Complete documentation suite"
echo "- Development tooling"
echo ""

# Confirmation
echo -e "${YELLOW}Are you sure you want to push to origin/main? (yes/no)${NC}"
read -r response

if [[ "$response" != "yes" ]]; then
    echo "Push cancelled"
    exit 0
fi

# Final validation
echo ""
echo -e "${CYAN}Running final validation...${NC}"
if ! ./scripts/validate-boundaries.sh > /dev/null 2>&1; then
    echo -e "${RED}Boundary validation failed!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Validation passed${NC}"

# Push to remote
echo ""
echo -e "${CYAN}Pushing to remote...${NC}"
if git push origin main; then
    echo ""
    echo -e "${GREEN}✅ Push successful!${NC}"
    echo ""
    echo -e "${BLUE}=== Next Steps ===${NC}"
    echo "1. Monitor CI/CD pipelines"
    echo "2. Send team notification (see HANDOFF_CHECKLIST.md)"
    echo "3. Share TEAM_ONBOARDING.md"
    echo "4. Be available for questions"
    echo ""
    echo -e "${GREEN}The Phoenix has risen! 🦅${NC}"
else
    echo ""
    echo -e "${RED}Push failed!${NC}"
    echo "Check your permissions and network connection"
    exit 1
fi