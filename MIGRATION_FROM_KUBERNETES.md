# Migration Guide: From Kubernetes to Docker Compose

This guide helps users migrate from the previous Kubernetes-based deployment to the new Docker Compose and single-VM deployment model.

## Overview

Phoenix Platform has transitioned from Kubernetes to a simpler, more maintainable deployment model using Docker Compose. This change reduces operational complexity while maintaining all core functionality.

## Why the Change?

1. **Reduced Complexity**: Docker Compose provides a simpler operational model
2. **Lower Resource Requirements**: Single-VM deployments require fewer resources
3. **Easier Maintenance**: No need for Kubernetes expertise
4. **Faster Development**: Simplified local development workflow
5. **Cost Effective**: Reduced infrastructure requirements

## Architecture Changes

### Before (Kubernetes)
- Deployments managed via kubectl and Helm
- DaemonSets for agent deployment
- ConfigMaps and Secrets for configuration
- Service mesh for inter-service communication
- Kubernetes-native service discovery

### After (Docker Compose)
- Services managed via docker-compose
- Agents deployed via systemd on host machines
- Environment files for configuration
- Direct container networking
- File-based service discovery

## Migration Steps

### 1. Export Data from Kubernetes

```bash
# Export experiments data
kubectl exec -n phoenix-system deployment/phoenix-api -- \
  pg_dump -U phoenix phoenix > phoenix_backup.sql

# Export Prometheus data (optional)
kubectl exec -n phoenix-system prometheus-0 -- \
  tar czf /tmp/prometheus-data.tar.gz /prometheus/*
kubectl cp phoenix-system/prometheus-0:/tmp/prometheus-data.tar.gz ./prometheus-data.tar.gz
```

### 2. Prepare Single-VM Environment

```bash
# Clone the repository
git clone https://github.com/phoenix/platform.git
cd platform

# Run setup script
./deployments/single-vm/scripts/setup-single-vm.sh
```

### 3. Configure Environment

Create `.env` file:
```bash
cat > deployments/single-vm/.env << EOF
# API Configuration
PHOENIX_API_PORT=8080
PHOENIX_API_HOST=0.0.0.0

# Database Configuration
POSTGRES_USER=phoenix
POSTGRES_PASSWORD=your-secure-password
POSTGRES_DB=phoenix

# Monitoring
PROMETHEUS_RETENTION=30d
GRAFANA_ADMIN_PASSWORD=your-admin-password

# Security
JWT_SECRET=your-jwt-secret
TLS_ENABLED=true
EOF
```

### 4. Import Data

```bash
# Start only the database
docker-compose -f deployments/single-vm/docker-compose.yml up -d postgres

# Import data
docker exec -i phoenix_postgres psql -U phoenix phoenix < phoenix_backup.sql

# Start all services
docker-compose -f deployments/single-vm/docker-compose.yml up -d
```

### 5. Deploy Agents

On each host machine:
```bash
# Download and run agent installer
curl -sSL https://your-phoenix-api/install-agent.sh | sudo bash -s -- \
  --api-url http://your-phoenix-api:8080 \
  --host-id $(hostname)
```

## Configuration Mapping

### Service Endpoints

| Kubernetes | Docker Compose |
|------------|----------------|
| `http://phoenix-api.phoenix-system:8080` | `http://phoenix-api:8080` |
| `http://prometheus.phoenix-system:9090` | `http://prometheus:9090` |
| `http://prometheus-pushgateway:9091` | `http://pushgateway:9091` |

### Agent Configuration

**Kubernetes (DaemonSet):**
```yaml
env:
- name: PHOENIX_API_URL
  value: "http://phoenix-api:8080"
- name: PHOENIX_HOST_ID
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
```

**Docker Compose (systemd):**
```ini
[Service]
Environment="PHOENIX_API_URL=http://your-phoenix-api:8080"
Environment="PHOENIX_HOST_ID=%H"
```

### Pipeline Deployments

**Kubernetes:**
```bash
kubectl apply -f pipeline-deployment.yaml
```

**Docker Compose:**
```bash
phoenix-cli pipeline deploy \
  --name production-optimization \
  --pipeline process-topk-v1 \
  --target "production" \
  --selector "app=webserver"
```

## Operational Changes

### Monitoring

- Prometheus and Grafana remain the same
- Access via exposed ports instead of Ingress
- Dashboards are automatically provisioned

### Scaling

**Kubernetes:** Horizontal Pod Autoscaler
**Docker Compose:** 
- Manual scaling via `docker-compose scale`
- Auto-scaling script: `./scripts/auto-scale-monitor.sh`

### Backups

```bash
# Automated backup script
./deployments/single-vm/scripts/backup.sh

# Restore from backup
./deployments/single-vm/scripts/restore.sh backup-20240115.tar.gz
```

### High Availability

For HA requirements:
1. Use external PostgreSQL cluster
2. Deploy multiple API instances behind load balancer
3. Configure shared storage for Prometheus

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check port usage
   sudo netstat -tlnp | grep -E '8080|9090|5432'
   ```

2. **Agent Connection Issues**
   ```bash
   # Check agent logs
   sudo journalctl -u phoenix-agent -f
   ```

3. **Database Migration**
   ```bash
   # Verify database connection
   docker exec phoenix_postgres psql -U phoenix -c "SELECT version();"
   ```

## Rollback Plan

If you need to rollback to Kubernetes:

1. Keep Kubernetes manifests in version control
2. Maintain database backups from both systems
3. Document any configuration changes
4. Test rollback procedure in staging

## Benefits After Migration

1. **Simplified Operations**: No Kubernetes cluster to maintain
2. **Reduced Costs**: Single VM vs. multi-node cluster
3. **Faster Deployments**: Direct Docker commands
4. **Easier Debugging**: Standard Linux tools
5. **Better Resource Utilization**: No Kubernetes overhead

## Support

For migration assistance:
- Documentation: `/docs/operations/`
- Single-VM Guide: `/deployments/single-vm/README.md`
- Issues: https://github.com/phoenix/platform/issues