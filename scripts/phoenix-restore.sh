#!/bin/bash
# Phoenix restore script - restores Phoenix from backup

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔄 Phoenix Restore"
echo "=================="
echo ""

# Check arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 <backup-file> [options]"
    echo ""
    echo "Options:"
    echo "  --skip-database    Skip database restoration"
    echo "  --skip-config      Skip configuration restoration"
    echo "  --skip-data        Skip data files restoration"
    echo "  --force            Don't ask for confirmation"
    echo ""
    exit 1
fi

# Configuration
BACKUP_FILE="$1"
shift
PHOENIX_DATA_DIR="${PHOENIX_DATA_DIR:-/var/lib/phoenix}"
TEMP_DIR="/tmp/phoenix-restore-$$"

# Parse options
SKIP_DATABASE=false
SKIP_CONFIG=false
SKIP_DATA=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-database)
            SKIP_DATABASE=true
            shift
            ;;
        --skip-config)
            SKIP_CONFIG=true
            shift
            ;;
        --skip-data)
            SKIP_DATA=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validate backup file
if [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}Backup file not found: $BACKUP_FILE${NC}"
    exit 1
fi

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ] && ! sudo -n true 2>/dev/null; then
    echo -e "${RED}This script requires sudo access${NC}"
    exit 1
fi

# Confirmation
if [ "$FORCE" = false ]; then
    echo -e "${RED}WARNING: This will restore Phoenix from backup!${NC}"
    echo "This operation will:"
    [ "$SKIP_DATABASE" = false ] && echo "  - Replace the current database"
    [ "$SKIP_CONFIG" = false ] && echo "  - Replace configuration files"
    [ "$SKIP_DATA" = false ] && echo "  - Replace data files"
    echo ""
    echo -n "Are you sure you want to continue? (yes/N): "
    read -r response
    if [ "$response" != "yes" ]; then
        echo "Restore cancelled"
        exit 0
    fi
fi

# Stop Phoenix services
echo -e "\n${YELLOW}Stopping Phoenix services...${NC}"
sudo systemctl stop phoenix-api 2>/dev/null || true
sudo systemctl stop phoenix-agent 2>/dev/null || true
sudo systemctl stop otel-collector 2>/dev/null || true

# Extract backup
echo -e "\n${YELLOW}Extracting backup...${NC}"
mkdir -p "$TEMP_DIR"
tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"

# Find the backup directory
BACKUP_DIR=$(find "$TEMP_DIR" -maxdepth 1 -type d -name "phoenix_backup_*" | head -1)
if [ -z "$BACKUP_DIR" ]; then
    echo -e "${RED}Invalid backup file format${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Check manifest
if [ -f "$BACKUP_DIR/manifest.json" ]; then
    echo -e "\n${BLUE}Backup Information:${NC}"
    cat "$BACKUP_DIR/manifest.json" | jq '.'
    echo ""
fi

# Restore database
if [ "$SKIP_DATABASE" = false ] && [ -f "$BACKUP_DIR/database.sql" ]; then
    echo -e "\n${YELLOW}Restoring database...${NC}"
    
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
    
    # Drop and recreate database
    echo "Recreating database..."
    PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres << EOF
DROP DATABASE IF EXISTS $DB_NAME;
CREATE DATABASE $DB_NAME;
EOF
    
    # Restore database
    echo "Restoring database content..."
    PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" < "$BACKUP_DIR/database.sql"
    
    echo -e "${GREEN}✓ Database restored${NC}"
fi

