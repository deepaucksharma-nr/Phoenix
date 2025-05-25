#!/bin/bash

# Service validation script - enforces service boundaries and dependencies
# Part of Phoenix Platform architectural governance

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVICES_DIR="$PROJECT_ROOT/services"
PKG_DIR="$PROJECT_ROOT/pkg"

# Allowed service dependencies (service -> allowed imports)
declare -A ALLOWED_DEPS=(
    ["experiment-controller"]="pkg/logger pkg/config pkg/db pkg/api/types pkg/clients"
    ["config-generator"]="pkg/logger pkg/config pkg/templates pkg/api/types"
    ["metrics-collector"]="pkg/logger pkg/config pkg/telemetry pkg/api/types"
    ["pipeline-validator"]="pkg/logger pkg/config pkg/validation pkg/api/types"
    ["cost-analyzer"]="pkg/logger pkg/config pkg/calculator pkg/api/types"
)

# Service port ranges (for conflict detection)
declare -A SERVICE_PORTS=(
    ["experiment-controller"]="8080,9090"
    ["config-generator"]="8081,9091"
    ["metrics-collector"]="8082,9092"
    ["pipeline-validator"]="8083,9093"
    ["cost-analyzer"]="8084,9094"
)

# Required files for each service
REQUIRED_FILES=(
    "main.go"
    "Dockerfile"
    "Makefile"
    "README.md"
    "internal/"
    "cmd/"
)

echo "🔍 Validating Phoenix Platform Services..."
echo "========================================="

VALIDATION_FAILED=false

# Function to check if a service exists
service_exists() {
    local service=$1
    [[ -d "$SERVICES_DIR/$service" ]]
}

# Function to validate service structure
validate_service_structure() {
    local service=$1
    local service_path="$SERVICES_DIR/$service"
    
    echo -e "\n📦 Validating structure for $service..."
    
    for file in "${REQUIRED_FILES[@]}"; do
        if [[ ! -e "$service_path/$file" ]]; then
            echo -e "${RED}✗ Missing required file/directory: $file${NC}"
            VALIDATION_FAILED=true
        else
            echo -e "${GREEN}✓ Found $file${NC}"
        fi
    done
    
    # Check for forbidden directories
    if [[ -d "$service_path/vendor" ]]; then
        echo -e "${YELLOW}⚠ Found vendor directory (should use go modules)${NC}"
    fi
    
    # Check go.mod exists and is valid
    if [[ -f "$service_path/go.mod" ]]; then
        local module_name=$(grep "^module" "$service_path/go.mod" | awk '{print $2}')
        local expected_module="github.com/phoenix-platform/services/$service"
        
        if [[ "$module_name" != "$expected_module" ]]; then
            echo -e "${RED}✗ Invalid module name: $module_name (expected: $expected_module)${NC}"
            VALIDATION_FAILED=true
        else
            echo -e "${GREEN}✓ Valid go.mod module name${NC}"
        fi
    else
        echo -e "${RED}✗ Missing go.mod file${NC}"
        VALIDATION_FAILED=true
    fi
}

# Function to validate service dependencies
validate_service_dependencies() {
    local service=$1
    local service_path="$SERVICES_DIR/$service"
    
    echo -e "\n🔗 Validating dependencies for $service..."
    
    # Get all Go imports from the service
    local imports=$(find "$service_path" -name "*.go" -exec grep -h "^import\|^\s*\"" {} \; | \
                    grep -E "\"github.com/phoenix-platform" | \
                    sed 's/.*"\(.*\)".*/\1/' | \
                    sort -u)
    
    local allowed="${ALLOWED_DEPS[$service]}"
    local has_violations=false
    
    while IFS= read -r import; do
        if [[ -z "$import" ]]; then
            continue
        fi
        
        # Extract the package category (e.g., "pkg/logger" from full import path)
        local pkg_category=$(echo "$import" | sed 's|github.com/phoenix-platform/||' | cut -d'/' -f1-2)
        
        # Check if it's importing from another service (forbidden)
        if [[ "$pkg_category" =~ ^services/ ]]; then
            local imported_service=$(echo "$pkg_category" | cut -d'/' -f2)
            if [[ "$imported_service" != "$service" ]]; then
                echo -e "${RED}✗ Forbidden cross-service import: $import${NC}"
                VALIDATION_FAILED=true
                has_violations=true
            fi
        # Check if it's an allowed pkg import
        elif [[ "$pkg_category" =~ ^pkg/ ]]; then
            if [[ ! " $allowed " =~ " $pkg_category " ]]; then
                echo -e "${RED}✗ Unauthorized package import: $import${NC}"
                VALIDATION_FAILED=true
                has_violations=true
            fi
        fi
    done <<< "$imports"
    
    if [[ "$has_violations" == "false" ]]; then
        echo -e "${GREEN}✓ All dependencies are valid${NC}"
    fi
}

