#!/bin/bash
# validate-structure.sh - Enforce Phoenix platform directory structure

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔍 Validating Phoenix Platform Structure..."

ERRORS=0
WARNINGS=0

# Function to check required directories
check_required_dir() {
    local dir="$1"
    local description="$2"
    
    if [ ! -d "$PROJECT_ROOT/$dir" ]; then
        echo -e "${RED}❌ Missing required directory: $dir ($description)${NC}"
        ((ERRORS++))
        return 1
    else
        echo -e "${GREEN}✓ Found: $dir${NC}"
        return 0
    fi
}

# Function to check forbidden patterns
check_forbidden() {
    local pattern="$1"
    local description="$2"
    
    if find "$PROJECT_ROOT" -path "$PROJECT_ROOT/.git" -prune -o -name "$pattern" -print | grep -q .; then
        echo -e "${RED}❌ Forbidden pattern found: $pattern ($description)${NC}"
        find "$PROJECT_ROOT" -path "$PROJECT_ROOT/.git" -prune -o -name "$pattern" -print
        ((ERRORS++))
        return 1
    fi
    return 0
}

# Function to check service structure
check_service() {
    local service="$1"
    
    echo -e "\n${YELLOW}Checking service: $service${NC}"
    
    if [ -d "$PROJECT_ROOT/cmd/$service" ]; then
        # Check for main.go
        if [ ! -f "$PROJECT_ROOT/cmd/$service/main.go" ]; then
            echo -e "${RED}❌ Missing main.go in cmd/$service${NC}"
            ((ERRORS++))
        else
            echo -e "${GREEN}✓ Found main.go${NC}"
        fi
        
        # Check for Dockerfile
        if [ ! -f "$PROJECT_ROOT/docker/$service/Dockerfile" ]; then
            echo -e "${YELLOW}⚠ Missing Dockerfile for $service${NC}"
            ((WARNINGS++))
        fi
    fi
}

echo -e "\n${YELLOW}=== Checking Required Directories ===${NC}"

# Core directories
check_required_dir "cmd" "Service entry points"
check_required_dir "pkg" "Shared packages"
check_required_dir "internal" "Internal packages"
check_required_dir "operators" "Kubernetes operators"
check_required_dir "dashboard" "Web UI"
check_required_dir "k8s" "Kubernetes manifests"
check_required_dir "helm" "Helm charts"
check_required_dir "scripts" "Build and utility scripts"
check_required_dir "docs" "Documentation"

echo -e "\n${YELLOW}=== Checking Service Structure ===${NC}"

# Check each service
for service in api controller generator simulator; do
    check_service "$service"
done

echo -e "\n${YELLOW}=== Checking Forbidden Patterns ===${NC}"

# Check for forbidden patterns
check_forbidden "*.exe" "Compiled binaries"
check_forbidden ".env" "Environment files (use .env.example)"

echo -e "\n${YELLOW}=== Checking Import Rules ===${NC}"

# Check for cross-service internal imports (simplified check)
CROSS_IMPORTS=$(grep -r "cmd/.*/internal" "$PROJECT_ROOT" --include="*.go" 2>/dev/null | grep -v "cmd/.*/.*\.go.*cmd/.*/internal" || true)
if [ -n "$CROSS_IMPORTS" ]; then
    echo -e "${RED}❌ Found cross-service internal imports:${NC}"
    echo "$CROSS_IMPORTS"
    ((ERRORS++))
else
    echo -e "${GREEN}✓ No cross-service internal imports found${NC}"
fi

echo -e "\n${YELLOW}=== Checking Documentation ===${NC}"

# Check for required documentation
REQUIRED_DOCS=(
    "docs/README.md"
    "docs/architecture.md"
    "docs/QUICK_START_GUIDE.md"
)

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ ! -f "$PROJECT_ROOT/$doc" ]; then
        echo -e "${YELLOW}⚠ Missing documentation: $doc${NC}"
        ((WARNINGS++))
    else
        echo -e "${GREEN}✓ Found: $doc${NC}"
    fi
done

# Check that no Phoenix docs are at root level
PHOENIX_DOCS_AT_ROOT=$(find "$PROJECT_ROOT" -maxdepth 1 -name "*.md" -not -name "CLAUDE.md" -not -name "README.md" | grep -E "(PHOENIX|PROJECT|IMPL|REVIEW|GUIDE)" || true)
if [ -n "$PHOENIX_DOCS_AT_ROOT" ]; then
    echo -e "${RED}❌ Phoenix documentation found at root level (should be in phoenix-platform/docs/):${NC}"
    echo "$PHOENIX_DOCS_AT_ROOT"
    ((ERRORS++))
fi

echo -e "\n${YELLOW}=== Summary ===${NC}"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}❌ Structure validation FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Structure validation PASSED${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}⚠ Please address warnings${NC}"
    fi
    exit 0
fi