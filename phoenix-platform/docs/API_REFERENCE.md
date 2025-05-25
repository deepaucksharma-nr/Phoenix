# Phoenix Platform API Reference

## Overview

The Phoenix Platform provides comprehensive APIs for experiment management, pipeline configuration, and metrics analysis. The platform supports multiple API protocols:

1. **REST API** (via gRPC-gateway) - HTTP/JSON interface at `https://api.phoenix.example.com/v1`
2. **gRPC API** - Native gRPC interface at `localhost:50051`
3. **Config Generator HTTP API** - Configuration generation at `http://localhost:8082`
4. **WebSocket API** - Real-time updates at `wss://api.phoenix.example.com/v1/ws`

## Authentication

### REST API Authentication

All REST API requests require JWT authentication:

```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  https://api.phoenix.example.com/v1/experiments
```

#### Obtaining a Token

```bash
POST /v1/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "password"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### gRPC Authentication

Currently, the gRPC APIs do not require authentication in development mode. In production:
- Use TLS certificates and JWT tokens
- Configure mutual TLS for service-to-service communication

## REST API

### Base URL

```
https://api.phoenix.example.com/v1
```

### Experiments API

#### List Experiments

```http
GET /v1/experiments
```

Query Parameters:
- `status` (optional): Filter by status (running, completed, failed)
- `limit` (optional): Number of results (default: 20, max: 100)
- `offset` (optional): Pagination offset

Response:
```json
{
  "experiments": [
    {
      "id": "exp-123",
      "name": "webserver-optimization",
      "status": "running",
      "baseline_pipeline": "process-baseline-v1",
      "candidate_pipeline": "process-priority-filter-v1",
      "created_at": "2024-11-24T10:00:00Z",
      "started_at": "2024-11-24T10:05:00Z",
      "target_nodes": {
        "selector": {
          "app": "webserver"
        }
      }
    }
  ],
  "total": 42,
  "has_more": true
}
```

#### Create Experiment

```http
POST /v1/experiments
Content-Type: application/json
```

Request Body:
```json
{
  "name": "database-optimization-test",
  "description": "Reduce database server metrics by 50%",
  "baseline_pipeline": "process-baseline-v1",
  "candidate_pipeline": "process-topk-v1",
  "duration": "24h",
  "target_nodes": {
    "selector": {
      "role": "database",
      "environment": "staging"
    }
  },
  "critical_processes": [
    "postgres",
    "pgbouncer",
    "patroni"
  ],
  "config_overrides": {
    "topk_count": 30,
    "memory_limit_mib": 256
  }
}
```

Response:
```json
{
  "id": "exp-456",
  "name": "database-optimization-test",
  "status": "pending",
  "created_at": "2024-11-24T12:00:00Z",
  "deployment_status": {
    "phase": "configuring",
    "message": "Generating pipeline configurations"
  }
}
```

#### Get Experiment Details

```http
GET /v1/experiments/{id}
```

Response:
```json
{
  "id": "exp-123",
  "name": "webserver-optimization",
  "status": "running",
  "baseline_pipeline": "process-baseline-v1",
  "candidate_pipeline": "process-priority-filter-v1",
  "created_at": "2024-11-24T10:00:00Z",
  "started_at": "2024-11-24T10:05:00Z",
  "metrics": {
    "baseline": {
      "cardinality": 50000,
      "ingestion_rate_dpm": 1000000,
      "critical_processes_retained": 25,
      "collector_cpu_cores": 0.5,
      "collector_memory_mib": 256
    },
    "candidate": {
      "cardinality": 12500,
      "ingestion_rate_dpm": 250000,
      "critical_processes_retained": 25,
      "collector_cpu_cores": 0.3,
      "collector_memory_mib": 200
    },
    "reduction_percentage": 75,
    "estimated_monthly_savings_usd": 875
  }
}
```

#### Update Experiment

```http
PATCH /v1/experiments/{id}
Content-Type: application/json
```

Request Body:
```json
{
  "duration": "48h",
  "description": "Extended test for weekend traffic"
}
```

#### Stop Experiment

```http
POST /v1/experiments/{id}/stop
```

Response:
```json
{
  "id": "exp-123",
  "status": "stopping",
  "message": "Experiment stop initiated"
}
```

#### Promote Experiment Variant

```http
POST /v1/experiments/{id}/promote
Content-Type: application/json
```

Request Body:
```json
{
  "variant": "candidate",
  "rollout_strategy": "immediate"
}
```

### Pipelines API

#### List Pipeline Templates

```http
GET /v1/pipelines
```

Response:
```json
{
  "pipelines": [
    {
      "name": "process-baseline-v1",
      "version": "1.0.0",
      "description": "No optimization, full process visibility",
      "type": "baseline",
      "expected_reduction": 0
    },
    {
      "name": "process-priority-filter-v1",
      "version": "1.0.0",
      "description": "Filter by process priority",
      "type": "optimization",
      "expected_reduction": 60,
      "configurable_parameters": [
        {
          "name": "critical_processes",
          "type": "array[string]",
          "required": true
        }
      ]
    }
  ]
}
```

#### Get Pipeline Configuration

```http
GET /v1/pipelines/{name}
```

Response:
```json
{
  "name": "process-priority-filter-v1",
  "version": "1.0.0",
  "description": "Filter by process priority",
  "configuration": {
    "receivers": {
      "hostmetrics": {
        "collection_interval": "10s",
        "scrapers": {
          "process": {
            "include": [".*"],
            "metrics": [
              "process.cpu.time",
              "process.memory.physical",
              "process.memory.virtual"
            ]
          }
        }
      }
    },
    "processors": {
      "memory_limiter": {
        "limit_mib": 512,
        "spike_limit_mib": 128
      },
      "transform/classify": {
        "metric_statements": [
          {
            "context": "resource",
            "statements": [
              "set(attributes[\"process.priority\"], \"critical\") where attributes[\"process.name\"] =~ \"^(nginx|mysql|redis)\""
            ]
          }
        ]
      }
    }
  }
}
```

#### Validate Pipeline Configuration

```http
POST /v1/pipelines/validate
Content-Type: application/json
```

Request Body:
```json
{
  "configuration": {
    "receivers": {},
    "processors": {},
    "exporters": {},
    "service": {}
  }
}
```

### Metrics API

#### Get Experiment Metrics

```http
GET /v1/experiments/{id}/metrics
```

Query Parameters:
- `start` (optional): Start time (RFC3339)
- `end` (optional): End time (RFC3339)
- `resolution` (optional): Data resolution (1m, 5m, 1h)

Response:
```json
{
  "experiment_id": "exp-123",
  "time_range": {
    "start": "2024-11-24T10:00:00Z",
    "end": "2024-11-24T12:00:00Z"
  },
  "series": {
    "baseline_cardinality": [
      {"timestamp": "2024-11-24T10:00:00Z", "value": 50000},
      {"timestamp": "2024-11-24T10:05:00Z", "value": 50100}
    ],
    "candidate_cardinality": [
      {"timestamp": "2024-11-24T10:00:00Z", "value": 12500},
      {"timestamp": "2024-11-24T10:05:00Z", "value": 12600}
    ]
  }
}
```

#### Get Cost Analysis

```http
GET /v1/experiments/{id}/cost-analysis
```

Response:
```json
{
  "experiment_id": "exp-123",
  "analysis": {
    "baseline_cost": {
      "hourly_usd": 1.73,
      "daily_usd": 41.52,
      "monthly_usd": 1250.00
    },
    "optimized_cost": {
      "hourly_usd": 0.52,
      "daily_usd": 12.48,
      "monthly_usd": 375.00
    },
    "savings": {
      "hourly_usd": 1.21,
      "daily_usd": 29.04,
      "monthly_usd": 875.00,
      "percentage": 70
    }
  }
}
```

### Load Simulation API

#### Create Load Simulation

```http
POST /v1/load-simulations
Content-Type: application/json
```

Request Body:
```json
{
  "name": "high-cardinality-test",
  "profile": "high-cardinality",
  "target_nodes": {
    "selector": {
      "test": "load-sim"
    }
  },
  "duration": "1h",
  "parameters": {
    "process_count": 2000,
    "churn_rate": 10
  }
}
```

#### List Load Profiles

```http
GET /v1/load-simulations/profiles
```

Response:
```json
{
  "profiles": [
    {
      "name": "realistic",
      "description": "Simulates typical production workload",
      "default_process_count": 100,
      "default_churn_rate": 1
    },
    {
      "name": "high-cardinality",
      "description": "Many unique processes",
      "default_process_count": 2000,
      "default_churn_rate": 5
    }
  ]
}
```

## gRPC API

### Service: ExperimentService

**Endpoint**: `localhost:50051`

### Methods

#### CreateExperiment
Creates a new A/B testing experiment.

**Request**:
```protobuf
message CreateExperimentRequest {
  string name = 1;
  string description = 2;
  string baseline_pipeline = 3;
  string candidate_pipeline = 4;
  map<string, string> target_nodes = 5;
}
```

**Response**:
```protobuf
message CreateExperimentResponse {
  Experiment experiment = 1;
}
```

**Example**:
```bash
grpcurl -plaintext -d '{
  "name": "Cardinality Reduction Test",
  "description": "Testing priority filter effectiveness",
  "baseline_pipeline": "process-baseline-v1",
  "candidate_pipeline": "process-priority-filter-v1",
  "target_nodes": {
    "node1": "prod-host-001",
    "node2": "prod-host-002"
  }
}' localhost:50051 phoenix.v1.ExperimentService/CreateExperiment
```

#### GetExperiment
Retrieves an experiment by ID.

**Request**:
```protobuf
message GetExperimentRequest {
  string id = 1;
}
```

**Response**:
```protobuf
message GetExperimentResponse {
  Experiment experiment = 1;
}
```

**Example**:
```bash
grpcurl -plaintext -d '{"id": "exp-123"}' \
  localhost:50051 phoenix.v1.ExperimentService/GetExperiment
