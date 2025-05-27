# Phoenix MVP Quick Reference

## 🚀 Key Commands

### Development
```bash
# Start all services
make dev-up

# Run MVP validation
./scripts/mvp-validation.sh

# Check fix checklist
./scripts/mvp-fix-checklist.sh

# Run E2E tests
go test ./tests/e2e -v -run TestMVPCompleteFlow
```

### Testing Individual Components
```bash
# Test CLI commands
phoenix experiment create --name test --baseline baseline --candidate topk
phoenix experiment start <id>
phoenix experiment metrics <id>
phoenix experiment stop <id>

# Test API directly
curl -X GET http://localhost:8080/api/v1/experiments
curl -X POST http://localhost:8080/api/v1/experiments/<id>/start

# Monitor WebSocket events
wscat -c ws://localhost:8080/api/v1/ws
```

## 🔧 Critical Fixes Needed

### 1. CLI (High Priority)
- [ ] Fix pipeline deployment endpoint path
- [ ] Fix experiment metrics endpoint
- [ ] Add experiment rollback command

**Files**: `projects/phoenix-cli/internal/client/api.go`, `cmd/experiment_metrics.go`

### 2. API (High Priority)
- [ ] Add unified `/experiments/{id}/metrics` endpoint
- [ ] Complete experiment lifecycle broadcasts
- [ ] Implement pipeline validation

**Files**: `projects/phoenix-api/internal/api/experiments.go`, `internal/controller/experiment_controller.go`

### 3. Agent (Medium Priority)
- [ ] Handle rollback action for deployments
- [ ] Verify process cleanup

**Files**: `projects/phoenix-agent/internal/supervisor/supervisor.go`

### 4. Metrics Engine (High Priority)
- [ ] Replace placeholder cost calculations
- [ ] Complete KPI computations

**Files**: `projects/phoenix-api/internal/services/cost_service.go`, `internal/analyzer/kpi_calculator.go`

## 📊 Success Metrics

### Functional
- All CLI commands work without errors
- Experiments complete with results
- WebSocket events broadcast correctly
- No orphan processes after stop/rollback

### Performance
- Experiment completion < 5 minutes
- 50-90% cardinality reduction achieved
- Cost calculations accurate within 5%

### Quality
- E2E tests pass consistently
- No critical TODOs in production paths
- Clear error messages for invalid inputs

## 🔍 Debugging Tips

### Check experiment status
```bash
# Via CLI
phoenix experiment status <id>

# Via API
curl http://localhost:8080/api/v1/experiments/<id> | jq .

# Via DB
psql phoenix -c "SELECT * FROM experiments WHERE id='<id>';"
```

### Monitor agent tasks
```bash
# Check agent logs
tail -f /var/log/phoenix-agent.log

# Check task queue
curl http://localhost:8080/api/v1/tasks | jq .

# Agent status
curl http://localhost:8080/api/v1/agents | jq .
```

### Verify metrics flow
```bash
# Check Prometheus
curl http://localhost:9090/api/v1/query?query=up

# Check OTel collector
curl http://localhost:13133/metrics

# Check Phoenix metrics
curl http://localhost:8080/api/v1/metrics/cost-flow | jq .
```

## 🚨 Common Issues

### "Pipeline deployment failed"
- Check agent is running and connected
- Verify template exists
- Check agent has permissions to start collectors

### "Experiment stuck in running"
- Check all tasks completed
- Verify agent heartbeats are recent
- Check for errors in agent logs

### "No metrics data"
- Verify OTel collector is running
- Check Prometheus is scraping
- Ensure experiment ran long enough

### "WebSocket not receiving events"
- Check WebSocket upgrade succeeds
- Verify hub.Broadcast() calls exist
- Check for panic recovery hiding errors

## 📝 Testing Checklist

Before declaring MVP ready:

- [ ] Run `./scripts/mvp-validation.sh` - all pass
- [ ] Complete one full experiment via CLI only
- [ ] Verify cost savings calculation accuracy
- [ ] Test stop/rollback leaves no orphans
- [ ] WebSocket shows real-time updates
- [ ] Dashboard displays experiment results
- [ ] E2E tests pass 3 times in a row
- [ ] Load test with 10 concurrent experiments
- [ ] Documentation matches implementation

## 🎯 MVP Definition of Done

1. **Core Flow Works**: Create → Deploy → Run → Analyze → Results
2. **All CLI Commands Function**: Per MVP checklist
3. **Metrics Are Accurate**: KPIs match manual calculations
4. **Real-time Updates Work**: WebSocket events for all transitions
5. **Error Handling Is Clear**: Invalid inputs produce helpful messages
6. **Tests Pass Consistently**: E2E suite green
7. **No Critical TODOs**: Production paths complete
8. **Documentation Current**: README and guides accurate