# Phoenix REST API Reference

## Overview

The Phoenix API provides a RESTful interface for managing experiments, pipelines, and agents. All endpoints are served on port 8080 with both REST and WebSocket support.

**Base URL**: `http://localhost:8080/api/v1`

## Authentication

### JWT Authentication
Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <jwt-token>
```

### Agent Authentication
Agent endpoints use host-based authentication:

```
X-Agent-Host-ID: <agent-host-id>
```

## Common Response Format

### Success Response
```json
{
  "data": {},
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-01-20T10:00:00Z"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {}
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-01-20T10:00:00Z"
  }
}
```

## Endpoints

### Health & Status

#### GET /health
Health check endpoint.

**Response**:
```json
{
  "status": "healthy",
  "version": "3.0.0",
  "timestamp": "2024-01-20T10:00:00Z"
}
```

#### GET /api/v1/fleet/status
Get fleet-wide status overview.

**Response**:
```json
{
  "data": {
    "total_agents": 150,
    "active_agents": 148,
    "total_experiments": 12,
    "active_experiments": 3,
    "metrics_per_second": 1250000,
    "cardinality_reduction": 0.72
  }
}
```

### Experiments

#### POST /api/v1/experiments
Create a new experiment.

**Request**:
```json
{
  "name": "Reduce app metrics cardinality",
  "description": "Test adaptive filter on application metrics",
  "baseline_pipeline_id": "pipeline-123",
  "candidate_pipeline_id": "pipeline-456",
  "traffic_split": 20,
  "duration_hours": 24,
  "success_criteria": {
    "min_cardinality_reduction": 0.6,
    "max_data_loss": 0.01,
    "max_latency_increase_ms": 10
  }
}
```

**Response**:
```json
{
  "data": {
    "id": "exp-789",
    "name": "Reduce app metrics cardinality",
    "status": "created",
    "created_at": "2024-01-20T10:00:00Z",
    "baseline_deployment_id": "dep-111",
    "candidate_deployment_id": "dep-222"
  }
}
```

#### GET /api/v1/experiments
List all experiments with filtering.

**Query Parameters**:
- `status` - Filter by status (created, running, completed, failed)
- `limit` - Results per page (default: 20, max: 100)
- `offset` - Pagination offset
- `sort` - Sort field (created_at, name, status)
- `order` - Sort order (asc, desc)

**Response**:
```json
{
  "data": [
    {
      "id": "exp-789",
      "name": "Reduce app metrics cardinality",
      "status": "running",
      "progress": 0.45,
      "metrics": {
        "baseline_mps": 500000,
        "candidate_mps": 150000,
        "reduction_rate": 0.70
      },
      "created_at": "2024-01-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 45,
    "limit": 20,
    "offset": 0
  }
}
```

#### GET /api/v1/experiments/{id}
Get experiment details.

**Response**:
```json
{
  "data": {
    "id": "exp-789",
    "name": "Reduce app metrics cardinality",
    "status": "running",
    "progress": 0.45,
    "baseline_pipeline": {
      "id": "pipeline-123",
      "name": "Default Pipeline",
      "type": "passthrough"
    },
    "candidate_pipeline": {
      "id": "pipeline-456",
      "name": "Adaptive Filter v2",
      "type": "adaptive_filter"
    },
    "deployments": {
      "baseline": {
        "id": "dep-111",
        "status": "healthy",
        "agent_count": 75
      },
      "candidate": {
        "id": "dep-222",
        "status": "healthy",
        "agent_count": 15
      }
    },
    "metrics": {
      "start_time": "2024-01-20T10:00:00Z",
      "duration_elapsed": "10h45m",
      "baseline_metrics_total": 45000000,
      "candidate_metrics_total": 13500000,
      "reduction_rate": 0.70,
      "error_rate": 0.0001,
      "p99_latency_ms": 5.2
    }
  }
}
```

#### POST /api/v1/experiments/{id}/start
Start an experiment.

**Response**:
```json
{
  "data": {
    "id": "exp-789",
    "status": "starting",
    "message": "Experiment starting, deploying to 90 agents"
  }
}
```

#### POST /api/v1/experiments/{id}/stop
Stop a running experiment.

**Request** (optional):
```json
{
  "reason": "Early success - 75% reduction achieved",
  "rollback": false
}
```

**Response**:
```json
{
  "data": {
    "id": "exp-789",
    "status": "stopping",
    "message": "Experiment stopping, rolling back 15 agents"
  }
}
```

#### GET /api/v1/experiments/{id}/metrics
Get detailed metrics for an experiment.

**Query Parameters**:
- `interval` - Time interval (1m, 5m, 1h, 1d)
- `start` - Start timestamp
- `end` - End timestamp

**Response**:
```json
{
  "data": {
    "summary": {
      "total_baseline_metrics": 45000000,
      "total_candidate_metrics": 13500000,
      "reduction_percentage": 70.0,
      "cost_savings_usd": 1250.50
    },
    "time_series": [
      {
        "timestamp": "2024-01-20T10:00:00Z",
        "baseline": {
          "metrics_per_second": 500000,
          "unique_series": 125000,
          "error_rate": 0.0001
        },
        "candidate": {
          "metrics_per_second": 150000,
          "unique_series": 37500,
          "error_rate": 0.0001
        }
      }
    ]
  }
}
```

#### POST /api/v1/experiments/{id}/promote
Promote experiment winner to production.

**Request**:
```json
{
  "winner": "candidate",
  "rollout_strategy": "gradual",
  "rollout_percentage_per_hour": 10
}
```

**Response**:
```json
{
  "data": {
    "promotion_id": "promo-456",
    "status": "in_progress",
    "target_deployment_count": 150,
    "current_deployment_count": 15
  }
}
```

### Pipelines

#### GET /api/v1/pipelines
List available pipeline templates.

**Response**:
```json
{
  "data": [
    {
      "id": "adaptive-filter-v2",
      "name": "Adaptive Filter v2",
      "type": "adaptive_filter",
      "description": "ML-based metric importance filtering",
      "parameters": {
        "importance_threshold": {
          "type": "float",
          "default": 0.7,
          "min": 0.1,
          "max": 0.99
        }
      }
    },
    {
      "id": "topk-sampler",
      "name": "TopK Sampler",
      "type": "topk",
      "description": "Keep only top K important metrics",
      "parameters": {
        "k": {
          "type": "integer",
          "default": 1000,
          "min": 10,
          "max": 100000
        }
      }
    }
  ]
}
```

#### POST /api/v1/pipelines/validate
Validate a pipeline configuration.

**Request**:
```json
{
  "type": "adaptive_filter",
  "config": {
    "importance_threshold": 0.8,
    "evaluation_interval": "5m",
    "min_sample_size": 1000
  }
}
```

**Response**:
```json
{
  "data": {
    "valid": true,
    "warnings": [
      "High importance_threshold may filter critical metrics"
    ]
  }
}
```

#### POST /api/v1/pipelines/render
Render a pipeline template with parameters.

**Request**:
```json
{
  "template_id": "adaptive-filter-v2",
  "parameters": {
    "importance_threshold": 0.75,
    "namespace_regex": "app_.*"
  },
  "collector_type": "nrdot"  // Optional: "otel" (default) or "nrdot"
}
```

**Response**:
```json
{
  "data": {
    "config": "receivers:\n  otlp:\n    protocols:\n      grpc:\n        endpoint: 0.0.0.0:4317\n\nprocessors:\n  adaptive_filter:\n    importance_threshold: 0.75\n    namespace_regex: app_.*\n\nexporters:\n  prometheus:\n    endpoint: 0.0.0.0:8889\n\nservice:\n  pipelines:\n    metrics:\n      receivers: [otlp]\n      processors: [adaptive_filter]\n      exporters: [prometheus]"
  }
}
```

For NRDOT collector, the config includes New Relic-specific exporters:
```yaml
exporters:
  nrdot:
    endpoint: ${NRDOT_OTLP_ENDPOINT}
    headers:
      api-key: ${NEW_RELIC_LICENSE_KEY}
```

### Pipeline Deployments

#### GET /api/v1/pipelines/deployments
List all pipeline deployments.

**Query Parameters**:
- `experiment_id` - Filter by experiment
- `status` - Filter by status
- `agent_id` - Filter by agent

**Response**:
```json
{
  "data": [
    {
      "id": "dep-222",
      "pipeline_id": "adaptive-filter-v2",
      "experiment_id": "exp-789",
      "variant": "candidate",
      "status": "active",
      "agent_count": 15,
      "metrics": {
        "metrics_per_second": 150000,
        "cardinality_reduction": 0.70,
        "error_rate": 0.0001
      },
      "created_at": "2024-01-20T10:00:00Z"
    }
  ]
}
```

#### POST /api/v1/pipelines/deployments
Create a new deployment (usually done automatically).

**Request**:
```json
{
  "pipeline_id": "adaptive-filter-v2",
  "experiment_id": "exp-789",
  "variant": "candidate",
  "target_agents": ["agent-1", "agent-2"],
  "config_overrides": {
    "importance_threshold": 0.8
  }
}
```

### Agent Operations

#### GET /api/v1/agent/tasks
Poll for pending tasks (Agent endpoint).

**Headers**:
```
X-Agent-Host-ID: agent-hostname-123
```

**Query Parameters**:
- `timeout` - Long-polling timeout in seconds (max: 30)

**Response**:
```json
{
  "data": {
    "task": {
      "id": "task-999",
      "type": "deploy_pipeline",
      "priority": 1,
      "payload": {
        "deployment_id": "dep-222",
        "pipeline_config": "receivers:\n  otlp:...",
        "rollback_on_error": true
      },
      "created_at": "2024-01-20T10:00:00Z"
    }
  }
}
```

#### POST /api/v1/agent/heartbeat
Send agent heartbeat.

**Headers**:
```
X-Agent-Host-ID: agent-hostname-123
```

**Request**:
```json
{
  "status": "healthy",
  "version": "1.2.3",
  "uptime_seconds": 86400,
  "collector_type": "nrdot",  // "otel" or "nrdot"
  "collector_version": "1.0.0",
  "metrics": {
    "cpu_percent": 45.2,
    "memory_percent": 62.1,
    "metrics_per_second": 125000
  }
}
```

**Response**:
```json
{
  "data": {
    "acknowledged": true,
    "server_time": "2024-01-20T10:00:00Z"
  }
}
```

#### POST /api/v1/agent/metrics
Report collected metrics.

**Headers**:
```
X-Agent-Host-ID: agent-hostname-123
```

**Request**:
```json
{
  "deployment_id": "dep-222",
  "interval_start": "2024-01-20T10:00:00Z",
  "interval_end": "2024-01-20T10:05:00Z",
  "metrics": {
    "total_series": 37500,
    "total_samples": 2250000,
    "dropped_samples": 5250000,
    "error_count": 12,
    "processing_time_ms": 4521
  }
}
```

### Cost Analysis

#### GET /api/v1/cost-flow
Get real-time cost flow data.

**Response**:
```json
{
  "data": {
    "current_mps": 1250000,
    "baseline_mps": 1250000,
    "optimized_mps": 375000,
    "reduction_percentage": 70.0,
    "cost_per_million_metrics": 0.50,
    "savings_per_hour": 262.50,
    "active_optimizations": [
      {
        "name": "Adaptive Filter - Apps",
        "reduction": 0.75,
        "metrics_filtered": 937500
      }
    ]
  }
}
```

#### GET /api/v1/analysis/summary
Get platform-wide analysis summary.

**Query Parameters**:
- `period` - Time period (24h, 7d, 30d)

**Response**:
```json
{
  "data": {
    "total_experiments": 145,
    "successful_experiments": 132,
    "average_reduction": 0.68,
    "total_cost_saved_usd": 45250.00,
    "top_performing_pipelines": [
      {
        "id": "adaptive-filter-v2",
        "avg_reduction": 0.72,
        "deployment_count": 89
      }
    ]
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Missing or invalid authentication |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `VALIDATION_ERROR` | Invalid request parameters |
| `CONFLICT` | Resource state conflict |
| `RATE_LIMITED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

## Rate Limiting

API requests are rate limited per authenticated user:
- **Standard endpoints**: 1000 requests/minute
- **Agent endpoints**: 10000 requests/minute
- **WebSocket connections**: 100 concurrent per user

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1642680000
```

## Pagination

List endpoints support pagination:

```
GET /api/v1/experiments?limit=20&offset=40
```

Response includes pagination metadata:
```json
{
  "meta": {
    "total": 145,
    "limit": 20,
    "offset": 40,
    "has_next": true,
    "has_previous": true
  }
}
```