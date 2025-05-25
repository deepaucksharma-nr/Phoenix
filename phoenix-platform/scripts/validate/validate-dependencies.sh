#!/bin/bash
# validate-dependencies.sh - Check for allowed/forbidden dependencies

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔍 Validating Phoenix Platform Dependencies..."

ERRORS=0
WARNINGS=0

# Allowed production dependencies
ALLOWED_DEPS=(
    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    "github.com/lib/pq"
    "github.com/go-redis/redis"
    "k8s.io/client-go"
    "github.com/prometheus/client_golang"
    "go.uber.org/zap"
    "github.com/spf13/viper"
    "sigs.k8s.io/controller-runtime"
    "github.com/golang-jwt/jwt"
    "github.com/gorilla/mux"
    "github.com/golang-migrate/migrate"
)

# Forbidden dependencies
FORBIDDEN_DEPS=(
    "github.com/pkg/errors"     # Use stdlib errors
    "github.com/sirupsen/logrus" # Use zap
    "gopkg.in/*"                # Avoid v1 style imports
    "github.com/jinzhu/gorm"    # Use database/sql
)

# Check if go.mod exists
if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
    echo -e "${RED}❌ go.mod not found in project root${NC}"
    exit 1
fi

echo -e "\n${YELLOW}=== Checking Forbidden Dependencies ===${NC}"

# Check for forbidden dependencies
for dep in "${FORBIDDEN_DEPS[@]}"; do
    if grep -q "$dep" "$PROJECT_ROOT/go.mod"; then
        echo -e "${RED}❌ Forbidden dependency found: $dep${NC}"
        ((ERRORS++))
    fi
done

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✓ No forbidden dependencies found${NC}"
fi

echo -e "\n${YELLOW}=== Checking Dependency Versions ===${NC}"

# Check for multiple versions of same dependency
duplicates=$(go mod graph | cut -d '@' -f 1 | sort | uniq -d | grep -v "^$" || true)
if [ -n "$duplicates" ]; then
    echo -e "${YELLOW}⚠️  Multiple versions of dependencies found:${NC}"
    echo "$duplicates"
    ((WARNINGS++))
else
    echo -e "${GREEN}✓ No duplicate dependency versions${NC}"
fi

echo -e "\n${YELLOW}=== Checking Indirect Dependencies ===${NC}"

# Count indirect dependencies
indirect_count=$(grep -c "// indirect" "$PROJECT_ROOT/go.mod" || echo "0")
if [ "$indirect_count" -gt 50 ]; then
    echo -e "${YELLOW}⚠️  High number of indirect dependencies: $indirect_count${NC}"
    echo "Consider upgrading direct dependencies to reduce indirect ones"
    ((WARNINGS++))
else
    echo -e "${GREEN}✓ Indirect dependencies: $indirect_count (acceptable)${NC}"
fi

echo -e "\n${YELLOW}=== Checking Security Vulnerabilities ===${NC}"

# Run go mod audit if available
if command -v gosec &> /dev/null; then
    echo "Running security check..."
    if ! gosec -quiet -fmt json ./... > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Security issues found. Run 'gosec ./...' for details${NC}"
        ((WARNINGS++))
    else
        echo -e "${GREEN}✓ No security issues found${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest${NC}"
    ((WARNINGS++))
fi

echo -e "\n${YELLOW}=== Checking License Compatibility ===${NC}"

# Simple license check for common problematic licenses
problematic_licenses=(
    "GPL"
    "AGPL"
    "LGPL"
)

echo "Checking for problematic licenses..."
license_issues=0
for license in "${problematic_licenses[@]}"; do
    if go mod graph | grep -i "$license" > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Potential $license licensed dependency found${NC}"
        ((license_issues++))
    fi
done

if [ $license_issues -eq 0 ]; then
    echo -e "${GREEN}✓ No obviously problematic licenses found${NC}"
else
    echo "Run a proper license audit tool for comprehensive checking"
    ((WARNINGS++))
fi

echo -e "\n${YELLOW}=== Checking go.mod Tidiness ===${NC}"

# Check if go.mod is tidy
if ! go mod tidy -v 2>&1 | grep -q "go.mod file is being rewritten"; then
    echo -e "${GREEN}✓ go.mod is tidy${NC}"
else
    echo -e "${YELLOW}⚠️  go.mod needs tidying. Run 'go mod tidy'${NC}"
    ((WARNINGS++))
fi

echo -e "\n${YELLOW}=== Summary ===${NC}"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}❌ Dependency validation FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Dependency validation PASSED${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}⚠️  Please review warnings${NC}"
    fi
    exit 0
fi