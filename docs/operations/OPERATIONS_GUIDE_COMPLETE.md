
# Phoenix Operations Guide

Complete operational guide for the Phoenix Platform's 70% cost reduction observability system. Covers deployment, monitoring, and troubleshooting of the agent-based architecture.

## Table of Contents

- [Platform Overview](#platform-overview)
- [Agent Deployment](#agent-deployment)
- [Task Management](#task-management)
- [Monitoring & Alerting](#monitoring--alerting)
- [Troubleshooting](#troubleshooting)
- [Cost Optimization](#cost-optimization)

## Platform Overview

Phoenix Platform reduces observability costs by 70% through:
- **Agent-based task polling** (30-second intervals)
- **A/B testing** for safe pipeline rollouts
- **Real-time WebSocket** monitoring
- **PostgreSQL task queue** for coordination

## Agent Deployment

### Prerequisites
- Phoenix API running on port 8080
- PostgreSQL database accessible
- Network connectivity to control plane

### Agent Installation
```bash
# Download and install agent
curl -fsSL https://phoenix.example.com/install-agent.sh | sudo bash

# Configure agent
sudo tee /etc/phoenix-agent/config.yaml <<EOF
api_url: https://phoenix.example.com:8080
host_id: $(hostname)-$(date +%s)
poll_interval: 30s
log_level: info
EOF

# Start agent service
sudo systemctl enable --now phoenix-agent
```

### Agent Authentication
Agents authenticate using X-Agent-Host-ID header:
```bash
curl -H "X-Agent-Host-ID: my-agent-001" \
     http://phoenix.example.com:8080/api/v1/agent/tasks
```

## Task Management

### Task Queue Architecture
Phoenix uses PostgreSQL-based task queue with long-polling:

```sql
-- Check pending tasks
SELECT id, host_id, type, action, status, created_at 
FROM tasks 
WHERE status = 'pending' 
ORDER BY priority DESC, created_at ASC;

-- Agent task polling
SELECT * FROM tasks 
WHERE host_id = $1 AND status = 'pending' 
FOR UPDATE SKIP LOCKED 
LIMIT 10;
```

### Task Types
- **deployment**: Pipeline deployment/update
- **metrics**: Metrics collection configuration
- **experiment**: A/B test execution
- **health**: Health check and status report

### Task Status Flow
```
pending → assigned → running → completed
                  ↓
                failed
```

## Monitoring & Alerting

### Key Metrics
```prometheus
# Cost reduction percentage
phoenix_cardinality_reduction_percent{experiment_id="exp-123"}

# Agent health
phoenix_agent_last_seen_seconds{host_id="agent-001"}

# Task queue depth
phoenix_task_queue_depth{status="pending"}

# Experiment status
phoenix_experiment_phase{experiment_id="exp-123"}
```

### Alert Rules
```yaml
groups:
- name: phoenix.rules
  rules:
  - alert: AgentDown
    expr: time() - phoenix_agent_last_seen_seconds > 300
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Phoenix agent {{ $labels.host_id }} is down"

  - alert: TaskQueueBackup
    expr: phoenix_task_queue_depth{status="pending"} > 100
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "Phoenix task queue backing up: {{ $value }} pending tasks"
```

## Troubleshooting

### Agent Not Appearing in Fleet
```bash
# Check agent logs
sudo journalctl -u phoenix-agent -f

# Verify connectivity
curl -H "X-Agent-Host-ID: test" http://phoenix-api:8080/api/v1/agent/tasks

# Check firewall rules
sudo iptables -L | grep 8080
```

### High Task Queue Depth
```bash
# Check database connections
SELECT count(*) FROM pg_stat_activity WHERE datname='phoenix';

# Monitor task processing
SELECT status, count(*) FROM tasks GROUP BY status;

# Restart task workers
docker-compose restart phoenix-api
```

### Experiment Stuck in Running State
```bash
# Check experiment status
curl http://phoenix-api:8080/api/v1/experiments/exp-123

# Review agent task completion
SELECT * FROM tasks WHERE experiment_id = 'exp-123' ORDER BY created_at DESC;

# Force experiment completion
curl -X POST http://phoenix-api:8080/api/v1/experiments/exp-123/stop
```

## Cost Optimization

### Monitoring Savings
Phoenix tracks cost reduction in real-time:

```bash
# Current cost flow
curl http://phoenix-api:8080/api/v1/cost-flow

# Experiment analysis
curl http://phoenix-api:8080/api/v1/experiments/exp-123/analysis
```

### Optimization Strategies
1. **Adaptive Filter**: ML-based metric importance
2. **TopK Sampling**: Keep only top K metrics
3. **Hybrid Approach**: Combination of strategies

### Typical Results
- **Before Phoenix**: $50,000/month
- **After Phoenix**: $15,000/month (70% reduction)
- **Annual Savings**: $420,000

## Systemd Service

Phoenix Agent service configuration `/etc/systemd/system/phoenix-agent.service`:

```ini
[Unit]
Description=Phoenix Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/otelcol --config /etc/otelcol/collector.yaml
Restart=always
User=otel
Group=otel

[Install]
WantedBy=multi-user.target
```

Enable and start the collector:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now otelcol
```

The collector will now run as a background service on the VM.
=======
# Phoenix Platform Operations Guide

This guide describes how to deploy pipelines, run experiments, and analyze results using the Phoenix Platform.

## 1. Deployment Workflow

1. **Bootstrap dependencies**
   ```bash
   make dev-up
   ```
2. **Deploy a pipeline**
   ```bash
   curl -X POST http://localhost:8080/api/v1/pipeline-deployments \
     -H "Content-Type: application/json" \
     -d '{"name":"demo","namespace":"default","template":"process-baseline-v1"}'
   ```
3. **Verify deployment**
   ```bash
   curl http://localhost:8080/api/v1/pipeline-deployments?namespace=default | jq .
   ```

## 2. Experiment Workflow

1. **Create an experiment**
   ```bash
   curl -X POST http://localhost:8080/api/v1/experiments \
     -H "Content-Type: application/json" \
     -d '{"name":"cost-opt","baseline_pipeline":"process-baseline-v1","candidate_pipeline":"process-intelligent-v1","target_namespaces":["default"]}'
   ```
2. **Monitor progress**
   ```bash
   curl http://localhost:8080/api/v1/experiments/<id> | jq .
   ```
3. **Generate configs**
   ```bash
   curl -X POST http://localhost:8082/api/v1/generate \
     -H "Content-Type: application/json" \
     -d '{"experiment_id":"<id>"}'
   ```
4. **Analyze results**
   ```bash
   curl http://localhost:8080/api/v1/experiments/<id>/results | jq .
   ```

## 3. Troubleshooting

- **Check service health**
  ```bash
  curl http://localhost:8080/health
  ```
- **Restart stack**
  ```bash
  make dev-down && make dev-up
  ```

