#!/usr/bin/env bash

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Counters
ERRORS=0
WARNINGS=0

# Functions
error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((ERRORS++))
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    ((WARNINGS++))
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

echo "=== Phoenix Platform Boundary Check ==="
echo

# Check if projects directory exists
if [ ! -d "projects" ]; then
    error "Projects directory not found"
    exit 1
fi

# Check for cross-project imports in Go files
echo "Checking Go project boundaries..."
for project in projects/*/; do
    if [ -d "$project" ] && [ -f "$project/go.mod" ]; then
        project_name=$(basename "$project")
        
        # Skip if no Go files
        if ! find "$project" -name "*.go" -type f | grep -q .; then
            continue
        fi
        
        echo "  Checking $project_name..."
        
        # Check for imports from other projects
        if grep -r "github.com/phoenix/platform/projects/" "$project" \
            --include="*.go" \
            --exclude-dir="vendor" \
            --exclude-dir=".git" | \
            grep -v "github.com/phoenix/platform/projects/$project_name"; then
            error "Cross-project import detected in $project_name"
            echo "    Projects must not import from each other directly"
            echo "    Use shared packages in /pkg instead"
        else
            success "No cross-project imports in $project_name"
        fi
        
        # Check for forbidden imports
        forbidden_imports=(
            "database/sql"
            "github.com/lib/pq"
            "github.com/go-sql-driver/mysql"
            "go.mongodb.org/mongo-driver"
        )
        
        for import in "${forbidden_imports[@]}"; do
            if grep -r "\"$import\"" "$project" \
                --include="*.go" \
                --exclude-dir="vendor" \
                --exclude-dir=".git" | \
                grep -v "_test.go"; then
                warning "Direct database driver import in $project_name"
                echo "    Use pkg/database abstractions instead"
            fi
        done
    fi
done

# Check for hardcoded secrets
echo
echo "Checking for hardcoded secrets..."
secret_patterns=(
    'password\s*[:=]\s*"[^"]*"'
    'secret\s*[:=]\s*"[^"]*"'
    'api[_-]?key\s*[:=]\s*"[^"]*"'
    'token\s*[:=]\s*"[^"]*"'
    'AWS_ACCESS_KEY_ID\s*[:=]\s*"[^"]*"'
    'AWS_SECRET_ACCESS_KEY\s*[:=]\s*"[^"]*"'
)

for pattern in "${secret_patterns[@]}"; do
    if grep -r -i "$pattern" . \
        --include="*.go" \
        --include="*.js" \
        --include="*.ts" \
        --include="*.py" \
        --include="*.yaml" \
        --include="*.yml" \
        --include="*.json" \
        --exclude-dir="vendor" \
        --exclude-dir="node_modules" \
        --exclude-dir=".git" \
        --exclude-dir="OLD_IMPLEMENTATION" \
        --exclude="*.example" \
        --exclude="*.test.*" \
        --exclude="*_test.go" | \
        grep -v "// Example:" | \
        grep -v "# Example:"; then
        error "Potential hardcoded secret detected"
    fi
done

# Check for production configuration in wrong places
echo
echo "Checking production configuration placement..."
if grep -r "production\|prod\." . \
    --include="*.go" \
    --include="*.yaml" \
    --include="*.yml" \
    --include="*.json" \
    --exclude-dir="vendor" \
    --exclude-dir="node_modules" \
    --exclude-dir=".git" \
    --exclude-dir="deployments/kubernetes/overlays/production" \
    --exclude-dir="docs" \
    --exclude-dir="OLD_IMPLEMENTATION" \
    --exclude=".github/workflows/*" | \
    grep -v "// " | \
    grep -v "# "; then
    warning "Production configuration found outside of designated areas"
fi

# Check shared package usage
echo
echo "Checking shared package usage..."
if [ -d "pkg" ]; then
    # Ensure pkg has its own go.mod
    if [ ! -f "pkg/go.mod" ]; then
        error "pkg directory missing go.mod file"
    else
        success "Shared packages properly configured"
    fi
    
    # Check if projects are using shared packages
    shared_packages=$(find pkg -type d -name "*" -maxdepth 2 | grep -v "^\.$" | wc -l)
    if [ "$shared_packages" -gt 0 ]; then
        projects_using_shared=0
        for project in projects/*/; do
            if [ -f "$project/go.mod" ]; then
                if grep -q "github.com/phoenix/platform/pkg" "$project/go.mod"; then
                    ((projects_using_shared++))
                fi
            fi
        done
        
        if [ "$projects_using_shared" -eq 0 ] && [ "$(ls -A projects 2>/dev/null)" ]; then
            warning "No projects are using shared packages"
        fi
    fi
fi

# Check for circular dependencies
echo
echo "Checking for circular dependencies..."
# This is a simplified check - in production, use more sophisticated tools
for project in projects/*/; do
    if [ -f "$project/go.mod" ]; then
        project_name=$(basename "$project")
        module_name=$(grep "^module" "$project/go.mod" | awk '{print $2}')
        
        # Check if any dependency depends back on this project
        for dep_project in projects/*/; do
            if [ "$dep_project" != "$project" ] && [ -f "$dep_project/go.mod" ]; then
                if grep -q "$module_name" "$dep_project/go.mod"; then
                    error "Potential circular dependency: $(basename $dep_project) -> $project_name"
                fi
            fi
        done
    fi
done

# Check file permissions for sensitive files
echo
echo "Checking file permissions..."
sensitive_files=(
    "*.pem"
    "*.key"
    "*.crt"
    ".env"
    "secrets.yaml"
)

for pattern in "${sensitive_files[@]}"; do
    while IFS= read -r -d '' file; do
        perms=$(stat -c "%a" "$file" 2>/dev/null || stat -f "%p" "$file" 2>/dev/null | tail -c 4)
        if [ "$perms" != "600" ] && [ "$perms" != "400" ]; then
            warning "Insecure permissions on $file (current: $perms, expected: 600 or 400)"
        fi
    done < <(find . -name "$pattern" -type f -print0 2>/dev/null)
done

# Summary
echo
echo "=== Boundary Check Summary ==="
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All boundary checks passed!${NC}"
    exit 0
else
    echo -e "Errors: ${RED}$ERRORS${NC}"
    echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
    
    if [ $ERRORS -gt 0 ]; then
        echo -e "${RED}✗ Boundary check failed${NC}"
        exit 1
    else
        echo -e "${YELLOW}⚠ Boundary check passed with warnings${NC}"
        exit 0
    fi
fi