```

#### ListExperiments
Lists experiments with optional filtering.

**Request**:
```protobuf
message ListExperimentsRequest {
  string status = 1;  // optional: filter by status
  int32 limit = 2;    // optional: max results
  int32 offset = 3;   // optional: pagination offset
}
```

**Response**:
```protobuf
message ListExperimentsResponse {
  repeated Experiment experiments = 1;
}
```

**Example**:
```bash
grpcurl -plaintext -d '{"status": "running", "limit": 10}' \
  localhost:50051 phoenix.v1.ExperimentService/ListExperiments
```

#### UpdateExperiment
Updates an existing experiment.

**Request**:
```protobuf
message UpdateExperimentRequest {
  Experiment experiment = 1;
}
```

**Response**:
```protobuf
message Experiment {
  string id = 1;
  string name = 2;
  string description = 3;
  // ... other fields
}
```

#### DeleteExperiment
Deletes an experiment.

**Request**:
```protobuf
message DeleteExperimentRequest {
  string id = 1;
}
```

**Response**:
```protobuf
message DeleteExperimentResponse {
  // Empty response
}
```

#### GetExperimentStatus
Gets the current status of an experiment.

**Request**:
```protobuf
message GetExperimentStatusRequest {
  string id = 1;
}
```

**Response**:
```protobuf
message ExperimentStatus {
  string status = 1;
  string message = 2;
}
```

## Config Generator API (HTTP)

### Base URL: `http://localhost:8082`

