# Missing Implementations in Phoenix Platform

**Date**: January 25, 2025  
**Priority**: High  
**Impact**: Production Readiness

## Overview

This document identifies specific implementations that are currently missing or mocked in the Phoenix Platform codebase, with detailed specifications for completion.

## 1. Metrics Collection Service

### Current State
The state machine in `cmd/controller/internal/controller/state_machine.go` uses hardcoded mock metrics:

```go
// CURRENT: Mock implementation
results := &ExperimentResults{
    BaselineMetrics: MetricsSnapshot{
        TimeSeriesCount:  10000,  // Hardcoded
        SamplesPerSecond: 1000,   // Hardcoded
    },
}
```

### Required Implementation

```go
// pkg/metrics/prometheus_client.go
package metrics

import (
    "context"
    "fmt"
    "time"
    
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
)

type PrometheusClient struct {
    api v1.API
}

func NewPrometheusClient(address string) (*PrometheusClient, error) {
    client, err := api.NewClient(api.Config{
        Address: address,
    })
    if err != nil {
        return nil, err
    }
    
    return &PrometheusClient{
        api: v1.NewAPI(client),
    }, nil
}

func (p *PrometheusClient) QueryExperimentMetrics(ctx context.Context, experimentID, variant string) (*MetricsSnapshot, error) {
    endTime := time.Now()
    startTime := endTime.Add(-5 * time.Minute)
    
    // Query time series count
    tsQuery := fmt.Sprintf(`count(count by (__name__)({experiment_id="%s",variant="%s"}))`, experimentID, variant)
    tsResult, _, err := p.api.Query(ctx, tsQuery, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to query time series count: %w", err)
    }
    
    // Query samples per second
    spsQuery := fmt.Sprintf(`rate(prometheus_tsdb_sample_appended_total{experiment_id="%s",variant="%s"}[5m])`, experimentID, variant)
    spsResult, _, err := p.api.Query(ctx, spsQuery, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to query samples per second: %w", err)
    }
    
    // Query CPU usage
    cpuQuery := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{pod=~"otel-collector-%s-%s-.*"}[5m])`, experimentID, variant)
    cpuResult, _, err := p.api.Query(ctx, cpuQuery, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to query CPU usage: %w", err)
    }
    
    return &MetricsSnapshot{
        Timestamp:        endTime,
        TimeSeriesCount:  extractScalarValue(tsResult),
        SamplesPerSecond: extractScalarValue(spsResult),
        CPUUsage:         extractScalarValue(cpuResult) * 100, // Convert to percentage
        MemoryUsage:      p.queryMemoryUsage(ctx, experimentID, variant),
        ProcessCount:     p.queryProcessCount(ctx, experimentID, variant),
    }, nil
}

func (p *PrometheusClient) CompareExperiments(ctx context.Context, experimentID string) (*ExperimentComparison, error) {
    baseline, err := p.QueryExperimentMetrics(ctx, experimentID, "baseline")
    if err != nil {
        return nil, fmt.Errorf("failed to query baseline metrics: %w", err)
    }
    
    candidate, err := p.QueryExperimentMetrics(ctx, experimentID, "candidate")
    if err != nil {
        return nil, fmt.Errorf("failed to query candidate metrics: %w", err)
    }
    
    return &ExperimentComparison{
        BaselineMetrics:      baseline,
        CandidateMetrics:     candidate,
        CardinalityReduction: calculateReduction(baseline.TimeSeriesCount, candidate.TimeSeriesCount),
        CPUImprovement:       calculateReduction(baseline.CPUUsage, candidate.CPUUsage),
        MemoryImprovement:    calculateReduction(baseline.MemoryUsage, candidate.MemoryUsage),
    }, nil
}
```

## 2. Process Simulator Integration

### Current State
Process simulator exists but isn't integrated with the experiment controller.

### Required Implementation

```go
// pkg/simulator/client.go
package simulator

