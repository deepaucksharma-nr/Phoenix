# Service Integration Test Scenarios

**Version**: 1.0  
**Purpose**: Define comprehensive integration test scenarios for Phoenix Platform services

## Overview

This document outlines critical integration test scenarios that validate service-to-service communication, data flow, and system behavior under various conditions.

## Test Environment Setup

```go
// test/integration/setup.go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

type TestEnvironment struct {
    PostgresContainer   testcontainers.Container
    PrometheusContainer testcontainers.Container
    RedisContainer      testcontainers.Container
    APIGateway          string
    ExperimentController string
    ConfigGenerator     string
}

func SetupIntegrationEnvironment(t *testing.T) (*TestEnvironment, func()) {
    ctx := context.Background()
    
    // Start PostgreSQL
    postgresContainer, _ := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        postgres.WithDatabase("phoenix_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    
    // Start Prometheus
    prometheusContainer, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "prom/prometheus:latest",
            ExposedPorts: []string{"9090/tcp"},
            WaitingFor:   wait.ForHTTP("/").WithPort("9090"),
        },
        Started: true,
    })
    
    // Cleanup function
    cleanup := func() {
        postgresContainer.Terminate(ctx)
        prometheusContainer.Terminate(ctx)
    }
    
    return env, cleanup
}
```

## Scenario 1: Complete Experiment Lifecycle

### Description
Test the full experiment lifecycle from creation through completion with real metrics collection.

### Test Implementation

```go
func TestCompleteExperimentLifecycle(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    // 1. Create experiment via API Gateway
    experiment := createExperiment(t, env.APIGateway, &CreateExperimentRequest{
        Name:              "integration-test-exp",
        BaselinePipeline:  "process-baseline-v1",
        CandidatePipeline: "process-aggregated-v1",
        TargetNodes:       []string{"node1", "node2"},
        Config: ExperimentConfig{
            Duration:         5 * time.Minute,
            TrafficSplit:     &TrafficSplit{Baseline: 50, Candidate: 50},
            SuccessCriteria: &SuccessCriteria{
                MinCardinalityReduction: 30.0,
                MaxLatencyIncrease:      10.0,
                CriticalProcessCoverage: 95.0,
            },
        },
    })
    
    // 2. Verify experiment transitions to initializing
    waitForState(t, env, experiment.ID, "initializing", 30*time.Second)
    
    // 3. Verify config generation
    configs := getGeneratedConfigs(t, env.ConfigGenerator, experiment.ID)
    assert.NotNil(t, configs.BaselineConfig)
    assert.NotNil(t, configs.CandidateConfig)
    
    // 4. Verify pipeline deployments
    deployments := getPipelineDeployments(t, env, experiment.ID)
    assert.Len(t, deployments, 2) // baseline and candidate
    
    // 5. Start load simulation
    startLoadSimulation(t, env, experiment.ID, LoadProfile{
        ProcessCount:     100,
        MetricsPerSecond: 1000,
        Duration:         5 * time.Minute,
    })
    
    // 6. Verify experiment running
    waitForState(t, env, experiment.ID, "running", 1*time.Minute)
    
    // 7. Monitor metrics collection
    time.Sleep(2 * time.Minute)
    metrics := getExperimentMetrics(t, env, experiment.ID)
    assert.Greater(t, metrics.BaselineMetrics.TimeSeriesCount, int64(0))
    assert.Greater(t, metrics.CandidateMetrics.TimeSeriesCount, int64(0))
    
    // 8. Wait for analysis
    waitForState(t, env, experiment.ID, "analyzing", 6*time.Minute)
    
    // 9. Verify completion
    waitForState(t, env, experiment.ID, "completed", 2*time.Minute)
    
    // 10. Validate results
    results := getExperimentResults(t, env, experiment.ID)
    assert.NotNil(t, results)
    assert.Greater(t, results.CardinalityReduction, 0.0)
    assert.NotEmpty(t, results.Recommendation)
}
```

## Scenario 2: Multi-Experiment Concurrency

### Description
Test multiple experiments running concurrently without interference.

### Test Implementation

```go
func TestMultiExperimentConcurrency(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    const numExperiments = 5
    experiments := make([]*Experiment, numExperiments)
    
    // Create multiple experiments concurrently
    var wg sync.WaitGroup
    for i := 0; i < numExperiments; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            experiments[idx] = createExperiment(t, env.APIGateway, &CreateExperimentRequest{
                Name:              fmt.Sprintf("concurrent-exp-%d", idx),
                BaselinePipeline:  "process-baseline-v1",
                CandidatePipeline: "process-aggregated-v1",
                TargetNodes:       []string{fmt.Sprintf("node%d", idx)},
            })
        }(i)
    }
    wg.Wait()
    
    // Verify all experiments are running independently
    for _, exp := range experiments {
        waitForState(t, env, exp.ID, "running", 2*time.Minute)
    }
    
    // Verify no cross-contamination of metrics
    for i, exp := range experiments {
        metrics := getExperimentMetrics(t, env, exp.ID)
        // Each experiment should only have metrics from its assigned node
        assert.Contains(t, metrics.Labels["node"], fmt.Sprintf("node%d", i))
    }
}
```

