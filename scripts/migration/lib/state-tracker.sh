#!/bin/bash
# state-tracker.sh - Migration state tracking system

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Phase status constants
readonly PHASE_PENDING="pending"
readonly PHASE_IN_PROGRESS="in_progress"
readonly PHASE_COMPLETED="completed"
readonly PHASE_FAILED="failed"
readonly PHASE_ROLLED_BACK="rolled_back"

# Initialize phase tracking
init_phase() {
    local phase_id=$1
    local phase_name=$2
    local can_parallelize=${3:-false}
    
    local phase_file=".migration/phases/${phase_id}.yaml"
    mkdir -p "$(dirname "$phase_file")"
    
    if [[ ! -f "$phase_file" ]]; then
        cat > "$phase_file" << EOF
phase:
  id: "$phase_id"
  name: "$phase_name"
  status: "$PHASE_PENDING"
  can_parallelize: $can_parallelize
  created_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  components: []
  validations: []
EOF
    fi
}

# Update phase status
update_phase_status() {
    local phase_id=$1
    local status=$2
    local agent_id=${3:-${AGENT_ID:-unknown}}
    
    local phase_file=".migration/phases/${phase_id}.yaml"
    
    # Record status change
    cat >> "$phase_file" << EOF

status_history:
  - status: "$status"
    changed_by: "$agent_id"
    changed_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
EOF
    
    # Update current status
    sed -i "s/status: .*/status: \"$status\"/" "$phase_file"
    
    log_info "Phase $phase_id status updated to: $status"
}

# Track component within a phase
track_component() {
    local phase_id=$1
    local component=$2
    local status=$3
    
    local phase_file=".migration/phases/${phase_id}.yaml"
    local component_file=".migration/components/${phase_id}/${component}.yaml"
    
    mkdir -p "$(dirname "$component_file")"
    
    cat > "$component_file" << EOF
component:
  name: "$component"
  phase: "$phase_id"
  status: "$status"
  updated_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
EOF
    
    # Update phase file
    echo "  - component: $component" >> "$phase_file"
    echo "    status: $status" >> "$phase_file"
}

# Check if phase is complete
is_phase_complete() {
    local phase_id=$1
    local phase_file=".migration/phases/${phase_id}.yaml"
    
    if [[ ! -f "$phase_file" ]]; then
        return 1
    fi
    
    grep -q "status: \"$PHASE_COMPLETED\"" "$phase_file"
}

# Check if phase is assigned
is_phase_assigned() {
    local phase_id=$1
    local phase_file=".migration/phases/${phase_id}.yaml"
    
    if [[ ! -f "$phase_file" ]]; then
        return 1
    fi
    
    grep -q "assigned_to:" "$phase_file"
}

# Get next available phase
get_next_phase() {
    local manifest_file="migration-manifest.yaml"
    
    if [[ ! -f "$manifest_file" ]]; then
        log_error "Migration manifest not found"
        return 1
    fi
    
    # Read phases from manifest
    local phases=($(grep "^  - id:" "$manifest_file" | awk '{print $3}' | tr -d '"'))
    
    for phase in "${phases[@]}"; do
        if ! is_phase_complete "$phase" && ! is_phase_assigned "$phase"; then
            echo "$phase"
            return 0
        fi
    done
    
    return 1
}

# Assign phase to agent
assign_phase_to_agent() {
    local phase_id=$1
    local agent_id=$2
    
    local phase_file=".migration/phases/${phase_id}.yaml"
    
    # Check if already assigned
    if is_phase_assigned "$phase_id"; then
        log_error "Phase $phase_id is already assigned"
        return 1
    fi
    
    # Atomically assign
    if acquire_lock "phase-assignment" "$agent_id" 10; then
        cat >> "$phase_file" << EOF

assignment:
  assigned_to: "$agent_id"
  assigned_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
EOF
        release_lock "phase-assignment" "$agent_id"
        
        log_success "Phase $phase_id assigned to agent $agent_id"
        update_phase_status "$phase_id" "$PHASE_IN_PROGRESS" "$agent_id"
        return 0
    else
        log_error "Failed to acquire lock for phase assignment"
        return 1
    fi
}

# Record validation result
record_validation() {
    local phase_id=$1
    local validation_name=$2
    local result=$3
    local details=${4:-}
    
    local validation_file=".migration/validations/${phase_id}/${validation_name}.yaml"
    mkdir -p "$(dirname "$validation_file")"
    
    cat > "$validation_file" << EOF
validation:
  name: "$validation_name"
  phase: "$phase_id"
  result: "$result"
  details: "$details"
  executed_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  executed_by: "${AGENT_ID:-unknown}"
EOF
    
    # Update phase file
    local phase_file=".migration/phases/${phase_id}.yaml"
    echo "  - validation: $validation_name" >> "$phase_file"
    echo "    result: $result" >> "$phase_file"
}

