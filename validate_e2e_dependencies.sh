#!/bin/bash

# Phoenix E2E Dependencies and Contracts Validation Script
# This script validates all dependencies and contracts required for e2e testing

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Phoenix E2E Dependencies and Contracts Validation ===${NC}\n"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if command_exists go; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        REQUIRED_VERSION="1.24.0"
        if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
            echo -e "${GREEN}✓ Go version $GO_VERSION meets requirement (>= $REQUIRED_VERSION)${NC}"
        else
            echo -e "${RED}✗ Go version $GO_VERSION does not meet requirement (>= $REQUIRED_VERSION)${NC}"
            exit 1
        fi
    else
        echo -e "${RED}✗ Go is not installed${NC}"
        exit 1
    fi
}

# Function to validate Go modules
validate_go_modules() {
    echo -e "\n${YELLOW}Validating Go modules...${NC}"
    
    # Check main pkg module
    if [ -f "pkg/go.mod" ]; then
        echo -e "${GREEN}✓ Found pkg/go.mod${NC}"
        cd pkg
        if go mod verify > /dev/null 2>&1; then
            echo -e "${GREEN}✓ pkg module dependencies verified${NC}"
        else
            echo -e "${RED}✗ pkg module dependencies verification failed${NC}"
            exit 1
        fi
        cd ..
    else
        echo -e "${RED}✗ pkg/go.mod not found${NC}"
        exit 1
    fi
    
    # Check all project modules
    for project in projects/*; do
        if [ -f "$project/go.mod" ]; then
            project_name=$(basename "$project")
            echo -e "${GREEN}✓ Found $project_name/go.mod${NC}"
            cd "$project"
            if go mod verify > /dev/null 2>&1; then
                echo -e "${GREEN}✓ $project_name module dependencies verified${NC}"
            else
                echo -e "${RED}✗ $project_name module dependencies verification failed${NC}"
                exit 1
            fi
            cd ../..
        fi
    done
}

# Function to check contracts
check_contracts() {
    echo -e "\n${YELLOW}Checking contracts...${NC}"
    
    # Check OpenAPI contracts
    if [ -f "pkg/contracts/openapi/control-api.yaml" ]; then
        echo -e "${GREEN}✓ Found OpenAPI contract: control-api.yaml${NC}"
    else
        echo -e "${RED}✗ OpenAPI contract not found${NC}"
        exit 1
    fi
    
    # Check Proto contracts
    proto_files=(
        "pkg/contracts/proto/v1/common.proto"
        "pkg/contracts/proto/v1/controller.proto"
        "pkg/contracts/proto/v1/experiment.proto"
        "pkg/contracts/proto/v1/generator.proto"
    )
    
    for proto in "${proto_files[@]}"; do
        if [ -f "$proto" ]; then
            echo -e "${GREEN}✓ Found Proto contract: $(basename "$proto")${NC}"
        else
            echo -e "${RED}✗ Proto contract not found: $proto${NC}"
            exit 1
        fi
    done
}

# Function to check E2E test files
check_e2e_tests() {
    echo -e "\n${YELLOW}Checking E2E test files...${NC}"
    
    e2e_tests=(
        "tests/e2e/simple_e2e_test.go"
        "tests/e2e/experiment_workflow_test.go"
    )
    
    for test in "${e2e_tests[@]}"; do
        if [ -f "$test" ]; then
            echo -e "${GREEN}✓ Found E2E test: $(basename "$test")${NC}"
            # Check for e2e build tag
            if grep -q "//go:build e2e" "$test" || grep -q "// +build e2e" "$test"; then
                echo -e "${GREEN}  ✓ Has e2e build tag${NC}"
            else
                echo -e "${YELLOW}  ⚠ Missing e2e build tag${NC}"
            fi
        else
            echo -e "${RED}✗ E2E test not found: $test${NC}"
            exit 1
        fi
    done
}

# Function to check required services for E2E
check_required_services() {
    echo -e "\n${YELLOW}Checking required services configuration...${NC}"
    
    # Check docker-compose.yml
    if [ -f "docker-compose.yml" ]; then
        echo -e "${GREEN}✓ Found docker-compose.yml${NC}"
        # Check for required services in docker-compose
        if grep -q "postgres" docker-compose.yml; then
            echo -e "${GREEN}  ✓ PostgreSQL service configured${NC}"
        else
            echo -e "${YELLOW}  ⚠ PostgreSQL service not found in docker-compose${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ docker-compose.yml not found${NC}"
    fi
    
    # Check for service configurations
    services=(
        "platform-api"
        "controller"
        "pipeline-operator"
    )
    
    for service in "${services[@]}"; do
        if [ -d "projects/$service" ]; then
            echo -e "${GREEN}✓ Found service: $service${NC}"
        else
            echo -e "${YELLOW}⚠ Service directory not found: $service${NC}"
        fi
    done
}

# Function to validate E2E test dependencies
validate_e2e_dependencies() {
    echo -e "\n${YELLOW}Validating E2E test dependencies...${NC}"
    
    # Required Go packages for E2E tests
    required_packages=(
        "github.com/stretchr/testify"
        "github.com/google/uuid"
        "google.golang.org/grpc"
        "k8s.io/client-go"
    )
    
    # Check if packages are in go.mod files
    found_packages=0
    for pkg in "${required_packages[@]}"; do
        if grep -r "$pkg" pkg/go.mod projects/*/go.mod > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Found dependency: $pkg${NC}"
            ((found_packages++))
        else
            echo -e "${YELLOW}⚠ Dependency not found in go.mod files: $pkg${NC}"
        fi
    done
    
    if [ $found_packages -eq ${#required_packages[@]} ]; then
        echo -e "${GREEN}✓ All E2E test dependencies found${NC}"
    else
        echo -e "${YELLOW}⚠ Some E2E test dependencies might be missing${NC}"
    fi
}

# Function to check environment setup
check_environment() {
    echo -e "\n${YELLOW}Checking environment setup...${NC}"
    
    # Check for .env.template
    if [ -f ".env.template" ]; then
        echo -e "${GREEN}✓ Found .env.template${NC}"
        # Check for required environment variables
        env_vars=(
            "DATABASE_URL"
            "NEW_RELIC_API_KEY"
            "NEW_RELIC_OTLP_ENDPOINT"
        )
        
        for var in "${env_vars[@]}"; do
            if grep -q "$var" .env.template; then
                echo -e "${GREEN}  ✓ Environment variable template found: $var${NC}"
            else
                echo -e "${YELLOW}  ⚠ Environment variable not in template: $var${NC}"
            fi
        done
    else
        echo -e "${YELLOW}⚠ .env.template not found${NC}"
    fi
}

# Function to run E2E tests (dry run)
run_e2e_test_check() {
    echo -e "\n${YELLOW}E2E Test Execution Check...${NC}"
    
    # Check if we can compile the E2E tests
    echo "Checking E2E test compilation..."
    cd tests/e2e
    if go test -tags e2e -c > /dev/null 2>&1; then
        echo -e "${GREEN}✓ E2E tests compile successfully${NC}"
        rm -f e2e.test  # Clean up compiled test binary
    else
        echo -e "${RED}✗ E2E tests compilation failed${NC}"
        echo "Run 'go test -tags e2e -c' in tests/e2e for details"
    fi
    cd ../..
}

# Main validation flow
main() {
    # Check basic requirements
    echo -e "${YELLOW}Checking basic requirements...${NC}"
    check_go_version
    
    # Validate modules
    validate_go_modules
    
    # Check contracts
    check_contracts
    
    # Check E2E tests
    check_e2e_tests
    
    # Check required services
    check_required_services
    
    # Validate E2E dependencies
    validate_e2e_dependencies
    
    # Check environment
    check_environment
    
    # E2E test compilation check
    run_e2e_test_check
    
    echo -e "\n${GREEN}=== E2E Dependencies and Contracts Validation Complete ===${NC}"
    echo -e "${BLUE}To run E2E tests, use: make test-e2e${NC}"
}

# Run main function
main
