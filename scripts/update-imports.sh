#!/bin/bash
# update-imports.sh - Update import paths to match new monorepo structure

set -euo pipefail

echo "=== Updating Go imports to match monorepo structure ==="

# Function to update imports in a Go file
update_go_imports() {
    local file=$1
    local temp_file="${file}.tmp"
    
    # Skip if file doesn't exist
    [[ ! -f "$file" ]] && return
    
    # Update imports
    sed -E \
        -e 's|"github\.com/phoenix/platform/cmd/controller/|"github.com/phoenix/platform/projects/controller/|g' \
        -e 's|"github\.com/phoenix/platform/cmd/api-gateway/|"github.com/phoenix/platform/projects/api/|g' \
        -e 's|"github\.com/phoenix/platform/cmd/generator/|"github.com/phoenix/platform/projects/generator/|g' \
        -e 's|"github\.com/phoenix/platform/cmd/phoenix-cli/|"github.com/phoenix/platform/projects/phoenix-cli/|g' \
        -e 's|"github\.com/phoenix/platform/operators/|"github.com/phoenix/platform/projects/|g' \
        -e 's|"github\.com/phoenix/platform/pkg/api/|"github.com/phoenix/platform/packages/contracts/proto/|g' \
        -e 's|"github\.com/phoenix/platform/pkg/|"github.com/phoenix/platform/packages/go-common/|g' \
        -e 's|phoenix-platform/pkg/|packages/go-common/|g' \
        -e 's|phoenix-platform/cmd/|projects/|g' \
        "$file" > "$temp_file"
    
    # Only update if changes were made
    if ! diff -q "$file" "$temp_file" > /dev/null 2>&1; then
        mv "$temp_file" "$file"
        echo "Updated: $file"
    else
        rm -f "$temp_file"
    fi
}

# Update imports in all Go files in projects directory
echo "Updating imports in projects..."
find projects -name "*.go" -type f | while read -r file; do
    update_go_imports "$file"
done

# Update imports in packages directory
echo "Updating imports in packages..."
find packages -name "*.go" -type f | while read -r file; do
    update_go_imports "$file"
done

# Update go.mod files to have correct module paths and replace directives
echo "Updating go.mod files..."

# Update projects go.mod files
for project_dir in projects/*/; do
    if [[ -f "$project_dir/go.mod" ]]; then
        project_name=$(basename "$project_dir")
        
        # Update module path
        sed -i.bak "s|^module .*|module github.com/phoenix/platform/projects/$project_name|" "$project_dir/go.mod"
        
        # Add replace directives if not present
        if ! grep -q "replace github.com/phoenix/platform/packages/go-common" "$project_dir/go.mod"; then
            echo "" >> "$project_dir/go.mod"
            echo "replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common" >> "$project_dir/go.mod"
            echo "replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts" >> "$project_dir/go.mod"
        fi
        
        rm -f "$project_dir/go.mod.bak"
        echo "Updated go.mod for $project_name"
    fi
done

# Create go.work file if it doesn't exist
if [[ ! -f "go.work" ]]; then
    echo "Creating go.work file..."
    cat > go.work << 'EOF'
go 1.21

use (
    ./packages/go-common
    ./packages/contracts
    ./projects/analytics
    ./projects/anomaly-detector
    ./projects/api
    ./projects/benchmark
    ./projects/collector
    ./projects/control-actuator-go
    ./projects/controller
    ./projects/dashboard
    ./projects/generator
    ./projects/loadsim-operator
    ./projects/phoenix-cli
    ./projects/pipeline-operator
    ./projects/platform-api
)
EOF
fi

echo "=== Import update complete ==="
echo ""
echo "Next steps:"
echo "1. Run 'go work sync' to sync workspace"
echo "2. Run 'go mod tidy' in each project directory"
echo "3. Test builds with 'go build ./...' in each project"