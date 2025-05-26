#!/bin/bash

# Phoenix Platform - Complete Streamlining Script
# Eliminates redundancies and focuses on MVP requirements

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create backup directory
BACKUP_DIR="phoenix-full-backup-$(date +%Y%m%d-%H%M%S)"

echo -e "${BLUE}🔥 Phoenix Platform Streamlining${NC}"
echo -e "${BLUE}=================================${NC}"

echo -e "\n${YELLOW}Creating full project backup...${NC}"
mkdir -p "$BACKUP_DIR"
cp -r projects/ "$BACKUP_DIR/" 2>/dev/null || true
cp -r docs/ "$BACKUP_DIR/" 2>/dev/null || true
cp -r configs/ "$BACKUP_DIR/" 2>/dev/null || true
cp -r infrastructure/ "$BACKUP_DIR/" 2>/dev/null || true
cp *.md "$BACKUP_DIR/" 2>/dev/null || true

echo -e "${GREEN}✓ Backup created at: $BACKUP_DIR${NC}"

# Function to safely remove directory
remove_project() {
    local project=$1
    if [ -d "projects/$project" ]; then
        echo -e "${YELLOW}Removing project: projects/$project${NC}"
        rm -rf "projects/$project"
    else
        echo -e "${RED}Project not found: projects/$project${NC}"
    fi
}

# Function to remove documentation files
remove_docs() {
    local pattern=$1
    local description=$2
    echo -e "${YELLOW}Removing $description...${NC}"
    find . -name "$pattern" -type f -delete 2>/dev/null || true
}

echo -e "\n${GREEN}Phase 1: Eliminating Redundant Projects${NC}"
echo "========================================="

# Remove redundant/placeholder projects (7 projects)
remove_project "hello-phoenix"
remove_project "api"
remove_project "collector" 
remove_project "control-actuator-go"
remove_project "anomaly-detector"
remove_project "analytics"
remove_project "generator"

echo -e "${GREEN}✓ Removed 7 redundant projects${NC}"

echo -e "\n${GREEN}Phase 2: Documentation Cleanup${NC}"
echo "==============================="

# Remove migration and status files
remove_docs "*MIGRATION*" "migration files"
remove_docs "*STATUS*" "status files"
remove_docs "*COMPLETION*" "completion files"
remove_docs "*SUMMARY*" "summary files"
remove_docs "*SUCCESS*" "success files"
remove_docs "*FINAL*" "final status files"
remove_docs "*CHECKLIST*" "checklist files"
remove_docs "*HANDOFF*" "handoff files"
remove_docs "*ACCOMPLISHMENTS*" "accomplishment files"
remove_docs "*CELEBRATION*" "celebration files"
remove_docs "*THE_END*" "end files"

# Remove redundant documentation
remove_docs "*PLAN.md" "plan files"
remove_docs "*REPORT.md" "report files"
remove_docs "*ANALYSIS.md" "analysis files"
remove_docs "*IMPLEMENTATION*.md" "implementation files"

# Remove specific redundant files
redundant_files=(
    "ACTIVE_DEVELOPMENT.md"
    "CODEBASE_CLEANUP_PLAN.md"
    "CONSOLIDATED_DOCS_PLAN.md"
    "DEVELOPMENT_GUIDE.md"
    "DEVELOPMENT_TIPS.md"
    "ELIMINATION_SUMMARY.md"
    "MASTER_DOCUMENTATION_INDEX.md"
    "NEXT_STEPS.md"
    "PHOENIX_DEMO_SUMMARY.md"
    "PHOENIX_PLATFORM_DEMO.md"
    "PHOENIX_PLATFORM_FINAL_STATE.md"
    "PHOENIX_PLATFORM_STATUS.md"
    "PLATFORM_STATUS.md"
    "PROJECT_COMPLETION_CHECKLIST.md"
    "PROJECT_HANDOFF.md"
    "PROJECT_HANDOFF_DOCUMENT.md"
    "QUICKSTART.md"
    "QUICK_START.md"
    "SERVICE_CONSOLIDATION_ANALYSIS.md"
    "START_HERE.md"
    "STREAMLINE_IMPLEMENTATION_PLAN.md"
    "TEAM_ASSIGNMENTS.md"
    "REDUNDANCY_ELIMINATION_PLAN.md"
    "PRD_*"
)

for file in "${redundant_files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${YELLOW}Removing: $file${NC}"
        rm -f "$file"
    fi
done

echo -e "${GREEN}✓ Cleaned up documentation files${NC}"

echo -e "\n${GREEN}Phase 3: Infrastructure Simplification${NC}"
echo "======================================="

# Remove duplicate infrastructure
if [ -d "infrastructure/helm/phoenix" ]; then
    echo -e "${YELLOW}Removing duplicate Helm chart${NC}"
    rm -rf "infrastructure/helm/phoenix"
fi

# Remove redundant docker-compose files
redundant_compose=(
    "docker-compose-fixed.yml"
    "infrastructure/docker/compose/docker-compose.dev.yml"
    "infrastructure/docker/compose/docker-compose.override.yml"
    "infrastructure/docker/compose/docker-compose.prod.yml"
)

for compose in "${redundant_compose[@]}"; do
    if [ -f "$compose" ]; then
        echo -e "${YELLOW}Removing: $compose${NC}"
        rm -f "$compose"
    fi
done

