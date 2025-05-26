# Phoenix API

The Phoenix API is the central control plane for the Phoenix Platform, managing experiments, task distribution, and metrics analysis.

## Overview

Phoenix API provides:
- RESTful API for all platform operations
- WebSocket support for real-time updates
- PostgreSQL-based task queue for agent work distribution
- Built-in metrics analysis and KPI calculation
- JWT-based authentication and authorization

## Architecture

```
┌─────────────────┐
│   HTTP/REST     │───► Experiments, Pipelines, Agents
│   WebSocket     │───► Real-time Updates
└────────┬────────┘
         │
    ┌────▼────┐
    │  Core   │
    │ Services│
    └────┬────┘
         │
    ┌────▼────┐
    │PostgreSQL│
    └─────────┘
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

# Run container
docker run -p 8080:8080 -p 8081:8081 \
  -e DATABASE_URL=postgresql://... \
  phoenix/api
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP API port | `8080` |
| `WEBSOCKET_PORT` | WebSocket port | `8081` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `PROMETHEUS_URL` | Prometheus server URL | `http://localhost:9090` |
| `PUSHGATEWAY_URL` | Pushgateway URL | `http://localhost:9091` |
| `JWT_SECRET` | JWT signing secret | Required |
| `LOG_LEVEL` | Logging level | `info` |

## API Endpoints

### Experiments
- `GET /api/v1/experiments` - List experiments
- `POST /api/v1/experiments` - Create experiment
- `GET /api/v1/experiments/{id}` - Get experiment details
- `POST /api/v1/experiments/{id}/start` - Start experiment
- `POST /api/v1/experiments/{id}/stop` - Stop experiment
- `POST /api/v1/experiments/{id}/promote` - Promote to production

### Agents
- `GET /api/v1/agents` - List registered agents
- `POST /api/v1/agents/heartbeat` - Agent heartbeat
- `GET /api/v1/agents/tasks` - Poll for tasks (long-polling)
- `PUT /api/v1/agents/tasks/{id}` - Update task status

### Pipelines
- `GET /api/v1/pipelines` - List pipeline templates
- `GET /api/v1/pipelines/{id}` - Get pipeline details
- `POST /api/v1/pipelines/{id}/deploy` - Deploy pipeline

## WebSocket Events

Connect to `ws://localhost:8081/ws` for real-time updates:

```javascript
// Event types
{
  "type": "experiment_started",
  "data": { "experiment_id": "...", "phase": "warmup" }
}

{
  "type": "metrics_updated", 
  "data": { "experiment_id": "...", "kpis": {...} }
}

{
  "type": "agent_status_changed",
  "data": { "host_id": "...", "status": "active" }
}
```

## Development

### Project Structure

```
phoenix-api/
├── cmd/api/          # Application entrypoint
├── internal/
│   ├── api/         # HTTP handlers
│   ├── controller/  # Business logic
│   ├── models/      # Data models
│   ├── store/       # Database layer
│   └── websocket/   # WebSocket hub
├── migrations/      # SQL migrations
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