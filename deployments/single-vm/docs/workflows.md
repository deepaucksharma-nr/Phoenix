# Phoenix Single-VM Workflows

This guide provides step-by-step instructions for common Phoenix operations.

## Table of Contents
1. [Initial Setup](#initial-setup)
2. [Agent Management](#agent-management)
3. [Pipeline Operations](#pipeline-operations)
4. [Experiment Workflows](#experiment-workflows)
5. [Cost Optimization](#cost-optimization)
6. [Monitoring & Troubleshooting](#monitoring--troubleshooting)
7. [Maintenance Tasks](#maintenance-tasks)

## Initial Setup

### Deploy Phoenix Control Plane
```bash
# 1. Clone repository
git clone https://github.com/phoenix-observability/phoenix
cd phoenix/deployments/single-vm

# 2. Run automated setup
sudo ./scripts/setup-single-vm.sh

# 3. Save credentials from output
# JWT Secret: xxx
# Agent Token: yyy
# Grafana Password: zzz
```

### Verify Installation
```bash
# Check all services
./scripts/health-check.sh

# View service logs
cd /opt/phoenix
sudo docker-compose logs -f

# Access UI
# https://phoenix.your-domain.com
```

## Agent Management

### Install Agent on New Host
```bash
# On the edge node:
curl -fsSL https://phoenix.your-domain.com/install-agent.sh | sudo bash

# Or with custom token:
export PHOENIX_TOKEN="your-token"
curl -fsSL https://phoenix.your-domain.com/install-agent.sh | sudo bash
```

### Check Agent Status
```bash
# On agent host:
sudo systemctl status phoenix-agent
sudo journalctl -u phoenix-agent -f

# From control plane:
curl -k https://phoenix.your-domain.com/api/v1/agents
```

### Update Agent
```bash
# On agent host:
sudo systemctl stop phoenix-agent
curl -L https://phoenix.your-domain.com/downloads/phoenix-agent-linux-amd64 \
  -o /opt/phoenix-agent/phoenix-agent
sudo chmod +x /opt/phoenix-agent/phoenix-agent
sudo systemctl start phoenix-agent
```

### Remove Agent
```bash
# On agent host:
sudo /opt/phoenix-agent/uninstall.sh
```

## Pipeline Operations

### Deploy Pipeline to All Hosts
```bash
# Using CLI
phoenix pipeline deploy process-topk-v1 --all

# Using API
curl -X POST https://phoenix.your-domain.com/api/v1/pipelines/deploy \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline": "process-topk-v1",
    "selector": {"all": true}
  }'
```

### Deploy to Specific Hosts
```bash
# By tags
phoenix pipeline deploy process-priority-filter \
  --selector env=prod,service=api

# By host IDs
phoenix pipeline deploy adaptive-sampling-v2 \
  --hosts host-001,host-002,host-003
```

### List Deployed Pipelines
```bash
# All deployments
phoenix pipeline list

# For specific host
phoenix pipeline list --host host-001

# Active deployments only
phoenix pipeline list --active
```

### Rollback Pipeline
```bash
# Rollback specific host
phoenix pipeline rollback --host host-001

# Rollback by tag
phoenix pipeline rollback --selector env=prod

# Rollback all hosts
phoenix pipeline rollback --all
```

## Experiment Workflows

### Create A/B Test Experiment
```bash
# 1. Create experiment
phoenix experiment create \
  --name "test-topk-optimization" \
  --description "Test Top-K filter on API servers" \
  --baseline process-baseline \
  --candidate process-topk-v1 \
  --selector service=api \
  --duration 1h \
  --traffic-split 50

# 2. Monitor in UI
# https://phoenix.your-domain.com/experiments

# 3. Check metrics
phoenix experiment metrics test-topk-optimization

# 4. Promote or rollback
phoenix experiment promote test-topk-optimization
# or
phoenix experiment rollback test-topk-optimization
```

### Progressive Rollout
```bash
# Start with 10% traffic
phoenix experiment create \
  --name "progressive-rollout" \
  --candidate new-filter \
  --traffic-split 10 \
  --duration 30m

# Increase to 50%
phoenix experiment update progressive-rollout \
  --traffic-split 50

# Full rollout
phoenix experiment promote progressive-rollout
```

### Emergency Rollback
```bash
# Immediate rollback
phoenix experiment rollback --emergency experiment-id

# Rollback all experiments
phoenix experiment rollback --all --emergency
```

## Cost Optimization

### Analyze Current Costs
```bash
# Get cost breakdown
phoenix cost analyze

# By service
phoenix cost analyze --group-by service

# By metric pattern
phoenix cost analyze --pattern "kubernetes.*"
```

### Find High-Cardinality Metrics
```bash
# Top 10 expensive metrics
phoenix metrics top --limit 10

# Search by pattern
phoenix metrics search "pod.cpu" --show-cost

# Export analysis
phoenix metrics analyze --export costs.csv
```

### Apply Quick Optimizations
```bash
# Deploy standard filters
# 1. Remove Kubernetes metrics (save ~40%)
phoenix pipeline deploy filter-kubernetes --all

# 2. Keep only critical metrics (save ~70%)
phoenix pipeline deploy priority-critical --all

# 3. Sample non-critical metrics (save ~50%)
phoenix pipeline deploy adaptive-sampling --all
```

### Create Custom Filter
```yaml
# custom-filter.yaml
apiVersion: v1
kind: ProcessorPipeline
metadata:
  name: custom-cost-saver
spec:
  processors:
    - type: filter
      config:
        exclude:
          - "*.debug.*"
          - "test.*"
          - "staging.*"
    - type: topk
      config:
        k: 20
        by: ["service", "endpoint"]
```

```bash
# Apply custom filter
phoenix pipeline create -f custom-filter.yaml
phoenix pipeline deploy custom-cost-saver --selector env=prod
```

## Monitoring & Troubleshooting

### Check System Health
```bash
# Full health check
/opt/phoenix/scripts/health-check.sh

# Quick status
docker-compose ps

# Service metrics
curl http://localhost:8080/metrics
```

### View Real-Time Metrics
```bash
# Agent metrics in Pushgateway
curl http://localhost:9091/metrics | grep phoenix_

# Query Prometheus
curl 'http://localhost:9090/api/v1/query?query=phoenix_agent_status'

# Cost savings
curl 'http://localhost:9090/api/v1/query?query=phoenix_cost_savings_total'
```

### Debug Agent Issues
```bash
# On agent host:
# Check logs
sudo journalctl -u phoenix-agent -n 100

# Test connectivity
curl -k https://phoenix.your-domain.com/health

# Check collector processes
ps aux | grep otelcol

# Force task poll
sudo systemctl restart phoenix-agent
```

### Debug Pipeline Issues
```bash
# Check pipeline status
phoenix pipeline status pipeline-name

# View pipeline logs
phoenix pipeline logs pipeline-name --tail 100

# Validate configuration
phoenix pipeline validate -f pipeline.yaml
```

## Maintenance Tasks

### Daily Tasks
```bash
# Check agent health
/opt/phoenix/scripts/health-check.sh

# Review cost savings
phoenix cost summary --today

# Check for alerts
docker-compose logs api | grep ERROR | tail -20
```

### Weekly Tasks
```bash
# Backup data
sudo /opt/phoenix/scripts/backup.sh

# Update agents
phoenix agent update --all

# Review metrics retention
phoenix metrics cleanup --dry-run
```

### Monthly Tasks
```bash
# Update Phoenix
cd /opt/phoenix
sudo docker-compose pull
sudo docker-compose up -d

# Optimize database
docker-compose exec db psql -U phoenix -c "VACUUM ANALYZE;"

# Review and rotate logs
find /opt/phoenix/data/logs -name "*.log" -mtime +30 -delete

# Test disaster recovery
# 1. Backup
sudo /opt/phoenix/scripts/backup.sh
# 2. Restore to test environment
sudo /opt/phoenix/scripts/restore.sh latest
```

### Certificate Renewal
```bash
# Check certificate expiry
openssl x509 -in /opt/phoenix/tls/fullchain.pem -noout -enddate

# Renew Let's Encrypt
sudo certbot renew

# Restart API to load new cert
cd /opt/phoenix
sudo docker-compose restart api
```

## Advanced Workflows

### Multi-Stage Experiments
```bash
# Stage 1: Test on canary hosts
phoenix experiment create \
  --name "multi-stage-test" \
  --candidate new-processor \
  --selector role=canary \
  --duration 1h

# Stage 2: Expand to 10% of production
phoenix experiment update multi-stage-test \
  --selector "role=canary OR (env=prod AND random<0.1)" \
  --duration 2h

# Stage 3: Full production
phoenix experiment promote multi-stage-test
```

### Scheduled Pipeline Changes
```bash
# Schedule off-peak optimization
cat > /etc/cron.d/phoenix-schedule << EOF
# Heavy filtering during business hours
0 9 * * 1-5 phoenix pipeline deploy aggressive-filter --all
# Relaxed filtering after hours
0 18 * * 1-5 phoenix pipeline deploy standard-filter --all
EOF
```

### Integration with CI/CD
```yaml
# .gitlab-ci.yml example
deploy_phoenix_pipeline:
  stage: deploy
  script:
    - |
      phoenix pipeline validate -f $CI_PROJECT_DIR/pipelines/prod.yaml
      phoenix pipeline create -f $CI_PROJECT_DIR/pipelines/prod.yaml
      phoenix experiment create \
        --name "deploy-$CI_COMMIT_SHA" \
        --candidate "pipeline-$CI_COMMIT_SHA" \
        --duration 30m
  only:
    - main
```

## Troubleshooting Quick Reference

| Issue | Check | Fix |
|-------|-------|-----|
| Agent offline | `systemctl status phoenix-agent` | `systemctl restart phoenix-agent` |
| No metrics | `curl pushgateway:9091/metrics` | Check agent connectivity |
| High memory | `docker stats` | Increase container limits |
| Slow queries | API logs for timing | Add database indexes |
| Cert expired | `openssl x509 -enddate` | `certbot renew` |

## Best Practices

1. **Start Small**: Test pipelines on a few hosts before deploying widely
2. **Monitor Impact**: Always check SLIs after deploying filters
3. **Use Experiments**: A/B test changes instead of direct deployment
4. **Regular Backups**: Automate daily backups and test restores
5. **Tag Everything**: Use consistent tags for easier management
6. **Document Changes**: Keep notes on what optimizations work

## Getting Help

```bash
# Built-in help
phoenix --help
phoenix experiment --help

# API documentation
curl https://phoenix.your-domain.com/api/docs

# Support
# - GitHub: https://github.com/phoenix-observability/phoenix/issues
# - Docs: https://docs.phoenix-observability.io
# - Slack: https://phoenix-community.slack.com
```