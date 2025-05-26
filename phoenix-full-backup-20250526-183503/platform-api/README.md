# Platform API Service

Core REST/gRPC API service for Phoenix Platform.

## Overview

The Platform API service provides the main interface for:
- Experiment management
- Pipeline configuration
- Metrics collection and analysis
- User authentication and authorization

## Architecture

```
platform-api/
├── cmd/api/              # Application entrypoint
├── internal/             # Private application code
│   ├── api/             # API layer (HTTP/gRPC)
│   ├── domain/          # Business logic
│   └── infrastructure/  # External dependencies
├── migrations/          # Database migrations
├── docs/               # Documentation
└── tests/              # Test files
```

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Make

### Quick Start

```bash
# Install dependencies
make deps

# Run tests
make test

# Build the service
make build

# Run locally
make run
```

### Configuration

The service is configured via environment variables:

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/phoenix?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Server
API_HOST=0.0.0.0
API_PORT=8080

# Auth
JWT_SECRET=your-secret-key
```

### API Documentation

- REST API: http://localhost:8080/swagger
- gRPC: See `api/proto/` for service definitions

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Generate coverage report
make test-coverage
```

## Deployment

### Docker

```bash
# Build image
make docker

# Run container
docker run -p 8080:8080 platform-api
```

### Kubernetes

```bash
kubectl apply -f deployments/k8s/
```

## Contributing

See [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines.