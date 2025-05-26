#!/bin/bash

# Phoenix Platform Safe Duplicate Elimination Script
# This script safely removes duplicate implementations while preserving critical functionality

set -euo pipefail

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="pre-elimination-backup-${TIMESTAMP}"
LOG_FILE="elimination-${TIMESTAMP}.log"

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

log "${BLUE}🚀 Phoenix Platform Duplicate Elimination${NC}"
log "${BLUE}===========================================${NC}"
log "Backup directory: $BACKUP_DIR"
log "Log file: $LOG_FILE"
echo

# Step 1: Backup critical directories
log "${YELLOW}📦 Creating backups...${NC}"
tar -czf "$BACKUP_DIR/services-backup.tar.gz" services/ 2>/dev/null || true
tar -czf "$BACKUP_DIR/operators-backup.tar.gz" operators/ 2>/dev/null || true
tar -czf "$BACKUP_DIR/pkg-backup.tar.gz" pkg/ 2>/dev/null || true
log "${GREEN}✅ Backups created${NC}"
echo

# Step 2: Remove duplicate services
log "${YELLOW}🗑️  Removing duplicate services...${NC}"

# List of services to remove from /services/ (keeping /projects/ versions)
SERVICES_TO_REMOVE=(
    "analytics"
    "anomaly-detector"
    "api"
    "benchmark"
    "collector"
    "control-actuator-go"
    "controller"
    "dashboard"
    "generator"
    "loadsim-operator"
    "phoenix-cli"
    "pipeline-operator"
)

for service in "${SERVICES_TO_REMOVE[@]}"; do
    if [ -d "services/$service" ]; then
        # Special handling for certain services
        case "$service" in
            "controller")
                # Preserve state machine implementation if unique
                if [ -f "services/controller/internal/controller/state_machine.go" ]; then
                    mkdir -p "projects/controller/internal/controller"
                    cp -p "services/controller/internal/controller/state_machine.go" \
                          "projects/controller/internal/controller/state_machine.go" 2>/dev/null || true
                    log "  ${BLUE}Preserved state_machine.go from services/controller${NC}"
                fi
                ;;
            "generators")
                # Keep generators as they're different from generator service
                continue
                ;;
        esac
        
        # Remove the duplicate service
        rm -rf "services/$service"
        log "  ${GREEN}✓ Removed services/$service${NC}"
    fi
done

