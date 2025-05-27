# Phoenix Platform - Complete Experiment Workflow Guide

This guide demonstrates the complete lifecycle of a Phoenix experiment from creation to completion.

## Prerequisites

- Phoenix API running on `http://localhost:8080`
- `curl` and `jq` installed
- Optional: `websocat` for WebSocket monitoring

## Experiment Lifecycle

### 1. Create a New Experiment

Create an experiment to compare baseline and optimized pipelines:

```bash
curl -X POST "http://localhost:8080/api/v1/experiments" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prometheus-optimization-experiment",
    "description": "Optimize Prometheus metric collection using intelligent sampling",
    "baseline_pipeline": "prometheus-baseline",
    "candidate_pipeline": "prometheus-optimized",
    "target_nodes": {
      "prometheus": "prometheus-0",
      "collector": "otel-collector-0"
    }
  }'
```

**Response:**
```json
{
  "id": "exp-123456",
  "name": "prometheus-optimization-experiment",
  "status": "created",
  "created_at": "2025-01-27T10:00:00Z"
}
```

### 2. Start the Experiment

Change experiment status to running:

```bash
curl -X PUT "http://localhost:8080/api/v1/experiments/exp-123456/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "running"}'
```

### 3. Monitor Metrics

During the experiment, metrics are collected from both pipelines:

**Expected Metrics:**
- **CPU Usage**: Baseline ~60% → Candidate ~45%
- **Memory Usage**: Baseline ~65% → Candidate ~50%
- **Cardinality**: Baseline ~12,000 → Candidate ~4,000

Monitor in real-time:
```bash
# Poll metrics endpoint
watch -n 5 'curl -s http://localhost:8080/api/v1/experiments/exp-123456/metrics | jq .'
```

### 4. Analyze Results

After sufficient data collection, analyze the results:

```bash
curl -s "http://localhost:8080/api/v1/experiments/exp-123456/analysis" | jq .
```

**Typical Analysis Results:**
```json
{
  "cost_reduction_percent": 75,
  "cardinality_reduction_percent": 67,
  "performance_impact_percent": 0.5,
  "recommendation": "PROMOTE",
  "confidence_score": 0.95
}
```

### 5. Complete the Experiment

Mark the experiment as completed:

```bash
curl -X PUT "http://localhost:8080/api/v1/experiments/exp-123456/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

### 6. List All Experiments

View all experiments and their results:

```bash
curl -s "http://localhost:8080/api/v1/experiments" | jq -r '.[] | "- \(.name) [\(.status)] - Cost Saving: \(.cost_saving_percent // 0)%"'
```

## WebSocket Real-Time Monitoring

Connect to WebSocket for real-time updates:

### Using websocat

```bash
# Install websocat
brew install websocat

# Connect to WebSocket
websocat ws://localhost:8080/ws
```

### Subscribe to Updates

```json
// Subscribe to experiment updates
{"type":"subscribe","data":{"topic":"experiment:exp-123456"}}

// Subscribe to metrics stream
{"type":"subscribe","data":{"topic":"metrics:exp-123456"}}

// Subscribe to all experiments
{"type":"subscribe","data":{"topic":"experiments:*"}}
```

### Example WebSocket Messages

```json
// Experiment status update
{
  "type": "experiment.status",
  "data": {
    "experiment_id": "exp-123456",
    "status": "running",
    "phase": "collecting_baseline"
  }
}

// Metrics update
{
  "type": "metrics.update",
  "data": {
    "experiment_id": "exp-123456",
    "timestamp": "2025-01-27T10:05:00Z",
    "baseline": {
      "cpu_percent": 62,
      "memory_percent": 68,
      "cardinality": 12543
    },
    "candidate": {
      "cpu_percent": 43,
      "memory_percent": 52,
      "cardinality": 3891
    }
  }
}
```

## Automation Script

Create a complete workflow script:

```bash
#!/bin/bash
# experiment-automation.sh

API_URL="http://localhost:8080"
EXPERIMENT_NAME="auto-experiment-$(date +%s)"

# Create experiment
echo "Creating experiment..."
RESPONSE=$(curl -s -X POST "$API_URL/api/v1/experiments" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$EXPERIMENT_NAME\",
    \"description\": \"Automated experiment\",
    \"baseline_pipeline\": \"baseline\",
    \"candidate_pipeline\": \"optimized\"
  }")

EXPERIMENT_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Created experiment: $EXPERIMENT_ID"

# Start experiment
echo "Starting experiment..."
curl -s -X PUT "$API_URL/api/v1/experiments/$EXPERIMENT_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "running"}'

# Monitor for 60 seconds
echo "Monitoring experiment for 60 seconds..."
sleep 60

# Get results
echo "Fetching results..."
curl -s "$API_URL/api/v1/experiments/$EXPERIMENT_ID/analysis" | jq .

# Complete experiment
echo "Completing experiment..."
curl -s -X PUT "$API_URL/api/v1/experiments/$EXPERIMENT_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'

echo "Experiment complete!"
```

## Best Practices

1. **Experiment Duration**: Run experiments for at least 1 hour for statistical significance
2. **Baseline Selection**: Ensure baseline represents typical production load
3. **Monitoring**: Use WebSocket for real-time monitoring during experiments
4. **Analysis**: Wait for sufficient data before making promotion decisions
5. **Rollback**: Always have a rollback plan if candidate performs poorly

## Troubleshooting

- **Experiment Stuck**: Check agent connectivity and pipeline status
- **No Metrics**: Verify collectors are running and connected
- **Analysis Failed**: Ensure minimum data collection period has elapsed
- **WebSocket Disconnects**: Check for network issues or API restarts