#!/bin/bash

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/phoenix?sslmode=disable}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./migrations}"
ACTION="${1:-up}"

echo -e "${GREEN}Phoenix Platform Database Migration Tool${NC}"
echo "========================================"

# Parse database URL
if [[ $DATABASE_URL =~ postgres://([^:]+):([^@]+)@([^:]+):([^/]+)/([^?]+) ]]; then
    DB_USER="${BASH_REMATCH[1]}"
    DB_PASS="${BASH_REMATCH[2]}"
    DB_HOST="${BASH_REMATCH[3]}"
    DB_PORT="${BASH_REMATCH[4]}"
    DB_NAME="${BASH_REMATCH[5]}"
else
    echo -e "${RED}Error: Invalid DATABASE_URL format${NC}"
    exit 1
fi

# Function to run SQL file
run_sql_file() {
    local file=$1
    echo -e "${YELLOW}Running migration: $(basename "$file")${NC}"
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$file"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Migration completed successfully${NC}"
    else
        echo -e "${RED}✗ Migration failed${NC}"
        exit 1
    fi
}

# Function to create migration table
create_migration_table() {
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME <<EOF
CREATE TABLE IF NOT EXISTS phoenix.schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
EOF
}

# Function to check if migration is applied
is_migration_applied() {
    local version=$1
    local result=$(PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM phoenix.schema_migrations WHERE version = '$version';")
    [ "$result" -gt 0 ]
}

# Function to record migration
record_migration() {
    local version=$1
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "INSERT INTO phoenix.schema_migrations (version) VALUES ('$version');"
}

# Function to remove migration record
remove_migration() {
    local version=$1
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM phoenix.schema_migrations WHERE version = '$version';"
}

case $ACTION in
    up)
        echo "Running migrations..."
        
        # Create migration table if it doesn't exist
        create_migration_table
        
        # Run all migrations in order
        for migration in $(ls $MIGRATIONS_DIR/*.sql | sort); do
            version=$(basename "$migration" .sql)
            
            if is_migration_applied "$version"; then
                echo -e "${YELLOW}Skipping migration $version (already applied)${NC}"
            else
                run_sql_file "$migration"
                record_migration "$version"
            fi
        done
        
        echo -e "${GREEN}All migrations completed!${NC}"
        ;;
        
    down)
        echo "Rolling back last migration..."
        
        # Get the last applied migration
        last_migration=$(PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT version FROM phoenix.schema_migrations ORDER BY applied_at DESC LIMIT 1;" | xargs)
        
        if [ -z "$last_migration" ]; then
            echo -e "${YELLOW}No migrations to rollback${NC}"
            exit 0
        fi
        
        # Look for down migration file
        down_file="$MIGRATIONS_DIR/${last_migration}_down.sql"
        if [ -f "$down_file" ]; then
            run_sql_file "$down_file"
            remove_migration "$last_migration"
            echo -e "${GREEN}Rolled back migration: $last_migration${NC}"
        else
            echo -e "${RED}No rollback file found for migration: $last_migration${NC}"
            exit 1
        fi
        ;;
        
    status)
        echo "Migration status:"
        PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version, applied_at FROM phoenix.schema_migrations ORDER BY applied_at;"
        ;;
        
    create)
        if [ -z "${2:-}" ]; then
            echo -e "${RED}Error: Migration name required${NC}"
            echo "Usage: $0 create <migration_name>"
            exit 1
        fi
        
        timestamp=$(date +%Y%m%d%H%M%S)
        filename="$MIGRATIONS_DIR/${timestamp}_${2}.sql"
        
        cat > "$filename" <<EOF
-- Migration: ${2}
-- Created: $(date)

SET search_path TO phoenix, public;

-- Add your migration SQL here

EOF
        
        echo -e "${GREEN}Created migration: $filename${NC}"
        ;;
        
    *)
        echo "Usage: $0 {up|down|status|create}"
        echo ""
        echo "Commands:"
        echo "  up      - Run all pending migrations"
        echo "  down    - Rollback the last migration"
        echo "  status  - Show migration status"
        echo "  create  - Create a new migration file"
        exit 1
        ;;
esac