# Phoenix Platform API Documentation

This directory contains comprehensive API documentation for the Phoenix Platform.

## 📚 API Documentation

### Core API References
- [Phoenix API v2](PHOENIX_API_v2.md) - Complete API v2 documentation with examples
- [REST API Reference](rest-api.md) - RESTful endpoint reference
- [WebSocket API](websocket-api.md) - Real-time WebSocket events
- [Pipeline Validation API](PIPELINE_VALIDATION_API.md) - Pipeline validation endpoints

### Quick Start

#### Base URLs
- **REST API**: `http://localhost:8080/api/v2`
- **WebSocket**: `ws://localhost:8080/ws` (same port as REST)
- **Health Check**: `http://localhost:8080/health`

#### Authentication

**For Agents:**
```bash
curl -H "X-Agent-Host-ID: agent-001" \
     http://localhost:8080/api/v2/agent/tasks
```

**For Users (optional in dev):**
```bash
curl -H "Authorization: Bearer <jwt-token>" \
     http://localhost:8080/api/v2/experiments
```

## 🚀 Common Operations

### 1. Create an Experiment
```bash
curl -X POST http://localhost:8080/api/v2/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Reduce API Costs",
    "baseline_pipeline": "baseline",
    "candidate_pipeline": "adaptive-filter",
    "duration": "24h"
  }'
```

### 2. Start Agent Polling
```bash
# Long-polling with 30s timeout
curl -H "X-Agent-Host-ID: agent-001" \
     http://localhost:8080/api/v2/agent/tasks
```

### 3. Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.on('open', () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    payload: { events: ['experiment_update'] }
  }));
});
```

## 🔌 Collector Support

### OpenTelemetry (Default)
```json
{
  "collector_type": "otel",
  "pipeline_template": "adaptive-filter"
}
```

### NRDOT (New Relic)
```json
{
  "collector_type": "nrdot",
  "pipeline_template": "nrdot-cardinality",
  "nrdot_config": {
    "license_key": "your-key",
    "otlp_endpoint": "otlp.nr-data.net:4317"
  }
}
```

## 📊 Key Endpoints

### Experiments
- `POST /api/v2/experiments` - Create experiment
- `GET /api/v2/experiments/{id}` - Get experiment details
- `POST /api/v2/experiments/{id}/start` - Start experiment
- `POST /api/v2/experiments/{id}/stop` - Stop experiment
- `GET /api/v2/experiments/{id}/metrics` - Get metrics

### Pipelines
- `GET /api/v2/pipelines/templates` - List templates
- `POST /api/v2/pipelines/validate` - Validate config
- `POST /api/v2/pipelines/render` - Render template

### Agent Operations
- `GET /api/v2/agent/tasks` - Poll for tasks (long-poll)
- `POST /api/v2/agent/heartbeat` - Send heartbeat
- `POST /api/v2/agent/metrics` - Report metrics

### Real-time Monitoring
- `WS /ws` - WebSocket connection
- `GET /api/v2/metrics/cost-flow` - Live cost data
- `GET /api/v2/fleet/status` - Agent fleet status

## 🛠️ SDKs and Tools

### CLI Tool
```bash
# Install Phoenix CLI
go install github.com/phoenix/platform/phoenix-cli

# Create experiment
phoenix-cli experiment create \
  --name "Test" \
  --baseline "baseline" \
  --candidate "topk"
```

### Language SDKs
- **Go**: Built-in client in `/pkg/client`
- **Python**: Coming soon
- **JavaScript/TypeScript**: Coming soon

## 📖 Additional Resources

- [API Changelog](../../CHANGELOG.md)
- [Error Codes Reference](PHOENIX_API_v2.md#error-handling)
- [Rate Limiting](PHOENIX_API_v2.md#rate-limiting)
- [OpenAPI Specification](http://localhost:8080/api/v2/openapi.json)

## 🔍 API Version History

- **v2** (Current) - REST + WebSocket, NRDOT support
- **v1** (Deprecated) - Legacy REST API

## 💡 Best Practices

1. **Use Long-Polling**: Agents should use 30s timeout
2. **Batch Operations**: Group related API calls
3. **Handle Retries**: Implement exponential backoff
4. **Monitor Rate Limits**: Check X-RateLimit headers
5. **Use WebSocket**: For real-time updates

## 🚨 Common Issues

| Issue | Solution |
|-------|----------|
| 401 Unauthorized | Check authentication headers |
| 404 Not Found | Verify API version (v2) |
| 429 Too Many Requests | Implement rate limiting |
| WebSocket drops | Implement reconnection logic |

## Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid pipeline configuration",
    "details": {
      "field": "config.threshold",
      "reason": "Must be between 0.1 and 0.99"
    }
  }
}
```

## Support

- [GitHub Issues](https://github.com/phoenix/platform/issues)
- [Discord Community](https://discord.gg/phoenix)
- [Documentation](https://docs.phoenix.io)