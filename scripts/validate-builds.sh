#!/bin/bash
# validate-builds.sh - Validate all service builds in the monorepo

set -euo pipefail

echo "=== Validating Phoenix Platform Builds ==="
echo ""

FAILED=0
SUCCEEDED=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Validate Go workspace
echo -e "${YELLOW}Syncing Go workspace...${NC}"
go work sync

# Build packages
echo ""
echo -e "${YELLOW}Building shared packages...${NC}"
for pkg in packages/go-common packages/contracts; do
    if [[ -d "$pkg" ]] && [[ -f "$pkg/go.mod" ]]; then
        echo -n "Building $pkg... "
        if (cd "$pkg" && go build ./...) > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
            ((SUCCEEDED++))
        else
            echo -e "${RED}✗${NC}"
            ((FAILED++))
        fi
    fi
done

# Build projects
echo ""
echo -e "${YELLOW}Building projects...${NC}"
for project in projects/*/; do
    if [[ -f "$project/go.mod" ]]; then
        project_name=$(basename "$project")
        echo -n "Building $project_name... "
        
        if (cd "$project" && go build ./...) > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
            ((SUCCEEDED++))
        else
            echo -e "${RED}✗${NC}"
            ((FAILED++))
            # Show error
            echo "  Error details:"
            (cd "$project" && go build ./... 2>&1 | head -5 | sed 's/^/    /')
        fi
    fi
done

# Build legacy services if they exist
if [[ -d "services" ]]; then
    echo ""
    echo -e "${YELLOW}Building legacy services...${NC}"
    for service in services/*/; do
        if [[ -f "$service/go.mod" ]]; then
            service_name=$(basename "$service")
            echo -n "Building $service_name... "
            
            if (cd "$service" && go build ./...) > /dev/null 2>&1; then
                echo -e "${GREEN}✓${NC}"
                ((SUCCEEDED++))
            else
                echo -e "${RED}✗${NC}"
                ((FAILED++))
            fi
        fi
    done
fi

# Summary
echo ""
echo "=== Build Summary ==="
echo -e "Succeeded: ${GREEN}$SUCCEEDED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"

if [[ $FAILED -gt 0 ]]; then
    echo ""
    echo -e "${RED}❌ Build validation FAILED${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}✅ All builds PASSED${NC}"
    exit 0
fi