#!/bin/bash
# ci-pipeline-optimizer.sh - Analyze and optimize CI/CD pipeline performance
# Created by Abhinav as part of the Pipeline Performance Optimization task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform CI/CD Pipeline Optimizer ===${NC}"

# Set default thresholds
DEFAULT_DURATION_THRESHOLD=10 # Minutes
DEFAULT_SIZE_THRESHOLD=500 # MB
REPO_ROOT=$(git rev-parse --show-toplevel)

# Parse command line arguments
duration_threshold=$DEFAULT_DURATION_THRESHOLD
size_threshold=$DEFAULT_SIZE_THRESHOLD
verbose=false

while getopts "d:s:v" opt; do
  case $opt in
    d) duration_threshold=$OPTARG ;;
    s) size_threshold=$OPTARG ;;
    v) verbose=true ;;
    \?) echo "Invalid option -$OPTARG" >&2; exit 1 ;;
  esac
done

# Function to check Docker images used in Dockerfiles
analyze_dockerfiles() {
    echo -e "\n${YELLOW}Analyzing Dockerfiles for optimization opportunities...${NC}"
    
    # Find all Dockerfiles in the repo
    dockerfiles=$(find "$REPO_ROOT" -name "Dockerfile" -type f)
    
    for dockerfile in $dockerfiles; do
        echo -e "\nAnalyzing ${BLUE}$dockerfile${NC}:"
        
        # Check for multi-stage builds
        if ! grep -q "^FROM .* AS " "$dockerfile"; then
            echo -e "- ${RED}Missing multi-stage build${NC}: Consider implementing to reduce final image size"
        else
            echo -e "- ${GREEN}Using multi-stage builds${NC}: Good practice for image size reduction"
        fi
        
        # Check base image
        base_image=$(grep "^FROM " "$dockerfile" | head -1 | awk '{print $2}')
        if [[ "$base_image" == *"latest"* ]]; then
            echo -e "- ${RED}Using 'latest' tag${NC}: Specify exact version for reproducible builds"
        fi
        
        # Check for .dockerignore
        dockerignore_path="${dockerfile%/*}/.dockerignore"
        if [[ ! -f "$dockerignore_path" ]]; then
            echo -e "- ${RED}Missing .dockerignore${NC}: Add one to speed up builds and reduce context size"
        else
            echo -e "- ${GREEN}Found .dockerignore${NC}: Good practice for optimizing build context"
        fi
        
        # Check for layer optimization
        if grep -q "RUN apt-get update && apt-get install" "$dockerfile"; then
            echo -e "- ${GREEN}Combining RUN commands${NC}: Good practice for reducing layers"
        fi
        
        # Check if using BuildKit
        if ! grep -q "# syntax=docker/dockerfile" "$dockerfile"; then
            echo -e "- ${YELLOW}Consider enabling BuildKit${NC}: For faster builds with caching"
        fi
    done
}

# Function to analyze GitHub workflow files
analyze_workflows() {
    echo -e "\n${YELLOW}Analyzing GitHub workflow files...${NC}"
    
    # Find all workflow files
    workflow_dir="$REPO_ROOT/.github/workflows"
    if [[ ! -d "$workflow_dir" ]]; then
        echo -e "${RED}No workflow directory found at $workflow_dir${NC}"
        return
    fi
    
    workflow_files=$(find "$workflow_dir" -name "*.yml" -o -name "*.yaml")
    
    for workflow in $workflow_files; do
        # Skip template files
        if [[ "$workflow" == *"/_templates/"* ]]; then
            continue
        fi
        
        echo -e "\nAnalyzing workflow: ${BLUE}$(basename "$workflow")${NC}"
        
        # Check for concurrency settings
        if grep -q "concurrency:" "$workflow"; then
            echo -e "- ${GREEN}Using concurrency${NC}: Avoids redundant workflow runs"
        else
            echo -e "- ${YELLOW}Missing concurrency setting${NC}: Add to prevent queued duplicate runs"
        fi
        
        # Check for caching
        if grep -q "actions/cache@" "$workflow" || grep -q "cache:" "$workflow"; then
            echo -e "- ${GREEN}Using caching${NC}: Improves build performance"
        else
            echo -e "- ${RED}Missing cache configuration${NC}: Add caching to speed up build times"
        fi
        
        # Check for matrix builds
        if grep -q "matrix:" "$workflow"; then
            echo -e "- ${GREEN}Using build matrix${NC}: Efficient parallel testing"
        fi
        
        # Check for excessive dependencies between jobs
        dependencies=$(grep -o "needs:" "$workflow" | wc -l)
        jobs=$(grep -o "jobs:" "$workflow" | wc -l)
        if (( dependencies > jobs*2/3 )) && (( jobs > 3 )); then
            echo -e "- ${YELLOW}High job dependencies${NC}: Consider simplifying the dependency chain"
        fi
    done
}

