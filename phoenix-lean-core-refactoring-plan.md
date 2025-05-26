# Phoenix Lean-Core Architecture Refactoring Plan

## Executive Summary

This document provides an ultra-detailed plan to refactor Phoenix from a multi-service Kubernetes-native architecture to a lean-core/rich-edge architecture that reduces operational complexity while maintaining all MVP functionality.

### Key Transformations
- **From**: 5+ microservices → **To**: 1 monolithic API + lightweight agents
- **From**: K8s CRDs & operators → **To**: Simple database state + polling agents
- **From**: Custom OTel processors → **To**: Stock OTel with Lua/expression configs
- **From**: Per-collector Prometheus scraping → **To**: Single Pushgateway + remote write

## 1. Current State Analysis

### 1.1 Existing Components to be Collapsed

| Component | Current Location | Lines of Code | Primary Functions | Destination |
|-----------|-----------------|---------------|-------------------|-------------|
| **Controller Service** | `projects/controller/` | ~2,500 | Experiment state machine, gRPC API | → Phoenix API module |
| **Platform API** | `projects/platform-api/` | ~3,000 | REST endpoints, WebSocket, aggregation | → Phoenix API (base) |
| **Pipeline Operator** | `projects/pipeline-operator/` | ~1,800 | Manage PhoenixProcessPipeline CRDs | → Agent task executor |
| **LoadSim Operator** | `projects/loadsim-operator/` | ~1,200 | Manage LoadSimulationJob CRDs | → Agent load generator |
| **Benchmark Service** | `projects/benchmark/` | ~800 | Cost/ingest calculations | → API KPI module |
| **Analytics Service** | `projects/analytics/` | ~1,000 | Experiment analysis | → API analysis module |

### 1.2 Custom Components to Replace

| Component | Location | Current Implementation | Lean Replacement |
|-----------|----------|----------------------|------------------|
| **Top-K Processor** | `pkg/otel/processors/topk/` | Go code (68 lines) | OTel transform + Lua |
| **Adaptive Filter** | `pkg/otel/processors/adaptivefilter/` | Go code (67 lines) | OTel filter + expressions |
| **Custom Collector** | `projects/collector/` | Custom binary | Stock OTel contrib |

### 1.3 Database Schema Analysis

Current tables (from `controller/internal/store/postgres.go`):
- `experiments` - Core experiment data
- `experiment_events` - State transitions
- `pipeline_templates` - Reusable configs
- `metrics_summaries` - Aggregated results

New tables needed:
- `agent_tasks` - Task queue for agents
- `agent_status` - Agent heartbeats & status
- `active_pipelines` - Current pipeline state per host

## 2. Target Architecture Components

### 2.1 Phoenix API (Monolith)

**Modules to create:**
```
cmd/phoenix-api/
├── main.go                    # Entry point
├── config/                    # Configuration
├── internal/
│   ├── api/                   # HTTP/WebSocket handlers
│   │   ├── experiments.go     # Experiment endpoints
│   │   ├── agents.go          # Agent task/status endpoints
│   │   ├── pipelines.go       # Pipeline management
│   │   └── websocket.go       # Real-time updates
│   ├── controller/            # Experiment state machine (from controller service)
│   ├── analyzer/              # KPI & cost analysis (from benchmark/analytics)
│   ├── store/                 # Database layer
│   │   ├── models.go          # All domain models
│   │   ├── migrations.go      # Schema migrations
│   │   └── queries.go         # Complex queries
│   └── tasks/                 # Task queue management
│       ├── scheduler.go       # Task scheduling logic
│       └── executor.go        # Task state tracking
```

### 2.2 Phoenix Agent

**Structure:**
```
cmd/phoenix-agent/
├── main.go                    # Entry point
├── internal/
│   ├── poller/               # API polling logic
│   │   ├── client.go         # HTTP client to API
│   │   └── backoff.go        # Retry/backoff logic
│   ├── supervisor/           # Process management
│   │   ├── collector.go      # OTel collector spawning
│   │   ├── loadsim.go        # Load generator management
│   │   └── process.go        # Process lifecycle
│   ├── metrics/              # Self-metrics
│   │   └── reporter.go       # Pushgateway client
│   └── config/               # Agent configuration
```

### 2.3 Configuration Templates

