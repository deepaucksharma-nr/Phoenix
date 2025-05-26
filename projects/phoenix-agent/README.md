# Phoenix Agent

Lightweight polling agent that manages OpenTelemetry collectors on target hosts.

## Overview

Phoenix Agent is a minimal footprint service that:
- Polls Phoenix API for work assignments using long-polling
- Manages multiple OTel collector processes
- Pushes metrics to Prometheus Pushgateway
- Self-registers with zero configuration
- Handles automatic reconnection and retries

## Features

- **Minimal Resource Usage**: <50MB RAM, <1% CPU
- **Zero Configuration**: Automatically registers with API
- **Multi-Collector Support**: Run 100+ collectors per agent
- **Fault Tolerant**: Automatic reconnection and retry logic
- **Secure**: No incoming connections required

## Quick Start

### Binary Installation

```bash
# Download latest release
curl -L https://github.com/phoenix/releases/latest/phoenix-agent -o phoenix-agent
chmod +x phoenix-agent

# Run agent
PHOENIX_API_URL=http://phoenix-api:8080 ./phoenix-agent
```

### Docker

```bash
docker run -d \
  --name phoenix-agent \
  -e PHOENIX_API_URL=http://phoenix-api:8080 \
  -e HOST_ID=$(hostname) \
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
| `HOST_ID` | Unique host identifier | Hostname |
| `POLL_INTERVAL` | Task polling interval | `10s` |
| `CONFIG_DIR` | OTel config directory | `/etc/phoenix/configs` |
| `LOG_LEVEL` | Logging level | `info` |
| `PUSHGATEWAY_URL` | Metrics pushgateway | From API |

## Architecture

```
┌─────────────────┐
│  Phoenix Agent  │
├─────────────────┤
│   Task Poller   │──────► Phoenix API
├─────────────────┤
│   Supervisor    │
├─────────────────┤
│ OTel Collectors │──────► Pushgateway
└─────────────────┘
```

## How It Works

1. **Registration**: Agent registers with API on startup
2. **Polling**: Long-polls API for task assignments
3. **Execution**: Starts/stops OTel collectors based on tasks
4. **Reporting**: Pushes metrics to Pushgateway
5. **Health**: Sends periodic heartbeats to API

## OTel Collector Management

The agent manages OTel collectors as child processes:

```go
// Task received from API
{
  "id": "exp-123-baseline",
  "action": "start",
  "config": {
    "config_url": "https://configs/baseline.yaml",
    "variables": {
      "BATCH_SIZE": "1000",
      "BATCH_TIMEOUT": "10s"
    }
  }
}
```

Agent will:
1. Download config from URL
2. Apply variable substitution
3. Start OTel collector process
4. Monitor process health
5. Report status back to API

## Monitoring

### Health Check

```bash
# Check agent status
curl http://localhost:8090/health
```

### Metrics

Agent exposes Prometheus metrics:

- `phoenix_agent_tasks_total` - Total tasks processed
- `phoenix_agent_collectors_active` - Active collectors
- `phoenix_agent_api_errors_total` - API communication errors
- `phoenix_agent_uptime_seconds` - Agent uptime

## Troubleshooting

### Agent not connecting to API?

```bash
# Check connectivity
curl $PHOENIX_API_URL/health

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

- No incoming network connections
- All communication is outbound (polling)
- Configs downloaded over HTTPS
- Process isolation for collectors
- Minimal system permissions required