#!/bin/bash
# monitoring-setup.sh - Set up monitoring & observability infrastructure for Phoenix Platform
# Created by Abhinav as part of monitoring & observability task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default configuration
NAMESPACE=${NAMESPACE:-monitoring}
DEPLOY_PROMETHEUS=${DEPLOY_PROMETHEUS:-true}
DEPLOY_GRAFANA=${DEPLOY_GRAFANA:-true}
DEPLOY_LOKI=${DEPLOY_LOKI:-true}
DEPLOY_JAEGER=${DEPLOY_JAEGER:-true}
DRY_RUN=${DRY_RUN:-false}
KUBE_CONTEXT=${KUBE_CONTEXT:-""}

# Display banner
echo -e "${BLUE}=== Phoenix Platform - Monitoring & Observability Setup ===${NC}"
echo ""
echo -e "This script will set up the following components:"
[ "$DEPLOY_PROMETHEUS" = "true" ] && echo -e "  - ${GREEN}✓${NC} Prometheus (metrics collection)"
[ "$DEPLOY_GRAFANA" = "true" ] && echo -e "  - ${GREEN}✓${NC} Grafana (visualization & dashboards)"
[ "$DEPLOY_LOKI" = "true" ] && echo -e "  - ${GREEN}✓${NC} Loki & Promtail (log aggregation)"
[ "$DEPLOY_JAEGER" = "true" ] && echo -e "  - ${GREEN}✓${NC} Jaeger (distributed tracing)"
echo ""
echo -e "Target namespace: ${YELLOW}$NAMESPACE${NC}"
[ -n "$KUBE_CONTEXT" ] && echo -e "Kubernetes context: ${YELLOW}$KUBE_CONTEXT${NC}"
[ "$DRY_RUN" = "true" ] && echo -e "${YELLOW}DRY RUN MODE - No actual changes will be applied${NC}"
echo ""

# Helper functions
function log_info() {
    echo -e "${BLUE}[INFO] $1${NC}"
}

function log_success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
}

function log_error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

function log_warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

# Kubernetes helpers
function kubectl_cmd() {
    if [ -n "$KUBE_CONTEXT" ]; then
        echo "kubectl --context=$KUBE_CONTEXT $*"
    else
        echo "kubectl $*"
    fi
}

# Check prerequisites
function check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl."
    fi
    
    # Check if can connect to Kubernetes
    local kubectl_cmd_str=$(kubectl_cmd "get nodes")
    if ! eval "$kubectl_cmd_str" &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    fi
    
    # Check if namespace exists, create if not
    local check_ns_cmd=$(kubectl_cmd "get namespace $NAMESPACE")
    if ! eval "$check_ns_cmd" &> /dev/null; then
        log_info "Namespace $NAMESPACE does not exist, creating..."
        
        local create_ns_cmd=$(kubectl_cmd "create namespace $NAMESPACE")
        if [ "$DRY_RUN" = "true" ]; then
            log_info "Would run: $create_ns_cmd"
        else
            if eval "$create_ns_cmd"; then
                log_success "Created namespace $NAMESPACE"
            else
                log_error "Failed to create namespace $NAMESPACE"
            fi
        fi
    else
        log_success "Namespace $NAMESPACE already exists"
    fi
    
    log_success "All prerequisites satisfied"
}

# Deploy Prometheus
function deploy_prometheus() {
    if [ "$DEPLOY_PROMETHEUS" != "true" ]; then
        return 0
    fi
    
    log_info "Deploying Prometheus..."
    
    # Apply Prometheus manifests
    local prometheus_dir="$REPO_ROOT/deployments/monitoring/prometheus"
    local deploy_cmd=$(kubectl_cmd "apply -f $prometheus_dir/configmap.yaml -f $prometheus_dir/deployment.yaml -f $prometheus_dir/rules.yaml")
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "Would run: $deploy_cmd"
    else
        if eval "$deploy_cmd"; then
            log_success "Prometheus deployed successfully"
        else
            log_error "Failed to deploy Prometheus"
        fi
    fi
}