**OTel Pipeline Templates** (replacing custom processors):
```yaml
# topk_approximation.yaml
processors:
  # Group by labels to track cardinality
  groupbyattrs:
    keys: [service.name, host.name, http.route]
  
  # Calculate metric rates
  metricsgeneration:
    metrics:
      - name: request_rate
        unit: "1/s"
        type: gauge
        value: rate(http.server.duration)
  
  # Filter top contributors using transform processor
  transform:
    metric_statements:
      - context: metric
        statements:
          # Lua script to keep only top K
          - |
            local sorted = {}
            for k,v in pairs(metrics) do
              table.insert(sorted, {key=k, value=v})
            end
            table.sort(sorted, function(a,b) return a.value > b.value end)
            -- Keep only top 10
            for i=11,#sorted do
              metrics[sorted[i].key] = nil
            end
```

## 3. Detailed Migration Plan

### 3.1 Database Migration

**Step 1: Add new tables (Week 1)**

```sql
-- Agent task queue
CREATE TABLE agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    task_type VARCHAR(50) NOT NULL, -- 'collector', 'loadsim'
    action VARCHAR(50) NOT NULL,     -- 'start', 'stop', 'update'
    config JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    INDEX idx_host_status (host_id, status)
);

-- Agent status tracking
CREATE TABLE agent_status (
    host_id VARCHAR(255) PRIMARY KEY,
    last_heartbeat TIMESTAMP NOT NULL,
    agent_version VARCHAR(50),
    capabilities JSONB,
    active_tasks JSONB,
    metrics JSONB,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Active pipelines per host
CREATE TABLE active_pipelines (
    host_id VARCHAR(255) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    experiment_id VARCHAR(50),
    config_url TEXT NOT NULL,
    process_info JSONB,
    started_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (host_id, variant)
);
```

### 3.2 Code Migration Strategy

#### Phase 1: API Consolidation (Sprint A - Week 1-2)

**1.1 Create phoenix-api base structure**
```bash
# Initialize module
cd projects/phoenix-api
go mod init github.com/phoenix/platform/projects/phoenix-api

# Copy base from platform-api
cp -r ../platform-api/cmd/api/main.go cmd/
cp -r ../platform-api/internal/api internal/
cp -r ../platform-api/internal/config internal/
```

**1.2 Migrate controller logic**
```go
// internal/controller/experiment_controller.go
// Port from projects/controller/internal/controller/
package controller

import (
    "context"
    "fmt"
    "github.com/phoenix/platform/projects/phoenix-api/internal/store"
    "github.com/phoenix/platform/projects/phoenix-api/internal/tasks"
)

type ExperimentController struct {
    store      *store.Store
    taskQueue  *tasks.Queue
    metrics    *MetricsClient
}

func (c *ExperimentController) StartExperiment(ctx context.Context, exp *Experiment) error {
    // Instead of creating CRDs, enqueue tasks
    for _, host := range exp.Config.TargetHosts {
        // Baseline collector task
        baselineTask := &tasks.Task{
            HostID:   host,
            Type:     "collector",
            Action:   "start",
            Config: map[string]interface{}{
                "id":         fmt.Sprintf("%s-baseline", exp.ID),
                "variant":    "baseline",
                "configUrl":  exp.Config.BaselineTemplate.URL,
                "vars":       exp.Config.BaselineTemplate.Variables,
            },
        }
        c.taskQueue.Enqueue(baselineTask)
        
        // Candidate collector task
        candidateTask := &tasks.Task{
            HostID:   host,
            Type:     "collector", 
            Action:   "start",
            Config: map[string]interface{}{
                "id":         fmt.Sprintf("%s-candidate", exp.ID),
                "variant":    "candidate",
                "configUrl":  exp.Config.CandidateTemplate.URL,
                "vars":       exp.Config.CandidateTemplate.Variables,
            },
        }
        c.taskQueue.Enqueue(candidateTask)
    }
    
    return c.store.UpdateExperimentPhase(ctx, exp.ID, "deploying")
}
```

