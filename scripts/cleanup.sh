#!/usr/bin/env bash
# Phoenix-vNext Cleanup Script
# Removes deployed resources from various environments

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
CLEANUP_TARGET="local"
ENVIRONMENT="development"
NAMESPACE="phoenix-system"
DRY_RUN=false
FORCE=false
KEEP_DATA=false

# Function to show usage
show_usage() {
    cat << EOF
Phoenix-vNext Cleanup Script

Usage: $0 [OPTIONS] TARGET

Targets:
  local     Clean up local Docker environment
  aws       Clean up AWS resources
  azure     Clean up Azure resources

Options:
  -e, --environment ENV     Environment to clean up (default: development)
  -n, --network NS          Docker network name (default: phoenix-network)
  -d, --dry-run            Show what would be cleaned without executing
  -f, --force              Force cleanup without confirmation
  -k, --keep-data          Keep persistent data volumes
  -h, --help               Show this help message

Examples:
  $0 local                          # Clean up local Docker environment
  $0 aws -e production             # Clean up AWS resources in production
  $0 azure --dry-run               # Show what would be cleaned in Azure

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
            -f|--force)
                FORCE=true
                shift
                ;;
            -k|--keep-data)
                KEEP_DATA=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            local|aws|azure)
                CLEANUP_TARGET="$1"
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

# Function to confirm cleanup
confirm_cleanup() {
    if [ "$FORCE" = true ]; then
        return 0
    fi
    
    log_warning "This will clean up Phoenix-vNext deployment:"
    log_warning "  Target: $CLEANUP_TARGET"
    log_warning "  Environment: $ENVIRONMENT"
    log_warning "  Namespace: $NAMESPACE"
    [ "$KEEP_DATA" = false ] && log_warning "  Data volumes will be DELETED"
    
    read -p "Are you sure you want to continue? [y/N] " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]] || { log_info "Cleanup cancelled"; exit 0; }
}

# Function to cleanup local environment
cleanup_local() {
    log_info "Cleaning up local Docker environment..."
    
    cd "$PROJECT_ROOT"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "Dry run - would execute:"
        log_info "  docker-compose down"
        [ "$KEEP_DATA" = false ] && log_info "  docker-compose down -v"
        log_info "  docker system prune -f"
        return
    fi
    
    # Stop and remove containers
    if [ "$KEEP_DATA" = true ]; then
        docker-compose down
    else
        docker-compose down -v
    fi
    
    # Clean up Phoenix-specific images
    log_info "Removing Phoenix images..."
    docker images --format "table {{.Repository}}:{{.Tag}}" | grep "phoenix/" | while read -r image; do
        log_info "Removing image: $image"
        docker rmi "$image" || true
    done
    
    # Clean up system resources
    docker system prune -f
    
    log_success "Local cleanup completed"
}

# Function to cleanup AWS resources
cleanup_aws() {
    log_info "Cleaning up AWS resources..."
    
    local region="${AWS_REGION:-us-west-2}"
    local stack_name="phoenix-${ENVIRONMENT}-stack"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "Dry run - would remove AWS CloudFormation stack"
        log_info "Stack name: $stack_name"
        log_info "Region: $region"
        return
    fi
    
    # Switch to AWS context if it exists
    if docker context ls | grep -q "aws-phoenix"; then
        log_info "Switching to AWS context..."
        docker context use aws-phoenix
        
        # Stop and remove stack
        log_info "Removing AWS stack: $stack_name"
        docker compose --project-name "$stack_name" down --volumes || true
        
        # Switch back to default context
        docker context use default
        
        # Remove AWS context
        docker context rm aws-phoenix --force || true
    else
        log_warning "AWS context not found, skipping stack cleanup"
    fi
    
    log_success "AWS cleanup completed"
}

# Function to cleanup Azure resources
cleanup_azure() {
    log_info "Cleaning up Azure resources..."
    
    local region="${AZURE_REGION:-eastus}"
    local resource_group="phoenix-${ENVIRONMENT}-rg"
    local container_group="phoenix-${ENVIRONMENT}-group"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "Dry run - would remove Azure Container Instances"
        log_info "Resource Group: $resource_group"
        log_info "Container Group: $container_group"
        return
    fi
    
    # Switch to Azure context if it exists
    if docker context ls | grep -q "azure-phoenix"; then
        log_info "Switching to Azure context..."
        docker context use azure-phoenix
        
        # Stop and remove container group
        log_info "Removing Azure container group: $container_group"
        docker compose --project-name "$container_group" down --volumes || true
        
        # Switch back to default context
        docker context use default
        
        # Remove Azure context
        docker context rm azure-phoenix --force || true
    else
        log_warning "Azure context not found, skipping container cleanup"
    fi
    
    # Optionally remove resource group (commented out for safety)
    # if [ "$KEEP_DATA" = false ]; then
    #     log_info "Removing resource group: $resource_group"
    #     az group delete --name "$resource_group" --yes --no-wait || true
    # fi
    
    log_success "Azure cleanup completed"
}


# Function to cleanup configuration files
cleanup_configs() {
    if [ "$KEEP_DATA" = false ]; then
        log_info "Cleaning up configuration files..."
        
        cd "$PROJECT_ROOT"
        
        # Remove generated files
        rm -rf data/
        rm -f CHECKSUMS.txt
        
        # Remove temporary files
        find . -name "*.log" -type f -delete
        find . -name "tmp/" -type d -exec rm -rf {} + 2>/dev/null || true
        
        log_success "Configuration cleanup completed"
    else
        log_info "Keeping configuration files as requested"
    fi
}

# Main execution
main() {
    log_info "Phoenix-vNext Cleanup"
    log_info "====================="
    
    parse_args "$@"
    
    log_info "Cleanup Configuration:"
    log_info "  Target: $CLEANUP_TARGET"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Namespace: $NAMESPACE"
    log_info "  Keep Data: $KEEP_DATA"
    log_info "  Dry Run: $DRY_RUN"
    
    confirm_cleanup
    
    case $CLEANUP_TARGET in
        local)
            cleanup_local
            ;;
        aws)
            cleanup_aws
            ;;
        azure)
            cleanup_azure
            ;;
        *)
            log_error "Unknown cleanup target: $CLEANUP_TARGET"
            show_usage
            exit 1
            ;;
    esac
    
    cleanup_configs
    
    log_success "Cleanup completed successfully!"
}

# Execute main function
main "$@"