## Scenario 3: Failure Recovery

### Description
Test system behavior when components fail and recover.

### Test Implementation

```go
func TestFailureRecovery(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    // Create and start experiment
    experiment := createAndStartExperiment(t, env)
    waitForState(t, env, experiment.ID, "running", 1*time.Minute)
    
    // Simulate config generator failure
    stopService(t, env.ConfigGenerator)
    time.Sleep(30 * time.Second)
    
    // Verify experiment continues running
    state := getExperimentState(t, env, experiment.ID)
    assert.Equal(t, "running", state)
    
    // Restart config generator
    startService(t, env.ConfigGenerator)
    time.Sleep(30 * time.Second)
    
    // Create new experiment to verify recovery
    newExp := createExperiment(t, env.APIGateway, &CreateExperimentRequest{
        Name: "post-recovery-exp",
    })
    
    // Verify new experiment processes normally
    waitForState(t, env, newExp.ID, "running", 2*time.Minute)
}
```

## Scenario 4: Pipeline Hot Reload

### Description
Test updating pipeline configuration without disrupting running experiments.

### Test Implementation

```go
func TestPipelineHotReload(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    // Start experiment with initial configuration
    experiment := createAndStartExperiment(t, env)
    
    // Collect baseline metrics for 1 minute
    time.Sleep(1 * time.Minute)
    beforeMetrics := getExperimentMetrics(t, env, experiment.ID)
    
    // Update pipeline configuration
    updatePipelineConfig(t, env, experiment.ID, map[string]interface{}{
        "processors": []map[string]interface{}{
            {
                "type": "filter",
                "config": map[string]interface{}{
                    "include_patterns": []string{"critical_*"},
                },
            },
        },
    })
    
    // Wait for configuration to propagate
    time.Sleep(30 * time.Second)
    
    // Verify metrics show configuration change effect
    afterMetrics := getExperimentMetrics(t, env, experiment.ID)
    assert.Less(t, afterMetrics.CandidateMetrics.TimeSeriesCount, 
                   beforeMetrics.CandidateMetrics.TimeSeriesCount)
}
```

## Scenario 5: Data Consistency

### Description
Verify data consistency across services during high load.

### Test Implementation

```go
func TestDataConsistency(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    const numRequests = 1000
    results := make(chan *Experiment, numRequests)
    
    // Create many experiments rapidly
    for i := 0; i < numRequests; i++ {
        go func(idx int) {
            exp := createExperiment(t, env.APIGateway, &CreateExperimentRequest{
                Name: fmt.Sprintf("consistency-test-%d", idx),
            })
            results <- exp
        }(i)
    }
    
    // Collect all created experiments
    experiments := make([]*Experiment, 0, numRequests)
    for i := 0; i < numRequests; i++ {
        experiments = append(experiments, <-results)
    }
    
    // Verify all experiments are in database
    for _, exp := range experiments {
        dbExp := getExperimentFromDB(t, env, exp.ID)
        assert.Equal(t, exp.ID, dbExp.ID)
        assert.Equal(t, exp.Name, dbExp.Name)
    }
    
    // Verify no duplicate IDs
    idMap := make(map[string]bool)
    for _, exp := range experiments {
        assert.False(t, idMap[exp.ID], "Duplicate ID found: %s", exp.ID)
        idMap[exp.ID] = true
    }
}
```

## Scenario 6: WebSocket Real-time Updates

### Description
Test WebSocket connections for real-time experiment updates.

### Test Implementation

```go
func TestWebSocketUpdates(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    // Connect WebSocket client
    ws := connectWebSocket(t, env.APIGateway, "/ws")
    defer ws.Close()
    
    // Subscribe to experiment updates
    subscribe(t, ws, "experiment.updates")
    
    // Create and start experiment
    experiment := createAndStartExperiment(t, env)
    
    // Collect updates
    updates := make([]ExperimentUpdate, 0)
    timeout := time.After(2 * time.Minute)
    
    for {
        select {
        case msg := <-readWebSocket(ws):
            var update ExperimentUpdate
            json.Unmarshal(msg, &update)
            updates = append(updates, update)
            
        case <-timeout:
            // Verify we received expected state transitions
            states := extractStates(updates)
            assert.Contains(t, states, "pending")
            assert.Contains(t, states, "initializing")
            assert.Contains(t, states, "running")
            return
        }
    }
}
```