# Function to validate service configuration
validate_service_config() {
    local service=$1
    local service_path="$SERVICES_DIR/$service"
    
    echo -e "\n⚙️  Validating configuration for $service..."
    
    # Check for hardcoded configuration
    local hardcoded=$(grep -r --include="*.go" -E "(localhost:|127\.0\.0\.1:|hardcoded|TODO|FIXME)" "$service_path" || true)
    
    if [[ -n "$hardcoded" ]]; then
        echo -e "${YELLOW}⚠ Found potential hardcoded values or TODOs:${NC}"
        echo "$hardcoded" | head -5
        echo -e "${YELLOW}  (showing first 5 matches)${NC}"
    fi
    
    # Check for environment variable usage
    local env_usage=$(grep -r --include="*.go" -E "os\.(Getenv|LookupEnv)" "$service_path" || true)
    
    if [[ -z "$env_usage" ]]; then
        echo -e "${YELLOW}⚠ No environment variable usage detected (might be using hardcoded values)${NC}"
    else
        echo -e "${GREEN}✓ Environment variables are being used${NC}"
    fi
}

# Function to validate API definitions
validate_service_api() {
    local service=$1
    local service_path="$SERVICES_DIR/$service"
    
    echo -e "\n🌐 Validating API definitions for $service..."
    
    # Check for OpenAPI spec
    if [[ -f "$service_path/api/openapi.yaml" ]] || [[ -f "$service_path/api/openapi.yml" ]]; then
        echo -e "${GREEN}✓ Found OpenAPI specification${NC}"
    else
        echo -e "${YELLOW}⚠ No OpenAPI specification found${NC}"
    fi
    
    # Check for proto definitions
    if ls "$service_path"/api/*.proto >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Found proto definitions${NC}"
        
        # Validate proto package names
        for proto in "$service_path"/api/*.proto; do
            local package_name=$(grep "^package" "$proto" | awk '{print $2}' | tr -d ';')
            local expected_prefix="phoenix.${service//-/_}.v1"
            
            if [[ ! "$package_name" =~ ^$expected_prefix ]]; then
                echo -e "${RED}✗ Invalid proto package name in $(basename "$proto"): $package_name${NC}"
                VALIDATION_FAILED=true
            fi
        done
    else
        echo -e "${YELLOW}⚠ No proto definitions found${NC}"
    fi
}

# Function to check for port conflicts
check_port_conflicts() {
    echo -e "\n🔌 Checking for port conflicts..."
    
    declare -A port_usage
    
    for service in "${!SERVICE_PORTS[@]}"; do
        IFS=',' read -ra ports <<< "${SERVICE_PORTS[$service]}"
        for port in "${ports[@]}"; do
            if [[ -n "${port_usage[$port]}" ]]; then
                echo -e "${RED}✗ Port conflict: $port used by both $service and ${port_usage[$port]}${NC}"
                VALIDATION_FAILED=true
            else
                port_usage[$port]=$service
            fi
        done
    done
    
    if [[ "$VALIDATION_FAILED" == "false" ]]; then
        echo -e "${GREEN}✓ No port conflicts detected${NC}"
    fi
}

# Function to validate service documentation
validate_service_docs() {
    local service=$1
    local service_path="$SERVICES_DIR/$service"
    
    echo -e "\n📚 Validating documentation for $service..."
    
    if [[ -f "$service_path/README.md" ]]; then
        # Check README has minimum required sections
        local required_sections=("Overview" "API" "Configuration" "Development" "Testing")
        local missing_sections=()
        
        for section in "${required_sections[@]}"; do
            if ! grep -q "^#.*$section" "$service_path/README.md"; then
                missing_sections+=("$section")
            fi
        done
        
        if [[ ${#missing_sections[@]} -eq 0 ]]; then
            echo -e "${GREEN}✓ README contains all required sections${NC}"
        else
            echo -e "${YELLOW}⚠ README missing sections: ${missing_sections[*]}${NC}"
        fi
    fi
}

# Main validation loop
echo -e "\n🏗️  Discovering services..."
services=()
for dir in "$SERVICES_DIR"/*; do
    if [[ -d "$dir" ]] && [[ -f "$dir/go.mod" ]]; then
        service_name=$(basename "$dir")
        services+=("$service_name")
        echo -e "  Found service: $service_name"
    fi
done

if [[ ${#services[@]} -eq 0 ]]; then
    echo -e "${RED}✗ No services found in $SERVICES_DIR${NC}"
    exit 1
fi

# Validate each service
for service in "${services[@]}"; do
    echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Validating Service: $service${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    validate_service_structure "$service"
    validate_service_dependencies "$service"
    validate_service_config "$service"
    validate_service_api "$service"
    validate_service_docs "$service"
done

# Check for global issues
echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Global Validations${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

check_port_conflicts

# Summary
echo -e "\n========================================="
if [[ "$VALIDATION_FAILED" == "true" ]]; then
    echo -e "${RED}❌ Validation FAILED${NC}"
    echo -e "${RED}Please fix the issues above before proceeding.${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All validations PASSED${NC}"
    echo -e "${GREEN}Services comply with Phoenix Platform architecture.${NC}"
fi

# Generate validation report
REPORT_FILE="$PROJECT_ROOT/validation-report-$(date +%Y%m%d-%H%M%S).txt"
{
    echo "Phoenix Platform Service Validation Report"
    echo "Generated: $(date)"
    echo "Services Validated: ${services[*]}"
    echo ""
    echo "Status: $([[ "$VALIDATION_FAILED" == "true" ]] && echo "FAILED" || echo "PASSED")"
} > "$REPORT_FILE"

echo -e "\n📄 Validation report saved to: $REPORT_FILE"