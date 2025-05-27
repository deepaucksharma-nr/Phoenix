# Phoenix Agent

Lightweight agent that polls Phoenix API for tasks and manages OpenTelemetry collectors for A/B testing experiments.

## Overview

Phoenix Agent is a minimal footprint service that:
- Polls Phoenix API for tasks using X-Agent-Host-ID authentication
- Manages baseline/candidate OTel collectors for A/B testing
- Executes pipeline templates (Adaptive Filter, TopK, Hybrid)
- Uses 30-second long-polling for task distribution
- Reports experiment results and metrics back to API

## Features

- **Minimal Resource Usage**: <50MB RAM, <1% CPU
- **Task Queue Design**: Long-polling with 30s timeout
- **A/B Testing Support**: Concurrent baseline/candidate pipelines
- **Authentication**: X-Agent-Host-ID header for secure polling
- **Secure**: Outbound-only connections (no incoming ports)

## Quick Start

### Binary Installation

```bash
# Download latest release
curl -L https://github.com/phoenix/releases/latest/phoenix-agent -o phoenix-agent
chmod +x phoenix-agent

# Run agent with host ID
PHOENIX_API_URL=http://phoenix-api:8080 \
AGENT_HOST_ID=$(hostname) \
./phoenix-agent
```

### Docker

```bash
docker run -d \
  --name phoenix-agent \
  -e PHOENIX_API_URL=http://phoenix-api:8080 \
  -e AGENT_HOST_ID=$(hostname) \
  -v /var/run/docker.sock:/var/run/docker.sock \
  phoenix/agent:latest
```

### Systemd Service

```bash
# Install service
sudo ./deployments/systemd/install.sh

# Configure
sudo vim /etc/phoenix/agent.env

# Start service
sudo systemctl start phoenix-agent
sudo systemctl enable phoenix-agent
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PHOENIX_API_URL` | Phoenix API URL | Required |
| `AGENT_HOST_ID` | Unique agent identifier | Hostname |
| `TASK_POLL_TIMEOUT` | Long-polling timeout | `30s` |
| `CONFIG_DIR` | OTel config directory | `/etc/phoenix/configs` |
| `LOG_LEVEL` | Logging level | `info` |
| `MAX_RETRIES` | Task retry attempts | `3` |
| `COLLECTOR_TYPE` | Collector type (`otel` or `nrdot`) | `otel` |
| `OTEL_COLLECTOR_ENDPOINT` | OpenTelemetry endpoint | `http://localhost:4317` |
| `NRDOT_OTLP_ENDPOINT` | NRDOT endpoint | `https://otlp.nr-data.net:4317` |
| `NEW_RELIC_LICENSE_KEY` | New Relic license key (for NRDOT) | - |

## Architecture

```
┌────────────────────────────────────┐
│           Phoenix Agent               │
├────────────────────────────────────┤
│ Task Poller (X-Agent-Host-ID)        │──► /api/v2/tasks/poll
├────────────────────────────────────┤      (30s timeout)
│          Supervisor                   │
├─────────────────┬──────────────────┤
│ Baseline OTel   │ Candidate OTel  │──► Metrics Backends
└─────────────────┴──────────────────┘
```

## How It Works

1. **Authentication**: Uses X-Agent-Host-ID header for identification
2. **Task Polling**: Long-polls `/api/v2/tasks/poll` with 30s timeout
3. **A/B Testing**: Runs baseline and candidate pipelines concurrently
4. **Pipeline Templates**: Executes Adaptive Filter, TopK, or Hybrid configs
5. **Result Reporting**: Updates task status and experiment metrics

## Collector Management

The agent supports both OpenTelemetry and NRDOT collectors:

### OpenTelemetry Collector (Default)

The agent manages OTel collectors as child processes:

```json
// Task received from API
{
  "id": "task-789",
  "type": "deploy_pipeline",
  "experiment_id": "exp-123",
  "action": "start",
  "config": {
    "pipeline_url": "http://api/configs/adaptive-filter-v1.yaml",
    "variant": "candidate",
    "variables": {
      "threshold": "0.8",
      "window_size": "60s"
    }
  },
  "priority": 1
}
```

Agent will:
1. Download pipeline config from URL
2. Apply variable substitution for template
3. Start OTel collector for specified variant (baseline/candidate)
4. Monitor metrics and cardinality reduction
5. Report results via `/api/v2/tasks/{id}/status`

### New Relic NRDOT

When configured for NRDOT, the agent:

```bash
# Set NRDOT environment
export COLLECTOR_TYPE=nrdot
export NEW_RELIC_LICENSE_KEY=your-license-key
export NRDOT_OTLP_ENDPOINT=https://otlp.nr-data.net:4317

# Start agent
./phoenix-agent
```

NRDOT-specific features:
- Direct integration with New Relic One
- Optimized for New Relic infrastructure
- License key authentication
- Enhanced metric compression

Pipeline configs are automatically adjusted for NRDOT exporters:
```yaml
exporters:
  nrdot:
    endpoint: ${NRDOT_OTLP_ENDPOINT}
    headers:
      api-key: ${NEW_RELIC_LICENSE_KEY}
```

## Monitoring

### Health Check

```bash
# Check agent status
curl http://localhost:8090/health
```

### Metrics

Agent exposes Prometheus metrics:

- `phoenix_agent_tasks_total` - Total tasks processed
- `phoenix_agent_collectors_active` - Active collectors by variant
- `phoenix_agent_poll_duration_seconds` - Task polling latency
- `phoenix_agent_experiment_status` - Current experiment status
- `phoenix_agent_cardinality_reduction` - Observed reduction percentage

## Troubleshooting

### Agent not connecting to API?

```bash
# Check connectivity
curl -H "X-Agent-Host-ID: $(hostname)" $PHOENIX_API_URL/api/v2/health

# Check logs
journalctl -u phoenix-agent -f

# Verify environment
systemctl show phoenix-agent | grep Environment
```

### Collectors not starting?

```bash
# Check config directory permissions
ls -la /etc/phoenix/configs/

# Verify OTel binary
which otelcol-contrib

# Check collector logs
tail -f /etc/phoenix/configs/*.log
```

## Development

### Building

```bash
# Build binary
make build

# Run tests
make test

# Build Docker image
make docker
```

### Testing

```bash
# Unit tests
go test ./...

# Integration test with API
PHOENIX_API_URL=http://localhost:8080 make test-integration
```

## Security

- No incoming network connections (outbound-only)
- Task polling with X-Agent-Host-ID authentication
- PostgreSQL task queue ensures atomic assignment
- Process isolation between baseline/candidate collectors
- Minimal system permissions required