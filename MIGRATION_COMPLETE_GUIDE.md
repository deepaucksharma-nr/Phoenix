# Phoenix Platform Complete Migration Guide

> **Version**: 2.0  
> **Last Updated**: January 2024  
> **Status**: Ready for Execution

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Migration Overview](#migration-overview)
3. [Quick Start Guide](#quick-start-guide)
4. [Architecture & Service Mapping](#architecture--service-mapping)
5. [Migration Framework](#migration-framework)
6. [Phase-by-Phase Execution](#phase-by-phase-execution)
7. [Multi-Agent Coordination](#multi-agent-coordination)
8. [Validation & Testing](#validation--testing)
9. [Troubleshooting & Recovery](#troubleshooting--recovery)
10. [Post-Migration Tasks](#post-migration-tasks)

---

## Executive Summary

This guide consolidates all Phoenix Platform migration documentation into a single, comprehensive resource. The migration transforms the legacy OLD_IMPLEMENTATION into a modern monorepo architecture with:

- **100% Project Independence**: Each service maintains its own lifecycle
- **70% Code Reuse**: Shared packages reduce duplication
- **Bulletproof Execution**: State tracking, rollback, and multi-agent coordination
- **Zero Downtime**: Blue-green deployment capability
- **Complete Validation**: Every step verified before proceeding

**Timeline**: 4-6 weeks | **Risk Level**: Medium | **Team Required**: 2-3 engineers

---

## Migration Overview

### Current State (OLD_IMPLEMENTATION)
```
OLD_IMPLEMENTATION/
├── phoenix-platform/      # Core platform services
├── apps/                  # Standalone applications
├── services/              # Supporting microservices
├── configs/               # Configuration files
└── docker-compose.yaml    # Development environment
```

### Target State (New Monorepo)
```
phoenix-vnext/
├── services/              # All microservices
├── packages/              # Shared code packages
├── infrastructure/        # Deployment configs
├── monitoring/            # Monitoring setup
├── tools/                 # Development tools
└── tests/                 # Cross-service tests
```

### Key Benefits
1. **Unified Development**: Single setup for entire platform
2. **Optimized Builds**: Turborepo for fast, cached builds
3. **Better Testing**: Comprehensive test infrastructure
4. **Improved Security**: Centralized security scanning
5. **Easier Onboarding**: Clear structure and documentation

---

## Quick Start Guide

### Prerequisites
```bash
# Check prerequisites
./scripts/migration/pre-flight-checks.sh

# Expected output: All checks PASSED
```

### Initialize Migration
```bash
# Set your agent ID
export AGENT_ID="agent-$(hostname)-$$"

# Initialize migration
./scripts/migration/migration-controller.sh init

# Check status
./scripts/migration/migration-controller.sh status
```

### Run Migration

#### Option 1: Complete Migration (Recommended for single agent)
```bash
./scripts/migration/migration-controller.sh run-all
```

#### Option 2: Phase-by-Phase (Recommended for multiple agents)
```bash
# Phase 0: Foundation (Single agent only)
./scripts/migration/migration-controller.sh run-phase phase-0-foundation

# Phase 1: Packages (Can be parallelized)
./scripts/migration/migration-controller.sh run-phase phase-1-packages

# Continue through all phases...
```

### Monitor Progress
```bash
# In a separate terminal
./scripts/migration/migration-controller.sh monitor
```

---

## Architecture & Service Mapping

### Service Migration Map

#### Core Services (Phase 2)
| OLD_IMPLEMENTATION | New Location | Description |
|-------------------|--------------|-------------|
| phoenix-platform/cmd/api-gateway | services/api-gateway | Main API gateway |
| phoenix-platform/cmd/control-service | services/control-service | Control plane service |
| phoenix-platform/cmd/controller | services/controller | Experiment controller |
| phoenix-platform/dashboard | services/dashboard | React web dashboard |

#### Supporting Services (Phase 3)
| OLD_IMPLEMENTATION | New Location | Description |
|-------------------|--------------|-------------|
| apps/anomaly-detector | services/anomaly-detector | Anomaly detection |
| apps/control-actuator-go | services/control-actuator-go | PID controller |
| services/analytics | services/analytics | Analytics engine |
| services/benchmark | services/benchmark | Performance benchmark |
| services/collector | services/collector | Metrics collector |

#### Operators & Tools (Phase 4)
| OLD_IMPLEMENTATION | New Location | Description |
|-------------------|--------------|-------------|
| phoenix-platform/cmd/phoenix-cli | services/phoenix-cli | CLI tool |
| phoenix-platform/operators/loadsim | services/loadsim-operator | Load simulation |
| phoenix-platform/operators/pipeline | services/pipeline-operator | Pipeline operator |

### Package Structure
```
packages/
├── common/               # TypeScript utilities
├── contracts/            # API contracts (OpenAPI, Proto)
├── config/               # Shared configurations
├── go-common/            # Go shared libraries
│   ├── auth/            # Authentication
│   ├── database/        # Database utilities
│   ├── telemetry/       # Metrics, tracing, logging
│   └── errors/          # Error handling
└── ui-components/        # Shared React components
```

### Infrastructure Layout
```
infrastructure/
├── docker/              # Docker configurations
├── kubernetes/          # K8s manifests
│   ├── base/           # Base resources
│   └── overlays/       # Environment-specific
├── helm/               # Helm charts
└── terraform/          # Infrastructure as code
```

---

## Migration Framework

### Core Principles

1. **Idempotency**: Every operation can be run multiple times safely
2. **Atomicity**: Each phase either completes fully or rolls back
3. **Validation**: Every step is validated before proceeding
4. **State Tracking**: Complete migration state is tracked and recoverable
5. **Multi-Agent Safe**: Coordination mechanisms prevent conflicts
6. **Rollback Ready**: Any phase can be rolled back safely

### State Management

All migration state is tracked in the `.migration/` directory:

```yaml
.migration/
├── state.yaml           # Overall migration state
├── phases/              # Phase-specific state
├── locks/               # Active resource locks
├── validations/         # Validation results
├── rollback-points/     # Git tags for rollback
├── reports/             # Generated reports
└── migration.log        # Complete activity log
```

### Lock System

Prevents conflicts between multiple agents:

```bash
# Acquire lock (automatic in migration controller)
acquire_lock "resource-name" "agent-id"

# Locks auto-expire after 1 hour to prevent deadlocks
# Manual cleanup if needed:
find .migration/locks -name "*.lock" -mmin +60 -delete
```

### Idempotent Operations

All operations check before executing:
- Directories: Created only if not exist
- Files: Marked with `.migrated` to prevent re-processing
- Git operations: Check for existing commits/tags
- Services: Verify not already migrated

---

## Phase-by-Phase Execution

### Phase 0: Foundation Setup
**Requirements**: Clean git state, no running services  
**Parallel**: No (exclusive)

Creates base directory structure and workspace configuration:
```bash
./scripts/migration/migration-controller.sh run-phase phase-0-foundation
```

**Validates**:
- Directory structure exists
- package.json and turbo.json created
- Makefile infrastructure ready
- Git configuration proper

### Phase 1: Shared Packages Migration
**Requirements**: Phase 0 complete  
**Parallel**: Yes

Migrates shared code to packages directory:
```bash
./scripts/migration/migration-controller.sh run-phase phase-1-packages
```

**Components**:
- Go common libraries → packages/go-common
- Contracts (OpenAPI, Proto) → packages/contracts
- TypeScript packages → packages/common, packages/ui-components

### Phase 2: Core Services Migration
**Requirements**: Phase 1 complete  
**Parallel**: Yes (per service)

Migrates critical platform services:
```bash
./scripts/migration/migration-controller.sh run-phase phase-2-core-services
```

**Services**: api-gateway, control-service, controller, dashboard

### Phase 3: Supporting Services Migration
**Requirements**: Phase 2 complete  
**Parallel**: Yes (per service)

Migrates auxiliary services:
```bash
./scripts/migration/migration-controller.sh run-phase phase-3-support-services
```

### Phase 4: Operators and Tools Migration
**Requirements**: Phase 3 complete  
**Parallel**: Yes

Migrates Kubernetes operators and CLI tools:
```bash
./scripts/migration/migration-controller.sh run-phase phase-4-operators
```

### Phase 5: Infrastructure Migration
**Requirements**: Phase 4 complete  
**Parallel**: No

Migrates deployment configurations:
```bash
./scripts/migration/migration-controller.sh run-phase phase-5-infrastructure
```

**Includes**:
- Kubernetes manifests
- Helm charts
- Docker configurations
- Monitoring setup

### Phase 6: Integration Testing
**Requirements**: Phase 5 complete  
**Parallel**: No

Validates all services work together:
```bash
./scripts/migration/migration-controller.sh run-phase phase-6-integration
```

### Phase 7: Finalization
**Requirements**: Phase 6 complete  
**Parallel**: No

Cleanup and final validation:
```bash
./scripts/migration/migration-controller.sh run-phase phase-7-finalization
```

---

## Multi-Agent Coordination

### Agent Setup
```bash
# Each agent needs unique ID
export AGENT_ID="agent-team1-01"

# Agent registration happens automatically
./scripts/migration/migration-controller.sh run-phase <phase>
```

### Parallel Execution Example

**Terminal 1 (Agent 1)**:
```bash
export AGENT_ID="agent-1"
./scripts/migration/migration-controller.sh run-phase phase-1-packages
# Works on go-common packages
```

**Terminal 2 (Agent 2)**:
```bash
export AGENT_ID="agent-2"
./scripts/migration/migration-controller.sh run-phase phase-1-packages
# Works on contracts packages
```

**Terminal 3 (Monitor)**:
```bash
./scripts/migration/migration-controller.sh monitor
# Shows real-time progress
```

### Coordination Rules
1. Only one agent can hold a lock on a resource
2. Phases marked `can_parallelize: true` support multiple agents
3. Components within a phase are distributed automatically
4. Stale locks (>1 hour) are automatically cleaned

---

## Validation & Testing

### Pre-Migration Validation
- [x] Git working directory clean
- [x] Required tools installed (Go 1.21+, Node 18+, Docker)
- [x] Sufficient disk space (10GB+)
- [x] No running Phoenix services
- [x] OLD_IMPLEMENTATION directory exists

### Per-Phase Validation
Each phase includes specific validations:

**Build Validation**:
```bash
cd services/<service> && make build
```

**Test Validation**:
```bash
cd services/<service> && make test
```

**Docker Validation**:
```bash
cd services/<service> && make docker-build
```

### Integration Testing
After all services migrated:
```bash
# Start services
docker-compose up -d

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e
```

### Validation Reports
Generated automatically after each phase:
```bash
# View phase report
cat .migration/reports/phase-<id>.md

# Generate final report
./scripts/migration/migration-controller.sh report
```

---

## Troubleshooting & Recovery

### Common Issues

#### "Failed to acquire lock"
```bash
# Check for stale locks
ls -la .migration/locks/

# Remove stale locks (older than 1 hour)
find .migration/locks -name "*.lock" -mmin +60 -delete
```

#### "Phase validation failed"
```bash
# Check validation details
cat .migration/validations/<phase>/*.yaml

# Fix issues and retry
./scripts/migration/migration-controller.sh run-phase <phase>
```

#### "Dependency not satisfied"
```bash
# Check phase status
./scripts/migration/migration-controller.sh status

# Run missing prerequisite phases first
```

### Rollback Procedures

#### Rollback Single Phase
```bash
./scripts/migration/migration-controller.sh rollback <phase-id>
```

#### Force Rollback (if normal rollback fails)
```bash
# Find rollback tag
git tag | grep rollback-<phase>

# Hard reset
git reset --hard <tag>

# Clean migration state
rm -rf .migration/phases/<phase>*
```

### Recovery from Failure

1. **Check Status**:
```bash
./scripts/migration/migration-controller.sh status
```

2. **Review Logs**:
```bash
tail -100 .migration/migration.log
```

3. **Fix Issues**:
- Address the specific error
- Clean up partial state if needed

4. **Resume Migration**:
```bash
# Migration automatically resumes from failure point
./scripts/migration/migration-controller.sh run-phase <failed-phase>
```

---

## Post-Migration Tasks

### 1. Verification
```bash
# Verify no OLD_IMPLEMENTATION references
grep -r "OLD_IMPLEMENTATION" --exclude-dir=.git --exclude-dir=OLD_IMPLEMENTATION .

# Run full test suite
make test

# Check all services build
make build
```

### 2. Documentation Update
- Update README.md with new structure
- Update CI/CD documentation
- Update deployment guides

### 3. CI/CD Pipeline Update
- Update GitHub Actions workflows
- Update build pipelines
- Configure Turborepo caching

### 4. Team Communication
- Announce migration completion
- Share new development workflows
- Schedule training if needed

### 5. Cleanup (After Verification)
```bash
# Archive OLD_IMPLEMENTATION
tar -czf OLD_IMPLEMENTATION-backup-$(date +%Y%m%d).tar.gz OLD_IMPLEMENTATION/

# Remove OLD_IMPLEMENTATION (only after thorough testing)
rm -rf OLD_IMPLEMENTATION/

# Clean migration state
./scripts/migration/migration-controller.sh cleanup
```

### 6. Performance Baseline
```bash
# Run benchmarks
make benchmark

# Compare with pre-migration baseline
```

---

## Success Criteria

### Technical Success
- ✅ All services migrated and building
- ✅ All tests passing (unit, integration, E2E)
- ✅ No performance regression (within 5%)
- ✅ Docker images building successfully
- ✅ Kubernetes manifests valid
- ✅ No references to OLD_IMPLEMENTATION

### Process Success
- ✅ Migration completed within timeline
- ✅ No data loss or corruption
- ✅ Zero downtime during migration
- ✅ All team members trained on new structure
- ✅ Documentation updated and accurate

### Operational Success
- ✅ Development velocity improved
- ✅ Build times reduced with Turborepo
- ✅ Easier to onboard new developers
- ✅ Simplified deployment process
- ✅ Better code reuse across services

---

## Appendix: Key Commands Reference

### Migration Control
```bash
# Initialize
./scripts/migration/migration-controller.sh init

# Status
./scripts/migration/migration-controller.sh status

# Run phase
./scripts/migration/migration-controller.sh run-phase <phase-id>

# Run all
./scripts/migration/migration-controller.sh run-all

# Monitor
./scripts/migration/migration-controller.sh monitor

# Rollback
./scripts/migration/migration-controller.sh rollback <phase-id>

# Report
./scripts/migration/migration-controller.sh report

# Cleanup
./scripts/migration/migration-controller.sh cleanup
```

### Development
```bash
# Setup workspace
make setup

# Start services
make dev-up

# Build everything
make build

# Run all tests
make test

# Stop services
make dev-down
```

### Service-Specific
```bash
# Build service
cd services/<name> && make build

# Test service
cd services/<name> && make test

# Run service
cd services/<name> && make run

# Build Docker image
cd services/<name> && make docker
```

---

*This guide consolidates all Phoenix Platform migration documentation. For questions or issues, consult the migration logs at `.migration/migration.log` or run `./scripts/migration/migration-controller.sh status` for current state.*