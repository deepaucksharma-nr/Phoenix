# Phoenix Platform Kubernetes Manifests

This directory contains Kubernetes manifests for deploying the Phoenix Platform using Kustomize.

## Directory Structure

```
k8s/
├── base/                    # Base manifests
│   ├── kustomization.yaml   # Base kustomization
│   ├── namespace.yaml       # Phoenix namespace
│   ├── rbac.yaml           # RBAC rules
│   ├── api-gateway.yaml    # API Gateway deployment
│   ├── experiment-controller.yaml
│   ├── config-generator.yaml
│   ├── pipeline-operator.yaml
│   ├── loadsim-operator.yaml
│   ├── dashboard.yaml      # Web dashboard
│   ├── secrets.yaml        # Secret templates
│   ├── configmaps.yaml     # Configuration
│   └── network-policies.yaml
├── crds/                   # Custom Resource Definitions
│   ├── experiment.yaml
│   ├── phoenixprocesspipeline.yaml
│   └── loadsimulationjob.yaml
└── overlays/               # Environment-specific overlays
    ├── development/        # Development environment
    └── production/         # Production environment
```

## Prerequisites

1. Kubernetes cluster (1.28+)
2. kubectl configured
3. Kustomize (or kubectl with kustomize support)

## Quick Start

### Install CRDs

```bash
kubectl apply -f crds/
```

### Deploy to Development

```bash
# Build and view the manifests
kubectl kustomize k8s/overlays/development

# Apply to cluster
kubectl apply -k k8s/overlays/development
```

### Deploy to Production

1. First, update the production overlay with your specific values:
   - Container image registry and tags
   - Ingress hostnames
   - TLS certificates
   - Database credentials

2. Deploy:

```bash
# Review the manifests
kubectl kustomize k8s/overlays/production

# Apply to cluster
kubectl apply -k k8s/overlays/production
```

## Configuration

### Secrets

**Important**: The secrets in `base/secrets.yaml` are templates with placeholder values. You MUST update these before deploying to production:

1. **JWT Secret**: Generate a strong random value
   ```bash
   openssl rand -base64 32
   ```

2. **Database Credentials**: Use your PostgreSQL credentials
   ```yaml
   stringData:
     username: "your-db-user"
     password: "your-secure-password"
   ```

3. **TLS Certificates**: Use cert-manager or provide your own certificates

### Environment Variables

Common environment variables can be configured in:
- `base/configmaps.yaml` - Non-sensitive configuration
- Overlay patches - Environment-specific overrides

### Resource Limits

Default resource limits are conservative. Adjust based on your workload:
- Development: Lower limits for cost savings
- Production: Higher limits with autoscaling

## Components

### API Gateway
- HTTP/gRPC gateway for external access
- JWT authentication
- CORS support
- Routes to internal services

### Experiment Controller
- Manages experiment lifecycle
- State machine implementation
- Database integration
- Periodic reconciliation

### Config Generator
- Generates OTel collector configurations
- Template-based pipeline generation
- Kubernetes manifest creation

### Pipeline Operator
- Manages OTel collector DaemonSets
- Watches PhoenixProcessPipeline CRDs
- Handles configuration updates

### LoadSim Operator
- Manages load simulation jobs
- Creates process simulators
- Watches LoadSimulationJob CRDs

### Dashboard
- React-based web UI
- Nginx serving static files
- Connects to API Gateway

## Networking

Network policies are configured to:
- Deny all traffic by default
- Allow specific service-to-service communication
- Permit metrics scraping from monitoring namespace
- Enable ingress for public-facing services

## Monitoring

All services expose Prometheus metrics on their metrics ports. Configure ServiceMonitors in your monitoring namespace to scrape these endpoints.

## Security

1. **Pod Security Standards**: Restricted policy enforced
2. **Non-root containers**: All containers run as non-root
3. **Read-only root filesystem**: Enabled where possible
4. **Network policies**: Strict ingress/egress rules
5. **RBAC**: Least privilege access

## Troubleshooting

### Check deployment status
```bash
kubectl -n phoenix-system get all
```

### View logs
```bash
# API Gateway logs
kubectl -n phoenix-system logs -l app.kubernetes.io/name=api-gateway

# Experiment Controller logs
kubectl -n phoenix-system logs -l app.kubernetes.io/name=experiment-controller
```

### Verify CRDs
```bash
kubectl get crd | grep phoenix
```

### Port forwarding for debugging
```bash
# API Gateway
kubectl -n phoenix-system port-forward svc/api-gateway 8080:8080

# Dashboard
kubectl -n phoenix-system port-forward svc/dashboard 3000:80
```

## Production Checklist

- [ ] Update all secret values
- [ ] Configure proper ingress hostnames
- [ ] Set up TLS certificates (cert-manager recommended)
- [ ] Configure external PostgreSQL database
- [ ] Set appropriate resource limits
- [ ] Configure monitoring and alerting
- [ ] Set up backup procedures
- [ ] Test disaster recovery
- [ ] Configure autoscaling policies
- [ ] Review and adjust network policies