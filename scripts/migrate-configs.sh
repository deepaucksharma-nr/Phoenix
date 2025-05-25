#!/bin/bash
# migrate-configs.sh - Migrate configuration files
# Usage: ./migrate-configs.sh <config-type>

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Validate arguments
if [ $# -lt 1 ]; then
    echo -e "${RED}Usage: $0 <config-type>${NC}"
    echo "Valid types: monitoring, otel, control, production"
    exit 1
fi

CONFIG_TYPE=$1
OLD_PATH="OLD_IMPLEMENTATION/configs/$CONFIG_TYPE"
NEW_PATH="configs/$CONFIG_TYPE"

echo -e "${YELLOW}Migrating $CONFIG_TYPE configs from $OLD_PATH to $NEW_PATH${NC}"

# Check if old path exists
if [ ! -d "$OLD_PATH" ]; then
    echo -e "${RED}Error: $OLD_PATH does not exist${NC}"
    exit 1
fi

# Create new directory structure
echo "Creating directory structure..."
mkdir -p "$NEW_PATH"

# Copy configuration files
echo "Copying configuration files..."
cp -r "$OLD_PATH"/* "$NEW_PATH/" 2>/dev/null || true

# Special handling for different config types
case "$CONFIG_TYPE" in
    monitoring)
        # Ensure proper structure for monitoring
        mkdir -p "$NEW_PATH/prometheus/rules"
        mkdir -p "$NEW_PATH/prometheus/alerts"
        mkdir -p "$NEW_PATH/grafana/dashboards"
        mkdir -p "$NEW_PATH/grafana/provisioning"
        
        # Move prometheus configs if needed
        if [ -f "$NEW_PATH/prometheus.yaml" ]; then
            mv "$NEW_PATH/prometheus.yaml" "$NEW_PATH/prometheus/"
        fi
        ;;
        
    otel)
        # Ensure proper structure for OpenTelemetry
        mkdir -p "$NEW_PATH/collectors"
        mkdir -p "$NEW_PATH/processors"
        mkdir -p "$NEW_PATH/exporters"
        mkdir -p "$NEW_PATH/receivers"
        ;;
        
    control)
        # Ensure proper structure for control configs
        mkdir -p "$NEW_PATH/profiles"
        mkdir -p "$NEW_PATH/policies"
        ;;
        
    production)
        # Ensure proper structure for production configs
        mkdir -p "$NEW_PATH/secrets"
        mkdir -p "$NEW_PATH/certificates"
        mkdir -p "$NEW_PATH/environments"
        
        # Create .gitignore for sensitive files
        cat > "$NEW_PATH/.gitignore" << EOF
# Sensitive files
secrets/*
certificates/*
*.key
*.pem
*.crt
.env
EOF
        ;;
esac

# Update any hardcoded paths
echo "Updating configuration paths..."
find "$NEW_PATH" -type f \( -name "*.yaml" -o -name "*.yml" -o -name "*.json" \) -exec sed -i '' \
    -e 's|OLD_IMPLEMENTATION/||g' \
    -e 's|configs/|config/|g' \
    {} + 2>/dev/null || true

# Create README for the config type
echo "Creating documentation..."
cat > "$NEW_PATH/README.md" << EOF
# $CONFIG_TYPE Configuration

## Overview
This directory contains $CONFIG_TYPE configuration files for the Phoenix platform.

## Structure
$(find "$NEW_PATH" -type d | sed "s|$NEW_PATH|.|" | grep -v "^.$" | sort)

## Usage
These configurations are used by various services in the Phoenix platform.
See individual configuration files for specific details.

## Environment Variables
Configuration files may reference environment variables for sensitive values.
Ensure all required variables are set before deployment.

## Validation
To validate configurations:
\`\`\`bash
make validate-configs CONFIG_TYPE=$CONFIG_TYPE
\`\`\`
EOF

# Summary
echo -e "${GREEN}✓ Migration completed successfully!${NC}"
echo ""
echo "Summary:"
echo "- Source: $OLD_PATH"
echo "- Destination: $NEW_PATH"
echo "- Files migrated: $(find $NEW_PATH -type f | wc -l | xargs)"
echo ""
echo "Next steps:"
echo "1. Review migrated configurations"
echo "2. Update any service-specific references"
echo "3. Test configurations with services"