# NRDOT Collector Integration Guide

This guide explains how to integrate New Relic's NRDOT collector into the Phoenix platform for enhanced cardinality reduction and New Relic-specific optimizations.

## Overview

NRDOT (New Relic Distribution of OpenTelemetry) is New Relic's optimized distribution of the OpenTelemetry Collector. It provides:

- Built-in cardinality reduction capabilities
- Optimized performance for New Relic backends
- Full compatibility with OpenTelemetry configurations
- Enhanced metric processing features

## Architecture

The Phoenix platform supports both standard OpenTelemetry collectors and NRDOT collectors:

```
┌─────────────────┐
│  Phoenix Agent  │
├─────────────────┤
│ Collector Mgr   │
├─────────────────┤
│ ┌─────────────┐ │
│ │   OTel or   │ │
│ │    NRDOT    │ │
│ └─────────────┘ │
└─────────────────┘
```

## Installation

### 1. Agent Installation (System)

When installing Phoenix agents, enable NRDOT support:

```bash
# Install with NRDOT support
export USE_NRDOT=true
export NEW_RELIC_LICENSE_KEY="your-license-key"
export NEW_RELIC_OTLP_ENDPOINT="otlp.nr-data.net:4317"

curl -fsSL https://phoenix.my-org.com/install-agent.sh | sudo bash
```

### 2. Docker Deployment

For Docker deployments, set environment variables:

```yaml
# docker-compose.yml
services:
  phoenix-agent:
    environment:
      - USE_NRDOT=true
      - NEW_RELIC_LICENSE_KEY=${NEW_RELIC_LICENSE_KEY}
      - NEW_RELIC_OTLP_ENDPOINT=otlp.nr-data.net:4317
```

### 3. Kubernetes Deployment

For Kubernetes, update the agent deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: phoenix-agent
spec:
  template:
    spec:
      containers:
      - name: phoenix-agent
        env:
        - name: USE_NRDOT
          value: "true"
        - name: NEW_RELIC_LICENSE_KEY
          valueFrom:
            secretKeyRef:
              name: newrelic-secret
              key: license-key
        - name: NEW_RELIC_OTLP_ENDPOINT
          value: "otlp.nr-data.net:4317"
```

## Configuration

### Pipeline Templates

Phoenix provides built-in NRDOT pipeline templates:

1. **nrdot-baseline**: Basic NRDOT configuration
2. **nrdot-cardinality**: NRDOT with cardinality reduction enabled

### Example: NRDOT Cardinality Reduction

```yaml
processors:
  newrelic/cardinality:
    enabled: true
    max_series: 10000
    reduction_target_percentage: 70
    preserve_critical_metrics: true
    critical_metrics_patterns:
      - "^system\\.cpu\\."
      - "^system\\.memory\\."
      - "^http\\.server\\.duration"
```

## Creating NRDOT Experiments

### CLI Example

```bash
# Create an experiment with NRDOT
phoenix-cli experiment create \
  --name "nrdot-cardinality-test" \
  --baseline "baseline" \
  --candidate "nrdot-cardinality" \
  --hosts "host1,host2,host3" \
  --config '{
    "nr_license_key": "your-key",
    "nr_otlp_endpoint": "otlp.nr-data.net:4317",
    "max_cardinality": 10000,
    "reduction_percentage": 70
  }'
```

### API Example

```bash
curl -X POST http://phoenix-api:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nrdot-optimization",
    "description": "Test NRDOT cardinality reduction",
    "baseline_pipeline": "baseline",
    "candidate_pipeline": "nrdot-cardinality",
    "target_hosts": ["host1", "host2"],
    "config": {
      "collector_type": "nrdot",
      "nr_license_key": "${NEW_RELIC_LICENSE_KEY}",
      "max_cardinality": 10000,
      "reduction_percentage": 70
    }
  }'
```

## A/B Testing: OTel vs NRDOT

Phoenix supports running experiments comparing standard OTel collectors with NRDOT:

```bash
# Create A/B test experiment
phoenix-cli experiment create \
  --name "otel-vs-nrdot" \
  --baseline "baseline" \
  --candidate "nrdot-cardinality" \
  --duration "1h" \
  --analysis-type "cardinality_comparison"
