# Documentation Migration Plan

## Overview
This document tracks the migration of documentation from OLD_IMPLEMENTATION to the new monorepo structure, noting differences and updates needed.

## Migration Status

### ✅ Already Migrated/Exists in New Structure
- **CLAUDE.md** - Already exists at root (updated for monorepo)
- **CONTRIBUTING.md** - Already exists at root
- **README.md** - Already exists at root (new version)
- **Architecture documentation** - Partially exists in PHOENIX_PLATFORM_ARCHITECTURE.md

### 📋 To Be Migrated

#### Root Level Documentation
- [ ] **CORE_REQUIREMENTS.md** → Merge into docs/requirements/
- [ ] **DEVELOPMENT_IMPROVEMENTS.md** → Merge into docs/guides/developer/
- [ ] **INFRASTRUCTURE.md** → Move to docs/infrastructure/
- [ ] **PROJECT_STRUCTURE.md** → Update and merge into PHOENIX_PLATFORM_ARCHITECTURE.md

#### API Documentation
- [ ] **docs/API.md** → docs/api/README.md
- [ ] **docs/api/rest.md** → docs/api/rest/README.md
- [ ] **docs/api/playground.md** → docs/api/playground/README.md
- [ ] **docs/assets/openapi.yaml** → docs/api/openapi/platform-api.yaml
- [ ] **phoenix-platform/docs/API_CONTRACT_SPECIFICATIONS.md** → docs/api/contracts/
- [ ] **phoenix-platform/docs/API_REFERENCE.md** → docs/api/reference/

#### Architecture Documentation
- [ ] **docs/ARCHITECTURE.md** → docs/architecture/README.md
- [ ] **docs/MONOREPO_STRUCTURE.md** → Already covered in PHOENIX_PLATFORM_ARCHITECTURE.md
- [ ] **phoenix-platform/docs/architecture.md** → Merge with docs/architecture/
- [ ] **phoenix-platform/docs/INTERFACE_ARCHITECTURE.md** → docs/architecture/interfaces/
- [ ] **phoenix-platform/docs/DATA_FLOW_AND_STATE_MANAGEMENT.md** → docs/architecture/data-flow/

#### Technical Specifications
- [ ] **phoenix-platform/docs/TECHNICAL_SPEC_*.md** → docs/architecture/services/

#### Development Guides
- [ ] **phoenix-platform/docs/DEVELOPMENT.md** → docs/guides/developer/development.md
- [ ] **phoenix-platform/docs/DEVELOPER_QUICK_START.md** → docs/guides/developer/quick-start.md
- [ ] **phoenix-platform/docs/BUILD_AND_RUN.md** → docs/guides/developer/build-and-run.md

#### Operations Documentation
- [ ] **docs/CLOUD_DEPLOYMENT.md** → docs/guides/operator/deployment/cloud.md
- [ ] **phoenix-platform/docs/DEPLOYMENT.md** → docs/guides/operator/deployment/
- [ ] **phoenix-platform/docs/OPERATIONAL_RUNBOOKS.md** → docs/runbooks/
- [ ] **phoenix-platform/docs/MONITORING_AND_ALERTING_STRATEGY.md** → docs/guides/operator/monitoring/

#### User Documentation
- [ ] **phoenix-platform/docs/USER_GUIDE.md** → docs/guides/user/
- [ ] **phoenix-platform/docs/PIPELINE_GUIDE.md** → docs/guides/user/pipelines/
- [ ] **phoenix-platform/docs/CLI_REFERENCE.md** → docs/reference/cli/

#### ADRs (Architecture Decision Records)
- [ ] All ADRs → docs/architecture/decisions/

### 🔄 Needs Update for New Implementation

1. **Service Locations**: OLD_IMPLEMENTATION had services in different structure
   - Update all service paths to reflect projects/ structure
   - Remove references to phoenix-platform/ subdirectory

2. **Import Paths**: Update all Go import paths
   - From: `github.com/phoenix/phoenix-platform/...`
   - To: `github.com/phoenix/platform/...`

3. **Docker & Kubernetes**: Update deployment configs
   - New docker-compose structure
   - Updated Kubernetes manifests in deployments/

4. **CI/CD**: Update pipeline documentation
   - New GitHub Actions structure
   - Monorepo-aware CI/CD

5. **Testing**: Update test documentation
   - New test structure in tests/
   - Integration test updates

## Migration Process

1. **Phase 1**: Core Documentation (Priority: High)
   - API documentation
   - Architecture documentation
   - Development guides

2. **Phase 2**: Operational Documentation (Priority: Medium)
   - Deployment guides
   - Runbooks
   - Monitoring documentation

3. **Phase 3**: Reference Documentation (Priority: Low)
   - ADRs
   - Technical specifications
   - Planning documents

## Key Differences in New Implementation

1. **Monorepo Structure**
   - All services under projects/
   - Shared packages in pkg/
   - Common tooling in tools/

2. **Build System**
   - Unified Makefile system
   - Go workspace (go.work)
   - Shared build infrastructure

3. **Service Independence**
   - Each project is completely independent
   - No cross-project imports allowed
   - Boundary enforcement via tools

4. **Development Workflow**
   - Single repository setup
   - Unified CI/CD
   - Consistent tooling across projects

5. **Documentation Structure**
   - Centralized in docs/
   - Project-specific docs in projects/*/docs/
   - Automated documentation generation