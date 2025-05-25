# Phoenix Platform Performance Tuning Guide

**Version**: 1.0  
**Last Updated**: January 25, 2025

## Overview

This guide provides comprehensive performance tuning recommendations for the Phoenix Platform to achieve optimal throughput, minimal latency, and efficient resource utilization.

## 1. Performance Targets

### 1.1 System-Wide Targets

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| API Latency (p99) | <100ms | TBD | 🟡 |
| Experiment Creation | <500ms | TBD | 🟡 |
| Metrics Ingestion | 1M/sec | TBD | 🟡 |
| Dashboard Load Time | <2s | TBD | 🟡 |
| OTel Collector CPU | <5% | TBD | 🟡 |
| Memory Efficiency | <512MB/collector | TBD | 🟡 |

### 1.2 Scalability Targets

- Support 1,000 concurrent experiments
- Handle 10,000 API requests/second
- Process 1M metrics/second across the platform
- Maintain performance with 10,000 WebSocket connections

## 2. Database Optimization

### 2.1 PostgreSQL Configuration

```sql
-- Performance-optimized postgresql.conf settings
-- Based on 32GB RAM, 8 vCPU instance

-- Memory Settings
shared_buffers = 8GB                    # 25% of RAM
effective_cache_size = 24GB             # 75% of RAM
work_mem = 64MB                         # RAM / (max_connections * 2)
maintenance_work_mem = 2GB              # For VACUUM, indexes
wal_buffers = 64MB                      # For write-heavy workload

-- Checkpoint Settings
checkpoint_timeout = 15min              # Reduce I/O spikes
checkpoint_completion_target = 0.9      # Spread checkpoint I/O
max_wal_size = 8GB
min_wal_size = 2GB

-- Connection Pooling
max_connections = 200                   # Use connection pooler instead
max_prepared_transactions = 100         # For 2PC if needed

-- Query Planning
random_page_cost = 1.1                  # For SSD storage
effective_io_concurrency = 200          # For SSD storage
default_statistics_target = 1000        # Better query plans

-- Parallel Query
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
parallel_leader_participation = on

-- Monitoring
shared_preload_libraries = 'pg_stat_statements,auto_explain'
pg_stat_statements.track = all
auto_explain.log_min_duration = '100ms'
```

### 2.2 Index Optimization

```sql
-- Critical indexes for performance

-- Experiments table
CREATE INDEX CONCURRENTLY idx_experiments_state_updated 
ON experiments(state, updated_at DESC) 
WHERE state IN ('running', 'analyzing');

CREATE INDEX CONCURRENTLY idx_experiments_owner_created 
ON experiments(owner_id, created_at DESC);

CREATE INDEX CONCURRENTLY idx_experiments_gin_config 
ON experiments USING gin(config jsonb_path_ops);

-- State transitions table (partitioned)
CREATE INDEX CONCURRENTLY idx_transitions_experiment_created 
ON state_transitions(experiment_id, created_at DESC);

-- Metrics metadata
CREATE INDEX CONCURRENTLY idx_metrics_experiment_variant 
ON metrics_metadata(experiment_id, variant) 
INCLUDE (time_series_count, samples_per_second);

-- Analyze tables regularly
ANALYZE experiments;
ANALYZE state_transitions;
```

### 2.3 Connection Pooling

```yaml
# PgBouncer configuration
[databases]
phoenix = host=postgres.phoenix.local port=5432 dbname=phoenix

[pgbouncer]
listen_port = 6432
listen_addr = *
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt

# Pool settings
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
min_pool_size = 5
reserve_pool_size = 5
reserve_pool_timeout = 5

# Performance settings
server_reset_query = DISCARD ALL
server_check_delay = 30
server_check_query = select 1
server_lifetime = 3600
server_idle_timeout = 600

# Logging
log_connections = 0
log_disconnections = 0
log_pooler_errors = 1
stats_period = 60
```

## 3. Application-Level Optimization

### 3.1 Go Service Optimization

```go
// Optimized HTTP server configuration
package main

import (
    "net/http"
    "time"
    "runtime"
)

func optimizedServer() *http.Server {
    // Set GOMAXPROCS to number of CPUs
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Enable memory ballast for GC optimization
    ballast := make([]byte, 256<<20) // 256MB
    _ = ballast
    
    return &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
        
        // Optimize connection handling
        MaxHeaderBytes: 1 << 20, // 1MB
    }
}

// Connection pool for database
func optimizedDBPool() *sql.DB {
    db, _ := sql.Open("postgres", connStr)
    
    // Connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(1 * time.Minute)
    
    return db
}

// Optimized context handling
func optimizedHandler(w http.ResponseWriter, r *http.Request) {
    // Use sync.Pool for object reuse
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    // Set appropriate timeouts
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    
    // Process request...
}
```