```

This will:
1. Run OTel collector as baseline
2. Run NRDOT collector as candidate
3. Compare cardinality reduction effectiveness
4. Measure resource usage differences

## Monitoring NRDOT Performance

### Metrics

NRDOT exposes additional metrics:

- `nrdot_cardinality_total`: Total unique metric series
- `nrdot_cardinality_dropped`: Dropped metric series
- `nrdot_reduction_percentage`: Actual reduction percentage
- `nrdot_processing_time_ms`: Processing latency

### Health Checks

NRDOT health endpoint:
```bash
curl http://localhost:13133/health/status
```

## Best Practices

### 1. Gradual Rollout

Start with a small percentage of hosts:
```yaml
rollout_strategy:
  type: "percentage"
  phases:
    - percentage: 10
      duration: "1h"
    - percentage: 50
      duration: "2h"
    - percentage: 100
```

### 2. Critical Metrics Protection

Always preserve critical metrics:
```yaml
critical_metrics_patterns:
  - "^system\\.cpu\\.utilization"
  - "^system\\.memory\\.usage"
  - "^http\\.server\\.request\\.duration"
  - "^db\\.query\\.duration"
  - "error\\.rate"
```

### 3. Resource Limits

Set appropriate resource limits:
```yaml
resources:
  memory_limit_mib: 512
  cpu_limit_percent: 10
```

## Troubleshooting

### NRDOT Not Starting

1. Check license key:
```bash
echo $NEW_RELIC_LICENSE_KEY
```

2. Verify binary installation:
```bash
which nrdot
ls -la /usr/local/bin/nrdot
```

3. Check agent logs:
```bash
journalctl -u phoenix-agent -f
```

### High Cardinality Despite NRDOT

1. Review cardinality settings:
```yaml
newrelic/cardinality:
  max_series: 10000  # Increase if needed
```

2. Check metric patterns:
```bash
# View top cardinality contributors
curl http://localhost:8888/metrics | grep nrdot_cardinality_by_metric
```

### Performance Issues

1. Enable debug logging:
```yaml
service:
  telemetry:
    logs:
      level: debug
```

2. Profile NRDOT:
```bash
curl http://localhost:1777/debug/pprof/profile > nrdot.prof
```

## Migration Guide

### From OTel to NRDOT

1. **Update agent configuration**:
```bash
# On each host
sudo sed -i 's/USE_NRDOT=false/USE_NRDOT=true/' /etc/phoenix-agent/config.yaml
sudo systemctl restart phoenix-agent
```

2. **Update experiments**:
```sql
-- Update existing experiments to use NRDOT
UPDATE experiments 
SET candidate_pipeline = 'nrdot-cardinality'
WHERE candidate_pipeline = 'adaptive-filter';
```

3. **Monitor transition**:
```bash
# Watch agent status
phoenix-cli agent list --watch

# Monitor metrics flow
curl http://prometheus:9090/api/v1/query?query=up{job="phoenix-experiment"}
```

## Security Considerations

1. **License Key Protection**:
   - Store in secure secrets management
   - Never commit to version control
   - Rotate regularly

2. **Network Security**:
   - NRDOT uses TLS by default
   - Ensure firewall allows outbound 4317/4318

3. **Access Control**:
   - Limit who can configure NRDOT settings
   - Audit configuration changes

## Performance Benchmarks

Typical improvements with NRDOT:

| Metric | OTel Baseline | NRDOT | Improvement |
|--------|---------------|-------|-------------|
| Cardinality | 100,000 | 30,000 | 70% reduction |
| Memory Usage | 512 MB | 350 MB | 31% reduction |
| CPU Usage | 10% | 8% | 20% reduction |
| Latency p99 | 100ms | 80ms | 20% improvement |

## Integration with New Relic UI

When using NRDOT, metrics appear in New Relic with additional metadata:

- `collector.type`: "nrdot"
- `phoenix.experiment_id`: Experiment identifier
- `phoenix.variant`: "baseline" or "candidate"

Query in NRQL:
```sql
SELECT count(*) 
FROM Metric 
WHERE collector.type = 'nrdot' 
AND phoenix.experiment_id = 'exp-123'
FACET phoenix.variant
```

## Summary

NRDOT integration provides Phoenix users with:

1. **Enhanced cardinality reduction** - Up to 70% reduction while preserving critical metrics
2. **Better New Relic integration** - Optimized for New Relic's backend
3. **A/B testing capability** - Compare OTel vs NRDOT performance
4. **Production-ready features** - Built-in safety mechanisms

For additional support, contact the Phoenix team or refer to New Relic's NRDOT documentation.