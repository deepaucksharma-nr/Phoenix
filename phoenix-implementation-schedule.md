# Phoenix Lean-Core Implementation Schedule & Tasks

## Sprint A: Control Plane Consolidation (Weeks 1-2)

### Week 1: API Consolidation & Database Setup

#### Day 1-2: Project Setup & Database Migration
**Owner: Backend Lead**

- [ ] Create `phoenix-api` project structure
  ```bash
  mkdir -p projects/phoenix-api/{cmd,internal/{api,controller,analyzer,store,tasks}}
  cd projects/phoenix-api && go mod init
  ```

- [ ] Set up database migrations framework
  ```go
  // internal/store/migrations/001_lean_core_tables.sql
  ```

- [ ] Create migration for new tables:
  - `agent_tasks`
  - `agent_status` 
  - `active_pipelines`
  - `metrics_cache`

- [ ] Add compatibility views for existing code

**Deliverable:** Empty phoenix-api structure with migrated database schema

#### Day 3-4: Merge Platform API & Controller Core
**Owner: Backend Team**

- [ ] Copy platform-api as base
  ```bash
  cp -r projects/platform-api/* projects/phoenix-api/
  ```

- [ ] Port controller state machine
  - [ ] Copy `controller/internal/controller/*.go`
  - [ ] Remove K8s client dependencies
  - [ ] Replace CRD operations with task queue calls

- [ ] Create task queue implementation
  ```go
  // internal/tasks/queue.go
  type Queue interface {
      Enqueue(ctx context.Context, task *Task) error
      GetPending(ctx context.Context, hostID string) ([]*Task, error)
      UpdateStatus(ctx context.Context, taskID string, status TaskStatus) error
  }
  ```

**Deliverable:** Merged API with working experiment creation (no deployment yet)

#### Day 5: Agent Endpoints & Task System
**Owner: Backend Developer 1**

- [ ] Implement agent endpoints:
  - [ ] `GET /agent/v1/tasks` (long-polling)
  - [ ] `POST /agent/v1/status`
  - [ ] `POST /agent/v1/heartbeat`
  - [ ] `POST /agent/v1/metrics`

- [ ] Add WebSocket broadcasting for agent events

- [ ] Create task scheduler logic:
  ```go
  // When experiment starts, create tasks for each host
  func (s *Scheduler) ScheduleExperiment(exp *Experiment) error
  ```

**Deliverable:** Working agent API endpoints with tests

### Week 2: Analysis Integration & Testing

#### Day 6-7: Port Benchmark & Analytics
**Owner: Backend Developer 2**

- [ ] Create analyzer package
  ```go
  // internal/analyzer/kpi_calculator.go
  // internal/analyzer/cost_analyzer.go
  ```

- [ ] Port PromQL queries from benchmark service
- [ ] Implement KPI calculation endpoints:
  - [ ] `POST /api/v1/experiments/:id/calculate-kpis`
  - [ ] `GET /api/v1/experiments/:id/kpis`

- [ ] Add Pushgateway client for metrics queries

**Deliverable:** Working KPI calculation without needing separate services

#### Day 8-9: Integration Testing
**Owner: QA + Backend Team**

- [ ] Create test environment setup
  ```go
  // tests/integration/setup_test.go
  ```

- [ ] Write integration tests:
  - [ ] Experiment lifecycle
  - [ ] Task queue operations
  - [ ] Agent polling simulation
  - [ ] KPI calculations

- [ ] Create mock agent for testing

**Deliverable:** Passing integration test suite

#### Day 10: Documentation & API Cleanup
**Owner: Whole Team**

- [ ] Update API documentation
- [ ] Remove dead code from merge
- [ ] Create migration guide for operators
- [ ] Performance testing of consolidated API

**Deliverable:** Clean, documented phoenix-api ready for agent integration

## Sprint B: Agent Implementation (Weeks 3-4)

