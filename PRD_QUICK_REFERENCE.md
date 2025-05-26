# Phoenix Platform - PRD Gap Quick Reference

## 🚨 Critical Missing Components

### 1. Load Simulation System (0% Complete)
```bash
# Missing CLI commands:
phoenix loadsim start --profile realistic --target-host node-01
phoenix loadsim stop --sim-job-name sim-123
phoenix loadsim status
phoenix loadsim list-profiles

# Missing implementation:
/projects/loadsim-operator/controllers/  # Empty - needs controller
/projects/loadsim-operator/internal/generator/  # Missing load generator
```

### 2. Pipeline Management CLI (6/8 Missing)
```bash
# Missing commands:
phoenix pipeline show process-baseline-v1
phoenix pipeline validate config.yaml
phoenix pipeline status --target-host node-01
phoenix pipeline get-active-config --crd-name my-pipeline
phoenix pipeline rollback my-pipeline --to-version v1
phoenix pipeline delete my-pipeline
```

### 3. Web Console Views (2 Missing)
```typescript
// Missing pages:
/projects/dashboard/src/pages/DeployedPipelines.tsx  # Host-pipeline mapping view
/projects/dashboard/src/pages/PipelineCatalog.tsx    # Read-only template browser
```

### 4. OTel Pipeline Configs (2/5 Missing)
```yaml
# Missing configs:
/configs/pipelines/catalog/process/process-topk-v1.yaml         # Top K by CPU/Memory
/configs/pipelines/catalog/process/process-adaptive-filter-v1.yaml  # Adaptive filtering
```

## ✅ Quick Wins (< 1 day each)

1. **Add Missing OTel Configs**
   ```yaml
   # process-topk-v1.yaml template:
   processors:
     groupbyattrs:
       keys: [process.executable.name]
     # Add top-k filtering logic
   ```

2. **Complete Pipeline Deployer Service**
   ```go
   // /projects/platform-api/internal/services/pipeline_deployment_service.go
   // Remove TODO comments and implement CRUD operations
   ```

3. **Add Experiment Delete Command**
   ```go
   // /projects/phoenix-cli/cmd/experiment_delete.go
   func NewDeleteCommand() *cobra.Command {
       // Implement delete with confirmation
   }
   ```

## 📋 Component Ownership Map

| Component | Owner Team | Priority | Effort |
|-----------|------------|----------|---------|
| LoadSim Operator | Control Plane | HIGH | 1 week |
| Load Generator | DevTools | HIGH | 1 week |
| Pipeline CLI Commands | CLI Team | HIGH | 1 week |
| Web Console Views | UI/UX Team | MEDIUM | 1 week |
| OTel Configs | Obs Guild | MEDIUM | 2 days |
| Pipeline Deployer | Control Plane | HIGH | 2 days |

## 🔧 Implementation Order

### Week 1: Foundation
1. Complete Pipeline Deployer Service
2. Create missing OTel configs
3. Start LoadSim Operator implementation

### Week 2: Load Simulation
1. Implement LoadSim controller
2. Create load generator
3. Add CLI commands

### Week 3: Pipeline Management
1. Add all missing pipeline CLI commands
2. Implement validation service
3. Add status aggregation

### Week 4: UI Completion
1. Create Deployed Pipelines view
2. Add Pipeline Catalog browser
3. Enhance experiment monitoring

### Week 5: Integration
1. Add missing experiment features
2. Implement acceptance tests
3. Fix integration issues

### Week 6: Polish
1. Documentation
2. Error message improvements
3. Performance optimization

## 🎯 Definition of Done

### For Each Component:
- [ ] Unit tests written (>80% coverage)
- [ ] Integration tests pass
- [ ] Documentation updated
- [ ] Error handling implemented
- [ ] Logging added
- [ ] Code reviewed

### For MVP Release:
- [ ] All 13 PRD acceptance tests pass
- [ ] All CLI commands documented
- [ ] UI views responsive and functional
- [ ] Load patterns generate correctly
- [ ] < 5% collector overhead verified
- [ ] End-to-end demo successful

## 📊 Progress Tracking

```bash
# Check implementation status:
make check-prd-compliance

# Run acceptance tests:
make test-acceptance

# Validate OTel configs:
make validate-pipelines
```

## 🚀 Getting Started

1. **Set up development environment:**
   ```bash
   ./scripts/setup-dev-env.sh
   make dev-up
   ```

2. **Pick a component from the ownership map**

3. **Follow the implementation pattern:**
   - Check PRD requirements
   - Write tests first
   - Implement incrementally
   - Validate against acceptance criteria

4. **Test your changes:**
   ```bash
   make test-unit
   make test-integration
   ```

---

**Remember**: The PRD is the source of truth. When in doubt, refer to `Phoenix Observability Platform – Process-Metrics MVP` PRD v2.0.