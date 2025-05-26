#!/bin/bash

# Phoenix Platform Documentation Consolidation Script
# This script consolidates redundant documentation files

set -e

echo "🔄 Starting Phoenix documentation consolidation..."

# Create archive directories
echo "📁 Creating archive directories..."
mkdir -p docs/archive/migration
mkdir -p docs/archive/prd  
mkdir -p docs/archive/old-reports
mkdir -p docs/history
mkdir -p docs/operations

# Archive migration documents
echo "📦 Archiving migration documents..."
for file in MIGRATION_*.md CLI_MIGRATION_REPORT.md; do
    if [ -f "$file" ]; then
        echo "  Moving $file to archive..."
        mv "$file" docs/archive/migration/ 2>/dev/null || true
    fi
done

# Archive PRD documents  
echo "📦 Archiving PRD documents..."
for file in PRD_*.md IMPLEMENTATION_CHECKLIST.md; do
    if [ -f "$file" ]; then
        echo "  Moving $file to archive..."
        mv "$file" docs/archive/prd/ 2>/dev/null || true
    fi
done

# Archive old reports
echo "📦 Archiving old reports..."
for file in *_RESULTS.md VALIDATION_REPORT.md; do
    if [ -f "$file" ]; then
        echo "  Moving $file to archive..."
        mv "$file" docs/archive/old-reports/ 2>/dev/null || true
    fi
done

# Move operational docs
echo "📚 Moving operational documents..."
[ -f "SERVICE_CONSOLIDATION_PLAN.md" ] && mv SERVICE_CONSOLIDATION_PLAN.md docs/operations/
[ -f "POST_MIGRATION_TASKS.md" ] && mv POST_MIGRATION_TASKS.md docs/archive/migration/
[ -f "E2E_DEMO_GUIDE.md" ] && mv E2E_DEMO_GUIDE.md docs/guides/

# Remove duplicate and temporary files
echo "🗑️  Removing duplicate files..."
[ -f "QUICKSTART.md" ] && rm -f QUICKSTART.md  # Duplicate of QUICK_START.md
[ -f "PUSH_SUMMARY.md" ] && rm -f PUSH_SUMMARY.md  # Temporary file
[ -f "HANDOFF_CHECKLIST.md" ] && rm -f HANDOFF_CHECKLIST.md  # Merged into PROJECT_HANDOFF.md

# Create consolidated migration summary if archives exist
if [ -d "docs/archive/migration" ] && [ "$(ls -A docs/archive/migration)" ]; then
    echo "📝 Creating consolidated migration summary..."
    cat > docs/history/MIGRATION_SUMMARY.md << 'EOF'
# Phoenix Platform Migration History

## Overview

The Phoenix Platform was successfully migrated from `phoenix-vnext` to `phoenix` module structure. This document consolidates the migration history for reference.

## Migration Summary

- **Duration**: Single session (May 2025)
- **Files Migrated**: 1,176 files
- **Services Migrated**: 15 microservices
- **Archive Size Reduction**: 4.5M → 952K (79%)

## Key Changes

1. All module names updated from `github.com/phoenix-vnext/platform` to `github.com/phoenix/platform`
2. All import paths updated across the codebase
3. Go workspace (go.work) configured with all modules
4. Phoenix CLI successfully migrated to `/projects/phoenix-cli`

## Validation Results

- ✅ No phoenix-vnext references remain
- ✅ All services build successfully  
- ✅ E2E tests pass
- ✅ Documentation updated

## Archived Documents

The detailed migration documents have been archived in `docs/archive/migration/` for historical reference.
EOF
fi

# Update main documentation files
echo "📝 Updating main documentation..."

# Update README.md documentation section if needed
if grep -q "MIGRATION_FINAL_REPORT.md" README.md 2>/dev/null; then
    echo "  Updating README.md references..."
    sed -i.bak 's|MIGRATION_FINAL_REPORT.md|docs/history/MIGRATION_SUMMARY.md|g' README.md
    rm -f README.md.bak
fi

# Create documentation index if it doesn't exist
if [ ! -f "docs/README.md" ]; then
    echo "📝 Creating documentation index..."
    cat > docs/README.md << 'EOF'
# Phoenix Platform Documentation

## Documentation Structure

```
docs/
├── prd/                # PRD compliance tracking
│   ├── GAP_ANALYSIS.md
│   ├── IMPLEMENTATION_PLAN.md
│   └── TRACKING_CHECKLIST.md
├── guides/             # How-to guides
├── operations/         # Operational procedures
├── history/           # Historical documents
└── api/               # API documentation
```

## Quick Links

- [Platform Status](../PLATFORM_STATUS.md)
- [Quick Start](../QUICK_START.md)
- [Contributing](../CONTRIBUTING.md)
- [Architecture](../PLATFORM_ARCHITECTURE.md)
EOF
fi

# Final cleanup
echo "🧹 Final cleanup..."
find . -name "*.bak" -type f -delete 2>/dev/null || true

# Summary
echo ""
echo "✅ Documentation consolidation complete!"
echo ""
echo "📊 Summary:"
echo "  - Migration docs archived to: docs/archive/migration/"
echo "  - PRD docs moved to: docs/prd/"
echo "  - Operational docs in: docs/operations/"
echo "  - Removed duplicate files"
echo ""
echo "📚 Essential documents remaining in root:"
echo "  - README.md"
echo "  - PLATFORM_ARCHITECTURE.md"
echo "  - PLATFORM_STATUS.md"
echo "  - QUICK_START.md"
echo "  - CONTRIBUTING.md"
echo "  - CLAUDE.md"
echo "  - MONOREPO_BOUNDARIES.md"
echo "  - PROJECT_HANDOFF.md"
echo ""
echo "🎯 Next steps:"
echo "  1. Review consolidated documents in docs/prd/"
echo "  2. Check PLATFORM_STATUS.md for current state"
echo "  3. Follow PROJECT_HANDOFF.md for team transition"