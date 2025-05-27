# Phoenix Platform API Documentation v2

## Overview

Phoenix Platform API provides a unified control plane for observability cost optimization through agent-based task distribution. The API follows REST principles with JSON payloads and includes WebSocket support for real-time updates on the same port.

## Base URLs

- **API**: `http://localhost:8080/api/v2`
- **WebSocket**: `ws://localhost:8080/ws` (same port as REST API)

## Table of Contents

- [Authentication](#authentication)
- [Experiments API](#experiments-api)
- [Fleet Management API](#fleet-management-api)
- [Pipeline API](#pipeline-api)
- [Metrics & Analytics API](#metrics--analytics-api)
- [Agent API](#agent-api)
- [WebSocket Events](#websocket-events)
- [Error Handling](#error-handling)

---

## Authentication

Agent endpoints require the `X-Agent-Host-ID` header for authentication. User-facing endpoints use JWT tokens (optional in development).

```bash
# Agent authentication with long-polling
curl -H "X-Agent-Host-ID: agent-001" http://localhost:8080/api/v2/tasks/poll
```

---

## Experiments API

### Create Experiment (Wizard)

**POST** `/api/v2/experiments/wizard`

Simplified experiment creation for UI.

```json
{
  "name": "Reduce API costs Q1",
  "description": "Optimize high-traffic API metrics",
  "host_selector": ["group:prod-api", "env=production"],
  "baseline_template": "standard",
  "candidate_template": "adaptive-filter-v1",
  "duration_hours": 24,
  "collector_type": "otel"
}
```

**With NRDOT Collector:**
```json
{
  "name": "NRDOT Cardinality Reduction",
  "description": "Use New Relic's advanced cardinality reduction",
  "host_selector": ["group:prod-api"],
  "baseline_template": "standard",
  "candidate_template": "nrdot-cardinality",
  "duration_hours": 24,
  "collector_type": "nrdot",
  "nrdot_config": {
    "license_key": "your-nr-license-key",
    "otlp_endpoint": "otlp.nr-data.net:4317",
    "max_cardinality": 10000,
    "reduction_percentage": 70
  }
}
```

**Response**: `201 Created`
```json
{
  "id": "exp-123",
  "name": "Reduce API costs Q1",
  "status": "pending",
  "created_at": "2024-01-15T10:00:00Z",
  "baseline_template": "standard",
  "candidate_template": "adaptive-filter-v1",
  "estimated_savings_percent": 70,
  "collector_type": "otel"
}
```

### List Experiments

**GET** `/api/v2/experiments`

**Response**: `200 OK`
```json
[
  {
    "id": "exp-123",
    "name": "Reduce API costs Q1",
    "status": "running",
    "phase": "baseline",
    "baseline_cost": 5000,
    "candidate_cost": 3000,
    "savings_percent": 40,
    "created_at": "2024-01-15T10:00:00Z"
  }
]
```

### Get Experiment Details

**GET** `/api/v2/experiments/{id}`

### Start Experiment

**POST** `/api/v2/experiments/{id}/start`

### Stop Experiment

**POST** `/api/v2/experiments/{id}/stop`

### Promote Experiment

**POST** `/api/v2/experiments/{id}/promote`

Promotes candidate pipeline to production.

### Instant Rollback

**POST** `/api/v2/experiments/{id}/rollback`

Immediately rolls back to baseline configuration.

**Response**: `200 OK`
```json
{
  "status": "success",
  "message": "Rollback initiated",
  "affected_hosts": 45
}
```

---

## Fleet Management API

### Get Fleet Status

**GET** `/api/v2/agents`

Returns comprehensive agent fleet status.

**Response**: `200 OK`
```json
{
  "total_agents": 150,
  "healthy_agents": 145,
  "offline_agents": 2,
  "updating_agents": 3,
  "total_savings": 125000,
  "agents": [
    {
      "host_id": "prod-api-001",
      "status": "healthy",
      "group": "prod-api",
      "active_tasks": [],
      "metrics": {
        "cpu_percent": 12.5,
        "memory_mb": 256,
        "metrics_per_sec": 45000,
        "dropped_count": 0
      },
      "cost_savings": 2500,
      "last_heartbeat": "2024-01-15T10:30:00Z",
      "location": {
        "region": "us-east",
        "zone": "us-east-1a"
      }
    }
  ]
}
```

### Get Agent Map

**GET** `/api/v2/agents/map`

Returns agents with geographical location for map visualization.

---

## Pipeline API

### List Pipeline Templates

**GET** `/api/v2/pipeline-templates`

**Response**: `200 OK`
```json
[
  {
    "id": "adaptive-filter-v1",
    "name": "Adaptive Filter",
    "description": "Dynamically filters low-value metrics",
    "category": "cost_optimization",
    "config_url": "/configs/adaptive-filter-v1.yaml",
    "estimated_savings_percent": 70,
    "collector_type": "otel"
  },
  {
    "id": "topk-v1",
    "name": "Top-K Filter",
    "description": "Keep only top K metrics by value",
    "category": "cost_optimization",
    "config_url": "/configs/topk-v1.yaml",
    "estimated_savings_percent": 65,
    "collector_type": "otel"
  },
  {
    "id": "hybrid-v1",
    "name": "Hybrid Optimization",
    "description": "Combines multiple filtering techniques",
    "category": "advanced",
    "config_url": "/configs/hybrid-v1.yaml",
    "estimated_savings_percent": 75,
    "collector_type": "otel"
  },
  {
    "id": "nrdot-baseline",
    "name": "NRDOT Baseline",
    "description": "New Relic collector with standard configuration",
    "category": "new_relic",
    "config_url": "/configs/nrdot-baseline.yaml",
    "estimated_savings_percent": 0,
    "collector_type": "nrdot",
    "requirements": {
      "license_key": true,
      "min_version": "1.0.0"
    }
  },
  {
    "id": "nrdot-cardinality",
    "name": "NRDOT Cardinality Reduction",
    "description": "New Relic collector with advanced cardinality reduction",
    "category": "new_relic",
    "config_url": "/configs/nrdot-cardinality.yaml",
    "estimated_savings_percent": 80,
    "collector_type": "nrdot",
    "requirements": {
      "license_key": true,
      "min_version": "1.0.0"
    }
  }
]
```

### Preview Pipeline Impact

**POST** `/api/v2/pipelines/preview`

Calculate impact without deploying.

```json
{
  "pipeline_config": {
    "processors": [
      {"type": "top_k", "config": {"k": 20}}
    ]
  },
  "target_hosts": ["group:prod-api"]
}
```

**Response**: `200 OK`
```json
{
  "estimated_cost_reduction": 65.5,
  "estimated_cardinality_reduction": 72.3,
  "estimated_cpu_impact": 1.2,
  "estimated_memory_impact": 45,
  "confidence_level": 0.85
}
```

### Quick Deploy Pipeline

**POST** `/api/v2/pipeline-deployments`

One-click pipeline deployment.

```json
{
  "pipeline_template_id": "adaptive-filter-v1",
  "target_hosts": ["group:prod-api"],
  "variant": "candidate",
  "experiment_id": "exp-123"
}
```

**Response**: `202 Accepted`
```json
{
  "deployment_id": "dep-456",
  "hosts_count": 45,
  "status": "deploying"
}
```

---

## Metrics & Analytics API

### Get Metric Cost Flow

**GET** `/api/v2/metrics/cost-flow`

Real-time metric cost breakdown.

**Response**: `200 OK`
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "total_cost_rate": 125.50,
  "top_metrics": [
    {
      "metric_name": "process.cpu.usage",
      "cost_per_minute": 45.25,
      "cardinality": 125000,
      "percentage": 36.1,
      "labels": {
        "service": "api-gateway",
        "env": "production"
      }
    }
  ],
  "by_service": {
    "api-gateway": 45.25,
    "auth-service": 23.10
  },
  "by_namespace": {
    "production": 98.50,
    "staging": 27.00
  }
}
```

### Get Cardinality Breakdown

**GET** `/api/v2/metrics/cardinality?namespace=production&service=api`

### Get Cost Analytics

**GET** `/api/v2/experiments/{id}/kpis`

Returns calculated KPIs including cardinality reduction and cost savings.

---

## Agent API

These endpoints are used by Phoenix agents.

### Poll for Tasks (Long-poll)

**GET** `/api/v2/tasks/poll`

Headers: `X-Agent-Host-ID: agent-001`

Long-polling endpoint with 30-second timeout. Returns immediately if tasks are available, otherwise waits up to 30 seconds.

**Response**: `200 OK`
```json
[
  {
    "id": "task-789",
    "type": "deploy_pipeline",
    "experiment_id": "exp-123",
    "config": {
      "pipeline_url": "http://api/configs/top-k-20.yaml",
      "variant": "candidate"
    },
    "priority": 1
  }
]
```

### Update Task Status

**POST** `/api/v2/tasks/{taskId}/status`

```json
{
  "status": "completed",
  "message": "Pipeline deployed successfully",
  "metrics": {
    "deployment_time_ms": 1250
  }
}
```

### Send Heartbeat

**POST** `/api/v2/agents/{hostId}/heartbeat`

```json
{
  "status": "healthy",
  "metrics": {
    "cpu_percent": 12.5,
    "memory_mb": 256,
    "metrics_per_sec": 45000
  },
  "active_pipelines": ["baseline", "candidate"],
  "agent_version": "1.0.0",
  "collector_info": {
    "type": "otel",
    "version": "0.91.0",
    "status": "running"
  }
}
```

**With NRDOT Collector:**
```json
{
  "status": "healthy",
  "metrics": {
    "cpu_percent": 12.5,
    "memory_mb": 256,
    "metrics_per_sec": 45000,
    "cardinality_reduction": 72.5
  },
  "active_pipelines": ["nrdot-cardinality"],
  "agent_version": "1.0.0",
  "collector_info": {
    "type": "nrdot",
    "version": "1.0.0",
    "status": "running",
    "new_relic_account": "123456"
  }
}
```

---

## WebSocket Events

Connect to `ws://localhost:8080/ws` for real-time updates. WebSocket runs on the same port as the REST API.

### Event Types

#### agent_status
```json
{
  "type": "agent_status",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "host_id": "prod-api-001",
    "status": "healthy",
    "metrics": {
      "cpu_percent": 12.5,
      "metrics_per_sec": 45000
    }
  }
}
```

#### experiment_update
```json
{
  "type": "experiment_update",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "experiment_id": "exp-123",
    "status": "running",
    "progress": 65,
    "metrics": {
      "baseline_cost": 5000,
      "candidate_cost": 3000,
      "savings_percent": 40
    }
  }
}
```

#### metric_flow
```json
{
  "type": "metric_flow",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "total_cost_rate": 125.50,
    "top_metrics": [...]
  }
}
```

#### task_progress
```json
{
  "type": "task_progress",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "task_id": "task-789",
    "progress": 80,
    "total_hosts": 45,
    "completed_hosts": 36
  }
}
```

### Subscribing to Events

Send subscription message after connecting:

```json
{
  "type": "subscribe",
  "payload": {
    "events": ["agent_status", "experiment_update"],
    "filters": {
      "experiments": ["exp-123"],
      "hosts": ["prod-api-*"]
    }
  }
}
```

---

## Task Management API

### Get Active Tasks

**GET** `/api/v2/tasks?status=running&limit=100`

### Get Task Queue Status

**GET** `/api/v2/tasks/queue`

**Response**: `200 OK`
```json
{
  "pending_tasks": 12,
  "running_tasks": 5,
  "completed_tasks": 145,
  "failed_tasks": 2,
  "tasks_by_type": {
    "deploy_pipeline": 8,
    "start_collector": 4
  },
  "average_wait_time": 2500
}
```

---

## Error Handling

All errors follow consistent format:

```json
{
  "error": "invalid_request",
  "message": "Pipeline template not found",
  "details": {
    "template": "unknown-template",
    "available": ["top-k-20", "priority-sli-slo"]
  }
}
```

Common status codes:
- `200 OK`: Success
- `201 Created`: Resource created
- `202 Accepted`: Async operation started
- `400 Bad Request`: Invalid parameters
- `401 Unauthorized`: Missing authentication
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Rate Limiting

- Default: 1000 requests/minute per IP
- Agent endpoints: 10,000 requests/minute
- WebSocket: 100 messages/second

Headers:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`

---

## SDK Examples

### JavaScript/TypeScript
```typescript
// Connect to WebSocket (same port as API)
const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', () => {
  // Subscribe to events
  ws.send(JSON.stringify({
    type: 'subscribe',
    payload: {
      events: ['metric_flow', 'agent_status']
    }
  }));
});

ws.on('message', (data) => {
  const event = JSON.parse(data);
  console.log(`Event: ${event.type}`, event.data);
});

// Create experiment via API
const response = await fetch('/api/v2/experiments/wizard', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: 'Optimize API costs',
    host_selector: ['env=prod'],
    baseline_template: 'standard',
    candidate_template: 'adaptive-filter-v1',
    duration_hours: 24
  })
});
```

### Python
```python
import requests
import websocket

# Get agent status
response = requests.get('http://localhost:8080/api/v2/agents')
fleet = response.json()
print(f"Healthy agents: {fleet['healthy_agents']}/{fleet['total_agents']}")

# Deploy pipeline
response = requests.post('http://localhost:8080/api/v2/pipeline-deployments',
    json={
        'pipeline_template_id': 'adaptive-filter-v1',
        'target_hosts': ['group:prod-api'],
        'variant': 'candidate'
    })
print(f"Deployment started: {response.json()['deployment_id']}")
```

### Go
```go
// Agent heartbeat
type Heartbeat struct {
    Status  string      `json:"status"`
    Metrics AgentMetrics `json:"metrics"`
}

client := &http.Client{}
data, _ := json.Marshal(Heartbeat{
    Status: "healthy",
    Metrics: AgentMetrics{
        CPUPercent: 12.5,
        MemoryMB: 256,
    },
})

req, _ := http.NewRequest("POST", "http://localhost:8080/api/v2/agents/agent-001/heartbeat", bytes.NewBuffer(data))
req.Header.Set("X-Agent-Host-ID", "agent-001")
req.Header.Set("Content-Type", "application/json")
client.Do(req)
```

---

## Key Implementation Details

- **Task Queue**: PostgreSQL-based with atomic assignment
- **Long-polling**: 30-second timeout for agent task polling
- **A/B Testing**: Baseline vs candidate pipeline comparison
- **Authentication**: X-Agent-Host-ID header for agents
- **WebSocket**: Real-time updates on same port as REST API
- **Pipeline Templates**: Adaptive Filter, TopK, Hybrid, and NRDOT approaches
- **Collector Support**: Both OpenTelemetry and NRDOT (New Relic) collectors
- **Cost Reduction**: Demonstrated 70-80% reduction in metrics cardinality
- **NRDOT Features**: Advanced cardinality reduction with New Relic processors

---

## OpenAPI Specification

Available at: `http://localhost:8080/api/v2/openapi.json`