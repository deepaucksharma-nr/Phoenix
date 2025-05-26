# Phoenix Platform - PRD Implementation Plan

## Overview

This implementation plan addresses the gaps identified in the PRD Alignment Report to achieve 100% compliance with the Process-Metrics MVP requirements. The plan is organized into sprints with clear deliverables and acceptance criteria.

## Sprint 0: Foundation & Critical Infrastructure (1 week)

### Goals
- Set up missing infrastructure components
- Complete critical service implementations
- Prepare for operator development

### Tasks

#### 1. Complete Pipeline Deployer Service
**Owner**: Control Plane Team  
**Location**: `/services/api/internal/services/pipeline_deployment_service.go`

```go
// Required implementations:
- CreateDeployment(ctx, request) (*Deployment, error)
- UpdateDeployment(ctx, id, request) (*Deployment, error)
- GetDeploymentStatus(ctx, id) (*DeploymentStatus, error)
- RollbackDeployment(ctx, id, version) error
- DeleteDeployment(ctx, id) error
```

#### 2. Set Up Load Generator Base
**Owner**: DevTools Team  
**Location**: `/pkg/loadgen/`

```go
// Create load generation framework:
- Process spawner interface
- CPU/Memory load patterns
- Process lifecycle management
- Metrics collection
```

#### 3. Create Missing OTel Pipeline Configs
**Owner**: Observability Guild  
**Location**: `/configs/pipelines/catalog/process/`

```yaml
# process-topk-v1.yaml
receivers:
  hostmetrics:
    scrapers:
      process:
        include:
          match_type: regexp
          names: [".*"]

processors:
  # Implement top-k logic
  groupbyattrs:
    keys: [process.executable.name, process.pid]
  
  # Sort and filter top K by CPU/Memory
  filter:
    metrics:
      # Keep only top K processes

# process-adaptive-filter-v1.yaml  
processors:
  # Adaptive threshold logic
  filter:
    metrics:
      # Dynamic filtering based on system load
```

### Acceptance Criteria
- [ ] Pipeline Deployer Service passes unit tests
- [ ] Load generator framework compiles and has basic tests
- [ ] Both missing OTel configs validated with `otelcol validate`

---

## Sprint 1: Load Simulation Implementation (2 weeks)

### Goals
- Implement LoadSim Operator
- Add CLI load simulation commands
- Create load generator profiles

### Tasks

#### 1. Implement LoadSim Operator Controller
**Owner**: Control Plane Team  
**Location**: `/projects/loadsim-operator/`

```go
// controllers/loadsimulationjob_controller.go
type LoadSimulationJobReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *LoadSimulationJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch LoadSimulationJob CR
    // 2. Create/Update simulator Job
    // 3. Monitor Job status
    // 4. Update CR status
    // 5. Handle cleanup on completion
}
```

#### 2. Implement Load Generator
**Owner**: DevTools Team  
**Location**: `/projects/loadsim-operator/internal/generator/`

```go
// Profiles implementation:
- RealisticProfile: Mix of long-running and short-lived processes
- HighCardinalityProfile: Many unique process names
- ProcessChurnProfile: Rapid process creation/destruction
- CustomProfile: User-defined patterns
```

#### 3. Add CLI Load Simulation Commands
**Owner**: CLI Team  
**Location**: `/projects/phoenix-cli/cmd/`

```go
// loadsim.go
- loadsimCmd (parent command)
- loadsimStartCmd
- loadsimStopCmd  
- loadsimStatusCmd
- loadsimListProfilesCmd
```

### Acceptance Criteria
- [ ] LoadSim operator successfully deploys simulator pods
- [ ] All three load profiles generate expected process patterns
- [ ] CLI commands can start/stop/monitor load simulations
- [ ] Integration test: Full load simulation workflow

---

## Sprint 2: Pipeline Management Enhancement (2 weeks)

