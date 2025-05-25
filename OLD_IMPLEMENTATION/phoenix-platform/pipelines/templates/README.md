# Phoenix Pipeline Template Library

This directory contains pre-validated OpenTelemetry collector configurations optimized for process metrics collection. Each template represents a different strategy for reducing telemetry volume while maintaining visibility for critical processes.

## Available Templates

### 1. process-baseline-v1.yaml
**Strategy**: Full Collection (Control Group)
**Expected Reduction**: 0% (baseline)
**Description**: Complete process metrics collection with no optimization. Used as the baseline for A/B testing experiments.

**Features**:
- Full metric collection for all processes
- Complete dimensionality preserved
- All OpenTelemetry metric types enabled
- Suitable for environments requiring 100% visibility

**Use Cases**:
- Baseline measurement for optimization experiments
- Compliance environments requiring full audit trails
- Initial deployment before optimization
- Small-scale environments where cost is not a concern

---

### 2. process-priority-filter-v1.yaml
**Strategy**: Priority-Based Filtering
**Expected Reduction**: 40-60%
**Description**: Intelligent process classification with filtering based on business importance.

**Features**:
- Automatic process classification (critical/high/medium/low)
- Rule-based filtering keeps only important processes
- Pattern matching for databases, web servers, applications
- Preserves critical infrastructure visibility

**Use Cases**:
- Production environments with known critical processes
- Organizations with clear service tier definitions
- Environments where specific applications must be monitored
- Teams needing predictable filtering behavior

**Classification Rules**:
- **Critical**: nginx, postgres, mysql, redis, kafka, kubelet
- **High**: systemd, sshd, system processes
- **Medium**: python, java, node with >100MB memory
- **Low**: Everything else (filtered out)

---

### 3. process-aggregated-v1.yaml
**Strategy**: Metric Aggregation
**Expected Reduction**: 30-50%
**Description**: Reduces cardinality by aggregating similar processes while preserving essential metrics.

**Features**:
- Groups processes by executable name and host
- Maintains visibility into resource consumption patterns
- Reduces time series count through intelligent aggregation
- Preserves critical alerting capabilities

**Use Cases**:
- High-cardinality environments with many similar processes
- Container platforms with many identical workloads
- Microservices architectures with repeated patterns
- Cost optimization with maintained alerting

---

### 4. process-topk-v1.yaml
**Strategy**: Top-K Resource Consumers
**Expected Reduction**: 60-80%
**Description**: Keeps only the highest resource-consuming processes plus critical infrastructure.

**Features**:
- Resource-based ranking (CPU + memory scoring)
- Always includes critical processes regardless of usage
- Configurable thresholds for inclusion
- Simulated top-k behavior with filtering

**Use Cases**:
- Performance troubleshooting focus
- Environments with clear resource bottlenecks
- Cost-sensitive deployments needing significant reduction
- Kubernetes clusters with many idle pods

**Selection Criteria**:
- Processes with >5% CPU utilization
- Processes with >100MB memory usage
- All database and web server processes
- Configurable sampling for remaining processes

---

### 5. process-adaptive-v1.yaml
**Strategy**: Adaptive Collection
**Expected Reduction**: 70-90%
**Description**: Dynamically adjusts collection based on process activity and type with intelligent sampling.

**Features**:
- Activity-based classification (active/idle)
- Process type-aware filtering
- Adaptive sampling rates by category
- Metric aggregation by process type
- Reduced collection frequency (30s intervals)

**Use Cases**:
- Dynamic environments with varying load patterns
- Cost-critical deployments requiring maximum reduction
- Environments with predictable quiet periods
- Auto-scaling infrastructures

**Adaptive Logic**:
- **Active Databases/Webservers**: 100% collection
- **Active Runtime Processes**: 100% collection
- **Idle Runtime Processes**: 10% sampling
- **System Processes**: Important only
- **Other Processes**: 5% sampling

---

### 6. process-minimal-v1.yaml
**Strategy**: Minimal Collection
**Expected Reduction**: 95%+
**Description**: Absolute minimum process metrics for cost-critical environments.