### Week 3: Core Agent Development

#### Day 11-12: Agent Foundation
**Owner: Backend Developer 3**

- [ ] Create agent project structure
  ```bash
  mkdir -p cmd/phoenix-agent/{main.go,internal/{poller,supervisor,metrics}}
  ```

- [ ] Implement API polling client:
  - [ ] HTTP client with retry/backoff
  - [ ] Task deserialization
  - [ ] Status reporting

- [ ] Basic agent loop:
  ```go
  for {
      tasks := poller.GetTasks()
      for _, task := range tasks {
          go supervisor.Execute(task)
      }
      time.Sleep(pollInterval)
  }
  ```

**Deliverable:** Agent that polls and logs tasks

#### Day 13-14: Process Supervisor
**Owner: Backend Developer 1**

- [ ] Implement collector supervisor:
  - [ ] Download OTel configs from URLs
  - [ ] Variable substitution in configs
  - [ ] Process spawning with proper isolation
  - [ ] Process monitoring and restart

- [ ] Add systemd/Docker support:
  ```go
  // Different execution strategies based on environment
  type Executor interface {
      Start(name string, cmd []string, env []string) (*Process, error)
      Stop(pid int) error
  }
  ```

**Deliverable:** Agent can spawn and manage OTel collectors

#### Day 15: Load Simulation Integration
**Owner: Backend Developer 2**

- [ ] Port load simulation profiles:
  - [ ] High cardinality generator
  - [ ] Normal load (stress-ng wrapper)
  - [ ] Custom workload scripts

- [ ] Implement load simulator supervisor:
  ```go
  // internal/supervisor/loadsim.go
  type LoadSimulator interface {
      StartProfile(profile string, duration time.Duration) error
      Stop() error
  }
  ```

**Deliverable:** Agent can run load simulations

### Week 4: Metrics & Production Readiness

#### Day 16-17: Metrics Pipeline
**Owner: Backend Developer 3**

- [ ] Implement Pushgateway client in agent
- [ ] Add self-metrics collection:
  - [ ] Active tasks
  - [ ] Resource usage
  - [ ] Process health

- [ ] Create OTel config templates:
  - [ ] Convert Top-K to transform+Lua
  - [ ] Convert adaptive filter to expressions
  - [ ] Add Pushgateway remote_write

**Deliverable:** Metrics flow from collectors → Pushgateway → Prometheus

#### Day 18: Kubernetes Integration
**Owner: DevOps + Backend Developer 1**

- [ ] Create agent DaemonSet manifest
- [ ] Add RBAC for agent (minimal)
- [ ] Test on multi-node cluster
- [ ] Add node affinity options

**Deliverable:** Agent runs successfully as DaemonSet

#### Day 19: VM Support
**Owner: Backend Developer 2**

- [ ] Create systemd unit file
- [ ] Add install script
- [ ] Test on Ubuntu/RHEL VMs
- [ ] Document VM deployment

**Deliverable:** Agent runs on VMs via systemd

#### Day 20: End-to-End Testing
**Owner: Whole Team**

- [ ] Full experiment flow test:
  1. Create experiment via API
  2. Agents pick up tasks
  3. Collectors start
  4. Metrics flow to Pushgateway
  5. KPIs calculate correctly
  6. Promotion works

- [ ] Performance benchmarks
- [ ] Chaos testing (agent disconnects, etc.)

**Deliverable:** Validated E2E flow

## Week 5: Production Deployment

### Day 21-22: Staging Deployment
**Owner: DevOps Team**

- [ ] Deploy to staging cluster:
  - [ ] Phoenix API (2 replicas)
  - [ ] PostgreSQL (with backups)
  - [ ] Prometheus + Pushgateway
  - [ ] Agent DaemonSet

- [ ] Run parallel experiments:
  - [ ] One via old system
  - [ ] One via new system
  - [ ] Compare results

