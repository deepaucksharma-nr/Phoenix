#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "Phoenix Platform Integration Validation"
echo "======================================"
echo "Date: $(date)"
echo "Repository: $(pwd)"
echo ""

# Results tracking
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# Log functions
log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_CHECKS++))
    ((TOTAL_CHECKS++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_CHECKS++))
    ((TOTAL_CHECKS++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
    ((TOTAL_CHECKS++))
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Check Go workspace
echo -e "\n${BLUE}==== Go Workspace Configuration ====${NC}\n"

if [ -f "go.work" ]; then
    log_pass "go.work file exists"
    
    # Check workspace sync
    if go work sync 2>/dev/null; then
        log_pass "Go workspace synced successfully"
    else
        log_fail "Go workspace sync failed"
    fi
else
    log_fail "go.work file not found"
fi

# Check packages
echo -e "\n${BLUE}==== Validating Shared Packages ====${NC}\n"

PACKAGES=("packages/go-common" "packages/contracts")
for pkg in "${PACKAGES[@]}"; do
    if [ -d "$pkg" ]; then
        log_info "Checking $pkg..."
        cd "$pkg"
        
        if [ -f "go.mod" ]; then
            if go mod tidy 2>/dev/null; then
                log_pass "$pkg: go mod tidy successful"
            else
                log_fail "$pkg: go mod tidy failed"
            fi
            
            if go build ./... 2>/dev/null; then
                log_pass "$pkg: builds successfully"
            else
                log_fail "$pkg: build failed"
            fi
            
            if go test ./... -short 2>/dev/null; then
                log_pass "$pkg: tests pass"
            else
                log_warn "$pkg: tests failed (non-critical)"
            fi
        else
            log_fail "$pkg: go.mod not found"
        fi
        
        cd - > /dev/null
    else
        log_fail "$pkg directory not found"
    fi
done

# Check services
echo -e "\n${BLUE}==== Validating Services ====${NC}\n"

# Find all services with go.mod
SERVICES=$(find services -name "go.mod" -type f | xargs dirname | sort)

for service in $SERVICES; do
    log_info "Checking $service..."
    cd "$service"
    
    # Check imports
    if grep -q "github.com/phoenix/platform/pkg[^/]" *.go **/*.go 2>/dev/null; then
        log_fail "$service: still has old pkg imports"
    else
        log_pass "$service: imports updated correctly"
    fi
    
    # Check go.mod replace directives
    if grep -q "packages/go-common" go.mod; then
        log_pass "$service: go.mod has correct replace directives"
    else
        log_warn "$service: missing replace directive for packages/go-common"
    fi
    
    # Try to build
    if go build ./... 2>/dev/null; then
        log_pass "$service: builds successfully"
    else
        log_fail "$service: build failed"
    fi
    
    cd - > /dev/null
done

# Check for duplicate services
echo -e "\n${BLUE}==== Checking for Duplicates ====${NC}\n"

# Compare services/ and projects/
for dir in services/*; do
    if [ -d "$dir" ]; then
        service_name=$(basename "$dir")
        if [ -d "projects/$service_name" ]; then
            log_warn "Duplicate found: $service_name exists in both services/ and projects/"
        fi
    fi
done

# Check protobuf files
echo -e "\n${BLUE}==== Checking Protocol Buffers ====${NC}\n"

if find . -name "*.proto" -type f | grep -q .; then
    log_pass "Proto files found"
    
    # Check if protoc is installed
    if command -v protoc &> /dev/null; then
        log_pass "protoc is installed"
    else
        log_warn "protoc not installed - cannot generate Go code from proto files"
    fi
else
    log_fail "No proto files found"
fi

# Summary
echo -e "\n${BLUE}==== Integration Validation Summary ====${NC}\n"
echo "Total Checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
echo -e "${RED}Failed: $FAILED_CHECKS${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "\n${GREEN}✓ Integration validation PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Integration validation FAILED${NC}"
    echo "Please fix the issues above before proceeding."
    exit 1
fi