#!/usr/bin/env bash
#
# Phoenix Backup Script
# Creates backups of database, Prometheus data, and configuration
# Supports incremental and full backups
#

set -euo pipefail

# Parse arguments
BACKUP_TYPE="full"
for arg in "$@"; do
    case $arg in
        --incremental)
            BACKUP_TYPE="incremental"
            ;;
        --full)
            BACKUP_TYPE="full"
            ;;
        --rotate-weekly)
            BACKUP_TYPE="rotate"
            ;;
    esac
done

# Configuration
PHOENIX_DIR="${PHOENIX_DIR:-/opt/phoenix}"
BACKUP_DIR="${BACKUP_DIR:-$PHOENIX_DIR/backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS="${RETENTION_DAYS:-7}"
S3_BUCKET="${S3_BUCKET:-}"  # Optional S3 backup

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
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

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Change to Phoenix directory
cd "$PHOENIX_DIR" || {
    error "Phoenix directory not found: $PHOENIX_DIR"
    exit 1
}

log "Starting Phoenix backup (type: $BACKUP_TYPE)..."

# Handle rotation
if [[ "$BACKUP_TYPE" == "rotate" ]]; then
    log "Rotating weekly backups..."
    mkdir -p "$BACKUP_DIR/weekly"
    
    # Move latest full backup to weekly
    latest_full=$(ls -t "$BACKUP_DIR"/phoenix_db_*.dump 2>/dev/null | grep -v incremental | head -1)
    if [[ -n "$latest_full" ]]; then
        cp "$latest_full" "$BACKUP_DIR/weekly/$(basename "$latest_full")"
        log "Archived weekly backup: $(basename "$latest_full")"
    fi
    
    # Clean old weekly backups (keep 4 weeks)
    find "$BACKUP_DIR/weekly" -name "*.dump" -mtime +28 -delete 2>/dev/null || true
    exit 0
fi

# For incremental backups, check if we need a full backup
if [[ "$BACKUP_TYPE" == "incremental" ]]; then
    # Check if last full backup is older than 24 hours
    last_full=$(ls -t "$BACKUP_DIR"/phoenix_db_*.dump 2>/dev/null | grep -v incremental | head -1)
    if [[ -z "$last_full" ]] || [[ $(find "$last_full" -mmin +1440 2>/dev/null | wc -l) -gt 0 ]]; then
        log "Last full backup is older than 24 hours, performing full backup instead"
        BACKUP_TYPE="full"
    fi
fi

# 1. Backup PostgreSQL database
log "Backing up PostgreSQL database..."
if docker-compose exec -T db pg_dump -U phoenix -Fc phoenix > "$BACKUP_DIR/phoenix_db_${TIMESTAMP}.dump" 2>/dev/null; then
    log "Database backup completed: phoenix_db_${TIMESTAMP}.dump"
    
    # Get database size
    db_size=$(docker-compose exec -T db psql -U phoenix -t -c "SELECT pg_database_size('phoenix');" | tr -d ' ')
    db_size_mb=$((db_size / 1024 / 1024))
    log "Database size: ${db_size_mb}MB"
else
    error "Failed to backup database"
    exit 1
fi

# 2. Backup Prometheus data
if [[ "$BACKUP_TYPE" == "incremental" ]]; then
    log "Skipping Prometheus backup for incremental backup"
