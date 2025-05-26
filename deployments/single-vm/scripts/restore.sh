#!/usr/bin/env bash
#
# Phoenix Restore Script
# Restores Phoenix from backup files
#
# Usage: ./restore.sh [timestamp]
#        ./restore.sh latest
#

set -euo pipefail

# Configuration
PHOENIX_DIR="${PHOENIX_DIR:-/opt/phoenix}"
BACKUP_DIR="${BACKUP_DIR:-$PHOENIX_DIR/backups}"
TIMESTAMP="${1:-latest}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper functions
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

# Show usage
usage() {
    echo "Usage: $0 [timestamp|latest]"
    echo
    echo "Examples:"
    echo "  $0 20240115_143022    # Restore specific backup"
    echo "  $0 latest             # Restore most recent backup"
    echo "  $0                    # Interactive mode"
    exit 1
}

# Interactive backup selection
select_backup() {
    echo "Available backups:"
    echo "────────────────────────────────────────────────"
    
    local backups=()
    local i=1
    
    # Find all backup manifests
    for manifest in $(ls -t "$BACKUP_DIR"/manifest_*.json 2>/dev/null); do
        local ts=$(basename "$manifest" | sed 's/manifest_\(.*\)\.json/\1/')
        local date=$(jq -r '.date' "$manifest" 2>/dev/null || echo "Unknown")
        local size=$(jq -r '.sizes.total' "$manifest" 2>/dev/null || echo "Unknown")
        
        echo "$i) $ts - $date (Total: $size)"
        backups+=("$ts")
        ((i++))
    done
    
    if [[ ${#backups[@]} -eq 0 ]]; then
        error "No backups found in $BACKUP_DIR"
        exit 1
    fi
    
    echo
    read -p "Select backup to restore (1-${#backups[@]}): " selection
    
    if [[ $selection -ge 1 && $selection -le ${#backups[@]} ]]; then
        TIMESTAMP="${backups[$((selection-1))]}"
        echo
        log "Selected backup: $TIMESTAMP"
    else
        error "Invalid selection"
        exit 1
    fi
}

# Main restore process
main() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}                   Phoenix Restore Script                       ${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo
    
    # Change to Phoenix directory
    cd "$PHOENIX_DIR" || {
        error "Phoenix directory not found: $PHOENIX_DIR"
        exit 1
    }
    
    # Handle timestamp parameter
    if [[ -z "$TIMESTAMP" ]]; then
        select_backup
    elif [[ "$TIMESTAMP" == "latest" ]]; then
        # Find latest backup
        if [[ -L "$BACKUP_DIR/manifest_latest.json" ]]; then
            TIMESTAMP=$(readlink "$BACKUP_DIR/manifest_latest.json" | sed 's/manifest_\(.*\)\.json/\1/')
            log "Using latest backup: $TIMESTAMP"
        else
            error "No latest backup symlink found"
            select_backup
        fi
    fi
    
    # Verify backup files exist
    local manifest="$BACKUP_DIR/manifest_${TIMESTAMP}.json"
    if [[ ! -f "$manifest" ]]; then
        error "Backup manifest not found: $manifest"
        exit 1
    fi
    
    # Show backup information
    echo "Backup Information:"
    echo "──────────────────"
    jq . "$manifest" 2>/dev/null || cat "$manifest"
    echo
    
    # Confirmation
    echo -e "${YELLOW}⚠️  WARNING: This will restore Phoenix to the selected backup state.${NC}"
    echo -e "${YELLOW}   Current data will be overwritten!${NC}"
    echo
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [[ "$confirm" != "yes" ]]; then
        echo "Restore cancelled."
        exit 0
    fi
    
    # Stop services
    log "Stopping Phoenix services..."
    docker-compose down || {
        warning "Failed to stop services gracefully"
    }
    
    # Wait for services to stop
    sleep 5
    
    # 1. Restore PostgreSQL database
    local db_backup="$BACKUP_DIR/phoenix_db_${TIMESTAMP}.dump"
    if [[ -f "$db_backup" ]]; then
        log "Restoring PostgreSQL database..."
        
        # Start only database service
        docker-compose up -d db
        
        # Wait for database to be ready
        info "Waiting for database to be ready..."
        local retries=30
        while ! docker-compose exec -T db pg_isready -U phoenix >/dev/null 2>&1; do
            retries=$((retries - 1))
            if [[ $retries -le 0 ]]; then
                error "Database failed to start"
                exit 1
            fi
            sleep 2
        done
        
        # Drop existing database and recreate
        docker-compose exec -T db psql -U phoenix -c "DROP DATABASE IF EXISTS phoenix;" postgres
        docker-compose exec -T db psql -U phoenix -c "CREATE DATABASE phoenix;" postgres
        
        # Restore database
        docker-compose exec -T db pg_restore -U phoenix -d phoenix < "$db_backup" 2>/dev/null || {
            # Try alternative restore method
            warning "pg_restore failed, trying psql restore..."
            docker-compose exec -T db psql -U phoenix phoenix < "$db_backup"
        }
        
        log "Database restored successfully"
    else
        error "Database backup not found: $db_backup"
    fi
    
    # 2. Restore Prometheus data
    local prom_backup="$BACKUP_DIR/prometheus_data_${TIMESTAMP}.tar.gz"
    if [[ -f "$prom_backup" ]]; then
        log "Restoring Prometheus data..."
        
        # Stop Prometheus if running
        docker-compose stop prometheus 2>/dev/null || true
        
        # Backup current data
        if [[ -d "$PHOENIX_DIR/data/prometheus" ]]; then
            mv "$PHOENIX_DIR/data/prometheus" "$PHOENIX_DIR/data/prometheus.old.$(date +%s)"
        fi
        
        # Extract backup
        mkdir -p "$PHOENIX_DIR/data"
        tar -xzf "$prom_backup" -C "$PHOENIX_DIR/data/" || {
            warning "Failed to restore Prometheus data"
        }
        
        log "Prometheus data restored"
    else
        warning "Prometheus backup not found: $prom_backup"
    fi
    
    # 3. Restore Grafana data
    local grafana_backup="$BACKUP_DIR/grafana_data_${TIMESTAMP}.tar.gz"
    if [[ -f "$grafana_backup" ]]; then
        log "Restoring Grafana data..."
        
        # Stop Grafana if running
        docker-compose stop grafana 2>/dev/null || true
        
        # Backup current data
        if [[ -d "$PHOENIX_DIR/data/grafana" ]]; then
            mv "$PHOENIX_DIR/data/grafana" "$PHOENIX_DIR/data/grafana.old.$(date +%s)"
        fi
        
        # Extract backup
        tar -xzf "$grafana_backup" -C "$PHOENIX_DIR/data/" || {
            warning "Failed to restore Grafana data"
        }
        
        log "Grafana data restored"
    else
        warning "Grafana backup not found: $grafana_backup"
    fi
    
    # 4. Restore configuration
    local config_backup="$BACKUP_DIR/config_${TIMESTAMP}.tar.gz"
    if [[ -f "$config_backup" ]]; then
        log "Restoring configuration..."
        
        # Backup current config
        cp "$PHOENIX_DIR/.env" "$PHOENIX_DIR/.env.before_restore.$(date +%s)" 2>/dev/null || true
        
        # Extract config (but preserve current .env by default)
        tar -xzf "$config_backup" -C "$PHOENIX_DIR/" --exclude='.env' || {
            warning "Failed to restore configuration"
        }
        
        read -p "Restore .env file from backup? (y/n): " restore_env
        if [[ "$restore_env" == "y" ]]; then
            tar -xzf "$config_backup" -C "$PHOENIX_DIR/" .env
            log ".env file restored from backup"
        else
            log "Keeping current .env file"
        fi
    else
        warning "Configuration backup not found: $config_backup"
    fi
    
    # 5. Restore TLS certificates (if needed)
    local tls_backup="$BACKUP_DIR/tls_${TIMESTAMP}.tar.gz.enc"
    if [[ -f "$tls_backup" ]]; then
        read -p "Restore TLS certificates? (y/n): " restore_tls
        if [[ "$restore_tls" == "y" ]]; then
            log "Restoring TLS certificates..."
            read -s -p "Enter TLS backup password: " tls_password
            echo
            
            openssl enc -d -aes-256-cbc -k "$tls_password" -in "$tls_backup" | \
                tar -xzf - -C "$PHOENIX_DIR/" || {
                    warning "Failed to restore TLS certificates"
                }
        fi
    fi
    
    # 6. Set proper permissions
    log "Setting permissions..."
    sudo chown -R $(id -u):$(id -g) "$PHOENIX_DIR/data" 2>/dev/null || true
    
    # 7. Start all services
    log "Starting Phoenix services..."
    docker-compose up -d
    
    # Wait for services to be healthy
    info "Waiting for services to be ready..."
    sleep 10
    
    # 8. Verify restoration
    log "Verifying restoration..."
    "$PHOENIX_DIR/scripts/health-check.sh" || {
        warning "Health check failed - services may need more time to start"
    }
    
    # 9. Summary
    echo
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}              Phoenix Restore Completed Successfully            ${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo
    echo "Restored from backup: $TIMESTAMP"
    echo
    echo "Next steps:"
    echo "  1. Verify services: $PHOENIX_DIR/scripts/health-check.sh"
    echo "  2. Check logs: cd $PHOENIX_DIR && docker-compose logs"
    echo "  3. Access UI: $(grep PHX_PUBLIC_URL "$PHOENIX_DIR/.env" | cut -d= -f2)"
    echo
    echo "Old data backed up with .old.* suffix in data directory"
    echo
}

# Run main function
main "$@"