**Deliverable:** Working staging environment

### Day 23: Production Prep
**Owner: DevOps + Backend Lead**

- [ ] Create rollback plan
- [ ] Set up monitoring/alerts:
  - [ ] API health
  - [ ] Agent connectivity
  - [ ] Task queue depth
  - [ ] Pushgateway availability

- [ ] Load test production config

**Deliverable:** Production-ready deployment plan

### Day 24-25: Gradual Rollout
**Owner: Whole Team**

- [ ] Enable feature flag for subset of users
- [ ] Monitor metrics closely
- [ ] Gather feedback
- [ ] Fix any issues

**Deliverable:** Lean architecture in production

## Task Breakdown by Developer

### Backend Developer 1 (Senior)
**Total: ~40 hours**
- Controller migration (8h)
- Agent endpoints (8h) 
- Process supervisor (8h)
- Kubernetes integration (8h)
- Code reviews (8h)

### Backend Developer 2
**Total: ~40 hours**
- KPI/Analytics port (8h)
- Load sim integration (8h)
- VM support (8h)
- Config templates (8h)
- Testing support (8h)

### Backend Developer 3
**Total: ~32 hours**
- Agent foundation (8h)
- Polling client (8h)
- Metrics pipeline (8h)
- Documentation (8h)

### DevOps Engineer
**Total: ~24 hours**
- Database migrations (4h)
- Kubernetes manifests (8h)
- Staging deployment (8h)
- Production rollout (4h)

### QA Engineer
**Total: ~24 hours**
- Integration test framework (8h)
- E2E test scenarios (8h)
- Performance testing (8h)

## Risk Mitigation Schedule

| Day | Risk Check | Mitigation Action |
|-----|------------|------------------|
| 5 | API performance with consolidated services | Add caching layer if needed |
| 10 | Database migration issues | Have rollback scripts ready |
| 15 | Agent reliability | Add circuit breakers |
| 18 | Pushgateway scalability | Test with sharding |
| 22 | Feature parity | Document any gaps |
| 25 | Production issues | 24/7 on-call schedule |

## Success Metrics Timeline

| Milestone | Target | Measurement |
|-----------|--------|-------------|
| Week 1 End | API responds < 100ms | Load test results |
| Week 2 End | All tests passing | CI/CD dashboard |
| Week 3 End | Agent handles 100 tasks/min | Stress test |
| Week 4 End | E2E < 5 min | Timer logs |
| Week 5 End | Zero downtime deploy | Monitoring |

## Dependencies & Prerequisites

### Required Before Start:
1. PostgreSQL 14+ deployed
2. Prometheus + Pushgateway deployed  
3. S3/GCS bucket for config storage
4. CI/CD pipelines updated
5. Team trained on new architecture

### External Dependencies:
- OTel Collector Contrib 0.95+ released
- Prometheus remote_write stable
- Go 1.21+ on build machines

## Contingency Plans

### If Behind Schedule:
1. **Day 10**: Skip analytics port, add in Sprint C
2. **Day 15**: Use basic shell scripts for load sim
3. **Day 20**: Reduce E2E test scenarios
4. **Day 25**: Staged rollout over 2 weeks

### If Blocking Issues:
1. **Database too slow**: Add Redis cache
2. **Agent unreliable**: Add local queue
3. **Pushgateway bottleneck**: Direct Prometheus scrape
4. **Config complexity**: GUI config builder

## Post-Implementation Tasks (Sprint C)

1. **Observability Enhancement** (1 week)
   - Distributed tracing
   - Enhanced dashboards
   - SLO monitoring

2. **Advanced Features** (1 week)
   - Multi-region support
   - Config versioning
   - A/B/n testing

3. **Developer Experience** (1 week)
   - CLI improvements
   - Config validators
   - Testing helpers

4. **Documentation** (3 days)
   - Architecture guide
   - Troubleshooting guide
   - Video tutorials