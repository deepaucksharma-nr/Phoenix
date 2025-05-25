#!/bin/bash

set -e

# Phoenix Platform Integration Test Runner
# This script sets up the environment and runs integration tests

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-phoenix}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-phoenix}"
TEST_DATABASE_NAME="${TEST_DATABASE_NAME:-phoenix_test}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "1.21" "$GO_VERSION" | sort -V | head -n1)" != "1.21" ]]; then
        log_error "Go 1.21 or later is required. Current version: $GO_VERSION"
        exit 1
    fi
    
    log_success "Go version: $GO_VERSION"
}

# Check PostgreSQL connectivity
check_postgres() {
    log_info "Checking PostgreSQL connectivity..."
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        log_warn "psql not found. Attempting to connect using Go..."
    else
        # Test connection with psql
        if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c '\q' &> /dev/null; then
            log_success "PostgreSQL connection successful"
        else
            log_error "Cannot connect to PostgreSQL at ${POSTGRES_HOST}:${POSTGRES_PORT}"
            log_info "Make sure PostgreSQL is running and accessible with:"
            echo "  Host: $POSTGRES_HOST"
            echo "  Port: $POSTGRES_PORT"
            echo "  User: $POSTGRES_USER"
            echo "  Password: $POSTGRES_PASSWORD"
            exit 1
        fi
    fi
}

# Setup test environment
setup_environment() {
    log_info "Setting up test environment..."
    
    # Export environment variables for tests
    export POSTGRES_HOST="$POSTGRES_HOST"
    export POSTGRES_PORT="$POSTGRES_PORT"
    export POSTGRES_USER="$POSTGRES_USER"
    export POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
    export TEST_DATABASE_NAME="$TEST_DATABASE_NAME"
    export TEST_DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${TEST_DATABASE_NAME}?sslmode=disable"
    
    log_success "Environment variables set"
}

# Run tests
run_tests() {
    log_info "Running integration tests..."
    
    cd "$PROJECT_DIR"
    
    # Build the projects first to catch any compilation errors
    log_info "Building controller..."
    if ! go build -o /tmp/phoenix-controller ./cmd/controller; then
        log_error "Failed to build controller"
        exit 1
    fi
    
    log_info "Building generator..."
    if ! go build -o /tmp/phoenix-generator ./cmd/generator; then
        log_error "Failed to build generator"
        exit 1
    fi
    
    log_success "Build successful"
    
    # Run integration tests with verbose output
    log_info "Executing integration tests..."
    
    TEST_FLAGS="-tags=integration -v -timeout=30m"
    
    # Allow running specific test packages or all
    if [[ $# -gt 0 ]]; then
        TEST_PACKAGE="$1"
        log_info "Running tests in package: $TEST_PACKAGE"
        go test $TEST_FLAGS "$TEST_PACKAGE"
    else
        log_info "Running all integration tests..."
        go test $TEST_FLAGS ./test/integration/...
    fi
    
    TEST_EXIT_CODE=$?
    
    # Cleanup temporary files
    rm -f /tmp/phoenix-controller /tmp/phoenix-generator
    
    if [[ $TEST_EXIT_CODE -eq 0 ]]; then
        log_success "All integration tests passed!"
    else
        log_error "Some integration tests failed"
        exit $TEST_EXIT_CODE
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    
    # Note: Test cleanup is handled by TestMain in the integration tests
    # This function is for any additional cleanup if needed
    
    log_success "Cleanup completed"
}

# Main execution
main() {
    log_info "Phoenix Platform Integration Test Runner"
    log_info "======================================"
    
    # Trap cleanup on exit
    trap cleanup EXIT
    
    # Run checks and setup
    check_prerequisites
    check_postgres
    setup_environment
    
    # Run tests
    run_tests "$@"
}

# Help function
show_help() {
    echo "Phoenix Platform Integration Test Runner"
    echo ""
    echo "Usage: $0 [OPTIONS] [TEST_PACKAGE]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  POSTGRES_HOST       PostgreSQL host (default: localhost)"
    echo "  POSTGRES_PORT       PostgreSQL port (default: 5432)"
    echo "  POSTGRES_USER       PostgreSQL user (default: phoenix)"
    echo "  POSTGRES_PASSWORD   PostgreSQL password (default: phoenix)"
    echo "  TEST_DATABASE_NAME  Test database name (default: phoenix_test)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all integration tests"
    echo "  $0 ./test/integration/experiment_controller_test.go  # Run specific test file"
    echo ""
    echo "Prerequisites:"
    echo "  - Go 1.21 or later"
    echo "  - PostgreSQL server running and accessible"
    echo "  - Test database will be created automatically"
}

# Check for help flag
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# Run main function
main "$@"