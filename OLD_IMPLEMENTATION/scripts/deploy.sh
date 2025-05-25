#!/usr/bin/env bash
# Unified Phoenix-vNext Deployment Script
# Supports local Docker, AWS EKS, and Azure AKS deployments

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Script configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default values
DEPLOYMENT_TARGET="local"
ENVIRONMENT="development"
NAMESPACE="phoenix-system"
DRY_RUN=false
VERBOSE=false
SKIP_BUILD=false
FORCE=false

# Function to show usage
show_usage() {
    cat << EOF
Phoenix-vNext Unified Deployment Script

Usage: $0 [OPTIONS] TARGET

Targets:
  local     Deploy to local Docker environment
  aws       Deploy to AWS with Docker Swarm
  azure     Deploy to Azure with Container Instances

Options:
  -e, --environment ENV     Deployment environment (default: development)
  -n, --network NS          Docker network name (default: phoenix-network)
  -d, --dry-run            Show what would be deployed without executing
  -v, --verbose            Enable verbose output
  -s, --skip-build         Skip building Docker images
  -f, --force              Force deployment without confirmation
  -h, --help               Show this help message

Examples:
  $0 local                          # Deploy locally with Docker Compose
  $0 aws -e production             # Deploy to AWS with Docker Swarm
  $0 azure --dry-run               # Show what would be deployed to Azure

Environment Variables:
  AWS_REGION                       # AWS region for deployment
  AZURE_REGION                     # Azure region for deployment
  PHOENIX_IMAGE_TAG                # Image tag to deploy (default: latest)
  NEW_RELIC_LICENSE_KEY           # New Relic license key (optional)

EOF
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -s|--skip-build)
                SKIP_BUILD=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            local|aws|azure)
                DEPLOYMENT_TARGET="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for $DEPLOYMENT_TARGET deployment..."
    
    case $DEPLOYMENT_TARGET in
        local)
            command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed"; exit 1; }
            command -v docker-compose >/dev/null 2>&1 || { log_error "Docker Compose is required but not installed"; exit 1; }
            ;;
        aws)
            command -v aws >/dev/null 2>&1 || { log_error "AWS CLI is required but not installed"; exit 1; }
            command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed"; exit 1; }
            ;;
        azure)
            command -v az >/dev/null 2>&1 || { log_error "Azure CLI is required but not installed"; exit 1; }
            command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed"; exit 1; }
            ;;
    esac
    
    log_success "Prerequisites check passed"
}

# Function to build images
build_images() {
    if [ "$SKIP_BUILD" = true ]; then
        log_info "Skipping image build as requested"
        return
    fi
    
    log_info "Building Phoenix images..."
    
    cd "$PROJECT_ROOT"
    
    if [ "$DEPLOYMENT_TARGET" = "local" ]; then
        docker-compose build
    else
        # Build for cloud deployment
        local tag="${PHOENIX_IMAGE_TAG:-latest}"
        
        # Build each service
        for service in control-actuator-go anomaly-detector benchmark synthetic-generator; do
            if [ -d "apps/$service" ] || [ -d "services/$service" ]; then
                local context_dir="apps/$service"
                [ -d "services/$service" ] && context_dir="services/$service"
                
                log_info "Building $service:$tag"
                docker build -t "phoenix/$service:$tag" "$context_dir"
            fi
        done
    fi
    
    log_success "Image build completed"
}

# Function to deploy locally
deploy_local() {
    log_info "Deploying Phoenix to local Docker environment..."
    
    cd "$PROJECT_ROOT"
    
    # Initialize environment if needed
    if [ ! -f ".env" ]; then
        log_info "Creating .env file from template"
        cp .env.template .env
    fi
    
    # Start services
    if [ "$DRY_RUN" = true ]; then
        log_info "Dry run - would execute: docker-compose up -d"
        docker-compose config
    else
        docker-compose up -d
        
        # Wait for services to be ready
        log_info "Waiting for services to start..."
        sleep 10
        
        # Check health
        local health_checks=(
            "http://localhost:13133"  # Main collector
            "http://localhost:13134"  # Observer
            "http://localhost:9090/-/healthy"  # Prometheus
        )
        
        for endpoint in "${health_checks[@]}"; do
            if curl -f -s "$endpoint" >/dev/null; then
                log_success "$(basename "$endpoint") is healthy"
            else
                log_warning "$(basename "$endpoint") health check failed"
            fi
        done
    fi
    
    log_success "Local deployment completed"
    log_info "Access points:"
    log_info "  Grafana: http://localhost:3000 (admin/admin)"
    log_info "  Prometheus: http://localhost:9090"
    log_info "  Control API: http://localhost:8081"
}