### Goals
- Complete pipeline management CLI commands
- Enhance pipeline lifecycle management
- Add validation capabilities

### Tasks

#### 1. Implement Missing Pipeline CLI Commands
**Owner**: CLI Team  
**Location**: `/projects/phoenix-cli/cmd/`

```go
// pipeline_show.go
func showPipelineCmd() *cobra.Command {
    // Display catalog pipeline YAML
}

// pipeline_validate.go
func validatePipelineCmd() *cobra.Command {
    // Validate local OTel config
}

// pipeline_status.go
func statusPipelineCmd() *cobra.Command {
    // Show deployment status with metrics
}

// pipeline_get_config.go
func getActiveConfigCmd() *cobra.Command {
    // Retrieve running configuration
}

// pipeline_rollback.go
func rollbackPipelineCmd() *cobra.Command {
    // Revert to previous version
}

// pipeline_delete.go
func deletePipelineCmd() *cobra.Command {
    // Remove deployed pipeline
}
```

#### 2. Add Pipeline Validation Service
**Owner**: Control Plane Team  
**Location**: `/pkg/validation/pipeline/`

```go
// Validation logic:
- OTel config syntax validation
- Required receivers check (hostmetrics)
- Phoenix processor validation
- Security checks (no hardcoded secrets)
```

#### 3. Implement Pipeline Status Aggregation
**Owner**: Control Plane Team  
**Location**: `/services/api/internal/services/`

```go
// Aggregate status from:
- PhoenixProcessPipeline CR status
- Prometheus metrics (input/output rates)
- Collector health metrics
- Critical process retention
```

### Acceptance Criteria
- [ ] All 6 missing pipeline commands functional
- [ ] Pipeline validation catches common errors
- [ ] Status command shows real-time metrics
- [ ] Rollback successfully reverts configurations

---

## Sprint 3: Web Console Completion (2 weeks)

### Goals
- Add missing UI views
- Enhance monitoring capabilities
- Improve user experience

### Tasks

#### 1. Implement Deployed Pipelines View
**Owner**: UI/UX Team  
**Location**: `/projects/dashboard/src/pages/DeployedPipelines.tsx`

```typescript
// Components needed:
- PipelineTable: Host, Pipeline, Status, Metrics
- PipelineFilters: By host, status, pipeline type
- MetricsSummary: Cardinality reduction, cost savings
- QuickActions: Start experiment, view details
```

#### 2. Create Pipeline Catalog Browser
**Owner**: UI/UX Team  
**Location**: `/projects/dashboard/src/pages/PipelineCatalog.tsx`

```typescript
// Read-only catalog view:
- CatalogList: Available pipeline templates
- PipelineDetails: Description, strategy, parameters
- YAMLViewer: Syntax-highlighted OTel config
- DeployButton: Link to CLI command or deployment flow
```

#### 3. Enhance Experiment Dashboard
**Owner**: UI/UX Team  
**Location**: `/projects/dashboard/src/components/ExperimentMonitor/`

```typescript
// Add missing features:
- Watch mode with auto-refresh
- Export options (JSON, HTML report)
- Detailed error states from operators
- Load simulation status indicator
```

### Acceptance Criteria
- [ ] Deployed Pipelines view shows all active pipelines
- [ ] Catalog browser displays all 5 pipeline templates
- [ ] Experiment dashboard has watch mode
- [ ] UI correctly displays operator error states

---

## Sprint 4: Experiment Enhancement (1 week)

### Goals
- Complete experiment management features
- Add missing CLI capabilities
- Improve error handling

### Tasks

#### 1. Add Experiment Delete Command
**Owner**: CLI Team  
**Location**: `/projects/phoenix-cli/cmd/experiment_delete.go`

```go
func deleteExperimentCmd() *cobra.Command {
    // Delete experiment and cleanup resources
    // Confirm before deletion
    // Handle cascading deletes
}
```

