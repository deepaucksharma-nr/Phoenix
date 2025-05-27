#!/bin/bash

# Phoenix Platform E2E Test Runner
# This script runs comprehensive end-to-end tests for the Phoenix platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
PHOENIX_API_URL="${PHOENIX_API_URL:-http://localhost:8080}"
POSTGRES_URL="${DATABASE_URL:-postgres://phoenix:phoenix@localhost/phoenix_test?sslmode=disable}"
TEST_TIMEOUT="${TEST_TIMEOUT:-10m}"
CLEANUP="${CLEANUP:-true}"

# Test directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo -e "${YELLOW}🧪 Phoenix Platform E2E Test Suite${NC}"
echo "=================================="
echo "API URL: $PHOENIX_API_URL"
echo "Database: $POSTGRES_URL"
echo "Timeout: $TEST_TIMEOUT"
echo ""

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed${NC}"
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed${NC}"
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All prerequisites met${NC}"
}

# Function to start test infrastructure
start_infrastructure() {
    echo -e "${YELLOW}Starting test infrastructure...${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Start dependencies (Postgres, Prometheus, etc.)
    docker-compose -f docker-compose-infra.yml up -d
    
    # Wait for Postgres
    echo "Waiting for PostgreSQL..."
    for i in {1..30}; do
        if docker-compose -f docker-compose-infra.yml exec -T postgres pg_isready -U phoenix &> /dev/null; then
            echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
            break
        fi
        sleep 1
    done
    
    # Run migrations
    echo "Running database migrations..."
    cd "$PROJECT_ROOT/projects/phoenix-api"
    go run cmd/api/main.go migrate up
    
    echo -e "${GREEN}✅ Infrastructure started${NC}"
}

# Function to build services
build_services() {
    echo -e "${YELLOW}Building services...${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Build Phoenix API
    echo "Building Phoenix API..."
    cd projects/phoenix-api
    make build
    
    # Build Phoenix Agent
    echo "Building Phoenix Agent..."
    cd ../phoenix-agent
    make build
    
    echo -e "${GREEN}✅ Services built${NC}"
}

# Function to start services
start_services() {
    echo -e "${YELLOW}Starting Phoenix services...${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Start Phoenix API
    echo "Starting Phoenix API..."
    cd projects/phoenix-api
    ./api &
    API_PID=$!
    
    # Wait for API to be ready
    echo "Waiting for Phoenix API..."
    for i in {1..30}; do
        if curl -s "$PHOENIX_API_URL/health" > /dev/null; then
            echo -e "${GREEN}✅ Phoenix API is ready${NC}"
            break
        fi
        sleep 1
    done
    
    # Start test agents
    echo "Starting test agents..."
    cd ../phoenix-agent
    for i in {1..3}; do
        AGENT_HOST_ID="e2e-test-agent-$i" ./phoenix-agent &
        AGENT_PIDS+=($!)
    done
    
    echo -e "${GREEN}✅ Services started${NC}"
}

# Function to run tests
run_tests() {
    echo -e "${YELLOW}Running E2E tests...${NC}"
    
    cd "$SCRIPT_DIR"
    
    # Set test environment
    export PHOENIX_API_URL
    export DATABASE_URL="$POSTGRES_URL"
    export GO_TEST_TIMEOUT="$TEST_TIMEOUT"
    
    # Run tests based on argument
    if [ "$1" = "simple" ]; then
        echo "Running simple E2E test..."
        go test -v -tags=e2e -timeout="$TEST_TIMEOUT" -run TestSimpleE2E ./...
    elif [ "$1" = "workflow" ]; then
        echo "Running experiment workflow test..."
        go test -v -tags=e2e -timeout="$TEST_TIMEOUT" -run TestExperimentWorkflowE2E ./...
    elif [ "$1" = "comprehensive" ]; then
        echo "Running comprehensive E2E test..."
        go test -v -tags=e2e -timeout="$TEST_TIMEOUT" -run TestComprehensiveE2E ./...
    else
        echo "Running all E2E tests..."
        go test -v -tags=e2e -timeout="$TEST_TIMEOUT" ./...
    fi
    
    TEST_EXIT_CODE=$?
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
    else
        echo -e "${RED}❌ Tests failed with exit code $TEST_EXIT_CODE${NC}"
    fi
    
    return $TEST_EXIT_CODE
}

# Function to cleanup
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    
    # Stop services
    if [ ! -z "$API_PID" ]; then
        echo "Stopping Phoenix API..."
        kill $API_PID 2>/dev/null || true
    fi
    
    for pid in "${AGENT_PIDS[@]}"; do
        echo "Stopping agent $pid..."
        kill $pid 2>/dev/null || true
    done
    
    # Stop infrastructure
    cd "$PROJECT_ROOT"
    docker-compose -f docker-compose-infra.yml down
    
    echo -e "${GREEN}✅ Cleanup complete${NC}"
}

# Main execution
main() {
    # Parse arguments
    TEST_TYPE="${1:-all}"
    
    # Check prerequisites
    check_prerequisites
    
    # Setup trap for cleanup
    if [ "$CLEANUP" = "true" ]; then
        trap cleanup EXIT
    fi
    
    # Start infrastructure
    start_infrastructure
    
    # Build services
    build_services
    
    # Start services
    start_services
    
    # Run tests
    run_tests "$TEST_TYPE"
    TEST_RESULT=$?
    
    # Exit with test result
    exit $TEST_RESULT
}

# Run main function
main "$@"