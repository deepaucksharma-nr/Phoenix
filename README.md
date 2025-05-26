# Phoenix Platform

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](go.mod)
[![Documentation](https://img.shields.io/badge/docs-latest-green)](docs/)

Phoenix is an observability cost optimization platform that reduces metrics cardinality by up to 90% while maintaining critical visibility. Using intelligent pipeline optimization and a lean-core architecture, Phoenix helps organizations cut observability costs without sacrificing insights.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/phoenix/platform.git
cd phoenix

# Start Phoenix with Docker Compose
./scripts/run-phoenix.sh

# Access the dashboard
open http://localhost:3000
```

See [QUICKSTART.md](QUICKSTART.md) for detailed setup instructions.

## 📋 Key Features

- **90% Metrics Reduction** - Intelligent filtering reduces cardinality without data loss
- **Real-time Cost Analytics** - See savings as they happen
- **Zero-Config Agents** - Self-registering agents with automatic discovery
- **A/B Testing Framework** - Compare pipeline configurations with production traffic
- **Visual Pipeline Builder** - Drag-and-drop interface for creating optimization rules
- **Enterprise Security** - JWT auth, RBAC, and full audit logging

## 🏗️ Architecture

Phoenix uses a lean-core architecture with three main components:

```
┌─────────────────┐         ┌─────────────────┐
│   Phoenix API   │◄────────┤   Dashboard     │
│  (Control Plane)│         │   (React UI)    │
└────────┬────────┘         └─────────────────┘
         │ Task Queue
    ┌────▼────┐
    │ Phoenix │────► Pushgateway ────► Prometheus
    │ Agents  │
    └─────────┘
```

See [Architecture Documentation](docs/architecture/PLATFORM_ARCHITECTURE.md) for details.

## 📚 Documentation

- [Architecture Overview](docs/architecture/PLATFORM_ARCHITECTURE.md)
- [Development Guide](DEVELOPMENT_GUIDE.md)
- [API Documentation](docs/api/)
- [Operations Guide](docs/operations/OPERATIONS_GUIDE_COMPLETE.md)
- [Contributing Guidelines](CONTRIBUTING.md)

## 🛠️ Development

### Prerequisites

- Go 1.24+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+

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