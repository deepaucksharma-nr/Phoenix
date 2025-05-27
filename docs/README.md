# Phoenix Platform Documentation

Welcome to the Phoenix Platform documentation! Phoenix is an observability cost optimization platform that reduces metrics cardinality by up to 70% while maintaining critical visibility.

## 📖 Documentation Index

### Getting Started
- **[Quick Start Guide](../QUICKSTART.md)** - Get Phoenix running in 5 minutes
- **[Core Concepts](getting-started/concepts.md)** - Understand key Phoenix concepts
- **[First Experiment](getting-started/first-experiment.md)** - Create your first cost optimization

### Architecture
- **[Architecture Overview](../ARCHITECTURE.md)** - High-level system architecture
- **[Platform Architecture](architecture/PLATFORM_ARCHITECTURE.md)** - Detailed platform design
- **[System Design](architecture/system-design.md)** - Component interactions
- **[Messaging Decision](architecture/MESSAGING_DECISION.md)** - Task queue design choices

### API Documentation
- **[API Documentation](api/README.md)** - API documentation index
- **[Phoenix API v2](api/PHOENIX_API_v2.md)** - Complete REST & WebSocket API
- **[REST API Reference](api/rest-api.md)** - HTTP endpoints
- **[WebSocket API](api/websocket-api.md)** - Real-time events
- **[Pipeline Validation](api/PIPELINE_VALIDATION_API.md)** - Pipeline validation API

### Operations
- **[Operations Index](operations/README.md)** - Operations documentation hub
- **[Operations Guide](operations/OPERATIONS_GUIDE_COMPLETE.md)** - Comprehensive ops manual
- **[Configuration](operations/configuration.md)** - All configuration options
- **[Docker Compose](operations/docker-compose.md)** - Container deployment
- **[NRDOT Integration](operations/nrdot-integration.md)** - New Relic collector setup
- **[NRDOT Troubleshooting](operations/nrdot-troubleshooting.md)** - NRDOT issues

### Development
- **[Development Guide](../DEVELOPMENT_GUIDE.md)** - Developer setup and workflows
- **[Contributing](../CONTRIBUTING.md)** - Contribution guidelines
- **[AI Assistant Guide](../CLAUDE.md)** - Guidelines for AI-assisted development

### Design & UX
- **[UX Design](design/UX_DESIGN.md)** - User experience documentation
- **[UX Design Review](design/ux-design-review.md)** - Design review outcomes
- **[UX Implementation Plan](design/ux-implementation-plan.md)** - Implementation roadmap
- **[UX Revolution Overview](design/ux-revolution-overview.md)** - UX transformation

### Deployment
- **[Single-VM Deployment](../deployments/single-vm/README.md)** - Production deployment guide
- **[Architecture Review](../deployments/single-vm/ARCHITECTURE_REVIEW.md)** - Deployment architecture
- **[Capacity Planning](../deployments/single-vm/docs/capacity-planning-template.md)** - Resource planning
- **[Scaling Decisions](../deployments/single-vm/docs/scaling-decision-tree.md)** - When to scale
- **[Troubleshooting](../deployments/single-vm/docs/troubleshooting.md)** - Common issues

### Additional Resources
- **[Load Simulation Profiles](LOAD_SIMULATION_PROFILES.md)** - Load testing configurations
- **[Changelog](../CHANGELOG.md)** - Version history
- **[License](../LICENSE)** - Apache 2.0 License

## 🚀 Quick Links

### For Users
1. Start with the [Quick Start Guide](../QUICKSTART.md)
2. Understand [Core Concepts](getting-started/concepts.md)
3. Create your [First Experiment](getting-started/first-experiment.md)
4. Explore the [Dashboard](design/UX_DESIGN.md)

### For Operators
1. Review the [Operations Guide](operations/OPERATIONS_GUIDE_COMPLETE.md)
2. Configure your [Deployment](operations/docker-compose.md)
3. Set up [Monitoring](operations/configuration.md#monitoring)
4. Plan for [Scaling](../deployments/single-vm/docs/scaling-decision-tree.md)

### For Developers
1. Set up your [Development Environment](../DEVELOPMENT_GUIDE.md)
2. Understand the [Architecture](../ARCHITECTURE.md)
3. Explore the [API Documentation](api/README.md)
4. Read [Contributing Guidelines](../CONTRIBUTING.md)

## 📊 Key Features

- **70% Cost Reduction** - Intelligent metrics filtering
- **Real-time Monitoring** - WebSocket-based live updates
- **A/B Testing** - Safe rollout with experiments
- **Dual Collector Support** - OpenTelemetry and NRDOT
- **Agent-Based** - Distributed, scalable architecture
- **Enterprise Ready** - Production-grade deployment

## 🔧 Technology Stack

- **Backend**: Go 1.21+
- **Frontend**: React 18, TypeScript, Vite
- **Database**: PostgreSQL 15+
- **Collectors**: OpenTelemetry, NRDOT
- **Container**: Docker & Docker Compose
- **Monitoring**: Prometheus & Grafana

## 📱 Component Documentation

### Core Components
- **[Phoenix API](../projects/phoenix-api/README.md)** - Central control plane
- **[Phoenix Agent](../projects/phoenix-agent/README.md)** - Distributed agents
- **[Phoenix CLI](../projects/phoenix-cli/README.md)** - Command-line interface
- **[Dashboard](../projects/dashboard/README.md)** - Web UI

### Shared Packages
- **[Common Interfaces](../pkg/common/interfaces/README.md)** - Shared interfaces
- **[Contracts](../pkg/contracts/README.md)** - API contracts

## 🌟 Getting Help

- **Documentation**: You're here!
- **Issues**: [GitHub Issues](https://github.com/phoenix/platform/issues)
- **Community**: [Discord](https://discord.gg/phoenix)
- **Email**: support@phoenix.io

## 📝 Documentation Standards

All documentation follows these principles:
- **Clear**: Simple language, avoid jargon
- **Complete**: Cover all features and edge cases
- **Current**: Keep synchronized with code
- **Consistent**: Use standard formatting
- **Accessible**: Easy navigation and search

## 🔄 Recent Updates

- Added NRDOT (New Relic) collector support
- Enhanced WebSocket real-time capabilities
- Improved single-VM deployment documentation
- Added comprehensive troubleshooting guides
- Updated API to v2 with better structure

---

*Documentation Version: 2.0.0 | Last Updated: 2024*