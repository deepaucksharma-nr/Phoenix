# Phoenix Platform Documentation

This directory contains all documentation for the Phoenix Observability Platform.

## Documentation Structure

### Getting Started
- [Architecture Overview](architecture.md) - System architecture and design
- [User Guide](user-guide.md) - How to use Phoenix for process metrics optimization
- [Development Guide](DEVELOPMENT.md) - Development setup and workflow
- [Deployment Guide](DEPLOYMENT.md) - Production deployment procedures

### Reference Documentation
- [API Reference](api-reference.md) - Complete API documentation
- [Pipeline Configuration Guide](pipeline-guide.md) - OpenTelemetry pipeline configuration
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

### Technical Specifications
- [Product Requirements Document](PRODUCT_REQUIREMENTS.md) - Detailed product requirements (v1.4)
- [Master Technical Specification](TECHNICAL_SPEC_MASTER.md) - Authoritative technical architecture
- [API Service Specification](TECHNICAL_SPEC_API_SERVICE.md) - Detailed API service implementation
- [Dashboard Specification](TECHNICAL_SPEC_DASHBOARD.md) - Frontend dashboard implementation
- [Experiment Controller Specification](TECHNICAL_SPEC_EXPERIMENT_CONTROLLER.md) - Controller service implementation
- [Pipeline Operator Specification](TECHNICAL_SPEC_PIPELINE_OPERATOR.md) - Kubernetes operator implementation

## Documentation Standards

### File Naming Conventions
- **User-facing guides**: lowercase with hyphens (e.g., `user-guide.md`, `api-reference.md`)
- **Technical specifications**: UPPERCASE with underscores (e.g., `TECHNICAL_SPEC_MASTER.md`)
- **Product documents**: Product name + type + version (e.g., `PHOENIX_PRD_v1.4.md`)

### Document Structure
Each document should include:
1. Title and metadata (version, status, last updated)
2. Table of contents
3. Clear sections with numbered headings
4. Code examples where appropriate
5. Links to related documents

### Maintenance
- Update documentation with each feature change
- Review quarterly for accuracy
- Version control all changes
- Keep examples current with latest API