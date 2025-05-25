#!/bin/bash
# migration-controller.sh - Main migration orchestration controller
# This ensures bulletproof migration execution with full state tracking

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Set agent ID if not provided
export AGENT_ID="${AGENT_ID:-agent-$(hostname)-$$}"

# Source libraries
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/state-tracker.sh"

# Parse command line arguments
COMMAND=${1:-help}
PHASE_ID=${2:-}
DRY_RUN=${DRY_RUN:-false}
FORCE=${FORCE:-false}

# Show usage
show_usage() {
    cat << EOF
Phoenix Migration Controller v2.0

Usage: $0 <command> [options]

Commands:
    init                Initialize migration (must be run first)
    status              Show migration status
    run-phase <id>      Run a specific phase
    run-all            Run complete migration
    validate <phase>    Validate a completed phase
    rollback <phase>    Rollback to a specific phase
    monitor            Monitor migration progress
    report             Generate migration report
    cleanup            Clean up migration state

Options:
    DRY_RUN=true       Run in dry-run mode
    FORCE=true         Force operation (skip confirmations)
    AGENT_ID=<id>      Set agent identifier

Examples:
    $0 init
    $0 run-phase phase-1-packages
    $0 status
    AGENT_ID=agent-1 $0 run-all

EOF
}

# Initialize migration
init_migration() {
    log_phase "Initializing Phoenix Migration"
    
    # Check if already initialized
    if [[ -f .migration/state.yaml ]] && [[ "$FORCE" != "true" ]]; then
        log_warning "Migration already initialized. Use FORCE=true to reinitialize."
        return 1
    fi
    
    # Run pre-flight checks
    log_info "Running pre-flight checks..."
    if ! "$SCRIPT_DIR/pre-flight-checks.sh"; then
        log_error "Pre-flight checks failed. Fix issues before proceeding."
        return 1
    fi
    
    # Initialize state
    init_migration_state
    
    # Create initial rollback point
    create_rollback_point "initialization" "Migration initialized"
    
    # Initialize all phases from manifest
    log_info "Loading migration manifest..."
    if [[ ! -f migration-manifest.yaml ]]; then
        log_error "Migration manifest not found"
        return 1
    fi
    
    # Parse phases from manifest and initialize
    while IFS= read -r phase_id; do
        phase_name=$(grep -A1 "id: \"$phase_id\"" migration-manifest.yaml | grep "name:" | cut -d'"' -f2)
        can_parallelize=$(grep -A2 "id: \"$phase_id\"" migration-manifest.yaml | grep "can_parallelize:" | awk '{print $2}')
        init_phase "$phase_id" "$phase_name" "$can_parallelize"
        log_info "Initialized phase: $phase_id"
    done < <(grep "^  - id:" migration-manifest.yaml | awk '{print $3}' | tr -d '"')
    
    # Record initialization
    update_migration_state "status" "initialized"
    update_migration_state "initialized_by" "$AGENT_ID"
    
    log_success "Migration initialized successfully"
    log_info "Agent ID: $AGENT_ID"
    log_info "Next step: Run '$0 run-phase phase-0-foundation' or '$0 run-all'"
}

