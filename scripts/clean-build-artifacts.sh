#!/bin/bash
# Clean build artifacts from Phoenix Platform
# This prevents large binaries from being accidentally committed

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Phoenix Platform - Clean Build Artifacts${NC}"
echo "========================================"
echo

# Find and remove binary files
echo -e "${YELLOW}Removing build artifacts...${NC}"

# Common binary locations
BINARY_PATHS=(
    "projects/phoenix-api/bin"
    "projects/phoenix-api/build/bin"
    "projects/phoenix-agent/bin"
    "projects/phoenix-agent/build/bin"
    "projects/phoenix-cli/bin"
    "projects/phoenix-cli/build/bin"
    "projects/dashboard/dist"
    "projects/dashboard/build"
)

# Remove binaries
for path in "${BINARY_PATHS[@]}"; do
    if [ -d "$path" ]; then
        echo "Cleaning $path..."
        rm -rf "$path"
    fi
done

# Find any remaining large files
echo
echo -e "${YELLOW}Checking for remaining large files...${NC}"
large_files=$(find . -type f -size +1M -not -path "*/node_modules/*" -not -path "*/.git/*" -not -path "*/vendor/*" 2>/dev/null || true)

if [ -n "$large_files" ]; then
    echo -e "${RED}Warning: Found large files that might cause issues:${NC}"
    echo "$large_files" | while read -r file; do
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "  - $file ($size)"
    done
else
    echo -e "${GREEN}✓ No large files found${NC}"
fi

# Check git status
echo
echo -e "${YELLOW}Checking git status...${NC}"
if git status --porcelain | grep -E "\.(exe|dll|so|dylib|bin)$" > /dev/null 2>&1; then
    echo -e "${RED}Warning: Binary files may be staged for commit${NC}"
    git status --porcelain | grep -E "\.(exe|dll|so|dylib|bin)$"
else
    echo -e "${GREEN}✓ No binary files staged${NC}"
fi

echo
echo -e "${GREEN}Clean complete!${NC}"
echo
echo "To build projects without committing binaries:"
echo "  - Use 'make build' which respects .gitignore"
echo "  - Binaries will be in build/ or bin/ directories"
echo "  - These directories are ignored by Git"