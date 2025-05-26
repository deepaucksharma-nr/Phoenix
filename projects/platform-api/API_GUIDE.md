# Phoenix Platform API Guide

The Phoenix Platform API is the core service for managing telemetry optimization experiments. It provides RESTful endpoints for experiment lifecycle management and WebSocket support for real-time updates.

## Quick Start

```bash
# Start the API with database setup
../../scripts/run-platform-api.sh

# Or manually:
export DATABASE_URL="postgres://phoenix:phoenix@localhost:5432/phoenix?sslmode=disable"
go run cmd/api/main.go
```

## API Endpoints

### Health Check
```bash
GET /health
```

### Experiments

#### List Experiments
```bash
GET /api/v1/experiments

# Response:
[
  {
    "id": "uuid",
    "name": "prometheus-optimization-v1",
    "status": "running",
    "cost_saving_percent": 45.2,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Create Experiment
```bash
POST /api/v1/experiments
Content-Type: application/json

{
  "name": "datadog-optimization",
  "description": "Reduce Datadog metric cardinality",
  "baseline_pipeline": "datadog-baseline",
  "candidate_pipeline": "datadog-optimized",
  "target_nodes": {
    "datadog": "datadog-agent-0"
  }
}
```

#### Get Experiment
```bash
GET /api/v1/experiments/{id}
```

#### Update Experiment Status
```bash
PUT /api/v1/experiments/{id}/status
Content-Type: application/json

{
  "status": "running" | "paused" | "completed" | "failed"
}
```

#### Delete Experiment
```bash
DELETE /api/v1/experiments/{id}
```

### Pipeline Deployments

#### List Deployments
```bash
GET /api/v1/pipelines/deployments?namespace=default&status=active
```

#### Create Deployment
```bash
POST /api/v1/pipelines/deployments
Content-Type: application/json

{
  "name": "prometheus-optimized",
  "namespace": "phoenix-system",
  "pipeline_name": "prometheus-topk-v1",
  "pipeline_config": {
    "processors": ["filter", "sample", "aggregate"],
    "sampling_rate": 0.1
  },
  "replicas": 3
}
```

## WebSocket Real-time Updates

Connect to the WebSocket endpoint for real-time experiment updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

// Subscribe to experiment updates
ws.send(JSON.stringify({
  type: 'subscribe',
  data: { topic: 'experiment:EXPERIMENT_ID' }
}));

// Subscribe to metrics
ws.send(JSON.stringify({
  type: 'subscribe',
  data: { topic: 'metrics:EXPERIMENT_ID' }
}));
```

### Message Types

- `experiment_update`: Experiment status changes
- `metric_update`: Real-time metric updates
- `status_change`: Pipeline status changes
- `notification`: System notifications
- `heartbeat`: Connection keep-alive

## Database Schema

The API uses PostgreSQL with the following main tables:

- `experiments`: Core experiment data
- `experiment_states`: State transition history
- `experiment_metrics`: Time-series metrics
- `experiment_results`: Analysis results
- `pipeline_deployments`: Pipeline configurations
- `audit_log`: Change tracking

## Examples

### Complete Experiment Workflow

```bash
# 1. Create experiment
experiment_id=$(curl -s -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prometheus-test-'$(date +%s)'",
    "description": "Test Prometheus optimization",
    "baseline_pipeline": "prometheus-baseline",
    "candidate_pipeline": "prometheus-optimized",
    "target_nodes": {"prometheus": "prometheus-0"}
  }' | jq -r '.id')

# 2. Start experiment
curl -X PUT http://localhost:8080/api/v1/experiments/$experiment_id/status \
  -H "Content-Type: application/json" \
  -d '{"status": "running"}'

# 3. Monitor progress (in another terminal)
curl http://localhost:8080/api/v1/experiments/$experiment_id

# 4. Complete experiment
curl -X PUT http://localhost:8080/api/v1/experiments/$experiment_id/status \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

### WebSocket Monitoring

Open `examples/websocket-client.html` in a browser or use websocat:

```bash
# Install websocat
brew install websocat

# Connect and subscribe
websocat ws://localhost:8080/ws
{"type":"subscribe","data":{"topic":"experiment:YOUR_EXPERIMENT_ID"}}
```

## Configuration

Environment variables:

- `DATABASE_URL`: PostgreSQL connection string (required)
- `PORT`: HTTP server port (default: 8080)
- `LOG_LEVEL`: Logging level (default: info)

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
make build
```

### Running with Docker
```bash
docker build -t phoenix-platform-api .
docker run -p 8080:8080 -e DATABASE_URL=... phoenix-platform-api
```

## Architecture

The Platform API follows a clean architecture pattern:

```
cmd/api/          # Application entry point
internal/
  services/       # Business logic
  store/          # Data persistence
  websocket/      # Real-time communication
  middleware/     # HTTP middleware
migrations/       # Database migrations
```

## Metrics

Prometheus metrics are exposed at `/metrics`:

- `phoenix_experiments_total`: Total experiments created
- `phoenix_experiments_active`: Currently active experiments
- `phoenix_api_requests_total`: API request count
- `phoenix_websocket_connections`: Active WebSocket connections

## Security

- CORS is configured for development (adjust for production)
- WebSocket connections support authentication tokens
- All changes are logged to the audit_log table

## Troubleshooting

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Test connection
psql -U phoenix -h localhost -d phoenix
```

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Migration Errors
```bash
# Reset database (development only!)
docker-compose exec postgres psql -U phoenix -c "DROP DATABASE phoenix;"
docker-compose exec postgres psql -U phoenix -c "CREATE DATABASE phoenix;"

# Re-run migrations
../../scripts/run-platform-api.sh
```