#!/bin/bash
# Documentation validation script for Phoenix Platform
# This script validates all markdown documentation files

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Phoenix Platform Documentation Validator${NC}"
echo "========================================"
echo

# Counters
total_files=0
valid_files=0
broken_links=0
issues_found=0

# Find all markdown files
echo -e "${YELLOW}Finding all markdown files...${NC}"
md_files=$(find . -name "*.md" -type f | grep -v node_modules | grep -v vendor | sort)
total_files=$(echo "$md_files" | wc -l)
echo "Found $total_files markdown files"
echo

# Function to check if file exists
check_file_exists() {
    local file=$1
    local base_dir=$(dirname "$2")
    
    # Handle relative paths
    if [[ "$file" =~ ^\.\. ]]; then
        # Go up directories
        file="$base_dir/$file"
    elif [[ "$file" =~ ^\. ]]; then
        # Current directory
        file="$base_dir/$(echo "$file" | sed 's|^\./||')"
    elif [[ ! "$file" =~ ^/ ]]; then
        # Relative path
        file="$base_dir/$file"
    fi
    
    # Normalize path
    file=$(cd "$(dirname "$file")" 2>/dev/null && pwd)/$(basename "$file") 2>/dev/null || echo "$file"
    
    if [[ -f "$file" ]]; then
        return 0
    else
        return 1
    fi
}

# Validate each markdown file
echo -e "${YELLOW}Validating markdown files...${NC}"
for file in $md_files; do
    echo -n "Checking $file... "
    
    # Check for broken internal links
    links=$(grep -oE '\[([^]]+)\]\(([^)]+\.md[^)]*)\)' "$file" | sed -E 's/.*\(([^)]+)\).*/\1/' | sed 's/#.*//')
    
    file_valid=true
    for link in $links; do
        # Skip URLs
        if [[ "$link" =~ ^https?:// ]]; then
            continue
        fi
        
        # Check if linked file exists
        if ! check_file_exists "$link" "$file"; then
            if $file_valid; then
                echo
                file_valid=false
            fi
            echo -e "  ${RED}✗ Broken link: $link${NC}"
            ((broken_links++))
        fi
    done
    
    # Check for common issues
    if grep -q "MIT License" "$file" 2>/dev/null; then
        if $file_valid; then
            echo
            file_valid=false
        fi
        echo -e "  ${RED}✗ Found MIT License reference (should be Apache 2.0)${NC}"
        ((issues_found++))
    fi
    
    if grep -q "/deployments/kubernetes" "$file" 2>/dev/null; then
        if $file_valid; then
            echo
            file_valid=false
        fi
        echo -e "  ${RED}✗ Found Kubernetes deployment reference (removed)${NC}"
        ((issues_found++))
    fi
    
    if $file_valid; then
        echo -e "${GREEN}✓${NC}"
        ((valid_files++))
    fi
done

echo
echo -e "${YELLOW}Checking documentation consistency...${NC}"

# Check for NRDOT consistency
echo -n "NRDOT terminology consistency... "
inconsistent=$(grep -r "nrdot\|Nrdot\|NRdot" --include="*.md" | grep -v "NRDOT" | grep -v "nrdot" | wc -l)
if [[ $inconsistent -eq 0 ]]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗ Found $inconsistent inconsistent NRDOT references${NC}"
    ((issues_found+=$inconsistent))
fi

# Check for API version consistency
echo -n "API version consistency... "
v1_refs=$(grep -r "/api/v1" --include="*.md" | grep -v "deprecated" | grep -v "legacy" | wc -l)
if [[ $v1_refs -eq 0 ]]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${YELLOW}⚠ Found $v1_refs references to API v1 (should be v2)${NC}"
fi

# Check for required files
echo
echo -e "${YELLOW}Checking required documentation files...${NC}"
required_files=(
    "README.md"
    "ARCHITECTURE.md"
    "QUICKSTART.md"
    "DEVELOPMENT_GUIDE.md"
    "CONTRIBUTING.md"
    "LICENSE"
    "docs/README.md"
    "docs/api/README.md"
    "docs/operations/README.md"
)

for req_file in "${required_files[@]}"; do
    echo -n "Checking $req_file... "
    if [[ -f "$req_file" ]]; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗ Missing${NC}"
        ((issues_found++))
    fi
done

# Summary
echo
echo -e "${GREEN}Documentation Validation Summary${NC}"
echo "================================"
echo "Total files checked: $total_files"
echo "Valid files: $valid_files"
echo "Files with issues: $((total_files - valid_files))"
echo "Broken links: $broken_links"
echo "Other issues: $issues_found"
echo

if [[ $broken_links -eq 0 && $issues_found -eq 0 ]]; then
    echo -e "${GREEN}✓ All documentation is valid!${NC}"
    exit 0
else
    echo -e "${RED}✗ Documentation validation failed${NC}"
    echo "Please fix the issues above and run this script again."
    exit 1
fi