**Features**:
- Only critical process types monitored
- Minimal metrics (CPU utilization, memory only)
- Heavy aggregation by process name
- Long collection intervals (60s)
- Resource-based filtering

**Use Cases**:
- Emergency cost reduction scenarios
- POC/development environments
- Compliance-only monitoring
- Budget-constrained deployments

**Limitations**:
- ⚠️ Minimal visibility - may miss important processes
- ⚠️ Not suitable for detailed performance analysis
- ⚠️ Limited alerting capabilities

---

### 7. process-intelligent-v1.yaml
**Strategy**: Intelligent Multi-Dimensional Classification
**Expected Reduction**: 50-75%
**Description**: Advanced classification using multiple dimensions with performance-aware filtering.

**Features**:
- Multi-tier process classification (tier1/tier2/tier3)
- Performance-based categorization (high/medium/low)
- Business criticality assessment
- Multi-dimensional filtering logic
- Adaptive sampling by importance

**Use Cases**:
- Production environments requiring balanced visibility and cost
- Complex multi-tier applications
- SRE teams needing intelligent alerting
- Mature organizations with defined service tiers

**Classification Matrix**:
| Tier | Performance | Retention Rate |
|------|-------------|----------------|
| Tier1 (Critical) | Any | 100% |
| Tier2 (Important) | High/Medium | 100% |
| Tier2 (Important) | Low | 0% |
| Tier3 (Normal) | High | 100% |
| Tier3 (Normal) | Medium | 30% |
| Tier3 (Normal) | Low | 5% |

## Template Selection Guide

### By Environment Type
- **Development/Testing**: process-minimal-v1.yaml
- **Production (High Visibility)**: process-intelligent-v1.yaml
- **Production (Cost Optimized)**: process-adaptive-v1.yaml
- **Emergency Cost Reduction**: process-minimal-v1.yaml
- **Compliance/Audit**: process-baseline-v1.yaml

### By Reduction Target
- **0-20% reduction**: process-baseline-v1.yaml
- **20-40% reduction**: process-aggregated-v1.yaml
- **40-60% reduction**: process-priority-filter-v1.yaml
- **60-80% reduction**: process-topk-v1.yaml or process-intelligent-v1.yaml
- **80%+ reduction**: process-adaptive-v1.yaml or process-minimal-v1.yaml

### By Use Case
- **A/B Testing Baseline**: process-baseline-v1.yaml
- **Known Critical Processes**: process-priority-filter-v1.yaml
- **High Cardinality Issues**: process-aggregated-v1.yaml
- **Performance Focus**: process-topk-v1.yaml
- **Dynamic Environments**: process-adaptive-v1.yaml
- **Cost Crisis**: process-minimal-v1.yaml
- **Complex Environments**: process-intelligent-v1.yaml

## Template Variables

All templates support the following standard variables:

### Required Variables
- `PHOENIX_EXPERIMENT_ID`: Unique identifier for the experiment
- `PHOENIX_VARIANT`: Either "baseline" or "candidate"
- `NEW_RELIC_API_KEY`: API key for New Relic export
- `NODE_NAME`: Kubernetes node name or hostname

### Optional Variables
- `NEW_RELIC_OTLP_ENDPOINT`: New Relic OTLP endpoint (default: https://otlp.nr-data.net)

### Template-Specific Variables
Some templates support additional configuration variables. Check the metadata section in each template file for details.

## Validation

All templates have been validated for:
- ✅ YAML syntax correctness
- ✅ OpenTelemetry configuration schema compliance
- ✅ Processor ordering best practices
- ✅ Resource utilization optimization
- ✅ New Relic compatibility

## Best Practices

1. **Start Conservative**: Begin with process-intelligent-v1.yaml for production
2. **Measure Impact**: Always run A/B experiments to measure actual reduction
3. **Monitor Alerting**: Ensure critical alerts still fire after optimization
4. **Gradual Rollout**: Deploy to small percentage of infrastructure first
5. **Have Rollback Plan**: Keep baseline configuration ready for quick revert

## Contributing

When adding new templates:
1. Follow the naming convention: `process-{strategy}-v{version}.yaml`
2. Include comprehensive metadata section
3. Test with real workloads
4. Document expected reduction percentages
5. Update this README with template details