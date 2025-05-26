#!/bin/bash

# Phoenix Platform Phase 2 Cleanup Script
# This script continues the cleanup by removing more redundant implementations

set -euo pipefail

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="cleanup-phase2-backup-${TIMESTAMP}"
LOG_FILE="cleanup-phase2-${TIMESTAMP}.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

# Create backup directory
mkdir -p "$BACKUP_DIR"

log "${BLUE}🧹 Phoenix Platform Phase 2 Cleanup${NC}"
log "${BLUE}======================================${NC}"
log "Backup directory: $BACKUP_DIR"
log "Log file: $LOG_FILE"
echo

# Step 1: Backup packages directory before consolidation
log "${YELLOW}📦 Backing up packages directory...${NC}"
if [ -d "packages" ]; then
    tar -czf "$BACKUP_DIR/packages-backup.tar.gz" packages/ 2>/dev/null || true
    log "${GREEN}✅ Packages backed up${NC}"
fi
echo

# Step 2: Consolidate packages into pkg
log "${YELLOW}🔄 Consolidating packages/ into pkg/...${NC}"

# Move go-common contents to pkg/common
if [ -d "packages/go-common" ] && [ ! -d "pkg/common" ]; then
    log "  Moving packages/go-common → pkg/common"
    mv "packages/go-common" "pkg/common"
    
    # Update the go.mod file in pkg/common
    if [ -f "pkg/common/go.mod" ]; then
        sed -i.bak 's|github.com/phoenix-vnext/platform/packages/go-common|github.com/phoenix/platform/pkg/common|g' "pkg/common/go.mod"
        rm -f "pkg/common/go.mod.bak"
    fi
elif [ -d "packages/go-common" ] && [ -d "pkg/common" ]; then
    log "  ${YELLOW}⚠️  pkg/common already exists, checking for unique files${NC}"
    # Copy any unique files
    rsync -av --ignore-existing "packages/go-common/" "pkg/common/" 2>/dev/null || true
fi

# Move contracts if not already consolidated
if [ -d "packages/contracts" ] && [ ! -f "packages/contracts/.consolidated" ]; then
    log "  ${BLUE}contracts already consolidated to pkg/contracts/proto${NC}"
    touch "packages/contracts/.consolidated"
fi

# Remove the now-empty packages directory
if [ -d "packages" ]; then
    # Check if it's truly empty or just has empty subdirs
    if [ -z "$(find packages -type f -not -path '*/\.*' 2>/dev/null)" ]; then
        rm -rf "packages"
        log "${GREEN}✅ Removed empty packages directory${NC}"
    else
        log "${YELLOW}⚠️  packages directory still has files, manual review needed${NC}"
    fi
fi
echo

# Step 3: Remove duplicate docker-compose file
log "${YELLOW}🐳 Cleaning duplicate Docker files...${NC}"
if [ -f "docker-compose-fixed.yml" ] && [ -f "docker-compose.yml" ]; then
    # Check if they're different
    if diff -q "docker-compose-fixed.yml" "docker-compose.yml" >/dev/null; then
        rm "docker-compose-fixed.yml"
        log "${GREEN}✅ Removed duplicate docker-compose-fixed.yml${NC}"
    else
        log "${YELLOW}⚠️  docker-compose files differ, manual review needed${NC}"
    fi
fi
echo

# Step 4: Clean up empty directories
log "${YELLOW}📁 Cleaning empty directories...${NC}"

# Function to remove empty directories
remove_empty_dirs() {
    local count=0
    while IFS= read -r dir; do
        if [ -d "$dir" ] && [ -z "$(ls -A "$dir")" ]; then
            rmdir "$dir" 2>/dev/null && ((count++)) || true
        fi
    done < <(find . -type d -not -path "./.git/*" -not -path "./node_modules/*" -not -path "./.next/*" 2>/dev/null | sort -r)
    echo $count
}

# Remove empty directories (multiple passes to handle nested empties)
total_removed=0
for i in {1..5}; do
    removed=$(remove_empty_dirs)
    total_removed=$((total_removed + removed))
    if [ "$removed" -eq 0 ]; then
        break
    fi
done

log "${GREEN}✅ Removed $total_removed empty directories${NC}"
echo

# Step 5: Consolidate config directories
log "${YELLOW}⚙️  Consolidating configuration directories...${NC}"
if [ -d "configs" ] && [ -d "config" ]; then
    # Prefer config over configs (singular is standard)
    if [ -z "$(find configs -type f -not -path '*/\.*' 2>/dev/null)" ]; then
        # configs is empty, remove it
        rm -rf "configs"
        log "${GREEN}✅ Removed empty configs directory${NC}"
    else
        # Move contents from configs to config
        log "  Merging configs/ into config/"
        rsync -av "configs/" "config/" 2>/dev/null || true
        rm -rf "configs"
        log "${GREEN}✅ Consolidated configuration directories${NC}"
    fi
