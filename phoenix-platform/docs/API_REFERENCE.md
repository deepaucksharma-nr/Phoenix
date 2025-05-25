# Phoenix Platform API Reference

## Overview

The Phoenix Platform provides two main APIs:
1. **Experiment Controller gRPC API** - For experiment management
2. **Config Generator HTTP API** - For configuration generation

## Experiment Controller API (gRPC)

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

## Error Codes

### gRPC Status Codes
- `OK (0)` - Success
- `INVALID_ARGUMENT (3)` - Invalid request parameters
- `NOT_FOUND (5)` - Resource not found
- `ALREADY_EXISTS (6)` - Resource already exists
- `PERMISSION_DENIED (7)` - Insufficient permissions
- `FAILED_PRECONDITION (9)` - Invalid state transition
- `INTERNAL (13)` - Internal server error

### HTTP Status Codes
- `200 OK` - Success
- `400 Bad Request` - Invalid request
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `500 Internal Server Error` - Server error

## Authentication

Currently, the APIs do not require authentication in development mode. In production:

1. **gRPC**: Use TLS certificates and JWT tokens
2. **HTTP**: Use API keys or OAuth 2.0

## Rate Limiting

Development mode has no rate limiting. Production recommendations:
- 100 requests/minute per client for read operations
- 10 requests/minute per client for write operations

## Best Practices

1. **Use idempotent operations** - Retry safely on network failures
2. **Handle errors gracefully** - Check error codes and messages
3. **Use appropriate timeouts** - 30s for most operations
4. **Batch operations** - Use ListExperiments instead of multiple GetExperiment calls
5. **Monitor metrics** - Track API usage and performance

## Client Libraries

### Go Client Example

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

### Python Client Example

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

## Webhooks (Future)

Planned webhook events:
- `experiment.created`
- `experiment.started`
- `experiment.completed`
- `experiment.failed`
- `experiment.cancelled`

## API Versioning

Current version: `v1`

Version is included in:
- gRPC package: `phoenix.v1`
- HTTP headers: `X-API-Version: v1`

## Support

For API issues or questions:
1. Check service logs
2. Review error messages
3. Consult integration tests for examples
4. Open an issue in the repository