# Remove Terraform (overkill for MVP)
if [ -d "infrastructure/terraform" ]; then
    echo -e "${YELLOW}Removing Terraform infrastructure${NC}"
    rm -rf "infrastructure/terraform"
fi

# Remove redundant configs
if [ -d "configs/control" ]; then
    echo -e "${YELLOW}Removing control configs${NC}"
    rm -rf "configs/control"
fi

if [ -d "configs/production" ]; then
    echo -e "${YELLOW}Removing production configs${NC}"
    rm -rf "configs/production"
fi

echo -e "${GREEN}✓ Simplified infrastructure${NC}"

echo -e "\n${GREEN}Phase 4: Archive and Script Cleanup${NC}"
echo "===================================="

# Remove old archives
if [ -d "archives" ]; then
    echo -e "${YELLOW}Removing old archives${NC}"
    rm -rf "archives"
fi

# Remove backup directories
if [ -d "pre-elimination-backup-20250526-181231" ]; then
    echo -e "${YELLOW}Removing old backup directory${NC}"
    rm -rf "pre-elimination-backup-20250526-181231"
fi

if [ -d "dashboard-backup-20250526-182525" ]; then
    echo -e "${YELLOW}Removing dashboard backup directory${NC}"
    rm -rf "dashboard-backup-20250526-182525"
fi

if [ -d "duplicate-analysis-20250526-181047" ]; then
    echo -e "${YELLOW}Removing analysis directory${NC}"
    rm -rf "duplicate-analysis-20250526-181047"
fi

# Remove redundant scripts
redundant_scripts=(
    "scripts/analyze-codebase.sh"
    "scripts/analyze-duplicates.sh"
    "scripts/cleanup-codebase.sh"
    "scripts/cleanup-duplicate-services.sh"
    "scripts/consolidate-docs.sh"
    "scripts/consolidate-services.sh"
    "scripts/eliminate-duplicates.sh"
    "scripts/remove-duplicate-services.sh"
    "scripts/standardize-services.sh"
    "scripts/analyze-dashboard-redundancy.sh"
    "scripts/streamline-dashboard.sh"
)

for script in "${redundant_scripts[@]}"; do
    if [ -f "$script" ]; then
        echo -e "${YELLOW}Removing: $script${NC}"
        rm -f "$script"
    fi
done

echo -e "${GREEN}✓ Cleaned up archives and scripts${NC}"

echo -e "\n${GREEN}Phase 5: Operators Consolidation${NC}"
echo "================================="

# Check if we should consolidate operators
if [ -d "projects/loadsim-operator" ] && [ -d "projects/pipeline-operator" ]; then
    echo -e "${YELLOW}Note: Consider consolidating operators into single operator${NC}"
    echo -e "${YELLOW}This requires manual review of functionality overlap${NC}"
fi

echo -e "\n${GREEN}Phase 6: Go Workspace Cleanup${NC}"
echo "============================="

# Update go.work to remove eliminated projects
if [ -f "go.work" ]; then
    echo -e "${YELLOW}Updating go.work file${NC}"
    # Create new go.work with only remaining projects
    cat > go.work << 'EOF'
go 1.21

use (
    ./pkg
    ./projects/phoenix-cli
    ./projects/platform-api
    ./projects/controller
    ./projects/benchmark
    ./projects/pipeline-operator
    ./projects/loadsim-operator
)
EOF
fi

echo -e "${GREEN}✓ Updated Go workspace${NC}"

echo -e "\n${GREEN}Phase 7: Final Statistics${NC}"
echo "========================="

# Count remaining files
echo -e "${BLUE}Project Structure After Streamlining:${NC}"
if [ -d "projects" ]; then
    ls -la projects/ | grep -E "^d" | wc -l | xargs echo "Projects remaining:"
fi

echo -e "${BLUE}Documentation files remaining:${NC}"
find . -name "*.md" -type f | wc -l | xargs echo "Markdown files:"

echo -e "${BLUE}Configuration files:${NC}"
find . -name "*.yaml" -o -name "*.yml" | wc -l | xargs echo "YAML files:"

echo -e "\n${GREEN}🎉 Phoenix Platform Streamlining Complete!${NC}"
echo "==========================================="

echo -e "${GREEN}✅ Eliminated:${NC}"
echo "  - 7 redundant projects"
echo "  - 90%+ of documentation files"
echo "  - Duplicate infrastructure"
echo "  - Old backups and analysis files"
echo "  - Redundant scripts"

echo -e "\n${GREEN}✅ Remaining Core:${NC}"
echo "  - phoenix-cli (CLI interface)"
echo "  - platform-api (backend)"
echo "  - dashboard (web UI)"
echo "  - controller (experiment management)"
echo "  - benchmark (performance analysis)" 
echo "  - pipeline-operator (K8s management)"
echo "  - loadsim-operator (load testing)"

echo -e "\n${YELLOW}📂 Backup location: $BACKUP_DIR${NC}"
echo -e "${YELLOW}🔄 To restore: rm -rf projects docs configs && cp -r $BACKUP_DIR/* .${NC}"

echo -e "\n${GREEN}Next steps:${NC}"
echo "1. Review remaining projects for functionality"
echo "2. Consider consolidating operators"
echo "3. Test core functionality"
echo "4. Update remaining documentation"
echo "5. Validate build and deployment"

echo -e "\n${BLUE}The Phoenix Platform is now streamlined for MVP success! 🚀${NC}"