### Endpoints

#### GET /health
Health check endpoint.

**Response**:
```json
{
  "status": "healthy",
  "service": "config-generator",
  "version": "0.1.0"
}
```

#### GET /templates
Lists available pipeline templates.

**Response**:
```json
{
  "templates": [
    "process-baseline-v1",
    "process-priority-filter-v1",
    "process-aggregated-v1"
  ]
}
```

**Example**:
```bash
curl http://localhost:8082/templates
```

#### GET /templates/{name}
Gets details of a specific template.

**Response**:
```json
{
  "name": "process-baseline-v1",
  "description": "Baseline pipeline configuration",
  "variables": ["EXPERIMENT_ID", "NAMESPACE"],
  "content": "..."
}
```

**Example**:
```bash
curl http://localhost:8082/templates/process-baseline-v1
```

#### POST /generate
Generates a configuration from a template.

**Request Body**:
```json
{
  "experiment_id": "exp-123",
  "baseline_pipeline": "process-baseline-v1",
  "candidate_pipeline": "process-priority-filter-v1",
  "target_nodes": ["node1", "node2"],
  "variables": {
    "EXPERIMENT_ID": "exp-123",
    "NAMESPACE": "phoenix-system",
    "NEW_RELIC_API_KEY_SECRET_NAME": "newrelic-secret"
  }
}
```

