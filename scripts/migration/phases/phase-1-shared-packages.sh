#!/bin/bash
# phase-1-shared-packages.sh - Shared packages migration phase
# This phase migrates shared packages from OLD_IMPLEMENTATION to new structure

set -euo pipefail

# Source libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/common.sh"
source "$SCRIPT_DIR/../lib/state-tracker.sh"

PHASE_ID="phase-1-shared-packages"

log_phase "Phase 1: Shared Packages Migration"

# Ensure we have the lock for this phase
if ! acquire_lock "$PHASE_ID" "$AGENT_ID"; then
    log_error "Failed to acquire lock for phase"
    exit 1
fi

# Track phase components
COMPONENTS=(
    "config-packages"
    "common-utilities"
    "go-common-packages"
    "contracts"
    "package-setup"
)

# Initialize component tracking
for component in "${COMPONENTS[@]}"; do
    track_component "$PHASE_ID" "$component" "pending"
done

# Component 1: Migrate config packages
log_info "Migrating config packages..."
track_component "$PHASE_ID" "config-packages" "in_progress"

# Create config package structure
create_directory "packages/config/src"
create_directory "packages/config/src/environments"
create_directory "packages/config/src/schemas"
create_directory "packages/config/src/validators"

# Create package.json for config package
if [[ ! -f packages/config/package.json ]]; then
    cat > packages/config/package.json << 'EOF'
{
  "name": "@phoenix/config",
  "version": "0.1.0",
  "description": "Phoenix Platform configuration management",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "test:watch": "jest --watch",
    "lint": "eslint src --ext .ts",
    "clean": "rm -rf dist coverage"
  },
  "dependencies": {
    "joi": "^17.11.0",
    "dotenv": "^16.3.1",
    "js-yaml": "^4.1.0"
  },
  "devDependencies": {
    "@types/node": "^20.10.0",
    "@types/jest": "^29.5.0",
    "typescript": "^5.3.0",
    "jest": "^29.7.0",
    "ts-jest": "^29.1.0",
    "eslint": "^8.54.0"
  }
}
EOF
    log_success "Created config package.json"
fi

# Create TypeScript config
if [[ ! -f packages/config/tsconfig.json ]]; then
    cat > packages/config/tsconfig.json << 'EOF'
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
EOF
    log_success "Created config tsconfig.json"
fi

