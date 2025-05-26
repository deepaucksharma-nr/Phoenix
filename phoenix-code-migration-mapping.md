# Phoenix Code Migration Mapping & Analysis

## 1. Detailed Code Inventory & Migration Paths

### 1.1 Controller Service Migration

**Current Structure:**
```
projects/controller/
├── cmd/controller/main.go (196 lines)
├── internal/
│   ├── clients/          # gRPC clients to other services
│   ├── controller/       # Core experiment logic
│   ├── grpc/            # gRPC server implementation
│   └── store/           # PostgreSQL store
```

**Migration Actions:**

| File/Package | Current Function | Migration Path | New Location |
|--------------|-----------------|----------------|--------------|
| `main.go` | gRPC server setup | Convert to HTTP handlers | `phoenix-api/cmd/main.go` |
| `clients/platform_client.go` | Call platform API | Remove (same process now) | N/A |
| `clients/benchmark_client.go` | Call benchmark service | Inline as analyzer module | `phoenix-api/internal/analyzer/` |
| `controller/experiment.go` | State machine | Keep, modify CRD calls → task queue | `phoenix-api/internal/controller/` |
| `controller/state_machine.go` | Phase transitions | Keep, add task generation | `phoenix-api/internal/controller/` |
| `grpc/server.go` | gRPC handlers | Convert to REST | `phoenix-api/internal/api/experiments.go` |
| `store/postgres.go` | DB operations | Merge with platform store | `phoenix-api/internal/store/` |

**Key Code Changes:**

```go
// BEFORE: controller/experiment.go
func (c *Controller) deployPipeline(ctx context.Context, exp *Experiment) error {
    ppp := &phoenixv1alpha1.PhoenixProcessPipeline{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("%s-baseline", exp.ID),
        },
        Spec: phoenixv1alpha1.PhoenixProcessPipelineSpec{
            ExperimentID: exp.ID,
            Variant:      "baseline",
            ConfigMap:    exp.Config.BaselineTemplate.ConfigMapName,
        },
    }
    return c.k8sClient.Create(ctx, ppp)
}

// AFTER: phoenix-api/internal/controller/experiment.go  
func (c *Controller) deployPipeline(ctx context.Context, exp *Experiment) error {
    for _, host := range exp.Config.TargetHosts {
        task := &Task{
            HostID: host,
            Type:   "collector",
            Action: "start",
            Config: map[string]interface{}{
                "id":        fmt.Sprintf("%s-baseline", exp.ID),
                "variant":   "baseline",
                "configUrl": c.getConfigURL(exp.Config.BaselineTemplate),
                "vars":      exp.Config.BaselineTemplate.Variables,
            },
        }
        if err := c.taskQueue.Enqueue(ctx, task); err != nil {
            return fmt.Errorf("failed to enqueue task for host %s: %w", host, err)
        }
    }
    return nil
}
```

### 1.2 Platform API Absorption

**Current Structure:**
```
projects/platform-api/
├── cmd/api/main.go
├── internal/
│   ├── api/              # REST endpoints
│   ├── auth/             # Authentication
│   ├── config/           # Configuration
│   ├── middleware/       # HTTP middleware
│   ├── services/         # Business logic
│   └── websocket/        # WebSocket handling
```

**This becomes the base for phoenix-api**

| Component | Action | Notes |
|-----------|--------|-------|
| `main.go` | Keep as base | Add agent endpoints |
| `api/handlers.go` | Extend | Add `/agent/v1/*` routes |
| `services/pipeline_status_aggregator.go` | Rewrite | Query agent_status table instead of K8s |
| `websocket/hub.go` | Keep | Add agent status broadcasts |

### 1.3 Pipeline Operator Decomposition

**Current CRD Controller:**
```go
// controllers/phoenixprocesspipeline_controller.go
func (r *PhoenixProcessPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Get CRD
    var ppp phoenixv1alpha1.PhoenixProcessPipeline
    if err := r.Get(ctx, req.NamespacedName, &ppp); err != nil {
        return ctrl.Result{}, err
    }
    
    // Create DaemonSet
    ds := r.buildDaemonSet(&ppp)
    if err := r.Create(ctx, ds); err != nil {
        return ctrl.Result{}, err
    }
}
```

**Becomes Agent Task Handler:**
```go
// phoenix-agent/internal/supervisor/task_handler.go
func (h *TaskHandler) HandleCollectorTask(task *Task) error {
    switch task.Action {
    case "start":
        config := task.Config
        return h.collector.Start(
            config["id"].(string),
            config["variant"].(string), 
            config["configUrl"].(string),
            config["vars"].(map[string]string),
        )
    case "stop":
        return h.collector.Stop(config["id"].(string))
    case "update":
        // Stop and restart with new config
        h.collector.Stop(config["id"].(string))
        return h.collector.Start(/* new config */)
    }
}
```