**Response**:
```json
{
  "success": true,
  "baseline_config": "apiVersion: v1\nkind: ConfigMap\n...",
  "candidate_config": "apiVersion: v1\nkind: ConfigMap\n...",
  "message": "Configurations generated successfully"
}
```

**Example**:
```bash
curl -X POST http://localhost:8082/generate \
  -H "Content-Type: application/json" \
  -d '{
    "experiment_id": "exp-123",
    "baseline_pipeline": "process-baseline-v1",
    "candidate_pipeline": "process-priority-filter-v1",
    "target_nodes": ["node1", "node2"],
    "variables": {
      "EXPERIMENT_ID": "exp-123",
      "NAMESPACE": "phoenix-system"
    }
  }'
```

#### POST /validate
Validates a configuration template.

**Request Body**:
```json
{
  "template": "process-custom-v1",
  "content": "processors:\n  - type: filter\n    ..."
}
```

**Response**:
```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

## WebSocket API

### Real-time Experiment Updates

```javascript
const ws = new WebSocket('wss://api.phoenix.example.com/v1/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    experiment_id: 'exp-123'
  }));
};

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  // Handle real-time metrics updates
};
```

Message Types:
- `metrics_update`: Real-time metrics
- `status_change`: Experiment status changes
- `alert`: Important notifications

## Metrics API

### Experiment Controller Metrics

**Endpoint**: `http://localhost:8081/metrics`

Available metrics:
- `phoenix_experiments_total` - Total number of experiments
- `phoenix_experiments_active` - Currently active experiments
- `phoenix_experiment_duration_seconds` - Experiment duration histogram
- `phoenix_experiment_state_transitions_total` - State transition counter
- `phoenix_api_requests_total` - API request counter
- `phoenix_api_request_duration_seconds` - API request duration

**Example**:
```bash
curl http://localhost:8081/metrics | grep phoenix_
```

## Error Handling

### REST API Error Format

All error responses follow this format:

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "Pipeline name is required",
    "details": {
      "field": "baseline_pipeline",
      "reason": "missing_required_field"
    }
  }
}
```

### Error Codes

#### gRPC Status Codes
- `OK (0)` - Success
- `INVALID_ARGUMENT (3)` - Invalid request parameters
- `NOT_FOUND (5)` - Resource not found
- `ALREADY_EXISTS (6)` - Resource already exists
- `PERMISSION_DENIED (7)` - Insufficient permissions
- `FAILED_PRECONDITION (9)` - Invalid state transition
- `INTERNAL (13)` - Internal server error

#### HTTP Status Codes
- `200 OK` - Success
- `400 Bad Request` - Invalid request
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `500 Internal Server Error` - Server error

## Rate Limiting

### REST API
- Authenticated: 1000 requests/hour
- Unauthenticated: 100 requests/hour

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1701234567
```

### gRPC API
Development mode has no rate limiting. Production recommendations:
- 100 requests/minute per client for read operations
- 10 requests/minute per client for write operations

## Client Libraries

### Go Client

#### gRPC Client
```go
import (
    pb "github.com/phoenix/platform/pkg/api/v1"
    "google.golang.org/grpc"
)

// Connect to server
conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
client := pb.NewExperimentServiceClient(conn)

// Create experiment
resp, err := client.CreateExperiment(ctx, &pb.CreateExperimentRequest{
    Name:             "Test Experiment",
    BaselinePipeline: "process-baseline-v1",
})
```

#### REST SDK
```go
import "github.com/phoenix/platform/sdk/go/phoenix"

client := phoenix.NewClient("https://api.phoenix.example.com")
client.SetToken("your-jwt-token")

experiment, err := client.CreateExperiment(&phoenix.ExperimentRequest{
    Name: "test-experiment",
    BaselinePipeline: "process-baseline-v1",
    CandidatePipeline: "process-topk-v1",
})
```

### Python Client

#### gRPC Client
```python
import grpc
from phoenix.api.v1 import experiment_pb2, experiment_pb2_grpc

# Connect to server
channel = grpc.insecure_channel('localhost:50051')
stub = experiment_pb2_grpc.ExperimentServiceStub(channel)

# Create experiment
response = stub.CreateExperiment(
    experiment_pb2.CreateExperimentRequest(
        name="Test Experiment",
        baseline_pipeline="process-baseline-v1"
    )
)
```