# Function to deploy to AWS
deploy_aws() {
    log_info "Deploying Phoenix to AWS with Docker Swarm..."
    
    local region="${AWS_REGION:-us-west-2}"
    local stack_name="phoenix-${ENVIRONMENT}-stack"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "Dry run - would create AWS CloudFormation stack"
        log_info "Stack name: $stack_name"
        log_info "Region: $region"
        return
    fi
    
    # Create Docker context for AWS
    if ! docker context ls | grep -q "aws-phoenix"; then
        log_info "Creating AWS Docker context..."
        docker context create ecs aws-phoenix --region "$region"
    fi
    
    # Use AWS context
    docker context use aws-phoenix
    
    log_info "Deploying Phoenix stack to AWS..."
    
    if [ "$FORCE" = false ]; then
        read -p "Continue with AWS deployment? [y/N] " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && { log_info "Deployment cancelled"; exit 0; }
    fi
    
    # Deploy using Docker Compose for AWS ECS
    docker compose --project-name "$stack_name" up --detach
    
    # Switch back to default context
    docker context use default
    
    log_success "AWS deployment completed"
    log_info "Stack: $stack_name"
    log_info "Region: $region"
}

# Function to deploy to Azure
deploy_azure() {
    log_info "Deploying Phoenix to Azure Container Instances..."
    
    local region="${AZURE_REGION:-eastus}"
    local resource_group="phoenix-${ENVIRONMENT}-rg"
    local container_group="phoenix-${ENVIRONMENT}-group"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "Dry run - would create Azure Container Instances"
        log_info "Resource Group: $resource_group"
        log_info "Container Group: $container_group"
        log_info "Region: $region"
        return
    fi
    
    # Create resource group if it doesn't exist
    if ! az group show --name "$resource_group" >/dev/null 2>&1; then
        log_info "Creating resource group: $resource_group"
        az group create --name "$resource_group" --location "$region"
    fi
    
    # Create Docker context for Azure
    if ! docker context ls | grep -q "azure-phoenix"; then
        log_info "Creating Azure Docker context..."
        docker context create aci azure-phoenix --resource-group "$resource_group" --location "$region"
    fi
    
    # Use Azure context
    docker context use azure-phoenix
    
    log_info "Deploying Phoenix to Azure Container Instances..."
    
    if [ "$FORCE" = false ]; then
        read -p "Continue with Azure deployment? [y/N] " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && { log_info "Deployment cancelled"; exit 0; }
    fi
    
    # Deploy using Docker Compose for Azure Container Instances
    docker compose --project-name "$container_group" up --detach
    
    # Switch back to default context
    docker context use default
    
    log_success "Azure deployment completed"
    log_info "Container Group: $container_group"
    log_info "Resource Group: $resource_group"
}


# Function to show deployment status
show_status() {
    log_info "Deployment Status for $DEPLOYMENT_TARGET"
    
    case $DEPLOYMENT_TARGET in
        local)
            docker-compose ps
            ;;
        aws|azure)
            docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
            ;;
    esac
}

# Main execution
main() {
    log_info "Phoenix-vNext Unified Deployment"
    log_info "================================="
    
    parse_args "$@"
    
    log_info "Configuration:"
    log_info "  Target: $DEPLOYMENT_TARGET"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Namespace: $NAMESPACE"
    log_info "  Dry Run: $DRY_RUN"
    
    check_prerequisites
    
    if [ "$SKIP_BUILD" = false ]; then
        build_images
    fi
    
    case $DEPLOYMENT_TARGET in
        local)
            deploy_local
            ;;
        aws)
            deploy_aws
            ;;
        azure)
            deploy_azure
            ;;
        *)
            log_error "Unknown deployment target: $DEPLOYMENT_TARGET"
            show_usage
            exit 1
            ;;
    esac
    
    if [ "$DRY_RUN" = false ]; then
        show_status
    fi
    
    log_success "Deployment completed successfully!"
}

# Execute main function
main "$@"