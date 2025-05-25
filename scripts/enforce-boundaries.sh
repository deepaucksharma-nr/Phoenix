#!/bin/bash
# enforce-boundaries.sh - Enforce monorepo boundaries during development

set -euo pipefail

echo "=== Enforcing Monorepo Boundaries ==="

# Run boundary validation
if ! ./scripts/validate-boundaries.sh; then
    echo ""
    echo "❌ Boundary violations detected!"
    echo ""
    echo "To fix violations:"
    echo "1. Move shared code to /packages/*"
    echo "2. Update imports to use the shared packages"
    echo "3. Ensure projects don't import from each other"
    exit 1
fi

# Check interface consistency
echo ""
echo "Checking interface consistency..."

# Ensure all projects that implement interfaces use the shared definitions
for project in projects/*/; do
    if [[ -d "$project/internal" ]]; then
        project_name=$(basename "$project")
        
        # Check if the project should implement certain interfaces
        case "$project_name" in
            "controller")
                if ! grep -r "interfaces.ExperimentService" "$project" > /dev/null 2>&1; then
                    echo "⚠️  WARNING: $project_name should implement ExperimentService interface"
                fi
                ;;
            "generator")
                if ! grep -r "interfaces.ConfigGenerator" "$project" > /dev/null 2>&1; then
                    echo "⚠️  WARNING: $project_name should implement ConfigGenerator interface"
                fi
                ;;
        esac
    fi
done

# Update go.work.sum if needed
echo ""
echo "Syncing Go workspace..."
go work sync

# Generate interface documentation
echo ""
echo "Generating interface documentation..."
cat > docs/INTERFACE_CONTRACTS.md << 'EOF'
# Phoenix Platform Interface Contracts

This document describes the interfaces that define contracts between services in the Phoenix Platform.

## Core Interfaces

### ExperimentService
Location: `packages/go-common/interfaces/experiment.go`

Implemented by:
- `projects/controller` - Core experiment management
- `projects/api` - REST API exposure

Consumed by:
- `projects/phoenix-cli` - CLI commands
- `projects/dashboard` - Web UI

### ConfigGenerator
Location: `packages/go-common/interfaces/pipeline.go`

Implemented by:
- `projects/generator` - Pipeline configuration generation

Consumed by:
- `projects/controller` - Experiment controller

### EventBus
Location: `packages/go-common/interfaces/events.go`

Implemented by:
- `packages/go-common/eventbus` - In-memory implementation

Consumed by:
- All services for event propagation

## Boundary Rules

1. **No Cross-Project Imports**: Projects cannot import from other projects
2. **Shared Code in Packages**: All shared code must be in `/packages/*`
3. **Interface-Based Communication**: Services communicate through defined interfaces
4. **Proto/OpenAPI Contracts**: External APIs defined in `/packages/contracts/*`

## Validation

Run `./scripts/validate-boundaries.sh` to check boundary compliance.
EOF

echo "✅ Boundary enforcement complete!"
echo ""
echo "Add this to your pre-commit hooks:"
echo "  ./scripts/enforce-boundaries.sh"