**1.3 Add agent endpoints**
```go
// internal/api/agents.go
package api

import (
    "encoding/json"
    "net/http"
    "time"
)

// GET /agent/v1/tasks?host=<id>
func (s *Server) handleAgentTasks(w http.ResponseWriter, r *http.Request) {
    hostID := r.URL.Query().Get("host")
    if hostID == "" {
        http.Error(w, "host parameter required", http.StatusBadRequest)
        return
    }
    
    // Long polling with 30s timeout
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    
    tasks, err := s.taskQueue.GetPendingTasks(ctx, hostID)
    if err != nil {
        s.logger.Error("failed to get tasks", zap.Error(err))
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}

// POST /agent/v1/status
func (s *Server) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
    var status AgentStatus
    if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    // Update agent heartbeat
    if err := s.store.UpdateAgentStatus(r.Context(), &status); err != nil {
        s.logger.Error("failed to update agent status", zap.Error(err))
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    
    // Broadcast status update via WebSocket
    s.broadcast <- WSMessage{
        Type: "agent_update",
        Data: status,
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

**1.4 Port benchmark/analytics logic**
```go
// internal/analyzer/kpi_calculator.go
package analyzer

import (
    "context"
    "fmt"
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
)

type KPICalculator struct {
    promClient v1.API
}

func (k *KPICalculator) CalculateExperimentKPIs(ctx context.Context, expID string, duration time.Duration) (*KPIResult, error) {
    endTime := time.Now()
    startTime := endTime.Add(-duration)
    
    // Calculate cardinality reduction
    baselineQuery := fmt.Sprintf(`
        count(count by (__name__)({experiment_id="%s",variant="baseline"}))
    `, expID)
    
    candidateQuery := fmt.Sprintf(`
        count(count by (__name__)({experiment_id="%s",variant="candidate"}))
    `, expID)
    
    baselineCardinality, err := k.queryValue(ctx, baselineQuery, endTime)
    if err != nil {
        return nil, fmt.Errorf("baseline cardinality query failed: %w", err)
    }
    
    candidateCardinality, err := k.queryValue(ctx, candidateQuery, endTime)
    if err != nil {
        return nil, fmt.Errorf("candidate cardinality query failed: %w", err)
    }
    
    reduction := (1 - candidateCardinality/baselineCardinality) * 100
    
    // Calculate resource usage
    cpuQuery := fmt.Sprintf(`
        avg(rate(container_cpu_usage_seconds_total{
            pod=~"otel-collector-.*",
            experiment_id="%s"
        }[5m])) by (variant)
    `, expID)
    
    // ... more KPI calculations
    
    return &KPIResult{
        CardinalityReduction: reduction,
        CPUUsage:            cpuResults,
        MemoryUsage:         memResults,
        IngestRate:          ingestResults,
    }, nil
}
```

#### Phase 2: Agent Implementation (Sprint B - Week 3-4)

**2.1 Agent core structure**
```go
// cmd/phoenix-agent/main.go
package main

import (
    "context"
    "flag"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/phoenix/platform/cmd/phoenix-agent/internal/poller"
    "github.com/phoenix/platform/cmd/phoenix-agent/internal/supervisor"
    "github.com/phoenix/platform/cmd/phoenix-agent/internal/metrics"
)

func main() {
    var (
        apiURL     = flag.String("api-url", "http://phoenix-api:8080", "Phoenix API URL")
        hostID     = flag.String("host-id", getHostID(), "Unique host identifier")
        pollInterval = flag.Duration("poll-interval", 15*time.Second, "Task poll interval")
    )
    flag.Parse()
    
    // Initialize components
    client := poller.NewClient(*apiURL, *hostID)
    sup := supervisor.New()
    reporter := metrics.NewReporter(*apiURL, *hostID)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Start polling loop
    go func() {
        ticker := time.NewTicker(*pollInterval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                tasks, err := client.GetTasks(ctx)
                if err != nil {
                    log.Printf("Failed to get tasks: %v", err)
                    continue
                }
                
                for _, task := range tasks {
                    if err := handleTask(sup, task); err != nil {
                        log.Printf("Task %s failed: %v", task.ID, err)
                        client.ReportTaskStatus(task.ID, "failed", err.Error())
                    }
                }
                
                // Report metrics
                reporter.Report(sup.GetMetrics())
                
            case <-ctx.Done():
                return
            }
        }
    }()
    
    // Wait for shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    
    log.Println("Shutting down agent...")
    sup.StopAll()
}
```

**2.2 Supervisor implementation**
```go
// internal/supervisor/collector.go
package supervisor

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "text/template"
)

type CollectorManager struct {
    processes map[string]*Process
    configDir string
}

