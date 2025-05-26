# Phoenix Platform

A modular observability cost optimization platform that reduces metrics cardinality by up to 90% while maintaining critical visibility.

## 📋 Current Status

- **Platform Status**: 65% PRD compliant, core functionality operational
- **Critical Issue**: LoadSim Operator needs fixing for experiments
- **Documentation**: See [PLATFORM_STATUS.md](./PLATFORM_STATUS.md) for details

## 🗺️ Documentation Navigation

### For Developers
- [Quick Start Guide](./QUICKSTART.md) - Get up and running in 5 minutes
- [Development Guide](./DEVELOPMENT_GUIDE.md) - Development workflow and standards
- [API Documentation](./docs/api/README.md) - API contracts and playground
- [Architecture Guide](./docs/architecture/PLATFORM_ARCHITECTURE.md) - System design

### For Operators
- [Operations Guide](./docs/operations/OPERATIONS_GUIDE_COMPLETE.md) - Deployment and operations
- [Monitoring Setup](./monitoring/README.md) - Prometheus and Grafana configuration
- [Helm Charts](./infrastructure/helm/phoenix/README.md) - Kubernetes deployment

### Key References
- [PRD Status](./PRD_STATUS.md) - Product requirements implementation status
- [Contributing](./CONTRIBUTING.md) - Contribution guidelines
- [Claude AI Guide](./CLAUDE.md) - AI assistant integration guide

## 🏗️ Monorepo Structure

```
phoenix/
├── packages/              # Shared packages
│   ├── go-common/        # Go utilities and interfaces
│   └── contracts/        # API contracts (proto, OpenAPI)
├── projects/             # Service implementations
│   ├── analytics/        # Analytics engine
│   ├── anomaly-detector/ # Anomaly detection service
│   ├── api/             # API gateway
│   ├── benchmark/       # Benchmarking service
│   ├── controller/      # Experiment controller
│   ├── generator/       # Configuration generator
│   ├── loadsim-operator/# Load simulation K8s operator
│   ├── pipeline-operator/# Pipeline management K8s operator
│   └── platform-api/    # Platform API service
├── infrastructure/      # Deployment configurations
│   ├── kubernetes/      # K8s manifests and CRDs
│   └── helm/           # Helm charts
├── monitoring/         # Observability configurations
│   ├── prometheus/     # Prometheus rules and config
│   └── grafana/       # Grafana dashboards
├── scripts/           # Operational scripts
└── tests/            # Integration and E2E tests
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Kubernetes cluster (optional)
- Make

### Local Development

1. **Clone and setup**:
   ```bash
   git clone https://github.com/phoenix/platform.git
   cd platform
   ./scripts/setup-dev-env.sh
   ```

2. **Start dependencies**:
   ```bash
   make -f Makefile.dev dev-up
   ```

3. **Run services**:
   ```bash
   # All services with hot reload
   goreman start
   
   # Or individual services
   make -f Makefile.dev run-api
   make -f Makefile.dev run-controller
   ```

4. **Access services**:
   - API: http://localhost:8080
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

### Kubernetes Deployment

1. **Development environment**:
   ```bash
   ./scripts/deploy-dev.sh
   ```

2. **Production with Helm**:
   ```bash
   helm install phoenix infrastructure/helm/phoenix/ \
     --namespace phoenix \
     --values values-prod.yaml
   ```

## 🏛️ Architecture

### Core Services
- **API Gateway**: REST/gRPC gateway for external access
- **Experiment Controller**: Manages A/B testing experiments
- **Config Generator**: Generates OpenTelemetry pipeline configurations
- **Analytics Engine**: Processes and analyzes metrics data
- **Anomaly Detector**: Detects unusual patterns in metrics

### Kubernetes Operators
- **Pipeline Operator**: Manages OTel collector DaemonSets
- **LoadSim Operator**: Orchestrates load testing jobs

### Shared Packages
- **go-common**: Shared Go libraries and interfaces
- **contracts**: API contracts and protobuf definitions

## 🔒 Monorepo Boundaries

This repository enforces strict modular boundaries:

1. **No cross-project imports**: Projects cannot import from other projects
2. **Shared code in packages**: All shared code must be in `/packages/*`
3. **Interface-based communication**: Services communicate through defined interfaces

Validate boundaries:
```bash
./scripts/validate-boundaries.sh
```

## 🧪 Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# E2E tests
make test-e2e

# All tests with coverage
make test-coverage
```

## 📊 Monitoring

The platform includes comprehensive observability:

- **Metrics**: Prometheus with custom recording rules
- **Dashboards**: Pre-built Grafana dashboards
- **Tracing**: OpenTelemetry integration (optional)
- **Logs**: Structured JSON logging

## 🛠️ Development

### Building
```bash
# Build all services
make build

# Build specific service
cd projects/api && go build ./...

# Build Docker images
make docker-build
```

### Code Quality
```bash
# Format code
make fmt

# Run linters
make lint

# Validate structure
make validate
```

### Git Hooks
Pre-commit hooks are automatically installed to:
- Validate monorepo boundaries
- Run linters
- Check code formatting

## 📚 Documentation

### Core Documentation
- [Architecture Overview](docs/architecture/PLATFORM_ARCHITECTURE.md)
- [Monorepo Boundaries](MONOREPO_BOUNDARIES.md)
- [Interface Contracts](docs/INTERFACE_CONTRACTS.md)
- [AI Assistant Guidelines](CLAUDE.md)

### Guides
- [Contributing Guide](CONTRIBUTING.md)
- [Team Onboarding](TEAM_ONBOARDING.md)
- [E2E Demo Guide](E2E_DEMO_GUIDE.md)

### Migration & Operations
- [Migration Summary](docs/migration/MIGRATION_SUMMARY_CONSOLIDATED.md)
- [Service Consolidation](docs/operations/SERVICE_CONSOLIDATION_PLAN.md)
- [Service Inventory](docs/generated/SERVICE_INVENTORY.md)

### Quick Reference
- [Consolidated Documentation Index](CONSOLIDATED_DOCUMENTATION.md)

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Ensure boundary validation passes (`./scripts/validate-boundaries.sh`)
4. Run tests (`make test`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🏷️ Version

Current version: v2.0.0 (Post-migration monorepo structure)

---

Built with ❤️ by the Phoenix Platform team