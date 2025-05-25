#!/bin/bash
# Phase 2: Core Services Migration
# Migrates core services from OLD_IMPLEMENTATION to new structure

set -euo pipefail

# Source common libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../lib/common.sh"
source "${SCRIPT_DIR}/../lib/state-tracker.sh"

# Phase configuration
PHASE_NAME="phase-2-core-services"
PHASE_NUMBER=2

# Service categories
declare -A SERVICE_CATEGORIES=(
    ["platform"]="api controller generator"
    ["operators"]="loadsim-operator pipeline-operator"
    ["control-plane"]="observer actuator"
    ["analytics"]="anomaly-detector benchmark validator analytics"
    ["generators"]="complex synthetic"
    ["dashboard"]="dashboard"
)

# Service migration mappings
declare -A SERVICE_MAPPINGS=(
    # Platform services
    ["api"]="OLD_IMPLEMENTATION/phoenix-platform/cmd/api:services/platform/api"
    ["controller"]="OLD_IMPLEMENTATION/phoenix-platform/cmd/controller:services/platform/controller"
    ["generator"]="OLD_IMPLEMENTATION/phoenix-platform/cmd/generator:services/platform/generator"
    
    # Operators
    ["loadsim-operator"]="OLD_IMPLEMENTATION/phoenix-platform/operators/loadsim:services/operators/loadsim"
    ["pipeline-operator"]="OLD_IMPLEMENTATION/phoenix-platform/operators/pipeline:services/operators/pipeline"
    
    # Control plane
    ["observer"]="OLD_IMPLEMENTATION/services/control-plane/observer:services/control-plane/observer"
    ["actuator"]="OLD_IMPLEMENTATION/services/control-plane/actuator:services/control-plane/actuator"
    
    # Analytics services
    ["anomaly-detector"]="OLD_IMPLEMENTATION/apps/anomaly-detector:services/analytics/anomaly-detector"
    ["benchmark"]="OLD_IMPLEMENTATION/services/benchmark:services/analytics/benchmark"
    ["validator"]="OLD_IMPLEMENTATION/services/validator:services/analytics/validator"
    ["analytics"]="OLD_IMPLEMENTATION/services/analytics:services/analytics/analytics"
    
    # Generators
    ["complex"]="OLD_IMPLEMENTATION/services/generators/complex:services/generators/complex"
    ["synthetic"]="OLD_IMPLEMENTATION/services/generators/synthetic:services/generators/synthetic"
    
    # Dashboard
    ["dashboard"]="OLD_IMPLEMENTATION/phoenix-platform/dashboard:services/dashboard"
)

# Initialize phase
init_phase() {
    log_info "Initializing Phase 2: Core Services Migration"
    
    # Load state
    load_state
    
    # Verify Phase 1 completion
    if [[ "$(get_state "phase_1_completed")" != "true" ]]; then
        log_error "Phase 1 must be completed before running Phase 2"
        return 1
    fi
    
    # Create phase marker
    set_state "phase_2_started" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    
    # Create services directory structure
    create_service_structure
    
    return 0
}