# Function to analyze build artifacts
analyze_artifacts() {
    echo -e "\n${YELLOW}Analyzing build artifacts in workflows...${NC}"
    
    # Find all workflow files
    workflow_dir="$REPO_ROOT/.github/workflows"
    workflow_files=$(find "$workflow_dir" -name "*.yml" -o -name "*.yaml")
    
    for workflow in $workflow_files; then
        # Extract artifact configurations
        artifacts=$(grep -A 10 "actions/upload-artifact@" "$workflow")
        
        if [[ -n "$artifacts" ]]; then
            echo -e "\nFound artifacts in ${BLUE}$(basename "$workflow")${NC}:"
            echo "$artifacts" | grep -E "name:|path:|retention-days:" | sort | uniq -c | sed 's/^/  /'
            
            # Check for retention days
            if ! echo "$artifacts" | grep -q "retention-days:"; then
                echo -e "- ${YELLOW}No retention days specified${NC}: Define to avoid keeping artifacts too long"
            fi
        fi
    done
}

# Function to optimize workflows
optimize_workflows() {
    echo -e "\n${YELLOW}Making optimization recommendations...${NC}"
    
    # 1. Check test/build parallelization
    echo -e "\n${BLUE}Test/Build Parallelization:${NC}"
    echo "- Use build matrices where possible: matrix: { os: [ubuntu-latest, windows-latest], node: [14, 16, 18] }"
    echo "- Split large test suites: npm test -- --shard=1/3"
    
    # 2. Caching recommendations
    echo -e "\n${BLUE}Caching Recommendations:${NC}"
    echo "- npm/yarn modules: actions/cache with package-lock.json as key"
    echo "- Go modules: actions/setup-go with cache: 'go' parameter"
    echo "- Docker layers: Use cache-from/cache-to in docker/build-push-action"
    echo "- Example cache configuration:"
    echo "  uses: actions/cache@v3"
    echo "  with:"
    echo "    path: |"
    echo "      ~/.npm"
    echo "      node_modules"
    echo "    key: \${{ runner.os }}-node-\${{ hashFiles('**/package-lock.json') }}"
    
    # 3. Conditional Job execution
    echo -e "\n${BLUE}Conditional Job Execution:${NC}"
    echo "- Use path-based conditionals for monorepo:"
    echo "  if: \${{ contains(github.event.pull_request.files.*.path, 'frontend/') }}"
    echo "- Skip CI completely with [skip ci] in commit message"
    echo "- Use environmental conditionals:"
    echo "  if: github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/')"
    
    # 4. GitHub-hosted runner optimization
    echo -e "\n${BLUE}GitHub-hosted Runner Optimization:${NC}"
    echo "- Minimize pre-installed tools: https://github.com/actions/virtual-environments"
    echo "- Use setup-* actions instead of installing from scratch"
    echo "- For large repos, checkout with fetch-depth: 1"
}

# Run the analysis
analyze_dockerfiles
analyze_workflows
analyze_artifacts
optimize_workflows

echo -e "\n${GREEN}CI/CD Pipeline optimization analysis complete!${NC}"
echo -e "Implement the recommended changes to achieve the target 30% reduction in pipeline execution time."
