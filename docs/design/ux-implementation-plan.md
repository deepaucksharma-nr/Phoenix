# Phoenix UX Revolution Implementation Plan

## Overview
This document outlines the monorepo-wide changes required to implement the Phoenix UX Revolution, focusing on the lean-core architecture with agent-based operations.

## Impact Analysis by Repository Structure

### 1. Top-Level Files Impact

#### `README.md`
- **Change**: Update to reflect new UI-first approach and simplified architecture
- **Add**: Screenshots of new UI features (cost flow, agent dashboard, visual pipeline builder)
- **Add**: Quick start focusing on UI experience rather than YAML configs

#### `QUICKSTART.md`
- **Rewrite**: Focus on UI-driven workflow
```markdown
# Old Flow
1. Write YAML configuration
2. Deploy via CLI
3. Monitor via CLI

# New Flow  
1. Open Phoenix Dashboard
2. Click "New Experiment"
3. Select hosts and pipeline template
4. Watch real-time cost savings
```

#### `docker-compose.yml`
```yaml
# Add new services for enhanced UI
services:
  # Existing services...
  
  # New: WebSocket hub for real-time updates
  phoenix-realtime:
    build: ./projects/phoenix-api
    ports:
      - "8081:8081"
    environment:
      - ENABLE_WEBSOCKET=true
      - WEBSOCKET_PORT=8081
  
  # New: Metrics aggregator for cost calculations
  phoenix-metrics-cache:
    image: redis:alpine
    ports:
      - "6379:6379"
```

### 2. `/pkg` - Shared Packages Updates

#### `/pkg/common/interfaces/`
```go
// New interfaces for real-time features
package interfaces

// AgentStatus represents real-time agent state
type AgentStatus struct {
    HostID          string
    Status          string // healthy, updating, offline
    ActiveTasks     []Task
    Metrics         AgentMetrics
    CostSavings     float64
    LastHeartbeat   time.Time
}

// MetricFlow represents real-time metric costs
type MetricFlow struct {
    MetricName      string
    CostPerMinute   float64
    Cardinality     int64
    Percentage      float64
}

// TaskProgress for UI updates
type TaskProgress struct {
    TaskID          string
    Type            string
    Progress        int
    TotalHosts      int
    CompletedHosts  int
    ETA             time.Duration
}
```

#### `/pkg/common/websocket/` (NEW)
```go
// New package for WebSocket management
package websocket

type EventType string

const (
    EventAgentStatus    EventType = "agent_status"
    EventExperimentUpdate EventType = "experiment_update"
    EventMetricFlow     EventType = "metric_flow"
    EventTaskProgress   EventType = "task_progress"
)

type Event struct {
    Type      EventType
    Timestamp time.Time
    Data      interface{}
}
```

#### `/pkg/common/metrics/`
```go
// Enhanced metrics for real-time cost calculation
package metrics

// CostCalculator provides real-time cost estimates
type CostCalculator interface {
    CalculateMetricCost(metric string, cardinality int64) float64
    GetCostBreakdown() map[string]float64
    ProjectMonthlySavings(current, optimized float64) float64
}
```

### 3. `/projects/phoenix-api/` - Major Changes

#### New Endpoints for UI
```go
// internal/api/routes.go
func SetupRoutes(router *gin.Engine) {
    // Existing routes...
    
    // New UI-focused endpoints
    api := router.Group("/api/v1")
    {
        // Real-time cost flow
        api.GET("/metrics/cost-flow", h.GetMetricCostFlow)
        api.GET("/metrics/cardinality", h.GetCardinalityBreakdown)
        
        // Agent fleet management
        api.GET("/fleet/status", h.GetFleetStatus)
        api.GET("/fleet/map", h.GetAgentMap)
        
        // Simplified experiment creation
        api.POST("/experiments/wizard", h.CreateExperimentWizard)
        api.GET("/pipelines/templates", h.GetPipelineTemplates)
        api.POST("/pipelines/preview", h.PreviewPipelineImpact)
        
        // Task visibility
        api.GET("/tasks/active", h.GetActiveTasks)
        api.GET("/tasks/queue", h.GetTaskQueue)
        
        // WebSocket endpoint
        api.GET("/ws", h.WebSocketHandler)
    }
}
```

