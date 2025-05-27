#!/bin/bash
# Phoenix Development Environment Setup
# One-time setup for local development

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Phoenix Development Environment Setup"
echo "======================================="
echo ""

# Configuration
PHOENIX_DEV_DIR="${PHOENIX_DEV_DIR:-$HOME/.phoenix-dev}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        echo "  Please install $1 and try again"
        exit 1
    else
        echo -e "${GREEN}✓ $1 is installed${NC}"
    fi
}

check_command docker
check_command docker-compose
check_command go
check_command node
check_command npm
check_command make

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MIN_GO_VERSION="1.21"
if [[ "$(printf '%s\n' "$MIN_GO_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_GO_VERSION" ]]; then
    echo -e "${RED}✗ Go version $GO_VERSION is too old (minimum: $MIN_GO_VERSION)${NC}"
    exit 1
fi

# Create development directories
echo -e "\n${YELLOW}Creating development directories...${NC}"
mkdir -p "$PHOENIX_DEV_DIR"/{data,logs,config,backups}
mkdir -p "$PROJECT_ROOT"/logs

# Create development docker-compose file
echo -e "\n${YELLOW}Creating docker-compose.dev.yml...${NC}"
cat > "$PROJECT_ROOT/docker-compose.dev.yml" << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: phoenix-postgres-dev
    environment:
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: phoenix-dev
      POSTGRES_DB: phoenix
    ports:
      - "5432:5432"
    volumes:
      - phoenix-postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U phoenix"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: phoenix-redis-dev
    ports:
      - "6379:6379"
    volumes:
      - phoenix-redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  prometheus:
    image: prom/prometheus:latest
    container_name: phoenix-prometheus-dev
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - phoenix-prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:9090/-/healthy"]
      interval: 10s
      timeout: 5s
      retries: 5

  grafana:
    image: grafana/grafana:latest
    container_name: phoenix-grafana-dev
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=phoenix-dev
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - phoenix-grafana-data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    depends_on:
      - prometheus

volumes:
  phoenix-postgres-data:
  phoenix-redis-data:
  phoenix-prometheus-data:
  phoenix-grafana-data:
EOF

# Create development environment file templates
echo -e "\n${YELLOW}Creating environment templates...${NC}"

# Phoenix API .env
cat > "$PROJECT_ROOT/projects/phoenix-api/.env.template" << 'EOF'
# Phoenix API Development Configuration
PORT=8080
DATABASE_URL=postgresql://phoenix:phoenix-dev@localhost:5432/phoenix?sslmode=disable
REDIS_URL=redis://localhost:6379
PROMETHEUS_URL=http://localhost:9090
JWT_SECRET=dev-secret-change-in-production
ENVIRONMENT=development
LOG_LEVEL=debug
SKIP_MIGRATIONS=false

# Feature flags
ENABLE_WEBSOCKET=true
ENABLE_METRICS_CACHE=true

# Timeouts
AGENT_POLL_TIMEOUT=30s
TASK_ASSIGN_TIMEOUT=5m
HEARTBEAT_INTERVAL=15s
EOF

# Copy templates if .env doesn't exist
if [ ! -f "$PROJECT_ROOT/projects/phoenix-api/.env" ]; then
    cp "$PROJECT_ROOT/projects/phoenix-api/.env.template" "$PROJECT_ROOT/projects/phoenix-api/.env"
    echo -e "${GREEN}✓ Created phoenix-api/.env${NC}"
fi

# Create development Makefile
echo -e "\n${YELLOW}Creating development Makefile...${NC}"
cat > "$PROJECT_ROOT/Makefile.dev" << 'EOF'
# Phoenix Development Makefile

.PHONY: help dev-up dev-down dev-logs dev-status dev-reset build-all test-all

help:
	@echo "Phoenix Development Commands:"
	@echo "  make dev-up      - Start all development services"
	@echo "  make dev-down    - Stop all development services"
	@echo "  make dev-logs    - Show logs from all services"
	@echo "  make dev-status  - Check status of all services"
	@echo "  make dev-reset   - Reset development environment"
	@echo "  make build-all   - Build all services"
	@echo "  make test-all    - Run all tests"

dev-up:
	@echo "Starting development services..."
	@docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@./scripts/dev-status.sh

dev-down:
	@echo "Stopping development services..."
	@docker-compose -f docker-compose.dev.yml down

dev-logs:
	@docker-compose -f docker-compose.dev.yml logs -f

dev-status:
	@./scripts/dev-status.sh

dev-reset:
	@./scripts/dev-reset.sh

build-all:
	@echo "Building all services..."
	@cd projects/phoenix-api && make build
	@cd projects/phoenix-agent && make build
	@cd projects/phoenix-cli && make build
	@echo "Build complete!"

test-all:
	@echo "Running all tests..."
	@cd projects/phoenix-api && make test
	@cd projects/phoenix-agent && make test
	@cd projects/phoenix-cli && make test
	@echo "Tests complete!"
EOF

# Create VS Code workspace settings
echo -e "\n${YELLOW}Creating VS Code workspace settings...${NC}"
mkdir -p "$PROJECT_ROOT/.vscode"
cat > "$PROJECT_ROOT/.vscode/settings.json" << 'EOF'
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "files.exclude": {
    "**/node_modules": true,
    "**/dist": true,
    "**/build": true,
    "**/.git": true
  },
  "search.exclude": {
    "**/node_modules": true,
    "**/dist": true,
    "**/vendor": true
  }
}
EOF

# Create Git hooks
echo -e "\n${YELLOW}Setting up Git hooks...${NC}"
mkdir -p "$PROJECT_ROOT/.git/hooks"
cat > "$PROJECT_ROOT/.git/hooks/pre-commit" << 'EOF'
#!/bin/bash
# Pre-commit hook for Phoenix

# Run validation checks
echo "Running pre-commit validation..."

# Check for cross-project imports
if ! ./scripts/phoenix-validate.sh --quick; then
    echo "Validation failed. Please fix the issues and try again."
    exit 1
fi

echo "Pre-commit validation passed!"
EOF
chmod +x "$PROJECT_ROOT/.git/hooks/pre-commit"

# Install Go dependencies
echo -e "\n${YELLOW}Installing Go dependencies...${NC}"
cd "$PROJECT_ROOT"
go work sync

# Install Node dependencies for dashboard
if [ -d "$PROJECT_ROOT/projects/dashboard" ]; then
    echo -e "\n${YELLOW}Installing dashboard dependencies...${NC}"
    cd "$PROJECT_ROOT/projects/dashboard"
    npm install
fi

# Create initial directories
echo -e "\n${YELLOW}Creating project directories...${NC}"
mkdir -p "$PROJECT_ROOT"/{bin,pkg,internal,configs,deployments,monitoring}

# Summary
echo -e "\n${GREEN}✅ Development environment setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Start infrastructure: make -f Makefile.dev dev-up"
echo "2. Build services: make -f Makefile.dev build-all"
echo "3. Start Phoenix: ./scripts/dev-start.sh"
echo ""
echo "Environment variables have been set in:"
echo "  - projects/phoenix-api/.env"
echo ""
echo "Development files created in: $PHOENIX_DEV_DIR"
echo "Logs will be stored in: $PROJECT_ROOT/logs"