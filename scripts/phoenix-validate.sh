#!/bin/bash
# Phoenix validation script - runs all validation checks

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔍 Phoenix Validation Suite"
echo "==========================="
echo ""

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
FAILURES=0

# Parse arguments
QUICK=false
VERBOSE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --quick)
            QUICK=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--quick] [--verbose]"
            exit 1
            ;;
    esac
done

# Function to run check
run_check() {
    local name=$1
    local cmd=$2
    
    echo -n "Checking $name... "
    
    if [ "$VERBOSE" = true ]; then
        echo ""
        if eval "$cmd"; then
            echo -e "${GREEN}✓ PASS${NC}"
        else
            echo -e "${RED}✗ FAIL${NC}"
            ((FAILURES++))
        fi
    else
        if output=$(eval "$cmd" 2>&1); then
            echo -e "${GREEN}✓ PASS${NC}"
        else
            echo -e "${RED}✗ FAIL${NC}"
            if [ -n "$output" ]; then
                echo "  Error: $(echo "$output" | head -1)"
            fi
            ((FAILURES++))
        fi
    fi
}

# 1. Code Structure Validation
echo -e "${BLUE}Code Structure:${NC}"
echo "---------------"

# Check for cross-project imports
run_check "cross-project imports" "
    ! grep -r --include='*.go' 'github.com/phoenix/platform/projects' $PROJECT_ROOT/projects 2>/dev/null | \
    grep -v '// Code generated' | \
    awk -F: '{print \$1}' | while read file; do
        dir=\$(dirname \"\$file\" | cut -d'/' -f1-4)
        if grep -q \"github.com/phoenix/platform/projects\" \"\$file\" | grep -v \"\$dir\"; then
            echo \"Cross-project import in \$file\"
            exit 1
        fi
    done
"

# Check for direct database imports
run_check "direct database access" "
    ! grep -r --include='*.go' -E '\"database/sql\"|\"github.com/lib/pq\"|\"github.com/jackc/pgx\"' \
    $PROJECT_ROOT/projects/*/internal $PROJECT_ROOT/projects/*/cmd 2>/dev/null | \
    grep -v '// Code generated'
"

# Check for hardcoded secrets
run_check "hardcoded secrets" "
    ! grep -r --include='*.go' --include='*.yaml' --include='*.yml' -E \
    '(password|secret|key|token)\\s*[:=]\\s*[\"'\''][^\"'\'']+[\"'\'']' \
    $PROJECT_ROOT 2>/dev/null | \
    grep -vE '(example|template|test|fake|dummy|TODO|CHANGE_ME)'
"

echo ""

# 2. Go Module Validation
echo -e "${BLUE}Go Modules:${NC}"
echo "-----------"

# Check go.work consistency
run_check "go.work consistency" "
    cd $PROJECT_ROOT && go work sync && \
    ! git diff --quiet go.work go.work.sum
"

# Check module dependencies
if [ "$QUICK" = false ]; then
    run_check "module dependencies" "
        cd $PROJECT_ROOT && \
        find projects -name go.mod -type f | while read mod; do
            dir=\$(dirname \"\$mod\")
            (cd \"\$dir\" && go mod verify) || exit 1
        done
    "
fi

echo ""

# 3. Build Validation
echo -e "${BLUE}Build Validation:${NC}"
echo "-----------------"

# Check if all services build
if [ "$QUICK" = false ]; then
    for service in phoenix-api phoenix-agent phoenix-cli; do
        if [ -d "$PROJECT_ROOT/projects/$service" ]; then
            run_check "$service build" "
                cd $PROJECT_ROOT/projects/$service && \
                go build ./... 2>&1 | grep -v 'no non-test Go files'
            "
        fi
    done
else
    echo "Skipping build checks (--quick mode)"
fi

echo ""

# 4. Security Validation
echo -e "${BLUE}Security Checks:${NC}"
echo "----------------"

# Check file permissions on sensitive files
run_check "sensitive file permissions" "
    ! find $PROJECT_ROOT -name '*.key' -o -name '*.pem' -o -name '.env*' | \
    xargs -I {} stat -c '%a %n' {} 2>/dev/null | \
    grep -v '^600' | grep -v '^400'
"

# Check for TODO security items
run_check "security TODOs" "
    ! grep -r --include='*.go' 'TODO.*security' $PROJECT_ROOT 2>/dev/null
"

echo ""

# 5. Configuration Validation
echo -e "${BLUE}Configuration:${NC}"
echo "--------------"

# Check for required config files
run_check "docker-compose.yml" "[ -f $PROJECT_ROOT/docker-compose.yml ]"
run_check "Makefile" "[ -f $PROJECT_ROOT/Makefile ]"

# Check for .env templates
run_check ".env templates" "
    [ -f $PROJECT_ROOT/projects/phoenix-api/.env.template ] || \
    [ -f $PROJECT_ROOT/projects/phoenix-api/.env ]
"

echo ""

# 6. Documentation Validation
if [ "$QUICK" = false ]; then
    echo -e "${BLUE}Documentation:${NC}"
    echo "--------------"
    
    # Check for README files
    run_check "main README" "[ -f $PROJECT_ROOT/README.md ]"
    run_check "API README" "[ -f $PROJECT_ROOT/projects/phoenix-api/README.md ]"
    
    # Check for API documentation
    run_check "API docs" "
        [ -f $PROJECT_ROOT/docs/api/README.md ] || \
        [ -f $PROJECT_ROOT/contracts/openapi/control-api.yaml ]
    "
    
    echo ""
fi

# 7. Git Validation
echo -e "${BLUE}Git Checks:${NC}"
echo "-----------"

# Check for large files
run_check "large files" "
    ! find $PROJECT_ROOT -type f -size +5M | grep -v '.git' | grep -v 'vendor' | grep -v 'node_modules'
"

# Check for merge conflicts
run_check "merge conflicts" "
    ! grep -r '<<<<<<<' $PROJECT_ROOT --exclude-dir=.git 2>/dev/null
"

echo ""

# Summary
echo -e "${BLUE}Summary:${NC}"
echo "--------"
if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✅ All validation checks passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ $FAILURES validation check(s) failed${NC}"
    echo ""
    echo "Please fix the issues before committing."
    exit 1
fi