# Deploy Grafana
function deploy_grafana() {
    if [ "$DEPLOY_GRAFANA" != "true" ]; then
        return 0
    fi
    
    log_info "Deploying Grafana..."
    
    # Apply Grafana manifests
    local grafana_dir="$REPO_ROOT/deployments/monitoring/grafana"
    local deploy_cmd=$(kubectl_cmd "apply -f $grafana_dir/config.yaml -f $grafana_dir/deployment.yaml")
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "Would run: $deploy_cmd"
    else
        if eval "$deploy_cmd"; then
            log_success "Grafana deployed successfully"
        else
            log_error "Failed to deploy Grafana"
        fi
    fi
}

# Deploy Loki and Promtail
function deploy_loki() {
    if [ "$DEPLOY_LOKI" != "true" ]; then
        return 0
    fi
    
    log_info "Deploying Loki & Promtail..."
    
    # Apply Loki manifests
    local loki_dir="$REPO_ROOT/deployments/monitoring/loki"
    local loki_cmd=$(kubectl_cmd "apply -f $loki_dir/statefulset.yaml")
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "Would run: $loki_cmd"
    else
        if eval "$loki_cmd"; then
            log_success "Loki deployed successfully"
        else
            log_error "Failed to deploy Loki"
        fi
    fi
    
    # Apply Promtail manifests
    local promtail_cmd=$(kubectl_cmd "apply -f $loki_dir/promtail.yaml")
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "Would run: $promtail_cmd"
    else
        if eval "$promtail_cmd"; then
            log_success "Promtail deployed successfully"
        else
            log_error "Failed to deploy Promtail"
        fi
    fi
}

# Deploy Jaeger
function deploy_jaeger() {
    if [ "$DEPLOY_JAEGER" != "true" ]; then
        return 0
    fi
    
    log_info "Deploying Jaeger..."
    
    # Apply Jaeger manifests
    local jaeger_dir="$REPO_ROOT/deployments/monitoring/jaeger"
    local deploy_cmd=$(kubectl_cmd "apply -f $jaeger_dir/deployment.yaml")
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "Would run: $deploy_cmd"
    else
        if eval "$deploy_cmd"; then
            log_success "Jaeger deployed successfully"
        else
            log_error "Failed to deploy Jaeger"
        fi
    fi
}

# Add Loki datasource to Grafana
function configure_grafana_datasource() {
    if [ "$DEPLOY_GRAFANA" != "true" ] || [ "$DEPLOY_LOKI" != "true" ] || [ "$DRY_RUN" = "true" ]; then
        return 0
    fi
    
    log_info "Configuring Grafana to use Loki..."
    
    # Wait for Grafana to be ready
    log_info "Waiting for Grafana to be ready..."
    local wait_cmd=$(kubectl_cmd "wait --for=condition=available deployment/grafana -n $NAMESPACE --timeout=120s")
    if ! eval "$wait_cmd"; then
        log_warning "Timed out waiting for Grafana to be ready. Loki datasource may need manual configuration."
        return 0
    fi
    
    # Get Grafana pod name
    local grafana_pod=$(eval "$(kubectl_cmd "get pods -n $NAMESPACE -l app=grafana -o jsonpath='{.items[0].metadata.name}'")")
    
    # Check if Loki datasource already exists
    local check_ds_cmd=$(kubectl_cmd "exec -n $NAMESPACE $grafana_pod -- curl -s http://admin:phoenix-admin-password@localhost:3000/api/datasources/name/Loki")
    if eval "$check_ds_cmd" | grep -q "id"; then
        log_info "Loki datasource already exists in Grafana"
        return 0
    fi
    
    # Create Loki datasource
    log_info "Adding Loki datasource..."
    local add_ds_cmd=$(kubectl_cmd "exec -n $NAMESPACE $grafana_pod -- curl -s -X POST -H \"Content-Type: application/json\" -d '{\"name\":\"Loki\",\"type\":\"loki\",\"url\":\"http://loki:3100\",\"access\":\"proxy\",\"isDefault\":false}' http://admin:phoenix-admin-password@localhost:3000/api/datasources")
    
    if eval "$add_ds_cmd" | grep -q "datasource added"; then
        log_success "Loki datasource added successfully"
    else
        log_warning "Failed to add Loki datasource. You may need to configure it manually."
    fi
}

