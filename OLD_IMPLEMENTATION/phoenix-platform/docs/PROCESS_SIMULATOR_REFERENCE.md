# Process Simulator Guide

## Overview

The Phoenix Process Simulator is a sophisticated tool designed to generate realistic process workloads for testing telemetry collection and optimization strategies. It simulates various process patterns, resource usage behaviors, and chaos scenarios to validate the effectiveness of Phoenix platform experiments.

## Architecture

The simulator consists of several key components:

1. **Process Simulation Engine**: Creates and manages simulated processes with configurable resource patterns
2. **Metrics Emitter**: Exposes process metrics in Prometheus format, mimicking OpenTelemetry hostmetrics
3. **Control API**: RESTful API for managing simulations and triggering chaos events
4. **Event Bus Integration**: Publishes simulation lifecycle events for monitoring and orchestration

## Features

### Simulation Profiles

#### 1. Realistic Profile
Simulates a typical production environment:
- **Processes**: nginx, postgres, redis, python apps, node services, browser tabs, cron jobs
- **Characteristics**: 
  - Stable long-running services (nginx, postgres, redis)
  - Variable application processes
  - Short-lived jobs and tasks
  - 10% churn rate per hour
- **Use Case**: Testing baseline telemetry collection

#### 2. High-Cardinality Profile
Creates extreme process diversity:
- **Processes**: 500+ unique process names with random patterns
- **Characteristics**:
  - Microservices with unique identifiers
  - Container-based workloads
  - Sidecar processes
  - 50% churn rate per hour
- **Use Case**: Testing cardinality reduction strategies

#### 3. High-Churn Profile  
Emphasizes process lifecycle dynamics:
- **Processes**: Short-lived jobs, batch processors, lambda-style functions
- **Characteristics**:
  - Process lifetimes from 15 seconds to 5 minutes
  - Constant creation and termination
  - 80% churn rate per hour
- **Use Case**: Testing process tracking and churn handling

#### 4. Chaos Profile
Includes unpredictable behaviors:
- **Processes**: Mix of critical and non-critical services
- **Characteristics**:
  - Random CPU/memory spikes
  - Unexpected process deaths
  - Resource leaks
  - 90% churn rate
- **Use Case**: Testing resilience and error handling

### Resource Patterns

#### CPU Patterns
- **steady**: Consistent CPU usage (20-25%)
- **spiky**: Alternates between low (10-20%) and high (70-90%) usage
- **growing**: Gradually increases over time (simulates CPU leaks)
- **random**: Completely unpredictable usage (0-100%)

#### Memory Patterns
- **steady**: Consistent memory usage (50-60MB)
- **spiky**: Periodic memory spikes (40-200MB)
- **growing**: Gradual memory increase (simulates memory leaks)
- **random**: Unpredictable memory usage (10-210MB)

### Process Classification

Processes are classified by priority:
- **critical**: Database servers, cache services (never churned)
- **high**: Web servers, load balancers
- **medium**: Application services
- **low**: Background jobs, temporary processes

## Metrics Exposed

The simulator exposes metrics compatible with OpenTelemetry hostmetrics receiver:

```
# Process CPU time
process_cpu_seconds_total{process_name="nginx-worker-1",pid="1234",priority="high",host="node-1"}

# Process memory usage
process_memory_bytes{process_name="postgres-1",pid="5678",priority="critical",host="node-1",type="rss"}
process_memory_bytes{process_name="postgres-1",pid="5678",priority="critical",host="node-1",type="vms"}

# Process threads
process_threads{process_name="java-app-1",pid="9012",priority="medium",host="node-1"}

# Open file descriptors
process_open_fds{process_name="redis-server-1",pid="3456",priority="critical",host="node-1"}

# Process start time and uptime
process_start_time_seconds{process_name="python-app-1",pid="7890",priority="medium",host="node-1"}
process_uptime_seconds{process_name="python-app-1",pid="7890",priority="medium",host="node-1"}
```

## Control API

### Create Simulation
```bash
curl -X POST http://localhost:8090/api/v1/simulations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-simulation",
    "type": "realistic",
    "duration": "1h",
    "parameters": {
      "process_count": 150,
      "enable_chaos": false
    }
  }'
```

### Start Simulation
```bash
curl -X POST http://localhost:8090/api/v1/simulations/{id}/start
```

### Get Results
```bash
curl http://localhost:8090/api/v1/simulations/{id}/results
```

### Trigger Chaos Events

#### CPU Spike
```bash
curl -X POST http://localhost:8090/api/v1/chaos/cpu-spike \
  -H "Content-Type: application/json" \
  -d '{
    "process_pattern": "python-app",
    "duration": "30s",
    "intensity": 90.0
  }'
```

