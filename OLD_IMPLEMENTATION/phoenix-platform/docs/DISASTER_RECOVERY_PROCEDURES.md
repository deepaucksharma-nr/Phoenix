# Phoenix Platform Disaster Recovery Procedures

**Version**: 1.0  
**Last Updated**: January 25, 2025  
**Classification**: Confidential

## Overview

This document outlines comprehensive disaster recovery (DR) procedures for the Phoenix Platform, ensuring business continuity and minimal data loss in case of system failures.

## 1. Recovery Objectives

### 1.1 Key Metrics

| Metric | Target | Maximum |
|--------|--------|---------|
| **RTO** (Recovery Time Objective) | 2 hours | 4 hours |
| **RPO** (Recovery Point Objective) | 15 minutes | 30 minutes |
| **MTTR** (Mean Time To Recovery) | 1 hour | 2 hours |
| **Data Integrity** | 100% | 99.99% |

### 1.2 Service Priority Levels

| Priority | Services | RTO | RPO |
|----------|----------|-----|-----|
| **P0 - Critical** | API Gateway, Experiment Controller, PostgreSQL | 30 min | 5 min |
| **P1 - High** | Config Generator, Pipeline Operator | 1 hour | 15 min |
| **P2 - Medium** | Dashboard, Monitoring | 2 hours | 30 min |
| **P3 - Low** | Process Simulator, Analytics | 4 hours | 1 hour |

## 2. Disaster Scenarios

### 2.1 Regional Failure

**Scenario**: Complete AWS region failure (us-east-1)

**Recovery Procedure**:
```bash
#!/bin/bash
# DR-001: Regional Failover Procedure

# 1. Verify primary region failure
aws health describe-events --region us-east-1 || PRIMARY_DOWN=true

# 2. Activate DR region (us-west-2)
if [ "$PRIMARY_DOWN" = true ]; then
    # Update Route53 weighted routing
    aws route53 change-resource-record-sets \
        --hosted-zone-id $ZONE_ID \
        --change-batch file://dr-routing-policy.json
    
    # Scale up DR region
    kubectl --context=eks-us-west-2 scale deployment phoenix-api --replicas=5
    
    # Restore latest database backup
    ./scripts/restore-database.sh --region us-west-2 --latest
fi
```

### 2.2 Database Corruption

**Scenario**: PostgreSQL data corruption detected

**Recovery Procedure**:
```sql
-- DR-002: Database Recovery Procedure

-- 1. Stop write traffic
ALTER DATABASE phoenix SET default_transaction_read_only = true;

-- 2. Identify corruption extent
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 3. Restore from backup
-- Use point-in-time recovery if available
```

```bash
# Restore procedure
pg_restore \
    --host=phoenix-dr.region.rds.amazonaws.com \
    --port=5432 \
    --username=phoenix \
    --dbname=phoenix_restore \
    --verbose \
    --no-owner \
    --no-privileges \
    /backups/phoenix_$(date -d "1 hour ago" +%Y%m%d_%H0000).dump

# Verify restoration
psql -h phoenix-dr.region.rds.amazonaws.com -U phoenix -d phoenix_restore \
    -c "SELECT COUNT(*) FROM experiments;"
```

### 2.3 Kubernetes Cluster Failure

**Scenario**: Complete EKS cluster failure

**Recovery Procedure**:
```yaml
# DR-003: Kubernetes Cluster Recovery
apiVersion: v1
kind: ConfigMap
metadata:
  name: dr-cluster-recovery
data:
  recovery.sh: |
    #!/bin/bash
    set -e
    
    # 1. Create new cluster
    eksctl create cluster -f cluster-dr.yaml
    
    # 2. Install cluster essentials
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/aws/deploy.yaml
    
    # 3. Restore Helm releases
    helm install phoenix ./helm/phoenix \
        --namespace phoenix-system \
        --create-namespace \
        --values values-dr.yaml
    
    # 4. Restore ConfigMaps and Secrets
    kubectl apply -f backups/configs/
    kubectl apply -f backups/secrets/
    
    # 5. Verify deployments
    kubectl get deployments -A
    kubectl get pods -A | grep -v Running
```

### 2.4 Data Center Network Partition