### 3.2 Caching Strategy

```go
// Multi-level caching implementation
package cache

import (
    "github.com/dgraph-io/ristretto"
    "github.com/go-redis/redis/v8"
)

type MultiLevelCache struct {
    l1      *ristretto.Cache  // In-memory (hot data)
    l2      *redis.Client     // Redis (warm data)
    l3      storage.Client    // S3/Database (cold data)
    metrics *CacheMetrics
}

func NewMultiLevelCache() (*MultiLevelCache, error) {
    // L1: In-memory cache (1GB)
    l1Config := &ristretto.Config{
        NumCounters: 1e7,     // 10 million
        MaxCost:     1 << 30, // 1GB
        BufferItems: 64,
        Metrics:     true,
    }
    l1, _ := ristretto.NewCache(l1Config)
    
    // L2: Redis configuration
    l2 := redis.NewClient(&redis.Options{
        Addr:         "redis:6379",
        PoolSize:     100,
        MinIdleConns: 10,
        MaxRetries:   3,
        PoolTimeout:  4 * time.Second,
    })
    
    return &MultiLevelCache{
        l1: l1,
        l2: l2,
    }, nil
}

// Optimized cache key patterns
const (
    ExperimentKey = "exp:%s"          // experiment:{id}
    MetricsKey    = "met:%s:%s"       // metrics:{exp_id}:{variant}
    PipelineKey   = "pip:%s:%s"       // pipeline:{name}:{version}
    UserKey       = "usr:%s"          // user:{id}
)

// Cache warming strategy
func (c *MultiLevelCache) WarmCache(ctx context.Context) error {
    // Preload frequently accessed data
    experiments, _ := db.GetActiveExperiments(ctx)
    for _, exp := range experiments {
        key := fmt.Sprintf(ExperimentKey, exp.ID)
        c.l1.Set(key, exp, int64(unsafe.Sizeof(exp)))
        
        // Also warm L2
        data, _ := json.Marshal(exp)
        c.l2.Set(ctx, key, data, 1*time.Hour)
    }
    
    return nil
}
```

### 3.3 Batch Processing

```go
// Optimized batch processing for metrics
package metrics

type BatchProcessor struct {
    batchSize     int
    flushInterval time.Duration
    workers       int
    queue         chan Metric
    batches       chan []Metric
}

func NewBatchProcessor() *BatchProcessor {
    bp := &BatchProcessor{
        batchSize:     1000,
        flushInterval: 100 * time.Millisecond,
        workers:       runtime.NumCPU(),
        queue:         make(chan Metric, 100000),
        batches:       make(chan []Metric, 100),
    }
    
    // Start workers
    for i := 0; i < bp.workers; i++ {
        go bp.worker()
    }
    
    // Start batcher
    go bp.batcher()
    
    return bp
}

func (bp *BatchProcessor) batcher() {
    batch := make([]Metric, 0, bp.batchSize)
    ticker := time.NewTicker(bp.flushInterval)
    
    for {
        select {
        case metric := <-bp.queue:
            batch = append(batch, metric)
            if len(batch) >= bp.batchSize {
                bp.batches <- batch
                batch = make([]Metric, 0, bp.batchSize)
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                bp.batches <- batch
                batch = make([]Metric, 0, bp.batchSize)
            }
        }
    }
}

func (bp *BatchProcessor) worker() {
    for batch := range bp.batches {
        bp.processBatch(batch)
    }
}

func (bp *BatchProcessor) processBatch(batch []Metric) {
    // Use prepared statements for batch insert
    stmt := `
        INSERT INTO metrics (experiment_id, variant, name, value, timestamp)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (experiment_id, variant, name, timestamp) 
        DO UPDATE SET value = EXCLUDED.value
    `
    
    tx, _ := db.Begin()
    defer tx.Rollback()
    
    pstmt, _ := tx.Prepare(stmt)
    defer pstmt.Close()
    
    for _, metric := range batch {
        pstmt.Exec(
            metric.ExperimentID,
            metric.Variant,
            metric.Name,
            metric.Value,
            metric.Timestamp,
        )
    }
    
    tx.Commit()
}
```

## 4. Kubernetes Optimization

### 4.1 Resource Allocation

```yaml
# Optimized deployment configuration
apiVersion: apps/v1
kind: Deployment
metadata:
  name: phoenix-api
spec:
  replicas: 3
  template:
    spec:
      # Node affinity for performance
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: node.kubernetes.io/instance-type
                operator: In
                values:
                - m5.xlarge
                - m5.2xlarge
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: phoenix-api
            topologyKey: kubernetes.io/hostname
      
      containers:
      - name: api
        image: phoenix/api:latest
        
        # Resource limits and requests
        resources:
          requests:
            cpu: 250m
            memory: 512Mi
            ephemeral-storage: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
            ephemeral-storage: 2Gi
        
        # JVM/Go runtime optimization
        env:
        - name: GOMAXPROCS
          value: "2"
        - name: GOMEMLIMIT
          value: "1800MiB"
        - name: GOGC
          value: "100"
        
        # Readiness and liveness probes
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          successThreshold: 1
          failureThreshold: 3
          
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
```

