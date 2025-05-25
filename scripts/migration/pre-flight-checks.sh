#!/bin/bash
# pre-flight-checks.sh - Comprehensive pre-migration validation
# This script MUST pass before any migration can begin

set -euo pipefail

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh" || {
    echo "ERROR: Cannot source common functions"
    exit 1
}

# Initialize results
CHECKS_PASSED=0
CHECKS_FAILED=0
WARNINGS=0

# Results file for other scripts to verify
RESULTS_FILE=".migration/pre-flight-results.json"
mkdir -p .migration

# Start results JSON
echo '{' > "$RESULTS_FILE"
echo '  "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",' >> "$RESULTS_FILE"
echo '  "checks": {' >> "$RESULTS_FILE"

# Function to record check result
record_check() {
    local name=$1
    local status=$2
    local message=$3
    
    if [[ "$status" == "passed" ]]; then
        ((CHECKS_PASSED++))
        log_success "$name: $message"
    elif [[ "$status" == "warning" ]]; then
        ((WARNINGS++))
        log_warning "$name: $message"
    else
        ((CHECKS_FAILED++))
        log_error "$name: $message"
    fi
    
    # Add to JSON (with comma handling)
    if [[ $((CHECKS_PASSED + CHECKS_FAILED + WARNINGS)) -gt 1 ]]; then
        echo ',' >> "$RESULTS_FILE"
    fi
    
    cat >> "$RESULTS_FILE" << EOF
    "$name": {
      "status": "$status",
      "message": "$message",
      "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    }
EOF
}

echo -e "${BLUE}=== Phoenix Migration Pre-flight Checks ===${NC}"
echo ""

# Check 1: Git State
echo -e "${YELLOW}Checking Git state...${NC}"
if [[ -n $(git status --porcelain) ]]; then
    record_check "git_state" "failed" "Uncommitted changes detected. Commit or stash changes before migration."
else
    record_check "git_state" "passed" "Git working directory is clean"
fi

# Check 2: Git Remote
if git remote -v | grep -q origin; then
    record_check "git_remote" "passed" "Git remote 'origin' is configured"
else
    record_check "git_remote" "warning" "No git remote 'origin' found. You may not be able to push changes."
fi

# Check 3: Required Tools
echo -e "\n${YELLOW}Checking required tools...${NC}"

# Go version
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_GO="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_GO" "$GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_GO" ]; then
        record_check "go_version" "passed" "Go version $GO_VERSION meets requirement (>=$REQUIRED_GO)"
    else
        record_check "go_version" "failed" "Go version $GO_VERSION is below required $REQUIRED_GO"
    fi
else
    record_check "go_version" "failed" "Go is not installed"
fi

# Node.js version
if command -v node &> /dev/null; then
    NODE_VERSION=$(node -v | sed 's/v//')
    REQUIRED_NODE="18.0.0"
    if [ "$(printf '%s\n' "$REQUIRED_NODE" "$NODE_VERSION" | sort -V | head -n1)" = "$REQUIRED_NODE" ]; then
        record_check "node_version" "passed" "Node.js version $NODE_VERSION meets requirement (>=$REQUIRED_NODE)"
    else
        record_check "node_version" "failed" "Node.js version $NODE_VERSION is below required $REQUIRED_NODE"
    fi
else
    record_check "node_version" "failed" "Node.js is not installed"
fi

# Docker
if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        record_check "docker" "passed" "Docker is installed and running"
    else
        record_check "docker" "failed" "Docker is installed but not running or accessible"
    fi
else
    record_check "docker" "failed" "Docker is not installed"
fi

# Make
if command -v make &> /dev/null; then
    record_check "make" "passed" "Make is installed"
else
    record_check "make" "failed" "Make is not installed"
fi

# Check 4: Disk Space
echo -e "\n${YELLOW}Checking disk space...${NC}"
AVAILABLE_SPACE=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
REQUIRED_SPACE=10

if (( AVAILABLE_SPACE >= REQUIRED_SPACE )); then
    record_check "disk_space" "passed" "${AVAILABLE_SPACE}GB available (required: ${REQUIRED_SPACE}GB)"
else
    record_check "disk_space" "failed" "Only ${AVAILABLE_SPACE}GB available (required: ${REQUIRED_SPACE}GB)"
fi

# Check 5: Running Services
echo -e "\n${YELLOW}Checking for running services...${NC}"
RUNNING_CONTAINERS=$(docker ps -q | wc -l)

if [[ "$RUNNING_CONTAINERS" -eq 0 ]]; then
    record_check "running_services" "passed" "No Docker containers running"
else
    # Check if they're Phoenix containers
    if docker ps --format "table {{.Names}}" | grep -E "(phoenix|control|telemetry)" &> /dev/null; then
        record_check "running_services" "failed" "Phoenix services are running. Stop them with: docker-compose down"
    else
        record_check "running_services" "warning" "$RUNNING_CONTAINERS non-Phoenix containers running"
    fi
