# Phoenix MVP Implementation Tracker

## Overview
This document tracks the implementation of all components required for Phoenix MVP readiness, based on the comprehensive development and validation plan.

## Component Status

### 1. Phoenix CLI (Command Fixes & Flow Finalization)

#### Issues to Fix:
- [ ] Pipeline deployment using wrong API path (`/api/v1/pipelines/deployments` vs `/api/v1/deployments`)
- [ ] Experiment metrics command calling non-existent endpoint
- [ ] Missing experiment rollback command implementation
- [ ] Missing pipeline rollback command alignment

#### Implementation Tasks:
```bash
# Files to modify:
# - projects/phoenix-cli/internal/client/api.go
# - projects/phoenix-cli/cmd/experiment_metrics.go
# - projects/phoenix-cli/cmd/pipeline_deploy.go
# - projects/phoenix-cli/cmd/experiment_rollback.go (create if missing)
```

#### Validation Signals:
- CLI commands return 2xx responses
- `phoenix experiment metrics <id>` returns real baseline vs candidate stats
- All commands in MVP checklist work without errors
- API logs show correct endpoints being hit

---

### 2. Phoenix API (Monolith) - Experiment & Deployment Lifecycle

#### Issues to Fix:
- [ ] Missing unified `/experiments/{id}/metrics` endpoint
- [ ] Incomplete experiment state transitions
- [ ] Stub implementations in handlers (validate, preview)
- [ ] Missing WebSocket events for experiment completion

#### Implementation Tasks:
```go
// Files to modify:
// - projects/phoenix-api/internal/api/experiments.go (add metrics endpoint)
// - projects/phoenix-api/internal/controller/experiment_controller.go
// - projects/phoenix-api/internal/api/pipelines.go (complete validation)
// - projects/phoenix-api/internal/websocket/hub.go (add missing events)
```

#### Validation Signals:
- Experiments progress through all phases correctly
- `GET /experiments/{id}` shows populated Results field when complete
- WebSocket broadcasts experiment_started, experiment_completed events
- Task queue shows proper state transitions

---

### 3. Phoenix Agent - Task Execution & Status Reporting

#### Issues to Fix:
- [ ] Missing rollback action handling for deployments
- [ ] Edge case handling for already running collectors
- [ ] Process cleanup verification needed

#### Implementation Tasks:
```go
// Files to modify:
// - projects/phoenix-agent/internal/supervisor/supervisor.go
// - projects/phoenix-agent/internal/supervisor/collector.go
// - projects/phoenix-agent/internal/supervisor/loadsim.go
```

#### Validation Signals:
- Agent logs show "Executing task... status=completed"
- No orphan processes after experiment stop
- Rollback tasks execute successfully
- Metrics pushed to API (HTTP 202 responses)

---

### 4. Pipeline Engine - Template Rendering & Validation

#### Issues to Fix:
- [ ] Incomplete template library
- [ ] Missing config validation implementation
- [ ] Variant tagging not verified
- [ ] Rollback logic needs definition

#### Implementation Tasks:
```yaml
# Files to create/modify:
# - configs/otel-templates/* (ensure all templates present)
# - projects/phoenix-api/internal/services/pipeline_template_renderer.go
# - projects/phoenix-api/internal/api/pipelines.go (validation endpoint)
```

#### Validation Signals:
- Templates render without errors
- Metrics in Prometheus have variant labels
- Config validation catches malformed YAML
- Rollback stops candidate pipeline cleanly

---

### 5. Load Simulation - Profile Standardization

#### Issues to Fix:
- [ ] Profile names need standardization
- [ ] Duration alignment with experiment
- [ ] Process cleanup verification
- [ ] Integration with experiment config

#### Implementation Tasks:
```bash
# Files to modify:
# - scripts/load-profiles/*.sh (standardize behavior)
# - projects/phoenix-agent/internal/supervisor/loadsim.go
# - projects/phoenix-api/internal/models/models.go (LoadProfile field)
```

#### Validation Signals:
- Load processes start/stop with experiment
- High-card profile generates expected cardinality
- No orphan processes remain
- CPU/memory metrics reflect load