### 4.2 HPA Configuration

```yaml
# Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: phoenix-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: phoenix-api
  minReplicas: 3
  maxReplicas: 20
  
  metrics:
  # CPU-based scaling
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
        
  # Memory-based scaling
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
        
  # Custom metrics scaling
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1000"
        
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
      - type: Pods
        value: 4
        periodSeconds: 60
      selectPolicy: Max
      
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### 4.3 Network Optimization

```yaml
# Optimized service mesh configuration
apiVersion: v1
kind: Service
metadata:
  name: phoenix-api
  annotations:
    # Session affinity for WebSocket connections
    service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled: "true"
    service.beta.kubernetes.io/aws-load-balancer-connection-draining-timeout: "60"
spec:
  type: LoadBalancer
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 3600
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  - name: grpc
    port: 50051
    targetPort: 50051
    protocol: TCP
```

## 5. OTel Collector Optimization

### 5.1 Collector Configuration

```yaml
# Optimized OpenTelemetry Collector config
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
        max_recv_msg_size_mib: 16
        max_concurrent_streams: 100
        keepalive:
          server_parameters:
            max_connection_idle: 60s
            max_connection_age: 300s

processors:
  # Batch processing for efficiency
  batch:
    send_batch_size: 8192
    timeout: 200ms
    send_batch_max_size: 10000
    
  # Memory limiter to prevent OOM
  memory_limiter:
    check_interval: 1s
    limit_mib: 400
    spike_limit_mib: 100
    
  # Resource detection
  resource:
    attributes:
      - key: service.instance.id
        from_attribute: k8s.pod.name
        action: insert
        
  # Sampling for high-volume metrics
  probabilistic_sampler:
    sampling_percentage: 10
    
  # Filtering
  filter:
    metrics:
      exclude:
        match_type: regexp
        metric_names:
          - .*_bucket
          - .*_info

exporters:
  # Prometheus exporter with optimizations
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: phoenix
    resource_to_telemetry_conversion:
      enabled: true
    enable_open_metrics: true
    
  # New Relic exporter with batching
  otlp/newrelic:
    endpoint: otlp.nr-data.net:4317
    compression: gzip
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 300s
    sending_queue:
      enabled: true
      num_consumers: 10
      queue_size: 5000

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource, filter]
      exporters: [prometheus, otlp/newrelic]
```

### 5.2 Collector Resource Management

```yaml
# DaemonSet with resource optimization
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: otel-collector
spec:
  template:
    spec:
      containers:
      - name: collector
        image: otel/opentelemetry-collector:latest
        
        # CPU and memory limits
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
            
        # Volume mounts for performance
        volumeMounts:
        - name: cache
          mountPath: /var/cache/otel
          
        env:
        # Garbage collection tuning
        - name: GOGC
          value: "80"
        - name: GOMEMLIMIT
          value: "450MiB"
          
      volumes:
      # EmptyDir for temporary cache
      - name: cache
        emptyDir:
          sizeLimit: 1Gi
          medium: Memory  # Use RAM for cache
```

## 6. Frontend Performance

### 6.1 React Optimization

```typescript
// Optimized React component patterns
import React, { memo, useMemo, useCallback, lazy, Suspense } from 'react';

// Lazy load heavy components
const HeavyChart = lazy(() => import('./components/HeavyChart'));

// Memoized component
const ExperimentCard = memo(({ experiment, onSelect }) => {
  // Memoize expensive calculations
  const costSavings = useMemo(() => {
    return calculateCostSavings(experiment.metrics);
  }, [experiment.metrics]);
  
  // Memoize callbacks
  const handleClick = useCallback(() => {
    onSelect(experiment.id);
  }, [experiment.id, onSelect]);
  
  return (
    <Card onClick={handleClick}>
      <h3>{experiment.name}</h3>
      <p>Savings: ${costSavings}</p>
    </Card>
  );
}, (prevProps, nextProps) => {
  // Custom comparison for re-rendering
  return prevProps.experiment.id === nextProps.experiment.id &&
         prevProps.experiment.state === nextProps.experiment.state;
});

// Virtual scrolling for large lists
import { FixedSizeList } from 'react-window';

