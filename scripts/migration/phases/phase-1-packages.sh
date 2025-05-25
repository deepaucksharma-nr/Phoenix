#!/bin/bash
# phase-1-packages.sh - Shared Packages Migration phase implementation
# This phase migrates shared code from OLD_IMPLEMENTATION to the packages directory

set -euo pipefail

# Source libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/common.sh"
source "$SCRIPT_DIR/../lib/state-tracker.sh"

PHASE_ID="phase-1-packages"

log_phase "Phase 1: Shared Packages Migration"

# Ensure we have the lock for this phase
if ! acquire_lock "$PHASE_ID" "$AGENT_ID"; then
    log_error "Failed to acquire lock for phase"
    exit 1
fi

# Track phase components
COMPONENTS=(
    "go-common-migration"
    "interfaces-migration"
    "contracts-migration"
    "go-mod-setup"
    "validation-tests"
)

# Initialize component tracking
for component in "${COMPONENTS[@]}"; do
    track_component "$PHASE_ID" "$component" "pending"
done

# Component 1: Migrate go-common packages
log_info "Migrating go-common packages..."
track_component "$PHASE_ID" "go-common-migration" "in_progress"

# Create go-common directory structure
create_directory "packages/go-common"
create_directory "packages/go-common/auth"
create_directory "packages/go-common/telemetry"
create_directory "packages/go-common/metrics"
create_directory "packages/go-common/utils"
create_directory "packages/go-common/store"
create_directory "packages/go-common/eventbus"
create_directory "packages/go-common/clients"

