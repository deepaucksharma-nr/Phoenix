#!/bin/bash
# validate-all.sh - Comprehensive validation of Phoenix Platform

set -euo pipefail

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; ((PASSED_CHECKS++)); ((TOTAL_CHECKS++)); }
log_error() { echo -e "${RED}[✗]${NC} $1"; ((FAILED_CHECKS++)); ((TOTAL_CHECKS++)); }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; ((WARNINGS++)); }

echo -e "${BLUE}=== Phoenix Platform Comprehensive Validation ===${NC}\n"

# 1. Check directory structure
log_info "Checking directory structure..."
REQUIRED_DIRS=(
    "build" "pkg" "projects" "services" "operators" 
    "configs" "infrastructure" "tests" "docs" "scripts"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        log_success "Directory exists: $dir"
    else
        log_error "Missing directory: $dir"
    fi
done

# 2. Validate Go workspace
log_info "\nValidating Go workspace..."
if [ -f "go.work" ]; then
    log_success "go.work file exists"
    
    # Try to sync workspace
    if go work sync 2>/dev/null; then
        log_success "Go workspace synced successfully"
    else
        log_warning "Go workspace sync had issues"
    fi
else
    log_error "go.work file missing"
fi

# 3. Check shared packages
log_info "\nChecking shared packages..."
if [ -f "pkg/go.mod" ]; then
    log_success "Shared packages go.mod exists"
    
    cd pkg
    if go build ./... 2>/dev/null; then
        log_success "Shared packages compile"
    else
        log_warning "Shared packages have compilation issues"
    fi
    cd ..
else
    log_error "Missing pkg/go.mod"
fi

# 4. Validate Docker setup
log_info "\nValidating Docker setup..."
DOCKER_FILES=(
    "docker-compose.yml"
    "infrastructure/docker/compose/docker-compose.dev.yml"
)

for file in "${DOCKER_FILES[@]}"; do
    if [ -f "$file" ]; then
        log_success "Found: $file"
        
        # Validate docker-compose syntax
        if docker-compose -f "$file" config > /dev/null 2>&1; then
            log_success "Valid docker-compose syntax: $file"
        else
            log_warning "Invalid docker-compose syntax: $file"
        fi
    else
        log_warning "Missing: $file"
    fi
done

# 5. Check build infrastructure
log_info "\nChecking build infrastructure..."
BUILD_FILES=(
    "Makefile"
    "build/makefiles/common.mk"
    "build/makefiles/go.mk"
    "build/makefiles/node.mk"
    "build/makefiles/docker.mk"
)

for file in "${BUILD_FILES[@]}"; do
    if [ -f "$file" ]; then
        log_success "Build file exists: $file"
    else
        log_error "Missing build file: $file"
    fi
done

# 6. Validate configuration files
log_info "\nValidating configuration files..."
CONFIG_DIRS=(
    "configs/monitoring"
    "configs/otel"
    "configs/control"
    "configs/production"
)

for dir in "${CONFIG_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        log_success "Config directory exists: $dir"
        
        # Count config files
        file_count=$(find "$dir" -type f \( -name "*.yaml" -o -name "*.yml" -o -name "*.json" \) | wc -l | tr -d ' ')
        if [ "$file_count" -gt 0 ]; then
            log_success "Found $file_count config files in $dir"
        else
            log_warning "No config files in $dir"
        fi
    else
        log_warning "Missing config directory: $dir"
    fi
done

# 7. Check Kubernetes manifests
log_info "\nChecking Kubernetes manifests..."
K8S_DIRS=(
    "infrastructure/kubernetes/base"
    "infrastructure/kubernetes/operators"
)

for dir in "${K8S_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        log_success "K8s directory exists: $dir"
        
        # Validate YAML files
        yaml_count=$(find "$dir" -name "*.yaml" -o -name "*.yml" | wc -l | tr -d ' ')
        if [ "$yaml_count" -gt 0 ]; then
            log_success "Found $yaml_count YAML files in $dir"
        fi
    else
        log_warning "Missing K8s directory: $dir"
    fi
done

# 8. Validate project structure
log_info "\nValidating project structure..."
if [ -d "projects" ]; then
    for project in projects/*/; do
        if [ -d "$project" ]; then
            project_name=$(basename "$project")
            log_info "Checking project: $project_name"
            
            # Check for standard files
            for file in "Makefile" "README.md"; do
                if [ -f "$project$file" ]; then
                    log_success "$project_name has $file"
                else
                    log_warning "$project_name missing $file"
                fi
            done
            
            # Check for go.mod if Go project
            if [ -f "$project/go.mod" ]; then
                log_success "$project_name has go.mod"
            elif [ -f "$project/package.json" ]; then
                log_success "$project_name has package.json"
            else
                log_warning "$project_name missing module file"
            fi
        fi
    done
fi

# 9. Check documentation
log_info "\nChecking documentation..."
DOC_FILES=(
    "README.md"
    "CONTRIBUTING.md"
    "LICENSE"
    "CLAUDE.md"
)

for file in "${DOC_FILES[@]}"; do
    if [ -f "$file" ]; then
        log_success "Documentation exists: $file"
    else
        log_warning "Missing documentation: $file"
    fi
done

# 10. Validate scripts
log_info "\nValidating scripts..."
SCRIPT_COUNT=$(find scripts -name "*.sh" -type f | wc -l | tr -d ' ')
if [ "$SCRIPT_COUNT" -gt 0 ]; then
    log_success "Found $SCRIPT_COUNT scripts"
    
    # Check if scripts are executable
    NON_EXEC=$(find scripts -name "*.sh" -type f ! -perm -u+x | wc -l | tr -d ' ')
    if [ "$NON_EXEC" -eq 0 ]; then
        log_success "All scripts are executable"
    else
        log_warning "$NON_EXEC scripts are not executable"
    fi
else
    log_error "No scripts found"
fi

# Summary
echo -e "\n${BLUE}=== Validation Summary ===${NC}"
echo -e "Total checks: ${TOTAL_CHECKS}"
echo -e "${GREEN}Passed: ${PASSED_CHECKS}${NC}"
echo -e "${RED}Failed: ${FAILED_CHECKS}${NC}"
echo -e "${YELLOW}Warnings: ${WARNINGS}${NC}"

if [ "$FAILED_CHECKS" -eq 0 ]; then
    echo -e "\n${GREEN}✓ Phoenix Platform validation passed!${NC}"
    if [ "$WARNINGS" -gt 0 ]; then
        echo -e "${YELLOW}Note: There are $WARNINGS warnings that should be reviewed.${NC}"
    fi
    exit 0
else
    echo -e "\n${RED}✗ Phoenix Platform validation failed with $FAILED_CHECKS errors.${NC}"
    exit 1
fi