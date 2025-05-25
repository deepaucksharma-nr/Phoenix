# Phoenix Platform Documentation

This directory contains all documentation for the Phoenix Observability Platform.

## Documentation Structure

### Getting Started
- [Overview & Concepts](OVERVIEW_QUICK_START.md) - Introduction to Phoenix platform
- [Developer Quick Start](DEVELOPER_QUICK_START.md) - 5-minute developer onboarding
- [Architecture Overview](ARCHITECTURE.md) - System architecture and design

### Development
- [Development Guide](DEVELOPMENT.md) - Detailed development environment setup
- [Build and Run Guide](BUILD_AND_RUN.md) - Quick commands for building and running
- [Implementation Status](IMPLEMENTATION_STATUS.md) - Current development status

### Deployment  
- [Deployment Guide](DEPLOYMENT.md) - Production deployment procedures

### Reference Documentation
- [API Reference](API_REFERENCE.md) - Complete API documentation
- [Pipeline Configuration Guide](PIPELINE_GUIDE.md) - OpenTelemetry pipeline configuration
- [Process Simulator Reference](PROCESS_SIMULATOR_REFERENCE.md) - Process simulator documentation
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions

### Technical Specifications
All technical specifications have been moved to `technical-specs/` subdirectory:
- [Master Technical Specification](TECHNICAL_SPEC_MASTER.md) - Authoritative technical architecture
- [API Service Specification](TECHNICAL_SPEC_API_SERVICE.md) - Detailed API service implementation
- [Dashboard Specification](TECHNICAL_SPEC_DASHBOARD.md) - Frontend dashboard implementation
- [Experiment Controller Specification](TECHNICAL_SPEC_EXPERIMENT_CONTROLLER.md) - Controller service implementation
- [Pipeline Operator Specification](TECHNICAL_SPEC_PIPELINE_OPERATOR.md) - Kubernetes operator implementation
- [Process Simulator Specification](TECHNICAL_SPEC_PROCESS_SIMULATOR.md) - Process simulator implementation

### Planning & Status
Project planning documents are in `planning/` subdirectory:
- [Product Requirements Document](PRODUCT_REQUIREMENTS.md) - Detailed product requirements (v1.4)
- [Project Status](planning/PROJECT_STATUS.md) - Real-time implementation tracking
- [CLI Implementation Plan](planning/CLI_IMPLEMENTATION_PLAN.md) - Phoenix CLI development plan
- [Pipeline Deployment API Design](planning/PIPELINE_DEPLOYMENT_API_DESIGN.md) - Direct pipeline deployment design
- [UI Error Handling Enhancement](planning/UI_ERROR_HANDLING_ENHANCEMENT.md) - UI improvements plan
- [Experiment Overlap Detection](planning/EXPERIMENT_OVERLAP_DETECTION_DESIGN.md) - Overlap detection design

### Reviews & Analysis
Review documents are in `reviews/` subdirectory:
- [Documentation Review](reviews/PHOENIX_DOCUMENTATION_REVIEW.md) - Comprehensive documentation analysis
- [Review Summary](reviews/COMPREHENSIVE_REVIEW_SUMMARY.md) - Executive review summary

## Documentation Standards

### File Naming Conventions
- **User-facing guides**: UPPERCASE with underscores (e.g., `PIPELINE_GUIDE.md`, `TROUBLESHOOTING.md`)
- **Technical documents**: UPPERCASE with underscores (e.g., `API_REFERENCE.md`, `DEVELOPER_QUICK_START.md`)
- **Technical specifications**: UPPERCASE with underscores (e.g., `TECHNICAL_SPEC_MASTER.md`)
- **Product documents**: Product name + type + version (e.g., `PRODUCT_REQUIREMENTS.md`)

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