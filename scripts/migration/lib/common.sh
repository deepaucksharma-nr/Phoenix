#!/bin/bash
# common.sh - Common functions for migration scripts

# Colors
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export BLUE='\033[0;34m'
export MAGENTA='\033[0;35m'
export CYAN='\033[0;36m'
export NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a .migration/migration.log
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a .migration/migration.log
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a .migration/migration.log
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a .migration/migration.log
}

log_phase() {
    echo "" | tee -a .migration/migration.log
    echo -e "${MAGENTA}════════════════════════════════════════════════════════════════${NC}" | tee -a .migration/migration.log
    echo -e "${MAGENTA}  $1${NC}" | tee -a .migration/migration.log
    echo -e "${MAGENTA}════════════════════════════════════════════════════════════════${NC}" | tee -a .migration/migration.log
    echo "" | tee -a .migration/migration.log
}

# State management functions
init_migration_state() {
    mkdir -p .migration/{locks,state,rollback,temp}
    
    if [[ ! -f .migration/state.yaml ]]; then
        cat > .migration/state.yaml << EOF
migration:
  id: "$(uuidgen || cat /proc/sys/kernel/random/uuid)"
  version: "2.0"
  started_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  started_by: "${AGENT_ID:-unknown}"
  status: "initializing"

phases: []
locks: []
rollback_points: []
EOF
    fi
    
    # Initialize log
    echo "=== Migration Started at $(date -u +%Y-%m-%dT%H:%M:%SZ) ===" > .migration/migration.log
}

update_migration_state() {
    local key=$1
    local value=$2
    local state_file=".migration/state.yaml"
    
    # Use yq or fall back to sed
    if command -v yq &> /dev/null; then
        yq eval ".migration.$key = \"$value\"" -i "$state_file"
    else
        # Simple sed replacement for basic cases
        sed -i "s|$key:.*|$key: \"$value\"|" "$state_file"
    fi
}

# Lock management
acquire_lock() {
    local resource=$1
    local agent_id=${2:-${AGENT_ID:-unknown}}
    local lock_file=".migration/locks/${resource}.lock"
    local max_wait=${3:-300} # 5 minutes default
    local waited=0
    
    mkdir -p "$(dirname "$lock_file")"
    
    while [[ $waited -lt $max_wait ]]; do
        if (set -C; echo "$agent_id:$(date -u +%s)" > "$lock_file") 2>/dev/null; then
            log_info "Lock acquired for $resource by $agent_id"
            return 0
        fi
        
        # Check if lock is stale (older than 1 hour)
        if [[ -f "$lock_file" ]]; then
            local lock_time=$(cut -d: -f2 "$lock_file")
            local current_time=$(date -u +%s)
            if (( current_time - lock_time > 3600 )); then
                log_warning "Removing stale lock for $resource"
                rm -f "$lock_file"
                continue
            fi
        fi
        
        sleep 5
        ((waited += 5))
        
        if (( waited % 30 == 0 )); then
            log_info "Waiting for lock on $resource... (${waited}s elapsed)"
        fi
    done
    
    log_error "Failed to acquire lock for $resource after ${max_wait}s"
    return 1
}

release_lock() {
    local resource=$1
    local agent_id=${2:-${AGENT_ID:-unknown}}
    local lock_file=".migration/locks/${resource}.lock"
    
    if [[ -f "$lock_file" ]] && grep -q "^$agent_id:" "$lock_file"; then
        rm -f "$lock_file"
        log_info "Lock released for $resource by $agent_id"
    else
        log_warning "Cannot release lock for $resource - not owned by $agent_id"
    fi
}

release_all_locks() {
    local agent_id=${1:-${AGENT_ID:-unknown}}
    
    for lock_file in .migration/locks/*.lock; do
        if [[ -f "$lock_file" ]] && grep -q "^$agent_id:" "$lock_file"; then
            local resource=$(basename "$lock_file" .lock)
            release_lock "$resource" "$agent_id"
        fi
    done
}

# Idempotent operations
create_directory() {
    local dir=$1
    
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
        log_info "Created directory: $dir"
    else
        log_info "Directory already exists: $dir"
    fi
}

copy_file() {
    local src=$1
    local dst=$2
    local transform=${3:-}
    
    if [[ ! -f "$src" ]]; then
        log_error "Source file not found: $src"
        return 1
    fi
    
    # Check if already migrated
    if [[ -f "$dst.migrated" ]]; then
        log_info "Already migrated: $dst"
        return 0
    fi
    
    # Create destination directory
    create_directory "$(dirname "$dst")"
    
    # Copy file
    cp "$src" "$dst"
    
    # Apply transformation if specified
    if [[ -n "$transform" ]] && [[ -x "$transform" ]]; then
        log_info "Applying transformation: $transform"
        "$transform" "$dst"
    fi
    
    # Mark as migrated
    touch "$dst.migrated"
    log_success "Migrated: $src → $dst"
}

copy_directory() {
    local src=$1
    local dst=$2
    local exclude_pattern=${3:-}
    
    if [[ ! -d "$src" ]]; then
        log_error "Source directory not found: $src"
        return 1
    fi
    
    create_directory "$dst"
    
    if [[ -n "$exclude_pattern" ]]; then
        rsync -av --exclude="$exclude_pattern" "$src/" "$dst/"
    else
        rsync -av "$src/" "$dst/"
    fi
    
    log_success "Migrated directory: $src → $dst"
}

# Validation helpers
run_validation() {
    local name=$1
    local command=$2
    local working_dir=${3:-.}
    
    log_info "Running validation: $name"
    
    if (cd "$working_dir" && eval "$command" > .migration/temp/validation.out 2>&1); then
        log_success "Validation passed: $name"
        return 0
    else
        log_error "Validation failed: $name"
        cat .migration/temp/validation.out
        return 1
    fi
}

# Rollback functions
create_rollback_point() {
    local phase=$1
    local description=$2
    
    # Ensure clean git state
    if [[ -n $(git status --porcelain) ]]; then
        git add -A
        git commit -m "Auto-commit before rollback point: $description" || true
    fi
    
    # Create tag
    local tag="rollback-${phase}-$(date +%s)"
    git tag "$tag"
    
    # Record in state
    cat >> .migration/rollback-points.yaml << EOF
- phase: "$phase"
  tag: "$tag"
  description: "$description"
  created_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
EOF
    
    log_success "Created rollback point: $tag"
}

# Process management
ensure_no_running_services() {
    local services=("$@")
    
    for service in "${services[@]}"; do
        if pgrep -f "$service" > /dev/null; then
            log_error "Service $service is running. Stop it before migration."
            return 1
        fi
    done
    
    # Check Docker containers
    if docker ps --format "{{.Names}}" | grep -E "(phoenix|control|telemetry)" > /dev/null; then
        log_error "Phoenix Docker containers are running. Stop with: docker-compose down"
        return 1
    fi
    
    return 0
}

# Cleanup on exit
cleanup_on_exit() {
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        log_error "Migration script exited with error code: $exit_code"
        release_all_locks
    fi
    
    # Always log completion
    echo "=== Migration Ended at $(date -u +%Y-%m-%dT%H:%M:%SZ) ===" >> .migration/migration.log
}

# Set trap for cleanup
trap cleanup_on_exit EXIT

# Export functions
export -f log_info log_success log_warning log_error log_phase
export -f init_migration_state update_migration_state
export -f acquire_lock release_lock release_all_locks
export -f create_directory copy_file copy_directory
export -f run_validation create_rollback_point
export -f ensure_no_running_services cleanup_on_exit