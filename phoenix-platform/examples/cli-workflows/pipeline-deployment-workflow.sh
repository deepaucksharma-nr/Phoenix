#!/bin/bash
# Pipeline Deployment Workflow Example
# This demonstrates direct pipeline deployment without experiments

set -e

# Configuration
API_URL=${PHOENIX_API_URL:-"http://localhost:8080"}
NAMESPACE="production"

echo "=== Phoenix Pipeline Deployment Workflow ==="
echo "This example shows how to deploy pipelines directly without experiments"
echo

# Check if logged in
echo "1. Checking authentication status..."
if ! phoenix auth status >/dev/null 2>&1; then
    echo "Not logged in. Please login first:"
    phoenix auth login
fi

# List available pipeline templates
echo -e "\n2. Listing available pipeline templates..."
phoenix pipeline templates list

# Deploy a specific pipeline configuration
echo -e "\n3. Deploying intelligent pipeline to production..."
DEPLOYMENT_ID=$(phoenix pipeline deploy \
    --name "prod-intelligent-pipeline" \
    --namespace "$NAMESPACE" \
    --template "process-intelligent-v1" \
    --description "Production intelligent pipeline deployment" \
    --config-override '{"sampling_rate": 0.1, "batch_size": 1000}' \
    --output json | jq -r '.id')

echo "Deployment created with ID: $DEPLOYMENT_ID"

# Check deployment status
echo -e "\n4. Monitoring deployment progress..."
phoenix pipeline deployment status "$DEPLOYMENT_ID" --follow

# List active deployments
echo -e "\n5. Listing all active deployments..."
phoenix pipeline deployments list --namespace "$NAMESPACE"

# Get deployment metrics
echo -e "\n6. Checking deployment metrics..."
phoenix pipeline deployment metrics "$DEPLOYMENT_ID"

# Update deployment configuration
echo -e "\n7. Updating deployment configuration..."
phoenix pipeline deployment update "$DEPLOYMENT_ID" \
    --config-override '{"sampling_rate": 0.05, "batch_size": 2000}' \
    --reason "Reducing sampling rate based on volume analysis"

# Rollback if needed (commented out for safety)
# echo -e "\n8. Rolling back deployment (if needed)..."
# phoenix pipeline deployment rollback "$DEPLOYMENT_ID" --reason "Performance regression detected"

# Get deployment history
echo -e "\n9. Viewing deployment history..."
phoenix pipeline deployment history "$DEPLOYMENT_ID"

# Export deployment configuration
echo -e "\n10. Exporting deployment configuration..."
phoenix pipeline deployment export "$DEPLOYMENT_ID" > "deployment-$DEPLOYMENT_ID.yaml"
echo "Configuration exported to deployment-$DEPLOYMENT_ID.yaml"

echo -e "\n=== Workflow Complete ==="
echo "Pipeline has been successfully deployed and configured."
echo "Use 'phoenix pipeline deployment status $DEPLOYMENT_ID' to check ongoing status."