#### 2. Implement Watch Mode for Status
**Owner**: CLI Team  
**Location**: `/projects/phoenix-cli/cmd/experiment_status.go`

```go
// Add --watch flag:
- Poll status every 5 seconds
- Display progress indicators
- Show real-time metric updates
- Handle Ctrl+C gracefully
```

#### 3. Add Output Formats for Compare
**Owner**: CLI Team  
**Location**: `/projects/phoenix-cli/cmd/experiment_compare.go`

```go
// Support output formats:
- Table (default): ASCII table
- JSON: Machine-readable format
- HTML: Generate report with charts
```

### Acceptance Criteria
- [ ] Experiment delete removes all resources
- [ ] Watch mode updates in real-time
- [ ] HTML reports include visualization
- [ ] All formats properly escape data

---

## Sprint 5: Integration & Testing (1 week)

### Goals
- Ensure all components work together
- Achieve acceptance test coverage
- Fix integration issues

### Tasks

#### 1. Implement Acceptance Test Suite
**Owner**: QE Team  
**Location**: `/tests/acceptance/`

```go
// Implement all PRD acceptance tests (AT-P01 to AT-P13):
- Pipeline deployment tests
- A/B experiment tests
- Load simulation tests
- Error handling tests
- Rollback tests
```

#### 2. Create E2E Test Automation
**Owner**: QE Team  
**Location**: `/.github/workflows/e2e-tests.yml`

```yaml
# GitHub Actions workflow:
- Set up KIND cluster
- Deploy Phoenix platform
- Run acceptance test suite
- Generate coverage reports
```

#### 3. Fix Integration Issues
**Owner**: All Teams  
- Address issues found during testing
- Improve error propagation
- Enhance logging and debugging

### Acceptance Criteria
- [ ] All 13 acceptance tests pass
- [ ] E2E tests run in CI/CD
- [ ] No critical integration issues
- [ ] Error messages are clear and actionable

---

## Sprint 6: Documentation & Polish (1 week)

### Goals
- Complete user documentation
- Polish user experience
- Prepare for release

### Tasks

#### 1. Create User Documentation
**Owner**: Documentation Team  
**Location**: `/docs/user-guide/`

- Getting Started Guide
- Pipeline Catalog Reference
- Experiment Tutorial
- Troubleshooting Guide
- CLI Command Reference

#### 2. Improve Error Messages
**Owner**: All Teams  
- Review all error messages
- Add helpful context
- Include resolution steps
- Test with users

#### 3. Performance Optimization
**Owner**: Platform Team  
- Optimize Prometheus queries
- Reduce API response times
- Improve UI loading times
- Cache where appropriate

### Acceptance Criteria
- [ ] Documentation covers all features
- [ ] Error messages are helpful
- [ ] Performance meets NFRs
- [ ] Ready for GA release

---

## Risk Mitigation

### Technical Risks
1. **OTel Version Compatibility**
   - Mitigation: Pin versions, extensive testing
   
2. **Operator Resource Usage**
   - Mitigation: Set conservative limits, monitor closely

3. **Prometheus Query Performance**
   - Mitigation: Use recording rules, optimize queries

### Schedule Risks
1. **Integration Complexity**
   - Mitigation: Daily integration tests, early detection

2. **Dependency on External Teams**
   - Mitigation: Clear interfaces, parallel development

## Success Metrics

- 100% PRD acceptance tests passing
- All CLI commands implemented and documented
- UI views complete with real-time updates
- Load simulation generating accurate patterns
- < 5% overhead for Phoenix collectors
- < 2 second API response times (p95)

## Conclusion

This implementation plan provides a structured approach to achieving full PRD compliance within 6-7 weeks. The sprints are organized to deliver value incrementally while building toward the complete Process-Metrics MVP.

---

*Plan Created: May 2025*  
*Target Completion: 6-7 weeks*  
*Teams Required: Control Plane, CLI, UI/UX, DevTools, QE, Documentation*