# Check if all validations passed for a phase
all_validations_passed() {
    local phase_id=$1
    local validation_dir=".migration/validations/${phase_id}"
    
    if [[ ! -d "$validation_dir" ]]; then
        log_warning "No validations found for phase $phase_id"
        return 1
    fi
    
    # Check each validation file
    for validation_file in "$validation_dir"/*.yaml; do
        if [[ -f "$validation_file" ]]; then
            if ! grep -q 'result: "passed"' "$validation_file"; then
                local validation_name=$(basename "$validation_file" .yaml)
                log_error "Validation failed: $validation_name"
                return 1
            fi
        fi
    done
    
    return 0
}

# Generate phase report
generate_phase_report() {
    local phase_id=$1
    local report_file=".migration/reports/phase-${phase_id}.md"
    
    mkdir -p "$(dirname "$report_file")"
    
    cat > "$report_file" << EOF
# Phase Report: $phase_id

**Generated at:** $(date -u +%Y-%m-%dT%H:%M:%SZ)

## Summary

- **Phase ID:** $phase_id
- **Status:** $(grep "status:" ".migration/phases/${phase_id}.yaml" | head -1 | awk '{print $2}' | tr -d '"')
- **Assigned to:** $(grep "assigned_to:" ".migration/phases/${phase_id}.yaml" | awk '{print $2}' | tr -d '"' || echo "Not assigned")

## Components

EOF
    
    # Add component status
    if [[ -d ".migration/components/${phase_id}" ]]; then
        for comp_file in ".migration/components/${phase_id}"/*.yaml; do
            if [[ -f "$comp_file" ]]; then
                local comp_name=$(basename "$comp_file" .yaml)
                local comp_status=$(grep "status:" "$comp_file" | awk '{print $2}' | tr -d '"')
                echo "- **$comp_name:** $comp_status" >> "$report_file"
            fi
        done
    fi
    
    echo "" >> "$report_file"
    echo "## Validations" >> "$report_file"
    echo "" >> "$report_file"
    
    # Add validation results
    if [[ -d ".migration/validations/${phase_id}" ]]; then
        for val_file in ".migration/validations/${phase_id}"/*.yaml; do
            if [[ -f "$val_file" ]]; then
                local val_name=$(basename "$val_file" .yaml)
                local val_result=$(grep "result:" "$val_file" | awk '{print $2}' | tr -d '"')
                echo "- **$val_name:** $val_result" >> "$report_file"
            fi
        done
    fi
    
    log_info "Generated phase report: $report_file"
}

# Monitor migration progress
monitor_migration_progress() {
    local refresh_interval=${1:-5}
    
    while true; do
        clear
        echo -e "${BLUE}=== Phoenix Migration Progress Monitor ===${NC}"
        echo -e "Time: $(date)"
        echo ""
        
        # Show phase status
        echo -e "${YELLOW}Phase Status:${NC}"
        for phase_file in .migration/phases/*.yaml; do
            if [[ -f "$phase_file" ]]; then
                local phase_id=$(basename "$phase_file" .yaml)
                local status=$(grep "status:" "$phase_file" | head -1 | awk '{print $2}' | tr -d '"')
                local assigned_to=$(grep "assigned_to:" "$phase_file" | awk '{print $2}' | tr -d '"' || echo "unassigned")
                
                case "$status" in
                    "$PHASE_COMPLETED")
                        echo -e "${GREEN}✓ $phase_id${NC} - Assigned to: $assigned_to"
                        ;;
                    "$PHASE_IN_PROGRESS")
                        echo -e "${YELLOW}⚡ $phase_id${NC} - Assigned to: $assigned_to"
                        ;;
                    "$PHASE_FAILED")
                        echo -e "${RED}✗ $phase_id${NC} - Assigned to: $assigned_to"
                        ;;
                    *)
                        echo -e "${CYAN}○ $phase_id${NC} - Status: $status"
                        ;;
                esac
            fi
        done
        
        echo ""
        echo -e "${YELLOW}Active Locks:${NC}"
        for lock_file in .migration/locks/*.lock; do
            if [[ -f "$lock_file" ]]; then
                local resource=$(basename "$lock_file" .lock)
                local lock_info=$(cat "$lock_file")
                echo "  - $resource: $lock_info"
            fi
        done
        
        echo ""
        echo -e "Press Ctrl+C to exit"
        sleep "$refresh_interval"
    done
}

# Export functions
export -f init_phase update_phase_status track_component
export -f is_phase_complete is_phase_assigned get_next_phase
export -f assign_phase_to_agent record_validation all_validations_passed
export -f generate_phase_report monitor_migration_progress