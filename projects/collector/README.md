# Phoenix Collector

OpenTelemetry collector distribution with Phoenix-specific processors.

## Overview

Phoenix Collector is a custom OTel collector build that includes:
- Phoenix adaptive filtering processor
- Top-K cardinality processor  
- Cost-aware sampling
- Metric enrichment pipelines

## Features

- **Adaptive Filtering**: Dynamic threshold-based filtering
- **Top-K Selection**: Keep only most important metrics
- **Cost Attribution**: Tag metrics with cost metadata
- **High Performance**: Optimized for high-volume processing

## Quick Start

```bash
# Build collector
make build

# Run with config
./phoenix-collector --config=config.yaml

# Run in container
docker run -v ./config.yaml:/etc/collector/config.yaml \
  phoenix/collector:latest
```

## Configuration

Example configuration with Phoenix processors:

```yaml
receivers:
  prometheus:
    config:
      scrape_configs:
        - job_name: 'phoenix'
          scrape_interval: 30s
          static_configs:
            - targets: ['localhost:9090']

processors:
  phoenix_adaptive_filter:
    cpu_threshold: 0.05      # 5% CPU threshold
    memory_threshold: 0.10   # 10% memory threshold
    
  phoenix_topk:
    k: 1000                  # Keep top 1000 series
    metric: cardinality      # Rank by cardinality
    
  batch:
    timeout: 10s
    send_batch_size: 1000

exporters:
  prometheusremotewrite:
    endpoint: http://prometheus:9090/api/v1/write

service:
  pipelines:
    metrics:
      receivers: [prometheus]
      processors: [phoenix_adaptive_filter, phoenix_topk, batch]
      exporters: [prometheusremotewrite]
```

## Phoenix Processors

### Adaptive Filter

Filters metrics based on dynamic thresholds:

```yaml
phoenix_adaptive_filter:
  cpu_threshold: 0.05        # Drop if CPU < 5%
  memory_threshold: 0.10     # Drop if memory < 10%
  disk_threshold: 1048576    # Drop if disk IO < 1MB/s
  network_threshold: 1048576 # Drop if network < 1MB/s
```

### Top-K Processor

Keeps only the top K metric series:

```yaml
phoenix_topk:
  k: 1000                    # Number of series to keep
  metric: cardinality        # Ranking metric
  window: 5m                 # Evaluation window
  by_labels: [service, job]  # Group by labels
```

## Building

```bash
# Build binary
make build

# Build with specific processors
make build-phoenix

# Build container
make docker

# Run tests
make test
```

## Performance

Benchmarks on 8-core machine:

| Processor | Throughput | Latency P99 | Memory |
|-----------|------------|-------------|---------|
| Baseline | 100K/sec | 1ms | 100MB |
| Adaptive Filter | 95K/sec | 1.2ms | 120MB |
| Top-K | 85K/sec | 2ms | 200MB |
| Combined | 80K/sec | 2.5ms | 250MB |

## Development

### Adding Processors

1. Implement processor in `internal/processor/`
2. Register in `components.go`
3. Add configuration struct
4. Write tests and benchmarks
5. Update builder manifest

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
make test-integration

# Load testing
make load-test
```