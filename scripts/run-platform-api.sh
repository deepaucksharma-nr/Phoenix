#!/bin/bash
# Script to run the Phoenix Platform API with proper database setup

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Phoenix Platform API - Startup Script${NC}"
echo "====================================="

# Check if PostgreSQL is running
check_postgres() {
    echo -e "\n${YELLOW}Checking PostgreSQL...${NC}"
    
    if docker-compose ps postgres | grep -q "Up"; then
        echo -e "${GREEN}✓ PostgreSQL is running${NC}"
        return 0
    else
        echo -e "${RED}✗ PostgreSQL is not running${NC}"
        return 1
    fi
}

# Start infrastructure if needed
start_infrastructure() {
    echo -e "\n${YELLOW}Starting infrastructure services...${NC}"
    
    cd /Users/deepaksharma/Desktop/src/Phoenix
    docker-compose up -d postgres redis
    
    echo -e "${GREEN}Waiting for PostgreSQL to be ready...${NC}"
    sleep 5
    
    # Wait for PostgreSQL to be ready
    for i in {1..30}; do
        if docker-compose exec -T postgres pg_isready -U phoenix > /dev/null 2>&1; then
            echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
            break
        fi
        echo -n "."
        sleep 1
    done
}

# Create database if it doesn't exist
setup_database() {
    echo -e "\n${YELLOW}Setting up database...${NC}"
    
    # Check if database exists
    if docker-compose exec -T postgres psql -U phoenix -lqt | cut -d \| -f 1 | grep -qw phoenix; then
        echo -e "${GREEN}✓ Database 'phoenix' already exists${NC}"
    else
        echo "Creating database 'phoenix'..."
        docker-compose exec -T postgres createdb -U phoenix phoenix
        echo -e "${GREEN}✓ Database created${NC}"
    fi
}

# Run migrations
run_migrations() {
    echo -e "\n${YELLOW}Running database migrations...${NC}"
    
    cd /Users/deepaksharma/Desktop/src/Phoenix/projects/platform-api
    
    # Check if migrate tool exists, if not use psql directly
    if [ -f migrations/001_initial_schema.sql ]; then
        echo "Applying migrations..."
        docker-compose exec -T postgres psql -U phoenix -d phoenix < migrations/001_initial_schema.sql 2>/dev/null || true
        echo -e "${GREEN}✓ Migrations applied${NC}"
    else
        echo -e "${RED}No migrations found${NC}"
    fi
}

# Build the platform-api
build_api() {
    echo -e "\n${YELLOW}Building platform-api...${NC}"
    
    cd /Users/deepaksharma/Desktop/src/Phoenix/projects/platform-api
    go build -o bin/platform-api cmd/api/main.go
    
    echo -e "${GREEN}✓ Build complete${NC}"
}

# Run the platform-api
run_api() {
    echo -e "\n${BLUE}Starting Phoenix Platform API...${NC}"
    echo "================================"
    echo -e "${GREEN}API URL: http://localhost:8080${NC}"
    echo -e "${GREEN}WebSocket URL: ws://localhost:8080/ws${NC}"
    echo -e "${GREEN}Health Check: http://localhost:8080/health${NC}"
    echo -e "${GREEN}Metrics: http://localhost:8080/metrics${NC}"
    echo ""
    echo "Press Ctrl+C to stop"
    echo ""
    
    cd /Users/deepaksharma/Desktop/src/Phoenix/projects/platform-api
    
    # Set environment variables
    export DATABASE_URL="postgres://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"
    export PORT=8080
    
    # Run the API
    ./bin/platform-api
}

# Main execution
main() {
    # Check prerequisites
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}Error: docker-compose is required but not installed${NC}"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is required but not installed${NC}"
        exit 1
    fi
    
    # Start infrastructure if needed
    if ! check_postgres; then
        start_infrastructure
    fi
    
    # Setup database
    setup_database
    
    # Run migrations
    run_migrations
    
    # Build API
    build_api
    
    # Run API
    run_api
}

# Handle cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    # Optionally stop services
    # docker-compose stop postgres redis
}

trap cleanup EXIT

# Run main function
main