### 1.4 Custom OTel Processors → Config Templates

**Current Top-K Processor (68 lines of Go):**
```go
// pkg/otel/processors/topk/topk.go
func (p *Processor) Process(metrics []otel.Metric) []otel.Metric {
    // Sort metrics by value
    sort.Slice(target, func(i, j int) bool {
        return target[i].Value > target[j].Value
    })
    // Keep only top K
    if len(target) > p.topK {
        target = target[:p.topK]
    }
    return append(rest, target...)
}
```

**Becomes OTel Config:**
```yaml
# templates/processors/topk.yaml
processors:
  # Step 1: Group by metric name
  groupbyattrs:
    keys: [__name__]
  
  # Step 2: Sort and filter using transform
  transform:
    error_mode: ignore
    metric_statements:
      - context: metric
        conditions:
          - name == "${METRIC_NAME}"
        statements:
          - |
            # Lua script for top-k filtering
            local points = {}
            for i, dp in ipairs(datapoints) do
              table.insert(points, {idx=i, val=dp.value})
            end
            table.sort(points, function(a,b) return a.val > b.val end)
            
            # Keep only top K
            local top_k = ${TOP_K}
            local new_points = {}
            for i = 1, math.min(top_k, #points) do
              new_points[i] = datapoints[points[i].idx]
            end
            datapoints = new_points
```

**Current Adaptive Filter (67 lines):**
```go
// pkg/otel/processors/adaptivefilter/adaptive_filter.go
func (p *Processor) Process(metrics []otel.Metric) []otel.Metric {
    var out []otel.Metric
    for _, m := range metrics {
        if m.Name == "process.cpu.utilization" && m.Value < p.cpu {
            continue
        }
        if m.Name == "process.memory.usage" && m.Value < p.mem {
            continue
        }
        out = append(out, m)
    }
}
```

**Becomes:**
```yaml
# templates/processors/adaptive_filter.yaml
processors:
  filter:
    error_mode: ignore
    metrics:
      datapoint:
        - 'metric.name == "process.cpu.utilization" and value < ${CPU_THRESHOLD}'
        - 'metric.name == "process.memory.usage" and value < ${MEMORY_THRESHOLD}'
```

## 2. Database Schema Migration Details

### 2.1 Current Schema Analysis

```sql
-- From store/postgres.go migrations
CREATE TABLE experiments (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phase VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    status JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE experiment_events (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50) REFERENCES experiments(id),
    event_type VARCHAR(50) NOT NULL,
    phase VARCHAR(50),
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE pipeline_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    config_url TEXT NOT NULL,
    variables JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### 2.2 New Tables Required

```sql
-- Task queue for agents
CREATE TABLE agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(50) REFERENCES experiments(id),
    task_type VARCHAR(50) NOT NULL CHECK (task_type IN ('collector', 'loadsim', 'command')),
    action VARCHAR(50) NOT NULL CHECK (action IN ('start', 'stop', 'update', 'execute')),
    config JSONB NOT NULL,
    priority INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'running', 'completed', 'failed')),
    assigned_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_host_status (host_id, status),
    INDEX idx_experiment (experiment_id),
    INDEX idx_created (created_at)
);

-- Agent heartbeat and status
CREATE TABLE agent_status (
    host_id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255),
    ip_address INET,
    agent_version VARCHAR(50),
    started_at TIMESTAMP,
    last_heartbeat TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'healthy' CHECK (status IN ('healthy', 'degraded', 'unhealthy', 'offline')),
    capabilities JSONB DEFAULT '{}',
    active_tasks JSONB DEFAULT '[]',
    resource_usage JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Active pipelines tracking (replaces CRD state)
CREATE TABLE active_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(50) REFERENCES experiments(id),
    variant VARCHAR(50) NOT NULL CHECK (variant IN ('baseline', 'candidate')),
    config_url TEXT NOT NULL,
    config_hash VARCHAR(64),
    process_info JSONB DEFAULT '{}',
    metrics_info JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'running' CHECK (status IN ('starting', 'running', 'stopping', 'stopped', 'failed')),
    started_at TIMESTAMP DEFAULT NOW(),
    stopped_at TIMESTAMP,
    UNIQUE(host_id, variant)
);

