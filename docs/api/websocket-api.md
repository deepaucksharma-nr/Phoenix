# Phoenix WebSocket API Reference

## Overview

The Phoenix WebSocket API provides real-time updates for experiments, metrics, and system events. WebSocket connections are available on the same port as the REST API.

**WebSocket URL**: `ws://localhost:8080/ws`

## Connection

### Authentication

Include the JWT token as a query parameter when establishing the connection:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=<jwt-token>');
```

### Connection Lifecycle

```javascript
// Connection established
ws.onopen = (event) => {
  console.log('Connected to Phoenix WebSocket');
  
  // Subscribe to events
  ws.send(JSON.stringify({
    type: 'subscribe',
    channels: ['experiments', 'metrics', 'alerts']
  }));
};

// Handle messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

// Connection closed
ws.onclose = (event) => {
  console.log('Disconnected:', event.code, event.reason);
};

// Handle errors
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

## Message Format

### Client to Server

```json
{
  "id": "msg-123",
  "type": "subscribe|unsubscribe|ping",
  "channels": ["experiments", "metrics"],
  "filters": {
    "experiment_id": "exp-789"
  }
}
```

### Server to Client

```json
{
  "id": "msg-123",
  "type": "event|response|error",
  "channel": "experiments",
  "event": "experiment.updated",
  "data": {},
  "timestamp": "2024-01-20T10:00:00Z"
}
```

## Channels

### experiments

Real-time updates for experiment lifecycle events.

#### Events

**experiment.created**
```json
{
  "type": "event",
  "channel": "experiments",
  "event": "experiment.created",
  "data": {
    "id": "exp-789",
    "name": "Reduce app metrics cardinality",
    "status": "created",
    "baseline_pipeline_id": "pipeline-123",
    "candidate_pipeline_id": "pipeline-456"
  }
}
```

**experiment.started**
```json
{
  "type": "event",
  "channel": "experiments",
  "event": "experiment.started",
  "data": {
    "id": "exp-789",
    "status": "running",
    "start_time": "2024-01-20T10:00:00Z",
    "deployment_count": 90
  }
}
```

**experiment.updated**
```json
{
  "type": "event",
  "channel": "experiments",
  "event": "experiment.updated",
  "data": {
    "id": "exp-789",
    "status": "running",
    "progress": 0.45,
    "metrics": {
      "baseline_mps": 500000,
      "candidate_mps": 150000,
      "reduction_rate": 0.70
    }
  }
}
```

**experiment.completed**
```json
{
  "type": "event",
  "channel": "experiments",
  "event": "experiment.completed",
  "data": {
    "id": "exp-789",
    "status": "completed",
    "winner": "candidate",
    "final_metrics": {
      "total_reduction": 0.72,
      "cost_saved_usd": 1250.50,
      "duration_hours": 24
    }
  }
}
```

**experiment.failed**
```json
{
  "type": "event",
  "channel": "experiments",
  "event": "experiment.failed",
  "data": {
    "id": "exp-789",
    "status": "failed",
    "error": "High error rate detected",
    "rollback_initiated": true
  }
}
```

### metrics

Real-time metrics updates aggregated at configurable intervals.

#### Events

**metrics.update**
```json
{
  "type": "event",
  "channel": "metrics",
  "event": "metrics.update",
  "data": {
    "timestamp": "2024-01-20T10:00:00Z",
    "global": {
      "total_mps": 1250000,
      "total_series": 312500,
      "active_agents": 148
    },
    "experiments": {
      "exp-789": {
        "baseline_mps": 500000,
        "candidate_mps": 150000,
        "reduction_rate": 0.70,
        "error_rate": 0.0001
      }
    }
  }
}
```

**metrics.anomaly**
```json
{
  "type": "event",
  "channel": "metrics",
  "event": "metrics.anomaly",
  "data": {
    "experiment_id": "exp-789",
    "type": "sudden_drop",
    "severity": "warning",
    "details": {
      "metric": "candidate_mps",
      "expected_range": [140000, 160000],
      "actual_value": 50000
    }
  }
}
```

### agents

Agent status and health updates.

#### Events

**agent.connected**
```json
{
  "type": "event",
  "channel": "agents",
  "event": "agent.connected",
  "data": {
    "agent_id": "agent-host-123",
    "hostname": "prod-app-01",
    "version": "1.2.3",
    "capabilities": ["adaptive_filter", "topk"]
  }
}
```

**agent.heartbeat**
```json
{
  "type": "event",
  "channel": "agents",
  "event": "agent.heartbeat",
  "data": {
    "agent_id": "agent-host-123",
    "status": "healthy",
    "metrics": {
      "cpu_percent": 45.2,
      "memory_percent": 62.1,
      "metrics_per_second": 125000
    }
  }
}
```

**agent.disconnected**
```json
{
  "type": "event",
  "channel": "agents",
  "event": "agent.disconnected",
  "data": {
    "agent_id": "agent-host-123",
    "last_seen": "2024-01-20T09:55:00Z",
    "tasks_reassigned": 2
  }
}
```

### deployments

Pipeline deployment status updates.

#### Events

