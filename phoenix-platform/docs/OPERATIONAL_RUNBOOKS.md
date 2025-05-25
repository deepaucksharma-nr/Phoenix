# Phoenix Platform Operational Runbooks

**Version**: 1.0  
**Last Updated**: January 25, 2025

## Overview

This document contains step-by-step operational procedures for common tasks, incident response, and maintenance activities for the Phoenix Platform.

## Table of Contents

1. [Service Health Checks](#1-service-health-checks)
2. [Incident Response Runbooks](#2-incident-response-runbooks)
3. [Deployment Procedures](#3-deployment-procedures)
4. [Database Operations](#4-database-operations)
5. [Monitoring and Alerting](#5-monitoring-and-alerting)
6. [Performance Troubleshooting](#6-performance-troubleshooting)
7. [Security Operations](#7-security-operations)
8. [Maintenance Procedures](#8-maintenance-procedures)

## 1. Service Health Checks

### 1.1 Quick Health Check

**Purpose**: Verify all Phoenix services are operational

**Steps**:
```bash
#!/bin/bash
# health-check.sh

echo "=== Phoenix Platform Health Check ==="
echo "Time: $(date)"

# Check API Gateway
echo -n "API Gateway: "
curl -s https://api.phoenix.io/health | jq -r '.status' || echo "FAIL"

# Check services
for service in api controller generator; do
    echo -n "$service: "
    kubectl get pod -n phoenix-system -l app=$service -o jsonpath='{.items[0].status.phase}'
    echo ""
done

# Check database
echo -n "PostgreSQL: "
kubectl exec -n phoenix-system deployment/postgres -- pg_isready || echo "FAIL"

# Check Redis
echo -n "Redis: "
kubectl exec -n phoenix-system deployment/redis -- redis-cli ping || echo "FAIL"

# Check Prometheus
echo -n "Prometheus: "
curl -s http://prometheus.phoenix.io/-/healthy | grep -q "Prometheus is Healthy" && echo "OK" || echo "FAIL"

echo "=== Health Check Complete ==="
```

### 1.2 Detailed Service Diagnostics

```bash
# Get detailed service status
kubectl get all -n phoenix-system

# Check recent events
kubectl get events -n phoenix-system --sort-by='.lastTimestamp' | tail -20

# Check pod logs for errors
for pod in $(kubectl get pods -n phoenix-system -o name); do
    echo "=== Checking $pod ==="
    kubectl logs $pod -n phoenix-system --tail=50 | grep -E "ERROR|FATAL|PANIC" || echo "No errors found"
done
```

## 2. Incident Response Runbooks

### 2.1 API Service Down

**Symptoms**: 
- API health check failing
- 5xx errors from API Gateway
- No response from https://api.phoenix.io

**Severity**: SEV1 (Critical)

**Response Steps**:

```bash
# 1. Verify the issue
curl -v https://api.phoenix.io/health

# 2. Check pod status
kubectl get pods -n phoenix-system -l app=phoenix-api

# 3. If pods are crashing, check logs
kubectl logs -n phoenix-system -l app=phoenix-api --tail=100

# 4. Check for resource issues
kubectl top pods -n phoenix-system -l app=phoenix-api

# 5. Restart if needed
kubectl rollout restart deployment/phoenix-api -n phoenix-system

# 6. Scale up if under load
kubectl scale deployment/phoenix-api --replicas=5 -n phoenix-system

# 7. If database connection issues
kubectl exec -n phoenix-system deployment/phoenix-api -- \
    psql $DATABASE_URL -c "SELECT 1"

# 8. Check recent deployments
kubectl rollout history deployment/phoenix-api -n phoenix-system

# 9. Rollback if recent deployment caused issues
kubectl rollout undo deployment/phoenix-api -n phoenix-system
```

**Escalation**: If not resolved in 15 minutes, page on-call SRE

### 2.2 High Experiment Failure Rate

**Symptoms**:
- Alert: `HighExperimentFailureRate`
- Multiple experiments in FAILED state
- Customer complaints about failed experiments

**Severity**: SEV2 (High)

**Response Steps**:

```sql
-- 1. Check recent experiment failures
SELECT 
    id,
    name,
    state,
    error_message,
    updated_at
FROM experiments
WHERE state = 'failed'
  AND updated_at > NOW() - INTERVAL '1 hour'
ORDER BY updated_at DESC
LIMIT 20;

-- 2. Check for common failure patterns
SELECT 
    error_message,
    COUNT(*) as failure_count
FROM experiments
WHERE state = 'failed'
  AND updated_at > NOW() - INTERVAL '1 hour'
GROUP BY error_message
ORDER BY failure_count DESC;
```

```bash
# 3. Check controller logs
kubectl logs -n phoenix-system deployment/experiment-controller --tail=200 | grep ERROR

# 4. Check if config generator is working
kubectl logs -n phoenix-system deployment/config-generator --tail=100

# 5. Verify Kubernetes API access
kubectl auth can-i create phoenixprocesspipelines --as=system:serviceaccount:phoenix-system:pipeline-operator

# 6. Check for resource quotas
kubectl describe resourcequota -n phoenix-experiments

# 7. Temporary mitigation - pause new experiments
kubectl scale deployment/experiment-controller --replicas=0 -n phoenix-system
# Fix issue...
kubectl scale deployment/experiment-controller --replicas=1 -n phoenix-system
```

### 2.3 Database Connection Pool Exhausted

**Symptoms**:
- Alert: `DatabaseConnectionPoolExhausted`
- Slow API responses
- Connection timeout errors

**Severity**: SEV1 (Critical)

**Response Steps**:

```sql
-- 1. Check current connections
SELECT 
    datname,
    usename,
    application_name,
    client_addr,
    state,
    query_start,
    state_change
FROM pg_stat_activity
WHERE datname = 'phoenix'
ORDER BY query_start;

-- 2. Count connections by application
SELECT 
    application_name,
    COUNT(*) as connection_count
FROM pg_stat_activity
WHERE datname = 'phoenix'
GROUP BY application_name
ORDER BY connection_count DESC;

-- 3. Kill idle connections older than 5 minutes
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'phoenix'
  AND state = 'idle'
  AND state_change < NOW() - INTERVAL '5 minutes';

-- 4. Kill long-running queries (>10 minutes)
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'phoenix'
  AND state != 'idle'
  AND query_start < NOW() - INTERVAL '10 minutes';
```

```bash
# 5. Restart connection pooler
kubectl rollout restart deployment/pgbouncer -n phoenix-system

# 6. Scale up database connections temporarily
kubectl edit configmap pgbouncer-config -n phoenix-system
# Increase max_client_conn and default_pool_size

# 7. Identify connection leaks
for pod in $(kubectl get pods -n phoenix-system -o name); do
    echo "=== $pod connection count ==="
    kubectl exec $pod -n phoenix-system -- netstat -an | grep 5432 | wc -l
done
```

### 2.4 OTel Collector High Memory Usage

**Symptoms**:
- Alert: `CollectorHighMemoryUsage`
- Collectors being OOM killed
- Metrics data loss

**Severity**: SEV2 (High)

**Response Steps**:

```bash
# 1. Check memory usage
kubectl top pods -n phoenix-experiments -l app=otel-collector

# 2. Check for memory pressure
kubectl describe nodes | grep -A 5 "Allocated resources"

# 3. Get collector metrics
curl http://otel-collector.phoenix-experiments:8888/metrics | grep memory

# 4. Reduce batch size temporarily
kubectl edit configmap otel-collector-config -n phoenix-experiments
# Reduce batch size from 8192 to 1000

# 5. Restart collectors
kubectl rollout restart daemonset/otel-collector -n phoenix-experiments

# 6. If specific nodes affected, cordon them
kubectl cordon node-xyz
kubectl drain node-xyz --ignore-daemonsets

# 7. Analyze pipeline configuration
kubectl logs -n phoenix-experiments daemonset/otel-collector | grep -E "processor|exporter"
```

## 3. Deployment Procedures

### 3.1 Standard Deployment

**Purpose**: Deploy new version to production

**Pre-requisites**:
- All tests passing in CI
- Approval from team lead
- No ongoing incidents

**Steps**:

```bash
#!/bin/bash
# deploy.sh

VERSION=$1
ENVIRONMENT=$2

echo "Deploying Phoenix Platform $VERSION to $ENVIRONMENT"

# 1. Pre-deployment checks
./scripts/pre-deploy-check.sh $ENVIRONMENT

# 2. Create deployment record
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: deployment-$VERSION
  namespace: phoenix-system
data:
  version: "$VERSION"
  timestamp: "$(date -Iseconds)"
  deployer: "$USER"
EOF

# 3. Update image tags
for service in api controller generator; do
    kubectl set image deployment/$service \
        $service=ghcr.io/phoenix/platform/$service:$VERSION \
        -n phoenix-system
done

# 4. Wait for rollout
for service in api controller generator; do
    kubectl rollout status deployment/$service -n phoenix-system
done

# 5. Run smoke tests
./scripts/smoke-test.sh $ENVIRONMENT

# 6. Update deployment status
if [ $? -eq 0 ]; then
    echo "Deployment successful"
    ./scripts/notify-slack.sh "✅ Deployed $VERSION to $ENVIRONMENT"
else
    echo "Deployment failed, initiating rollback"
    ./scripts/rollback.sh $ENVIRONMENT
fi
```

### 3.2 Emergency Hotfix

**Purpose**: Deploy critical fix bypassing normal process

**Authorization Required**: VP Engineering

**Steps**:

```bash
# 1. Create hotfix branch
git checkout -b hotfix/critical-fix

# 2. Apply fix and test locally
# ... make changes ...
go test ./...

# 3. Build and push hotfix image
docker build -t ghcr.io/phoenix/platform/api:hotfix-$(git rev-parse --short HEAD) .
docker push ghcr.io/phoenix/platform/api:hotfix-$(git rev-parse --short HEAD)

# 4. Deploy to single pod first
kubectl patch deployment phoenix-api -n phoenix-system --type='json' \
  -p='[{"op": "replace", "path": "/spec/replicas", "value": 1}]'

kubectl set image deployment/phoenix-api \
  api=ghcr.io/phoenix/platform/api:hotfix-$(git rev-parse --short HEAD) \
  -n phoenix-system

# 5. Monitor for 5 minutes
watch 'kubectl logs -n phoenix-system deployment/phoenix-api --tail=50'

# 6. If stable, scale back up
kubectl scale deployment/phoenix-api --replicas=3 -n phoenix-system

# 7. Create proper PR for main branch
git push origin hotfix/critical-fix
```

## 4. Database Operations

### 4.1 Database Backup

**Purpose**: Manual backup before major changes

**Steps**:

```bash
#!/bin/bash
# backup-database.sh

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="phoenix_manual_backup_$TIMESTAMP"

# 1. Create backup
kubectl exec -n phoenix-system deployment/postgres -- \
    pg_dump -U phoenix -d phoenix -Fc > $BACKUP_NAME.dump

# 2. Verify backup
pg_restore -l $BACKUP_NAME.dump | head -20

# 3. Upload to S3
aws s3 cp $BACKUP_NAME.dump s3://phoenix-backups/manual/$BACKUP_NAME.dump

# 4. Create backup record
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: backup-$TIMESTAMP
  namespace: phoenix-system
data:
  backup_name: "$BACKUP_NAME"
  size: "$(ls -lh $BACKUP_NAME.dump | awk '{print $5}')"
  created_by: "$USER"
  reason: "Manual backup before deployment"
EOF

echo "Backup completed: $BACKUP_NAME"
```

### 4.2 Database Migration

**Purpose**: Apply schema changes

**Steps**:

```bash
# 1. Backup database first
./scripts/backup-database.sh

# 2. Check migration status
kubectl exec -n phoenix-system deployment/postgres -- \
    psql -U phoenix -d phoenix -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 10;"

# 3. Test migration in staging
kubectl exec -n phoenix-staging deployment/postgres -- \
    psql -U phoenix -d phoenix -f /migrations/new_migration.sql

# 4. Apply migration in production
kubectl exec -n phoenix-system deployment/postgres -- \
    psql -U phoenix -d phoenix -f /migrations/new_migration.sql

# 5. Verify migration
kubectl exec -n phoenix-system deployment/postgres -- \
    psql -U phoenix -d phoenix -c "\d+ table_name"

# 6. Update migration record
kubectl exec -n phoenix-system deployment/postgres -- \
    psql -U phoenix -d phoenix -c "INSERT INTO schema_migrations (version) VALUES ('$(date +%Y%m%d%H%M%S)');"
```

## 5. Monitoring and Alerting

### 5.1 Add New Alert

**Purpose**: Create new Prometheus alert rule

**Steps**:

```yaml
# 1. Edit prometheus rules
kubectl edit configmap prometheus-rules -n monitoring

# 2. Add new rule
- alert: NewAlertName
  expr: your_prometheus_query > threshold
  for: 5m
  labels:
    severity: warning
    team: platform
  annotations:
    summary: "Brief description"
    description: "Detailed description with {{ $value }}"
    runbook_url: "https://wiki.phoenix.io/runbooks/new-alert"

# 3. Reload Prometheus
curl -X POST http://prometheus:9090/-/reload

# 4. Verify alert is loaded
curl http://prometheus:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.name=="NewAlertName")'

# 5. Test alert
# Trigger condition and verify in AlertManager
```

### 5.2 Dashboard Creation

**Purpose**: Create new Grafana dashboard

**Steps**:

```bash
# 1. Export existing dashboard as template
curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  http://grafana.phoenix.io/api/dashboards/uid/template > dashboard-template.json

# 2. Create new dashboard JSON
cat > new-dashboard.json <<EOF
{
  "dashboard": {
    "title": "Phoenix - New Feature",
    "panels": [
      {
        "title": "Metric Name",
        "targets": [
          {
            "expr": "rate(phoenix_metric_name[5m])"
          }
        ]
      }
    ]
  }
}
EOF

# 3. Import dashboard
curl -X POST -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -H "Content-Type: application/json" \
  -d @new-dashboard.json \
  http://grafana.phoenix.io/api/dashboards/db

# 4. Add to dashboard provisioning
kubectl edit configmap grafana-dashboards -n monitoring
```

## 6. Performance Troubleshooting

### 6.1 API Slow Response

**Symptoms**: P99 latency > 500ms

**Investigation Steps**:

```bash
# 1. Check current latency
curl -w "@curl-format.txt" -o /dev/null -s https://api.phoenix.io/health

# 2. Get slow query log
kubectl exec -n phoenix-system deployment/postgres -- \
    psql -U phoenix -d phoenix -c "SELECT query, mean_time, calls FROM pg_stat_statements WHERE mean_time > 100 ORDER BY mean_time DESC LIMIT 10;"

# 3. Check API resource usage
kubectl top pod -n phoenix-system -l app=phoenix-api

# 4. Profile the service
kubectl port-forward -n phoenix-system deployment/phoenix-api 6060:6060 &
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 5. Check cache hit rates
curl http://phoenix-api:8080/metrics | grep cache_hit

# 6. Analyze traces
# Query Jaeger for slow traces
curl "http://jaeger.phoenix.io/api/traces?service=phoenix-api&minDuration=500ms&limit=20"
```

### 6.2 High CPU Usage

**Investigation Steps**:

```bash
# 1. Identify high CPU pods
kubectl top pods -n phoenix-system --sort-by=cpu

# 2. Get CPU profile
POD=$(kubectl get pod -n phoenix-system -l app=phoenix-api -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n phoenix-system $POD -- kill -USR1 1
kubectl cp phoenix-system/$POD:/tmp/cpu.prof ./cpu.prof
go tool pprof -http=:8080 cpu.prof

# 3. Check goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1 | head -20

# 4. Analyze CPU metrics over time
cat > cpu-query.promql <<EOF
rate(container_cpu_usage_seconds_total{pod=~"phoenix-api-.*"}[5m])
EOF
```

## 7. Security Operations

### 7.1 Rotate Secrets

**Purpose**: Regular secret rotation

**Steps**:

```bash
#!/bin/bash
# rotate-secrets.sh

# 1. Generate new secrets
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 24)
API_KEY=$(uuidgen)

# 2. Update Kubernetes secrets
kubectl create secret generic phoenix-secrets \
  --from-literal=jwt-secret=$JWT_SECRET \
  --from-literal=db-password=$DB_PASSWORD \
  --from-literal=api-key=$API_KEY \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Update database password
kubectl exec -n phoenix-system deployment/postgres -- \
  psql -U postgres -c "ALTER USER phoenix PASSWORD '$DB_PASSWORD';"

# 4. Restart services to pick up new secrets
for deployment in api controller generator; do
    kubectl rollout restart deployment/$deployment -n phoenix-system
    kubectl rollout status deployment/$deployment -n phoenix-system
done

# 5. Verify services are healthy
./scripts/health-check.sh

# 6. Update external services (New Relic, etc)
# ... manual steps ...
```

### 7.2 Security Audit

**Purpose**: Regular security check

**Steps**:

```bash
# 1. Check for vulnerabilities in images
for image in $(kubectl get pods -n phoenix-system -o jsonpath='{.items[*].spec.containers[*].image}' | tr ' ' '\n' | sort -u); do
    echo "Scanning $image"
    trivy image $image
done

# 2. Check RBAC permissions
kubectl auth can-i --list --as=system:serviceaccount:phoenix-system:default

# 3. Review network policies
kubectl get networkpolicies -n phoenix-system

# 4. Check for exposed services
kubectl get services -A -o wide | grep -E "LoadBalancer|NodePort"

# 5. Audit secret access
kubectl logs -n kube-system -l component=kube-apiserver | grep -i secret | tail -50

# 6. Check pod security policies
kubectl get psp
kubectl get pods -n phoenix-system -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.securityContext}{"\n"}{end}'
```

## 8. Maintenance Procedures

### 8.1 Cluster Maintenance

**Purpose**: Kubernetes cluster updates

**Steps**:

```bash
# 1. Notify users
./scripts/notify-maintenance.sh "Cluster maintenance starting in 30 minutes"

# 2. Backup critical data
./scripts/backup-all.sh

# 3. Cordon nodes one by one
for node in $(kubectl get nodes -o name); do
    kubectl cordon $node
    kubectl drain $node --ignore-daemonsets --delete-emptydir-data
    # Perform maintenance on node
    kubectl uncordon $node
    sleep 300  # Wait 5 minutes between nodes
done

# 4. Verify cluster health
kubectl get nodes
kubectl get pods -A | grep -v Running

# 5. Run test suite
./scripts/run-integration-tests.sh

# 6. Notify completion
./scripts/notify-maintenance.sh "Cluster maintenance completed"
```

### 8.2 Certificate Renewal

**Purpose**: Renew TLS certificates

**Steps**:

```bash
# 1. Check certificate expiry
echo | openssl s_client -servername api.phoenix.io -connect api.phoenix.io:443 2>/dev/null | openssl x509 -noout -dates

# 2. Generate new certificate
certbot certonly --dns-cloudflare \
  --dns-cloudflare-credentials /etc/letsencrypt/cloudflare.ini \
  -d api.phoenix.io \
  -d "*.phoenix.io"

# 3. Update Kubernetes secret
kubectl create secret tls phoenix-tls \
  --cert=/etc/letsencrypt/live/phoenix.io/fullchain.pem \
  --key=/etc/letsencrypt/live/phoenix.io/privkey.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# 4. Restart ingress controller
kubectl rollout restart deployment/ingress-nginx-controller -n ingress-nginx

# 5. Verify new certificate
echo | openssl s_client -servername api.phoenix.io -connect api.phoenix.io:443 2>/dev/null | openssl x509 -noout -dates
```

## Quick Reference

### Important Commands

```bash
# Get all Phoenix resources
kubectl get all -n phoenix-system

# Tail logs from all API pods
kubectl logs -n phoenix-system -l app=phoenix-api -f --max-log-requests=10

# Execute SQL in production
kubectl exec -it -n phoenix-system deployment/postgres -- psql -U phoenix

# Port forward to service
kubectl port-forward -n phoenix-system svc/phoenix-api 8080:8080

# Get events sorted by time
kubectl get events -A --sort-by='.lastTimestamp'

# Check resource usage
kubectl top nodes
kubectl top pods -A --sort-by=memory
```

### Emergency Contacts

- On-Call Engineer: Check PagerDuty
- Platform Lead: platform-lead@phoenix.io
- Database Admin: dba-team@phoenix.io
- Security Team: security@phoenix.io
- AWS Support: +1-xxx-xxx-xxxx

### Useful Links

- [Monitoring Dashboard](https://grafana.phoenix.io)
- [Logs](https://logs.phoenix.io)
- [Traces](https://jaeger.phoenix.io)
- [Wiki](https://wiki.phoenix.io)
- [Status Page](https://status.phoenix.io)