-- Metrics cache for faster queries
CREATE TABLE metrics_cache (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50) REFERENCES experiments(id),
    timestamp TIMESTAMP NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    host_id VARCHAR(255),
    value DOUBLE PRECISION,
    labels JSONB DEFAULT '{}',
    INDEX idx_experiment_time (experiment_id, timestamp),
    INDEX idx_metric_variant (metric_name, variant)
);
```

### 2.3 Migration Scripts

```sql
-- Migration 001: Add lean-core tables
BEGIN;

-- Create new tables
CREATE TABLE agent_tasks (...);
CREATE TABLE agent_status (...);
CREATE TABLE active_pipelines (...);
CREATE TABLE metrics_cache (...);

-- Add columns to existing tables
ALTER TABLE experiments ADD COLUMN deployment_mode VARCHAR(50) DEFAULT 'kubernetes';
ALTER TABLE experiments ADD COLUMN target_hosts TEXT[];

-- Create views for compatibility
CREATE VIEW deployment_status AS
SELECT 
    e.id as experiment_id,
    e.phase,
    COUNT(DISTINCT ap.host_id) as active_hosts,
    jsonb_agg(jsonb_build_object(
        'host', ap.host_id,
        'baseline_status', MAX(CASE WHEN ap.variant = 'baseline' THEN ap.status END),
        'candidate_status', MAX(CASE WHEN ap.variant = 'candidate' THEN ap.status END)
    )) as host_details
FROM experiments e
LEFT JOIN active_pipelines ap ON e.id = ap.experiment_id
GROUP BY e.id, e.phase;

COMMIT;
```

## 3. API Endpoint Mapping

### 3.1 Service Consolidation Map

| Current Service | Endpoint | New Phoenix API Endpoint | Changes |
|-----------------|----------|-------------------------|---------|
| **Controller** | gRPC :50051 | | |
| | `CreateExperiment` | `POST /api/v1/experiments` | gRPC → REST |
| | `GetExperiment` | `GET /api/v1/experiments/:id` | gRPC → REST |
| | `UpdatePhase` | `PUT /api/v1/experiments/:id/phase` | gRPC → REST |
| | `StreamEvents` | `WS /api/v1/experiments/:id/events` | gRPC stream → WebSocket |
| **Platform API** | HTTP :8080 | | |
| | `GET /api/pipelines` | `GET /api/v1/pipelines` | Query active_pipelines table |
| | `GET /api/metrics/summary` | `GET /api/v1/metrics/summary` | Query Pushgateway |
| | `WS /ws` | `WS /api/v1/ws` | Add agent events |
| **Benchmark** | gRPC :50052 | | |
| | `CalculateKPIs` | `POST /api/v1/experiments/:id/kpis` | Inline calculation |
| | `GetCostAnalysis` | `GET /api/v1/experiments/:id/cost` | Inline with PromQL |
| **Analytics** | HTTP :8082 | | |
| | `POST /analyze` | `POST /api/v1/experiments/:id/analyze` | Inline analyzer |

### 3.2 New Agent Endpoints

```go
// phoenix-api/internal/api/routes.go
func SetupRoutes(r *chi.Mux, s *Server) {
    // Existing endpoints (from platform-api)
    r.Route("/api/v1", func(r chi.Router) {
        // ... existing routes ...
        
        // New agent endpoints
        r.Route("/agent", func(r chi.Router) {
            r.Use(AgentAuthMiddleware)
            
            // Task polling (long-poll with 30s timeout)
            r.Get("/tasks", s.handleAgentGetTasks)
            
            // Task status updates
            r.Post("/tasks/{taskId}/status", s.handleTaskStatusUpdate)
            
            // Agent heartbeat
            r.Post("/heartbeat", s.handleAgentHeartbeat)
            
            // Metrics push (batch)
            r.Post("/metrics", s.handleAgentMetrics)
            
            // Log streaming
            r.Post("/logs", s.handleAgentLogs)
        })
    })
}
```

## 4. Configuration File Transformations

### 4.1 Pipeline Templates Migration

**Current: ConfigMap-based**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: baseline-config
data:
  config.yaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
    processors:
      batch:
        timeout: 1s
    exporters:
      prometheus:
        endpoint: 0.0.0.0:8889
```

**New: URL-based with variables**
```yaml
# Stored in S3/GCS/HTTP server
# URL: https://configs.phoenix.io/pipelines/baseline-v1.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  # Add experiment metadata
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
  
  batch:
    timeout: ${BATCH_TIMEOUT:-1s}
    send_batch_size: ${BATCH_SIZE:-1000}

exporters:
  # Push to central gateway
  prometheusremotewrite:
    endpoint: ${PUSHGATEWAY_URL}/metrics/job/collector/instance/${HOST_ID}
    external_labels:
      experiment_id: ${EXPERIMENT_ID}
      variant: ${VARIANT}
```

