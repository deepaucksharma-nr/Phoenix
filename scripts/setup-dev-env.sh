#!/bin/bash
# setup-dev-env.sh - Set up local development environment for Phoenix Platform

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform Development Environment Setup ===${NC}"
echo ""

# Create necessary directories
setup_directories() {
    echo -e "${YELLOW}Setting up directories...${NC}"
    
    # Data directories for local development
    mkdir -p data/{prometheus,grafana,postgres}
    mkdir -p logs/{services,tests}
    mkdir -p tmp/builds
    
    echo -e "${GREEN}✓ Directories created${NC}"
}

# Generate development environment file
generate_env_file() {
    echo -e "${YELLOW}Generating .env file...${NC}"
    
    if [[ -f ".env" ]]; then
        echo "  .env already exists, creating .env.new"
        ENV_FILE=".env.new"
    else
        ENV_FILE=".env"
    fi
    
    cat > "$ENV_FILE" << 'EOF'
# Phoenix Platform Development Environment

# Database
DATABASE_URL=postgres://phoenix:phoenix@localhost:5432/phoenix_dev?sslmode=disable
TEST_DATABASE_URL=postgres://phoenix:phoenix@localhost:5433/phoenix_test?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# Service Ports
API_PORT=8080
CONTROLLER_PORT=8081
GENERATOR_PORT=8082
DASHBOARD_PORT=3001

# gRPC Ports
GRPC_API_PORT=50051
GRPC_CONTROLLER_PORT=50052
GRPC_GENERATOR_PORT=50053

# Monitoring
PROMETHEUS_URL=http://localhost:9090
GRAFANA_URL=http://localhost:3000

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Authentication (Development Only!)
JWT_SECRET=development-secret-do-not-use-in-production
AUTH_ENABLED=false

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_SERVICE_NAME=phoenix-dev

# Feature Flags
ENABLE_PROFILING=true
ENABLE_METRICS=true
ENABLE_TRACING=false

# Development Mode
DEVELOPMENT_MODE=true
HOT_RELOAD=true
EOF
    
    echo -e "${GREEN}✓ Environment file created: $ENV_FILE${NC}"
}

# Set up Docker Compose for dependencies
setup_docker_compose() {
    echo -e "${YELLOW}Creating docker-compose.dev.yml...${NC}"
    
    cat > docker-compose.dev.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: phoenix-postgres
    environment:
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: phoenix
      POSTGRES_DB: phoenix_dev
    ports:
      - "5432:5432"
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U phoenix"]
      interval: 10s
      timeout: 5s
      retries: 5

  postgres-test:
    image: postgres:15-alpine
    container_name: phoenix-postgres-test
    environment:
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: phoenix
      POSTGRES_DB: phoenix_test
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: phoenix-redis
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - ./data/redis:/data

  prometheus:
    image: prom/prometheus:latest
    container_name: phoenix-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/prometheus.yaml:/etc/prometheus/prometheus.yml
      - ./monitoring/prometheus/rules:/etc/prometheus/rules
      - ./data/prometheus:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    container_name: phoenix-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources.yaml:/etc/grafana/provisioning/datasources/datasources.yaml
      - ./data/grafana:/var/lib/grafana

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: phoenix-jaeger
    ports:
      - "6831:6831/udp"
      - "16686:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true

networks:
  default:
    name: phoenix-dev
EOF
    
    echo -e "${GREEN}✓ docker-compose.dev.yml created${NC}"
}

# Create Makefile for development
create_dev_makefile() {
    echo -e "${YELLOW}Creating development Makefile...${NC}"
    
    cat > Makefile.dev << 'EOF'
# Phoenix Platform Development Makefile

.PHONY: help dev-up dev-down dev-logs dev-clean test test-unit test-integration lint fmt

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development Environment
dev-up: ## Start development dependencies
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running migrations..."
	@make migrate-up

dev-down: ## Stop development dependencies
	docker-compose -f docker-compose.dev.yml down

dev-logs: ## Show development logs
	docker-compose -f docker-compose.dev.yml logs -f

dev-clean: ## Clean development data
	docker-compose -f docker-compose.dev.yml down -v
	rm -rf data/

# Database
migrate-up: ## Run database migrations
	@for service in projects/*/; do \
		if [ -d "$$service/migrations" ]; then \
			echo "Running migrations for $$service..."; \
			DATABASE_URL=$${DATABASE_URL} go run $$service/cmd/migrate/main.go up || true; \
		fi \
	done

migrate-down: ## Rollback database migrations
	@for service in projects/*/; do \
		if [ -d "$$service/migrations" ]; then \
			echo "Rolling back migrations for $$service..."; \
			DATABASE_URL=$${DATABASE_URL} go run $$service/cmd/migrate/main.go down || true; \
		fi \
	done

# Services
run-api: ## Run API service
	cd projects/api && go run cmd/main.go

run-controller: ## Run Controller service
	cd projects/controller && go run cmd/main.go

run-generator: ## Run Generator service
	cd projects/generator && go run cmd/main.go

run-all: ## Run all services (requires goreman)
	goreman start

# Testing
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -short ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -tags=integration ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Code Quality
lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# Validation
validate: ## Validate monorepo structure
	./scripts/validate-boundaries.sh
	./scripts/validate-builds.sh

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all > API_DOCS.txt
EOF
    
    echo -e "${GREEN}✓ Makefile.dev created${NC}"
}