# Show migration status
show_status() {
    log_phase "Migration Status"
    
    if [[ ! -f .migration/state.yaml ]]; then
        log_error "Migration not initialized. Run '$0 init' first."
        return 1
    fi
    
    # Show overall status
    echo -e "${YELLOW}Overall Status:${NC}"
    echo -e "Migration ID: $(grep "id:" .migration/state.yaml | head -1 | awk '{print $2}' | tr -d '"')"
    echo -e "Status: $(grep "status:" .migration/state.yaml | head -1 | awk '{print $2}' | tr -d '"')"
    echo -e "Started: $(grep "started_at:" .migration/state.yaml | head -1 | awk '{print $2}' | tr -d '"')"
    echo ""
    
    # Show phase status
    echo -e "${YELLOW}Phase Status:${NC}"
    printf "%-25s %-15s %-20s %s\n" "Phase" "Status" "Assigned To" "Updated"
    printf "%-25s %-15s %-20s %s\n" "-----" "------" "-----------" "-------"
    
    for phase_file in .migration/phases/*.yaml; do
        if [[ -f "$phase_file" ]]; then
            phase_id=$(basename "$phase_file" .yaml)
            phase_name=$(grep "name:" "$phase_file" | head -1 | awk -F'"' '{print $2}')
            status=$(grep "status:" "$phase_file" | head -1 | awk '{print $2}' | tr -d '"')
            assigned=$(grep "assigned_to:" "$phase_file" 2>/dev/null | awk '{print $2}' | tr -d '"' || echo "-")
            
            # Color code status
            case "$status" in
                "completed")
                    status_display="${GREEN}$status${NC}"
                    ;;
                "in_progress")
                    status_display="${YELLOW}$status${NC}"
                    ;;
                "failed")
                    status_display="${RED}$status${NC}"
                    ;;
                *)
                    status_display="${CYAN}$status${NC}"
                    ;;
            esac
            
            printf "%-25s %-25s %-20s\n" "$phase_id" "$status_display" "$assigned"
        fi
    done
    
    echo ""
    
    # Show active locks
    if ls .migration/locks/*.lock 2>/dev/null | grep -q .; then
        echo -e "${YELLOW}Active Locks:${NC}"
        for lock_file in .migration/locks/*.lock; do
            resource=$(basename "$lock_file" .lock)
            lock_info=$(cat "$lock_file")
            echo "  - $resource: $lock_info"
        done
        echo ""
    fi
    
    # Show recent activity
    echo -e "${YELLOW}Recent Activity:${NC}"
    tail -5 .migration/migration.log | sed 's/^/  /'
}

# Run a specific phase
run_phase() {
    local phase_id=$1
    
    if [[ -z "$phase_id" ]]; then
        log_error "Phase ID required"
        show_usage
        return 1
    fi
    
    log_phase "Running Phase: $phase_id"
    
    # Check if migration is initialized
    if [[ ! -f .migration/state.yaml ]]; then
        log_error "Migration not initialized. Run '$0 init' first."
        return 1
    fi
    
    # Check if phase exists
    if [[ ! -f ".migration/phases/${phase_id}.yaml" ]]; then
        log_error "Phase not found: $phase_id"
        return 1
    fi
    
    # Check if phase is already complete
    if is_phase_complete "$phase_id"; then
        log_warning "Phase already completed: $phase_id"
        return 0
    fi
    
    # Check dependencies
    log_info "Checking phase dependencies..."
    dependencies=$(grep -A10 "id: \"$phase_id\"" migration-manifest.yaml | grep -A5 "dependencies:" | grep "^    -" | awk '{print $2}' | tr -d '"' || true)
    
    for dep in $dependencies; do
        if ! is_phase_complete "$dep"; then
            log_error "Dependency not satisfied: $dep must be completed first"
            return 1
        fi
    done
    
    # Try to assign phase to this agent
    if ! assign_phase_to_agent "$phase_id" "$AGENT_ID"; then
        log_error "Failed to assign phase to agent"
        return 1
    fi
    
    # Create rollback point
    create_rollback_point "$phase_id" "Before phase $phase_id"
    
    # Execute phase based on type
    local phase_script="$SCRIPT_DIR/phases/${phase_id}.sh"
    
    if [[ -f "$phase_script" ]]; then
        log_info "Executing phase script: $phase_script"
        
        if [[ "$DRY_RUN" == "true" ]]; then
            log_warning "DRY RUN: Would execute $phase_script"
        else
            if "$phase_script"; then
                update_phase_status "$phase_id" "$PHASE_COMPLETED" "$AGENT_ID"
                log_success "Phase completed: $phase_id"
                
                # Run validations
                validate_phase "$phase_id"
            else
                update_phase_status "$phase_id" "$PHASE_FAILED" "$AGENT_ID"
                log_error "Phase failed: $phase_id"
                return 1
            fi
        fi
    else
        log_error "Phase script not found: $phase_script"
        update_phase_status "$phase_id" "$PHASE_FAILED" "$AGENT_ID"
        return 1
    fi
    
    # Generate phase report
    generate_phase_report "$phase_id"
}

# Validate a phase
validate_phase() {
    local phase_id=$1
    
    log_info "Validating phase: $phase_id"
    
    # Get validations from manifest
    local validations=$(mktemp)
    awk "/id: \"$phase_id\"/,/^  - id:/" migration-manifest.yaml | grep -A20 "validations:" | grep -E "name:|command:" > "$validations"
    
    local all_passed=true
    
    while IFS= read -r line; do
        if [[ "$line" =~ name:\ \"(.*)\" ]]; then
            validation_name="${BASH_REMATCH[1]}"
        elif [[ "$line" =~ command:\ \"(.*)\" ]]; then
            validation_cmd="${BASH_REMATCH[1]}"
            
            log_info "Running validation: $validation_name"
            
            if eval "$validation_cmd" > /dev/null 2>&1; then
                record_validation "$phase_id" "$validation_name" "passed"
                log_success "Validation passed: $validation_name"
            else
                record_validation "$phase_id" "$validation_name" "failed"
                log_error "Validation failed: $validation_name"
                all_passed=false
            fi
        fi
    done < "$validations"
    
    rm -f "$validations"
    
    if [[ "$all_passed" == "true" ]]; then
        log_success "All validations passed for phase: $phase_id"
        return 0
    else
        log_error "Some validations failed for phase: $phase_id"
        return 1
    fi
}

# Run all phases
run_all() {
    log_phase "Running Complete Migration"
    
    # Get all phases in order
    local phases=($(grep "^  - id:" migration-manifest.yaml | awk '{print $3}' | tr -d '"'))
    
    log_info "Found ${#phases[@]} phases to execute"
    
    for phase in "${phases[@]}"; do
        if is_phase_complete "$phase"; then
            log_info "Skipping completed phase: $phase"
            continue
        fi
        
        log_info "Executing phase: $phase"
        
        if ! run_phase "$phase"; then
            log_error "Migration stopped due to phase failure: $phase"
            return 1
        fi
        
        # Small delay between phases
        sleep 2
    done
    
    log_success "All phases completed successfully!"
    
    # Update overall status
    update_migration_state "status" "completed"
    update_migration_state "completed_at" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}

# Rollback to a phase
rollback_phase() {
    local phase_id=$1
    
    log_phase "Rollback to Phase: $phase_id"
    
    if [[ "$FORCE" != "true" ]]; then
        read -p "Are you sure you want to rollback? This will lose all progress after $phase_id. (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Rollback cancelled"
            return 0
        fi
    fi
    
    # Find rollback script
    local rollback_script="$SCRIPT_DIR/rollback/rollback-to-${phase_id}.sh"
    
    if [[ -f "$rollback_script" ]]; then
        log_info "Executing rollback script: $rollback_script"
        "$rollback_script"
    else
        log_warning "No specific rollback script found, using git rollback"
        
        # Find git tag for phase
        local tag=$(grep -A10 "phase: \"$phase_id\"" .migration/rollback-points.yaml | grep "tag:" | head -1 | awk '{print $2}' | tr -d '"')
        
        if [[ -n "$tag" ]]; then
            git reset --hard "$tag"
            log_success "Rolled back to tag: $tag"
        else
            log_error "No rollback point found for phase: $phase_id"
            return 1
        fi
    fi
    
    # Update migration state
    update_migration_state "status" "rolled_back"
    update_migration_state "rolled_back_to" "$phase_id"
}

# Generate final report
generate_report() {
    log_phase "Generating Migration Report"
    
    local report_file=".migration/reports/final-report-$(date +%Y%m%d-%H%M%S).md"
    mkdir -p "$(dirname "$report_file")"
    
    "$SCRIPT_DIR/lib/report-generator.sh" "$report_file"
    
    log_success "Report generated: $report_file"
}

# Cleanup migration state
cleanup_migration() {
    log_phase "Cleaning Up Migration State"
    
    if [[ "$FORCE" != "true" ]]; then
        read -p "This will remove all migration tracking. Are you sure? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Cleanup cancelled"
            return 0
        fi
    fi
    
    # Archive migration data
    local archive_name="migration-archive-$(date +%Y%m%d-%H%M%S).tar.gz"
    tar -czf "$archive_name" .migration/
    log_info "Migration data archived to: $archive_name"
    
    # Remove migration directory
    rm -rf .migration/
    log_success "Migration state cleaned up"
}

# Main execution
cd "$ROOT_DIR"

case "$COMMAND" in
    init)
        init_migration
        ;;
    status)
        show_status
        ;;
    run-phase)
        run_phase "$PHASE_ID"
        ;;
    run-all)
        run_all
        ;;
    validate)
        validate_phase "$PHASE_ID"
        ;;
    rollback)
        rollback_phase "$PHASE_ID"
        ;;
    monitor)
        monitor_migration_progress
        ;;
    report)
        generate_report
        ;;
    cleanup)
        cleanup_migration
        ;;
    help|*)
        show_usage
        ;;
esac