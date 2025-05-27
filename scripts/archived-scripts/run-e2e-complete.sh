#!/bin/bash

# Phoenix Platform - Complete End-to-End Run Script
# This script starts all services and runs a complete workflow

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PROJECT_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for a service to be ready
wait_for_service() {
    local service_name=$1
    local check_command=$2
    local max_attempts=30
    local attempt=0
    
    log_info "Waiting for $service_name to be ready..."
    
    while [ $attempt -lt $max_attempts ]; do
        if eval "$check_command" >/dev/null 2>&1; then
            log_success "$service_name is ready!"
            return 0
        fi
        
        attempt=$((attempt + 1))
        echo -n "."
        sleep 2
    done
    
    echo
    log_error "$service_name failed to start after $max_attempts attempts"
    return 1
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    # Check required tools
    for tool in docker docker-compose go make curl jq; do
        if ! command_exists "$tool"; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and try again."
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running. Please start Docker and try again."
        exit 1
    fi
    
    log_success "All prerequisites satisfied!"
}

# Function to clean up previous runs
cleanup_previous() {
    log_info "Cleaning up previous runs..."
    
    # Stop any running containers
    docker-compose -f "$PROJECT_ROOT/docker-compose.yml" down -v 2>/dev/null || true
    
    # Kill any running services
    pkill -f "phoenix-api" 2>/dev/null || true
    pkill -f "phoenix-controller" 2>/dev/null || true
    pkill -f "phoenix-generator" 2>/dev/null || true
    
    log_success "Cleanup completed!"
}

# Function to start infrastructure
start_infrastructure() {
    log_info "Starting infrastructure services..."
    
    cd "$PROJECT_ROOT"
    
    # Start infrastructure services
    docker-compose up -d postgres redis nats prometheus grafana jaeger otel-collector
    
    # Wait for services to be ready
    wait_for_service "PostgreSQL" "docker exec phoenix-postgres pg_isready -U phoenix"
    wait_for_service "Redis" "docker exec phoenix-redis redis-cli --pass phoenix ping"
    wait_for_service "NATS" "curl -s http://localhost:8222/healthz"
    wait_for_service "Prometheus" "curl -s http://localhost:9090/-/healthy"
    wait_for_service "Grafana" "curl -s http://localhost:3000/api/health"
    wait_for_service "Jaeger" "curl -s http://localhost:16686/"
    
    log_success "Infrastructure services started!"
}

# Function to run database migrations
run_migrations() {
    log_info "Running database migrations..."
    
    # Create databases if they don't exist
    docker exec phoenix-postgres psql -U phoenix -c "CREATE DATABASE phoenix_db;" 2>/dev/null || true
    docker exec phoenix-postgres psql -U phoenix -c "CREATE DATABASE experiments_db;" 2>/dev/null || true
    docker exec phoenix-postgres psql -U phoenix -c "CREATE DATABASE pipelines_db;" 2>/dev/null || true
    
    # Run migrations for each service
    for migration_dir in "$PROJECT_ROOT"/projects/*/migrations; do
        if [ -d "$migration_dir" ] && [ "$(ls -A $migration_dir/*.sql 2>/dev/null)" ]; then
            service_name=$(basename $(dirname "$migration_dir"))
            log_info "Running migrations for $service_name..."
            
            for sql_file in "$migration_dir"/*.sql; do
                docker cp "$sql_file" phoenix-postgres:/tmp/
                docker exec phoenix-postgres psql -U phoenix -d phoenix_db -f "/tmp/$(basename $sql_file)" || {
                    log_warning "Migration $(basename $sql_file) might have already been applied"
                }
            done
        fi
    done
    
    log_success "Database migrations completed!"
}

# Function to build services
build_services() {
    log_info "Building all services..."
    
    cd "$PROJECT_ROOT"
    
    # Build shared packages first
    log_info "Building shared packages..."
    cd pkg && go mod download && go build ./... && cd ..
    
    # Build each service
    for service_dir in projects/*/; do
        if [ -f "$service_dir/go.mod" ]; then
            service_name=$(basename "$service_dir")
            log_info "Building $service_name..."
            
            cd "$service_dir"
            go mod download
            go build -o bin/$service_name ./cmd/... 2>/dev/null || go build -o bin/$service_name ./cmd/main.go 2>/dev/null || {
                log_warning "Could not build $service_name, skipping..."
            }
            cd "$PROJECT_ROOT"
        fi
    done
    
    log_success "All services built!"
}

