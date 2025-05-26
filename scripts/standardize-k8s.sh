#!/bin/bash
# standardize-k8s.sh - Generate standardized Kubernetes manifests for Phoenix services
# Created by Abhinav as part of K8s manifest standardization task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default values
VERSION=$(cat VERSION || echo "latest")
DEFAULT_REPLICAS=2
DEFAULT_HTTP_PORT=8080
DEFAULT_METRICS_PORT=9090
DEFAULT_CPU_REQUEST="100m"
DEFAULT_CPU_LIMIT="500m"
DEFAULT_MEMORY_REQUEST="128Mi"
DEFAULT_MEMORY_LIMIT="512Mi"
IMAGE_REPOSITORY="ghcr.io/phoenix"

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TEMPLATE_DIR="$REPO_ROOT/deployments/kubernetes/base"
PROJECT_DIRS=("$REPO_ROOT/projects/"*)

# Function to generate manifests
generate_manifest() {
    local service_name="$1"
    local project_dir="$2"
    local output_dir="$3"
    local replicas="${4:-$DEFAULT_REPLICAS}"
    local http_port="${5:-$DEFAULT_HTTP_PORT}"
    local metrics_port="${6:-$DEFAULT_METRICS_PORT}"
    local cpu_request="${7:-$DEFAULT_CPU_REQUEST}"
    local cpu_limit="${8:-$DEFAULT_CPU_LIMIT}"
    local memory_request="${9:-$DEFAULT_MEMORY_REQUEST}"
    local memory_limit="${10:-$DEFAULT_MEMORY_LIMIT}"
    
    echo -e "${BLUE}Generating K8s manifests for $service_name${NC}"
    
    # Create output directory if it doesn't exist
    mkdir -p "$output_dir"
    
    # Generate deployment.yaml
    sed -e "s/\${SERVICE_NAME}/$service_name/g" \
        -e "s/\${REPLICAS}/$replicas/g" \
        -e "s/\${VERSION}/$VERSION/g" \
        -e "s/\${HTTP_PORT}/$http_port/g" \
        -e "s/\${METRICS_PORT}/$metrics_port/g" \
        -e "s/\${CPU_REQUEST}/$cpu_request/g" \
        -e "s/\${CPU_LIMIT}/$cpu_limit/g" \
        -e "s/\${MEMORY_REQUEST}/$memory_request/g" \
        -e "s/\${MEMORY_LIMIT}/$memory_limit/g" \
        -e "s/\${IMAGE_REPOSITORY}/$IMAGE_REPOSITORY/g" \
        "$TEMPLATE_DIR/deployment-template.yaml" > "$output_dir/deployment.yaml"
    
    # Generate service.yaml
    sed -e "s/\${SERVICE_NAME}/$service_name/g" \
        -e "s/\${HTTP_PORT}/$http_port/g" \
        -e "s/\${METRICS_PORT}/$metrics_port/g" \
        "$TEMPLATE_DIR/service-template.yaml" > "$output_dir/service.yaml"
    
    # Generate kustomization.yaml
    cat > "$output_dir/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- deployment.yaml
- service.yaml

commonLabels:
  app.kubernetes.io/name: $service_name
  app.kubernetes.io/part-of: phoenix-platform
  app.kubernetes.io/version: $VERSION
EOF
    
    echo -e "${GREEN}Successfully generated manifests for $service_name${NC}"
}

# Process all project directories
for project_dir in "${PROJECT_DIRS[@]}"; do
    # Extract service name from project directory
    service_name=$(basename "$project_dir")
    
    # Skip if not a valid service
    if [[ ! -d "$project_dir/cmd" && ! -d "$project_dir/src" ]]; then
        continue
    fi
    
    # Set output directory
    output_dir="$project_dir/deployments/k8s"
    
    # Generate manifests
    generate_manifest "$service_name" "$project_dir" "$output_dir"
done

# Create environment overlays
for env in dev staging prod; do
    mkdir -p "$REPO_ROOT/deployments/kubernetes/overlays/$env"
    cat > "$REPO_ROOT/deployments/kubernetes/overlays/$env/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
$(for project_dir in "${PROJECT_DIRS[@]}"; do
    service_name=$(basename "$project_dir")
    if [[ -d "$project_dir/deployments/k8s" ]]; then
        echo "- ../../../projects/$service_name/deployments/k8s"
    fi
done)

namespace: phoenix-$env

commonLabels:
  environment: $env

patches:
- path: resource-patches.yaml
  target:
    kind: Deployment
EOF

    # Create resource patches for different environments
    if [[ "$env" == "dev" ]]; then
        cat > "$REPO_ROOT/deployments/kubernetes/overlays/$env/resource-patches.yaml" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: all-deployments
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: all-containers
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 100m
            memory: 128Mi
EOF
    elif [[ "$env" == "staging" ]]; then
        cat > "$REPO_ROOT/deployments/kubernetes/overlays/$env/resource-patches.yaml" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: all-deployments
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: all-containers
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
EOF
    else
        cat > "$REPO_ROOT/deployments/kubernetes/overlays/$env/resource-patches.yaml" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: all-deployments
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: all-containers
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 512Mi
EOF
    fi
done

echo -e "${GREEN}K8s manifest standardization complete!${NC}"
