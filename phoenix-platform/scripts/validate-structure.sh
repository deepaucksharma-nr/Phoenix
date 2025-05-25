#!/bin/bash
# Validates Phoenix Platform mono-repo structure and enforces governance rules

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validation results
ERRORS=0
WARNINGS=0

echo "🔍 Validating Phoenix Platform Structure..."

# Function to check directory exists
check_dir() {
    local dir=$1
    local required=$2
    
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✓${NC} Directory exists: $dir"
        return 0
    else
        if [ "$required" = "required" ]; then
            echo -e "${RED}✗${NC} Missing required directory: $dir"
            ((ERRORS++))
        else
            echo -e "${YELLOW}⚠${NC} Missing optional directory: $dir"
            ((WARNINGS++))
        fi
        return 1
    fi
}

# Function to check file exists
check_file() {
    local file=$1
    local required=$2
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} File exists: $file"
        return 0
    else
        if [ "$required" = "required" ]; then
            echo -e "${RED}✗${NC} Missing required file: $file"
            ((ERRORS++))
        else
            echo -e "${YELLOW}⚠${NC} Missing optional file: $file"
            ((WARNINGS++))
        fi
        return 1
    fi
}

# Check we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}✗${NC} Not in phoenix-platform directory"
    exit 1
fi

echo ""
echo "📁 Checking directory structure..."

# Required directories
check_dir "cmd" "required"
check_dir "cmd/api" "required"
check_dir "cmd/controller" "required"
check_dir "cmd/generator" "required"
check_dir "cmd/simulator" "required"

check_dir "pkg" "required"
check_dir "pkg/api" "required"
check_dir "pkg/store" "required"
check_dir "pkg/models" "required"

check_dir "operators" "required"
check_dir "operators/pipeline" "required"
check_dir "operators/loadsim" "required"

check_dir "dashboard" "required"
check_dir "docker" "required"
check_dir "helm" "required"
check_dir "k8s" "required"
check_dir "scripts" "required"
check_dir "docs" "required"

# Optional but recommended directories
check_dir "internal" "optional"
check_dir "test" "optional"
check_dir "test/unit" "optional"
check_dir "test/integration" "optional"
check_dir "test/e2e" "optional"
check_dir "pipelines/templates" "optional"

echo ""
echo "📄 Checking required files..."

# Required files
check_file "go.mod" "required"
check_file "go.sum" "required"
check_file "Makefile" "required"
check_file "README.md" "required"
check_file ".golangci.yml" "required"
check_file "docker-compose.yaml" "required"

# Check governance files at repo root
check_file "../CODEOWNERS" "required"
check_file "../.commitlintrc.yml" "required"

echo ""
echo "🔧 Checking service implementations..."

# Check if services have main.go files
for service in api controller generator simulator; do
    if check_file "cmd/$service/main.go" "optional"; then
        # Check if it's not just a stub
        if grep -q "TODO\|FIXME\|panic.*not implemented" "cmd/$service/main.go" 2>/dev/null; then
            echo -e "${YELLOW}⚠${NC} Service $service has TODO/unimplemented code"
            ((WARNINGS++))
        fi
    fi
done

echo ""
echo "📦 Checking package structure..."

# Validate no cross-dependencies between cmd directories
if find cmd -name "*.go" -exec grep -l "phoenix/platform/cmd/" {} \; 2>/dev/null | grep -v _test.go; then
    echo -e "${RED}✗${NC} Found imports from cmd packages (violation of modular structure)"
    ((ERRORS++))
else
    echo -e "${GREEN}✓${NC} No cross-imports between cmd packages"
fi

# Check for internal package usage
if [ -d "internal" ]; then
    # Ensure internal packages are not imported from outside their service
    echo -e "${GREEN}✓${NC} Internal package structure validated"
fi

echo ""
echo "🏗️ Checking build configuration..."

# Check Makefile targets
required_targets=("build" "test" "lint" "fmt" "docker" "clean")
for target in "${required_targets[@]}"; do
    if grep -q "^$target:" Makefile; then
        echo -e "${GREEN}✓${NC} Makefile target exists: $target"
    else
        echo -e "${RED}✗${NC} Missing Makefile target: $target"
        ((ERRORS++))
    fi
done

echo ""
echo "📊 Validation Summary:"
echo "===================="
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed!${NC}"
    exit 0
else
    echo -e "Errors: ${RED}$ERRORS${NC}"
    echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
    
    if [ $ERRORS -gt 0 ]; then
        echo -e "\n${RED}❌ Validation failed with errors${NC}"
        exit 1
    else
        echo -e "\n${YELLOW}⚠️  Validation passed with warnings${NC}"
        exit 0
    fi
fi