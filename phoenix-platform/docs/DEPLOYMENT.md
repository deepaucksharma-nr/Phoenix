# Phoenix Platform Deployment Guide

This guide covers deployment procedures for the Phoenix platform across different environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Deployment Architectures](#deployment-architectures)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Helm Deployment](#helm-deployment)
5. [Configuration Management](#configuration-management)
6. [Production Deployment](#production-deployment)
7. [Monitoring Setup](#monitoring-setup)
8. [Troubleshooting Deployment](#troubleshooting-deployment)

## Prerequisites

### Required Tools
- kubectl 1.28+
- Helm 3.12+
- Docker (for building images)
- Access to container registry
- New Relic account with OTLP endpoint access

### Required Permissions
- Kubernetes cluster admin (for CRD installation)
- Namespace creation privileges
- Secret management access
- Ingress configuration rights

## Deployment Architectures

### Single Region Deployment
```
┌─────────────────┐
│   Ingress       │
└────────┬────────┘
         │
┌────────┴────────┐
│  Phoenix API    │
└────────┬────────┘
         │
┌────────┴────────────────┐
│  Experiment Controller   │
└────────┬────────────────┘
         │
┌────────┴────────┐
│  OTel Collectors │
│  (DaemonSet)     │
└──────────────────┘
```

### Multi-Region Deployment
- Federated Prometheus
- Cross-region replication
- Global load balancing

## Kubernetes Deployment

### 1. Create Namespace

```bash
kubectl create namespace phoenix-system
kubectl create namespace phoenix-experiments
```

### 2. Install CRDs

```bash
kubectl apply -f k8s/crds/
```

### 3. Create Secrets

```bash
# New Relic API Key
kubectl create secret generic newrelic-api-key \
  --namespace phoenix-system \
  --from-literal=api-key='YOUR_API_KEY'

# JWT Secret
kubectl create secret generic jwt-secret \
  --namespace phoenix-system \
  --from-literal=secret='YOUR_JWT_SECRET'

# Git Token (for config generation)
kubectl create secret generic git-credentials \
  --namespace phoenix-system \
  --from-literal=token='YOUR_GIT_TOKEN'
```

### 4. Deploy Core Components

```bash
# Using Kustomize
kubectl apply -k k8s/overlays/production

# Or using raw manifests
kubectl apply -f k8s/base/
```

## Helm Deployment

### 1. Add Helm Repository

```bash
helm repo add phoenix https://phoenix.io/helm-charts
helm repo update
```

### 2. Install Phoenix

```bash
helm install phoenix phoenix/phoenix \
  --namespace phoenix-system \
  --create-namespace \
  --values values-production.yaml
```

### 3. Custom Values File

```yaml
# values-production.yaml
global:
  domain: phoenix.example.com
  environment: production

api:
  replicas: 3
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2000m
      memory: 2Gi

postgresql:
  enabled: true
  auth:
    database: phoenix
    username: phoenix
  primary:
    persistence:
      size: 100Gi

newrelic:
  apiKey:
    secretName: newrelic-api-key
    secretKey: api-key
  endpoint: https://otlp.nr-data.net:4318

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  tls:
    - secretName: phoenix-tls
      hosts:
        - phoenix.example.com
```

## Configuration Management

### Environment-Specific Configs

1. **Development**
   ```yaml
   # k8s/overlays/development/config.yaml
   collectors:
     logLevel: debug
     sampling: 100
   ```

2. **Staging**
   ```yaml
   # k8s/overlays/staging/config.yaml
   collectors:
     logLevel: info
     sampling: 50
   ```

3. **Production**
   ```yaml
   # k8s/overlays/production/config.yaml
   collectors:
     logLevel: warn
     sampling: 10
   ```

### Pipeline Configuration

Deploy pipeline templates:
```bash
kubectl create configmap pipeline-templates \
  --namespace phoenix-system \
  --from-file=pipelines/templates/
```

## Production Deployment

### Pre-deployment Checklist

- [ ] Database backups configured
- [ ] Monitoring alerts set up
- [ ] Resource limits defined
- [ ] Network policies applied
- [ ] TLS certificates ready
- [ ] Backup/restore procedures tested

### Deployment Steps

1. **Build and Push Images**
   ```bash
   make docker VERSION=v1.0.0
   make push VERSION=v1.0.0
   ```

2. **Deploy Database**
   ```bash
   helm install postgresql bitnami/postgresql \
     --namespace phoenix-system \
     --values postgresql-values.yaml
   ```

3. **Run Migrations**
   ```bash
   kubectl run migrations \
     --namespace phoenix-system \
     --image=phoenix/migrations:v1.0.0 \
     --rm -it --restart=Never
   ```

4. **Deploy Phoenix**
   ```bash
   helm upgrade --install phoenix ./helm/phoenix \
     --namespace phoenix-system \
     --values values-production.yaml \
     --wait
   ```

5. **Verify Deployment**
   ```bash
   kubectl get pods -n phoenix-system
   kubectl get phoenixexperiments -A
   ```

### Rolling Updates

```bash
# Update API service
kubectl set image deployment/phoenix-api \
  api=phoenix/api:v1.1.0 \
  -n phoenix-system

# Monitor rollout
kubectl rollout status deployment/phoenix-api -n phoenix-system
```

### Rollback Procedures

```bash
# Rollback to previous version
kubectl rollout undo deployment/phoenix-api -n phoenix-system

# Rollback Helm release
helm rollback phoenix 1 -n phoenix-system
```

## Monitoring Setup

### 1. Deploy Prometheus

```bash
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values prometheus-values.yaml
```

### 2. Deploy Grafana Dashboards

```bash
kubectl create configmap grafana-dashboards \
  --namespace monitoring \
  --from-file=configs/monitoring/grafana/dashboards/
```

### 3. Configure Alerts

```yaml
# configs/monitoring/prometheus/rules/phoenix-alerts.yaml
groups:
  - name: phoenix
    rules:
      - alert: PhoenixAPIDown
        expr: up{job="phoenix-api"} == 0
        for: 5m
        annotations:
          summary: Phoenix API is down
          
      - alert: HighCollectorMemory
        expr: container_memory_usage_bytes{pod=~"collector-.*"} > 500000000
        for: 10m
        annotations:
          summary: Collector using high memory
```

## Troubleshooting Deployment

### Common Issues

1. **CRDs Not Found**
   ```bash
   # Reinstall CRDs
   kubectl apply -f k8s/crds/ --server-side
   ```

2. **Image Pull Errors**
   ```bash
   # Check image pull secrets
   kubectl get secret regcred -n phoenix-system
   
   # Create if missing
   kubectl create secret docker-registry regcred \
     --namespace phoenix-system \
     --docker-server=registry.example.com \
     --docker-username=user \
     --docker-password=pass
   ```

3. **Database Connection Issues**
   ```bash
   # Check database service
   kubectl get svc postgresql -n phoenix-system
   
   # Test connection
   kubectl run -it --rm debug \
     --image=postgres:15 \
     --restart=Never \
     -- psql -h postgresql -U phoenix
   ```

### Health Checks

```bash
# API Health
curl https://phoenix.example.com/health

# Metrics endpoint
curl https://phoenix.example.com/metrics

# Experiment status
kubectl get phoenixexperiments -A
```

### Logs Collection

```bash
# API logs
kubectl logs -n phoenix-system deployment/phoenix-api -f

# Controller logs
kubectl logs -n phoenix-system deployment/experiment-controller -f

# Collector logs (specific node)
kubectl logs -n phoenix-experiments -l app=otel-collector,node=node1
```

## Post-Deployment

### Verification Steps

1. **Create Test Experiment**
   ```bash
   phoenix experiment create \
     --name test-deployment \
     --baseline process-baseline-v1 \
     --candidate process-priority-filter-v1
   ```

2. **Check Metrics Flow**
   - Verify in Prometheus: http://prometheus.example.com
   - Check New Relic: Infrastructure > Hosts

3. **Load Testing**
   ```bash
   phoenix loadsim start \
     --profile realistic \
     --duration 10m
   ```

### Backup Configuration

```bash
# Backup CRDs
kubectl get phoenixexperiments -A -o yaml > experiments-backup.yaml

# Backup configurations
kubectl get configmaps -n phoenix-system -o yaml > configmaps-backup.yaml
```