**Scenario**: Network split between availability zones

**Recovery Procedure**:
```go
// DR-004: Network Partition Handler
package dr

func HandleNetworkPartition(ctx context.Context) error {
    // 1. Detect partition
    if err := detectPartition(); err != nil {
        return fmt.Errorf("partition detection failed: %w", err)
    }
    
    // 2. Enter degraded mode
    config.SetDegradedMode(true)
    
    // 3. Fence minority partition
    if isMinorityPartition() {
        // Stop accepting writes
        db.SetReadOnly(true)
        
        // Redirect traffic to majority
        return redirectToMajority()
    }
    
    // 4. Continue in majority partition
    return continueMajorityOperations()
}
```

## 3. Backup Strategies

### 3.1 Database Backups

```yaml
# Automated backup configuration
backup_config:
  postgresql:
    schedule:
      full: "0 2 * * *"  # Daily at 2 AM
      incremental: "*/15 * * * *"  # Every 15 minutes
      wal_archive: continuous
    retention:
      daily: 7
      weekly: 4
      monthly: 12
    encryption: AES-256
    storage:
      primary: s3://phoenix-backups-primary/postgres/
      secondary: s3://phoenix-backups-dr/postgres/
```

### 3.2 Application State Backups

```bash
#!/bin/bash
# backup-application-state.sh

# Backup Kubernetes resources
kubectl get all,cm,secret -A -o yaml > k8s-resources-$(date +%Y%m%d-%H%M%S).yaml

# Backup etcd
ETCDCTL_API=3 etcdctl snapshot save etcd-snapshot-$(date +%Y%m%d-%H%M%S).db \
  --endpoints=https://etcd-cluster:2379 \
  --cacert=/etc/etcd/ca.crt \
  --cert=/etc/etcd/client.crt \
  --key=/etc/etcd/client.key

# Backup Git configurations
git clone https://github.com/phoenix/configs.git
tar czf git-configs-$(date +%Y%m%d-%H%M%S).tar.gz configs/

# Upload to S3
aws s3 sync ./backups s3://phoenix-backups-primary/state/
```

### 3.3 Metrics Data Backup

```yaml
# Prometheus backup job
apiVersion: batch/v1
kind: CronJob
metadata:
  name: prometheus-backup
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: prom/prometheus:latest
            command:
            - /bin/sh
            - -c
            - |
              # Create snapshot
              curl -XPOST http://prometheus:9090/api/v1/admin/tsdb/snapshot
              
              # Upload to S3
              aws s3 sync /prometheus/snapshots/ s3://phoenix-backups/prometheus/
```

## 4. Recovery Procedures

### 4.1 Full System Recovery

```bash
#!/bin/bash
# DR-005: Full System Recovery

# Phase 1: Infrastructure
echo "Phase 1: Restoring Infrastructure"
terraform apply -var-file=dr.tfvars -auto-approve

# Phase 2: Kubernetes
echo "Phase 2: Restoring Kubernetes"
eksctl create cluster -f cluster-dr.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/aws/deploy.yaml

# Phase 3: Database
echo "Phase 3: Restoring Database"
./scripts/restore-database.sh --latest --verify

# Phase 4: Applications
echo "Phase 4: Deploying Applications"
helm install phoenix ./helm/phoenix -f values-dr.yaml

# Phase 5: Verification
echo "Phase 5: Running Verification"
./scripts/dr-verification.sh
```

### 4.2 Partial Recovery

```go
// Service-specific recovery
package recovery

type ServiceRecovery struct {
    service  string
    priority int
    steps    []RecoveryStep
}

func (r *ServiceRecovery) Execute(ctx context.Context) error {
    log.Printf("Starting recovery for service: %s", r.service)
    
    for i, step := range r.steps {
        log.Printf("Executing step %d: %s", i+1, step.Name)
        
        if err := step.Execute(ctx); err != nil {
            // Attempt rollback
            if rbErr := r.rollback(ctx, i); rbErr != nil {
                return fmt.Errorf("recovery failed and rollback failed: %v, %v", err, rbErr)
            }
            return fmt.Errorf("recovery failed at step %s: %w", step.Name, err)
        }
        
        // Verify step completion
        if err := step.Verify(ctx); err != nil {
            return fmt.Errorf("verification failed for step %s: %w", step.Name, err)
        }
    }
    
    return nil
}
```