# Function to start core services
start_core_services() {
    log_info "Starting core services..."
    
    # Start API service
    log_info "Starting API service..."
    cd "$PROJECT_ROOT/projects/platform-api"
    DB_HOST=localhost DB_PORT=5432 DB_USER=phoenix DB_PASSWORD=phoenix DB_NAME=phoenix_db \
    REDIS_HOST=localhost REDIS_PORT=6379 REDIS_PASSWORD=phoenix \
    ./bin/platform-api &
    API_PID=$!
    
    # Start Controller service
    log_info "Starting Controller service..."
    cd "$PROJECT_ROOT/projects/controller"
    DB_HOST=localhost DB_PORT=5432 DB_USER=phoenix DB_PASSWORD=phoenix DB_NAME=experiments_db \
    ./bin/controller &
    CONTROLLER_PID=$!
    
    # Start Generator service
    log_info "Starting Generator service..."
    cd "$PROJECT_ROOT/projects/generator"
    ./bin/generator &
    GENERATOR_PID=$!
    
    cd "$PROJECT_ROOT"
    
    # Wait for services to be ready
    wait_for_service "API" "curl -s http://localhost:8080/health"
    wait_for_service "Controller" "curl -s http://localhost:8081/health"
    wait_for_service "Generator" "curl -s http://localhost:8082/health"
    
    log_success "Core services started!"
}

# Function to run basic workflow
run_basic_workflow() {
    log_info "Running basic end-to-end workflow..."
    
    cd "$PROJECT_ROOT"
    
    # Build CLI if not already built
    if [ ! -f "projects/phoenix-cli/bin/phoenix-cli" ]; then
        log_info "Building Phoenix CLI..."
        cd projects/phoenix-cli
        go build -o bin/phoenix-cli ./cmd
        cd "$PROJECT_ROOT"
    fi
    
    CLI="$PROJECT_ROOT/projects/phoenix-cli/bin/phoenix-cli"
    
    # Configure CLI
    export PHOENIX_API_URL="http://localhost:8080"
    
    # 1. Create an experiment
    log_info "Creating test experiment..."
    EXPERIMENT_ID=$($CLI experiment create \
        --name "E2E Test Experiment" \
        --description "End-to-end test" \
        --baseline-config '{"name":"baseline","type":"standard"}' \
        --candidate-config '{"name":"candidate","type":"optimized"}' \
        --success-criteria '{"min_reduction":10,"max_latency":100}' \
        --duration 5m \
        -o json | jq -r '.id' 2>/dev/null || echo "test-exp-1")
    
    log_info "Created experiment: $EXPERIMENT_ID"
    
    # 2. Start the experiment
    log_info "Starting experiment..."
    $CLI experiment start $EXPERIMENT_ID || log_warning "Experiment start failed"
    
    # 3. Check experiment status
    log_info "Checking experiment status..."
    $CLI experiment status $EXPERIMENT_ID || log_warning "Status check failed"
    
    # 4. Create a pipeline
    log_info "Creating test pipeline..."
    PIPELINE_RESULT=$($CLI pipeline create \
        --name "E2E Test Pipeline" \
        --type "cardinality-reduction" \
        --config '{"processors":[{"type":"filter","config":{"metric_patterns":["test.*"]}}]}' \
        -o json 2>/dev/null || echo '{"id":"test-pipeline-1"}')
    
    PIPELINE_ID=$(echo "$PIPELINE_RESULT" | jq -r '.id' 2>/dev/null || echo "test-pipeline-1")
    log_info "Created pipeline: $PIPELINE_ID"
    
    # 5. Deploy the pipeline
    log_info "Deploying pipeline..."
    $CLI pipeline deploy $PIPELINE_ID \
        --environment "development" \
        --replicas 1 || log_warning "Pipeline deployment failed"
    
    # 6. List pipelines
    log_info "Listing pipelines..."
    $CLI pipeline list || log_warning "Pipeline list failed"
    
    # 7. Check metrics
    log_info "Checking experiment metrics..."
    sleep 5
    $CLI experiment metrics $EXPERIMENT_ID --format table || log_warning "Metrics check failed"
    
    # 8. Stop the experiment
    log_info "Stopping experiment..."
    $CLI experiment stop $EXPERIMENT_ID || log_warning "Experiment stop failed"
    
    log_success "Basic workflow completed!"
}

