# Phoenix API

The Phoenix API is the central control plane for the Phoenix Platform, managing experiments through agent-based task distribution, A/B testing, and real-time metrics analysis.

## Overview

Phoenix API provides:
- RESTful API v2 with WebSocket on port 8080
- PostgreSQL-based task queue with 30-second long-polling
- A/B testing framework with baseline/candidate pipelines
- Real-time KPI calculation (70% cost reduction demonstrated)
- Agent authentication via X-Agent-Host-ID header

## Architecture

```
┌───────────────────────────────┐
│   Phoenix API (Port 8080)     │
│   REST API v2 + WebSocket     │
└─────────────┬────────────────┘
              │
    ┌─────────▼─────────┐
    │   Core Services    │
    │  - Experiments     │
    │  - Task Queue      │
    │  - KPI Analysis    │
    └─────────┬─────────┘
              │
    ┌─────────▼─────────┐
    │   PostgreSQL       │
    │  (Task Queue DB)   │
    └───────────────────┘
```

## Quick Start

### Development

```bash
# Install dependencies
go mod download

# Run database migrations
make migrate

# Start the API
make run

# Run tests
make test
```

### Docker

```bash
# Build image
docker build -t phoenix/api .

# Run container (single port for REST + WebSocket)
docker run -p 8080:8080 \
  -e DATABASE_URL=postgresql://... \
  phoenix/api
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API port (REST + WebSocket) | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `TASK_POLL_TIMEOUT` | Long-polling timeout | `30s` |
| `LOG_LEVEL` | Logging level | `info` |
| `ENABLE_AUTH` | Enable authentication | `true` |
| `METRICS_INTERVAL` | KPI calculation interval | `30s` |

## API Endpoints

### Experiments
- `GET /api/v2/experiments` - List experiments
- `POST /api/v2/experiments` - Create A/B test experiment
- `GET /api/v2/experiments/{id}` - Get experiment details
- `POST /api/v2/experiments/{id}/start` - Start experiment
- `POST /api/v2/experiments/{id}/stop` - Stop experiment
- `POST /api/v2/experiments/{id}/promote` - Promote candidate
- `GET /api/v2/experiments/{id}/kpis` - Get calculated KPIs

### Agents
- `GET /api/v2/agents` - List registered agents
- `POST /api/v2/agents/{hostId}/heartbeat` - Agent heartbeat
- `GET /api/v2/tasks/poll` - Poll for tasks (30s long-poll)
- `POST /api/v2/tasks/{id}/status` - Update task status

### Pipeline Templates
- `GET /api/v2/pipeline-templates` - List templates
- `GET /api/v2/pipeline-templates/{id}` - Get template details
- `POST /api/v2/pipeline-deployments` - Deploy pipeline

## WebSocket Events

Connect to `ws://localhost:8080/ws` for real-time updates (same port as REST):

```javascript
// Event types
{
  "type": "experiment_update",
  "data": { 
    "experiment_id": "exp-123", 
    "phase": "running",
    "baseline_cost": 5000,
    "candidate_cost": 1500,
    "savings_percent": 70
  }
}

{
  "type": "agent_status", 
  "data": { 
    "host_id": "agent-001", 
    "status": "healthy",
    "active_tasks": ["task-123"]
  }
}

{
  "type": "metric_flow",
  "data": { 
    "total_cost_rate": 125.50,
    "cardinality_reduction": 70
  }
}
```

## Development

### Project Structure

```
phoenix-api/
├── cmd/api/          # Application entrypoint
├── internal/
│   ├── api/         # HTTP handlers + WebSocket
│   ├── controller/  # Experiment controller
│   ├── models/      # Data models
│   ├── services/    # Pipeline services
│   ├── store/       # PostgreSQL store
│   ├── tasks/       # Task queue implementation
│   └── websocket/   # WebSocket hub
├── migrations/      # PostgreSQL migrations
└── Dockerfile      # Container definition
```

### Testing

```bash
# Unit tests
make test

# Integration tests
make test-integration

# Coverage report
make test-coverage
```

## Deployment

See [Operations Guide](../../docs/operations/OPERATIONS_GUIDE_COMPLETE.md) for production deployment instructions.