#### REST SDK
```python
from phoenix_sdk import PhoenixClient

client = PhoenixClient(
    base_url="https://api.phoenix.example.com",
    token="your-jwt-token"
)

experiment = client.create_experiment(
    name="test-experiment",
    baseline_pipeline="process-baseline-v1",
    candidate_pipeline="process-topk-v1"
)
```

### Pipeline Deployments API

#### Create Pipeline Deployment

```http
POST /v1/pipeline-deployments
Content-Type: application/json
```

Request Body:
```json
{
  "name": "production-intelligent-pipeline",
  "namespace": "production",
  "template": "process-intelligent-v1",
  "config": {
    "sampling_rate": 0.1,
    "batch_size": 1000,
    "memory_limit_mib": 512
  },
  "description": "Production deployment with intelligent filtering"
}
```

Response:
```json
{
  "id": "dep-789",
  "name": "production-intelligent-pipeline",
  "namespace": "production",
  "template": "process-intelligent-v1",
  "status": "active",
  "created_at": "2024-11-24T14:00:00Z",
  "created_by": "user@example.com"
}
```

#### List Pipeline Deployments

```http
GET /v1/pipeline-deployments?namespace={namespace}
```

Query Parameters:
- `namespace` (required): Filter by namespace
- `status` (optional): Filter by status (active, suspended, rollback)

Response:
```json
{
  "deployments": [
    {
      "id": "dep-789",
      "name": "production-intelligent-pipeline",
      "namespace": "production",
      "template": "process-intelligent-v1",
      "status": "active",
      "created_at": "2024-11-24T14:00:00Z",
      "updated_at": "2024-11-24T14:00:00Z"
    }
  ]
}
```

#### Get Pipeline Deployment

```http
GET /v1/pipeline-deployments/{id}
```

Response includes full deployment details with configuration.

#### Update Pipeline Deployment Configuration

```http
PATCH /v1/pipeline-deployments/{id}
Content-Type: application/json
```

Request Body:
```json
{
  "config": {
    "sampling_rate": 0.05,
    "batch_size": 2000
  },
  "reason": "Reducing sampling rate based on volume analysis"
}
```

#### Update Pipeline Deployment Status

```http
PUT /v1/pipeline-deployments/{id}/status
Content-Type: application/json
```

Request Body:
```json
{
  "status": "suspended",
  "reason": "Maintenance window"
}
```

#### Get Pipeline Deployment History

```http
GET /v1/pipeline-deployments/{id}/history
```

Response:
```json
{
  "history": [
    {
      "id": "hist-1",
      "deployment_id": "dep-789",
      "action": "create",
      "config": {...},
      "created_at": "2024-11-24T14:00:00Z",
      "created_by": "user@example.com"
    },
    {
      "id": "hist-2",
      "deployment_id": "dep-789",
      "action": "update",
      "reason": "Configuration optimization",
      "created_at": "2024-11-24T15:00:00Z",
      "created_by": "admin@example.com"
    }
  ]
}
```

#### Rollback Pipeline Deployment

```http
POST /v1/pipeline-deployments/{id}/rollback
Content-Type: application/json
```

Request Body:
```json
{
  "history_id": "hist-1",
  "reason": "Reverting to previous stable configuration"
}
```

#### Delete Pipeline Deployment

```http
DELETE /v1/pipeline-deployments/{id}
```

Response: 204 No Content

#### Export Pipeline Deployment

```http
GET /v1/pipeline-deployments/{id}/export
```

Response:
```json
{
  "deployment": {
    "id": "dep-789",
    "name": "production-intelligent-pipeline",
    "namespace": "production",
    "template": "process-intelligent-v1",
    "config": {...},
    "status": "active"
  },
  "history": [...],
  "exported_at": "2024-11-24T16:00:00Z"
}
```

### CLI Usage

The Phoenix CLI provides a comprehensive command-line interface for managing experiments and pipeline deployments.

#### Installation

```bash
# Download and install
curl -sSL https://get.phoenix.example.com/cli | bash

# Or build from source
git clone https://github.com/phoenix/platform
cd platform/phoenix-platform
make build-cli
sudo mv bin/phoenix /usr/local/bin/
```

#### Authentication

```bash
# Login interactively
phoenix auth login

# Login with credentials
phoenix auth login --username user@example.com --password yourpassword

# Check authentication status
phoenix auth status

# Logout
phoenix auth logout
```

#### Configuration Management