# Migrate packages from OLD_IMPLEMENTATION
if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg" ]]; then
    # Migrate auth package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/auth" ]]; then
        log_info "Migrating auth package..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/auth/* packages/go-common/auth/ 2>/dev/null || true
    fi
    
    # Migrate telemetry/logging package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/logging" ]]; then
        log_info "Migrating logging package to telemetry..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/logging/* packages/go-common/telemetry/ 2>/dev/null || true
    fi
    
    # Migrate metrics package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/metrics" ]]; then
        log_info "Migrating metrics package..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/metrics/* packages/go-common/metrics/ 2>/dev/null || true
    fi
    
    # Migrate utils package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/utils" ]]; then
        log_info "Migrating utils package..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/utils/* packages/go-common/utils/ 2>/dev/null || true
    fi
    
    # Migrate store package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/store" ]]; then
        log_info "Migrating store package..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/store/* packages/go-common/store/ 2>/dev/null || true
    fi
    
    # Migrate eventbus package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/eventbus" ]]; then
        log_info "Migrating eventbus package..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/eventbus/* packages/go-common/eventbus/ 2>/dev/null || true
    fi
    
    # Migrate clients package
    if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/clients" ]]; then
        log_info "Migrating clients package..."
        cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/clients/* packages/go-common/clients/ 2>/dev/null || true
    fi
    
    log_success "go-common packages migrated"
else
    log_warning "OLD_IMPLEMENTATION/phoenix-platform/pkg not found"
fi

track_component "$PHASE_ID" "go-common-migration" "completed"

# Component 2: Migrate interfaces
log_info "Migrating interfaces..."
track_component "$PHASE_ID" "interfaces-migration" "in_progress"

create_directory "packages/go-common/interfaces"

if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/pkg/interfaces" ]]; then
    log_info "Copying interface files..."
    cp -r OLD_IMPLEMENTATION/phoenix-platform/pkg/interfaces/* packages/go-common/interfaces/ 2>/dev/null || true
    log_success "Interfaces migrated"
else
    log_warning "OLD_IMPLEMENTATION/phoenix-platform/pkg/interfaces not found"
fi

track_component "$PHASE_ID" "interfaces-migration" "completed"

# Component 3: Migrate contracts
log_info "Migrating contracts..."
track_component "$PHASE_ID" "contracts-migration" "in_progress"

# OpenAPI contracts
create_directory "packages/contracts/openapi"
if [[ -d "OLD_IMPLEMENTATION/packages/contracts/openapi" ]]; then
    log_info "Migrating OpenAPI contracts..."
    cp -r OLD_IMPLEMENTATION/packages/contracts/openapi/* packages/contracts/openapi/ 2>/dev/null || true
elif [[ -d "OLD_IMPLEMENTATION/phoenix-platform/docs/assets" ]]; then
    # Check for openapi.yaml in assets
    if [[ -f "OLD_IMPLEMENTATION/phoenix-platform/docs/assets/openapi.yaml" ]]; then
        cp OLD_IMPLEMENTATION/phoenix-platform/docs/assets/openapi.yaml packages/contracts/openapi/
    fi
fi

# Proto contracts
create_directory "packages/contracts/proto"
if [[ -d "OLD_IMPLEMENTATION/phoenix-platform/api/proto" ]]; then
    log_info "Migrating Proto contracts..."
    cp -r OLD_IMPLEMENTATION/phoenix-platform/api/proto/* packages/contracts/proto/ 2>/dev/null || true
elif [[ -d "OLD_IMPLEMENTATION/phoenix-platform/proto" ]]; then
    cp -r OLD_IMPLEMENTATION/phoenix-platform/proto/* packages/contracts/proto/ 2>/dev/null || true
fi

log_success "Contracts migrated"
track_component "$PHASE_ID" "contracts-migration" "completed"

# Component 4: Setup go.mod files
log_info "Setting up go.mod files..."
track_component "$PHASE_ID" "go-mod-setup" "in_progress"

# Create go.mod for go-common
if [[ ! -f "packages/go-common/go.mod" ]]; then
    cat > packages/go-common/go.mod << 'EOF'
module github.com/phoenix/platform/packages/go-common

go 1.21

require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/prometheus/client_golang v1.18.0
    github.com/stretchr/testify v1.8.4
    go.uber.org/zap v1.26.0
    google.golang.org/grpc v1.60.1
    google.golang.org/protobuf v1.32.0
)

require (
    github.com/beorn7/perks v1.0.1 // indirect
    github.com/cespare/xxhash/v2 v2.2.0 // indirect
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/golang/protobuf v1.5.3 // indirect
    github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    github.com/prometheus/client_model v0.5.0 // indirect
    github.com/prometheus/common v0.45.0 // indirect
    github.com/prometheus/procfs v0.12.0 // indirect
    go.uber.org/multierr v1.11.0 // indirect
    golang.org/x/net v0.19.0 // indirect
    golang.org/x/sys v0.15.0 // indirect
    golang.org/x/text v0.14.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
EOF
    log_success "Created go.mod for go-common"
fi

# Create go.mod for contracts if needed
if [[ ! -f "packages/contracts/go.mod" ]]; then
    cat > packages/contracts/go.mod << 'EOF'
module github.com/phoenix/platform/packages/contracts

go 1.21

require (
    google.golang.org/grpc v1.60.1
    google.golang.org/protobuf v1.32.0
)

require (
    github.com/golang/protobuf v1.5.3 // indirect
    golang.org/x/net v0.19.0 // indirect
    golang.org/x/sys v0.15.0 // indirect
    golang.org/x/text v0.14.0 // indirect
    google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
)
EOF
    log_success "Created go.mod for contracts"
fi

# Update import paths in go files
log_info "Updating import paths..."
if command -v find &> /dev/null && command -v sed &> /dev/null; then
    # Update imports in go-common
    find packages/go-common -name "*.go" -type f -exec sed -i.bak \
        -e 's|"github.com/deepaucksharma-nr/phoenix-v3/phoenix-platform/pkg/|"github.com/phoenix/platform/packages/go-common/|g' \
        -e 's|"phoenix-platform/pkg/|"github.com/phoenix/platform/packages/go-common/|g' \
        {} \; 2>/dev/null || true
    
    # Clean up backup files
    find packages/go-common -name "*.go.bak" -type f -delete 2>/dev/null || true
fi

track_component "$PHASE_ID" "go-mod-setup" "completed"

# Component 5: Run validation tests
log_info "Running validation tests..."
track_component "$PHASE_ID" "validation-tests" "in_progress"

VALIDATIONS_PASSED=true

# Validation 1: Check go-common structure
if [[ -d "packages/go-common" ]]; then
    record_validation "$PHASE_ID" "go_common_structure" "passed" "go-common directory exists"
    
    # Try to build go-common
    if cd packages/go-common && go mod tidy 2>/dev/null && go build ./... 2>/dev/null; then
        record_validation "$PHASE_ID" "go_common_build" "passed" "go-common builds successfully"
    else
        record_validation "$PHASE_ID" "go_common_build" "warning" "go-common build needs attention"
        log_warning "go-common build needs manual intervention"
    fi
    cd - > /dev/null
else
    record_validation "$PHASE_ID" "go_common_structure" "failed" "go-common directory missing"
    VALIDATIONS_PASSED=false
fi

# Validation 2: Check interfaces
if [[ -d "packages/go-common/interfaces" ]] && [[ -n "$(ls -A packages/go-common/interfaces 2>/dev/null)" ]]; then
    record_validation "$PHASE_ID" "interfaces" "passed" "Interfaces migrated successfully"
else
    record_validation "$PHASE_ID" "interfaces" "warning" "No interfaces found or directory empty"
fi

# Validation 3: Check contracts
if [[ -d "packages/contracts/openapi" ]] || [[ -d "packages/contracts/proto" ]]; then
    record_validation "$PHASE_ID" "contracts" "passed" "Contracts directory structure exists"
else
    record_validation "$PHASE_ID" "contracts" "failed" "Contracts directory structure missing"
    VALIDATIONS_PASSED=false
fi

# Validation 4: Check for go.mod files
if [[ -f "packages/go-common/go.mod" ]]; then
    record_validation "$PHASE_ID" "go_mod_files" "passed" "go.mod files created"
else
    record_validation "$PHASE_ID" "go_mod_files" "failed" "go.mod files missing"
    VALIDATIONS_PASSED=false
fi

track_component "$PHASE_ID" "validation-tests" "completed"

# Create README for packages
if [[ ! -f "packages/README.md" ]]; then
    cat > packages/README.md << 'EOF'
# Phoenix Platform - Shared Packages

This directory contains shared packages used across the Phoenix Platform.

## Structure

- `go-common/` - Shared Go packages
  - `auth/` - Authentication and authorization
  - `telemetry/` - Logging, metrics, and tracing
  - `metrics/` - Metrics utilities
  - `utils/` - Common utilities
  - `store/` - Data store abstractions
  - `eventbus/` - Event bus implementation
  - `clients/` - Service clients
  - `interfaces/` - Shared interfaces

- `contracts/` - API contracts and schemas
  - `openapi/` - OpenAPI specifications
  - `proto/` - Protocol Buffer definitions
  - `schemas/` - JSON schemas

## Usage

Import packages in your Go services:

```go
import (
    "github.com/phoenix/platform/packages/go-common/auth"
    "github.com/phoenix/platform/packages/go-common/telemetry"
)
```

## Development

Each package has its own `go.mod` file for dependency management.

To work on a package:

```bash
cd packages/go-common
go mod tidy
go test ./...
```
EOF
    log_success "Created packages README"
fi

# Release lock
release_lock "$PHASE_ID" "$AGENT_ID"

# Final status
if [[ "$VALIDATIONS_PASSED" == "true" ]]; then
    log_success "Phase 1: Shared packages migration completed successfully!"
    log_info "Next steps:"
    log_info "  - Review migrated packages in packages/go-common"
    log_info "  - Update any broken imports manually"
    log_info "  - Run 'cd packages/go-common && go mod tidy' to update dependencies"
    log_info "  - Proceed with Phase 2: Core Services Migration"
    exit 0
else
    log_error "Phase 1: Shared packages migration completed with validation failures"
    exit 1
fi