# Create Procfile for running all services
create_procfile() {
    echo -e "${YELLOW}Creating Procfile for goreman...${NC}"
    
    cat > Procfile << 'EOF'
# Phoenix Platform Services

api: cd projects/platform-api && go run cmd/api/main.go
controller: cd projects/controller && go run cmd/main.go
generator: cd projects/generator && go run cmd/main.go
EOF
    
    echo -e "${GREEN}✓ Procfile created${NC}"
}

# Set up Git hooks
setup_git_hooks() {
    echo -e "${YELLOW}Setting up Git hooks...${NC}"
    
    mkdir -p .git/hooks
    
    # Pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for Phoenix Platform

echo "Running pre-commit checks..."

# Validate boundaries
if ! ./scripts/validate-boundaries.sh > /dev/null 2>&1; then
    echo "❌ Boundary validation failed!"
    echo "Run: ./scripts/validate-boundaries.sh"
    exit 1
fi

# Run linters
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run --fast; then
        echo "❌ Linting failed!"
        exit 1
    fi
fi

# Format check
if ! go fmt ./... | grep -q .; then
    echo "✅ Code formatting OK"
else
    echo "❌ Code formatting issues found!"
    echo "Run: go fmt ./..."
    exit 1
fi

echo "✅ Pre-commit checks passed"
EOF
    
    chmod +x .git/hooks/pre-commit
    echo -e "${GREEN}✓ Git hooks installed${NC}"
}

# Install development tools
install_dev_tools() {
    echo -e "${YELLOW}Installing development tools...${NC}"
    
    # Check and install tools
    tools=(
        "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        "golang.org/x/tools/cmd/goimports@latest"
        "github.com/mattn/goreman@latest"
        "github.com/cosmtrek/air@latest"
    )
    
    for tool in "${tools[@]}"; do
        tool_name=$(echo $tool | rev | cut -d'/' -f1 | rev | cut -d'@' -f1)
        echo -n "Installing $tool_name... "
        if go install "$tool" 2>/dev/null; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${YELLOW}skipped${NC}"
        fi
    done
}

# Generate VS Code settings
setup_vscode() {
    echo -e "${YELLOW}Setting up VS Code configuration...${NC}"
    
    mkdir -p .vscode
    
    cat > .vscode/settings.json << 'EOF'
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.formatTool": "goimports",
    "go.formatFlags": ["-local", "github.com/phoenix/platform"],
    "go.testFlags": ["-v"],
    "go.testTimeout": "10s",
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.5)"
    },
    "files.exclude": {
        "**/.git": true,
        "**/.DS_Store": true,
        "**/node_modules": true,
        "data/": true,
        "tmp/": true
    },
    "editor.formatOnSave": true,
    "[go]": {
        "editor.codeActionsOnSave": {
            "source.organizeImports": true
        }
    }
}
EOF
    
    cat > .vscode/launch.json << 'EOF'
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch API",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/projects/platform-api/cmd/api",
            "env": {
                "LOG_LEVEL": "debug"
            },
            "args": []
        },
        {
            "name": "Launch Controller",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/projects/controller/cmd",
            "env": {
                "LOG_LEVEL": "debug"
            },
            "args": []
        }
    ]
}
EOF
    
    echo -e "${GREEN}✓ VS Code configuration created${NC}"
}

# Main setup flow
main() {
    echo "This will set up your local development environment."
    echo ""
    
    setup_directories
    generate_env_file
    setup_docker_compose
    create_dev_makefile
    create_procfile
    setup_git_hooks
    install_dev_tools
    setup_vscode
    
    echo ""
    echo -e "${GREEN}=== Development Environment Setup Complete ===${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Start dependencies: make -f Makefile.dev dev-up"
    echo "2. Run migrations: make -f Makefile.dev migrate-up"
    echo "3. Start services:"
    echo "   - All services: goreman start"
    echo "   - Individual: make -f Makefile.dev run-api"
    echo "4. Run tests: make -f Makefile.dev test"
    echo ""
    echo "VS Code users: Reload window to activate settings"
}

# Run main function
main "$@"