```bash
# Set API URL
phoenix config set api_url https://api.phoenix.example.com

# Set default namespace
phoenix config set default_namespace production

# View all configuration
phoenix config list

# Get specific config value
phoenix config get api_url

# Reset configuration
phoenix config reset
```

#### Experiment Management

```bash
# Create experiment
phoenix experiment create \
  --name "cost-optimization-test" \
  --namespace "production" \
  --pipeline-a "process-baseline-v1" \
  --pipeline-b "process-intelligent-v1" \
  --traffic-split "50/50" \
  --duration "2h" \
  --selector "app=webserver" \
  --min-cost-reduction 20 \
  --max-data-loss 2

# List experiments
phoenix experiment list --namespace production --status running

# Get experiment status
phoenix experiment status exp-123

# Follow experiment status (real-time updates)
phoenix experiment status exp-123 --follow

# Start experiment
phoenix experiment start exp-123

# Get experiment metrics
phoenix experiment metrics exp-123

# Stop experiment
phoenix experiment stop exp-123 --reason "Manual intervention required"

# Promote experiment
phoenix experiment promote exp-123 --reason "Met all success criteria"

# Export experiment configuration
phoenix experiment export exp-123 > experiment-config.yaml
```

#### Pipeline Deployment Management

```bash
# Deploy a pipeline directly (without experiment)
phoenix pipeline deploy \
  --name "prod-intelligent-pipeline" \
  --namespace "production" \
  --template "process-intelligent-v1" \
  --description "Production intelligent pipeline" \
  --config-override '{"sampling_rate": 0.1, "batch_size": 1000}'

# List pipeline deployments
phoenix pipeline deployments list --namespace production

# Get deployment status
phoenix pipeline deployment status dep-789

# Update deployment configuration
phoenix pipeline deployment update dep-789 \
  --config-override '{"sampling_rate": 0.05}' \
  --reason "Reduce sampling based on volume"

# View deployment history
phoenix pipeline deployment history dep-789

# Rollback to previous configuration
phoenix pipeline deployment rollback dep-789 \
  --history-id hist-1 \
  --reason "Performance regression detected"

# Export deployment configuration
phoenix pipeline deployment export dep-789 > deployment-backup.yaml

# List available pipeline templates
phoenix pipeline templates list
```

#### Output Formats

All commands support multiple output formats:

```bash
# Table format (default)
phoenix experiment list

# JSON format
phoenix experiment list --output json

# YAML format
phoenix experiment list --output yaml

# Custom table columns
phoenix experiment list --columns id,name,status,duration
```

#### Shell Completion

```bash
# Generate bash completion
phoenix completion bash > /etc/bash_completion.d/phoenix

# Generate zsh completion
phoenix completion zsh > "${fpath[1]}/_phoenix"

# Generate fish completion
phoenix completion fish > ~/.config/fish/completions/phoenix.fish
```

#### Advanced Usage

```bash
# Batch operations with jq
phoenix experiment list --output json | \
  jq -r '.experiments[] | select(.status == "completed") | .id' | \
  xargs -I {} phoenix experiment export {} > {}.yaml

# Monitor multiple experiments
watch -n 5 'phoenix experiment list --namespace production --output table'

# CI/CD integration
export PHOENIX_API_TOKEN="your-ci-token"
phoenix experiment create --name "ci-test-$BUILD_ID" ...
```

## Best Practices

1. **Use idempotent operations** - Retry safely on network failures
2. **Handle errors gracefully** - Check error codes and messages
3. **Use appropriate timeouts** - 30s for most operations
4. **Batch operations** - Use ListExperiments instead of multiple GetExperiment calls
5. **Monitor metrics** - Track API usage and performance
6. **Use WebSocket for real-time updates** - Avoid polling for status changes
7. **Validate configurations locally** - Use the validate endpoint before creating experiments

## API Versioning

Current version: `v1`

Version is included in:
- REST API path: `/v1/experiments`
- gRPC package: `phoenix.v1`
- HTTP headers: `X-API-Version: v1`

## Webhooks (Future)

Planned webhook events:
- `experiment.created`
- `experiment.started`
- `experiment.completed`
- `experiment.failed`
- `experiment.cancelled`

## Support

For API issues or questions:
1. Check service logs
2. Review error messages
3. Consult integration tests for examples
4. Open an issue in the repository