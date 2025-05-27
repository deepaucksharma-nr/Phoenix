#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS=linux;;
        Darwin*)    OS=darwin;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*) OS=windows;;
        *)          OS="unknown";;
    esac
    echo "$OS"
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64)     ARCH=amd64;;
        arm64|aarch64) ARCH=arm64;;
        *)          ARCH="unknown";;
    esac
    echo "$ARCH"
}

OS=$(detect_os)
ARCH=$(detect_arch)

print_info "Phoenix Platform Development Environment Setup"
print_info "OS: $OS, Architecture: $ARCH"
echo

# Check prerequisites
print_info "Checking prerequisites..."

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION found"
else
    print_error "Go not found. Please install Go 1.21 or later"
    echo "Visit: https://golang.org/dl/"
    exit 1
fi

# Check Node.js
if command_exists node; then
    NODE_VERSION=$(node --version)
    print_success "Node.js $NODE_VERSION found"
else
    print_error "Node.js not found. Please install Node.js 18 or later"
    echo "Visit: https://nodejs.org/"
    exit 1
fi

# Check Docker
if command_exists docker; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    print_success "Docker $DOCKER_VERSION found"
else
    print_error "Docker not found. Please install Docker"
    echo "Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Docker Compose
if command_exists docker-compose || docker compose version >/dev/null 2>&1; then
    print_success "Docker Compose found"
else
    print_error "Docker Compose not found. Please install Docker Compose"
    echo "Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

# Check Make
if command_exists make; then
    print_success "Make found"
else
    print_error "Make not found. Please install Make"
    exit 1
fi

# Install pnpm if not present
if ! command_exists pnpm; then
    print_info "Installing pnpm..."
    npm install -g pnpm
    print_success "pnpm installed"
else
    PNPM_VERSION=$(pnpm --version)
    print_success "pnpm $PNPM_VERSION found"
fi

# Install Go tools
print_info "Installing Go development tools..."

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
go install mvdan.cc/gofumpt@latest
go install github.com/golang/mock/mockgen@v1.6.0
go install github.com/google/wire/cmd/wire@latest
go install github.com/swaggo/swag/cmd/swag@v1.16.2
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2
go install github.com/bufbuild/buf/cmd/buf@v1.28.1
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

print_success "Go tools installed"

# Install global Node tools
print_info "Installing Node.js development tools..."

pnpm add -g typescript@latest
pnpm add -g @types/node@latest
pnpm add -g eslint@latest
pnpm add -g prettier@latest
pnpm add -g npm-check-updates@latest

print_success "Node.js tools installed"

# Install optional tools
print_info "Checking optional tools..."

# kubectl
if ! command_exists kubectl; then
    print_warning "kubectl not found. Install it for Kubernetes development"
    echo "Visit: https://kubernetes.io/docs/tasks/tools/"
fi

# helm
if ! command_exists helm; then
    print_warning "Helm not found. Install it for Kubernetes package management"
    echo "Visit: https://helm.sh/docs/intro/install/"
fi

# kind
if ! command_exists kind; then
    print_warning "kind not found. Install it for local Kubernetes testing"
    echo "Visit: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
fi

# Setup Go workspace
print_info "Setting up Go workspace..."
go work sync
print_success "Go workspace synced"

# Create local configuration files
print_info "Creating local configuration files..."

# Create .env file if not exists
if [ ! -f .env ]; then
    cat > .env << EOF
# Phoenix Platform Local Development Environment

# Database
DATABASE_URL=postgres://phoenix:phoenix@localhost:5432/phoenix_db?sslmode=disable
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5

# Redis
REDIS_URL=redis://:phoenix@localhost:6379/0
REDIS_MAX_CONNECTIONS=10

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=phoenix-dev

# NATS
NATS_URL=nats://localhost:4222

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Observability
PROMETHEUS_URL=http://localhost:9090
GRAFANA_URL=http://localhost:3000
JAEGER_URL=http://localhost:16686
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

# API Configuration
API_HOST=0.0.0.0
API_PORT=8080
API_READ_TIMEOUT=30s
API_WRITE_TIMEOUT=30s

# Authentication
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRATION=24h

# Feature Flags
FEATURE_NEW_UI=true
FEATURE_ML_OPTIMIZATION=false

# Development
DEBUG=true
LOG_LEVEL=debug
ENVIRONMENT=development
EOF
    print_success "Created .env file"
else
    print_info ".env file already exists"
fi

# Create directories for local data
print_info "Creating local data directories..."
mkdir -p data/{postgres,redis,kafka,prometheus,grafana}

# Initialize git hooks
print_info "Setting up git hooks..."
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for Phoenix Platform

echo "Running pre-commit checks..."

# Run Go formatting check
echo "Checking Go formatting..."
if ! make fmt-check 2>/dev/null; then
    echo "Go formatting issues found. Run 'make fmt' to fix."
    exit 1
fi

# Run linting
echo "Running linters..."
if ! make lint 2>/dev/null; then
    echo "Linting issues found. Please fix them before committing."
    exit 1
fi

echo "Pre-commit checks passed!"
EOF

chmod +x .git/hooks/pre-commit
print_success "Git hooks configured"

# Start development services
print_info "Starting development services..."
docker-compose up -d

# Wait for services to be ready
print_info "Waiting for services to be ready..."
sleep 10

# Check service health
print_info "Checking service health..."

# PostgreSQL
if docker-compose exec -T postgres pg_isready -U phoenix >/dev/null 2>&1; then
    print_success "PostgreSQL is ready"
else
    print_error "PostgreSQL is not ready"
fi

# Redis
if docker-compose exec -T redis redis-cli --pass phoenix ping >/dev/null 2>&1; then
    print_success "Redis is ready"
else
    print_error "Redis is not ready"
fi

# Prometheus
if curl -s http://localhost:9090/-/healthy >/dev/null 2>&1; then
    print_success "Prometheus is ready"
else
    print_error "Prometheus is not ready"
fi

# Grafana
if curl -s http://localhost:3000/api/health >/dev/null 2>&1; then
    print_success "Grafana is ready"
else
    print_error "Grafana is not ready"
fi

echo
print_success "Development environment setup complete!"
echo
print_info "Available services:"
echo "  - PostgreSQL: localhost:5432 (user: phoenix, pass: phoenix)"
echo "  - Redis: localhost:6379 (pass: phoenix)"
echo "  - Kafka: localhost:9092"
echo "  - NATS: localhost:4222"
echo "  - MinIO: http://localhost:9001 (user: minioadmin, pass: minioadmin)"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (user: admin, pass: phoenix)"
echo "  - Jaeger: http://localhost:16686"
echo
print_info "To view all available make targets:"
echo "  make help"
echo
print_info "To stop services:"
echo "  make dev-down"
echo
print_info "To view logs:"
echo "  make dev-logs"
echo
print_success "Happy coding! 🚀"