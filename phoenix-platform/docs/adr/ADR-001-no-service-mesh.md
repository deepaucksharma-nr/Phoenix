# ADR-001: No Service Mesh for A/B Testing

## Status
Accepted

## Context
The Phoenix platform needs to run A/B tests comparing two different OpenTelemetry pipeline configurations. The typical approach would be to use a service mesh (like Istio) to route traffic between different versions.

## Decision
We will NOT use a service mesh. Instead, we will run two OpenTelemetry collectors (baseline and candidate) side-by-side on the same host, both collecting the same metrics but exporting to different endpoints.

## Rationale
1. **Simplicity**: Service mesh adds significant complexity for a simple use case
2. **Performance**: Avoids service mesh overhead (important for metrics collection)
3. **Cost**: No additional infrastructure required
4. **Accuracy**: Both collectors see identical data, ensuring fair comparison

## Implementation
- Deploy two DaemonSets per experiment (baseline and candidate)
- Both collectors run on same nodes with different names
- Collectors export to different Prometheus targets
- Use labels to differentiate metrics

## Consequences
### Positive
- Simpler architecture
- Lower operational overhead
- Better performance
- Easier debugging

### Negative
- Cannot do traffic splitting (not needed for our use case)
- Both collectors consume resources (acceptable for testing)

## Alternatives Considered
1. **Service Mesh**: Too complex for simple A/B testing
2. **Single Collector with Feature Flags**: Wouldn't give true A/B comparison
3. **Time-based Testing**: Sequential testing wouldn't account for workload variations

## References
- Original design in TECHNICAL_SPEC_MASTER.md
- A/B testing requirements in PRODUCT_REQUIREMENTS.md