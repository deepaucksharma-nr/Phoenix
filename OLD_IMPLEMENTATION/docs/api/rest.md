# REST API Reference

The Phoenix Platform REST API provides HTTP endpoints for managing experiments, pipelines, and metrics.

## Base URL

```
https://api.phoenix.example.com/v1
```

## Authentication

All API requests require JWT authentication:

```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  https://api.phoenix.example.com/v1/experiments
```

### Obtaining a Token

=== "cURL"

    ```bash
    curl -X POST https://api.phoenix.example.com/v1/auth/login \
      -H "Content-Type: application/json" \
      -d '{
        "username": "user@example.com",
        "password": "password"
      }'
    ```

=== "JavaScript"

    ```javascript
    const response = await fetch('https://api.phoenix.example.com/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: 'user@example.com',
        password: 'password'
      })
    });
    const { token } = await response.json();
    ```

=== "Python"

    ```python
    import requests
    
    response = requests.post(
        'https://api.phoenix.example.com/v1/auth/login',
        json={
            'username': 'user@example.com',
            'password': 'password'
        }
    )
    token = response.json()['token']
    ```

## Experiments

### List Experiments

<div class="api-endpoint">
  <span class="method get">GET</span>
  <code>/v1/experiments</code>
</div>

List all experiments with optional filtering.

#### Query Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `status` | string | Filter by status (running, completed, failed) | - |
| `limit` | integer | Number of results | 20 |
| `offset` | integer | Pagination offset | 0 |

#### Response

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

### Create Experiment

<div class="api-endpoint">
  <span class="method post">POST</span>
  <code>/v1/experiments</code>
</div>

Create a new optimization experiment.

#### Request Body

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
  ]
}
```

#### Response

!!! success "201 Created"
    ```json
    {
      "id": "exp-456",
      "name": "database-optimization-test",
      "status": "pending",
      "created_at": "2024-11-24T12:00:00Z"
    }
    ```

### Get Experiment

<div class="api-endpoint">
  <span class="method get">GET</span>
  <code>/v1/experiments/{id}</code>
</div>

Get detailed information about a specific experiment.

#### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Experiment ID |

#### Response

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

## Pipelines

### List Pipeline Templates

<div class="api-endpoint">
  <span class="method get">GET</span>
  <code>/v1/pipelines</code>
</div>

Get available pipeline optimization templates.

#### Response

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

## Metrics

### Get Experiment Metrics

<div class="api-endpoint">
  <span class="method get">GET</span>
  <code>/v1/experiments/{id}/metrics</code>
</div>

Retrieve time-series metrics for an experiment.

#### Query Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `start` | string | Start time (RFC3339) | 1 hour ago |
| `end` | string | End time (RFC3339) | now |
| `resolution` | string | Data resolution (1m, 5m, 1h) | 5m |

#### Response

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

## Error Responses

All errors follow a consistent format:

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

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_ARGUMENT` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `PERMISSION_DENIED` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `ALREADY_EXISTS` | Resource already exists |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL` | Internal server error |

## Rate Limiting

API requests are rate limited:

- **Authenticated**: 1000 requests/hour
- **Unauthenticated**: 100 requests/hour

Rate limit information is included in response headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1701234567
```

<style>
.api-endpoint {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 1rem 0;
  padding: 0.5rem;
  background: var(--md-code-bg-color);
  border-radius: 4px;
}

.method {
  font-weight: bold;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  color: white;
  font-size: 0.875rem;
}

.method.get { background: #61affe; }
.method.post { background: #49cc90; }
.method.put { background: #fca130; }
.method.delete { background: #f93e3e; }
.method.patch { background: #50e3c2; }
</style>