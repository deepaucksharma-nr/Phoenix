#!/bin/bash
# update-module-names.sh - Update Go module names to use phoenix-vnext

set -euo pipefail

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Updating Go Module Names ===${NC}"

# Update projects
for project in projects/*/go.mod; do
    if [ -f "$project" ]; then
        project_dir=$(dirname "$project")
        project_name=$(basename "$project_dir")
        
        echo -e "${YELLOW}Updating module: $project_name${NC}"
        
        # Update module name
        sed -i.bak "s|^module .*|module github.com/phoenix-vnext/platform/projects/$project_name|" "$project"
        rm -f "$project.bak"
        
        echo -e "${GREEN}✓ Updated $project_name${NC}"
    fi
done

# Update services
for service in services/*/go.mod; do
    if [ -f "$service" ]; then
        service_dir=$(dirname "$service")
        service_name=$(basename "$service_dir")
        
        echo -e "${YELLOW}Updating module: $service_name${NC}"
        
        # Update module name
        sed -i.bak "s|^module .*|module github.com/phoenix-vnext/platform/services/$service_name|" "$service"
        rm -f "$service.bak"
        
        echo -e "${GREEN}✓ Updated $service_name${NC}"
    fi
done

# Update nested services
for service in services/*/*/go.mod; do
    if [ -f "$service" ]; then
        service_dir=$(dirname "$service")
        parent_dir=$(dirname "$service_dir")
        parent_name=$(basename "$parent_dir")
        service_name=$(basename "$service_dir")
        
        echo -e "${YELLOW}Updating module: $parent_name/$service_name${NC}"
        
        # Update module name
        sed -i.bak "s|^module .*|module github.com/phoenix-vnext/platform/services/$parent_name/$service_name|" "$service"
        rm -f "$service.bak"
        
        echo -e "${GREEN}✓ Updated $parent_name/$service_name${NC}"
    fi
done

# Update operators
for operator in operators/*/go.mod; do
    if [ -f "$operator" ]; then
        operator_dir=$(dirname "$operator")
        operator_name=$(basename "$operator_dir")
        
        echo -e "${YELLOW}Updating module: $operator_name${NC}"
        
        # Update module name
        sed -i.bak "s|^module .*|module github.com/phoenix-vnext/platform/operators/$operator_name|" "$operator"
        rm -f "$operator.bak"
        
        echo -e "${GREEN}✓ Updated $operator_name${NC}"
    fi
done

echo -e "${GREEN}Module name updates complete!${NC}"