## 5. Data Recovery

### 5.1 Point-in-Time Recovery

```sql
-- PostgreSQL PITR
-- Restore to specific transaction
SELECT pg_create_restore_point('before_critical_update');

-- If recovery needed
-- 1. Stop PostgreSQL
-- 2. Clear data directory
-- 3. Restore base backup
-- 4. Configure recovery.conf
```

```conf
# recovery.conf
restore_command = 'aws s3 cp s3://phoenix-wal-archive/%f %p'
recovery_target_name = 'before_critical_update'
recovery_target_inclusive = true
recovery_target_action = 'promote'
```

### 5.2 Experiment Data Recovery

```go
// Reconstruct experiment state from events
func RecoverExperimentState(ctx context.Context, experimentID string) (*Experiment, error) {
    // 1. Load events from event store
    events, err := eventStore.LoadEvents(ctx, experimentID)
    if err != nil {
        return nil, fmt.Errorf("failed to load events: %w", err)
    }
    
    // 2. Replay events to rebuild state
    experiment := &Experiment{ID: experimentID}
    for _, event := range events {
        if err := experiment.Apply(event); err != nil {
            log.Printf("Failed to apply event %s: %v", event.ID, err)
            continue // Skip corrupted events
        }
    }
    
    // 3. Validate reconstructed state
    if err := experiment.Validate(); err != nil {
        return nil, fmt.Errorf("invalid experiment state: %w", err)
    }
    
    // 4. Save to primary store
    if err := store.SaveExperiment(ctx, experiment); err != nil {
        return nil, fmt.Errorf("failed to save experiment: %w", err)
    }
    
    return experiment, nil
}
```

## 6. Communication Procedures

### 6.1 Incident Communication Matrix

| Severity | Internal | External | Escalation |
|----------|----------|----------|------------|
| **SEV1** | Immediate | Within 15 min | VP Engineering |
| **SEV2** | Within 30 min | Within 1 hour | Engineering Manager |
| **SEV3** | Within 1 hour | Within 4 hours | Team Lead |
| **SEV4** | Within 4 hours | Next business day | On-call Engineer |

### 6.2 Status Page Updates

```yaml
# Status page update template
incident:
  title: "Service Degradation - Experiment Creation"
  severity: major
  affected_components:
    - API Service
    - Experiment Controller
  updates:
    - timestamp: "2025-01-25T10:00:00Z"
      status: investigating
      message: "We are investigating issues with experiment creation"
    - timestamp: "2025-01-25T10:30:00Z"
      status: identified
      message: "Database connection pool exhaustion identified"
    - timestamp: "2025-01-25T11:00:00Z"
      status: monitoring
      message: "Fix deployed, monitoring recovery"
    - timestamp: "2025-01-25T12:00:00Z"
      status: resolved
      message: "Service fully restored"
```

## 7. Testing and Drills

### 7.1 DR Test Schedule

| Test Type | Frequency | Duration | Scope |
|-----------|-----------|----------|-------|
| **Backup Verification** | Daily | 30 min | Automated |
| **Component Failover** | Weekly | 2 hours | Single service |
| **Regional Failover** | Monthly | 4 hours | Full DR site |
| **Full DR Simulation** | Quarterly | 8 hours | Complete system |

### 7.2 DR Test Scenarios

```bash
#!/bin/bash
# dr-test-runner.sh

# Test 1: Database failover
echo "Test 1: Database Failover"
./tests/database-failover.sh
sleep 300
./tests/verify-database-recovery.sh

# Test 2: Service failure
echo "Test 2: Service Failure Recovery"
kubectl delete deployment phoenix-api -n phoenix-system
sleep 60
./tests/verify-service-recovery.sh

# Test 3: Network partition
echo "Test 3: Network Partition"
./tests/simulate-network-partition.sh
sleep 120
./tests/verify-partition-recovery.sh

# Generate report
./tests/generate-dr-report.sh > dr-test-$(date +%Y%m%d).report
```

## 8. Recovery Verification

