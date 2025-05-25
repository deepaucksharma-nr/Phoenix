# ADR-005: Dual Metrics Export Pattern

## Status
Accepted

## Context
During A/B testing of OpenTelemetry pipelines, we need to compare metrics from baseline and candidate configurations. We also need to ensure production metrics continue flowing to New Relic.

## Decision
All OpenTelemetry collectors will export metrics to BOTH Prometheus (for comparison) AND New Relic (for production monitoring) simultaneously.

## Rationale
1. **Comparison**: Prometheus provides local storage for A/B analysis
2. **Continuity**: New Relic export ensures no monitoring gaps
3. **Safety**: Production monitoring continues during experiments
4. **Analysis**: Local Prometheus enables detailed comparison
5. **Cost**: Temporary dual export is acceptable for testing

## Implementation

### OTel Pipeline Configuration
```yaml
exporters:
  # Production metrics
  otlp/newrelic:
    endpoint: "https://otlp.nr-data.net:4317"
    headers:
      api-key: "${NEW_RELIC_API_KEY}"
  
  # Comparison metrics
  prometheus:
    endpoint: "0.0.0.0:8888"
    namespace: "phoenix"
    const_labels:
      pipeline: "${PIPELINE_NAME}"
      experiment: "${EXPERIMENT_ID}"

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [batch, attributes]
      exporters: [otlp/newrelic, prometheus]  # BOTH
```

### Metrics Flow
```
Host Metrics
    ↓
OTel Collector
    ├─→ New Relic (Production)
    └─→ Prometheus (Comparison)
         └─→ Phoenix Analysis Service
```

## Comparison Strategy
1. **Baseline**: Exports with label `variant="baseline"`
2. **Candidate**: Exports with label `variant="candidate"`
3. **Analysis**: Query Prometheus for both variants
4. **Duration**: Keep Prometheus data for experiment duration only

## Consequences
### Positive
- No production monitoring gaps
- Accurate local comparison
- Can analyze without New Relic API limits
- Rollback doesn't affect production metrics

### Negative
- Temporary double resource usage
- Need local Prometheus storage
- Complexity of dual export config

## Storage Considerations
- Prometheus retention: 7 days (experiment duration)
- Automatic cleanup after experiment
- Estimated 10GB per experiment
- Use persistent volumes in production

## Alternatives Considered
1. **New Relic Only**: API limits and cost for comparison queries
2. **Prometheus Only**: Risk of production monitoring gaps
3. **Split Traffic**: Would miss some production metrics
4. **Export After Analysis**: Too late for real-time comparison

## References
- Metrics analysis in TECHNICAL_SPEC_CONTROLLER.md
- A/B testing requirements in PRODUCT_REQUIREMENTS.md
- Prometheus configuration in monitoring/