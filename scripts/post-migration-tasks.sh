#!/bin/bash
# post-migration-tasks.sh - Run post-migration tasks and generate report

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform Post-Migration Tasks ===${NC}"
echo ""

# Task tracking
declare -a COMPLETED_TASKS=()
declare -a FAILED_TASKS=()

# Function to run a task
run_task() {
    local task_name="$1"
    local task_cmd="$2"
    
    echo -ne "${YELLOW}Running:${NC} $task_name... "
    
    if eval "$task_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        COMPLETED_TASKS+=("$task_name")
    else
        echo -e "${RED}✗${NC}"
        FAILED_TASKS+=("$task_name")
    fi
}

# 1. Sync Go modules
echo -e "${BLUE}1. Go Module Maintenance${NC}"
run_task "Sync Go workspace" "go work sync"
run_task "Download dependencies" "go mod download"

# 2. Generate documentation
echo ""
echo -e "${BLUE}2. Documentation Generation${NC}"
if [[ ! -d docs/generated ]]; then
    mkdir -p docs/generated
fi

# Create service inventory
cat > docs/generated/SERVICE_INVENTORY.md << 'EOF'
# Phoenix Platform Service Inventory

Generated: $(date)

## Projects (Migrated Services)

| Service | Type | Language | Status |
|---------|------|----------|--------|
EOF

for project in projects/*/; do
    if [[ -d "$project" ]]; then
        service_name=$(basename "$project")
        lang="Go"
        if [[ -f "$project/package.json" ]]; then
            lang="JavaScript/TypeScript"
        fi
        echo "| $service_name | Microservice | $lang | ✅ Migrated |" >> docs/generated/SERVICE_INVENTORY.md
    fi
done

cat >> docs/generated/SERVICE_INVENTORY.md << 'EOF'

## Legacy Services (Not Yet Migrated)

| Service | Type | Language | Status |
|---------|------|----------|--------|
| validator | Microservice | Go | ⏳ Pending |
| generators/complex | Tool | Bash | ⏳ Pending |
| generators/synthetic | Generator | Go | ⏳ Pending |
| control-plane/observer | Monitor | JavaScript | ⏳ Pending |
EOF

run_task "Generate service inventory" "[[ -f docs/generated/SERVICE_INVENTORY.md ]]"

# 3. Create development shortcuts
echo ""
echo -e "${BLUE}3. Development Shortcuts${NC}"

# Create a quick-start script
cat > scripts/quick-start.sh << 'EOF'
#!/bin/bash
# quick-start.sh - Quick start development environment

echo "🚀 Phoenix Platform Quick Start"
echo ""

# Check Docker
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check Go
if ! go version > /dev/null 2>&1; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

echo "✅ Prerequisites checked"
echo ""

# Setup local environment
echo "Setting up local development environment..."
./scripts/setup-dev-env.sh

echo ""
echo "✅ Development environment ready!"
echo ""
echo "Available commands:"
echo "  make dev        - Start all services locally"
echo "  make test       - Run all tests"
echo "  make build      - Build all services"
echo "  make validate   - Run validation checks"
echo ""
echo "To deploy to Kubernetes:"
echo "  ./scripts/deploy-dev.sh"
EOF

chmod +x scripts/quick-start.sh
run_task "Create quick-start script" "[[ -x scripts/quick-start.sh ]]"

# 4. Update git hooks
echo ""
echo -e "${BLUE}4. Git Hooks${NC}"

mkdir -p .git/hooks

cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for Phoenix Platform

echo "Running pre-commit checks..."

# Check for cross-project imports
if ! ./scripts/validate-boundaries.sh > /dev/null 2>&1; then
    echo "❌ Cross-project import violations detected!"
    echo "Run ./scripts/validate-boundaries.sh for details"
    exit 1
fi

# Run go fmt
echo "Running go fmt..."
go fmt ./...

# Run go vet
echo "Running go vet..."
go vet ./...

echo "✅ Pre-commit checks passed"
EOF

chmod +x .git/hooks/pre-commit
run_task "Install pre-commit hook" "[[ -x .git/hooks/pre-commit ]]"

# 5. Create migration rollback plan
echo ""
echo -e "${BLUE}5. Rollback Plan${NC}"

cat > docs/ROLLBACK_PLAN.md << 'EOF'
# Migration Rollback Plan

If you need to rollback the migration:

## 1. Restore OLD_IMPLEMENTATION
```bash
# Extract the archive
tar -xzf archives/OLD_IMPLEMENTATION-*.tar.gz

# Remove new structure (BE CAREFUL!)
rm -rf projects/ packages/
```

## 2. Restore go.mod files
```bash
# The archive contains all original go.mod files
# No additional action needed
```

## 3. Update imports
```bash
# Imports in OLD_IMPLEMENTATION use original paths
# No changes needed
```

## 4. Cleanup
```bash
# Remove new scripts
rm -rf scripts/deploy-dev.sh scripts/setup-dev-env.sh
rm -rf scripts/validate-*.sh
```

## ⚠️ Warning
Rollback should only be used in emergencies. The new structure provides:
- Better modularity
- Improved security
- Easier maintenance
- Clearer boundaries
EOF

run_task "Create rollback plan" "[[ -f docs/ROLLBACK_PLAN.md ]]"

# 6. Generate final report
echo ""
echo -e "${BLUE}6. Final Report${NC}"

REPORT_FILE="MIGRATION_REPORT_$(date +%Y%m%d_%H%M%S).txt"

cat > "$REPORT_FILE" << EOF
Phoenix Platform Migration Report
=================================
Generated: $(date)

Migration Summary
-----------------
- Total Services Migrated: $(ls -1 projects/ | wc -l | tr -d ' ')
- Shared Packages Created: $(ls -1 packages/ | wc -l | tr -d ' ')
- Archive Size: $(ls -lh archives/OLD_IMPLEMENTATION-*.tar.gz | awk '{print $5}')
- Git Commits: $(git rev-list --count HEAD)

Completed Tasks
---------------
EOF

for task in "${COMPLETED_TASKS[@]}"; do
    echo "✓ $task" >> "$REPORT_FILE"
done

if [[ ${#FAILED_TASKS[@]} -gt 0 ]]; then
    echo "" >> "$REPORT_FILE"
    echo "Failed Tasks" >> "$REPORT_FILE"
    echo "------------" >> "$REPORT_FILE"
    for task in "${FAILED_TASKS[@]}"; do
        echo "✗ $task" >> "$REPORT_FILE"
    done
fi

cat >> "$REPORT_FILE" << EOF

Next Steps
----------
1. Run quick start: ./scripts/quick-start.sh
2. Deploy to dev: ./scripts/deploy-dev.sh
3. Run E2E tests: Follow E2E_DEMO_GUIDE.md
4. Update CI/CD pipelines

Resources
---------
- Documentation: README.md, CLAUDE.md
- Migration Summary: MIGRATION_SUMMARY.md
- Service Inventory: docs/generated/SERVICE_INVENTORY.md
- Rollback Plan: docs/ROLLBACK_PLAN.md
EOF

echo ""
echo -e "${GREEN}=== Post-Migration Tasks Complete ===${NC}"
echo ""
echo "Summary:"
echo "- Completed tasks: ${#COMPLETED_TASKS[@]}"
echo "- Failed tasks: ${#FAILED_TASKS[@]}"
echo "- Report saved to: $REPORT_FILE"
echo ""
echo "Quick start development:"
echo "  ${GREEN}./scripts/quick-start.sh${NC}"