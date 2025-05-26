# Phoenix Platform Documentation

Welcome to the Phoenix Platform documentation. This guide will help you understand, deploy, and operate Phoenix.

## 📚 Documentation Structure

### Getting Started
- [Quick Start Guide](../QUICKSTART.md) - Get Phoenix running in 5 minutes
- [Development Guide](../DEVELOPMENT_GUIDE.md) - Set up your development environment
- [Architecture Overview](architecture/PLATFORM_ARCHITECTURE.md) - Understand the system design

### API & Integration
- [API Documentation](api/) - RESTful API reference
- [API v2 Specification](api/PHOENIX_API_v2.md) - Detailed API documentation
- [CLI Documentation](../projects/phoenix-cli/README.md) - Command-line interface

### Architecture & Design
- [Platform Architecture](architecture/PLATFORM_ARCHITECTURE.md) - System architecture
- [Component Interactions](architecture/PHOENIX_COMPONENT_INTERACTIONS.mmd) - Service communication
- [Data Model](architecture/PHOENIX_DATA_MODEL.mmd) - Database schema
- [Network Topology](architecture/PHOENIX_NETWORK_TOPOLOGY.mmd) - Network design
- [Messaging Decision](architecture/MESSAGING_DECISION.md) - Architecture decisions

### User Experience
- [UX Design Overview](design/UX_DESIGN.md) - Design philosophy and principles
- [UX Revolution](design/ux-revolution-overview.md) - Revolutionary UX features
- [Implementation Plan](design/ux-implementation-plan.md) - UX implementation roadmap
- [Design Review](design/ux-design-review.md) - Design decisions and learnings

### Operations
- [Operations Guide](operations/OPERATIONS_GUIDE_COMPLETE.md) - Production deployment and management
- [Single VM Deployment](../deployments/single-vm/README.md) - Simple deployment option
- [Troubleshooting Guide](../deployments/single-vm/docs/troubleshooting.md) - Common issues
- [Workflows](../deployments/single-vm/docs/workflows.md) - Operational workflows

### Configuration
- [Control Configuration](../configs/control/README.md) - Control plane settings
- [Monitoring Configuration](../configs/monitoring/README.md) - Prometheus & Grafana
- [OTel Configuration](../configs/otel/README.md) - OpenTelemetry setup
- [Production Configuration](../configs/production/README.md) - Production settings

### Development
- [Contributing Guidelines](../CONTRIBUTING.md) - How to contribute
- [Shared Interfaces](../pkg/common/interfaces/README.md) - Common interfaces
- [Contract Definitions](../pkg/contracts/README.md) - API contracts
- [Test Documentation](../tests/e2e/README.md) - Testing guidelines

## 🔍 Quick Links by Role

### For Operators
1. [Quick Start](../QUICKSTART.md)
2. [Operations Guide](operations/OPERATIONS_GUIDE_COMPLETE.md)
3. [Troubleshooting](../deployments/single-vm/docs/troubleshooting.md)
4. [Monitoring Setup](../configs/monitoring/README.md)

### For Developers
1. [Development Guide](../DEVELOPMENT_GUIDE.md)
2. [API Documentation](api/)
3. [Architecture](architecture/PLATFORM_ARCHITECTURE.md)
4. [Contributing](../CONTRIBUTING.md)

### For Users
1. [Dashboard Guide](design/ux-revolution-overview.md)
2. [CLI Reference](../projects/phoenix-cli/README.md)
3. [Experiment Workflows](../deployments/single-vm/docs/workflows.md)

## 📖 Documentation Standards

### File Naming
- Use lowercase with hyphens for file names
- Use `.md` extension for all documentation
- Keep names descriptive but concise

### Content Structure
1. Start with a clear title and purpose
2. Include a table of contents for long documents
3. Use code examples liberally
4. Add diagrams where helpful
5. Include troubleshooting sections

### Markdown Standards
- Use ATX-style headers (`#`, `##`, etc.)
- Indent code blocks with language hints
- Use tables for structured data
- Include links to related documents

## 🤝 Contributing to Docs

To contribute to documentation:

1. Follow the standards above
2. Check for broken links
3. Ensure examples are tested
4. Update the index when adding new docs
5. Submit PR with clear description

See [Contributing Guidelines](../CONTRIBUTING.md) for more details.