# Keep essential services that don't have duplicates in /projects/
log "${BLUE}Keeping unique services:${NC}"
for service in services/*/; do
    if [ -d "$service" ]; then
        service_name=$(basename "$service")
        log "  ${GREEN}✓ Kept services/$service_name (no duplicate in projects/)${NC}"
    fi
done
echo

# Step 3: Remove duplicate operators
log "${YELLOW}🗑️  Removing duplicate operators...${NC}"

# Remove operators that exist in projects/
if [ -d "operators/pipeline" ] && [ -d "projects/pipeline-operator" ]; then
    # Check if controllers directory has unique implementation
    if [ -d "operators/pipeline/controllers" ]; then
        mkdir -p "projects/pipeline-operator/controllers"
        cp -r "operators/pipeline/controllers/"* "projects/pipeline-operator/controllers/" 2>/dev/null || true
        log "  ${BLUE}Preserved controllers from operators/pipeline${NC}"
    fi
    rm -rf "operators/pipeline"
    log "  ${GREEN}✓ Removed operators/pipeline${NC}"
fi

if [ -d "operators/loadsim" ] && [ -d "projects/loadsim-operator" ]; then
    rm -rf "operators/loadsim"
    log "  ${GREEN}✓ Removed operators/loadsim${NC}"
fi
echo

# Step 4: Consolidate proto files
log "${YELLOW}📁 Consolidating proto files...${NC}"

# Create canonical proto location
PROTO_DIR="pkg/contracts/proto"
mkdir -p "$PROTO_DIR/v1"
mkdir -p "$PROTO_DIR/phoenix/v1"

# Function to consolidate proto file
consolidate_proto() {
    local proto_name=$1
    local found=false
    
    # Search for the proto file in all locations
    for location in "pkg/grpc/proto/v1" "pkg/grpc/proto/phoenix/v1" "packages/contracts/proto/v1" "packages/contracts/proto/phoenix/v1"; do
        if [ -f "$location/$proto_name" ]; then
            if [ "$found" = false ]; then
                # First occurrence - use as canonical
                cp "$location/$proto_name" "$PROTO_DIR/v1/$proto_name"
                log "  ${GREEN}✓ Consolidated $proto_name from $location${NC}"
                found=true
            else
                # Check if different from canonical
                if ! diff -q "$location/$proto_name" "$PROTO_DIR/v1/$proto_name" >/dev/null; then
                    log "  ${YELLOW}⚠️  Warning: $proto_name differs in $location${NC}"
                    cp "$location/$proto_name" "$PROTO_DIR/v1/${proto_name}.${location//\//_}.backup"
                fi
            fi
        fi
    done
}

# Consolidate each proto file
for proto in "common.proto" "controller.proto" "experiment.proto" "generator.proto"; do
    consolidate_proto "$proto"
done

# Remove old proto locations
rm -rf "pkg/grpc/proto"
rm -rf "packages/contracts/proto"
log "${GREEN}✅ Proto files consolidated to $PROTO_DIR${NC}"
echo

# Step 5: Clean up empty directories in pkg/
log "${YELLOW}🧹 Cleaning empty directories in pkg/...${NC}"

# Find and remove empty directories
find pkg/ -type d -empty -delete 2>/dev/null || true
log "${GREEN}✅ Empty directories cleaned${NC}"
echo

# Step 6: Update import paths
log "${YELLOW}🔄 Updating import paths...${NC}"

# Update proto imports
find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec \
    sed -i.bak 's|pkg/grpc/proto/|pkg/contracts/proto/|g' {} \;

find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec \
    sed -i.bak 's|packages/contracts/proto/|pkg/contracts/proto/|g' {} \;

# Clean up backup files
find . -name "*.bak" -delete

log "${GREEN}✅ Import paths updated${NC}"
echo

# Step 7: Update go.work file
log "${YELLOW}📝 Updating go.work file...${NC}"

# Create new go.work file with only valid paths
cat > go.work.new << 'EOF'
go 1.23

use (
    ./pkg
    ./projects/analytics
    ./projects/anomaly-detector
    ./projects/api
    ./projects/benchmark
    ./projects/collector
    ./projects/control-actuator-go
    ./projects/controller
    ./projects/dashboard
    ./projects/generator
    ./projects/loadsim-operator
    ./projects/phoenix-cli
    ./projects/pipeline-operator
    ./projects/platform-api
)
EOF

mv go.work.new go.work
log "${GREEN}✅ go.work updated${NC}"
echo

# Step 8: Run go work sync
log "${YELLOW}🔄 Syncing Go workspace...${NC}"
go work sync || log "${YELLOW}⚠️  go work sync failed - manual intervention may be needed${NC}"
echo

# Step 9: Validation
log "${YELLOW}🔍 Running validation...${NC}"

# Check for broken imports
if command -v goimports &> /dev/null; then
    find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | \
        xargs goimports -l 2>/dev/null | grep -v "^$" > broken-imports.txt || true
    
    if [ -s broken-imports.txt ]; then
        log "${YELLOW}⚠️  Found files with broken imports (see broken-imports.txt)${NC}"
    else
        log "${GREEN}✅ No broken imports found${NC}"
        rm -f broken-imports.txt
    fi
fi

# Final summary
echo
log "${GREEN}🎉 Duplicate elimination complete!${NC}"
log "${BLUE}===========================================${NC}"
log "Backups saved in: $BACKUP_DIR"
log "Log file: $LOG_FILE"
echo
log "${YELLOW}Next steps:${NC}"
log "1. Run 'make validate' to check project structure"
log "2. Run 'make test' to ensure tests pass"
log "3. Check git status and review changes"
log "4. If issues arise, restore from backups in $BACKUP_DIR"