#### WebSocket Hub Implementation
```go
// internal/websocket/hub.go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan Event
    register   chan *Client
    unregister chan *Client
    
    // Channels for different event types
    agentUpdates     chan AgentStatus
    experimentUpdates chan ExperimentUpdate
    metricFlows      chan MetricFlow
    taskProgress     chan TaskProgress
}

func (h *Hub) Run() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Broadcast real-time updates
            h.broadcastMetricFlow()
            h.broadcastAgentStatus()
            
        case client := <-h.register:
            h.clients[client] = true
            // Send initial state
            h.sendInitialState(client)
            
        case update := <-h.agentUpdates:
            h.broadcast <- Event{
                Type: EventAgentStatus,
                Data: update,
            }
        }
    }
}
```

### 4. `/projects/dashboard/` - Complete Overhaul

#### Package.json Updates
```json
{
  "dependencies": {
    // Remove heavy dependencies
    "- @mui/material": "removed",
    "- redux": "removed",
    
    // Add performance-focused libraries
    "+ zustand": "^4.5.0",
    "+ @tanstack/react-virtual": "^3.0.0",
    "+ d3": "^7.8.0",
    "+ react-use-websocket": "^4.5.0",
    "+ framer-motion": "^11.0.0",
    "+ comlink": "^4.4.1"
  }
}
```

#### New Component Structure
```
src/
├── components/
│   ├── CostFlow/
│   │   ├── LiveCostMonitor.tsx
│   │   ├── MetricBreakdown.tsx
│   │   └── CostFlowChart.tsx
│   ├── Fleet/
│   │   ├── AgentMap.tsx
│   │   ├── FleetStatus.tsx
│   │   └── TaskQueue.tsx
│   ├── PipelineBuilder/
│   │   ├── DragDropCanvas.tsx
│   │   ├── ProcessorBlocks.tsx
│   │   └── LivePreview.tsx
│   ├── Experiments/
│   │   ├── WizardFlow.tsx
│   │   ├── RealtimeComparison.tsx
│   │   └── QuickActions.tsx
│   └── Analytics/
│       ├── CardinalityExplorer.tsx
│       ├── SunburstChart.tsx
│       └── ExecutiveDashboard.tsx
├── hooks/
│   ├── useRealtimeUpdates.ts
│   ├── useAgentStatus.ts
│   ├── useMetricFlow.ts
│   └── useKeyboardShortcuts.ts
├── workers/
│   ├── metricsAggregator.worker.ts
│   └── searchIndex.worker.ts
└── stores/
    ├── agentStore.ts
    ├── experimentStore.ts
    └── metricStore.ts
```

#### Core UI Components

```typescript
// components/CostFlow/LiveCostMonitor.tsx
export const LiveCostMonitor: React.FC = () => {
  const { metricFlows } = useMetricFlow();
  const [selectedMetric, setSelectedMetric] = useState<string | null>(null);
  
  return (
    <Card>
      <CardHeader>
        <Typography variant="h5">Live Cost Flow Monitor</Typography>
        <Chip label={`₹${totalCost}/min`} color="primary" />
      </CardHeader>
      <CardContent>
        <CostFlowVisualization 
          flows={metricFlows}
          onMetricSelect={setSelectedMetric}
        />
        {selectedMetric && (
          <MetricActions 
            metric={selectedMetric}
            onDeploy={handleQuickDeploy}
          />
        )}
      </CardContent>
    </Card>
  );
};
```

### 5. `/projects/phoenix-agent/` - Enhanced for UI Support

#### New Reporting Capabilities
```go
// internal/reporter/ui_metrics.go
type UIMetricsReporter struct {
    apiClient *client.APIClient
    ticker    *time.Ticker
}

func (r *UIMetricsReporter) Start() {
    for range r.ticker.C {
        metrics := r.collectUIMetrics()
        r.apiClient.ReportMetrics(metrics)
    }
}

func (r *UIMetricsReporter) collectUIMetrics() UIMetrics {
    return UIMetrics{
        CPUUsage:       getCurrentCPU(),
        MemoryUsage:    getCurrentMemory(),
        MetricsPerSec:  getMetricsThroughput(),
        ActivePipelines: getActivePipelines(),
        CostSavings:    calculateLocalSavings(),
    }
}
```

### 6. `/configs/` - New UI Configuration

