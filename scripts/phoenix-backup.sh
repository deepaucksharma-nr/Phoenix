#!/bin/bash
# Phoenix backup script - backs up all Phoenix data

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "💾 Phoenix Backup"
echo "================="
echo ""

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/phoenix}"
PHOENIX_DATA_DIR="${PHOENIX_DATA_DIR:-/var/lib/phoenix}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="phoenix_backup_${TIMESTAMP}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Parse arguments
BACKUP_TYPE="full"
COMPRESS=true
INCLUDE_METRICS=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --incremental)
            BACKUP_TYPE="incremental"
            shift
            ;;
        --no-compress)
            COMPRESS=false
            shift
            ;;
        --include-metrics)
            INCLUDE_METRICS=true
            shift
            ;;
        --output)
            BACKUP_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--incremental] [--no-compress] [--include-metrics] [--output DIR]"
            exit 1
            ;;
    esac
done

# Create backup directory
echo -e "${YELLOW}Creating backup directory...${NC}"
sudo mkdir -p "$BACKUP_DIR"
BACKUP_PATH="$BACKUP_DIR/$BACKUP_NAME"
sudo mkdir -p "$BACKUP_PATH"

# Function to backup database
backup_database() {
    echo -e "\n${YELLOW}Backing up database...${NC}"
    
    # Get database credentials
    if [ -f /etc/phoenix/phoenix-api.env ]; then
        source /etc/phoenix/phoenix-api.env
        DB_URL="$DATABASE_URL"
    else
        DB_URL="postgresql://phoenix:phoenix@localhost:5432/phoenix"
    fi
    
    # Parse database URL
    DB_HOST=$(echo $DB_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
    DB_PORT=$(echo $DB_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
    DB_NAME=$(echo $DB_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')
    DB_USER=$(echo $DB_URL | sed -n 's/.*\/\/\([^:]*\):.*/\1/p')
    DB_PASS=$(echo $DB_URL | sed -n 's/.*:\([^@]*\)@.*/\1/p')
    
    # Dump database
    if [ "$BACKUP_TYPE" = "incremental" ] && [ "$INCLUDE_METRICS" = false ]; then
        # Exclude metrics table for incremental backups
        echo "Creating incremental database backup (excluding metrics)..."
        PGPASSWORD="$DB_PASS" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            --exclude-table=metric_cache \
            -f "$BACKUP_PATH/database.sql"
    else
        echo "Creating full database backup..."
        PGPASSWORD="$DB_PASS" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
            -f "$BACKUP_PATH/database.sql"
    fi
    
    echo -e "${GREEN}✓ Database backed up${NC}"
}

# Function to backup configuration
backup_config() {
    echo -e "\n${YELLOW}Backing up configuration...${NC}"
    
    # Backup /etc/phoenix (excluding secrets)
    if [ -d /etc/phoenix ]; then
        sudo tar -czf "$BACKUP_PATH/config.tar.gz" \
            --exclude='*.key' \
            --exclude='*.pem' \
            --exclude='*.env' \
            -C /etc phoenix
        echo -e "${GREEN}✓ Configuration backed up${NC}"
    fi
    
    # Save environment variables (sanitized)
    if [ -f /etc/phoenix/phoenix-api.env ]; then
        sudo grep -v -E '(PASSWORD|SECRET|KEY)' /etc/phoenix/phoenix-api.env > "$BACKUP_PATH/env.template"
        echo -e "${GREEN}✓ Environment template saved${NC}"
    fi
}

# Function to backup data files
backup_data() {
    echo -e "\n${YELLOW}Backing up data files...${NC}"
    
    if [ -d "$PHOENIX_DATA_DIR" ]; then
        # Determine what to backup
        if [ "$BACKUP_TYPE" = "incremental" ]; then
            # Find files modified in last 24 hours
            echo "Finding recently modified files..."
            sudo find "$PHOENIX_DATA_DIR" -type f -mtime -1 -print0 | \
                sudo tar -czf "$BACKUP_PATH/data.tar.gz" --null -T -
        else
            # Full backup
            sudo tar -czf "$BACKUP_PATH/data.tar.gz" -C "$PHOENIX_DATA_DIR" .
        fi
        echo -e "${GREEN}✓ Data files backed up${NC}"
    fi
}

# Function to backup Docker volumes
backup_docker_volumes() {
    echo -e "\n${YELLOW}Backing up Docker volumes...${NC}"
    
    # Get Phoenix-related volumes
    volumes=$(docker volume ls -q | grep phoenix)
    
    for volume in $volumes; do
        echo "Backing up volume: $volume"
        docker run --rm -v $volume:/data -v "$BACKUP_PATH":/backup \
            alpine tar -czf "/backup/${volume}.tar.gz" -C /data .
    done
    
    if [ -n "$volumes" ]; then
        echo -e "${GREEN}✓ Docker volumes backed up${NC}"
    fi
}

# Function to create backup manifest
create_manifest() {
    echo -e "\n${YELLOW}Creating backup manifest...${NC}"
    
    cat > "$BACKUP_PATH/manifest.json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "type": "$BACKUP_TYPE",
  "version": "$(cat /usr/local/bin/phoenix-api 2>/dev/null | strings | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')",
  "hostname": "$(hostname)",
  "components": {
    "database": $([ -f "$BACKUP_PATH/database.sql" ] && echo "true" || echo "false"),
    "config": $([ -f "$BACKUP_PATH/config.tar.gz" ] && echo "true" || echo "false"),
    "data": $([ -f "$BACKUP_PATH/data.tar.gz" ] && echo "true" || echo "false"),
    "docker_volumes": $(ls "$BACKUP_PATH"/*.tar.gz 2>/dev/null | grep -v -E '(config|data)\.tar\.gz' | wc -l)
  },
  "metrics_included": $INCLUDE_METRICS
}
EOF
    
    echo -e "${GREEN}✓ Manifest created${NC}"
}

# Perform backup
echo -e "${BLUE}Starting $BACKUP_TYPE backup...${NC}"

# Check if running as root or with sudo access
if [ "$EUID" -ne 0 ] && ! sudo -n true 2>/dev/null; then
    echo -e "${RED}This script requires sudo access${NC}"
    exit 1
fi

# Backup components
backup_database
backup_config
backup_data
backup_docker_volumes
create_manifest

# Compress entire backup if requested
if [ "$COMPRESS" = true ]; then
    echo -e "\n${YELLOW}Compressing backup...${NC}"
    cd "$BACKUP_DIR"
    sudo tar -czf "${BACKUP_NAME}.tar.gz" "$BACKUP_NAME"
    sudo rm -rf "$BACKUP_NAME"
    FINAL_PATH="${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
    SIZE=$(du -h "$FINAL_PATH" | cut -f1)
    echo -e "${GREEN}✓ Backup compressed to $SIZE${NC}"
else
    FINAL_PATH="$BACKUP_PATH"
    SIZE=$(du -sh "$FINAL_PATH" | cut -f1)
fi

# Clean up old backups
echo -e "\n${YELLOW}Cleaning up old backups...${NC}"
find "$BACKUP_DIR" -name "phoenix_backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete
echo -e "${GREEN}✓ Old backups cleaned (retained: $RETENTION_DAYS days)${NC}"

# Summary
echo -e "\n${GREEN}✅ Backup complete!${NC}"
echo ""
echo "Backup details:"
echo "  Type: $BACKUP_TYPE"
echo "  Location: $FINAL_PATH"
echo "  Size: $SIZE"
echo "  Retention: $RETENTION_DAYS days"
echo ""
echo "To restore from this backup:"
echo "  ./phoenix-restore.sh $FINAL_PATH"