# Configure Jaeger datasource for Grafana
function configure_jaeger_datasource() {
    if [ "$DEPLOY_GRAFANA" != "true" ] || [ "$DEPLOY_JAEGER" != "true" ] || [ "$DRY_RUN" = "true" ]; then
        return 0
    fi
    
    log_info "Configuring Grafana to use Jaeger..."
    
    # Get Grafana pod name
    local grafana_pod=$(eval "$(kubectl_cmd "get pods -n $NAMESPACE -l app=grafana -o jsonpath='{.items[0].metadata.name}'")")
    
    # Check if Jaeger datasource already exists
    local check_ds_cmd=$(kubectl_cmd "exec -n $NAMESPACE $grafana_pod -- curl -s http://admin:phoenix-admin-password@localhost:3000/api/datasources/name/Jaeger")
    if eval "$check_ds_cmd" | grep -q "id"; then
        log_info "Jaeger datasource already exists in Grafana"
        return 0
    fi
    
    # Create Jaeger datasource
    log_info "Adding Jaeger datasource..."
    local add_ds_cmd=$(kubectl_cmd "exec -n $NAMESPACE $grafana_pod -- curl -s -X POST -H \"Content-Type: application/json\" -d '{\"name\":\"Jaeger\",\"type\":\"jaeger\",\"url\":\"http://jaeger-query:16686\",\"access\":\"proxy\",\"isDefault\":false}' http://admin:phoenix-admin-password@localhost:3000/api/datasources")
    
    if eval "$add_ds_cmd" | grep -q "datasource added"; then
        log_success "Jaeger datasource added successfully"
    else
        log_warning "Failed to add Jaeger datasource. You may need to configure it manually."
    fi
}

# Show endpoints
function show_service_info() {
    if [ "$DRY_RUN" = "true" ]; then
        return 0
    fi
    
    echo ""
    echo -e "${BLUE}=== Monitoring Services Information ===${NC}"
    echo ""
    
    # Get service info
    local get_services_cmd=$(kubectl_cmd "get services -n $NAMESPACE")
    eval "$get_services_cmd"
    
    echo ""
    echo "To access services locally, use port-forwarding:"
    [ "$DEPLOY_PROMETHEUS" = "true" ] && echo "  - Prometheus: kubectl port-forward -n $NAMESPACE svc/prometheus 9090:9090"
    [ "$DEPLOY_GRAFANA" = "true" ] && echo "  - Grafana: kubectl port-forward -n $NAMESPACE svc/grafana 3000:3000"
    [ "$DEPLOY_JAEGER" = "true" ] && echo "  - Jaeger UI: kubectl port-forward -n $NAMESPACE svc/jaeger-query 16686:16686"
    echo ""
    
    [ "$DEPLOY_GRAFANA" = "true" ] && echo "Grafana credentials:"
    [ "$DEPLOY_GRAFANA" = "true" ] && echo "  - Username: admin"
    [ "$DEPLOY_GRAFANA" = "true" ] && echo "  - Password: phoenix-admin-password (change this in production)"
    echo ""
}

# Main function
function main() {
    check_prerequisites
    deploy_prometheus
    deploy_grafana
    deploy_loki
    deploy_jaeger
    configure_grafana_datasource
    configure_jaeger_datasource
    show_service_info
    
    log_success "Monitoring setup completed successfully"
}

# Display help
function show_help() {
    echo "Usage: $(basename "$0") [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -n, --namespace NAME    Target Kubernetes namespace (default: monitoring)"
    echo "  -c, --context CONTEXT   Use specific Kubernetes context"
    echo "  --skip-prometheus       Skip Prometheus deployment"
    echo "  --skip-grafana          Skip Grafana deployment"
    echo "  --skip-loki             Skip Loki & Promtail deployment"
    echo "  --skip-jaeger           Skip Jaeger deployment"
    echo "  --dry-run               Show commands without executing them"
    echo ""
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -c|--context)
            KUBE_CONTEXT="$2"
            shift 2
            ;;
        --skip-prometheus)
            DEPLOY_PROMETHEUS=false
            shift
            ;;
        --skip-grafana)
            DEPLOY_GRAFANA=false
            shift
            ;;
        --skip-loki)
            DEPLOY_LOKI=false
            shift
            ;;
        --skip-jaeger)
            DEPLOY_JAEGER=false
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            ;;
    esac
done

# Execute main function
main
