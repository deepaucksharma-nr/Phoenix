# Phoenix Platform - Experiment Simulation Guide

This guide demonstrates how to simulate Phoenix Platform experiment workflows using the API.

## Prerequisites

- Phoenix API running on `http://localhost:8080`
- `curl` and `jq` installed for API interactions
- Basic understanding of Phoenix experiments

## Simulation Steps

### Step 1: View Current Optimization Experiments

Check all active experiments and their cost savings:

```bash
curl -s http://localhost:8080/api/v1/experiments | jq '.experiments[] | {id, name, status, savings: .cost_saving_percent}'
```

### Step 2: Check Platform-Wide Metrics

Get overall platform optimization metrics:

```bash
curl -s http://localhost:8080/api/v1/metrics | jq .
```

### Step 3: Get Experiment Details

Retrieve detailed information about a specific experiment:

```bash
curl -s http://localhost:8080/api/v1/experiments/exp-001 | jq .
```

### Step 4: Monitor Real-Time Performance

Monitor optimization performance in real-time:

```bash
# Watch metrics updates
watch -n 5 'curl -s http://localhost:8080/api/v1/metrics | jq ".processed_metrics, .cost_reduction_percent"'
```

### Step 5: Calculate Projected Savings

Calculate annual savings based on current performance:

```bash
# Get monthly savings
monthly_savings=$(curl -s http://localhost:8080/api/v1/metrics | jq -r '.monthly_savings_usd')

# Calculate annual projection
annual_savings=$((monthly_savings * 12))
echo "Projected Annual Savings: \$$annual_savings"
```

### Step 6: Review Optimization Recommendations

Phoenix provides AI-powered recommendations based on current metrics:

- **Cardinality Reduction**: Increase threshold to 90% for maximum savings
- **Adaptive Sampling**: Enable for high-volume metrics
- **Tag Consolidation**: Merge similar Kubernetes labels
- **Intelligent Aggregation**: Activate for time-series optimization

## Expected Results

A successful simulation should show:

- **Cardinality Reduction**: 85-90% reduction in unique metric series
- **Cost Savings**: 70-80% reduction in observability costs
- **Performance**: No impact on critical metric visibility
- **Data Quality**: Maintained accuracy for important metrics

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/experiments` | GET | List all experiments |
| `/api/v1/experiments/{id}` | GET | Get experiment details |
| `/api/v1/metrics` | GET | Get platform metrics |
| `/api/v1/experiments` | POST | Create new experiment |
| `/api/v1/experiments/{id}/status` | PUT | Update experiment status |

## WebSocket Real-Time Updates

Connect to WebSocket for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Real-time update:', data);
};
```

## Troubleshooting

1. **API Connection Failed**: Ensure Phoenix API is running on port 8080
2. **No Experiments Found**: Create an experiment first using the API
3. **Metrics Not Updating**: Check that agents are connected and sending data

## Next Steps

- Create a real experiment using the Phoenix CLI
- Deploy agents to collect actual metrics
- Monitor cost savings in production