---

### 6. Metrics Engine - KPI Computation

#### Issues to Fix:
- [ ] Placeholder cost calculations
- [ ] Missing accuracy checks
- [ ] Incomplete KPI calculations
- [ ] UI endpoints returning stub data

#### Implementation Tasks:
```go
// Files to modify:
// - projects/phoenix-api/internal/services/cost_service.go
// - projects/phoenix-api/internal/analyzer/kpi_calculator.go
// - projects/phoenix-api/internal/services/analysis_service.go
// - projects/phoenix-api/internal/api/ui_handlers.go
```

#### Validation Signals:
- KPI results match manual calculations
- Cost estimates use real metrics data
- No TODO warnings in logs
- Experiment Results populated correctly

---

### 7. WebSocket & Real-Time Feedback

#### Missing Events:
- [ ] experiment_started
- [ ] experiment_completed
- [ ] experiment_analyzed
- [ ] deployment_started
- [ ] kpi_computed

#### Implementation Tasks:
```go
// Broadcast points to add:
// - ExperimentController.StartExperiment()
// - ExperimentController.completeExperiment()
// - AnalysisService.AnalyzeExperiment()
// - PipelineDeploymentService methods
```

#### Validation Signals:
- WebSocket client receives all lifecycle events
- Events contain sufficient data
- Events arrive in correct order
- Near-instantaneous broadcast

---

### 8. End-to-End Test Coverage

#### Test Scenarios to Implement:
- [ ] Happy path: Create → Start → Complete → Results
- [ ] Stop mid-flight with cleanup verification
- [ ] Pipeline deploy and rollback
- [ ] Invalid input handling
- [ ] Multi-agent simulation (optional)

#### Implementation Tasks:
```bash
# Files to create/modify:
# - tests/e2e/experiment_complete_test.go
# - tests/e2e/pipeline_lifecycle_test.go
# - scripts/run-mvp-validation.sh
# - Makefile (add test-mvp target)
```

#### Validation Signals:
- All E2E tests pass consistently
- Reasonable completion times
- No errors in component logs
- Clean resource state after tests

---

## Execution Order

### Phase 1: Core Fixes (Week 1)
1. CLI endpoint corrections
2. API missing handlers
3. Agent rollback support
4. Basic WebSocket events

### Phase 2: Integration (Week 2)
5. Pipeline templates & validation
6. Metrics engine real data
7. Load simulation standardization
8. Cost calculation accuracy

### Phase 3: Polish & Testing (Week 3)
9. Complete WebSocket coverage
10. E2E test implementation
11. Performance validation
12. Documentation updates

---

## Validation Checklist

### Component Integration
- [ ] CLI → API: All commands work
- [ ] API → Agent: Tasks dispatch correctly
- [ ] Agent → API: Status updates flow
- [ ] Metrics → Analysis: KPIs computed
- [ ] WebSocket: Events broadcast

### Data Flow
- [ ] Experiments persist correctly
- [ ] Metrics reach Prometheus
- [ ] KPIs stored in database
- [ ] Results retrievable via CLI

### Resource Management
- [ ] Processes start/stop cleanly
- [ ] No orphan collectors
- [ ] Memory usage stable
- [ ] Database connections pooled

### Error Handling
- [ ] Invalid inputs rejected gracefully
- [ ] Failed tasks marked appropriately
- [ ] Rollback works on errors
- [ ] Clear error messages

---

## Success Criteria

The MVP is ready when:
1. Complete experiment workflow executes without manual intervention
2. All CLI commands in the MVP checklist work
3. Metrics show expected cardinality reduction (50-90%)
4. Cost savings calculations are accurate
5. WebSocket provides real-time visibility
6. E2E tests pass consistently
7. No critical TODOs remain in code
8. Documentation reflects current state

---

## Notes

- Single-VM architecture simplifies agent discovery
- Process-level metrics reduce complexity vs container metrics  
- Focus on core flow reliability over advanced features
- Pilot readiness is priority over optimization