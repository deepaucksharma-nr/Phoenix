#!/bin/bash
# Script to run CLI and API integration tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== Phoenix CLI and API Test Runner ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi
    
    # Check if PostgreSQL is running
    if ! pg_isready -h localhost -p 5432 &> /dev/null; then
        echo -e "${YELLOW}Warning: PostgreSQL doesn't seem to be running on localhost:5432${NC}"
        echo "Integration tests will be skipped if database is not available"
    fi
    
    echo -e "${GREEN}Prerequisites check passed${NC}"
    echo
}

# Setup test environment
setup_test_env() {
    echo "Setting up test environment..."
    
    # Create test database if it doesn't exist
    if pg_isready -h localhost -p 5432 &> /dev/null; then
        echo "Creating test database..."
        createdb -h localhost -p 5432 -U phoenix phoenix_test 2>/dev/null || true
        
        # Run migrations on test database
        echo "Running migrations on test database..."
        export DATABASE_URL="postgres://phoenix:phoenix@localhost:5432/phoenix_test?sslmode=disable"
        cd "$PROJECT_ROOT"
        go run scripts/migrate.go up
    fi
    
    echo -e "${GREEN}Test environment ready${NC}"
    echo
}

# Run unit tests
run_unit_tests() {
    echo "Running unit tests..."
    cd "$PROJECT_ROOT"
    
    # CLI unit tests
    echo "Testing CLI components..."
    go test -v ./cmd/phoenix-cli/... -short
    
    # API unit tests
    echo "Testing API components..."
    go test -v ./pkg/api/... -short
    
    echo -e "${GREEN}Unit tests completed${NC}"
    echo
}

# Run integration tests
run_integration_tests() {
    echo "Running integration tests..."
    
    # Check if API server is running
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${YELLOW}Warning: API server not running on localhost:8080${NC}"
        echo "Starting API server in background..."
        
        # Build and start API server
        cd "$PROJECT_ROOT"
        go build -o api-server ./cmd/api
        ./api-server &
        API_PID=$!
        
        # Wait for server to start
        echo "Waiting for API server to start..."
        for i in {1..30}; do
            if curl -s http://localhost:8080/health > /dev/null 2>&1; then
                echo -e "${GREEN}API server started${NC}"
                break
            fi
            sleep 1
        done
        
        if [ $i -eq 30 ]; then
            echo -e "${RED}Error: API server failed to start${NC}"
            kill $API_PID 2>/dev/null || true
            exit 1
        fi
    fi
    
    # Set test environment variables
    export PHOENIX_API_URL="http://localhost:8080"
    export TEST_DATABASE_URL="postgres://phoenix:phoenix@localhost:5432/phoenix_test?sslmode=disable"
    
    # Run integration tests
    cd "$PROJECT_ROOT"
    go test -v ./test/integration/... -tags=integration
    
    # Clean up API server if we started it
    if [ ! -z "$API_PID" ]; then
        echo "Stopping API server..."
        kill $API_PID 2>/dev/null || true
    fi
    
    echo -e "${GREEN}Integration tests completed${NC}"
    echo
}

# Generate coverage report
generate_coverage() {
    echo "Generating test coverage report..."
    cd "$PROJECT_ROOT"
    
    # Run tests with coverage
    go test -coverprofile=coverage.out \
        ./cmd/phoenix-cli/... \
        ./pkg/api/... \
        -short
    
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    
    # Display coverage summary
    echo "Coverage summary:"
    go tool cover -func=coverage.out | grep total
    
    echo -e "${GREEN}Coverage report generated: coverage.html${NC}"
    echo
}

# Main execution
main() {
    check_prerequisites
    
    # Parse command line arguments
    RUN_UNIT=true
    RUN_INTEGRATION=false
    RUN_COVERAGE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit-only)
                RUN_INTEGRATION=false
                shift
                ;;
            --integration-only)
                RUN_UNIT=false
                RUN_INTEGRATION=true
                shift
                ;;
            --all)
                RUN_INTEGRATION=true
                shift
                ;;
            --coverage)
                RUN_COVERAGE=true
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --unit-only       Run only unit tests (default)"
                echo "  --integration-only Run only integration tests"
                echo "  --all            Run both unit and integration tests"
                echo "  --coverage       Generate coverage report"
                echo "  --help           Show this help message"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Setup test environment
    if [ "$RUN_INTEGRATION" = true ]; then
        setup_test_env
    fi
    
    # Run tests
    if [ "$RUN_UNIT" = true ]; then
        run_unit_tests
    fi
    
    if [ "$RUN_INTEGRATION" = true ]; then
        run_integration_tests
    fi
    
    if [ "$RUN_COVERAGE" = true ]; then
        generate_coverage
    fi
    
    echo -e "${GREEN}All tests completed successfully!${NC}"
}

# Run main function
main "$@"