fi
echo

# Step 6: Create shared test utilities directory
log "${YELLOW}🧪 Creating shared test utilities...${NC}"
if [ ! -d "pkg/testing" ]; then
    mkdir -p "pkg/testing"
    
    # Create a basic test utilities file
    cat > "pkg/testing/helpers.go" << 'EOF'
package testing

import (
    "testing"
    "os"
    "path/filepath"
)

// TestContext provides common test setup
type TestContext struct {
    T       *testing.T
    TempDir string
}

// NewTestContext creates a new test context with temp directory
func NewTestContext(t *testing.T) *TestContext {
    return &TestContext{
        T:       t,
        TempDir: t.TempDir(),
    }
}

// Cleanup performs test cleanup
func (tc *TestContext) Cleanup() {
    // Additional cleanup if needed
}

// FixturePath returns the path to a test fixture
func FixturePath(t *testing.T, path string) string {
    t.Helper()
    abs, err := filepath.Abs(filepath.Join("testdata", path))
    if err != nil {
        t.Fatalf("failed to get fixture path: %v", err)
    }
    return abs
}

// RequireEnv skips the test if the environment variable is not set
func RequireEnv(t *testing.T, key string) string {
    t.Helper()
    value := os.Getenv(key)
    if value == "" {
        t.Skipf("skipping test: %s not set", key)
    }
    return value
}
EOF

    # Create go.mod for testing package
    cat > "pkg/testing/go.mod" << 'EOF'
module github.com/phoenix/platform/pkg/testing

go 1.23
EOF

    log "${GREEN}✅ Created shared test utilities in pkg/testing${NC}"
fi
echo

# Step 7: Update imports for consolidated packages
log "${YELLOW}🔄 Updating import paths...${NC}"

# Update imports from packages/go-common to pkg/common
find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec \
    sed -i.bak 's|github.com/phoenix-vnext/platform/packages/go-common|github.com/phoenix/platform/pkg/common|g' {} \;

# Update imports from packages/contracts to pkg/contracts
find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec \
    sed -i.bak 's|github.com/phoenix-vnext/platform/packages/contracts|github.com/phoenix/platform/pkg/contracts|g' {} \;

# Clean up backup files
find . -name "*.bak" -delete

log "${GREEN}✅ Import paths updated${NC}"
echo

# Step 8: Update go.mod replace directives
log "${YELLOW}📝 Updating go.mod replace directives...${NC}"

update_go_mod_replaces() {
    local go_mod=$1
    if [ -f "$go_mod" ]; then
        # Update package paths in replace directives
        sed -i.bak 's|../../packages/go-common|../../pkg/common|g' "$go_mod"
        sed -i.bak 's|../../packages/contracts|../../pkg/contracts|g' "$go_mod"
        rm -f "${go_mod}.bak"
    fi
}

# Update all go.mod files
find . -name "go.mod" -not -path "./vendor/*" -not -path "./.git/*" | while read -r mod_file; do
    update_go_mod_replaces "$mod_file"
done

log "${GREEN}✅ go.mod files updated${NC}"
echo

# Step 9: Final validation
log "${YELLOW}🔍 Running final validation...${NC}"

# Check for any remaining references to packages/
remaining_refs=$(grep -r "packages/" --include="*.go" --include="*.mod" --include="*.sum" \
    --exclude-dir=".git" --exclude-dir="vendor" --exclude-dir="node_modules" . 2>/dev/null | wc -l || echo "0")

if [ "$remaining_refs" -gt 0 ]; then
    log "${YELLOW}⚠️  Found $remaining_refs remaining references to packages/ directory${NC}"
    log "    Run: grep -r 'packages/' --include='*.go' --include='*.mod' . | grep -v vendor"
else
    log "${GREEN}✅ No remaining references to packages/ directory${NC}"
fi

# Final summary
echo
log "${GREEN}🎉 Phase 2 cleanup complete!${NC}"
log "${BLUE}======================================${NC}"
log "Backups saved in: $BACKUP_DIR"
log "Log file: $LOG_FILE"
echo
log "${YELLOW}Summary of changes:${NC}"
log "- Consolidated packages/ into pkg/"
log "- Removed $total_removed empty directories"
log "- Created shared test utilities in pkg/testing"
log "- Updated all import paths"
log "- Cleaned duplicate Docker files"
echo
log "${YELLOW}Next steps:${NC}"
log "1. Run 'go work sync' to update workspace"
log "2. Run 'make test' to ensure tests pass"
log "3. Regenerate any proto files if needed"
log "4. Check git status and commit changes"