const ExperimentList = ({ experiments }) => {
  const Row = ({ index, style }) => (
    <div style={style}>
      <ExperimentCard experiment={experiments[index]} />
    </div>
  );
  
  return (
    <FixedSizeList
      height={600}
      itemCount={experiments.length}
      itemSize={120}
      width="100%"
    >
      {Row}
    </FixedSizeList>
  );
};
```

### 6.2 Bundle Optimization

```javascript
// webpack.config.js
module.exports = {
  optimization: {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          priority: 10,
        },
        common: {
          minChunks: 2,
          priority: 5,
          reuseExistingChunk: true,
        },
      },
    },
    // Tree shaking
    usedExports: true,
    // Minification
    minimize: true,
    // Module concatenation
    concatenateModules: true,
  },
  
  // Performance hints
  performance: {
    maxEntrypointSize: 300000,
    maxAssetSize: 250000,
    hints: 'warning',
  },
};
```

## 7. Monitoring Performance

### 7.1 Performance Metrics

```go
// Custom performance metrics
var (
    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "phoenix_request_duration_seconds",
            Help: "Request duration distribution",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint", "status"},
    )
    
    DatabaseQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "phoenix_db_query_duration_seconds",
            Help: "Database query duration",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
        },
        []string{"query_type", "table"},
    )
    
    CacheHitRate = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "phoenix_cache_hits_total",
            Help: "Cache hit/miss counts",
        },
        []string{"cache_level", "result"},
    )
)
```

### 7.2 Performance Dashboard

```json
{
  "dashboard": {
    "title": "Phoenix Performance Dashboard",
    "panels": [
      {
        "title": "API Latency Distribution",
        "query": "histogram_quantile(0.99, rate(phoenix_request_duration_seconds_bucket[5m]))"
      },
      {
        "title": "Database Query Performance",
        "query": "avg by (query_type) (rate(phoenix_db_query_duration_seconds_sum[5m]) / rate(phoenix_db_query_duration_seconds_count[5m]))"
      },
      {
        "title": "Cache Hit Rate",
        "query": "rate(phoenix_cache_hits_total{result=\"hit\"}[5m]) / (rate(phoenix_cache_hits_total{result=\"hit\"}[5m]) + rate(phoenix_cache_hits_total{result=\"miss\"}[5m]))"
      },
      {
        "title": "Goroutines",
        "query": "go_goroutines{job=\"phoenix-api\"}"
      },
      {
        "title": "GC Pause Duration",
        "query": "rate(go_gc_duration_seconds_sum[5m])"
      }
    ]
  }
}
```

## 8. Load Testing

### 8.1 Load Test Scenarios

```go
// k6 load test script
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up
    { duration: '5m', target: 1000 },  // Stay at 1000 users
    { duration: '2m', target: 10000 }, // Spike test
    { duration: '5m', target: 1000 },  // Back to normal
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<100'], // 99% of requests under 100ms
    errors: ['rate<0.01'],            // Error rate under 1%
  },
};

export default function() {
  // Create experiment
  const payload = JSON.stringify({
    name: `Load test ${Date.now()}`,
    baseline_pipeline: 'process-baseline-v1',
    candidate_pipeline: 'process-aggregated-v1',
    target_nodes: ['node1', 'node2'],
  });
  
  const res = http.post('https://api.phoenix.io/v1/experiments', payload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + __ENV.API_TOKEN,
    },
  });
  
  const success = check(res, {
    'status is 201': (r) => r.status === 201,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(!success);
  
  sleep(1);
}
```

## 9. Performance Troubleshooting

### 9.1 Common Issues and Solutions

| Issue | Symptoms | Solution |
|-------|----------|----------|
| **Slow API Response** | p99 > 200ms | Enable query logging, check slow queries |
| **High Memory Usage** | OOM kills | Tune GOMEMLIMIT, add memory limits |
| **Database Locks** | Timeouts | Check long-running transactions |
| **Cache Misses** | High latency variance | Implement cache warming |
| **GC Pressure** | CPU spikes | Tune GOGC, reduce allocations |

### 9.2 Performance Debugging Tools

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Trace analysis
wget http://localhost:6060/debug/pprof/trace?seconds=10
go tool trace trace

# Database slow query log
tail -f /var/log/postgresql/postgresql-slow.log
```

## 10. Best Practices

### 10.1 Code-Level Optimizations

1. **Avoid Allocations in Hot Paths**
2. **Use sync.Pool for Object Reuse**
3. **Batch Database Operations**
4. **Implement Circuit Breakers**
5. **Use Context Timeouts**
6. **Profile Before Optimizing**

### 10.2 Infrastructure Optimizations

1. **Right-size Kubernetes Resources**
2. **Use Node Affinity for Performance**
3. **Enable Cluster Autoscaling**
4. **Optimize Network Policies**
5. **Use SSD Storage for Databases**
6. **Implement Geographic Load Balancing**