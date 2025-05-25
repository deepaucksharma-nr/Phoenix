#!/bin/bash
# deploy-dev.sh - Deploy Phoenix Platform to development environment

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phoenix Platform Development Deployment ===${NC}"
echo ""

# Configuration
NAMESPACE="phoenix-dev"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-localhost:5000}"
VERSION="${VERSION:-dev-$(date +%Y%m%d-%H%M%S)}"

# Check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    
    # Check for required tools
    for tool in kubectl helm docker; do
        if ! command -v $tool &> /dev/null; then
            echo -e "${RED}✗ $tool is not installed${NC}"
            exit 1
        fi
    done
    
    # Check Kubernetes connectivity
    if ! kubectl cluster-info &> /dev/null; then
        echo -e "${RED}✗ Cannot connect to Kubernetes cluster${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ All prerequisites met${NC}"
}

# Build Docker images
build_images() {
    echo ""
    echo -e "${YELLOW}Building Docker images...${NC}"
    
    # Build Go services
    for project in projects/*/; do
        if [[ -f "$project/Dockerfile" ]] || [[ -f "$project/build/Dockerfile" ]]; then
            project_name=$(basename "$project")
            echo -n "Building $project_name... "
            
            dockerfile="$project/Dockerfile"
            if [[ -f "$project/build/Dockerfile" ]]; then
                dockerfile="$project/build/Dockerfile"
            fi
            
            if docker build -t "${DOCKER_REGISTRY}/phoenix/${project_name}:${VERSION}" \
                -f "$dockerfile" "$project" > /tmp/docker-build.log 2>&1; then
                echo -e "${GREEN}✓${NC}"
                
                # Push to registry
                if [[ "$DOCKER_REGISTRY" != "localhost:5000" ]]; then
                    docker push "${DOCKER_REGISTRY}/phoenix/${project_name}:${VERSION}"
                fi
            else
                echo -e "${RED}✗${NC}"
                echo "  Error details:"
                tail -10 /tmp/docker-build.log | sed 's/^/    /'
            fi
        fi
    done
}

# Create namespace
create_namespace() {
    echo ""
    echo -e "${YELLOW}Setting up namespace...${NC}"
    
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        echo "Namespace $NAMESPACE already exists"
    else
        kubectl create namespace "$NAMESPACE"
        echo -e "${GREEN}✓ Created namespace $NAMESPACE${NC}"
    fi
    
    # Set as default namespace
    kubectl config set-context --current --namespace="$NAMESPACE"
}

# Deploy infrastructure
deploy_infrastructure() {
    echo ""
    echo -e "${YELLOW}Deploying infrastructure components...${NC}"
    
    # Deploy CRDs
    if [[ -d "infrastructure/kubernetes/operators" ]]; then
        echo -n "Deploying CRDs... "
        if kubectl apply -f infrastructure/kubernetes/operators/ > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    fi
    
    # Deploy base manifests using Kustomize
    if [[ -f "infrastructure/kubernetes/base/kustomization.yaml" ]]; then
        echo -n "Deploying base resources... "
        if kubectl apply -k infrastructure/kubernetes/base/ > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
    fi
}

# Deploy services with Helm
deploy_helm() {
    echo ""
    echo -e "${YELLOW}Deploying Phoenix services with Helm...${NC}"
    
    # Update Helm dependencies
    if [[ -d "infrastructure/helm/phoenix" ]]; then
        cd infrastructure/helm/phoenix
        
        # Create values override for dev
        cat > values-dev.yaml << EOF
global:
  imageRegistry: ${DOCKER_REGISTRY}
  imageTag: ${VERSION}
  environment: development

# Development overrides
replicaCount: 1

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# Enable debug logging
logLevel: debug

# Disable production features
ingress:
  enabled: false
  
monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
    adminPassword: admin
EOF
        
        # Deploy with Helm
        echo "Installing Phoenix Helm chart..."
        helm upgrade --install phoenix . \
            --namespace "$NAMESPACE" \
            --values values-dev.yaml \
            --wait \
            --timeout 10m
        
        cd - > /dev/null
    fi
}

# Deploy monitoring stack
deploy_monitoring() {
    echo ""
    echo -e "${YELLOW}Deploying monitoring stack...${NC}"
    
    # Deploy Prometheus
    echo -n "Deploying Prometheus... "
    kubectl apply -f monitoring/prometheus/prometheus.yaml -n "$NAMESPACE" > /dev/null 2>&1 && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}"
    
    # Deploy Grafana dashboards
    if [[ -d "monitoring/grafana/dashboards" ]]; then
        echo -n "Deploying Grafana dashboards... "
        kubectl create configmap grafana-dashboards \
            --from-file=monitoring/grafana/dashboards/ \
            -n "$NAMESPACE" \
            --dry-run=client -o yaml | kubectl apply -f - > /dev/null 2>&1
        echo -e "${GREEN}✓${NC}"
    fi
}

# Wait for services to be ready
wait_for_services() {
    echo ""
    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    
    # Wait for deployments
    kubectl wait --for=condition=available --timeout=300s \
        deployment --all -n "$NAMESPACE" > /dev/null 2>&1 || true
    
    # Check pod status
    echo ""
    echo "Pod Status:"
    kubectl get pods -n "$NAMESPACE" | grep -v "^NAME"
}

# Port forwarding for local access
setup_port_forwarding() {
    echo ""
    echo -e "${YELLOW}Setting up port forwarding...${NC}"
    
    cat > /tmp/phoenix-port-forward.sh << 'EOF'
#!/bin/bash
# Port forwarding script for Phoenix Platform

NAMESPACE="phoenix-dev"

echo "Starting port forwarding for Phoenix services..."

# API Gateway
kubectl port-forward -n $NAMESPACE svc/phoenix-api 8080:8080 &

# Prometheus
kubectl port-forward -n $NAMESPACE svc/prometheus 9090:9090 &

# Grafana
kubectl port-forward -n $NAMESPACE svc/grafana 3000:3000 &

echo "Port forwarding established:"
echo "  API Gateway: http://localhost:8080"
echo "  Prometheus: http://localhost:9090"
echo "  Grafana: http://localhost:3000 (admin/admin)"
echo ""
echo "Press Ctrl+C to stop port forwarding"

wait
EOF
    
    chmod +x /tmp/phoenix-port-forward.sh
    echo -e "${GREEN}✓ Port forwarding script created${NC}"
    echo "  Run: /tmp/phoenix-port-forward.sh"
}

# Generate deployment report
generate_report() {
    echo ""
    echo -e "${YELLOW}Generating deployment report...${NC}"
    
    cat > DEPLOYMENT_REPORT.md << EOF
# Phoenix Platform Deployment Report

**Date**: $(date)  
**Environment**: Development  
**Namespace**: $NAMESPACE  
**Version**: $VERSION

## Deployment Summary

### Services Deployed
$(kubectl get deployments -n $NAMESPACE -o name | sed 's|deployment.apps/|- |')

### Pod Status
\`\`\`
$(kubectl get pods -n $NAMESPACE)
\`\`\`

### Service Endpoints
\`\`\`
$(kubectl get services -n $NAMESPACE)
\`\`\`

## Access Information

### Local Port Forwarding
Run the port forwarding script:
\`\`\`bash
/tmp/phoenix-port-forward.sh
\`\`\`

### Service URLs (with port forwarding)
- API Gateway: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## Next Steps

1. Verify service health:
   \`\`\`bash
   kubectl get pods -n $NAMESPACE
   \`\`\`

2. Check logs:
   \`\`\`bash
   kubectl logs -n $NAMESPACE -l app=phoenix-controller
   \`\`\`

3. Access Grafana dashboards:
   - Navigate to http://localhost:3000
   - Login with admin/admin
   - View Phoenix dashboards

4. Run integration tests:
   \`\`\`bash
   ./scripts/test-integration.sh
   \`\`\`
EOF
    
    echo -e "${GREEN}✓ Deployment report saved to DEPLOYMENT_REPORT.md${NC}"
}

# Main deployment flow
main() {
    check_prerequisites
    create_namespace
    
    # Optional: Build images
    if [[ "${BUILD_IMAGES:-false}" == "true" ]]; then
        build_images
    fi
    
    deploy_infrastructure
    deploy_helm
    deploy_monitoring
    wait_for_services
    setup_port_forwarding
    generate_report
    
    echo ""
    echo -e "${GREEN}=== Deployment Complete ===${NC}"
    echo "Namespace: $NAMESPACE"
    echo "Version: $VERSION"
    echo ""
    echo "To access services locally, run:"
    echo "  /tmp/phoenix-port-forward.sh"
}

# Run main function
main "$@"