func (m *CollectorManager) Start(id, variant, configURL string, vars map[string]string) error {
    // Download and process config
    config, err := m.downloadConfig(configURL)
    if err != nil {
        return fmt.Errorf("failed to download config: %w", err)
    }
    
    // Apply variable substitution
    processedConfig, err := m.applyVariables(config, vars)
    if err != nil {
        return fmt.Errorf("failed to apply variables: %w", err)
    }
    
    // Write config to disk
    configPath := filepath.Join(m.configDir, fmt.Sprintf("%s.yaml", id))
    if err := os.WriteFile(configPath, []byte(processedConfig), 0644); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    // Prepare command
    cmd := exec.Command(
        "otelcol-contrib",
        "--config", configPath,
        "--set", fmt.Sprintf("service.telemetry.metrics.address=:0"), // Disable default metrics
    )
    
    // Set environment variables
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("EXPERIMENT_ID=%s", id),
        fmt.Sprintf("VARIANT=%s", variant),
        fmt.Sprintf("HOST_ID=%s", m.hostID),
        // Add Pushgateway config
        "METRICS_PUSHGATEWAY_URL=http://prometheus-pushgateway:9091",
    )
    
    // Start process
    process := &Process{
        ID:      id,
        Variant: variant,
        Cmd:     cmd,
    }
    
    if err := process.Start(); err != nil {
        return fmt.Errorf("failed to start collector: %w", err)
    }
    
    m.processes[id] = process
    return nil
}
```

**2.3 Load simulator integration**
```go
// internal/supervisor/loadsim.go
package supervisor

import (
    "context"
    "fmt"
    "os/exec"
    "time"
)

type LoadSimManager struct {
    activeJob *exec.Cmd
}

func (m *LoadSimManager) StartLoadProfile(profile string, duration time.Duration) error {
    if m.activeJob != nil {
        return fmt.Errorf("load simulation already running")
    }
    
    script := m.getProfileScript(profile)
    
    ctx, cancel := context.WithTimeout(context.Background(), duration)
    m.activeJob = exec.CommandContext(ctx, "bash", "-c", script)
    
    go func() {
        defer cancel()
        if err := m.activeJob.Run(); err != nil {
            log.Printf("Load simulation ended: %v", err)
        }
        m.activeJob = nil
    }()
    
    return nil
}

func (m *LoadSimManager) getProfileScript(profile string) string {
    switch profile {
    case "high-card":
        return `
            while true; do
                # Generate high cardinality metrics
                for i in {1..1000}; do
                    curl -X POST http://localhost:4318/v1/metrics \
                        -H "Content-Type: application/json" \
                        -d "{
                            \"resource\": {
                                \"attributes\": [{
                                    \"key\": \"service.name\",
                                    \"value\": {\"stringValue\": \"load-test\"}
                                }]
                            },
                            \"scopeMetrics\": [{
                                \"metrics\": [{
                                    \"name\": \"test.metric\",
                                    \"gauge\": {
                                        \"dataPoints\": [{
                                            \"asDouble\": $RANDOM,
                                            \"attributes\": [{
                                                \"key\": \"user.id\",
                                                \"value\": {\"stringValue\": \"user-$i\"}
                                            }]
                                        }]
                                    }
                                }]
                            }]
                        }"
                done
                sleep 1
            done
        `
    case "normal":
        return `stress-ng --cpu 2 --io 2 --vm 1 --vm-bytes 128M --timeout 60s`
    default:
        return `echo "Unknown profile: ${profile}"`
    }
}
```

### 3.3 OTel Configuration Migration

**Transform custom processors to stock configs:**

```yaml
# Replace topk processor
processors:
  # Step 1: Add experiment labels
  attributes:
    actions:
      - key: experiment_id
        value: ${EXPERIMENT_ID}
        action: insert
      - key: variant
        value: ${VARIANT}
        action: insert
      - key: host
        value: ${HOST_ID}
        action: insert
  
  # Step 2: Group metrics for cardinality tracking
  groupbyattrs:
    keys: [service.name, http.route, http.method]
  
  # Step 3: Filter with expressions (replacing adaptive filter)
  filter/adaptive:
    metrics:
      metric:
        # Drop low-value metrics
        - 'name == "process.cpu.utilization" and value < 0.01'
        - 'name == "process.memory.usage" and value < 10485760'  # 10MB
  
  # Step 4: Transform for top-k approximation
  transform:
    metric_statements:
      - context: datapoint
        statements:
          # Add ranking attribute based on value
          - set(attributes["rank"], value)
    
  # Step 5: Use tail sampling for top contributors
  probabilistic_sampler:
    sampling_percentage: 10  # Keep top 10%
    attribute_source: "rank"