**deployment.started**
```json
{
  "type": "event",
  "channel": "deployments",
  "event": "deployment.started",
  "data": {
    "deployment_id": "dep-222",
    "pipeline_id": "adaptive-filter-v2",
    "agent_count": 15,
    "variant": "candidate"
  }
}
```

**deployment.progress**
```json
{
  "type": "event",
  "channel": "deployments",
  "event": "deployment.progress",
  "data": {
    "deployment_id": "dep-222",
    "completed": 12,
    "total": 15,
    "failed": 0
  }
}
```

**deployment.completed**
```json
{
  "type": "event",
  "channel": "deployments",
  "event": "deployment.completed",
  "data": {
    "deployment_id": "dep-222",
    "status": "active",
    "duration_seconds": 45,
    "agent_count": 15
  }
}
```

### alerts

System alerts and notifications.

#### Events

**alert.triggered**
```json
{
  "type": "event",
  "channel": "alerts",
  "event": "alert.triggered",
  "data": {
    "alert_id": "alert-456",
    "name": "High Error Rate",
    "severity": "critical",
    "experiment_id": "exp-789",
    "condition": "error_rate > 0.01",
    "current_value": 0.015
  }
}
```

**alert.resolved**
```json
{
  "type": "event",
  "channel": "alerts",
  "event": "alert.resolved",
  "data": {
    "alert_id": "alert-456",
    "resolved_at": "2024-01-20T10:05:00Z",
    "duration_minutes": 5
  }
}
```

### cost-flow

Real-time cost analysis updates.

#### Events

**cost.update**
```json
{
  "type": "event",
  "channel": "cost-flow",
  "event": "cost.update",
  "data": {
    "timestamp": "2024-01-20T10:00:00Z",
    "baseline_cost_per_hour": 375.00,
    "optimized_cost_per_hour": 112.50,
    "savings_per_hour": 262.50,
    "active_optimizations": [
      {
        "experiment_id": "exp-789",
        "contribution_usd": 125.00
      }
    ]
  }
}
```

## Commands

### Subscribe to Channels

```json
{
  "type": "subscribe",
  "channels": ["experiments", "metrics"],
  "filters": {
    "experiment_id": "exp-789"
  }
}
```

**Response**:
```json
{
  "type": "response",
  "status": "ok",
  "subscribed": ["experiments", "metrics"]
}
```

### Unsubscribe from Channels

```json
{
  "type": "unsubscribe",
  "channels": ["alerts"]
}
```

### Ping/Pong

Keep connection alive:

```json
{
  "type": "ping"
}
```

**Response**:
```json
{
  "type": "pong",
  "timestamp": "2024-01-20T10:00:00Z"
}
```

## Error Handling

### Error Response Format

```json
{
  "type": "error",
  "error": {
    "code": "INVALID_CHANNEL",
    "message": "Channel 'invalid' does not exist",
    "details": {
      "valid_channels": ["experiments", "metrics", "agents", "deployments", "alerts", "cost-flow"]
    }
  }
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Invalid or expired token |
| `INVALID_MESSAGE` | Malformed message format |
| `INVALID_CHANNEL` | Unknown channel name |
| `RATE_LIMITED` | Too many messages |
| `INTERNAL_ERROR` | Server error |

## Connection Management

### Reconnection Strategy

```javascript
class PhoenixWebSocket {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.reconnectDelay = 1000;
    this.maxReconnectDelay = 30000;
    this.reconnectAttempts = 0;
  }

  connect() {
    this.ws = new WebSocket(`${this.url}?token=${this.token}`);
    
    this.ws.onopen = () => {
      console.log('Connected');
      this.reconnectDelay = 1000;
      this.reconnectAttempts = 0;
    };
    
    this.ws.onclose = () => {
      console.log('Disconnected, reconnecting...');
      this.scheduleReconnect();
    };
  }

  scheduleReconnect() {
    setTimeout(() => {
      this.reconnectAttempts++;
      this.connect();
    }, Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts), this.maxReconnectDelay));
  }
}
```

### Rate Limiting

- **Messages per second**: 100 per connection
- **Subscriptions**: 10 channels per connection
- **Connection limit**: 100 concurrent per user

## Best Practices

1. **Subscribe Selectively**: Only subscribe to channels you need
2. **Use Filters**: Apply filters to reduce message volume
3. **Handle Reconnection**: Implement exponential backoff
4. **Process Asynchronously**: Don't block on message processing
5. **Monitor Connection**: Track connection health metrics

## Example Implementation

```javascript
// React Hook Example
import { useEffect, useState, useCallback } from 'react';

export function usePhoenixWebSocket(token) {
  const [connected, setConnected] = useState(false);
  const [messages, setMessages] = useState([]);
  const [ws, setWs] = useState(null);

  useEffect(() => {
    const websocket = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
    
    websocket.onopen = () => {
      setConnected(true);
      websocket.send(JSON.stringify({
        type: 'subscribe',
        channels: ['experiments', 'metrics']
      }));
    };

    websocket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      setMessages(prev => [...prev, message]);
    };

    websocket.onclose = () => {
      setConnected(false);
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, [token]);

  const subscribe = useCallback((channels) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'subscribe',
        channels
      }));
    }
  }, [ws]);

  return { connected, messages, subscribe };
}
```