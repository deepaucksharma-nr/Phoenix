#!/bin/bash
# validate-migration.sh - Validate the migration was successful
# Usage: ./validate-migration.sh [service-name]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Validation results
ERRORS=0
WARNINGS=0

# Log functions
log_error() {
    echo -e "${RED}✗ $1${NC}"
    ((ERRORS++))
}

log_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
    ((WARNINGS++))
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

echo -e "${BLUE}=== Phoenix Migration Validation ===${NC}"
echo ""

# Check if specific service provided
SERVICE_NAME=$1
if [ -n "${SERVICE_NAME:-}" ]; then
    echo -e "${YELLOW}Validating migration for: $SERVICE_NAME${NC}"
    PROJECTS=("projects/$SERVICE_NAME")
else
    echo -e "${YELLOW}Validating all migrated services${NC}"
    PROJECTS=(projects/*)
fi

# Validate directory structure
echo -e "\n${BLUE}Checking Directory Structure${NC}"
REQUIRED_DIRS=(
    "build"
    "deployments" 
    "pkg"
    "projects"
    "scripts"
    "tests"
    "tools"
    "docs"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        log_success "Directory exists: $dir"
    else
        log_error "Missing directory: $dir"
    fi
done

# Validate shared packages
echo -e "\n${BLUE}Checking Shared Packages${NC}"
if [ -f "pkg/go.mod" ]; then
    log_success "Shared packages go.mod exists"
    
    # Check if packages compile
    cd pkg
    if go build ./... 2>/dev/null; then
        log_success "Shared packages compile successfully"
    else
        log_error "Shared packages have compilation errors"
    fi
    cd ..
else
    log_error "Missing pkg/go.mod"
fi

# Validate each project
for project_path in "${PROJECTS[@]}"; do
    if [ ! -d "$project_path" ]; then
        continue
    fi
    
    project_name=$(basename "$project_path")
    echo -e "\n${BLUE}Validating Project: $project_name${NC}"
    
    # Check required files
    REQUIRED_FILES=(
        "Makefile"
        "README.md"
        ".gitignore"
        "VERSION"
    )
    
    for file in "${REQUIRED_FILES[@]}"; do
        if [ -f "$project_path/$file" ]; then
            log_success "$file exists"
        else
            log_error "Missing $file"
        fi
    done
    
    # Check directory structure
    REQUIRED_PROJECT_DIRS=(
        "build"
        "deployments"
        "docs"
    )
    
    for dir in "${REQUIRED_PROJECT_DIRS[@]}"; do
        if [ -d "$project_path/$dir" ]; then
            log_success "Directory $dir exists"
        else
            log_warning "Missing directory: $dir"
        fi
    done
    
    # Language-specific checks
    if [ -f "$project_path/go.mod" ]; then
        echo -e "\n${YELLOW}Go Service Validation${NC}"
        
        # Check Go module name
        module_name=$(grep "^module" "$project_path/go.mod" | awk '{print $2}')
        if [[ $module_name == *"phoenix-vnext"* ]]; then
            log_success "Go module name updated correctly"
        else
            log_error "Go module name not updated: $module_name"
        fi
        
        # Check for old imports
        if grep -r "github.com/phoenix/" "$project_path" --include="*.go" 2>/dev/null | grep -v "phoenix-vnext"; then
            log_error "Found old import paths"
        else
            log_success "Import paths updated"
        fi
        
        # Check if it compiles
        cd "$project_path"
        if go build ./... 2>/dev/null; then
            log_success "Project compiles successfully"
        else
            log_warning "Project has compilation errors"
        fi
        cd - > /dev/null
        
    elif [ -f "$project_path/package.json" ]; then
        echo -e "\n${YELLOW}Node.js Service Validation${NC}"
        
        # Check package.json name
        package_name=$(jq -r .name "$project_path/package.json" 2>/dev/null)
        if [ -n "$package_name" ]; then
            log_success "package.json valid"
        else
            log_error "Invalid package.json"
        fi
        
        # Check for lock file
        if [ -f "$project_path/package-lock.json" ] || [ -f "$project_path/yarn.lock" ] || [ -f "$project_path/pnpm-lock.yaml" ]; then
            log_success "Lock file present"
        else
            log_warning "No lock file found"
        fi
    fi
    
    # Check Dockerfile
    if [ -f "$project_path/build/Dockerfile" ]; then
        log_success "Dockerfile present"
        
        # Check if Dockerfile references correct paths
        if grep -q "OLD_IMPLEMENTATION" "$project_path/build/Dockerfile"; then
            log_error "Dockerfile contains OLD_IMPLEMENTATION references"
        fi
    else
        log_warning "No Dockerfile found"
    fi
    
    # Check Kubernetes manifests
    if [ -d "$project_path/deployments/k8s" ]; then
        log_success "Kubernetes manifests present"
        
        # Validate YAML syntax
        if command -v yamllint > /dev/null; then
            if yamllint "$project_path/deployments/k8s"/*.yaml 2>/dev/null; then
                log_success "Kubernetes manifests valid"
            else
                log_warning "Kubernetes manifests have syntax issues"
            fi
        fi
    else
        log_warning "No Kubernetes manifests found"
    fi
done

# Check for remaining OLD_IMPLEMENTATION references
echo -e "\n${BLUE}Checking for OLD_IMPLEMENTATION References${NC}"
if grep -r "OLD_IMPLEMENTATION" . --exclude-dir="OLD_IMPLEMENTATION" --exclude-dir=".git" --exclude="*.sh" 2>/dev/null; then
    log_error "Found references to OLD_IMPLEMENTATION in migrated code"
else
    log_success "No OLD_IMPLEMENTATION references found"
fi

# Validate build infrastructure
echo -e "\n${BLUE}Checking Build Infrastructure${NC}"
REQUIRED_MAKEFILES=(
    "build/makefiles/common.mk"
    "build/makefiles/go.mk"
    "build/makefiles/node.mk"
    "build/makefiles/docker.mk"
)

for makefile in "${REQUIRED_MAKEFILES[@]}"; do
    if [ -f "$makefile" ]; then
        log_success "Build file exists: $makefile"
    else
        log_error "Missing build file: $makefile"
    fi
done

# Test root Makefile
if [ -f "Makefile" ]; then
    if make help > /dev/null 2>&1; then
        log_success "Root Makefile is functional"
    else
        log_error "Root Makefile has errors"
    fi
fi

# Summary
echo -e "\n${BLUE}=== Validation Summary ===${NC}"
echo -e "Errors: ${RED}$ERRORS${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"

if [ $ERRORS -eq 0 ]; then
    echo -e "\n${GREEN}✓ Migration validation passed!${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}Note: There are $WARNINGS warnings that should be addressed.${NC}"
    fi
    exit 0
else
    echo -e "\n${RED}✗ Migration validation failed with $ERRORS errors.${NC}"
    echo -e "${YELLOW}Please fix the errors before proceeding.${NC}"
    exit 1
fi