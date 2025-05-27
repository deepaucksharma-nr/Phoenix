# Load Simulation Profiles

## Overview

The Phoenix Agent includes built-in load simulation capabilities to test system behavior under various metric load patterns. This document describes the available load profiles, their characteristics, and usage guidelines.

## Available Profiles

### 1. High Cardinality (`high-cardinality`, `high-card`)

Simulates a scenario with extremely high metric cardinality, typical of poorly configured applications with unbounded label values.

**Characteristics:**
- Generates 1000 unique metrics per second
- Each metric has unique user IDs and endpoint paths
- Creates exponential growth in time series
- Tests system behavior under cardinality explosion

**Use Cases:**
- Testing cardinality reduction processors
- Validating memory limits and safeguards
- Demonstrating the need for Phoenix optimization

**Generated Metrics:**
- Metric: `http.request.duration` (histogram)
- Labels: `user.id` (unique UUID), `endpoint` (unique path), `status_code`
- Rate: 1000 metrics/second

### 2. Normal/Realistic Load (`realistic`, `normal`)

Simulates typical production workload with moderate resource usage.

**Characteristics:**
- Uses system stress tools (stress-ng) when available
- Falls back to CPU-intensive calculations
- Generates consistent system load
- Duration: 60 seconds default

**Use Cases:**
- Baseline performance testing
- Comparing optimized vs non-optimized pipelines
- General system validation

**System Impact:**
- CPU: 2 cores at ~50% utilization
- I/O: 2 concurrent I/O workers
- Memory: 128MB working set
- Network: Minimal impact

### 3. Spike Load (`spike`)

Simulates sudden traffic spikes followed by return to normal, testing auto-scaling and recovery.

**Characteristics:**
- 30 seconds normal load (1 metric/second)
- 10 seconds spike (100x normal rate)
- 20 seconds recovery to normal
- Tests system elasticity

**Use Cases:**
- Testing auto-scaling triggers
- Validating buffer and queue handling
- Checking recovery behavior

**Load Pattern:**
```
Rate
100x |     ████████
     |     ████████
 10x |     ████████
     |█████        █████
  1x |█████        █████
     +----+--------+----+
     0   30s      40s  60s
```

### 4. Steady Load (`steady`, `process-churn`)

Maintains consistent metric generation rate for extended periods.

**Characteristics:**
- Constant 10 requests/second
- Minimal variance
- Suitable for long-running tests
- Low resource overhead

**Use Cases:**
- Baseline measurements
- Long-term stability testing
- Resource leak detection

**Generated Metrics:**
- Metric: `http.requests` (counter)
- Labels: `method` (GET)
- Rate: 10 metrics/second constant

## Usage

### Starting Load Simulation

Via Phoenix CLI:
```bash
phoenix-cli loadsim start --profile <profile-name> --duration <duration>
```

Via API:
```bash
curl -X POST http://localhost:8080/api/v1/agent/tasks \
  -H "Content-Type: application/json" \
  -H "X-Agent-Host-ID: agent-1" \
  -d '{
    "type": "loadsim",
    "action": "start",
    "config": {
      "profile": "high-cardinality",
      "duration": "5m"
    }
  }'
```

### Stopping Load Simulation

Via CLI:
```bash
phoenix-cli loadsim stop
```

Via API:
```bash
curl -X POST http://localhost:8080/api/v1/agent/tasks \
  -H "Content-Type: application/json" \
  -H "X-Agent-Host-ID: agent-1" \
  -d '{
    "type": "loadsim",
    "action": "stop"
  }'
```

### Monitoring Load Simulation

Check status via CLI:
```bash
phoenix-cli loadsim status
```

Agent metrics will include:
```json
{
  "load_sim_active": true,
  "pid": 12345
}
```

## Profile Selection Guide

| Scenario | Recommended Profile | Duration | Expected Result |
|----------|-------------------|-----------|-----------------|
| Testing cardinality reduction | `high-cardinality` | 2-5 minutes | 70%+ reduction in series |
| Baseline performance | `normal` | 5-10 minutes | Stable resource usage |
| Auto-scaling validation | `spike` | 1-2 minutes | Successful scale-out/in |
| Long-term testing | `steady` | 30+ minutes | No resource leaks |
| A/B experiment validation | `realistic` | Duration of experiment | Clear performance delta |

## Resource Requirements

