#!/bin/bash

# Cleanup duplicate services that have been migrated to projects/

set -e

echo "🧹 Cleaning up duplicate services..."

# Services that have been successfully migrated
MIGRATED_SERVICES=(
    "api"           # -> projects/platform-api
    "controller"    # -> projects/controller
    "generator"     # -> projects/generator
    "analytics"     # -> projects/analytics
    "benchmark"     # -> projects/benchmark
    "anomaly-detector" # -> projects/anomaly-detector
    "validator"     # -> projects/validator
)

# Check for uncommitted changes
if [[ -n $(git status --porcelain services/) ]]; then
    echo "⚠️  Warning: Uncommitted changes found in services/"
    echo "Please commit or stash changes before cleaning up."
    git status --short services/
    echo ""
    read -p "Do you want to discard these changes and continue? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cleanup cancelled."
        exit 1
    fi
fi

# Remove migrated services
for service in "${MIGRATED_SERVICES[@]}"; do
    if [ -d "services/$service" ]; then
        echo "🗑️  Removing services/$service (migrated to projects/)"
        git rm -rf "services/$service" 2>/dev/null || rm -rf "services/$service"
    fi
done

# List remaining services
echo ""
echo "📁 Remaining services in services/ directory:"
if [ -d "services" ]; then
    ls -la services/ 2>/dev/null || echo "No services directory found"
else
    echo "Services directory has been removed"
fi

echo ""
echo "✅ Cleanup complete!"
echo ""
echo "Note: The following mappings were used:"
echo "  - services/api -> projects/platform-api"
echo "  - services/controller -> projects/controller" 
echo "  - services/generator -> projects/generator"
echo "  - services/analytics -> projects/analytics"
echo "  - services/benchmark -> projects/benchmark"
echo "  - services/anomaly-detector -> projects/anomaly-detector"
echo "  - services/validator -> projects/validator"