### 4.2 Helm Values Transformation

**Current: Complex multi-service**
```yaml
# Current values.yaml (simplified)
controller:
  enabled: true
  image: phoenix/controller:latest
  
platformApi:
  enabled: true
  image: phoenix/platform-api:latest
  
pipelineOperator:
  enabled: true
  image: phoenix/pipeline-operator:latest
  
prometheusOperator:
  enabled: true
```

**New: Lean deployment**
```yaml
# Lean values.yaml
phoenix:
  api:
    image: phoenix/api:latest
    replicas: 2
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
    env:
      - name: DATABASE_URL
        valueFrom:
          secretKeyRef:
            name: phoenix-db
            key: url
  
  agent:
    image: phoenix/agent:latest
    updateStrategy: RollingUpdate
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
    # Agent runs as DaemonSet
    nodeSelector: {}
    tolerations:
      - operator: Exists
        
monitoring:
  prometheus:
    retention: 7d
    storage: 50Gi
  
  pushgateway:
    persistence: true
    size: 10Gi
```

## 5. Testing Migration

### 5.1 Test File Mappings

| Current Test | Location | Migration Needed | New Location |
|--------------|----------|------------------|--------------|
| Controller integration | `projects/controller/tests/` | Update K8s calls → API calls | `tests/integration/controller_test.go` |
| Pipeline operator E2E | `projects/pipeline-operator/e2e/` | Rewrite for agent model | `tests/e2e/agent_pipeline_test.go` |
| Benchmark calculations | `projects/benchmark/tests/` | Move to API tests | `tests/integration/kpi_test.go` |
| Load simulation | `projects/loadsim-operator/tests/` | Test via agent | `tests/integration/loadsim_test.go` |

### 5.2 New Test Patterns

```go
// tests/e2e/lean_architecture_test.go
func TestAgentTaskExecution(t *testing.T) {
    // Setup
    env := newTestEnvironment(t)
    defer env.Cleanup()
    
    // Start mock agent
    agent := env.StartAgent("test-host-1")
    
    // Create experiment
    exp := &Experiment{
        ID:   "exp-test1234",
        Name: "Test Experiment",
        Config: ExperimentConfig{
            TargetHosts: []string{"test-host-1"},
            BaselineTemplate: Template{
                URL: env.ConfigServer.URL + "/baseline.yaml",
            },
        },
    }
    
    err := env.API.CreateExperiment(exp)
    require.NoError(t, err)
    
    // Verify agent picks up task
    require.Eventually(t, func() bool {
        tasks := agent.GetActiveTasks()
        return len(tasks) > 0 && tasks[0].Type == "collector"
    }, 30*time.Second, 1*time.Second)
    
    // Verify metrics appear in Pushgateway
    require.Eventually(t, func() bool {
        metrics := env.QueryPushgateway("experiment_id", exp.ID)
        return len(metrics) > 0
    }, 30*time.Second, 1*time.Second)
}
```

## 6. Rollback Plan

### 6.1 Feature Flags

```go
// phoenix-api/internal/config/features.go
type Features struct {
    UseLeanAgents      bool `env:"FEATURE_LEAN_AGENTS" default:"false"`
    UseKubernetesMode  bool `env:"FEATURE_K8S_MODE" default:"true"`
    UsePushgateway     bool `env:"FEATURE_PUSHGATEWAY" default:"false"`
}

// Dual mode support in controller
func (c *Controller) deployExperiment(ctx context.Context, exp *Experiment) error {
    if c.features.UseLeanAgents {
        return c.deployViaAgents(ctx, exp)
    }
    return c.deployViaOperators(ctx, exp)
}
```

### 6.2 Data Compatibility Layer

```sql
-- View to make new tables look like old for compatibility
CREATE VIEW kubernetes_deployments AS
SELECT 
    ap.experiment_id || '-' || ap.variant as name,
    ap.experiment_id,
    ap.variant,
    'DaemonSet' as kind,
    'phoenix-system' as namespace,
    jsonb_build_object(
        'replicas', COUNT(DISTINCT ap.host_id),
        'readyReplicas', COUNT(DISTINCT CASE WHEN ap.status = 'running' THEN ap.host_id END)
    ) as status
FROM active_pipelines ap
GROUP BY ap.experiment_id, ap.variant;
```