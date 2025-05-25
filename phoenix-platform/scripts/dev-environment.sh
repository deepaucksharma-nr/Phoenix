#!/bin/bash

# Development environment management script for Phoenix Platform
# This script helps manage the local development environment using docker-compose

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.dev.yml"
ENV_FILE="$PROJECT_ROOT/.env"

# Function to display usage
usage() {
    echo -e "${BLUE}Phoenix Platform Development Environment${NC}"
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  up        Start all services"
    echo "  down      Stop all services"
    echo "  restart   Restart all services"
    echo "  logs      Show logs for all services"
    echo "  logs-f    Follow logs for all services"
    echo "  status    Show status of all services"
    echo "  build     Build all service images"
    echo "  clean     Stop services and remove volumes"
    echo "  migrate   Run database migrations"
    echo "  test      Run integration tests"
    echo "  help      Show this help message"
    echo ""
    echo "Service-specific commands:"
    echo "  logs <service>    Show logs for specific service"
    echo "  restart <service> Restart specific service"
    echo "  exec <service>    Execute shell in service container"
    echo ""
    echo "Available services:"
    echo "  - postgres"
    echo "  - redis"
    echo "  - prometheus"
    echo "  - grafana"
    echo "  - experiment-controller"
    echo "  - config-generator"
    echo "  - control-service"
    echo "  - api-gateway"
    echo "  - dashboard"
    echo "  - process-simulator"
    echo "  - nats"
}

# Function to check if .env file exists
check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}⚠ .env file not found. Creating from template...${NC}"
        if [ -f "$PROJECT_ROOT/.env.example" ]; then
            cp "$PROJECT_ROOT/.env.example" "$ENV_FILE"
            echo -e "${GREEN}✓ Created .env file from template${NC}"
            echo -e "${YELLOW}  Please update the values in .env as needed${NC}"
        else
            echo -e "${RED}✗ .env.example not found. Creating minimal .env file...${NC}"
            cat > "$ENV_FILE" << EOF
# Phoenix Platform Environment Variables
ENVIRONMENT=development
GIT_TOKEN=
EOF
            echo -e "${GREEN}✓ Created minimal .env file${NC}"
        fi
    fi
}

# Function to start services
start_services() {
    echo -e "${BLUE}Starting Phoenix Platform services...${NC}"
    check_env_file
    
    docker-compose -f "$COMPOSE_FILE" up -d
    
    echo -e "${GREEN}✓ Services started${NC}"
    echo ""
    echo -e "${BLUE}Service URLs:${NC}"
    echo "  API Gateway:    http://localhost:8080"
    echo "  Dashboard:      http://localhost:5173"
    echo "  Prometheus:     http://localhost:9090"
    echo "  Grafana:        http://localhost:3000 (admin/admin)"
    echo ""
    echo -e "${BLUE}gRPC Endpoints:${NC}"
    echo "  Experiment:     localhost:50051"
    echo "  Generator:      localhost:50052"
    echo "  Controller:     localhost:50053"
}

# Function to stop services
stop_services() {
    echo -e "${BLUE}Stopping Phoenix Platform services...${NC}"
    docker-compose -f "$COMPOSE_FILE" down
    echo -e "${GREEN}✓ Services stopped${NC}"
}

# Function to restart services
restart_services() {
    if [ -n "${1:-}" ]; then
        echo -e "${BLUE}Restarting service: $1...${NC}"
        docker-compose -f "$COMPOSE_FILE" restart "$1"
        echo -e "${GREEN}✓ Service $1 restarted${NC}"
    else
        echo -e "${BLUE}Restarting all services...${NC}"
        docker-compose -f "$COMPOSE_FILE" restart
        echo -e "${GREEN}✓ All services restarted${NC}"
    fi
}

# Function to show logs
show_logs() {
    if [ -n "${1:-}" ]; then
        docker-compose -f "$COMPOSE_FILE" logs "$1"
    else
        docker-compose -f "$COMPOSE_FILE" logs
    fi
}

# Function to follow logs
follow_logs() {
    if [ -n "${1:-}" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f "$1"
    else
        docker-compose -f "$COMPOSE_FILE" logs -f
    fi
}

# Function to show status
show_status() {
    echo -e "${BLUE}Phoenix Platform Service Status:${NC}"
    docker-compose -f "$COMPOSE_FILE" ps
}

# Function to build images
build_images() {
    echo -e "${BLUE}Building Phoenix Platform images...${NC}"
    docker-compose -f "$COMPOSE_FILE" build
    echo -e "${GREEN}✓ Images built${NC}"
}

# Function to clean environment
clean_environment() {
    echo -e "${YELLOW}⚠ This will stop all services and remove data volumes${NC}"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}Cleaning Phoenix Platform environment...${NC}"
        docker-compose -f "$COMPOSE_FILE" down -v
        echo -e "${GREEN}✓ Environment cleaned${NC}"
    else
        echo -e "${YELLOW}Cancelled${NC}"
    fi
}

# Function to run migrations
run_migrations() {
    echo -e "${BLUE}Running database migrations...${NC}"
    
    # Wait for postgres to be ready
    echo "Waiting for PostgreSQL to be ready..."
    docker-compose -f "$COMPOSE_FILE" exec postgres pg_isready -U phoenix
    
    # Run migration script
    cd "$PROJECT_ROOT"
    go run scripts/migrate.go up
    
    echo -e "${GREEN}✓ Migrations completed${NC}"
}

# Function to execute shell in container
exec_shell() {
    if [ -z "${1:-}" ]; then
        echo -e "${RED}✗ Service name required${NC}"
        echo "Usage: $0 exec <service>"
        exit 1
    fi
    
    echo -e "${BLUE}Executing shell in $1...${NC}"
    docker-compose -f "$COMPOSE_FILE" exec "$1" /bin/sh
}

# Function to run integration tests
run_tests() {
    echo -e "${BLUE}Running integration tests...${NC}"
    
    # Ensure services are running
    docker-compose -f "$COMPOSE_FILE" up -d
    
    # Wait for services to be healthy
    echo "Waiting for services to be ready..."
    sleep 10
    
    # Run tests
    cd "$PROJECT_ROOT"
    go test -v -tags=integration ./test/integration/...
    
    echo -e "${GREEN}✓ Tests completed${NC}"
}

# Main script logic
case "${1:-help}" in
    up)
        start_services
        ;;
    down)
        stop_services
        ;;
    restart)
        restart_services "${2:-}"
        ;;
    logs)
        show_logs "${2:-}"
        ;;
    logs-f)
        follow_logs "${2:-}"
        ;;
    status)
        show_status
        ;;
    build)
        build_images
        ;;
    clean)
        clean_environment
        ;;
    migrate)
        run_migrations
        ;;
    exec)
        exec_shell "${2:-}"
        ;;
    test)
        run_tests
        ;;
    help)
        usage
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        usage
        exit 1
        ;;
esac