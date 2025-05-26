# Phoenix Lean-Core Architecture

This document describes the lean-core architecture implementation for Phoenix Platform, which simplifies the system from multiple microservices to a consolidated API + lightweight agents model.

## Overview

The lean architecture replaces the complex Kubernetes-native microservice architecture with:
- **1 Phoenix API** (monolithic control plane)
- **N Phoenix Agents** (lightweight data plane agents)
- **Standard monitoring stack** (Prometheus + Pushgateway)

## Architecture Components

### Phoenix API (`projects/phoenix-api`)
Consolidates functionality from:
- Controller Service → Experiment state machine
- Platform API → REST endpoints & WebSocket
- Benchmark Service → KPI calculation module
- Analytics Service → Analysis module
- Pipeline Operator → Task queue system
- LoadSim Operator → Load profile tasks

Key features:
- Single PostgreSQL database with all state
- Task queue for agent work distribution
- Long-polling agent endpoints
- WebSocket for real-time updates
- Integrated KPI calculations

### Phoenix Agent (`projects/phoenix-agent`)
Lightweight agent that runs on each host:
- Polls API for tasks via long-polling
- Manages OTel collector processes
- Executes load simulations
- Reports metrics and status
- Self-contained with minimal dependencies

## Database Schema

New tables for lean architecture:
```sql
-- Task queue for agents
agent_tasks (
  id, host_id, experiment_id, type, action, 
  config, priority, status, timestamps...
)

-- Agent heartbeat tracking
agent_status (
  host_id, hostname, agent_version, status,
  capabilities, active_tasks, resource_usage...
)

-- Active pipeline tracking
active_pipelines (
  id, host_id, experiment_id, variant,
  config_url, status, process_info...
)

-- Metrics cache for queries
metrics_cache (
  id, experiment_id, timestamp, metric_name,
  variant, host_id, value, labels...
)
```

## Deployment

### Kubernetes
```bash
# Deploy the lean stack
kubectl apply -f deployments/kubernetes/lean-architecture/

# Components deployed:
# - Phoenix API (Deployment, 2 replicas)
# - Phoenix Agent (DaemonSet)
# - PostgreSQL (StatefulSet)
# - Prometheus + Pushgateway
```

### VM/Bare Metal
```bash
# Install agent
cd projects/phoenix-agent
make install
sudo systemctl start phoenix-agent

# Configure API endpoint
export PHOENIX_API_URL=http://phoenix-api:8080
```

## Configuration Templates

OTel configurations moved from custom processors to standard configs:

### Baseline (No filtering)
- `configs/otel-templates/baseline/config.yaml`
- Collects all metrics without filtering

### Candidate (Top-K approximation)
- `configs/otel-templates/candidate/topk-config.yaml`
- Uses transform + Lua for top contributor filtering
- Probabilistic sampling for cardinality reduction

### Candidate (Adaptive filtering)
- `configs/otel-templates/candidate/adaptive-filter-config.yaml`
- Dynamic threshold-based filtering
- Prioritizes anomalies and errors

## API Changes

### New Agent Endpoints
```
GET  /api/v1/agent/tasks          # Long-poll for tasks (30s timeout)
POST /api/v1/agent/tasks/{id}/status  # Update task status
POST /api/v1/agent/heartbeat      # Agent health check
POST /api/v1/agent/metrics        # Push metrics batch
POST /api/v1/agent/logs           # Stream logs
```

### Experiment Flow
1. Create experiment via API
2. API creates tasks in queue
3. Agents poll and pick up tasks
4. Agents start OTel collectors
5. Metrics flow: Collectors → Pushgateway → Prometheus
6. API queries Prometheus for KPIs
7. Results available via REST/WebSocket

## Migration from K8s-Native

### Feature Flags
Control migration with environment variables:
```bash
FEATURE_LEAN_AGENTS=true      # Use agent architecture
FEATURE_K8S_MODE=false        # Disable K8s operators
FEATURE_PUSHGATEWAY=true      # Use Pushgateway for metrics
```

### Parallel Testing
1. Deploy lean architecture alongside existing
2. Mirror experiments to both systems
3. Compare results
4. Gradual cutover via feature flags

## Benefits

1. **Simplified Operations**
   - Single API to manage (vs 5+ services)
   - Standard OTel configs (no custom processors)
   - Fewer moving parts

2. **Better Debugging**
   - All logic in one place
   - Simple task queue visibility
   - Direct database queries

3. **Cross-Platform**
   - Works on K8s, VMs, bare metal
   - No CRD dependencies
   - Standard HTTP/WebSocket APIs

4. **Reduced Resource Usage**
   - ~60% less code
   - 50% fewer containers
   - Lower memory footprint

## Development

### Running Locally
```bash
# Start API
cd projects/phoenix-api
make dev

# Start Agent
cd projects/phoenix-agent
make run

# Run integration tests
make test-integration
```

### Building
```bash
# Build all lean components
make build-phoenix-api build-phoenix-agent

# Build Docker images
make docker-phoenix-api docker-phoenix-agent
```

## Monitoring

Grafana dashboards available:
- Phoenix Overview (experiments, KPIs)
- Agent Status (health, tasks)
- Resource Usage (CPU, memory)
- Metrics Flow (ingestion rates)

## Troubleshooting

### Agent Issues
```bash
# Check agent logs
journalctl -u phoenix-agent -f

# Debug mode
phoenix-agent -log-level=debug

# Test connectivity
curl -H "X-Agent-Host-ID: test" http://api:8080/api/v1/agent/tasks
```

### Task Queue
```sql
-- Check pending tasks
SELECT * FROM agent_tasks WHERE status = 'pending';

-- Check agent status
SELECT * FROM agent_status WHERE last_heartbeat < NOW() - INTERVAL '1 minute';

-- Task execution history
SELECT host_id, status, count(*) 
FROM agent_tasks 
GROUP BY host_id, status;
```

## Future Enhancements

1. **Multi-region support**
   - Region-aware task routing
   - Cross-region replication

2. **Enhanced security**
   - mTLS for agent communication
   - Task signing/verification

3. **Advanced scheduling**
   - Priority queues
   - Resource-aware placement

4. **Config management**
   - Version control for configs
   - A/B testing templates