# Migrate config files from OLD_IMPLEMENTATION
if [[ -d OLD_IMPLEMENTATION/configs ]]; then
    log_info "Copying configuration files from OLD_IMPLEMENTATION..."
    
    # Copy YAML configs
    if [[ -d OLD_IMPLEMENTATION/configs/control ]]; then
        cp -r OLD_IMPLEMENTATION/configs/control/* packages/config/src/environments/ 2>/dev/null || true
    fi
    
    # Copy monitoring configs
    if [[ -d OLD_IMPLEMENTATION/configs/monitoring ]]; then
        create_directory "packages/config/src/monitoring"
        cp -r OLD_IMPLEMENTATION/configs/monitoring/* packages/config/src/monitoring/ 2>/dev/null || true
    fi
    
    # Copy OTEL configs
    if [[ -d OLD_IMPLEMENTATION/configs/otel ]]; then
        create_directory "packages/config/src/otel"
        cp -r OLD_IMPLEMENTATION/configs/otel/* packages/config/src/otel/ 2>/dev/null || true
    fi
    
    log_success "Configuration files migrated"
fi

track_component "$PHASE_ID" "config-packages" "completed"

# Component 2: Migrate common utilities
log_info "Migrating common utilities..."
track_component "$PHASE_ID" "common-utilities" "in_progress"

# Create common package structure
create_directory "packages/common/src"
create_directory "packages/common/src/utils"
create_directory "packages/common/src/types"
create_directory "packages/common/src/constants"

# Create package.json for common package
if [[ ! -f packages/common/package.json ]]; then
    cat > packages/common/package.json << 'EOF'
{
  "name": "@phoenix/common",
  "version": "0.1.0",
  "description": "Phoenix Platform common utilities and types",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "test:watch": "jest --watch",
    "lint": "eslint src --ext .ts",
    "clean": "rm -rf dist coverage"
  },
  "dependencies": {
    "lodash": "^4.17.21",
    "uuid": "^9.0.1",
    "winston": "^3.11.0"
  },
  "devDependencies": {
    "@types/node": "^20.10.0",
    "@types/lodash": "^4.14.202",
    "@types/uuid": "^9.0.7",
    "@types/jest": "^29.5.0",
    "typescript": "^5.3.0",
    "jest": "^29.7.0",
    "ts-jest": "^29.1.0"
  }
}
EOF
    log_success "Created common package.json"
fi

# Create index files
cat > packages/common/src/index.ts << 'EOF'
export * from './utils';
export * from './types';
export * from './constants';
EOF

track_component "$PHASE_ID" "common-utilities" "completed"

# Component 3: Migrate Go common packages
log_info "Migrating Go common packages..."
track_component "$PHASE_ID" "go-common-packages" "in_progress"

# Create go-common structure
create_directory "packages/go-common/pkg/logger"
create_directory "packages/go-common/pkg/metrics"
create_directory "packages/go-common/pkg/errors"
create_directory "packages/go-common/pkg/middleware"
create_directory "packages/go-common/pkg/database"
create_directory "packages/go-common/pkg/utils"

# Create go.mod for go-common
if [[ ! -f packages/go-common/go.mod ]]; then
    cat > packages/go-common/go.mod << 'EOF'
module github.com/phoenix-platform/phoenix/packages/go-common

go 1.21

require (
    go.uber.org/zap v1.26.0
    github.com/prometheus/client_golang v1.17.0
    github.com/google/uuid v1.4.0
    github.com/pkg/errors v0.9.1
)
EOF
    log_success "Created go-common go.mod"
fi

# Create README for go-common
cat > packages/go-common/README.md << 'EOF'
# Phoenix Go Common Packages

Shared Go packages for Phoenix Platform services.

## Packages

- `pkg/logger`: Structured logging with zap
- `pkg/metrics`: Prometheus metrics utilities
- `pkg/errors`: Error handling and wrapping
- `pkg/middleware`: Common HTTP/gRPC middleware
- `pkg/database`: Database connection utilities
- `pkg/utils`: General utilities

## Usage

```go
import (
    "github.com/phoenix-platform/phoenix/packages/go-common/pkg/logger"
    "github.com/phoenix-platform/phoenix/packages/go-common/pkg/metrics"
)
```
EOF

# Migrate existing Go common code if available
if [[ -d OLD_IMPLEMENTATION/pkg ]]; then
    log_info "Migrating existing Go common packages..."
    
    # Copy telemetry/logging
    if [[ -d OLD_IMPLEMENTATION/pkg/telemetry/logging ]]; then
        cp -r OLD_IMPLEMENTATION/pkg/telemetry/logging/* packages/go-common/pkg/logger/ 2>/dev/null || true
    fi
    
    # Copy auth utilities
    if [[ -d OLD_IMPLEMENTATION/pkg/auth ]]; then
        create_directory "packages/go-common/pkg/auth"
        cp -r OLD_IMPLEMENTATION/pkg/auth/* packages/go-common/pkg/auth/ 2>/dev/null || true
    fi
fi

track_component "$PHASE_ID" "go-common-packages" "completed"

# Component 4: Migrate contracts
log_info "Migrating contract definitions..."
track_component "$PHASE_ID" "contracts" "in_progress"

# Create contracts structure
create_directory "packages/contracts/openapi/specs"
create_directory "packages/contracts/proto/phoenix/v1"
create_directory "packages/contracts/schemas/json"
create_directory "packages/contracts/schemas/avro"

# Create package.json for contracts
if [[ ! -f packages/contracts/package.json ]]; then
    cat > packages/contracts/package.json << 'EOF'
{
  "name": "@phoenix/contracts",
  "version": "0.1.0",
  "description": "Phoenix Platform API contracts and schemas",
  "scripts": {
    "build": "npm run build:openapi && npm run build:proto",
    "build:openapi": "echo 'OpenAPI generation placeholder'",
    "build:proto": "echo 'Protobuf generation placeholder'",
    "validate": "npm run validate:openapi",
    "validate:openapi": "echo 'OpenAPI validation placeholder'",
    "clean": "rm -rf generated"
  },
  "devDependencies": {
    "@apidevtools/swagger-cli": "^4.0.4",
    "@grpc/proto-loader": "^0.7.10",
    "openapi-typescript": "^6.7.0"
  }
}
EOF
    log_success "Created contracts package.json"
fi

# Create README for contracts
cat > packages/contracts/README.md << 'EOF'
# Phoenix Platform Contracts

API contracts and schema definitions for Phoenix Platform.

## Structure

- `openapi/`: OpenAPI 3.0 specifications
- `proto/`: Protocol Buffer definitions
- `schemas/`: JSON Schema and Avro schemas

## Building

```bash
npm run build        # Build all contracts
npm run validate     # Validate specifications
```
EOF

# Migrate existing proto files
if [[ -d OLD_IMPLEMENTATION/phoenix-platform/api/proto ]]; then
    log_info "Migrating protocol buffer definitions..."
    cp -r OLD_IMPLEMENTATION/phoenix-platform/api/proto/* packages/contracts/proto/ 2>/dev/null || true
fi

# Migrate OpenAPI specs
if [[ -f OLD_IMPLEMENTATION/docs/assets/openapi.yaml ]]; then
    cp OLD_IMPLEMENTATION/docs/assets/openapi.yaml packages/contracts/openapi/specs/phoenix-api.yaml
    log_success "Migrated OpenAPI specification"
fi

track_component "$PHASE_ID" "contracts" "completed"

# Component 5: Setup package management
log_info "Setting up package management..."
track_component "$PHASE_ID" "package-setup" "in_progress"

# Create base TypeScript config if not exists
if [[ ! -f tsconfig.base.json ]]; then
    cat > tsconfig.base.json << 'EOF'
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020"],
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "moduleResolution": "node",
    "allowJs": false,
    "noEmit": false
  },
  "exclude": ["node_modules", "dist", "build", "coverage"]
}
EOF
    log_success "Created base TypeScript configuration"
fi

# Create lerna.json for monorepo management
if [[ ! -f lerna.json ]]; then
    cat > lerna.json << 'EOF'
{
  "version": "independent",
  "npmClient": "npm",
  "packages": [
    "packages/*",
    "services/*"
  ],
  "command": {
    "publish": {
      "conventionalCommits": true,
      "message": "chore(release): publish",
      "registry": "https://registry.npmjs.org",
      "allowBranch": ["main", "release/*"]
    },
    "version": {
      "conventionalCommits": true
    }
  }
}
EOF
    log_success "Created lerna configuration"
fi

track_component "$PHASE_ID" "package-setup" "completed"

# Run validations
log_info "Running phase validations..."

VALIDATIONS_PASSED=true

# Validation 1: Package directories
REQUIRED_PACKAGES=("config" "common" "go-common" "contracts")
for pkg in "${REQUIRED_PACKAGES[@]}"; do
    if [[ -d "packages/$pkg" ]]; then
        record_validation "$PHASE_ID" "${pkg}_package" "passed" "Package $pkg directory exists"
    else
        record_validation "$PHASE_ID" "${pkg}_package" "failed" "Missing package $pkg directory"
        VALIDATIONS_PASSED=false
    fi
done

# Validation 2: Package.json files
for pkg in "config" "common" "contracts"; do
    if [[ -f "packages/$pkg/package.json" ]]; then
        record_validation "$PHASE_ID" "${pkg}_package_json" "passed" "Package $pkg has package.json"
    else
        record_validation "$PHASE_ID" "${pkg}_package_json" "failed" "Missing package.json for $pkg"
        VALIDATIONS_PASSED=false
    fi
done

# Validation 3: Go module
if [[ -f "packages/go-common/go.mod" ]]; then
    record_validation "$PHASE_ID" "go_common_module" "passed" "Go common module initialized"
else
    record_validation "$PHASE_ID" "go_common_module" "failed" "Missing go.mod for go-common"
    VALIDATIONS_PASSED=false
fi

# Validation 4: TypeScript configuration
if [[ -f "tsconfig.base.json" ]]; then
    record_validation "$PHASE_ID" "typescript_config" "passed" "Base TypeScript configuration exists"
else
    record_validation "$PHASE_ID" "typescript_config" "failed" "Missing base TypeScript configuration"
    VALIDATIONS_PASSED=false
fi

# Release lock
release_lock "$PHASE_ID" "$AGENT_ID"

# Final status
if [[ "$VALIDATIONS_PASSED" == "true" ]]; then
    log_success "Phase 1: Shared packages migration completed successfully!"
    exit 0
else
    log_error "Phase 1: Shared packages migration completed with validation failures"
    exit 1
fi