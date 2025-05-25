# Phoenix Platform Migration Plan v2.0

## Current Status
- ✅ Phase 0: Foundation Setup - COMPLETE
- ✅ Phase 1: Shared Packages - COMPLETE  
- ✅ Phase 2: Core Services - COMPLETE (api, controller, generator, dashboard)
- 🔄 Phase 3-7: Remaining work

## Revised Migration Strategy

### Parallel Execution Groups
To enable multiple agents to work simultaneously, we're organizing remaining work into independent parallel groups:

### Group A: Go Services (Can run in parallel)
```bash
# Agent 1
- anomaly-detector (apps/anomaly-detector)
- control-actuator-go (apps/control-actuator-go)

# Agent 2  
- analytics (services/analytics)
- benchmark (services/benchmark)
- validator (services/validator)

# Agent 3
- synthetic-generator (services/generators/synthetic)
```

### Group B: Node/Script Services (Can run in parallel)
```bash
# Agent 4
- collector (services/collector)
- control-plane/observer (services/control-plane/observer)

# Agent 5
- control-plane/actuator (services/control-plane/actuator)
- generators/complex (services/generators/complex)
```

### Group C: Operators (Sequential - depends on Groups A&B)
```bash
# Any available agent
- loadsim-operator (phoenix-platform/operators/loadsim)
- pipeline-operator (phoenix-platform/operators/pipeline)
```

### Group D: Configurations (Can run in parallel)
```bash
# Agent 6
- monitoring configs (configs/monitoring)
- prometheus rules
- grafana dashboards

# Agent 7
- otel configs (configs/otel)
- control configs (configs/control)
- production configs (configs/production)
```

### Group E: Infrastructure (Sequential - depends on all above)
```bash
# Any available agent
- Docker Compose files
- Kubernetes manifests
- Helm charts
- CI/CD pipelines
```

## Execution Commands

### For Service Migration
```bash
# Generic command for any agent
export AGENT_ID="agent-$(hostname)-$$"
./scripts/migrate-service-corrected.sh <service-name> <old-path> <type>

# Examples:
./scripts/migrate-service-corrected.sh anomaly-detector apps/anomaly-detector go
./scripts/migrate-service-corrected.sh collector services/collector node
```

### For Config Migration
```bash
# Create a config migration script
./scripts/migrate-configs.sh <config-type>

# Examples:
./scripts/migrate-configs.sh monitoring
./scripts/migrate-configs.sh otel
```

## Validation Checkpoints

### After Each Service Migration
1. Check go.mod/package.json exists
2. Verify main entry point exists
3. Update import paths
4. Run basic build test
5. Create/update Makefile
6. Add to workspace configuration

### After Each Group Completion
1. Run integration tests for the group
2. Verify no broken dependencies
3. Update documentation
4. Commit changes

### Final Validation (Phase 6)
1. Full system build
2. Integration tests
3. E2E tests
4. Docker compose up
5. Kubernetes deployment test

## Quick Progress Tracker

```
Phase 3: Supporting Services
├── [ ] anomaly-detector
├── [ ] control-actuator-go  
├── [ ] analytics
├── [ ] benchmark
├── [ ] collector
├── [ ] validator
├── [ ] control-plane/actuator
├── [ ] control-plane/observer
├── [ ] generators/complex
└── [ ] generators/synthetic

Phase 4: Operators & Tools
├── [ ] loadsim-operator
├── [ ] pipeline-operator
└── [ ] phoenix-cli (already in phoenix-platform/cmd)

Phase 5: Infrastructure & Config
├── [ ] monitoring configs
├── [ ] otel configs
├── [ ] control configs
├── [ ] production configs
├── [ ] docker-compose files
├── [ ] kubernetes manifests
└── [ ] helm charts

Phase 6: Integration Testing
├── [ ] Update all imports
├── [ ] Build all services
├── [ ] Run unit tests
├── [ ] Run integration tests
└── [ ] Run E2E tests

Phase 7: Finalization
├── [ ] Archive OLD_IMPLEMENTATION
├── [ ] Update documentation
├── [ ] Create migration report
└── [ ] Tag release
```

## Time Estimate
- With 7 parallel agents: ~2-3 hours
- With 3 parallel agents: ~4-5 hours  
- With 1 agent (sequential): ~8-10 hours

## Next Steps
1. Assign agents to groups
2. Start parallel execution
3. Monitor progress via lock files
4. Coordinate at checkpoints