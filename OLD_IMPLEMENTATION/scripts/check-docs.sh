#!/bin/bash
# Documentation quality checks for Phoenix Platform

set -euo pipefail

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
WARNINGS=0
ERRORS=0

# Output functions
print_check() { 
    ((TOTAL_CHECKS++))
    echo -e "${BLUE}[CHECK]${NC} $1" 
}

print_pass() { 
    ((PASSED_CHECKS++))
    echo -e "${GREEN}  ✓${NC} $1" 
}

print_warning() { 
    ((WARNINGS++))
    echo -e "${YELLOW}  ⚠${NC} $1" 
}

print_error() { 
    ((ERRORS++))
    echo -e "${RED}  ✗${NC} $1" 
}

# Check for broken links in markdown files
check_broken_links() {
    print_check "Checking for broken internal links..."
    
    local broken_links=0
    
    # Find all markdown files
    find "$PROJECT_ROOT" -name "*.md" -type f | grep -v node_modules | while read -r file; do
        # Extract markdown links
        grep -oE '\[([^]]+)\]\(([^)]+)\)' "$file" | grep -oE '\]\([^)]+\)' | sed 's/](\(.*\))/\1/' | while read -r link; do
            # Skip external links
            if [[ "$link" =~ ^https?:// ]] || [[ "$link" =~ ^mailto: ]]; then
                continue
            fi
            
            # Skip anchors
            if [[ "$link" =~ ^# ]]; then
                continue
            fi
            
            # Get the directory of the current file
            local file_dir=$(dirname "$file")
            
            # Resolve relative path
            local target_path="$file_dir/$link"
            target_path=$(echo "$target_path" | sed 's/#.*//')  # Remove anchor
            
            # Check if file exists
            if [ ! -f "$target_path" ]; then
                print_warning "Broken link in $(basename "$file"): $link"
                ((broken_links++))
            fi
        done
    done
    
    if [ $broken_links -eq 0 ]; then
        print_pass "All internal links are valid"
    fi
}

# Check for required sections in documentation
check_required_sections() {
    print_check "Checking for required sections in key documents..."
    
    # Check README files
    for readme in "$PROJECT_ROOT/phoenix-platform/README.md" "$PROJECT_ROOT/phoenix-platform/docs/README.md"; do
        if [ -f "$readme" ]; then
            local missing_sections=()
            
            # Required sections for README
            local required=("## Overview" "## Installation" "## Usage" "## Contributing")
            
            for section in "${required[@]}"; do
                if ! grep -q "^$section" "$readme"; then
                    missing_sections+=("$section")
                fi
            done
            
            if [ ${#missing_sections[@]} -eq 0 ]; then
                print_pass "$(basename "$readme") has all required sections"
            else
                print_warning "$(basename "$readme") missing sections: ${missing_sections[*]}"
            fi
        fi
    done
}

# Check for outdated information
check_outdated_content() {
    print_check "Checking for potentially outdated content..."
    
    # Look for old version numbers
    local old_versions=("0.9" "0.8" "2023")
    
    for version in "${old_versions[@]}"; do
        local files_with_old_version=$(grep -r "$version" "$PROJECT_ROOT" --include="*.md" | grep -v node_modules | wc -l)
        
        if [ "$files_with_old_version" -gt 0 ]; then
            print_warning "Found $files_with_old_version files mentioning potentially old version: $version"
        fi
    done
    
    print_pass "Outdated content check complete"
}

# Check documentation file naming
check_file_naming() {
    print_check "Checking documentation file naming conventions..."
    
    local invalid_names=0
    
    # Find markdown files not following convention
    find "$PROJECT_ROOT/phoenix-platform/docs" -name "*.md" -type f | while read -r file; do
        local filename=$(basename "$file")
        
        # Skip README.md (standard exception)
        if [ "$filename" = "README.md" ]; then
            continue
        fi
        
        # Check if follows UPPERCASE_WITH_UNDERSCORES
        if ! [[ "$filename" =~ ^[A-Z_]+\.md$ ]]; then
            print_warning "Non-standard filename: $filename"
            ((invalid_names++))
        fi
    done
    
    if [ $invalid_names -eq 0 ]; then
        print_pass "All documentation files follow naming convention"
    fi
}

# Check for missing documentation
check_missing_docs() {
    print_check "Checking for missing documentation..."
    
    # Check if key services have documentation
    local services=("api" "controller" "generator" "api-gateway" "control-service")
    
    for service in "${services[@]}"; do
        local service_dir="$PROJECT_ROOT/phoenix-platform/cmd/$service"
        
        if [ -d "$service_dir" ]; then
            local has_readme=false
            
            # Check for README in service directory
            if [ -f "$service_dir/README.md" ]; then
                has_readme=true
            fi
            
            # Check for technical spec
            if [ -f "$PROJECT_ROOT/phoenix-platform/docs/TECHNICAL_SPEC_${service^^}.md" ]; then
                has_readme=true
            fi
            
            if $has_readme; then
                print_pass "Documentation found for $service"
            else
                print_warning "Missing documentation for service: $service"
            fi
        fi
    done
}

# Check code examples in documentation
check_code_examples() {
    print_check "Checking code examples in documentation..."
    
    # Check for code blocks without language specification
    local unspecified_blocks=$(find "$PROJECT_ROOT" -name "*.md" -type f -exec grep -l '```$' {} \; | grep -v node_modules | wc -l)
    
    if [ "$unspecified_blocks" -gt 0 ]; then
        print_warning "Found $unspecified_blocks files with code blocks missing language specification"
    else
        print_pass "All code blocks have language specification"
    fi
}

# Check for TODO items in documentation
check_todos() {
    print_check "Checking for TODO items in documentation..."
    
    local todo_count=$(grep -r "TODO\|FIXME\|XXX" "$PROJECT_ROOT" --include="*.md" | grep -v node_modules | wc -l)
    
    if [ "$todo_count" -gt 0 ]; then
        print_warning "Found $todo_count TODO/FIXME items in documentation"
        grep -r "TODO\|FIXME\|XXX" "$PROJECT_ROOT" --include="*.md" | grep -v node_modules | head -5
    else
        print_pass "No TODO items found in documentation"
    fi
}

# Check documentation consistency
check_consistency() {
    print_check "Checking documentation consistency..."
    
    # Check if all technical specs follow the same structure
    local tech_specs=$(find "$PROJECT_ROOT/phoenix-platform/docs" -name "TECHNICAL_SPEC_*.md" -type f)
    local required_sections=("## Overview" "## Architecture" "## API" "## Configuration" "## Security")
    
    for spec in $tech_specs; do
        local missing=()
        
        for section in "${required_sections[@]}"; do
            if ! grep -q "^$section" "$spec"; then
                missing+=("$section")
            fi
        done
        
        if [ ${#missing[@]} -eq 0 ]; then
            print_pass "$(basename "$spec") has consistent structure"
        else
            print_warning "$(basename "$spec") missing sections: ${missing[*]}"
        fi
    done
}

# Check mkdocs configuration
check_mkdocs() {
    print_check "Checking MkDocs configuration..."
    
    if [ -f "$PROJECT_ROOT/mkdocs.yml" ]; then
        # Check if all files in nav exist
        local missing_files=0
        
        # Extract file paths from mkdocs.yml nav section
        grep -E "^\s+- .+: .+\.md" "$PROJECT_ROOT/mkdocs.yml" | sed 's/.*: //' | while read -r doc_path; do
            if [ ! -f "$PROJECT_ROOT/$doc_path" ]; then
                print_warning "File referenced in mkdocs.yml not found: $doc_path"
                ((missing_files++))
            fi
        done
        
        if [ $missing_files -eq 0 ]; then
            print_pass "All files in mkdocs.yml navigation exist"
        fi
        
        # Check if mkdocs builds successfully
        if command -v mkdocs &> /dev/null; then
            if mkdocs build --strict --quiet 2>/dev/null; then
                print_pass "MkDocs builds successfully"
            else
                print_error "MkDocs build failed"
            fi
        else
            print_warning "mkdocs not installed, skipping build check"
        fi
    else
        print_error "mkdocs.yml not found"
    fi
}

# Main execution
main() {
    echo "Phoenix Platform Documentation Quality Checks"
    echo "==========================================="
    echo ""
    
    # Run all checks
    check_broken_links
    check_required_sections
    check_outdated_content
    check_file_naming
    check_missing_docs
    check_code_examples
    check_todos
    check_consistency
    check_mkdocs
    
    # Summary
    echo ""
    echo "==========================================="
    echo "Summary:"
    echo "  Total checks: $TOTAL_CHECKS"
    echo -e "  ${GREEN}Passed: $PASSED_CHECKS${NC}"
    echo -e "  ${YELLOW}Warnings: $WARNINGS${NC}"
    echo -e "  ${RED}Errors: $ERRORS${NC}"
    
    # Exit code
    if [ $ERRORS -gt 0 ]; then
        echo ""
        echo -e "${RED}Documentation quality check failed!${NC}"
        exit 1
    elif [ $WARNINGS -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}Documentation quality check passed with warnings.${NC}"
        exit 0
    else
        echo ""
        echo -e "${GREEN}Documentation quality check passed!${NC}"
        exit 0
    fi
}

# Run main
main "$@"