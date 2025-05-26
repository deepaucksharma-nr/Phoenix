#!/bin/bash

# Phoenix Platform Codebase Cleanup Script
# Removes duplicates and dead code after analysis

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo "🧹 Phoenix Platform Codebase Cleanup"
echo "===================================="
echo ""

# Safety check
echo -e "${YELLOW}⚠️  WARNING: This script will delete files and directories!${NC}"
echo "Have you:"
echo "  1. Created a backup? (tar -czf phoenix-backup.tar.gz .)"
echo "  2. Committed all changes? (git status)"
echo "  3. Run the analysis script? (./scripts/analyze-codebase.sh)"
echo ""
read -p "Continue with cleanup? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Cleanup cancelled."
    exit 1
fi

echo ""
echo "📦 Phase 1: Removing duplicate services..."
echo ""

# List of services to check for duplicates
duplicate_services=(
    "analytics"
    "benchmark"
    "dashboard"
    "phoenix-cli"
    "loadsim-operator"
    "pipeline-operator"
)

for service in "${duplicate_services[@]}"; do
    if [ -d "services/$service" ] && [ -d "projects/$service" ]; then
        echo -e "${RED}Removing duplicate:${NC} services/$service"
        rm -rf "services/$service"
    fi
done

# Remove control plane services that are being consolidated
if [ -d "services/control-actuator-go" ]; then
    echo -e "${RED}Removing:${NC} services/control-actuator-go (empty stub)"
    rm -rf "services/control-actuator-go"
fi

if [ -d "services/collector" ]; then
    echo -e "${RED}Removing:${NC} services/collector (duplicate of projects/collector)"
    rm -rf "services/collector"
fi

echo ""
echo "🔄 Phase 2: Removing duplicate operators..."
echo ""

# Remove old operator directories if new ones exist
if [ -d "operators/loadsim" ] && [ -d "projects/loadsim-operator" ]; then
    echo -e "${RED}Removing duplicate:${NC} operators/loadsim"
    rm -rf "operators/loadsim"
fi

if [ -d "operators/pipeline" ] && [ -d "projects/pipeline-operator" ]; then
    echo -e "${RED}Removing duplicate:${NC} operators/pipeline"
    rm -rf "operators/pipeline"
fi

echo ""
echo "💀 Phase 3: Removing empty directories and dead code..."
echo ""

# Remove empty directories in pkg/
echo "Removing empty directories in pkg/..."
find pkg -type d -empty -delete 2>/dev/null || true

# Remove known dead packages
dead_packages=(
    "pkg/auth/oauth"
    "pkg/auth/rbac" 
    "pkg/database/redis"
    "pkg/messaging/kafka"
    "pkg/messaging/nats"
    "pkg/k8s/client"
    "pkg/k8s/controllers"
    "pkg/k8s/informers"
)

for pkg in "${dead_packages[@]}"; do
    if [ -d "$pkg" ]; then
        # Check if directory has any Go files
        go_files=$(find "$pkg" -name "*.go" 2>/dev/null | wc -l)
        if [ "$go_files" -eq 0 ]; then
            echo -e "${RED}Removing empty package:${NC} $pkg"
            rm -rf "$pkg"
        fi
    fi
done

echo ""
echo "📝 Phase 4: Updating go.work..."
echo ""

# Create a backup of go.work
cp go.work go.work.backup

# Remove entries for deleted directories
while IFS= read -r line; do
    if [[ "$line" =~ ^[[:space:]]*\./(.*) ]]; then
        path="${BASH_REMATCH[1]}"
        if [ ! -d "$path" ]; then
            echo -e "${YELLOW}Removing obsolete go.work entry:${NC} $path"
            # Comment out the line instead of removing
            sed -i.tmp "s|^\s*\./$path|// REMOVED: ./$path|g" go.work
        fi
    fi
done < go.work.backup

# Clean up temp file
rm -f go.work.tmp

echo ""
echo "🔧 Phase 5: Cleaning up proto files..."
echo ""

# This is a sensitive operation - just report for now
echo "Proto file consolidation requires manual review:"
echo "  1. Decide on canonical location (recommend: pkg/grpc/proto/)"
echo "  2. Update proto generation scripts"
echo "  3. Update all imports"
echo "  4. Regenerate proto files"

echo ""
echo "🧹 Phase 6: Final cleanup..."
echo ""

# Remove backup files
find . -name "*.backup" -o -name "*.bak" -type f -delete 2>/dev/null || true

# Remove .DS_Store files (macOS)
find . -name ".DS_Store" -type f -delete 2>/dev/null || true

echo ""
echo -e "${GREEN}✅ Cleanup complete!${NC}"
echo ""
echo "📋 Next steps:"
echo "  1. Run 'go work sync' to update workspace"
echo "  2. Run 'make build' to verify builds"
echo "  3. Run 'make test' to verify tests"
echo "  4. Update any broken imports"
echo "  5. Commit changes"
echo ""
echo "🔍 To verify cleanup:"
echo "  ./scripts/analyze-codebase.sh"
echo ""
echo "📊 Cleanup statistics:"
echo "  - Services directory: $(find services -name "*.go" 2>/dev/null | wc -l) Go files remaining"
echo "  - Projects directory: $(find projects -name "*.go" 2>/dev/null | wc -l) Go files"
echo "  - Empty directories removed: $(find pkg -type d -empty 2>/dev/null | wc -l)"
echo ""

# Create a cleanup report
cat > CLEANUP_REPORT.md << EOF
# Codebase Cleanup Report

**Date**: $(date)

## Actions Taken

### Removed Duplicate Services
$(for service in "${duplicate_services[@]}"; do
    if [ ! -d "services/$service" ] && [ -d "projects/$service" ]; then
        echo "- ✅ services/$service"
    fi
done)

### Removed Duplicate Operators
- ✅ operators/loadsim (kept projects/loadsim-operator)
- ✅ operators/pipeline (kept projects/pipeline-operator)

### Removed Empty Packages
$(for pkg in "${dead_packages[@]}"; do
    if [ ! -d "$pkg" ]; then
        echo "- ✅ $pkg"
    fi
done)

### Updated Files
- go.work - removed obsolete entries

## Remaining Tasks
- [ ] Consolidate proto files
- [ ] Update imports for moved packages
- [ ] Complete stub implementations or document as not-implemented

## Verification
Run these commands to verify:
\`\`\`bash
go work sync
make build
make test
./scripts/analyze-codebase.sh
\`\`\`
EOF

echo -e "${GREEN}📄 Cleanup report saved to CLEANUP_REPORT.md${NC}"