import (
    "context"
    "fmt"
    
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

type SimulatorClient struct {
    k8sClient kubernetes.Interface
    namespace string
}

func (s *SimulatorClient) StartLoadSimulation(ctx context.Context, req *LoadSimulationRequest) error {
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("loadsim-%s", req.ExperimentID),
            Namespace: s.namespace,
            Labels: map[string]string{
                "app":           "process-simulator",
                "experiment-id": req.ExperimentID,
            },
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        {
                            Name:  "simulator",
                            Image: "phoenix/process-simulator:latest",
                            Env: []corev1.EnvVar{
                                {Name: "EXPERIMENT_ID", Value: req.ExperimentID},
                                {Name: "PROFILE", Value: req.Profile},
                                {Name: "DURATION", Value: req.Duration.String()},
                                {Name: "PROCESS_COUNT", Value: fmt.Sprintf("%d", req.ProcessCount)},
                                {Name: "METRICS_INTERVAL", Value: req.MetricsInterval.String()},
                                {Name: "TARGET_ENDPOINT", Value: req.TargetEndpoint},
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("100m"),
                                    corev1.ResourceMemory: resource.MustParse("256Mi"),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("500m"),
                                    corev1.ResourceMemory: resource.MustParse("512Mi"),
                                },
                            },
                        },
                    },
                },
            },
            BackoffLimit:            ptr.Int32(3),
            TTLSecondsAfterFinished: ptr.Int32(300),
        },
    }
    
    _, err := s.k8sClient.BatchV1().Jobs(s.namespace).Create(ctx, job, metav1.CreateOptions{})
    return err
}
```

## 3. Statistical Analysis Engine

### Current State
No statistical significance calculation for experiment results.

### Required Implementation

```go
// pkg/analysis/statistics.go
package analysis

import (
    "math"
    "gonum.org/v1/gonum/stat"
    "gonum.org/v1/gonum/stat/distuv"
)

type StatisticalAnalyzer struct {
    confidenceLevel float64
}

func NewStatisticalAnalyzer(confidenceLevel float64) *StatisticalAnalyzer {
    return &StatisticalAnalyzer{
        confidenceLevel: confidenceLevel,
    }
}

func (s *StatisticalAnalyzer) AnalyzeExperiment(baseline, candidate *MetricsSeries) (*StatisticalResult, error) {
    // Calculate basic statistics
    baselineMean, baselineStdDev := stat.MeanStdDev(baseline.Values, nil)
    candidateMean, candidateStdDev := stat.MeanStdDev(candidate.Values, nil)
    
    // Perform t-test
    tStat, pValue := s.performTTest(baseline.Values, candidate.Values)
    
    // Calculate confidence interval
    ci := s.calculateConfidenceInterval(
        baselineMean, candidateMean,
        baselineStdDev, candidateStdDev,
        len(baseline.Values), len(candidate.Values),
    )
    
    // Calculate effect size (Cohen's d)
    pooledStdDev := math.Sqrt(
        ((float64(len(baseline.Values)-1)*math.Pow(baselineStdDev, 2) +
          float64(len(candidate.Values)-1)*math.Pow(candidateStdDev, 2)) /
         float64(len(baseline.Values)+len(candidate.Values)-2)),
    )
    effectSize := (candidateMean - baselineMean) / pooledStdDev
    
    return &StatisticalResult{
        BaselineMean:      baselineMean,
        CandidateMean:     candidateMean,
        PercentChange:     ((candidateMean - baselineMean) / baselineMean) * 100,
        PValue:            pValue,
        ConfidenceInterval: ci,
        EffectSize:        effectSize,
        IsSignificant:     pValue < (1 - s.confidenceLevel),
        SampleSizeBaseline: len(baseline.Values),
        SampleSizeCandidate: len(candidate.Values),
    }, nil
}

func (s *StatisticalAnalyzer) performTTest(x, y []float64) (float64, float64) {
    // Welch's t-test for unequal variances
    n1, n2 := float64(len(x)), float64(len(y))
    mean1, var1 := stat.MeanVariance(x, nil)
    mean2, var2 := stat.MeanVariance(y, nil)
    
    // Calculate t-statistic
    se := math.Sqrt(var1/n1 + var2/n2)
    t := (mean1 - mean2) / se
    
    // Calculate degrees of freedom (Welch-Satterthwaite equation)
    df := math.Pow(var1/n1+var2/n2, 2) / 
         (math.Pow(var1/n1, 2)/(n1-1) + math.Pow(var2/n2, 2)/(n2-1))
    
    // Calculate p-value
    dist := distuv.StudentsT{Nu: df}
    pValue := 2 * (1 - dist.CDF(math.Abs(t)))
    
    return t, pValue
}
```

## 4. Cost Estimation Service

### Current State
No actual cost calculation based on metrics volume and cloud provider pricing.

### Required Implementation

```go
// pkg/cost/estimator.go
package cost

import (
    "context"
    "math"
)

type CostEstimator struct {
    providers map[string]PricingProvider
}

type PricingProvider interface {
    CalculateMetricsCost(volume MetricsVolume) (float64, error)
    CalculateStorageCost(sizeGB float64) (float64, error)
    CalculateComputeCost(cpu, memory float64) (float64, error)
}

