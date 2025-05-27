# Phoenix Platform

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](go.mod)
[![Documentation](https://img.shields.io/badge/docs-latest-green)](docs/)

Phoenix is an observability cost optimization platform that reduces metrics cardinality by up to 70% while maintaining critical visibility. Using intelligent pipeline optimization and agent-based architecture, Phoenix helps organizations cut observability costs without sacrificing insights.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/phoenix/platform.git
cd platform

# Run the setup script for single-VM deployment
./deployments/single-vm/scripts/setup-single-vm.sh

# Start Phoenix with Docker Compose
cd deployments/single-vm
docker-compose up -d

# Access the dashboard
open http://localhost:3000

# Install agents on target hosts
curl -sSL http://localhost:8080/install-agent.sh | sudo bash
```

See [QUICKSTART.md](QUICKSTART.md) for detailed setup instructions.

## 📋 Key Features

- **70% Cost Reduction** - Intelligent metrics filtering reduces cardinality without losing critical data
- **Real-time Monitoring** - WebSocket-based live updates for experiments and metrics
- **Agent-Based Architecture** - Distributed agents with task polling and heartbeat monitoring
- **A/B Testing Framework** - Safe rollout with baseline vs candidate pipeline comparison
- **Pipeline Templates** - Pre-built optimization strategies (Adaptive Filter, TopK, Hybrid)
- **Dual Collector Support** - Choose between OpenTelemetry or NRDOT (New Relic Distribution)
- **NRDOT Integration** - Advanced cardinality reduction with New Relic's optimized collectors
- **Enterprise Ready** - PostgreSQL storage, TLS support, comprehensive monitoring

## 🏗️ Architecture Overview

Phoenix uses a modular monorepo structure with agent-based architecture:

```
┌─────────────────┐         ┌─────────────────┐
│   Phoenix API   │◄────────┤   Dashboard     │
│  (Port 8080)    │         │   (React 18)    │
│  REST + WS      │         └─────────────────┘
└────────┬────────┘
         │ Task Queue (PostgreSQL)
         │ Long-polling (30s timeout)
    ┌────▼────┐
    │ Phoenix │────► OTel/NRDOT ────► Observability
    │ Agents  │      Collector       │ Backends
    └─────────┘                      └─────────────
```

### Core Components

- **Phoenix API** - Central control plane with REST/WebSocket APIs
- **Phoenix Agent** - Distributed agents deploying pipeline configurations
- **Phoenix CLI** - Command-line interface for operations
- **Dashboard** - React-based UI for monitoring and management

## 📚 Documentation

### Getting Started
- [Quick Start Guide](QUICKSTART.md) - Get running in 5 minutes
- [Concepts & Terminology](docs/getting-started/concepts.md) - Core concepts
- [First Experiment](docs/getting-started/first-experiment.md) - Create your first cost optimization

### Architecture & Design
- [System Architecture](docs/architecture/system-design.md) - High-level design
- [Component Details](docs/architecture/components.md) - Service descriptions
- [Data Flow](docs/architecture/data-flow.md) - Request and data paths
- [Security Model](docs/architecture/security.md) - Authentication and authorization

### API Reference
- [REST API](docs/api/rest-api.md) - HTTP endpoints reference
- [WebSocket API](docs/api/websocket-api.md) - Real-time updates
- [OpenAPI Spec](docs/api/openapi.yaml) - Machine-readable specification

### User Guides
- [Dashboard Guide](docs/user-guide/dashboard.md) - UI walkthrough
- [Managing Experiments](docs/user-guide/experiments.md) - A/B testing workflows
- [Pipeline Management](docs/user-guide/pipelines.md) - Configuration and deployment
- [Monitoring & Alerts](docs/user-guide/monitoring.md) - Observability setup

### Developer Resources
- [Development Setup](DEVELOPMENT_GUIDE.md) - Local environment setup
- [Project Structure](docs/developer-guide/project-structure.md) - Codebase organization
- [Testing Guide](docs/developer-guide/testing.md) - Test strategies and execution
- [Contributing](CONTRIBUTING.md) - Contribution guidelines

### Operations
- [Single-VM Deployment](deployments/single-vm/README.md) - Production-ready single VM setup
- [Docker Compose Guide](docs/operations/docker-compose.md) - Container orchestration
- [Configuration Reference](docs/operations/configuration.md) - All config options
- [Production Guide](docs/operations/OPERATIONS_GUIDE_COMPLETE.md) - Production deployment
- [NRDOT Integration](docs/operations/nrdot-integration.md) - New Relic collector setup
- [Scaling & Performance](docs/operations/scaling.md) - Scaling strategies
- [Migration from Kubernetes](MIGRATION_FROM_KUBERNETES.md) - Migration guide

### Tutorials
- [Reduce Cardinality by 70%](docs/tutorials/reduce-cardinality.md)
- [Building Custom Pipelines](docs/tutorials/custom-pipelines.md)
- [Integration Guide](docs/tutorials/integration-guide.md)

## 🔌 Collector Support

Phoenix supports multiple telemetry collectors:

### OpenTelemetry Collector (Default)
- Industry-standard collector with wide ecosystem support
- Configurable processors for basic cardinality reduction
- Compatible with any OTLP-compliant backend

### NRDOT (New Relic Distribution of OpenTelemetry)
- Advanced cardinality reduction with New Relic's algorithms
- Up to 80% reduction in metric volume
- Automatic preservation of critical metrics
- Native integration with New Relic platform

**Quick Start with NRDOT:**
```bash
# Using environment variables
export USE_NRDOT=true
export NEW_RELIC_LICENSE_KEY=your-license-key

# Or using CLI
phoenix-cli experiment create \
  --name "NRDOT Test" \
  --use-nrdot \
  --nr-license-key $NEW_RELIC_LICENSE_KEY \
  --candidate-pipeline nrdot-cardinality
```

See [NRDOT Integration Guide](docs/operations/nrdot-integration.md) for detailed setup.

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+ (primary database)

### Build from Source

```bash
# Install dependencies
make setup

# Build all components
make build

# Run tests
make test

# Start development environment
make dev-up
```

See [Development Guide](DEVELOPMENT_GUIDE.md) for detailed instructions.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Good First Issues

- Check out issues labeled [`good first issue`](https://github.com/phoenix/platform/issues?q=label%3A%22good+first+issue%22)
- Join our [Discord community](https://discord.gg/phoenix) for help

## 📄 License

This project is licensed under the Apache License 2.0 - see [LICENSE](LICENSE) for details.

## 🔗 Links

- [Documentation](docs/)
- [Issue Tracker](https://github.com/phoenix/platform/issues)
- [Discord Community](https://discord.gg/phoenix)
- [Blog](https://phoenix.io/blog)