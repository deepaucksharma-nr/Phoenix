#!/bin/bash
# validate-boundaries.sh - Validate monorepo modular boundaries

set -euo pipefail

echo "=== Validating Monorepo Boundaries ==="

VIOLATIONS=0
WARNINGS=0

# Function to check for boundary violations
check_violations() {
    local file=$1
    local violations=""
    
    # Check for cross-project imports
    if [[ "$file" =~ projects/([^/]+)/ ]]; then
        local current_project="${BASH_REMATCH[1]}"
        
        # Check if this project imports from other projects
        if grep -E "github\.com/phoenix/platform/projects/[^/]+/" "$file" | grep -v "projects/$current_project/" > /dev/null 2>&1; then
            violations=$(grep -E "github\.com/phoenix/platform/projects/[^/]+/" "$file" | grep -v "projects/$current_project/" | head -3)
            echo "❌ VIOLATION in $file:"
            echo "   Cross-project import detected:"
            echo "$violations" | sed 's/^/     /'
            ((VIOLATIONS++))
        fi
    fi
    
    # Check that pkg doesn't import from projects
    if [[ "$file" =~ pkg/ ]]; then
        if grep -E "github\.com/phoenix/platform/projects/" "$file" > /dev/null 2>&1; then
            violations=$(grep -E "github\.com/phoenix/platform/projects/" "$file" | head -3)
            echo "❌ VIOLATION in $file:"
            echo "   Package importing from projects:"
            echo "$violations" | sed 's/^/     /'
            ((VIOLATIONS++))
        fi
    fi
    
    # Check for old import paths
    if grep -E "phoenix-platform/(cmd|pkg|operators)" "$file" > /dev/null 2>&1; then
        violations=$(grep -E "phoenix-platform/(cmd|pkg|operators)" "$file" | head -3)
        echo "⚠️  WARNING in $file:"
        echo "   Old import path detected:"
        echo "$violations" | sed 's/^/     /'
        ((WARNINGS++))
    fi
}

# Validate Go files in projects
echo "Checking project boundaries..."
find projects -name "*.go" -type f | while read -r file; do
    check_violations "$file"
done

# Validate Go files in pkg
echo "Checking pkg boundaries..."
find pkg -name "*.go" -type f 2>/dev/null | while read -r file; do
    check_violations "$file"
done

# Check go.mod files for proper replace directives
echo ""
echo "Checking go.mod replace directives..."
for mod_file in projects/*/go.mod; do
    if [[ -f "$mod_file" ]]; then
        project=$(dirname "$mod_file")
        
        # Check for pkg/common replace directive
        if ! grep -q "replace github.com/phoenix/platform/pkg/common" "$mod_file"; then
            echo "⚠️  WARNING: $mod_file missing pkg/common replace directive"
            ((WARNINGS++))
        fi
        
        # Check for pkg/contracts replace directive
        if ! grep -q "replace github.com/phoenix/platform/pkg/contracts" "$mod_file"; then
            echo "⚠️  WARNING: $mod_file missing pkg/contracts replace directive"  
            ((WARNINGS++))
        fi
    fi
done

# Summary
echo ""
echo "=== Boundary Validation Summary ==="
echo "Violations: $VIOLATIONS"
echo "Warnings: $WARNINGS"

if [[ $VIOLATIONS -gt 0 ]]; then
    echo ""
    echo "❌ FAILED: Found $VIOLATIONS boundary violations that must be fixed!"
    echo ""
    echo "Rules:"
    echo "1. Projects in /projects/* cannot import from other projects"
    echo "2. Packages in /pkg/* cannot import from /projects/*"
    echo "3. All shared code must be in /pkg/*"
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    echo ""
    echo "⚠️  PASSED with warnings: Consider fixing $WARNINGS warnings"
    exit 0
else
    echo ""
    echo "✅ PASSED: All boundaries are properly maintained!"
    exit 0
fi