#### Memory Leak
```bash
curl -X POST http://localhost:8090/api/v1/chaos/memory-leak \
  -H "Content-Type: application/json" \
  -d '{
    "process_pattern": "node-service",
    "leak_rate": "10MB/min"
  }'
```

#### Process Kill
```bash
curl -X POST http://localhost:8090/api/v1/chaos/process-kill \
  -H "Content-Type: application/json" \
  -d '{
    "process_pattern": "worker",
    "count": 5
  }'
```

## Deployment

### Docker
```bash
docker run -d \
  -p 8090:8090 \
  -p 8888:8888 \
  -e PROFILE=realistic \
  -e DURATION=1h \
  -e AUTO_START=true \
  phoenix/process-simulator:latest
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: process-simulator
  namespace: phoenix-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: process-simulator
  template:
    metadata:
      labels:
        app: process-simulator
    spec:
      containers:
      - name: simulator
        image: phoenix/process-simulator:latest
        ports:
        - containerPort: 8090  # Control API
        - containerPort: 8888  # Metrics
        env:
        - name: PROFILE
          value: "high-cardinality"
        - name: DURATION
          value: "2h"
        - name: PROCESS_COUNT
          value: "500"
        - name: AUTO_START
          value: "true"
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONTROL_PORT` | Control API port | 8090 |
| `METRICS_PORT` | Prometheus metrics port | 8888 |
| `AUTO_START` | Start simulation on launch | false |
| `PROFILE` | Simulation profile | realistic |
| `DURATION` | Simulation duration | 1h |
| `PROCESS_COUNT` | Target process count | 100 |
| `ENABLE_CHAOS` | Enable chaos engineering | false |
| `TARGET_HOST` | Hostname in metrics labels | localhost |

## Integration with Phoenix Platform

### 1. Experiment Setup
Configure experiments to use the simulator:
```yaml
apiVersion: phoenix.newrelic.com/v1alpha1
kind: LoadSimulationJob
metadata:
  name: experiment-load
spec:
  profile: high-cardinality
  duration: 30m
  targetNodes:
  - node-1
  - node-2
```

### 2. Metrics Collection
OpenTelemetry collectors will automatically discover and collect metrics from:
- Prometheus endpoint: `http://simulator:8888/metrics`
- Process metrics matching hostmetrics format

### 3. Event Integration
The simulator publishes events to the EventBus:
- `SimulationCreated`
- `SimulationStarted`
- `SimulationCompleted`
- `SimulationFailed`

## Performance Considerations

### Resource Usage
- **CPU**: 0.5-2 cores depending on process count
- **Memory**: 200MB-2GB depending on simulation complexity
- **Network**: Minimal (metrics endpoint only)

### Scaling
- Single simulator instance can handle up to 1000 processes
- For larger simulations, deploy multiple instances
- Use node affinity to co-locate with collectors

## Troubleshooting

### Common Issues

1. **High CPU usage**
   - Reduce process count
   - Use "steady" CPU patterns
   - Disable chaos features

2. **Metrics not appearing**
   - Check metrics endpoint: `curl http://localhost:8888/metrics`
   - Verify process creation in logs
   - Ensure Prometheus scraping is configured

3. **Processes not churning**
   - Verify profile churn rate
   - Check process lifetimes in configuration
   - Review simulator logs for errors

### Debug Commands
```bash
# Check simulator health
curl http://localhost:8090/api/v1/health

# View current simulations
curl http://localhost:8090/api/v1/simulations

# Check metrics
curl -s http://localhost:8888/metrics | grep process_

# View logs
kubectl logs -n phoenix-system deployment/process-simulator
```

## Best Practices

1. **Start Small**: Begin with realistic profile and 50-100 processes
2. **Monitor Resources**: Watch simulator CPU/memory usage
3. **Use Appropriate Profiles**: Match profile to testing scenario
4. **Enable Chaos Gradually**: Start without chaos, add complexity
5. **Coordinate with Experiments**: Align simulation duration with experiments

## Example Scenarios

### Scenario 1: Testing Cardinality Reduction
```bash
# High-cardinality simulation
PROFILE=high-cardinality PROCESS_COUNT=1000 DURATION=1h ./simulator

# Expected outcome: 1000+ unique time series
# Phoenix should reduce to <500 without losing critical processes
```

### Scenario 2: Testing Process Priority
```bash
# Mixed priority processes with chaos
PROFILE=chaos ENABLE_CHAOS=true DURATION=30m ./simulator

# Expected outcome: Critical processes always retained
# Low priority processes filtered during high load
```

### Scenario 3: Performance Testing
```bash
# Sustained high load
PROFILE=realistic PROCESS_COUNT=500 DURATION=6h ./simulator

# Expected outcome: Stable collection performance
# No memory leaks or CPU creep in collectors
```