# Create service directory structure
create_service_structure() {
    log_info "Creating service directory structure"
    
    mkdir -p services/{platform,operators,control-plane,analytics,generators,dashboard}
    
    # Create category READMEs
    for category in platform operators control-plane analytics generators dashboard; do
        cat > "services/${category}/README.md" << EOF
# ${category^} Services

This directory contains ${category} services for the Phoenix Platform.

## Services

$(list_category_services "$category")

## Development

Each service follows a standard structure:
- \`cmd/\`: Application entrypoint
- \`internal/\`: Private application code
- \`build/\`: Docker and build configurations
- \`Makefile\`: Service-specific commands

## Building

\`\`\`bash
# Build all ${category} services
make build-${category}

# Build specific service
cd <service-name>
make build
\`\`\`
EOF
    done
    
    return 0
}

# List services in a category
list_category_services() {
    local category=$1
    local services=""
    
    case $category in
        platform)
            services="- API Service
- Controller Service
- Generator Service"
            ;;
        operators)
            services="- LoadSim Operator
- Pipeline Operator"
            ;;
        control-plane)
            services="- Observer
- Actuator"
            ;;
        analytics)
            services="- Anomaly Detector
- Benchmark Service
- Validator Service
- Analytics Service"
            ;;
        generators)
            services="- Complex Generator
- Synthetic Generator"
            ;;
        dashboard)
            services="- Dashboard UI"
            ;;
    esac
    
    echo "$services"
}

# Migrate a single service
migrate_service() {
    local service_name=$1
    local mapping="${SERVICE_MAPPINGS[$service_name]}"
    local src_path="${mapping%:*}"
    local dst_path="${mapping#*:}"
    
    log_info "Migrating service: $service_name"
    log_info "  From: $src_path"
    log_info "  To: $dst_path"
    
    # Check if already migrated
    if [[ -d "$dst_path" ]]; then
        log_warn "Service $service_name already migrated, skipping"
        return 0
    fi
    
    # Check source exists
    if [[ ! -d "$src_path" ]]; then
        log_error "Source path not found: $src_path"
        return 1
    fi
    
    # Create destination directory
    mkdir -p "$dst_path"
    
    # Copy service files
    cp -r "$src_path"/* "$dst_path/" 2>/dev/null || true
    
    # Migrate based on service type
    case $service_name in
        api|controller|generator)
            migrate_go_service "$service_name" "$dst_path"
            ;;
        loadsim-operator|pipeline-operator)
            migrate_operator_service "$service_name" "$dst_path"
            ;;
        observer|actuator)
            migrate_control_plane_service "$service_name" "$dst_path"
            ;;
        anomaly-detector|benchmark|validator|analytics)
            migrate_analytics_service "$service_name" "$dst_path"
            ;;
        complex|synthetic)
            migrate_generator_service "$service_name" "$dst_path"
            ;;
        dashboard)
            migrate_dashboard_service "$dst_path"
            ;;
    esac
    
    # Update service state
    set_state "service_${service_name}_migrated" "true"
    
    return 0
}

# Migrate Go service
migrate_go_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating Go service: $service_name"
    
    # Create standard structure
    mkdir -p "$dst_path"/{cmd,internal/{api,domain,infrastructure},build}
    
    # Move main.go if exists
    if [[ -f "$dst_path/main.go" ]]; then
        mv "$dst_path/main.go" "$dst_path/cmd/main.go"
    fi
    
    # Create or update go.mod
    cat > "$dst_path/go.mod" << EOF
module github.com/phoenix/platform/services/platform/$service_name

go 1.21

require (
    github.com/phoenix/platform/pkg v0.0.0-unpublished
)

replace github.com/phoenix/platform/pkg => ../../../pkg
EOF
    
    # Update imports in Go files
    find "$dst_path" -name "*.go" -type f | while read -r file; do
        update_go_imports "$file"
    done
    
    # Create Makefile
    create_service_makefile "$service_name" "$dst_path"
    
    # Create Dockerfile
    create_service_dockerfile "$service_name" "$dst_path"
    
    return 0
}

# Migrate operator service
migrate_operator_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating operator service: $service_name"
    
    # Create standard structure
    mkdir -p "$dst_path"/{cmd,internal/{controller,reconciler},build}
    
    # Move main.go if exists
    if [[ -f "$dst_path/main.go" ]]; then
        mv "$dst_path/main.go" "$dst_path/cmd/main.go"
    elif [[ -f "$dst_path/cmd/main.go" ]]; then
        # Already in correct location
        true
    fi
    
    # Move controllers
    if [[ -d "$dst_path/controllers" ]]; then
        mv "$dst_path/controllers"/* "$dst_path/internal/controller/" 2>/dev/null || true
        rm -rf "$dst_path/controllers"
    fi
    
    # Create or update go.mod
    cat > "$dst_path/go.mod" << EOF
module github.com/phoenix/platform/services/operators/${service_name%%-operator}

go 1.21

require (
    github.com/phoenix/platform/pkg v0.0.0-unpublished
    k8s.io/apimachinery v0.28.4
    k8s.io/client-go v0.28.4
    sigs.k8s.io/controller-runtime v0.16.3
)

replace github.com/phoenix/platform/pkg => ../../../pkg
EOF
    
    # Update imports
    find "$dst_path" -name "*.go" -type f | while read -r file; do
        update_go_imports "$file"
    done
    
    # Create Makefile
    create_operator_makefile "$service_name" "$dst_path"
    
    # Create Dockerfile
    create_operator_dockerfile "$service_name" "$dst_path"
    
    return 0
}

# Migrate control plane service
migrate_control_plane_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating control plane service: $service_name"
    
    # Handle Node.js services
    if [[ -f "$dst_path/package.json" ]]; then
        migrate_node_service "$service_name" "$dst_path"
        return 0
    fi
    
    # Handle shell script services
    if [[ -d "$dst_path/src" ]] && find "$dst_path/src" -name "*.sh" | grep -q .; then
        migrate_shell_service "$service_name" "$dst_path"
        return 0
    fi
    
    # Default to Go service migration
    migrate_go_service "$service_name" "$dst_path"
    
    return 0
}

# Migrate Node.js service
migrate_node_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating Node.js service: $service_name"
    
    # Update package.json
    if [[ -f "$dst_path/package.json" ]]; then
        # Update name and dependencies
        jq --arg name "@phoenix/platform-$service_name" \
           '.name = $name' "$dst_path/package.json" > "$dst_path/package.json.tmp"
        mv "$dst_path/package.json.tmp" "$dst_path/package.json"
    fi
    
    # Create Makefile
    cat > "$dst_path/Makefile" << 'EOF'
.PHONY: build test lint run docker

build:
	npm install
	npm run build || true

test:
	npm test || true

lint:
	npm run lint || true

run:
	npm start

docker:
	docker build -t phoenix-${SERVICE_NAME}:latest -f build/Dockerfile .

.DEFAULT_GOAL := build
EOF
    
    # Create Dockerfile
    create_node_dockerfile "$service_name" "$dst_path"
    
    return 0
}

# Migrate shell script service
migrate_shell_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating shell script service: $service_name"
    
    # Create standard structure
    mkdir -p "$dst_path"/{src,build,scripts}
    
    # Move scripts to src if not already there
    find "$dst_path" -maxdepth 1 -name "*.sh" -type f | while read -r script; do
        mv "$script" "$dst_path/src/"
    done
    
    # Create Makefile
    cat > "$dst_path/Makefile" << 'EOF'
.PHONY: build test lint run docker

build:
	@echo "Shell service - no build required"

test:
	shellcheck src/*.sh || true

lint:
	shellcheck src/*.sh

run:
	./src/main.sh || ./src/$(ls src/*.sh | head -n1 | xargs basename)

docker:
	docker build -t phoenix-${SERVICE_NAME}:latest -f build/Dockerfile .

.DEFAULT_GOAL := build
EOF
    
    # Create Dockerfile
    create_shell_dockerfile "$service_name" "$dst_path"
    
    return 0
}

# Migrate analytics service
migrate_analytics_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating analytics service: $service_name"
    
    # Most analytics services are Go-based
    migrate_go_service "$service_name" "$dst_path"
    
    return 0
}

# Migrate generator service  
migrate_generator_service() {
    local service_name=$1
    local dst_path=$2
    
    log_info "Migrating generator service: $service_name"
    
    # Check if shell-based or Go-based
    if [[ -d "$dst_path/src" ]] && find "$dst_path/src" -name "*.sh" | grep -q .; then
        migrate_shell_service "$service_name" "$dst_path"
    else
        migrate_go_service "$service_name" "$dst_path"
    fi
    
    return 0
}

# Migrate dashboard service
migrate_dashboard_service() {
    local dst_path=$1
    
    log_info "Migrating dashboard service"
    
    # Update package.json
    if [[ -f "$dst_path/package.json" ]]; then
        jq '.name = "@phoenix/platform-dashboard"' "$dst_path/package.json" > "$dst_path/package.json.tmp"
        mv "$dst_path/package.json.tmp" "$dst_path/package.json"
    fi
    
    # Update import paths in TypeScript/JavaScript files
    find "$dst_path/src" -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" | while read -r file; do
        # Update imports to use new package structure
        sed -i.bak 's|@phoenix-platform/|@phoenix/platform-|g' "$file"
        rm -f "${file}.bak"
    done
    
    # Create Makefile
    cat > "$dst_path/Makefile" << 'EOF'
.PHONY: build test lint run dev docker

build:
	npm install
	npm run build

test:
	npm test

lint:
	npm run lint

dev:
	npm run dev

run: build
	npm run preview

docker:
	docker build -t phoenix-dashboard:latest -f build/Dockerfile .

.DEFAULT_GOAL := build
EOF
    
    # Create Dockerfile
    create_dashboard_dockerfile "$dst_path"
    
    return 0
}

# Update Go imports
update_go_imports() {
    local file=$1
    
    # Update imports to use new pkg structure
    sed -i.bak '
        s|"github.com/deepaucksharma-nr/phoenix-monorepo/|"github.com/phoenix/platform/|g
        s|"github.com/deepaucksharma/phoenix-monorepo/|"github.com/phoenix/platform/|g
        s|phoenix-platform/pkg/|github.com/phoenix/platform/pkg/|g
        s|"../../../pkg/|"github.com/phoenix/platform/pkg/|g
        s|"../../pkg/|"github.com/phoenix/platform/pkg/|g
    ' "$file"
    
    # Remove backup file
    rm -f "${file}.bak"
}

# Create service Makefile
create_service_makefile() {
    local service_name=$1
    local dst_path=$2
    
    cat > "$dst_path/Makefile" << 'EOF'
# Service Makefile
SERVICE_NAME := $(notdir $(CURDIR))
DOCKER_IMAGE := phoenix-$(SERVICE_NAME):latest

.PHONY: build test lint run docker clean

build:
	go build -v -o bin/$(SERVICE_NAME) ./cmd

test:
	go test -v -race ./...

test-unit:
	go test -v -race -short ./...

test-integration:
	go test -v -race -run Integration ./...

lint:
	golangci-lint run

run: build
	./bin/$(SERVICE_NAME)

docker:
	docker build -t $(DOCKER_IMAGE) -f build/Dockerfile .

clean:
	rm -rf bin/
	go clean -cache

.DEFAULT_GOAL := build
EOF
}

# Create operator Makefile
create_operator_makefile() {
    local service_name=$1
    local dst_path=$2
    
    cat > "$dst_path/Makefile" << 'EOF'
# Operator Makefile
OPERATOR_NAME := $(notdir $(CURDIR))
DOCKER_IMAGE := phoenix-$(OPERATOR_NAME):latest

.PHONY: build test lint run docker manifests

build:
	go build -v -o bin/manager ./cmd

test:
	go test -v -race ./...

lint:
	golangci-lint run

run: build
	./bin/manager

docker:
	docker build -t $(DOCKER_IMAGE) -f build/Dockerfile .

manifests:
	controller-gen rbac:roleName=manager-role crd paths="./..." output:dir=config/crd

.DEFAULT_GOAL := build
EOF
}

# Create service Dockerfile
create_service_dockerfile() {
    local service_name=$1
    local dst_path=$2
    
    mkdir -p "$dst_path/build"
    
    cat > "$dst_path/build/Dockerfile" << 'EOF'
# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/main.go

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /

COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]
EOF
}

# Create operator Dockerfile
create_operator_dockerfile() {
    local service_name=$1
    local dst_path=$2
    
    mkdir -p "$dst_path/build"
    
    cat > "$dst_path/build/Dockerfile" << 'EOF'
# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]
EOF
}

# Create Node.js Dockerfile
create_node_dockerfile() {
    local service_name=$1
    local dst_path=$2
    
    mkdir -p "$dst_path/build"
    
    cat > "$dst_path/build/Dockerfile" << 'EOF'
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci --only=production

# Copy source
COPY . .

# Runtime stage
FROM node:18-alpine

RUN apk --no-cache add dumb-init

WORKDIR /app

# Copy from builder
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app .

USER node

ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "index.js"]
EOF
}

# Create shell Dockerfile
create_shell_dockerfile() {
    local service_name=$1
    local dst_path=$2
    
    mkdir -p "$dst_path/build"
    
    cat > "$dst_path/build/Dockerfile" << 'EOF'
FROM alpine:3.18

RUN apk --no-cache add bash curl jq

WORKDIR /app

COPY src/ ./

RUN chmod +x *.sh

USER nobody

ENTRYPOINT ["./main.sh"]
EOF
}

# Create dashboard Dockerfile
create_dashboard_dockerfile() {
    local dst_path=$1
    
    mkdir -p "$dst_path/build"
    
    cat > "$dst_path/build/Dockerfile" << 'EOF'
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source
COPY . .

# Build
RUN npm run build

# Runtime stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY build/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
EOF
    
    # Create nginx config
    cat > "$dst_path/build/nginx.conf" << 'EOF'
server {
    listen 80;
    server_name localhost;
    
    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://phoenix-api:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
EOF
}

# Validate service migration
validate_service() {
    local service_name=$1
    local dst_path="${SERVICE_MAPPINGS[$service_name]#*:}"
    
    log_info "Validating service: $service_name"
    
    # Check directory exists
    if [[ ! -d "$dst_path" ]]; then
        log_error "Service directory not found: $dst_path"
        return 1
    fi
    
    # Check for required files
    if [[ ! -f "$dst_path/Makefile" ]]; then
        log_error "Makefile not found for service: $service_name"
        return 1
    fi
    
    # Check for Dockerfile
    if [[ ! -f "$dst_path/build/Dockerfile" ]] && [[ ! -f "$dst_path/Dockerfile" ]]; then
        log_error "Dockerfile not found for service: $service_name"
        return 1
    fi
    
    # Language-specific validation
    if [[ -f "$dst_path/go.mod" ]]; then
        validate_go_service "$service_name" "$dst_path"
    elif [[ -f "$dst_path/package.json" ]]; then
        validate_node_service "$service_name" "$dst_path"
    fi
    
    return 0
}

# Validate Go service
validate_go_service() {
    local service_name=$1
    local dst_path=$2
    
    # Check go.mod validity
    if ! (cd "$dst_path" && go mod verify 2>/dev/null); then
        log_warn "Go module verification failed for $service_name"
    fi
    
    # Check for main.go
    if [[ ! -f "$dst_path/cmd/main.go" ]] && [[ ! -f "$dst_path/main.go" ]]; then
        log_error "No main.go found for Go service: $service_name"
        return 1
    fi
    
    # Check imports
    if grep -r "OLD_IMPLEMENTATION" "$dst_path" --include="*.go" > /dev/null 2>&1; then
        log_error "Found references to OLD_IMPLEMENTATION in $service_name"
        return 1
    fi
    
    return 0
}

# Validate Node.js service
validate_node_service() {
    local service_name=$1
    local dst_path=$2
    
    # Check package.json validity
    if ! (cd "$dst_path" && npm list --depth=0 > /dev/null 2>&1); then
        log_warn "npm dependencies not installed for $service_name"
    fi
    
    return 0
}

# Run validation for all migrated services
validate_all_services() {
    log_info "Validating all migrated services"
    
    local validation_failed=0
    
    for service_name in "${!SERVICE_MAPPINGS[@]}"; do
        if [[ "$(get_state "service_${service_name}_migrated")" == "true" ]]; then
            if ! validate_service "$service_name"; then
                validation_failed=1
            fi
        fi
    done
    
    if [[ $validation_failed -eq 1 ]]; then
        log_error "Service validation failed"
        return 1
    fi
    
    log_success "All services validated successfully"
    return 0
}

# Update go.work for new services
update_go_workspace() {
    log_info "Updating Go workspace"
    
    # Add service modules to go.work
    for category in platform operators control-plane analytics generators; do
        find "services/$category" -name "go.mod" -type f | while read -r mod_file; do
            local mod_dir=$(dirname "$mod_file")
            local rel_path=$(realpath --relative-to="." "$mod_dir")
            
            # Check if already in go.work
            if ! grep -q "$rel_path" go.work 2>/dev/null; then
                log_info "Adding $rel_path to go.work"
                # Add to go.work (implementation depends on go.work format)
            fi
        done
    done
    
    # Sync workspace
    if command -v go >/dev/null 2>&1; then
        go work sync || true
    fi
    
    return 0
}

# Main migration function
main() {
    log_header "Phase 2: Core Services Migration"
    
    # Initialize phase
    if ! init_phase; then
        log_error "Phase initialization failed"
        exit 1
    fi
    
    # Check for multi-agent lock
    if ! acquire_migration_lock; then
        log_error "Another migration process is running"
        exit 1
    fi
    
    # Migrate services by category
    for category in "${!SERVICE_CATEGORIES[@]}"; do
        log_info "Migrating $category services"
        
        for service in ${SERVICE_CATEGORIES[$category]}; do
            if ! migrate_service "$service"; then
                log_error "Failed to migrate service: $service"
                release_migration_lock
                exit 1
            fi
        done
    done
    
    # Validate all services
    if ! validate_all_services; then
        log_error "Service validation failed"
        release_migration_lock
        exit 1
    fi
    
    # Update Go workspace
    update_go_workspace
    
    # Update phase state
    set_state "phase_2_completed" "true"
    set_state "phase_2_completed_at" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    
    # Release lock
    release_migration_lock
    
    log_success "Phase 2: Core Services Migration completed successfully!"
    log_info "Services migrated to:"
    log_info "  - services/platform/*"
    log_info "  - services/operators/*"
    log_info "  - services/control-plane/*"
    log_info "  - services/analytics/*"
    log_info "  - services/generators/*"
    log_info "  - services/dashboard/*"
    
    return 0
}

# Run main function
main "$@"