type NewRelicPricing struct {
    pricePerMillionDataPoints float64
    pricePerGBMonth          float64
}

func (n *NewRelicPricing) CalculateMetricsCost(volume MetricsVolume) (float64, error) {
    // New Relic charges per million data points
    millionDataPoints := float64(volume.DataPointsPerMonth) / 1_000_000
    metricsCost := millionDataPoints * n.pricePerMillionDataPoints
    
    // Add storage cost (30-day retention)
    storageSizeGB := float64(volume.TimeSeriesCount) * 0.001 * 30 // Rough estimate
    storageCost := storageSizeGB * n.pricePerGBMonth
    
    return metricsCost + storageCost, nil
}

func (c *CostEstimator) EstimateExperimentSavings(ctx context.Context, experiment *Experiment) (*CostAnalysis, error) {
    // Get metrics volumes
    baselineVolume := c.calculateVolume(experiment.BaselineMetrics)
    candidateVolume := c.calculateVolume(experiment.CandidateMetrics)
    
    // Calculate costs for each provider
    savings := make(map[string]ProviderSavings)
    
    for provider, pricing := range c.providers {
        baselineCost, err := pricing.CalculateMetricsCost(baselineVolume)
        if err != nil {
            return nil, err
        }
        
        candidateCost, err := pricing.CalculateMetricsCost(candidateVolume)
        if err != nil {
            return nil, err
        }
        
        savings[provider] = ProviderSavings{
            BaselineCostPerMonth:  baselineCost,
            CandidateCostPerMonth: candidateCost,
            SavingsPerMonth:       baselineCost - candidateCost,
            SavingsPercentage:     ((baselineCost - candidateCost) / baselineCost) * 100,
            ProjectedAnnualSavings: (baselineCost - candidateCost) * 12,
        }
    }
    
    return &CostAnalysis{
        ExperimentID:     experiment.ID,
        ProviderSavings:  savings,
        RecommendedAction: c.generateRecommendation(savings),
        BreakEvenTime:    c.calculateBreakEvenTime(savings),
    }, nil
}
```

## 5. Multi-Tenancy Support

### Current State
Basic tenant ID in JWT claims but no actual isolation.

### Required Implementation

```go
// pkg/multitenancy/isolation.go
package multitenancy

import (
    "context"
    "fmt"
    
    "github.com/casbin/casbin/v2"
    gormadapter "github.com/casbin/gorm-adapter/v3"
)

type TenantIsolator struct {
    enforcer *casbin.Enforcer
    db       *gorm.DB
}

func NewTenantIsolator(db *gorm.DB) (*TenantIsolator, error) {
    adapter, err := gormadapter.NewAdapterByDB(db)
    if err != nil {
        return nil, err
    }
    
    enforcer, err := casbin.NewEnforcer("configs/rbac_model.conf", adapter)
    if err != nil {
        return nil, err
    }
    
    return &TenantIsolator{
        enforcer: enforcer,
        db:       db,
    }, nil
}

// Middleware for tenant isolation
func (t *TenantIsolator) IsolateByTenant() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID, exists := GetTenantID(c)
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "tenant not found"})
            c.Abort()
            return
        }
        
        // Add tenant filter to database queries
        ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantID)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}

// Database hooks for automatic tenant filtering
func (t *TenantIsolator) RegisterGormCallbacks() {
    // Automatically add tenant_id to queries
    t.db.Callback().Query().Before("gorm:query").Register("tenant:query", func(db *gorm.DB) {
        if tenantID, ok := db.Statement.Context.Value("tenant_id").(string); ok {
            db.Where("tenant_id = ?", tenantID)
        }
    })
    
    // Automatically add tenant_id to creates
    t.db.Callback().Create().Before("gorm:create").Register("tenant:create", func(db *gorm.DB) {
        if tenantID, ok := db.Statement.Context.Value("tenant_id").(string); ok {
            db.Statement.SetColumn("tenant_id", tenantID)
        }
    })
}

// Namespace isolation for Kubernetes resources
func (t *TenantIsolator) GetTenantNamespace(tenantID string) string {
    return fmt.Sprintf("phoenix-tenant-%s", tenantID)
}

