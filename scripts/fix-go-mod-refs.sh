#!/bin/bash

# Fix all go.mod references to old packages location

echo "Fixing go.mod references..."

# List of go.mod files to update
GO_MOD_FILES=(
    "./projects/benchmark/go.mod"
    "./projects/pipeline-operator/go.mod"
    "./projects/platform-api/go.mod"
    "./projects/controller/go.mod"
    "./projects/anomaly-detector/go.mod"
    "./projects/control-actuator-go/go.mod"
    "./projects/loadsim-operator/go.mod"
    "./projects/analytics/go.mod"
)

for go_mod in "${GO_MOD_FILES[@]}"; do
    if [ -f "$go_mod" ]; then
        echo "Updating $go_mod"
        
        # Update replace directives
        sed -i.bak 's|../../packages/go-common|../../pkg/common|g' "$go_mod"
        sed -i.bak 's|../../packages/contracts|../../pkg/contracts|g' "$go_mod"
        
        # Update module paths
        sed -i.bak 's|github.com/phoenix-vnext/platform/packages/go-common|github.com/phoenix/platform/pkg/common|g' "$go_mod"
        sed -i.bak 's|github.com/phoenix/platform/packages/go-common|github.com/phoenix/platform/pkg/common|g' "$go_mod"
        sed -i.bak 's|github.com/phoenix-vnext/platform/packages/contracts|github.com/phoenix/platform/pkg/contracts|g' "$go_mod"
        sed -i.bak 's|github.com/phoenix/platform/packages/contracts|github.com/phoenix/platform/pkg/contracts|g' "$go_mod"
        
        # Remove backup
        rm -f "${go_mod}.bak"
    fi
done

echo "✅ Fixed all go.mod references"

# Also fix any Go source files
echo "Updating Go import paths..."
find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r go_file; do
    if grep -q "packages/go-common\|packages/contracts" "$go_file"; then
        sed -i.bak 's|github.com/phoenix-vnext/platform/packages/go-common|github.com/phoenix/platform/pkg/common|g' "$go_file"
        sed -i.bak 's|github.com/phoenix/platform/packages/go-common|github.com/phoenix/platform/pkg/common|g' "$go_file"
        sed -i.bak 's|github.com/phoenix-vnext/platform/packages/contracts|github.com/phoenix/platform/pkg/contracts|g' "$go_file"
        sed -i.bak 's|github.com/phoenix/platform/packages/contracts|github.com/phoenix/platform/pkg/contracts|g' "$go_file"
        rm -f "${go_file}.bak"
    fi
done

echo "✅ Updated all Go import paths"