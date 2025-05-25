#!/bin/bash
# archive-old-implementation.sh - Archive the OLD_IMPLEMENTATION directory

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Archiving OLD_IMPLEMENTATION ===${NC}"
echo ""

# Configuration
ARCHIVE_DIR="archives"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
ARCHIVE_NAME="OLD_IMPLEMENTATION-${TIMESTAMP}.tar.gz"

# Check if OLD_IMPLEMENTATION exists
if [[ ! -d "OLD_IMPLEMENTATION" ]]; then
    echo -e "${YELLOW}OLD_IMPLEMENTATION directory not found. Nothing to archive.${NC}"
    exit 0
fi

# Calculate size
SIZE=$(du -sh OLD_IMPLEMENTATION | cut -f1)
echo "Directory size: $SIZE"
echo ""

# Ask for confirmation
echo -e "${YELLOW}This will:${NC}"
echo "1. Create an archive: $ARCHIVE_DIR/$ARCHIVE_NAME"
echo "2. Remove the OLD_IMPLEMENTATION directory"
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Create archive directory
mkdir -p "$ARCHIVE_DIR"

# Create archive
echo ""
echo -e "${YELLOW}Creating archive...${NC}"
ARCHIVE_PATH="$ARCHIVE_DIR/$ARCHIVE_NAME"
tar -czf "$ARCHIVE_PATH" \
    --exclude='.git' \
    --exclude='node_modules' \
    --exclude='vendor' \
    --exclude='dist' \
    --exclude='build' \
    --exclude='*.log' \
    --exclude='*.tmp' \
    OLD_IMPLEMENTATION/

# Verify archive
if [[ -f "$ARCHIVE_DIR/$ARCHIVE_NAME" ]]; then
    ARCHIVE_SIZE=$(du -h "$ARCHIVE_DIR/$ARCHIVE_NAME" | cut -f1)
    echo -e "${GREEN}✓ Archive created: $ARCHIVE_DIR/$ARCHIVE_NAME (${ARCHIVE_SIZE})${NC}"
    
    # Create checksum
    echo -n "Creating checksum... "
    if command -v sha256sum &> /dev/null; then
        sha256sum "$ARCHIVE_DIR/$ARCHIVE_NAME" > "$ARCHIVE_DIR/$ARCHIVE_NAME.sha256"
    else
        shasum -a 256 "$ARCHIVE_DIR/$ARCHIVE_NAME" > "$ARCHIVE_DIR/$ARCHIVE_NAME.sha256"
    fi
    echo -e "${GREEN}✓${NC}"
    
    # Create archive manifest
    echo -e "${YELLOW}Creating manifest...${NC}"
    cat > "$ARCHIVE_DIR/OLD_IMPLEMENTATION-${TIMESTAMP}.manifest" << EOF
Archive Manifest
================
Date: $(date)
Original Size: $SIZE
Archive Size: $ARCHIVE_SIZE
Files Archived: $(find OLD_IMPLEMENTATION -type f | wc -l | tr -d ' ')
Directories: $(find OLD_IMPLEMENTATION -type d | wc -l | tr -d ' ')

Excluded:
- .git directories
- node_modules
- vendor directories
- dist/build directories
- Log files
- Temporary files

Checksum:
$(cat "$ARCHIVE_DIR/$ARCHIVE_NAME.sha256")

Notes:
This archive contains the pre-migration Phoenix Platform implementation.
The code has been successfully migrated to the new monorepo structure.
EOF
    
    echo -e "${GREEN}✓ Manifest created${NC}"
    
    # Remove OLD_IMPLEMENTATION
    echo ""
    echo -e "${YELLOW}Removing OLD_IMPLEMENTATION directory...${NC}"
    rm -rf OLD_IMPLEMENTATION
    echo -e "${GREEN}✓ Directory removed${NC}"
    
    # Update .gitignore
    if ! grep -q "^OLD_IMPLEMENTATION" .gitignore 2>/dev/null; then
        echo "OLD_IMPLEMENTATION/" >> .gitignore
        echo -e "${GREEN}✓ Updated .gitignore${NC}"
    fi
    
    echo ""
    echo -e "${GREEN}=== Archive Complete ===${NC}"
    echo "Archive location: $ARCHIVE_DIR/$ARCHIVE_NAME"
    echo "Manifest: $ARCHIVE_DIR/OLD_IMPLEMENTATION-${TIMESTAMP}.manifest"
    echo ""
    echo "To restore if needed:"
    echo "  tar -xzf $ARCHIVE_DIR/$ARCHIVE_NAME"
else
    echo -e "${RED}✗ Failed to create archive${NC}"
    exit 1
fi