exporters:
  # Push to central gateway instead of individual endpoints
  prometheusremotewrite:
    endpoint: ${METRICS_PUSHGATEWAY_URL}/metrics/job/phoenix-collector/instance/${HOST_ID}
    add_metric_suffixes: false
    external_labels:
      experiment_id: ${EXPERIMENT_ID}
      variant: ${VARIANT}
      host: ${HOST_ID}
```

### 3.4 Deployment Migration

**From Helm charts with operators to simple deployment:**

```yaml
# phoenix-lean/helm/values.yaml
phoenix-api:
  image: phoenix/api:latest
  replicas: 2
  env:
    DATABASE_URL: postgresql://phoenix:phoenix@postgres:5432/phoenix
    PROMETHEUS_URL: http://prometheus:9090
    PUSHGATEWAY_URL: http://prometheus-pushgateway:9091
  
phoenix-agent:
  image: phoenix/agent:latest
  daemonset: true
  env:
    API_URL: http://phoenix-api:8080
    POLL_INTERVAL: 15s
  hostNetwork: true
  hostPID: true
  volumes:
    - name: dockersock
      hostPath:
        path: /var/run/docker.sock
    - name: cgroup
      hostPath:
        path: /sys/fs/cgroup

prometheus-stack:
  prometheus:
    config:
      global:
        scrape_interval: 15s
      scrape_configs:
        - job_name: pushgateway
          static_configs:
            - targets: ['prometheus-pushgateway:9091']
  
  pushgateway:
    enabled: true
    persistence:
      enabled: true
      size: 10Gi
```

## 4. Testing & Validation Strategy

### 4.1 Integration Tests

```go
// tests/integration/lean_architecture_test.go
func TestLeanArchitectureEndToEnd(t *testing.T) {
    // Start test environment
    env := startTestEnv(t)
    defer env.Cleanup()
    
    // Test 1: Agent picks up tasks
    t.Run("AgentTaskPolling", func(t *testing.T) {
        // Create experiment
        exp := createTestExperiment(t, env.API)
        
        // Wait for agent to pick up tasks
        require.Eventually(t, func() bool {
            status := getAgentStatus(t, env.API, "test-host")
            return len(status.ActiveTasks) == 2 // baseline + candidate
        }, 30*time.Second, 1*time.Second)
    })
    
    // Test 2: Metrics flow to Pushgateway
    t.Run("MetricsPushgateway", func(t *testing.T) {
        // Query Pushgateway
        metrics := queryPushgateway(t, env.Prometheus, "phoenix_collector")
        require.NotEmpty(t, metrics)
        
        // Verify labels
        require.Contains(t, metrics[0].Labels, "experiment_id")
        require.Contains(t, metrics[0].Labels, "variant")
    })
    
    // Test 3: KPI calculation works
    t.Run("KPICalculation", func(t *testing.T) {
        result := calculateKPIs(t, env.API, exp.ID)
        require.Greater(t, result.CardinalityReduction, 50.0)
    })
}
```

### 4.2 Migration Testing

**Parallel run strategy:**
1. Deploy lean architecture alongside existing
2. Mirror experiments to both systems
3. Compare results
4. Gradual cutover

## 5. Rollout Plan

### Week 1-2 (Sprint A)
- [ ] Create phoenix-api structure
- [ ] Migrate controller logic
- [ ] Add agent endpoints
- [ ] Port analyzer modules
- [ ] Database migrations
- [ ] Basic integration tests

### Week 3-4 (Sprint B)  
- [ ] Implement phoenix-agent
- [ ] Create supervisor modules
- [ ] OTel config templates
- [ ] Pushgateway integration
- [ ] E2E testing
- [ ] Documentation

### Week 5 (Deployment)
- [ ] Deploy to staging
- [ ] Parallel run tests
- [ ] Performance validation
- [ ] Gradual production rollout

## 6. Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Agent polling misses tasks | High | Implement reliable queue with acknowledgments |
| Pushgateway becomes bottleneck | Medium | Use sharding, multiple instances |
| Config template errors | High | Extensive validation, dry-run mode |
| Migration data loss | Critical | Full backup, reversible migrations |

## 7. Success Metrics

- **Code reduction**: >60% fewer lines of code
- **Deployment time**: <5 minutes (from 20+ minutes)
- **Debug time**: <30 minutes to trace issues (from 2+ hours)
- **Resource usage**: 50% less CPU/memory for control plane
- **Operational toil**: 80% reduction in kubectl commands needed