#### `/configs/ui/` (NEW)
```yaml
# dashboard-config.yaml
ui:
  performance:
    max_websocket_connections: 10000
    metric_update_interval: 1s
    agent_poll_interval: 30s
    
  features:
    enable_cost_flow: true
    enable_cardinality_explorer: true
    enable_visual_pipeline_builder: true
    enable_keyboard_shortcuts: true
    
  thresholds:
    cost_alert_threshold: 10000  # ₹/hour
    cardinality_warning: 1000000  # metrics
    agent_offline_timeout: 60s
```

### 7. `/deployments/kubernetes/` - UI Service Additions

```yaml
# lean-architecture/phoenix-ui-enhanced.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: phoenix-dashboard
spec:
  replicas: 3  # Increased for WebSocket load
  template:
    spec:
      containers:
      - name: dashboard
        image: phoenix-dashboard:latest
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        env:
        - name: ENABLE_WEBSOCKET
          value: "true"
        - name: API_ENDPOINT
          value: "http://phoenix-api:8080"
---
apiVersion: v1
kind: Service
metadata:
  name: phoenix-dashboard
spec:
  ports:
  - port: 3000
    name: http
  - port: 8081
    name: websocket
```

### 8. `/tests/` - New UI Test Suites

#### `/tests/e2e/ui_workflows_test.go`
```go
func TestUIWorkflows(t *testing.T) {
    tests := []struct {
        name     string
        workflow func(*testing.T)
    }{
        {
            name: "Create experiment via wizard",
            workflow: testExperimentWizardFlow,
        },
        {
            name: "Real-time cost monitoring",
            workflow: testCostFlowUpdates,
        },
        {
            name: "Visual pipeline builder",
            workflow: testPipelineBuilderDragDrop,
        },
        {
            name: "One-click rollback",
            workflow: testInstantRollback,
        },
    }
}
```

### 9. `/scripts/` - UI Development Scripts

#### `/scripts/start-ui-dev.sh` (NEW)
```bash
#!/bin/bash
# Start UI development environment

# Start backend services
docker-compose up -d postgres phoenix-api

# Start mock agent fleet
./scripts/mock-agents.sh 10

# Start dashboard with hot reload
cd projects/dashboard
npm run dev

# Open browser
open http://localhost:3000
```

### 10. Database Migrations

#### `/projects/phoenix-api/migrations/003_ui_enhancements.sql`
```sql
-- UI-specific tables for performance

-- Metric cost cache for instant calculations
CREATE TABLE metric_cost_cache (
    metric_name TEXT PRIMARY KEY,
    cardinality BIGINT,
    cost_per_minute DECIMAL(10,2),
    last_updated TIMESTAMP,
    labels JSONB
);

-- Agent UI state
CREATE TABLE agent_ui_state (
    host_id TEXT PRIMARY KEY,
    display_name TEXT,
    group_name TEXT,
    location JSONB, -- for map view
    ui_metadata JSONB
);

-- Pipeline templates for wizard
CREATE TABLE pipeline_templates (
    id UUID PRIMARY KEY,
    name TEXT,
    description TEXT,
    category TEXT,
    config JSONB,
    estimated_savings_percent INT,
    ui_preview JSONB
);

-- Create indexes for UI performance
CREATE INDEX idx_metric_cost_cache_cost ON metric_cost_cache(cost_per_minute DESC);
CREATE INDEX idx_agent_tasks_status_host ON agent_tasks(status, host_id);
```

## Implementation Timeline

### Phase 1: Foundation (Weeks 1-2)
- Set up WebSocket infrastructure in phoenix-api
- Create base UI component library
- Implement real-time data stores

### Phase 2: Core Features (Weeks 3-4)
- Build cost flow monitor
- Implement agent fleet dashboard
- Create experiment wizard

### Phase 3: Advanced Features (Weeks 5-6)
- Visual pipeline builder
- Cardinality explorer
- Time machine rollback

### Phase 4: Polish & Performance (Weeks 7-8)
- Keyboard shortcuts
- Performance optimization
- E2E testing

## Success Criteria

1. **Performance**: All UI operations < 100ms response time
2. **Adoption**: 90% of users prefer UI over CLI within 30 days
3. **Efficiency**: Time to first optimization < 2 minutes
4. **Scale**: Support 10,000 concurrent WebSocket connections

## Risk Mitigation

1. **WebSocket Scale**: Use Redis pub/sub for multi-instance coordination
2. **Data Volume**: Implement smart aggregation and sampling
3. **Browser Performance**: Use Web Workers for heavy computations
4. **Backwards Compatibility**: Maintain CLI/API for automation