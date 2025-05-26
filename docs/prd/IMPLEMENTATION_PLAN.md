# Phoenix Platform - PRD Implementation Plan

## Overview

This plan provides a complete roadmap to achieve 100% PRD compliance from the current 65% baseline within 6-7 weeks.

## Team Structure

### Required Resources (4-5 Engineers)
- **2 Backend Engineers** - Operators, services (Go, Kubernetes)
- **1 CLI Engineer** - Command implementation (Go, Cobra)
- **1 Frontend Engineer** - Web Console views (React, TypeScript)
- **1 DevOps Engineer** (part-time) - OTel configs, deployment

## Implementation Roadmap

### Week 0: Foundation (Current Week)

#### Quick Wins (1-2 days each)
1. **Complete Pipeline Deployer Service**
   ```go
   // Location: /projects/platform-api/internal/services/pipeline_deployment_service.go
   - Implement CreateDeployment()
   - Implement UpdateDeployment()
   - Implement GetDeploymentStatus()
   - Implement RollbackDeployment()
   - Implement DeleteDeployment()
   ```

2. **Create Missing OTel Configs**
   ```bash
   make generate-topk-pipeline
   make generate-adaptive-pipeline
   ```

3. **Add Experiment Delete Command**
   ```go
   // Location: /projects/phoenix-cli/cmd/experiment_delete.go
   func NewDeleteCommand() *cobra.Command {
       // Simple CRUD operation
   }
   ```

### Weeks 1-2: Load Simulation System

#### Week 1: Operator Foundation
- **LoadSim Operator Controller**
  - Reconciliation loop
  - Job management
  - Status tracking
  - Resource cleanup

- **Load Generator Framework**
  - Process spawner interface
  - Profile structure
  - Metrics collection

#### Week 2: Complete Implementation
- **Load Profiles**
  - Realistic (mixed workload)
  - High-cardinality (unique names)
  - Process-churn (rapid lifecycle)
  - Custom (user-defined)

- **CLI Integration**
  ```bash
  phoenix loadsim start --profile realistic --target-host node-01
  phoenix loadsim stop --sim-job-name sim-123
  phoenix loadsim status
  phoenix loadsim list-profiles
  ```

### Weeks 3-4: Feature Completion

#### Week 3: CLI Commands
Implement 6 missing pipeline commands:

```bash
# Pipeline management
phoenix pipeline show process-baseline-v1          # Display YAML
phoenix pipeline validate config.yaml              # Validate config
phoenix pipeline status --target-host node-01      # Show status
phoenix pipeline get-active-config --crd-name app  # Get running config
phoenix pipeline rollback app --to-version v1      # Rollback
phoenix pipeline delete app                        # Delete pipeline
```

Enhanced experiment features:
- Add `--watch` flag to status command
- Add output formats (table, json, html) to compare
- Complete experiment delete command

#### Week 4: Web Console

1. **Deployed Pipelines View**
   - Host-pipeline mapping table
   - Real-time metrics display
   - Cardinality reduction percentages
   - Cost savings estimates
   - Quick action buttons

2. **Pipeline Catalog View**
   - Template browser (read-only)
   - YAML configuration viewer
   - Parameter documentation
   - CLI command generation

### Weeks 5-6: Integration & Polish

#### Week 5: Testing
- Implement all 13 PRD acceptance tests (AT-P01 to AT-P13)
- Performance validation (< 5% overhead)
- End-to-end experiment workflow
- Load simulation verification

#### Week 6: Documentation & Release
- User documentation
- CLI command reference
- Deployment guides
- Performance optimization
- Bug fixes
- GA preparation

## Implementation Examples

### LoadSim Operator Controller
```go
// /projects/loadsim-operator/controllers/loadsimulationjob_controller.go
func (r *LoadSimulationJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch LoadSimulationJob CR
    var loadSimJob phoenixv1alpha1.LoadSimulationJob
    if err := r.Get(ctx, req.NamespacedName, &loadSimJob); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // 2. Create/Update simulator Job
    job := r.constructJobForLoadSim(&loadSimJob)
    if err := ctrl.SetControllerReference(&loadSimJob, job, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }
    
    // 3. Deploy and monitor
    // ... implementation details in PRD_IMPLEMENTATION_EXAMPLES.md
}
```

### Pipeline Show Command
```go
// /projects/phoenix-cli/cmd/pipeline_show.go
func NewPipelineShowCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "show <pipeline-name>",
        Short: "Display pipeline configuration",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            config, err := apiClient.GetPipelineConfig(ctx, args[0])
            if err != nil {
                return err
            }
            fmt.Println(config.YAML)
            return nil
        },
    }
}
```

### Deployed Pipelines View
```typescript
// /projects/dashboard/src/pages/DeployedPipelines.tsx
export const DeployedPipelines: React.FC = () => {
    const [pipelines, setPipelines] = useState<DeployedPipeline[]>([]);
    
    // Real-time updates via WebSocket
    useEffect(() => {
        const ws = new WebSocket('/api/v1/pipelines/stream');
        ws.onmessage = (event) => {
            setPipelines(JSON.parse(event.data));
        };
    }, []);
    
    return (
        <Table>
            <TableHead>
                <TableRow>
                    <TableCell>Host</TableCell>
                    <TableCell>Pipeline</TableCell>
                    <TableCell>Reduction %</TableCell>
                    <TableCell>Est. Savings</TableCell>
                    <TableCell>Actions</TableCell>
                </TableRow>
            </TableHead>
            {/* ... render pipelines */}
        </Table>
    );
};
```

## Tracking & Validation

### Weekly Milestones

**Week 2**: LoadSim operator deploys pods ✓  
**Week 4**: All CLI commands working ✓  
**Week 6**: All acceptance tests passing ✓

### Daily Checklist
- [ ] Update TRACKING_CHECKLIST.md
- [ ] Run relevant tests
- [ ] Check integration impacts
- [ ] Commit with PRD reference

### Success Metrics
- All 17 CLI commands implemented
- Both operators functional
- 4 Web Console views complete
- 5 OTel configs validated
- Load simulation operational
- 13 acceptance tests passing
- < 5% performance overhead
- Complete documentation

## Risk Mitigation

1. **Technical Risks**
   - Pin OTel versions
   - Conservative resource limits
   - Extensive testing

2. **Schedule Risks**
   - Daily integration tests
   - Parallel development
   - 1-week buffer

3. **Quality Risks**
   - Continuous testing
   - Code reviews
   - Acceptance criteria

## Commands Reference

```bash
# Check compliance
make -f Makefile.prd check-prd-compliance

# Generate stubs
make -f Makefile.prd create-missing-files

# Run tests
make test-acceptance

# Validate configs
make validate-pipelines
```

## Conclusion

This implementation plan provides a clear path from 65% to 100% PRD compliance within 6-7 weeks. The plan balances quick wins with complex implementations, ensures continuous validation, and maintains focus on delivering business value through the Process-Metrics MVP.