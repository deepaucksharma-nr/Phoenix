#!/bin/bash
# fix-all-imports.sh - Fix all import paths to use phoenix-vnext

set -euo pipefail

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Fixing All Import Paths ===${NC}"

# Find and fix all Go files
find . -name "*.go" -type f \
    -not -path "./OLD_IMPLEMENTATION/*" \
    -not -path "./.git/*" \
    -not -path "./vendor/*" \
    -not -path "./node_modules/*" | while read -r file; do
    
    # Check if file contains any phoenix imports
    if grep -q "github.com/phoenix/" "$file" 2>/dev/null; then
        echo -e "${YELLOW}Fixing imports in: $file${NC}"
        
        # Replace all variations of phoenix imports
        sed -i '' 's|"github.com/phoenix/platform/|"github.com/phoenix-vnext/platform/|g' "$file"
        sed -i '' 's|"github.com/phoenix/|"github.com/phoenix-vnext/|g' "$file"
    fi
done

# Also update go.mod replace statements
find . -name "go.mod" -type f \
    -not -path "./OLD_IMPLEMENTATION/*" \
    -not -path "./.git/*" | while read -r file; do
    
    if grep -q "github.com/phoenix/" "$file" 2>/dev/null; then
        echo -e "${YELLOW}Fixing go.mod: $file${NC}"
        
        # Update replace statements
        sed -i '' 's|github.com/phoenix/platform/|github.com/phoenix-vnext/platform/|g' "$file"
        sed -i '' 's|github.com/phoenix/|github.com/phoenix-vnext/|g' "$file"
    fi
done

echo -e "${GREEN}Import path fixes complete!${NC}"