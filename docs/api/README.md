# Phoenix API Documentation

The Phoenix Platform provides comprehensive REST and WebSocket APIs for managing experiments, pipelines, and real-time monitoring.

## API Documentation

### REST API
- [REST API Reference](rest-api.md) - Complete HTTP endpoint documentation
- [Authentication Guide](authentication.md) - JWT and agent authentication
- [OpenAPI Specification](openapi.yaml) - Machine-readable API spec

### WebSocket API
- [WebSocket API Reference](websocket-api.md) - Real-time event streaming
- Channels: experiments, metrics, agents, deployments, alerts, cost-flow

### Legacy Documentation
- [Phoenix API v2](PHOENIX_API_v2.md) - Previous API version (deprecated)

## Quick Start

### 1. Obtain Authentication Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

### 2. Create an Experiment
```bash
curl -X POST http://localhost:8080/api/v1/experiments \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Reduce metrics cardinality",
    "baseline_pipeline_id": "default",
    "candidate_pipeline_id": "adaptive-filter-v2",
    "traffic_split": 20
  }'
```

### 3. Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=<token>');
ws.onmessage = (event) => {
  console.log('Real-time update:', JSON.parse(event.data));
};
```

## API Features

### Core Capabilities
- **Experiment Management** - Create, monitor, and analyze A/B tests
- **Pipeline Configuration** - Deploy and validate optimization pipelines
- **Agent Orchestration** - Manage distributed agent fleet
- **Real-time Monitoring** - WebSocket streaming for live updates
- **Cost Analysis** - Track savings and optimization impact

### Key Endpoints

#### Experiments
- `POST /api/v1/experiments` - Create new experiment
- `GET /api/v1/experiments/{id}` - Get experiment details
- `POST /api/v1/experiments/{id}/start` - Start experiment
- `GET /api/v1/experiments/{id}/metrics` - Get experiment metrics

#### Pipelines
- `GET /api/v1/pipelines` - List available pipelines
- `POST /api/v1/pipelines/validate` - Validate configuration
- `POST /api/v1/pipelines/render` - Render pipeline template

#### Agents
- `GET /api/v1/agent/tasks` - Poll for tasks (agent-only)
- `POST /api/v1/agent/heartbeat` - Send heartbeat
- `POST /api/v1/agent/metrics` - Report metrics

#### Analysis
- `GET /api/v1/cost-flow` - Real-time cost analysis
- `GET /api/v1/fleet/status` - Fleet overview

## Authentication

### JWT Authentication (Users/CLI)
```
Authorization: Bearer <jwt-token>
```

### Agent Authentication
```
X-Agent-Host-ID: <agent-host-id>
```

## Error Handling

Standard error response format:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid pipeline configuration",
    "details": {
      "field": "config.importance_threshold",
      "reason": "Must be between 0.1 and 0.99"
    }
  }
}
```

## Rate Limits

- Standard endpoints: 1000 req/min
- Agent endpoints: 10000 req/min
- WebSocket: 100 concurrent connections

## SDK Support

Official SDKs available:
- Go SDK: `github.com/phoenix/platform/sdk/go`
- Python SDK: `pip install phoenix-platform`
- Node.js SDK: `npm install @phoenix/platform-sdk`

## Support

- [Issue Tracker](https://github.com/phoenix/platform/issues)
- [Discord Community](https://discord.gg/phoenix)
- Email: api-support@phoenix.io