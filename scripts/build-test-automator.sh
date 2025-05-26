#!/bin/bash
# build-test-automator.sh - Automated build and test script for Phoenix Platform
# Created by Abhinav as part of build & test automation task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
SKIP_DOCKER=${SKIP_DOCKER:-false}
SKIP_TESTS=${SKIP_TESTS:-false}
BUILD_ALL=${BUILD_ALL:-false}
CI_MODE=${CI_MODE:-false}
VERBOSE=${VERBOSE:-false}
PARALLEL=${PARALLEL:-true}
DRY_RUN=${DRY_RUN:-false}

# Target specific project or component
TARGET_PROJECT=${TARGET_PROJECT:-""}

# Define build output directory
BUILD_DIR="$REPO_ROOT/bin"
LOG_DIR="$REPO_ROOT/logs/build"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
BUILD_LOG="$LOG_DIR/build-$TIMESTAMP.log"
TEST_LOG="$LOG_DIR/test-$TIMESTAMP.log"

# Counters for summary
PROJECTS_BUILT=0
PROJECTS_FAILED=0
TESTS_PASSED=0
TESTS_FAILED=0
START_TIME=$(date +%s)

# Create output directories
mkdir -p "$BUILD_DIR" "$LOG_DIR"

# Display banner
echo -e "${BLUE}=================================================================================${NC}"
echo -e "${BLUE}                 Phoenix Platform - Build & Test Automator                       ${NC}"
echo -e "${BLUE}=================================================================================${NC}"
echo ""
echo -e "Build Date: $(date)"
echo -e "Log Files: ${LOG_DIR}"
if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}DRY RUN MODE - No actual builds or tests will be performed${NC}"
fi
echo ""

# Helper functions
function log_info() {
    echo -e "${BLUE}[INFO] $1${NC}"
    if [ "$VERBOSE" = "true" ]; then
        echo -e "[INFO] $1" >> "$BUILD_LOG"
    fi
}

function log_success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
    echo -e "[SUCCESS] $1" >> "$BUILD_LOG"
}

function log_error() {
    echo -e "${RED}[ERROR] $1${NC}"
    echo -e "[ERROR] $1" >> "$BUILD_LOG"
}

function log_warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
    echo -e "[WARNING] $1" >> "$BUILD_LOG"
}

# Execute a command with proper logging
function execute_command() {
    local cmd="$1"
    local description="$2"
    local log_file="$3"
    local error_msg="$4"
    
    echo -e "${BLUE}Executing: $description${NC}"
    echo "Command: $cmd" >> "$log_file"
    echo "----------------------------------------" >> "$log_file"
    
    if [ "$DRY_RUN" = "true" ]; then
        echo -e "${YELLOW}[DRY RUN] Would execute: $cmd${NC}"
        return 0
    fi
    
    if [ "$VERBOSE" = "true" ]; then
        if eval "$cmd" 2>&1 | tee -a "$log_file"; then
            log_success "$description completed successfully"
            return 0
        else
            log_error "$error_msg"
            return 1
        fi
    else
        if eval "$cmd" >> "$log_file" 2>&1; then
            log_success "$description completed successfully"
            return 0
        else
            log_error "$error_msg"
            echo -e "${YELLOW}Check log file for details: $log_file${NC}"
            return 1
        fi
    fi
}

# Find all Go projects
function find_go_projects() {
    find "$REPO_ROOT/projects" -name "go.mod" -not -path "*/vendor/*" | while read -r mod_file; do
        dirname "$mod_file"
    done
}

# Find all Node.js projects
function find_node_projects() {
    find "$REPO_ROOT/projects" -name "package.json" -not -path "*/node_modules/*" | while read -r pkg_file; do
        dirname "$pkg_file"
    done
}

# Determine if project should be included based on TARGET_PROJECT
function should_process_project() {
    local project_path="$1"
    local project_name=$(basename "$project_path")
    
    if [ -z "$TARGET_PROJECT" ]; then
        return 0  # Process all projects if no target specified
    elif [ "$project_name" = "$TARGET_PROJECT" ]; then
        return 0  # Process if project name matches target
    else
        return 1  # Skip if project name doesn't match target
    fi
}

