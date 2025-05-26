#!/usr/bin/env bash
#
# Phoenix Incremental Backup Script
# Performs lightweight hourly backups of critical data
#

set -euo pipefail

# Configuration
PHOENIX_DIR="${PHOENIX_DIR:-/opt/phoenix}"
BACKUP_DIR="${BACKUP_DIR:-$PHOENIX_DIR/backups/incremental}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_HOURS="${RETENTION_HOURS:-24}"  # Keep hourly backups for 24 hours

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Helper functions
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# 1. Backup database (incremental - only changed data)
log "Starting incremental backup..."

# Check if database is accessible
if docker-compose -f "$PHOENIX_DIR/docker-compose.yml" exec -T db pg_isready -U phoenix >/dev/null 2>&1; then
    # Export only recent changes (last hour)
    docker-compose -f "$PHOENIX_DIR/docker-compose.yml" exec -T db psql -U phoenix -c "
        COPY (
            SELECT * FROM experiments WHERE updated_at > NOW() - INTERVAL '1 hour'
        ) TO STDOUT WITH CSV HEADER;
    " > "$BACKUP_DIR/experiments_${TIMESTAMP}.csv" 2>/dev/null || true
    
    docker-compose -f "$PHOENIX_DIR/docker-compose.yml" exec -T db psql -U phoenix -c "
        COPY (
            SELECT * FROM agent_tasks WHERE created_at > NOW() - INTERVAL '1 hour'
        ) TO STDOUT WITH CSV HEADER;
    " > "$BACKUP_DIR/agent_tasks_${TIMESTAMP}.csv" 2>/dev/null || true
    
    docker-compose -f "$PHOENIX_DIR/docker-compose.yml" exec -T db psql -U phoenix -c "
        COPY (
            SELECT * FROM metrics_cache WHERE timestamp > NOW() - INTERVAL '1 hour'
        ) TO STDOUT WITH CSV HEADER;
    " > "$BACKUP_DIR/metrics_cache_${TIMESTAMP}.csv" 2>/dev/null || true
    
    # Backup current sequence values
    docker-compose -f "$PHOENIX_DIR/docker-compose.yml" exec -T db psql -U phoenix -t -c "
        SELECT sequence_name, last_value 
        FROM information_schema.sequences 
        JOIN pg_sequences ON sequence_name = schemaname || '.' || sequencename;
    " > "$BACKUP_DIR/sequences_${TIMESTAMP}.txt" 2>/dev/null || true
    
    log "Database incremental backup completed"
else
    log "WARNING: Database not accessible, skipping incremental backup"
fi

# 2. Backup recent logs (last hour)
if [[ -d "$PHOENIX_DIR/data/logs" ]]; then
    find "$PHOENIX_DIR/data/logs" -name "*.log" -mmin -60 -exec cp {} "$BACKUP_DIR/" \; 2>/dev/null || true
fi

# 3. Save current metrics state
curl -s http://localhost:9090/api/v1/query?query=phoenix_cost_savings_total > \
    "$BACKUP_DIR/cost_savings_${TIMESTAMP}.json" 2>/dev/null || true

# 4. Clean old incremental backups
find "$BACKUP_DIR" -name "*.csv" -mmin +$((RETENTION_HOURS * 60)) -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "*.txt" -mmin +$((RETENTION_HOURS * 60)) -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "*.json" -mmin +$((RETENTION_HOURS * 60)) -delete 2>/dev/null || true

# 5. Create incremental manifest
cat > "$BACKUP_DIR/manifest_${TIMESTAMP}.json" << EOF
{
  "type": "incremental",
  "timestamp": "$TIMESTAMP",
  "files": [
    "experiments_${TIMESTAMP}.csv",
    "agent_tasks_${TIMESTAMP}.csv",
    "metrics_cache_${TIMESTAMP}.csv",
    "sequences_${TIMESTAMP}.txt",
    "cost_savings_${TIMESTAMP}.json"
  ]
}
EOF

log "Incremental backup completed successfully"