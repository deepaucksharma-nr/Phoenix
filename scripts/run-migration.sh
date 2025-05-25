#!/bin/bash
# run-migration.sh - Orchestrate the complete migration process
# Usage: ./run-migration.sh [--phase <phase-number>] [--dry-run]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Parse arguments
PHASE=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --phase)
            PHASE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--phase <phase-number>] [--dry-run]"
            exit 1
            ;;
    esac
done

# Migration configuration
declare -A PHASE1_SERVICES=(
    ["anomaly-detector"]="apps/anomaly-detector go"
    ["control-actuator"]="apps/control-actuator-go go"
)

declare -A PHASE2_SERVICES=(
    ["platform-api"]="phoenix-platform/cmd/api-gateway go"
    ["control-service"]="phoenix-platform/cmd/control-service go"
    ["experiment-controller"]="phoenix-platform/cmd/controller go"
    ["config-generator"]="phoenix-platform/cmd/generator go"
    ["phoenix-cli"]="phoenix-platform/cmd/phoenix-cli go"
    ["web-dashboard"]="phoenix-platform/dashboard react"
)

declare -A PHASE3_SERVICES=(
    ["analytics-engine"]="services/analytics go"
    ["benchmark-service"]="services/benchmark go"
    ["telemetry-collector"]="services/collector node"
    ["config-validator"]="services/validator go"
)

declare -A PHASE4_SERVICES=(
    ["loadsim-operator"]="phoenix-platform/operators/loadsim go"
    ["pipeline-operator"]="phoenix-platform/operators/pipeline go"
    ["process-simulator"]="phoenix-platform/cmd/simulator go"
)

# Functions
log_phase() {
    echo ""
    echo -e "${MAGENTA}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${MAGENTA}  $1${NC}"
    echo -e "${MAGENTA}════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

log_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
}

execute_cmd() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}[DRY RUN] Would execute: $1${NC}"
    else
        eval "$1"
    fi
}

# Migration functions
setup_foundation() {
    log_phase "PHASE 0: Setting Up Foundation"
    
    log_info "Creating directory structure..."
    execute_cmd "mkdir -p build/makefiles build/docker/base build/scripts"
    execute_cmd "mkdir -p deployments/{kubernetes,helm,terraform,ansible}"
    execute_cmd "mkdir -p pkg tests/{integration,e2e,performance,security,contracts}"
    execute_cmd "mkdir -p tools/{dev-env,generators,linters,analyzers,migration}"
    execute_cmd "mkdir -p projects scripts docs configs"
    
    log_info "Creating build infrastructure..."
    if [ ! -f "build/makefiles/common.mk" ] && [ "$DRY_RUN" = false ]; then
        log_error "Build makefiles not found. Please ensure they are created first."
        exit 1
    fi
    
    log_success "Foundation setup complete"
}

migrate_shared_packages() {
    log_phase "PHASE 1: Migrating Shared Packages"
    
    if [ -f "scripts/migrate-shared-packages.sh" ]; then
        log_info "Running shared package migration..."
        execute_cmd "./scripts/migrate-shared-packages.sh"
    else
        log_error "Shared package migration script not found"
        return 1
    fi
    
    log_success "Shared packages migrated"
}

migrate_services() {
    local phase_name=$1
    shift
    local -n services=$1
    
    log_phase "$phase_name"
    
    for service in "${!services[@]}"; do
        IFS=' ' read -r old_path service_type <<< "${services[$service]}"
        
        log_info "Migrating $service..."
        
        if [ -f "scripts/migrate-service.sh" ]; then
            execute_cmd "./scripts/migrate-service.sh '$service' '$old_path' '$service_type'"
            
            # Validate the migration
            if [ "$DRY_RUN" = false ]; then
                if ./scripts/validate-migration.sh "$service" > /dev/null 2>&1; then
                    log_success "$service migrated successfully"
                else
                    log_error "$service migration validation failed"
                fi
            fi
        else
            log_error "Service migration script not found"
            return 1
        fi
        
        # Add a small delay between migrations
        sleep 1
    done
}

migrate_configurations() {
    log_phase "PHASE 5: Migrating Configurations"
    
    log_info "Migrating environment configurations..."
    execute_cmd "cp -r OLD_IMPLEMENTATION/configs/production configs/ 2>/dev/null || true"
    execute_cmd "cp -r OLD_IMPLEMENTATION/configs/staging configs/ 2>/dev/null || true"
    execute_cmd "mkdir -p configs/development"
    
    log_info "Migrating monitoring configurations..."
    execute_cmd "mkdir -p deployments/kubernetes/base/infrastructure/{prometheus,grafana}"
    execute_cmd "cp -r OLD_IMPLEMENTATION/configs/monitoring/prometheus/* deployments/kubernetes/base/infrastructure/prometheus/ 2>/dev/null || true"
    execute_cmd "cp -r OLD_IMPLEMENTATION/configs/monitoring/grafana/* deployments/kubernetes/base/infrastructure/grafana/ 2>/dev/null || true"
    
    log_success "Configurations migrated"
}