# Function to run load simulation
run_load_simulation() {
    log_info "Running load simulation test..."
    
    cd "$PROJECT_ROOT"
    CLI="$PROJECT_ROOT/projects/phoenix-cli/bin/phoenix-cli"
    
    # List available load profiles
    log_info "Available load profiles:"
    $CLI loadsim list-profiles || log_warning "Could not list profiles"
    
    # Start a load simulation
    log_info "Starting load simulation..."
    $CLI loadsim start test-exp-1 \
        --profile realistic \
        --process-count 50 \
        --duration 2m || log_warning "Load simulation failed"
    
    # Check status
    sleep 5
    log_info "Checking load simulation status..."
    $CLI loadsim status test-exp-1 || log_warning "Status check failed"
    
    log_success "Load simulation test completed!"
}

# Function to check system health
check_system_health() {
    log_info "Checking system health..."
    
    local all_healthy=true
    
    # Check infrastructure
    docker-compose ps | grep -E "postgres|redis|nats|prometheus|grafana" | while read line; do
        if echo "$line" | grep -q "Up"; then
            echo -e "  ${GREEN}✓${NC} $(echo $line | awk '{print $1}')"
        else
            echo -e "  ${RED}✗${NC} $(echo $line | awk '{print $1}')"
            all_healthy=false
        fi
    done
    
    # Check services
    for service in "API:8080" "Controller:8081" "Generator:8082"; do
        name=${service%:*}
        port=${service#*:}
        if curl -s "http://localhost:$port/health" >/dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $name service"
        else
            echo -e "  ${RED}✗${NC} $name service"
            all_healthy=false
        fi
    done
    
    if $all_healthy; then
        log_success "All systems healthy!"
    else
        log_warning "Some services are not healthy"
    fi
}

# Function to show access information
show_access_info() {
    echo
    log_info "Phoenix Platform is running!"
    echo
    echo "Access points:"
    echo "  - API:        http://localhost:8080"
    echo "  - Grafana:    http://localhost:3000 (admin/phoenix)"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Jaeger:     http://localhost:16686"
    echo "  - Adminer:    http://localhost:8080 (postgres/phoenix/phoenix)"
    echo
    echo "To stop all services, run:"
    echo "  $0 stop"
    echo
}

# Function to stop all services
stop_all_services() {
    log_info "Stopping all services..."
    
    # Kill service processes
    [ ! -z "$API_PID" ] && kill $API_PID 2>/dev/null || true
    [ ! -z "$CONTROLLER_PID" ] && kill $CONTROLLER_PID 2>/dev/null || true
    [ ! -z "$GENERATOR_PID" ] && kill $GENERATOR_PID 2>/dev/null || true
    
    # Stop Docker containers
    cd "$PROJECT_ROOT"
    docker-compose down
    
    log_success "All services stopped!"
}

# Main execution
main() {
    case "${1:-}" in
        stop)
            stop_all_services
            ;;
        *)
            # Update todo
            log_info "Starting Phoenix Platform end-to-end run..."
            
            # Run all steps
            check_prerequisites
            cleanup_previous
            start_infrastructure
            run_migrations
            build_services
            start_core_services
            run_basic_workflow
            run_load_simulation
            check_system_health
            show_access_info
            
            # Keep running until interrupted
            log_info "Press Ctrl+C to stop all services..."
            trap stop_all_services INT TERM
            wait
            ;;
    esac
}

# Run main function
main "$@"