// Resource quotas per tenant
func (t *TenantIsolator) ApplyTenantQuota(ctx context.Context, tenantID string) error {
    namespace := t.GetTenantNamespace(tenantID)
    
    quota := &corev1.ResourceQuota{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "tenant-quota",
            Namespace: namespace,
        },
        Spec: corev1.ResourceQuotaSpec{
            Hard: corev1.ResourceList{
                corev1.ResourcePods:                   resource.MustParse("100"),
                corev1.ResourceRequestsCPU:            resource.MustParse("50"),
                corev1.ResourceRequestsMemory:         resource.MustParse("100Gi"),
                corev1.ResourcePersistentVolumeClaims: resource.MustParse("10"),
            },
        },
    }
    
    _, err := t.k8sClient.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{})
    return err
}
```

## 6. WebSocket Real-time Updates

### Current State
WebSocket endpoint defined but not implemented.

### Required Implementation

```go
// pkg/websocket/hub.go
package websocket

import (
    "context"
    "encoding/json"
    "sync"
    
    "github.com/gorilla/websocket"
    "go.uber.org/zap"
)

type Hub struct {
    clients    map[string]map[*Client]bool // experimentID -> clients
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
    logger     *zap.Logger
}

type Client struct {
    hub          *Hub
    conn         *websocket.Conn
    send         chan []byte
    experimentID string
    userID       string
}

type Message struct {
    Type         string      `json:"type"`
    ExperimentID string      `json:"experiment_id"`
    Data         interface{} `json:"data"`
    Timestamp    time.Time   `json:"timestamp"`
}

func NewHub(logger *zap.Logger) *Hub {
    return &Hub{
        clients:    make(map[string]map[*Client]bool),
        broadcast:  make(chan Message, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        logger:     logger,
    }
}

func (h *Hub) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
            
        case client := <-h.register:
            h.mu.Lock()
            if h.clients[client.experimentID] == nil {
                h.clients[client.experimentID] = make(map[*Client]bool)
            }
            h.clients[client.experimentID][client] = true
            h.mu.Unlock()
            
            h.logger.Info("client connected",
                zap.String("user_id", client.userID),
                zap.String("experiment_id", client.experimentID),
            )
            
        case client := <-h.unregister:
            h.mu.Lock()
            if clients, ok := h.clients[client.experimentID]; ok {
                if _, ok := clients[client]; ok {
                    delete(clients, client)
                    close(client.send)
                    if len(clients) == 0 {
                        delete(h.clients, client.experimentID)
                    }
                }
            }
            h.mu.Unlock()
            
        case message := <-h.broadcast:
            h.mu.RLock()
            clients := h.clients[message.ExperimentID]
            h.mu.RUnlock()
            
            data, err := json.Marshal(message)
            if err != nil {
                h.logger.Error("failed to marshal message", zap.Error(err))
                continue
            }
            
            for client := range clients {
                select {
                case client.send <- data:
                default:
                    // Client's send channel is full, close it
                    h.unregister <- client
                }
            }
        }
    }
}

// Integration with experiment updates
func (h *Hub) PublishExperimentUpdate(experimentID string, update interface{}) {
    h.broadcast <- Message{
        Type:         "experiment.update",
        ExperimentID: experimentID,
        Data:         update,
        Timestamp:    time.Now(),
    }
}

// Client connection handler
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        h.logger.Error("websocket upgrade failed", zap.Error(err))
        return
    }
    
    // Extract user and experiment info from request
    userID := r.Header.Get("X-User-ID")
    experimentID := r.URL.Query().Get("experiment_id")
    
    client := &Client{
        hub:          h,
        conn:         conn,
        send:         make(chan []byte, 256),
        experimentID: experimentID,
        userID:       userID,
    }
    
    h.register <- client
    
    go client.writePump()
    go client.readPump()
}
```

## Implementation Priority

### Phase 1 (Week 1-2) - Critical
1. **Prometheus Metrics Client** - Replace mock metrics
2. **Statistical Analysis** - Add significance testing
3. **WebSocket Implementation** - Real-time updates

### Phase 2 (Week 3-4) - Important
1. **Process Simulator Integration** - Automated load generation
2. **Cost Estimation Service** - Accurate savings calculation
3. **Multi-tenancy Isolation** - Database and namespace separation

### Phase 3 (Week 5-6) - Enhancement
1. **Advanced Analytics** - ML-based recommendations
2. **Audit Logging** - Compliance requirements
3. **Performance Optimizations** - Caching, batching

## Testing Requirements

Each implementation must include:
1. Unit tests with >90% coverage
2. Integration tests with test containers
3. Performance benchmarks
4. Documentation updates
5. Example usage

## Success Criteria

- All mock implementations replaced with real services
- Integration tests passing for all scenarios
- Performance targets met (latency, throughput)
- Security audit passed
- Documentation complete