## Scenario 7: Security Boundaries

### Description
Test tenant isolation and access control.

### Test Implementation

```go
func TestTenantIsolation(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    // Create experiments for different tenants
    tenant1Token := loginAsTenant(t, env, "tenant1")
    tenant2Token := loginAsTenant(t, env, "tenant2")
    
    exp1 := createExperimentWithToken(t, env.APIGateway, tenant1Token, &CreateExperimentRequest{
        Name: "tenant1-experiment",
    })
    
    exp2 := createExperimentWithToken(t, env.APIGateway, tenant2Token, &CreateExperimentRequest{
        Name: "tenant2-experiment",
    })
    
    // Verify tenant1 cannot access tenant2's experiment
    _, err := getExperimentWithToken(t, env.APIGateway, tenant1Token, exp2.ID)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "forbidden")
    
    // Verify tenant2 cannot access tenant1's experiment
    _, err = getExperimentWithToken(t, env.APIGateway, tenant2Token, exp1.ID)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "forbidden")
    
    // Verify each tenant can access their own experiments
    t1Exp := getExperimentWithToken(t, env.APIGateway, tenant1Token, exp1.ID)
    assert.Equal(t, exp1.ID, t1Exp.ID)
    
    t2Exp := getExperimentWithToken(t, env.APIGateway, tenant2Token, exp2.ID)
    assert.Equal(t, exp2.ID, t2Exp.ID)
}
```

## Scenario 8: Performance Under Load

### Description
Test system performance under sustained high load.

### Test Implementation

```go
func TestPerformanceUnderLoad(t *testing.T) {
    env, cleanup := SetupIntegrationEnvironment(t)
    defer cleanup()
    
    // Metrics collection
    latencies := make([]time.Duration, 0)
    errors := make([]error, 0)
    var mu sync.Mutex
    
    // Generate load
    const concurrent = 100
    const requestsPerClient = 100
    
    start := time.Now()
    var wg sync.WaitGroup
    
    for i := 0; i < concurrent; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()
            
            for j := 0; j < requestsPerClient; j++ {
                reqStart := time.Now()
                
                _, err := createExperiment(t, env.APIGateway, &CreateExperimentRequest{
                    Name: fmt.Sprintf("load-test-%d-%d", clientID, j),
                })
                
                elapsed := time.Since(reqStart)
                
                mu.Lock()
                latencies = append(latencies, elapsed)
                if err != nil {
                    errors = append(errors, err)
                }
                mu.Unlock()
            }
        }(i)
    }
    
    wg.Wait()
    totalTime := time.Since(start)
    
    // Calculate metrics
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })
    
    p50 := latencies[len(latencies)*50/100]
    p95 := latencies[len(latencies)*95/100]
    p99 := latencies[len(latencies)*99/100]
    
    // Assert performance requirements
    assert.Less(t, p99, 100*time.Millisecond, "P99 latency exceeds 100ms")
    assert.Less(t, float64(len(errors))/float64(len(latencies)), 0.01, "Error rate exceeds 1%")
    
    throughput := float64(concurrent*requestsPerClient) / totalTime.Seconds()
    assert.Greater(t, throughput, 1000.0, "Throughput less than 1000 req/s")
    
    t.Logf("Performance Results: P50=%v, P95=%v, P99=%v, Errors=%d, Throughput=%.2f req/s",
        p50, p95, p99, len(errors), throughput)
}
```

## Test Execution Strategy

### Continuous Integration
```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Start test environment
        run: docker-compose -f test/docker-compose.test.yml up -d
        
      - name: Run integration tests
        run: |
          go test -tags=integration \
            -timeout=30m \
            -run="TestCompleteExperimentLifecycle|TestMultiExperiment" \
            ./test/integration/...
            
      - name: Run performance tests
        if: github.ref == 'refs/heads/main'
        run: |
          go test -tags=integration,performance \
            -timeout=1h \
            -run="TestPerformanceUnderLoad" \
            ./test/integration/...
```

### Local Development
```bash
# Run specific scenario
make test-integration SCENARIO=TestCompleteExperimentLifecycle

# Run all integration tests
make test-integration-all

# Run with verbose output
make test-integration VERBOSE=1
```

## Success Criteria

Each integration test scenario must:
1. Complete within defined timeout
2. Pass all assertions
3. Clean up resources after completion
4. Be idempotent (same result on repeated runs)
5. Not interfere with other tests

## Monitoring Integration Tests

### Metrics to Track
- Test execution time
- Test failure rate
- Resource usage during tests
- Test coverage of API endpoints

### Alerts
- Test failure rate > 5%
- Test execution time > 2x baseline
- Resource exhaustion during tests