migrate_deployment_infrastructure() {
    log_phase "PHASE 6: Migrating Deployment Infrastructure"
    
    log_info "Migrating Kubernetes resources..."
    execute_cmd "cp -r OLD_IMPLEMENTATION/phoenix-platform/k8s/crds deployments/kubernetes/operators/ 2>/dev/null || true"
    execute_cmd "cp -r OLD_IMPLEMENTATION/phoenix-platform/k8s/base/* deployments/kubernetes/base/ 2>/dev/null || true"
    execute_cmd "cp -r OLD_IMPLEMENTATION/phoenix-platform/k8s/overlays/* deployments/kubernetes/overlays/ 2>/dev/null || true"
    
    log_info "Migrating Helm charts..."
    execute_cmd "cp -r OLD_IMPLEMENTATION/phoenix-platform/helm/* deployments/helm/ 2>/dev/null || true"
    
    log_info "Migrating Docker configurations..."
    execute_cmd "cp OLD_IMPLEMENTATION/docker-compose.yaml docker-compose.yml 2>/dev/null || true"
    execute_cmd "cp OLD_IMPLEMENTATION/docker-compose.prod.yml . 2>/dev/null || true"
    
    log_success "Deployment infrastructure migrated"
}

migrate_documentation() {
    log_phase "PHASE 7: Migrating Documentation"
    
    log_info "Migrating architecture documentation..."
    execute_cmd "mkdir -p docs/architecture/{decisions,diagrams,patterns}"
    execute_cmd "cp -r OLD_IMPLEMENTATION/phoenix-platform/docs/adr/* docs/architecture/decisions/ 2>/dev/null || true"
    
    log_info "Migrating guides..."
    execute_cmd "mkdir -p docs/guides/{developer,operator,user}"
    execute_cmd "cp OLD_IMPLEMENTATION/phoenix-platform/docs/*_GUIDE.md docs/guides/ 2>/dev/null || true"
    
    log_info "Migrating runbooks..."
    execute_cmd "cp -r OLD_IMPLEMENTATION/runbooks/* docs/runbooks/ 2>/dev/null || true"
    
    log_success "Documentation migrated"
}

run_final_validation() {
    log_phase "FINAL VALIDATION"
    
    if [ "$DRY_RUN" = false ]; then
        log_info "Running complete validation..."
        if ./scripts/validate-migration.sh; then
            log_success "Migration validation passed!"
        else
            log_error "Migration validation failed"
            return 1
        fi
    else
        log_info "[DRY RUN] Would run final validation"
    fi
}

# Main execution
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Phoenix Platform Migration Runner                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}Running in DRY RUN mode - no changes will be made${NC}"
    echo ""
fi

# Execute phases based on argument
case "$PHASE" in
    "0")
        setup_foundation
        ;;
    "1")
        migrate_shared_packages
        ;;
    "2")
        migrate_services "PHASE 2: Core Services Migration" PHASE2_SERVICES
        ;;
    "3")
        migrate_services "PHASE 3: Supporting Services Migration" PHASE3_SERVICES
        ;;
    "4")
        migrate_services "PHASE 4: Operators and Tools Migration" PHASE4_SERVICES
        ;;
    "5")
        migrate_configurations
        ;;
    "6")
        migrate_deployment_infrastructure
        ;;
    "7")
        migrate_documentation
        ;;
    "")
        # Run all phases
        setup_foundation
        migrate_shared_packages
        migrate_services "PHASE 2: Core Services Migration" PHASE2_SERVICES
        migrate_services "PHASE 3: Supporting Services Migration" PHASE3_SERVICES
        migrate_services "PHASE 4: Operators and Tools Migration" PHASE4_SERVICES
        migrate_configurations
        migrate_deployment_infrastructure
        migrate_documentation
        run_final_validation
        ;;
    *)
        log_error "Invalid phase: $PHASE"
        echo "Valid phases: 0-7"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Migration ${PHASE:+Phase $PHASE }Complete!${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"

if [ -z "$PHASE" ] && [ "$DRY_RUN" = false ]; then
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "1. Review the MIGRATION_PLAN.md for detailed information"
    echo "2. Test individual services: cd projects/<service> && make test"
    echo "3. Update CI/CD pipelines to use the new structure"
    echo "4. Remove OLD_IMPLEMENTATION directory when ready"
    echo "5. Commit the changes to the repository"
fi