fi

# Check 6: OLD_IMPLEMENTATION Directory
echo -e "\n${YELLOW}Checking source directory...${NC}"
if [[ -d "OLD_IMPLEMENTATION" ]]; then
    FILE_COUNT=$(find OLD_IMPLEMENTATION -type f | wc -l)
    record_check "source_directory" "passed" "OLD_IMPLEMENTATION exists with $FILE_COUNT files"
else
    record_check "source_directory" "failed" "OLD_IMPLEMENTATION directory not found"
fi

# Check 7: Target Directories
echo -e "\n${YELLOW}Checking target directories...${NC}"
EXISTING_TARGETS=""
for dir in services packages infrastructure monitoring; do
    if [[ -d "$dir" ]] && [[ -n "$(ls -A "$dir" 2>/dev/null)" ]]; then
        EXISTING_TARGETS="$EXISTING_TARGETS $dir"
    fi
done

if [[ -z "$EXISTING_TARGETS" ]]; then
    record_check "target_directories" "passed" "Target directories are empty or don't exist"
else
    record_check "target_directories" "warning" "Target directories already exist and contain files:$EXISTING_TARGETS"
fi

# Check 8: Network Connectivity (for package downloads)
echo -e "\n${YELLOW}Checking network connectivity...${NC}"
if curl -s --head --connect-timeout 5 https://registry.npmjs.org > /dev/null; then
    record_check "network_npm" "passed" "Can reach npm registry"
else
    record_check "network_npm" "warning" "Cannot reach npm registry - offline mode may be needed"
fi

if curl -s --head --connect-timeout 5 https://proxy.golang.org > /dev/null; then
    record_check "network_go" "passed" "Can reach Go module proxy"
else
    record_check "network_go" "warning" "Cannot reach Go module proxy - offline mode may be needed"
fi

# Check 9: Migration Lock
echo -e "\n${YELLOW}Checking for existing migration...${NC}"
if [[ -f ".migration/migration.lock" ]]; then
    LOCK_AGE=$(( $(date +%s) - $(stat -f %m ".migration/migration.lock" 2>/dev/null || stat -c %Y ".migration/migration.lock") ))
    if (( LOCK_AGE > 3600 )); then
        record_check "migration_lock" "warning" "Stale migration lock found (${LOCK_AGE}s old). Remove with: rm .migration/migration.lock"
    else
        record_check "migration_lock" "failed" "Active migration lock found. Another migration may be in progress."
    fi
else
    record_check "migration_lock" "passed" "No existing migration lock"
fi

# Check 10: Critical Files
echo -e "\n${YELLOW}Checking critical files...${NC}"
CRITICAL_FILES=(
    "OLD_IMPLEMENTATION/phoenix-platform/go.mod"
    "OLD_IMPLEMENTATION/phoenix-platform/dashboard/package.json"
    "OLD_IMPLEMENTATION/docker-compose.yaml"
)

MISSING_FILES=""
for file in "${CRITICAL_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        MISSING_FILES="$MISSING_FILES $file"
    fi
done

if [[ -z "$MISSING_FILES" ]]; then
    record_check "critical_files" "passed" "All critical files present"
else
    record_check "critical_files" "failed" "Missing critical files:$MISSING_FILES"
fi

# Close JSON
echo '  },' >> "$RESULTS_FILE"
echo '  "summary": {' >> "$RESULTS_FILE"
echo '    "total_checks": '$((CHECKS_PASSED + CHECKS_FAILED + WARNINGS))',' >> "$RESULTS_FILE"
echo '    "passed": '$CHECKS_PASSED',' >> "$RESULTS_FILE"
echo '    "failed": '$CHECKS_FAILED',' >> "$RESULTS_FILE"
echo '    "warnings": '$WARNINGS >> "$RESULTS_FILE"
echo '  }' >> "$RESULTS_FILE"
echo '}' >> "$RESULTS_FILE"

# Summary
echo ""
echo -e "${BLUE}=== Pre-flight Check Summary ===${NC}"
echo -e "Total Checks: $((CHECKS_PASSED + CHECKS_FAILED + WARNINGS))"
echo -e "${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "${RED}Failed: $CHECKS_FAILED${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"

# Determine exit status
if [[ $CHECKS_FAILED -gt 0 ]]; then
    echo ""
    echo -e "${RED}✗ Pre-flight checks FAILED${NC}"
    echo -e "${YELLOW}Fix the issues above before proceeding with migration${NC}"
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    echo ""
    echo -e "${YELLOW}⚠ Pre-flight checks passed with warnings${NC}"
    echo -e "${YELLOW}Review warnings above and proceed with caution${NC}"
    exit 0
else
    echo ""
    echo -e "${GREEN}✓ All pre-flight checks PASSED${NC}"
    echo -e "${GREEN}Ready to proceed with migration${NC}"
    exit 0
fi