### High Cardinality Profile
- CPU: Low (mostly I/O bound)
- Memory: Can grow significantly without limits
- Network: ~1MB/s outbound
- Disk: Minimal

### Normal/Realistic Profile
- CPU: 2-4 cores at 50%
- Memory: 128-256MB
- Network: Minimal
- Disk: Low I/O

### Spike Profile
- CPU: Variable (low to high)
- Memory: Stable
- Network: Bursts up to 10MB/s
- Disk: Minimal

### Steady Profile
- CPU: <1 core
- Memory: <50MB
- Network: ~100KB/s constant
- Disk: None

## Best Practices

1. **Duration Planning**
   - Short tests (1-5 min): Profile validation
   - Medium tests (5-30 min): Performance comparison
   - Long tests (30+ min): Stability and leak detection

2. **Resource Monitoring**
   - Always monitor agent resource usage
   - Set appropriate memory limits
   - Watch for OOM kills

3. **Isolation**
   - Run load tests on dedicated agents when possible
   - Avoid mixing with production workloads
   - Use separate namespaces/labels

4. **Incremental Testing**
   - Start with `steady` profile
   - Progress to `normal`, then `spike`
   - Use `high-cardinality` for stress testing

## Customization

### Environment Variables

All profiles respect these environment variables:

- `OTEL_ENDPOINT`: Target OTLP endpoint (default: `http://localhost:4318`)
- `METRICS_PUSHGATEWAY_URL`: Alternative metric destination
- `LOAD_SIM_RATE_MULTIPLIER`: Scale all rates (default: 1.0)

### Custom Profiles

For custom load patterns, create a bash script following this template:

```bash
#!/bin/bash
# Custom load profile

ENDPOINT="${OTEL_ENDPOINT:-http://localhost:4318}"
RATE="${CUSTOM_RATE:-5}"

while true; do
    for i in $(seq 1 $RATE); do
        # Generate your custom metrics here
        curl -X POST "${ENDPOINT}/v1/metrics" \
            -H "Content-Type: application/json" \
            -d '{
                "resourceMetrics": [{
                    "resource": {
                        "attributes": [{
                            "key": "service.name",
                            "value": {"stringValue": "custom-load"}
                        }]
                    },
                    "scopeMetrics": [{
                        "metrics": [{
                            "name": "custom.metric",
                            "gauge": {
                                "dataPoints": [{
                                    "timeUnixNano": "'$(date +%s)'000000000",
                                    "asDouble": '$RANDOM'
                                }]
                            }
                        }]
                    }]
                }]
            }' 2>/dev/null &
    done
    sleep 1
done
```

## Troubleshooting

### Load Simulation Won't Start
- Check agent logs for errors
- Verify OTLP endpoint is accessible
- Ensure no other load sim is running

### High Memory Usage
- Expected for `high-cardinality` profile
- Set memory limits in agent config
- Use shorter durations

### No Metrics Generated
- Verify OTLP endpoint configuration
- Check network connectivity
- Look for curl/stress-ng availability

### Process Won't Stop
- Agent implements graceful shutdown (2s timeout)
- Force kill after timeout if needed
- Check for zombie processes

## Integration with Experiments

Load profiles are commonly used during A/B experiments:

1. Start experiment
2. Begin load simulation on target agents
3. Monitor cardinality and resource metrics
4. Compare baseline vs candidate performance
5. Stop load simulation
6. Analyze results

Example workflow:
```bash
# Start experiment
phoenix-cli experiment start exp-123

# Start load on all agents
phoenix-cli loadsim start --profile high-cardinality --duration 10m --agents all

# Monitor in real-time
phoenix-cli experiment status exp-123 --watch

# Stop load
phoenix-cli loadsim stop --agents all

# Get results
phoenix-cli experiment metrics exp-123
```

## Safety Considerations

1. **Production Systems**
   - Never run `high-cardinality` in production
   - Use `steady` profile for production validation
   - Always set resource limits

2. **Duration Limits**
   - Default max duration: 1 hour
   - Automatic cleanup after duration + 30s
   - Manual stop always available

3. **Concurrent Simulations**
   - Only one simulation per agent
   - Attempting to start another will fail
   - Must stop current before starting new

4. **Resource Protection**
   - Agent monitors its own resources
   - Automatic throttling if limits approached
   - Graceful degradation under pressure