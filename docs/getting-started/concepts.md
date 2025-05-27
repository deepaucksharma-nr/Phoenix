# Phoenix Platform Core Concepts

## Overview

This guide introduces the fundamental concepts and terminology used throughout the Phoenix Platform. Understanding these concepts will help you effectively use Phoenix to optimize your observability costs.

## Key Concepts

### Metrics Cardinality

**Definition**: The number of unique time series in your metrics data, determined by the combination of metric names and their label values.

**Example**:
```
http_requests_total{method="GET", status="200", endpoint="/api/users"} 
http_requests_total{method="POST", status="201", endpoint="/api/users"}
http_requests_total{method="GET", status="404", endpoint="/api/posts"}
```
This creates 3 unique time series (cardinality = 3).

**Why it matters**: High cardinality directly impacts storage costs and query performance in observability platforms.

### Experiments

**Definition**: Controlled A/B tests that compare different pipeline configurations to measure their effectiveness at reducing cardinality while maintaining data quality.

**Components**:
- **Baseline Pipeline**: Current production configuration (control group)
- **Candidate Pipeline**: New optimization configuration (test group)
- **Traffic Split**: Percentage of agents running the candidate (e.g., 20%)
- **Success Criteria**: Thresholds for cardinality reduction and data quality

**Lifecycle**:
1. **Created**: Experiment defined but not running
2. **Running**: Actively collecting comparison data
3. **Analyzing**: Processing results after completion
4. **Completed**: Results available for decision making

### Pipelines

**Definition**: OpenTelemetry Collector configurations that process metrics before they reach your observability backend.

**Types**:
1. **Passthrough**: No modification (baseline)
2. **Adaptive Filter**: ML-based importance scoring
3. **TopK**: Keep only top K important metrics
4. **Hybrid**: Combination of strategies

**Pipeline Flow**:
```
Metrics Source → OTel Receiver → Processors → Exporter → Backend
                                      ↑
                                Phoenix Pipeline
```

### Agents

**Definition**: Lightweight services deployed alongside OpenTelemetry Collectors that execute pipeline configurations and report metrics.

**Responsibilities**:
- Poll control plane for tasks
- Deploy pipeline configurations
- Collect performance metrics
- Report experiment results

**Authentication**: Each agent uses a unique `X-Agent-Host-ID` for identification.

### Task Queue

**Definition**: PostgreSQL-based queue system for distributing work to agents without requiring persistent connections.

**Benefits**:
- Resilient to network interruptions
- Supports thousands of agents
- Enables gradual rollouts
- Provides audit trail

**Flow**:
1. API creates task in queue
2. Agent polls for tasks (30s timeout)
3. Agent executes task
4. Agent reports completion

### Cardinality Reduction

**Definition**: The process of intelligently filtering metrics to reduce the number of unique time series without losing critical observability data.

**Strategies**:
1. **Importance-based**: Keep metrics that matter most
2. **Sampling**: Reduce frequency of less important metrics
3. **Aggregation**: Combine similar metrics
4. **Dropping**: Remove truly unnecessary metrics

**Measurement**: 
```
Reduction Rate = (Baseline Cardinality - Optimized Cardinality) / Baseline Cardinality
```

### Signal Preservation

**Definition**: A measure of how well the optimized pipeline maintains the ability to detect issues and anomalies compared to the baseline.

**Calculation**: Based on:
- Alert accuracy
- Anomaly detection rate
- Query result similarity
- Statistical distribution preservation

**Target**: Maintain >99% signal preservation while reducing cardinality.

### Cost Flow

**Definition**: Real-time visualization and calculation of cost savings from cardinality reduction.

**Components**:
- **Baseline Cost**: Current metrics ingestion cost
- **Optimized Cost**: Cost with reduction applied
- **Savings Rate**: Dollar savings per hour/day/month

**Formula**:
```
Savings = (Baseline MPS - Optimized MPS) × Cost per Million Metrics × Time
```

### Variants

**Definition**: Different versions of a pipeline deployed as part of an experiment.

**Types**:
- **Baseline Variant**: Control group configuration
- **Candidate Variant**: Test group configuration

**Purpose**: Enable safe testing of optimizations with ability to roll back.

### Deployment

**Definition**: The process of applying a pipeline configuration to one or more agents.

**Attributes**:
- **Deployment ID**: Unique identifier
- **Target Agents**: Which agents receive the config
- **Status**: pending, active, failed
- **Metrics**: Performance data from deployment

### WebSocket Channels

**Definition**: Real-time communication channels for live updates.

**Available Channels**:
- `experiments`: Experiment lifecycle events
- `metrics`: Live metrics updates
- `agents`: Agent status changes
- `deployments`: Deployment progress
- `alerts`: System alerts
- `cost-flow`: Cost savings updates

## Architecture Concepts

### Control Plane

**Definition**: Central management system (Phoenix API) that orchestrates experiments, manages configurations, and aggregates results.

**Responsibilities**:
- Experiment lifecycle management
- Pipeline configuration storage
- Task distribution
- Metrics aggregation
- WebSocket event broadcasting

### Data Plane

**Definition**: Distributed agents that execute pipeline configurations and process metrics.

**Components**:
- Phoenix Agents
- OpenTelemetry Collectors
- Metrics sources

### Monorepo Structure

**Definition**: Single repository containing all Phoenix components with shared packages.

**Benefits**:
- Consistent versioning
- Shared code reuse
- Simplified testing
- Atomic changes

**Structure**:
```
/projects/*   - Independent services
/pkg/*        - Shared packages
/deployments/* - Deployment configs
```

## Operational Concepts

### Long Polling

**Definition**: Technique where agents maintain open HTTP connections to receive tasks without constant polling.

**Configuration**:
- Timeout: 30 seconds
- Automatic reconnection
- Zero message loss

### Heartbeat

**Definition**: Regular health check messages from agents to the control plane.

**Purpose**:
- Detect offline agents
- Monitor agent health
- Track resource usage
- Maintain agent registry

### Rollback

**Definition**: The ability to quickly revert to previous pipeline configuration if issues are detected.

**Triggers**:
- High error rate
- Performance degradation
- Manual intervention
- Experiment failure

### Promotion

**Definition**: Process of rolling out a successful experiment's configuration to all agents.

**Strategies**:
- **Immediate**: Deploy to all agents at once
- **Gradual**: Increase percentage over time
- **Canary**: Start with small percentage

## Success Metrics

### Cardinality Reduction Rate
- **Target**: 70% reduction
- **Measurement**: Comparing unique series count

### Cost Savings
- **Calculation**: Based on vendor pricing
- **Tracking**: Per experiment and cumulative

### Signal Quality
- **Preservation**: >99% of critical signals
- **Validation**: Alert accuracy testing

### Performance Impact
- **Latency**: <10ms additional processing time
- **CPU**: <5% overhead
- **Memory**: <100MB per collector

## Next Steps

Now that you understand the core concepts:

1. [Create your first experiment](first-experiment.md)
2. [Explore pipeline types](../user-guide/pipelines.md)
3. [Monitor cost savings](../user-guide/monitoring.md)
4. [Read the architecture guide](../architecture/system-design.md)