else
    log "Backing up Prometheus data..."
    if [[ -d "$PHOENIX_DIR/data/prometheus" ]]; then
    # Create snapshot via Prometheus API
    snapshot_result=$(curl -X POST http://localhost:9090/api/v1/admin/tsdb/snapshot 2>/dev/null || echo "{}")
    snapshot_name=$(echo "$snapshot_result" | jq -r '.data.name' 2>/dev/null || echo "")
    
    if [[ -n "$snapshot_name" && "$snapshot_name" != "null" ]]; then
        # Compress snapshot
        tar -czf "$BACKUP_DIR/prometheus_data_${TIMESTAMP}.tar.gz" \
            -C "$PHOENIX_DIR/data/prometheus/snapshots" \
            "$snapshot_name" 2>/dev/null
        
        # Remove snapshot after backup
        rm -rf "$PHOENIX_DIR/data/prometheus/snapshots/$snapshot_name"
        
        log "Prometheus backup completed: prometheus_data_${TIMESTAMP}.tar.gz"
    else
        # Fallback: direct backup
        warning "Prometheus snapshot API failed, using direct backup"
        tar -czf "$BACKUP_DIR/prometheus_data_${TIMESTAMP}.tar.gz" \
            -C "$PHOENIX_DIR/data" \
            --exclude='prometheus/wal' \
            --exclude='prometheus/chunks_head' \
            prometheus/ 2>/dev/null || {
                warning "Prometheus data backup failed"
            }
    fi
else
    warning "Prometheus data directory not found"
fi
fi

# 3. Backup Grafana dashboards and data
if [[ "$BACKUP_TYPE" == "incremental" ]]; then
    log "Skipping Grafana backup for incremental backup"
else
    log "Backing up Grafana data..."
    if [[ -d "$PHOENIX_DIR/data/grafana" ]]; then
    tar -czf "$BACKUP_DIR/grafana_data_${TIMESTAMP}.tar.gz" \
        -C "$PHOENIX_DIR/data" \
        grafana/ 2>/dev/null && \
        log "Grafana backup completed: grafana_data_${TIMESTAMP}.tar.gz" || \
        warning "Grafana backup failed"
else
    warning "Grafana data directory not found"
fi
fi

# 4. Backup configuration files
log "Backing up configuration..."
tar -czf "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" \
    -C "$PHOENIX_DIR" \
    --exclude='tls/*.pem' \
    config/ \
    .env \
    docker-compose.yml \
    2>/dev/null && \
    log "Configuration backup completed: config_${TIMESTAMP}.tar.gz" || \
    error "Configuration backup failed"

# 5. Backup TLS certificates separately (encrypted)
if [[ -d "$PHOENIX_DIR/tls" ]] && [[ -f "$PHOENIX_DIR/tls/fullchain.pem" ]]; then
    log "Backing up TLS certificates (encrypted)..."
    tar -czf - -C "$PHOENIX_DIR" tls/ | \
        openssl enc -aes-256-cbc -salt -k "${TLS_BACKUP_PASSWORD:-phoenix}" \
        -out "$BACKUP_DIR/tls_${TIMESTAMP}.tar.gz.enc" 2>/dev/null && \
        log "TLS backup completed: tls_${TIMESTAMP}.tar.gz.enc" || \
        warning "TLS backup failed"
fi

# 6. Create backup manifest
log "Creating backup manifest..."
if [[ "$BACKUP_TYPE" == "incremental" ]]; then
    backup_suffix="_incremental"
else
    backup_suffix=""
fi
cat > "$BACKUP_DIR/manifest_${TIMESTAMP}${backup_suffix}.json" << EOF
{
  "timestamp": "$TIMESTAMP",
  "date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "version": "$(docker-compose exec -T api phoenix-api --version 2>/dev/null | tr -d '\r\n' || echo 'unknown')",
  "backup_type": "$BACKUP_TYPE",
  "files": {
    "database": "phoenix_db_${TIMESTAMP}.dump",
    "prometheus": $(if [[ "$BACKUP_TYPE" == "incremental" ]]; then echo '"skipped"'; else echo "\"prometheus_data_${TIMESTAMP}.tar.gz\""; fi),
    "grafana": $(if [[ "$BACKUP_TYPE" == "incremental" ]]; then echo '"skipped"'; else echo "\"grafana_data_${TIMESTAMP}.tar.gz\""; fi),
    "config": "config_${TIMESTAMP}.tar.gz",
    "tls": "tls_${TIMESTAMP}.tar.gz.enc"
  },
  "sizes": {
    "database": "$(du -h "$BACKUP_DIR/phoenix_db_${TIMESTAMP}.dump" 2>/dev/null | cut -f1 || echo 'N/A')",
    "prometheus": "$(du -h "$BACKUP_DIR/prometheus_data_${TIMESTAMP}.tar.gz" 2>/dev/null | cut -f1 || echo 'N/A')",
    "total": "$(du -sh "$BACKUP_DIR" | cut -f1)"
  }
}
EOF

# 7. Clean old backups
log "Cleaning old backups (older than $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "phoenix_db_*.dump" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "prometheus_data_*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "grafana_data_*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "config_*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "tls_*.tar.gz.enc" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "manifest_*.json" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true

# 8. Upload to S3 (optional)
if [[ -n "$S3_BUCKET" ]]; then
    log "Uploading backup to S3..."
    if command -v aws >/dev/null 2>&1; then
        # Create daily folder
        s3_prefix="phoenix-backups/$(date +%Y/%m/%d)"
        
        # Upload files
        for file in \
            "phoenix_db_${TIMESTAMP}.dump" \
            "prometheus_data_${TIMESTAMP}.tar.gz" \
            "grafana_data_${TIMESTAMP}.tar.gz" \
            "config_${TIMESTAMP}.tar.gz" \
            "tls_${TIMESTAMP}.tar.gz.enc" \
            "manifest_${TIMESTAMP}.json"
        do
            if [[ -f "$BACKUP_DIR/$file" ]]; then
                aws s3 cp "$BACKUP_DIR/$file" "s3://$S3_BUCKET/$s3_prefix/$file" && \
                    log "Uploaded $file to S3" || \
                    warning "Failed to upload $file to S3"
            fi
        done
    else
        warning "AWS CLI not found, skipping S3 upload"
    fi
fi

# 9. Create latest symlinks
ln -sf "phoenix_db_${TIMESTAMP}.dump" "$BACKUP_DIR/phoenix_db_latest.dump"
ln -sf "config_${TIMESTAMP}.tar.gz" "$BACKUP_DIR/config_latest.tar.gz"
ln -sf "manifest_${TIMESTAMP}.json" "$BACKUP_DIR/manifest_latest.json"

# 10. Summary
echo
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}              Phoenix Backup Completed Successfully             ${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo
echo "Backup location: $BACKUP_DIR"
echo "Backup timestamp: $TIMESTAMP"
echo
echo "Files created:"
ls -lh "$BACKUP_DIR"/*_${TIMESTAMP}* 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
echo
echo "Total backup size: $(du -sh "$BACKUP_DIR" | cut -f1)"
echo "Free disk space: $(df -h "$BACKUP_DIR" | tail -1 | awk '{print $4}')"
echo
echo "To restore from this backup:"
echo "  $PHOENIX_DIR/scripts/restore.sh $TIMESTAMP"
echo

# Exit successfully
exit 0