### 8.1 Automated Verification

```go
// Verification suite
package verification

type VerificationSuite struct {
    checks []HealthCheck
}

func (v *VerificationSuite) RunAll(ctx context.Context) (*Report, error) {
    report := &Report{
        StartTime: time.Now(),
        Checks:    make([]CheckResult, 0, len(v.checks)),
    }
    
    for _, check := range v.checks {
        result := CheckResult{
            Name:      check.Name,
            StartTime: time.Now(),
        }
        
        err := check.Run(ctx)
        result.EndTime = time.Now()
        result.Duration = result.EndTime.Sub(result.StartTime)
        
        if err != nil {
            result.Status = "FAILED"
            result.Error = err.Error()
        } else {
            result.Status = "PASSED"
        }
        
        report.Checks = append(report.Checks, result)
    }
    
    report.EndTime = time.Now()
    report.OverallStatus = v.determineOverallStatus(report.Checks)
    
    return report, nil
}
```

### 8.2 Manual Verification Checklist

- [ ] All services responding to health checks
- [ ] Database connections established
- [ ] No data loss detected
- [ ] Performance within acceptable range
- [ ] All experiments resumed correctly
- [ ] Monitoring and alerting functional
- [ ] Client connections restored
- [ ] No duplicate data created

## 9. Post-Recovery Actions

### 9.1 Immediate Actions

1. **Verify System Stability** (0-2 hours)
   - Monitor error rates
   - Check performance metrics
   - Validate data integrity

2. **Communicate Status** (0-4 hours)
   - Update status page
   - Send customer notifications
   - Brief executive team

3. **Document Timeline** (0-24 hours)
   - Record all actions taken
   - Note decision points
   - Capture lessons learned

### 9.2 Follow-up Actions

1. **Root Cause Analysis** (24-48 hours)
2. **Post-Mortem Meeting** (48-72 hours)
3. **Update DR Procedures** (1 week)
4. **Implement Improvements** (2-4 weeks)

## 10. DR Automation

### 10.1 Automated Failover

```yaml
# Kubernetes operator for automated DR
apiVersion: dr.phoenix.io/v1alpha1
kind: DisasterRecovery
metadata:
  name: phoenix-dr
spec:
  primaryRegion: us-east-1
  drRegion: us-west-2
  
  triggers:
    - type: HealthCheck
      threshold: 3
      interval: 30s
    - type: ErrorRate
      threshold: 50
      window: 5m
      
  actions:
    - type: DatabaseFailover
      priority: 1
    - type: TrafficRedirect
      priority: 2
    - type: ServiceScaling
      priority: 3
      
  notifications:
    - type: PagerDuty
      severity: P1
    - type: Slack
      channel: "#incidents"
```

## 11. Compliance and Audit

### 11.1 DR Compliance Requirements

- **SOC2**: Annual DR testing required
- **ISO 27001**: Documented procedures and test results
- **GDPR**: Data recovery within 72 hours
- **HIPAA**: Encryption of backups (if applicable)

### 11.2 Audit Trail

All DR activities must be logged:
```json
{
  "event": "dr_activation",
  "timestamp": "2025-01-25T10:00:00Z",
  "triggered_by": "automated",
  "reason": "region_failure",
  "actions": [
    "database_failover",
    "traffic_redirect",
    "notification_sent"
  ],
  "outcome": "successful",
  "duration_minutes": 45
}
```

## 12. Contact Information

### Emergency Contacts

| Role | Name | Phone | Email |
|------|------|-------|-------|
| Incident Commander | On-call rotation | +1-xxx-xxx-xxxx | incidents@phoenix.io |
| VP Engineering | John Doe | +1-xxx-xxx-xxxx | john.doe@phoenix.io |
| Database Admin | Jane Smith | +1-xxx-xxx-xxxx | jane.smith@phoenix.io |
| AWS TAM | Mike Johnson | +1-xxx-xxx-xxxx | mike@aws.com |

### Vendor Support

- **AWS Support**: +1-xxx-xxx-xxxx (Enterprise Support)
- **New Relic Support**: +1-xxx-xxx-xxxx
- **PagerDuty**: +1-xxx-xxx-xxxx