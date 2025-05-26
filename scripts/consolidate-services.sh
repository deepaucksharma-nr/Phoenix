#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Phoenix Platform Service Consolidation${NC}"
echo "======================================"
echo "This script will help consolidate duplicate services"
echo ""

# Dry run mode by default
DRY_RUN=${DRY_RUN:-true}

if [ "$DRY_RUN" = "true" ]; then
    echo -e "${YELLOW}Running in DRY RUN mode. No changes will be made.${NC}"
    echo "To actually consolidate, run: DRY_RUN=false $0"
    echo ""
fi

# Function to compare directories
compare_service() {
    local service=$1
    local service_dir="services/$service"
    local project_dir="projects/$service"
    
    echo -e "\n${BLUE}Analyzing: $service${NC}"
    echo "----------------------------------------"
    
    # Check if both exist
    if [ ! -d "$service_dir" ] || [ ! -d "$project_dir" ]; then
        echo -e "${YELLOW}Skipping - not in both directories${NC}"
        return
    fi
    
    # Count files
    service_files=$(find "$service_dir" -type f -name "*.go" | wc -l | tr -d ' ')
    project_files=$(find "$project_dir" -type f -name "*.go" | wc -l | tr -d ' ')
    
    echo "Go files in services/: $service_files"
    echo "Go files in projects/: $project_files"
    
    # Check for key files
    echo -n "Makefile: "
    [ -f "$service_dir/Makefile" ] && echo -n "services=✓ " || echo -n "services=✗ "
    [ -f "$project_dir/Makefile" ] && echo "projects=✓" || echo "projects=✗"
    
    echo -n "README: "
    [ -f "$service_dir/README.md" ] && echo -n "services=✓ " || echo -n "services=✗ "
    [ -f "$project_dir/README.md" ] && echo "projects=✓" || echo "projects=✗"
    
    # Check main.go locations
    if [ -f "$service_dir/cmd/main.go" ]; then
        echo "Main in services/: cmd/main.go"
    elif [ -f "$service_dir/main.go" ]; then
        echo "Main in services/: main.go"
    fi
    
    if [ -f "$project_dir/cmd/$service/main.go" ]; then
        echo "Main in projects/: cmd/$service/main.go"
    elif [ -f "$project_dir/cmd/main.go" ]; then
        echo "Main in projects/: cmd/main.go"
    elif [ -f "$project_dir/main.go" ]; then
        echo "Main in projects/: main.go"
    fi
    
    # Recommendation
    if [ $project_files -gt 0 ] && [ -f "$project_dir/Makefile" ]; then
        echo -e "${GREEN}Recommendation: Keep projects/ (better structure)${NC}"
        if [ "$DRY_RUN" = "false" ]; then
            echo -e "${YELLOW}Would remove: $service_dir${NC}"
        fi
    elif [ $service_files -gt $project_files ]; then
        echo -e "${YELLOW}Recommendation: Merge services/ into projects/${NC}"
        if [ "$DRY_RUN" = "false" ]; then
            echo -e "${YELLOW}Would merge: $service_dir -> $project_dir${NC}"
        fi
    else
        echo -e "${BLUE}Recommendation: Manual review needed${NC}"
    fi
}

# List of known duplicates
DUPLICATES=(
    "analytics"
    "anomaly-detector"
    "api"
    "benchmark"
    "collector"
    "control-actuator-go"
    "controller"
    "dashboard"
    "generator"
    "loadsim-operator"
    "phoenix-cli"
    "pipeline-operator"
)

# Analyze each duplicate
for service in "${DUPLICATES[@]}"; do
    compare_service "$service"
done

# Summary
echo -e "\n${BLUE}Summary${NC}"
echo "======="
echo "Total duplicates: ${#DUPLICATES[@]}"

if [ "$DRY_RUN" = "true" ]; then
    echo -e "\n${YELLOW}This was a DRY RUN. To actually consolidate:${NC}"
    echo "1. Review the recommendations above"
    echo "2. Run: DRY_RUN=false $0"
    echo "3. Update go.work to remove deleted paths"
    echo "4. Run: go work sync"
else
    echo -e "\n${GREEN}Consolidation complete!${NC}"
    echo "Next steps:"
    echo "1. Update go.work to remove deleted paths"
    echo "2. Run: go work sync"
    echo "3. Commit changes"
fi

# Check for services only in services/
echo -e "\n${BLUE}Services only in services/ directory:${NC}"
for dir in services/*; do
    if [ -d "$dir" ]; then
        service=$(basename "$dir")
        if [ ! -d "projects/$service" ]; then
            echo "- $service"
        fi
    fi
done

# Check for services only in projects/
echo -e "\n${BLUE}Services only in projects/ directory:${NC}"
for dir in projects/*; do
    if [ -d "$dir" ]; then
        service=$(basename "$dir")
        if [ ! -d "services/$service" ] && [ -f "$dir/go.mod" ]; then
            echo "- $service"
        fi
    fi
done