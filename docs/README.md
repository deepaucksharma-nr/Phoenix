# Phoenix Platform Documentation

Welcome to the Phoenix Platform documentation. Phoenix is an observability cost optimization platform that reduces metrics cardinality by up to 70% while maintaining critical visibility.

## Documentation by Role

### 👤 For Users

**Getting Started**
- [Quick Start Guide](../QUICKSTART.md) - Get Phoenix running in 5 minutes
- [Core Concepts](getting-started/concepts.md) - Understand Phoenix terminology
- [First Experiment](getting-started/first-experiment.md) - Create your first optimization

**User Guides**
- [Dashboard Overview](user-guide/dashboard.md) - Navigate the Phoenix UI
- [Managing Experiments](user-guide/experiments.md) - Run A/B tests safely
- [Pipeline Configuration](user-guide/pipelines.md) - Configure optimization strategies
- [Monitoring Setup](user-guide/monitoring.md) - Track performance and savings

### 👨‍💻 For Developers

**Development**
- [Development Setup](../DEVELOPMENT_GUIDE.md) - Set up your local environment
- [Project Structure](developer-guide/project-structure.md) - Understand the codebase
- [API Reference](api/rest-api.md) - REST API documentation
- [WebSocket API](api/websocket-api.md) - Real-time updates

**Contributing**
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute
- [Testing Guide](developer-guide/testing.md) - Write and run tests
- [Best Practices](developer-guide/best-practices.md) - Coding standards

### 🔧 For Operators

**Deployment**
- [Kubernetes Deployment](operations/deployment/kubernetes.md) - Production K8s setup
- [Docker Compose](operations/deployment/docker-compose.md) - Quick deployment
- [Single VM Setup](operations/deployment/single-vm.md) - Standalone installation

**Operations**
- [Configuration Reference](operations/configuration.md) - All configuration options
- [Production Guide](operations/OPERATIONS_GUIDE_COMPLETE.md) - Production best practices
- [Scaling Strategies](operations/scaling.md) - Handle growth
- [Backup & Recovery](operations/backup-recovery.md) - Data protection

### 🏛️ For Architects

**Architecture**
- [System Design](architecture/system-design.md) - High-level architecture
- [Component Overview](architecture/components.md) - Service descriptions
- [Data Flow](architecture/data-flow.md) - Request and data paths
- [Security Architecture](architecture/security.md) - Security model

**Design Decisions**
- [Messaging Architecture](architecture/MESSAGING_DECISION.md) - Why task polling
- [Platform Architecture](architecture/PLATFORM_ARCHITECTURE.md) - Detailed design

## Quick Links

### Essential Documentation
- 🚀 [Quick Start](../QUICKSTART.md)
- 📖 [Architecture Overview](architecture/system-design.md)
- 🔌 [API Reference](api/rest-api.md)
- 🛠️ [Development Guide](../DEVELOPMENT_GUIDE.md)

### Common Tasks
- [Create an Experiment](user-guide/experiments.md#creating-experiments)
- [Deploy a Pipeline](user-guide/pipelines.md#deployment)
- [Monitor Cost Savings](user-guide/monitoring.md#cost-tracking)
- [Troubleshoot Issues](user-guide/troubleshooting.md)

### Tutorials
- [Reduce Metrics by 70%](tutorials/reduce-cardinality.md)
- [Build Custom Pipelines](tutorials/custom-pipelines.md)
- [Integrate with Existing Systems](tutorials/integration-guide.md)

## Component Documentation

### Core Services
- [Phoenix API](../projects/phoenix-api/README.md) - Central control plane
- [Phoenix Agent](../projects/phoenix-agent/README.md) - Distributed agents
- [Phoenix CLI](../projects/phoenix-cli/README.md) - Command-line interface
- [Dashboard](../projects/dashboard/README.md) - Web UI

### Shared Packages
- [Authentication](../pkg/auth/) - JWT-based auth
- [Database](../pkg/database/) - Database abstractions
- [Telemetry](../pkg/telemetry/) - Logging and metrics
- [Contracts](../pkg/contracts/) - API contracts

## Documentation Standards

All documentation follows these principles:
- **Clear and Concise** - Direct, actionable content
- **Example-Driven** - Real-world examples included
- **Up-to-Date** - Reflects current implementation
- **Cross-Referenced** - Links to related content

## Getting Help

- 💬 [Discord Community](https://discord.gg/phoenix) - Ask questions
- 🐛 [Issue Tracker](https://github.com/phoenix/platform/issues) - Report bugs
- 📧 [Email Support](mailto:support@phoenix.io) - Enterprise support

## Version

This documentation is for Phoenix Platform v3.0.0. For other versions, see the [releases page](https://github.com/phoenix/platform/releases).