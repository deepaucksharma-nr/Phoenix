# Changelog

All notable changes to the Phoenix Platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **BREAKING**: Removed all Kubernetes deployment support in favor of Docker Compose
- Simplified deployment model to single-VM with Docker Compose orchestration
- Agents now deployed via systemd instead of Kubernetes DaemonSets
- Replaced Kubernetes service discovery with file-based configuration
- Updated all documentation to reflect Docker Compose deployment model

### Added
- Single-VM deployment scripts and documentation
- Auto-scaling monitor script for Docker Compose deployments
- Comprehensive migration guide from Kubernetes to Docker Compose
- Backup and restore scripts for single-VM deployments
- Health check and validation scripts

### Removed
- All Kubernetes manifests and Helm charts
- Kubernetes-specific CLI commands and client code
- Kubernetes deployment directories from all projects
- Kubernetes targets from Makefiles
- References to kubectl, kubeconfig, and Kubernetes concepts

### Migration
- Users running on Kubernetes should follow [MIGRATION_FROM_KUBERNETES.md](MIGRATION_FROM_KUBERNETES.md)
- Database migration scripts are provided for data portability
- Zero-downtime migration path available with proper planning

## [0.9.0] - 2024-01-15

### Added
- WebSocket support for real-time updates
- A/B testing framework for safe pipeline rollouts
- Task queue implementation using PostgreSQL
- Agent authentication via X-Agent-Host-ID header
- Live cost monitoring dashboard
- Pipeline templates (Adaptive Filter, TopK, Hybrid)

### Changed
- Improved agent polling mechanism with 30-second timeout
- Enhanced experiment lifecycle management
- Better error handling and retry logic

### Fixed
- Agent connection stability issues
- Memory leaks in long-running experiments
- Dashboard WebSocket reconnection logic

## [0.8.0] - 2023-12-01

### Added
- Initial Phoenix Platform release
- Core API with experiment management
- Agent-based architecture
- Basic dashboard UI
- PostgreSQL backend
- Prometheus integration

[Unreleased]: https://github.com/phoenix/platform/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/phoenix/platform/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/phoenix/platform/releases/tag/v0.8.0