# Restore configuration
if [ "$SKIP_CONFIG" = false ] && [ -f "$BACKUP_DIR/config.tar.gz" ]; then
    echo -e "\n${YELLOW}Restoring configuration...${NC}"
    
    # Backup current config
    sudo tar -czf "/tmp/phoenix-config-backup-$(date +%Y%m%d_%H%M%S).tar.gz" -C /etc phoenix 2>/dev/null || true
    
    # Extract new config
    sudo tar -xzf "$BACKUP_DIR/config.tar.gz" -C /etc
    
    # Restore permissions
    sudo chown -R root:root /etc/phoenix
    sudo chmod 600 /etc/phoenix/*.env 2>/dev/null || true
    sudo chmod 600 /etc/phoenix/certs/* 2>/dev/null || true
    
    echo -e "${GREEN}✓ Configuration restored${NC}"
    
    if [ -f "$BACKUP_DIR/env.template" ]; then
        echo ""
        echo "Note: Environment template saved to $BACKUP_DIR/env.template"
        echo "You may need to update sensitive values in /etc/phoenix/*.env"
    fi
fi

# Restore data files
if [ "$SKIP_DATA" = false ] && [ -f "$BACKUP_DIR/data.tar.gz" ]; then
    echo -e "\n${YELLOW}Restoring data files...${NC}"
    
    # Backup current data
    if [ -d "$PHOENIX_DATA_DIR" ]; then
        sudo tar -czf "/tmp/phoenix-data-backup-$(date +%Y%m%d_%H%M%S).tar.gz" -C "$PHOENIX_DATA_DIR" . 2>/dev/null || true
    fi
    
    # Create data directory
    sudo mkdir -p "$PHOENIX_DATA_DIR"
    
    # Extract data
    sudo tar -xzf "$BACKUP_DIR/data.tar.gz" -C "$PHOENIX_DATA_DIR"
    
    # Fix permissions
    sudo chown -R phoenix:phoenix "$PHOENIX_DATA_DIR"
    
    echo -e "${GREEN}✓ Data files restored${NC}"
fi

# Restore Docker volumes
docker_volumes=$(find "$BACKUP_DIR" -name "phoenix*.tar.gz" -not -name "config.tar.gz" -not -name "data.tar.gz" 2>/dev/null)
if [ -n "$docker_volumes" ]; then
    echo -e "\n${YELLOW}Restoring Docker volumes...${NC}"
    
    for volume_backup in $docker_volumes; do
        volume_name=$(basename "$volume_backup" .tar.gz)
        echo "Restoring volume: $volume_name"
        
        # Create volume if it doesn't exist
        docker volume create "$volume_name" 2>/dev/null || true
        
        # Restore volume data
        docker run --rm -v "$volume_name":/data -v "$BACKUP_DIR":/backup \
            alpine tar -xzf "/backup/$(basename "$volume_backup")" -C /data
    done
    
    echo -e "${GREEN}✓ Docker volumes restored${NC}"
fi

# Clean up
echo -e "\n${YELLOW}Cleaning up...${NC}"
rm -rf "$TEMP_DIR"

# Start services
echo -e "\n${YELLOW}Starting Phoenix services...${NC}"
sudo systemctl start phoenix-api 2>/dev/null || echo "Phoenix API not configured"
sudo systemctl start phoenix-agent 2>/dev/null || echo "Phoenix Agent not configured"

# Wait for services
sleep 5

# Verify restoration
echo -e "\n${YELLOW}Verifying restoration...${NC}"
if systemctl is-active --quiet phoenix-api; then
    echo -e "Phoenix API: ${GREEN}✓ Running${NC}"
    
    # Check API health
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "API Health: ${GREEN}✓ Healthy${NC}"
    else
        echo -e "API Health: ${RED}✗ Unhealthy${NC}"
    fi
else
    echo -e "Phoenix API: ${RED}✗ Not running${NC}"
fi

if systemctl is-active --quiet phoenix-agent; then
    echo -e "Phoenix Agent: ${GREEN}✓ Running${NC}"
else
    echo -e "Phoenix Agent: ${YELLOW}⚠ Not running (may not be configured)${NC}"
fi

# Summary
echo -e "\n${GREEN}✅ Restore complete!${NC}"
echo ""
echo "Please verify:"
echo "  1. Check service logs: journalctl -u phoenix-api -f"
echo "  2. Update any sensitive configuration in /etc/phoenix/*.env"
echo "  3. Verify data integrity in the application"
echo ""
echo "Backup files saved in /tmp/ for safety"