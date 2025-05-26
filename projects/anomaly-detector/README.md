# Phoenix Anomaly Detector

Real-time anomaly detection service for identifying metric irregularities and system issues.

## Overview

The Anomaly Detector monitors metrics streams to identify:
- Sudden metric spikes or drops
- Unusual cardinality patterns
- System performance degradation
- Configuration drift
- Cost anomalies

## Features

- **Real-time Detection**: Sub-second anomaly identification
- **Multiple Algorithms**: Statistical, ML-based, and rule-based detection
- **Configurable Sensitivity**: Adjust detection thresholds per metric
- **Alert Integration**: Webhook and notification support

## Quick Start

```bash
# Install dependencies
go mod download

# Generate protobuf code
make generate

# Run service
make run

# Run tests
make test
```

## API

The service exposes a gRPC API:

```protobuf
service AnomalyService {
  rpc DetectAnomalies(stream MetricData) returns (stream AnomalyEvent);
  rpc GetAnomalyHistory(HistoryRequest) returns (HistoryResponse);
  rpc UpdateDetectionRules(RulesUpdate) returns (RulesResponse);
}
```

## Detection Algorithms

1. **Statistical Methods**
   - Z-score detection
   - Moving average deviation
   - Seasonal decomposition

2. **Machine Learning**
   - Isolation Forest
   - LSTM prediction
   - Clustering-based detection

3. **Rule-Based**
   - Threshold alerts
   - Rate of change
   - Pattern matching

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50054` |
| `PROMETHEUS_URL` | Prometheus server | Required |
| `DETECTION_INTERVAL` | Check interval | `30s` |
| `SENSITIVITY` | Global sensitivity | `medium` |
| `ALERT_WEBHOOK` | Alert endpoint | Optional |

## Architecture

```
┌─────────────────┐
│ Metrics Stream  │
├─────────────────┤
│ Detection Engine│
├─────────────────┤
│   Algorithms    │
├─────────────────┤
│ Alert Manager   │
└─────────────────┘
```

## Development

### Testing

```bash
# Unit tests
make test

# Integration tests with Prometheus
make test-integration

# Benchmark detection algorithms
make benchmark
```

### Adding New Algorithms

1. Implement the `Detector` interface
2. Register in `internal/service/registry.go`
3. Add configuration options
4. Write tests and benchmarks