# Build Go project
function build_go_project() {
    local project_path="$1"
    local project_name=$(basename "$project_path")
    local project_log="$LOG_DIR/build-go-$project_name-$TIMESTAMP.log"
    
    if ! should_process_project "$project_path"; then
        return 0
    fi
    
    log_info "Building Go project: $project_name"
    
    # Execute build using project's Makefile if it exists
    if [ -f "$project_path/Makefile" ]; then
        if execute_command "cd '$project_path' && make build" \
            "Building $project_name using Makefile" "$project_log" \
            "Failed to build $project_name"; then
            PROJECTS_BUILT=$((PROJECTS_BUILT + 1))
        else
            PROJECTS_FAILED=$((PROJECTS_FAILED + 1))
            if [ "$CI_MODE" = "true" ]; then
                return 1
            fi
        fi
    else
        # Fall back to direct go build command
        local main_files=$(find "$project_path" -name "main.go" | grep -v "_test.go")
        
        if [ -z "$main_files" ]; then
            log_warning "No main.go found in $project_name, skipping build"
            return 0
        fi
        
        for main_file in $main_files; do
            local dir_name=$(dirname "$main_file")
            local binary_name=$(basename "$(dirname "$dir_name")")
            
            # If main.go is in cmd/something, use 'something' as the binary name
            if [[ "$dir_name" == */cmd/* ]]; then
                binary_name=$(basename "$dir_name")
            fi
            
            local output_path="$BUILD_DIR/$binary_name"
            
            if execute_command "cd '$project_path' && go build -o '$output_path' '$main_file'" \
                "Building $project_name/$binary_name" "$project_log" \
                "Failed to build $project_name/$binary_name"; then
                PROJECTS_BUILT=$((PROJECTS_BUILT + 1))
            else
                PROJECTS_FAILED=$((PROJECTS_FAILED + 1))
                if [ "$CI_MODE" = "true" ]; then
                    return 1
                fi
            fi
        done
    fi
}

# Build Node.js project
function build_node_project() {
    local project_path="$1"
    local project_name=$(basename "$project_path")
    local project_log="$LOG_DIR/build-node-$project_name-$TIMESTAMP.log"
    
    if ! should_process_project "$project_path"; then
        return 0
    fi
    
    log_info "Building Node.js project: $project_name"
    
    # Check if project needs to be built (has build script)
    if ! grep -q '"build"' "$project_path/package.json"; then
        log_warning "No build script found in $project_name/package.json, skipping build"
        return 0
    fi
    
    # Determine package manager
    local package_manager="npm"
    if [ -f "$project_path/pnpm-lock.yaml" ]; then
        package_manager="pnpm"
    elif [ -f "$project_path/yarn.lock" ]; then
        package_manager="yarn"
    fi
    
    # Install dependencies if node_modules doesn't exist
    if [ ! -d "$project_path/node_modules" ]; then
        log_info "Installing dependencies for $project_name"
        
        local install_cmd=""
        case "$package_manager" in
            pnpm) install_cmd="pnpm install" ;;
            yarn) install_cmd="yarn install --frozen-lockfile" ;;
            npm) install_cmd="npm ci" ;;
        esac
        
        if ! execute_command "cd '$project_path' && $install_cmd" \
            "Installing dependencies for $project_name" "$project_log" \
            "Failed to install dependencies for $project_name"; then
            if [ "$CI_MODE" = "true" ]; then
                return 1
            fi
        fi
    fi
    
    # Execute build
    local build_cmd=""
    case "$package_manager" in
        pnpm) build_cmd="pnpm build" ;;
        yarn) build_cmd="yarn build" ;;
        npm) build_cmd="npm run build" ;;
    esac
    
    if execute_command "cd '$project_path' && $build_cmd" \
        "Building $project_name" "$project_log" \
        "Failed to build $project_name"; then
        PROJECTS_BUILT=$((PROJECTS_BUILT + 1))
    else
        PROJECTS_FAILED=$((PROJECTS_FAILED + 1))
        if [ "$CI_MODE" = "true" ]; then
            return 1
        fi
    fi
}

# Test Go project
function test_go_project() {
    if [ "$SKIP_TESTS" = "true" ]; then
        return 0
    fi
    
    local project_path="$1"
    local project_name=$(basename "$project_path")
    local project_log="$LOG_DIR/test-go-$project_name-$TIMESTAMP.log"
    
    if ! should_process_project "$project_path"; then
        return 0
    fi
    
    log_info "Testing Go project: $project_name"
    
    # Execute tests using project's Makefile if it exists
    if [ -f "$project_path/Makefile" ] && grep -q "test:" "$project_path/Makefile"; then
        if execute_command "cd '$project_path' && make test" \
            "Testing $project_name using Makefile" "$project_log" \
            "Tests failed for $project_name"; then
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
            if [ "$CI_MODE" = "true" ]; then
                return 1
            fi
        fi
    else
        # Fall back to direct go test command
        if execute_command "cd '$project_path' && go test -v ./..." \
            "Testing $project_name" "$project_log" \
            "Tests failed for $project_name"; then
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
            if [ "$CI_MODE" = "true" ]; then
                return 1
            fi
        fi
    fi
}

# Test Node.js project
function test_node_project() {
    if [ "$SKIP_TESTS" = "true" ]; then
        return 0
    fi
    
    local project_path="$1"
    local project_name=$(basename "$project_path")
    local project_log="$LOG_DIR/test-node-$project_name-$TIMESTAMP.log"
    
    if ! should_process_project "$project_path"; then
        return 0
    fi
    
    # Check if project has tests (has test script)
    if ! grep -q '"test"' "$project_path/package.json"; then
        log_warning "No test script found in $project_name/package.json, skipping tests"
        return 0
    fi
    
    log_info "Testing Node.js project: $project_name"
    
    # Determine package manager
    local package_manager="npm"
    if [ -f "$project_path/pnpm-lock.yaml" ]; then
        package_manager="pnpm"
    elif [ -f "$project_path/yarn.lock" ]; then
        package_manager="yarn"
    fi
    
    # Execute tests
    local test_cmd=""
    case "$package_manager" in
        pnpm) test_cmd="pnpm test" ;;
        yarn) test_cmd="yarn test" ;;
        npm) test_cmd="npm test" ;;
    esac
    
    if execute_command "cd '$project_path' && $test_cmd" \
        "Testing $project_name" "$project_log" \
        "Tests failed for $project_name"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        if [ "$CI_MODE" = "true" ]; then
            return 1
        fi
    fi
}

# Build Docker images
function build_docker_images() {
    if [ "$SKIP_DOCKER" = "true" ]; then
        return 0
    fi
    
    log_info "Building Docker images"
    local docker_log="$LOG_DIR/docker-build-$TIMESTAMP.log"
    
    # Find all Dockerfiles in the projects
    local dockerfiles=$(find "$REPO_ROOT/projects" -name "Dockerfile" | sort)
    local docker_builds=0
    local docker_failures=0
    
    for dockerfile in $dockerfiles; do
        local project_dir=$(dirname "$dockerfile")
        local project_name=$(basename "$(dirname "$dockerfile")")
        
        if ! should_process_project "$project_dir"; then
            continue
        fi
        
        # If Dockerfile is in a subdir like "build/docker", adjust project_name
        if [[ "$(basename "$project_dir")" == "docker" ]]; then
            project_name=$(basename "$(dirname "$(dirname "$dockerfile")")")
        fi
        
        log_info "Building Docker image for $project_name"
        
        # Check if project has a Makefile with docker target
        if [ -f "$project_dir/../Makefile" ] && grep -q "docker:" "$project_dir/../Makefile"; then
            if execute_command "cd '$(dirname "$project_dir")' && make docker" \
                "Building Docker image for $project_name using Makefile" "$docker_log" \
                "Failed to build Docker image for $project_name"; then
                docker_builds=$((docker_builds + 1))
            else
                docker_failures=$((docker_failures + 1))
            fi
        else
            # Fall back to direct docker build
            local tag="phoenix/$project_name:latest"
            if execute_command "docker build -t '$tag' -f '$dockerfile' '$(dirname "$project_dir")'" \
                "Building Docker image $tag" "$docker_log" \
                "Failed to build Docker image for $project_name"; then
                docker_builds=$((docker_builds + 1))
            else
                docker_failures=$((docker_failures + 1))
            fi
        fi
    done
    
    log_info "Docker build summary: $docker_builds successful, $docker_failures failed"
}

# Run parallel builds and tests
function run_parallel_builds_and_tests() {
    local go_projects=($(find_go_projects))
    local node_projects=($(find_node_projects))
    local pids=()
    
    log_info "Starting parallel builds..."
    
    # Build Go projects in parallel
    for project in "${go_projects[@]}"; do
        if should_process_project "$project"; then
            build_go_project "$project" & 
            pids+=($!)
        fi
    done
    
    # Build Node.js projects in parallel
    for project in "${node_projects[@]}"; do
        if should_process_project "$project"; then
            build_node_project "$project" &
            pids+=($!)
        fi
    done
    
    # Wait for all builds to complete
    for pid in "${pids[@]}"; do
        wait "$pid" || true
    done
    
    # Clear pids array
    pids=()
    
    if [ "$SKIP_TESTS" != "true" ]; then
        log_info "Starting parallel tests..."
        
        # Test Go projects in parallel
        for project in "${go_projects[@]}"; do
            if should_process_project "$project"; then
                test_go_project "$project" &
                pids+=($!)
            fi
        done
        
        # Test Node.js projects in parallel
        for project in "${node_projects[@]}"; do
            if should_process_project "$project"; then
                test_node_project "$project" &
                pids+=($!)
            fi
        done
        
        # Wait for all tests to complete
        for pid in "${pids[@]}"; do
            wait "$pid" || true
        done
    fi
    
    # Build Docker images if needed
    if [ "$SKIP_DOCKER" != "true" ]; then
        build_docker_images
    fi
}

# Run sequential builds and tests
function run_sequential_builds_and_tests() {
    local go_projects=($(find_go_projects))
    local node_projects=($(find_node_projects))
    
    log_info "Starting sequential builds..."
    
    # Build Go projects
    for project in "${go_projects[@]}"; do
        if should_process_project "$project"; then
            build_go_project "$project"
        fi
    done
    
    # Build Node.js projects
    for project in "${node_projects[@]}"; do
        if should_process_project "$project"; then
            build_node_project "$project"
        fi
    done
    
    if [ "$SKIP_TESTS" != "true" ]; then
        log_info "Starting sequential tests..."
        
        # Test Go projects
        for project in "${go_projects[@]}"; do
            if should_process_project "$project"; then
                test_go_project "$project"
            fi
        done
        
        # Test Node.js projects
        for project in "${node_projects[@]}"; do
            if should_process_project "$project"; then
                test_node_project "$project"
            fi
        done
    fi
    
    # Build Docker images if needed
    if [ "$SKIP_DOCKER" != "true" ]; then
        build_docker_images
    fi
}

# Display help message
function show_help() {
    echo "Usage: $(basename "$0") [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  -p, --project NAME   Build specific project/component"
    echo "  -a, --all            Build all projects"
    echo "  -s, --sequential     Run builds sequentially (default: parallel)"
    echo "  --skip-docker        Skip Docker image builds"
    echo "  --skip-tests         Skip running tests"
    echo "  -v, --verbose        Show verbose output"
    echo "  --ci-mode            Exit on first failure (for CI usage)"
    echo "  --dry-run            Show what would be done without actually doing it"
    echo ""
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            ;;
        -p|--project)
            TARGET_PROJECT="$2"
            shift 2
            ;;
        -a|--all)
            BUILD_ALL=true
            shift
            ;;
        -s|--sequential)
            PARALLEL=false
            shift
            ;;
        --skip-docker)
            SKIP_DOCKER=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --ci-mode)
            CI_MODE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            ;;
    esac
done

# Main execution
log_info "Starting build and test process"
log_info "Target project: ${TARGET_PROJECT:-"all"}"
log_info "Build mode: $([ "$PARALLEL" = "true" ] && echo "parallel" || echo "sequential")"
log_info "Skip Docker builds: $SKIP_DOCKER"
log_info "Skip tests: $SKIP_TESTS"
log_info "CI mode: $CI_MODE"

# Run builds and tests
if [ "$PARALLEL" = "true" ]; then
    run_parallel_builds_and_tests
else
    run_sequential_builds_and_tests
fi

# Calculate elapsed time
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
MINUTES=$((ELAPSED / 60))
SECONDS=$((ELAPSED % 60))

# Display summary
echo ""
echo -e "${BLUE}=================================================================================${NC}"
echo -e "${BLUE}                         Build & Test Summary                                    ${NC}"
echo -e "${BLUE}=================================================================================${NC}"
echo ""
echo -e "Total time: ${MINUTES}m ${SECONDS}s"
echo -e "Projects built successfully: ${GREEN}${PROJECTS_BUILT}${NC}"
echo -e "Projects failed to build: ${RED}${PROJECTS_FAILED}${NC}"

if [ "$SKIP_TESTS" != "true" ]; then
    echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"
fi

if [ "$PROJECTS_FAILED" -gt 0 ] || [ "$TESTS_FAILED" -gt 0 ]; then
    echo -e "${RED}Build and test process completed with errors${NC}"
    echo -e "Check logs in: $LOG_DIR"
    exit 1
else
    echo -e "${GREEN}Build and test process completed successfully${NC}"
    echo -e "Logs saved to: $LOG_DIR"
    exit 0
fi
