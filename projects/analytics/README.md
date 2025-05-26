# Phoenix Analytics Service

Advanced analytics service for deep metrics analysis and cost optimization insights.

## Overview

The Analytics service provides:
- Cardinality analysis and trending
- Cost projection models
- Anomaly detection in metric patterns
- Optimization recommendations
- Historical analysis and reporting

## Features

- **Real-time Analysis**: Stream processing of metrics data
- **ML-Powered Insights**: Machine learning models for pattern detection
- **Cost Modeling**: Accurate cost projections based on cardinality
- **API Integration**: gRPC API for high-performance analytics

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

The service exposes a gRPC API defined in `api/analytics.proto`:

```protobuf
service AnalyticsService {
  rpc AnalyzeCardinality(CardinalityRequest) returns (CardinalityResponse);
  rpc ProjectCosts(CostProjectionRequest) returns (CostProjectionResponse);
  rpc GetOptimizationSuggestions(OptimizationRequest) returns (OptimizationResponse);
}
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50053` |
| `PROMETHEUS_URL` | Prometheus server | Required |
| `MODEL_PATH` | ML models directory | `/models` |
| `LOG_LEVEL` | Logging level | `info` |

## Architecture

```
┌─────────────────┐
│   gRPC API      │
├─────────────────┤
│ Analytics Engine│
├─────────────────┤
│   ML Models     │
├─────────────────┤
│ Prometheus Query│
└─────────────────┘
```

## Development

### Project Structure

```
analytics/
├── api/           # Protocol definitions
├── cmd/           # Service entrypoint
├── internal/
│   ├── server/   # gRPC server
│   └── service/  # Business logic
└── scripts/      # Build scripts
```

### Building

```bash
# Generate protobuf code
make generate

# Build binary
make build

# Build Docker image
make docker
```