# Archived Scripts

This directory contains scripts that were archived during the Phoenix script streamlining process.

## Archive Date
May 27, 2025

## Reason for Archiving
These scripts were replaced by the new streamlined script set that provides:
- Better organization (dev-*, deploy-*, phoenix-* naming)
- Consolidated functionality
- Improved error handling
- Consistent patterns

## Archived Scripts Categories

### Demo Scripts
- demo-complete.sh
- demo-docker.sh
- demo-flow.sh
- demo-local.sh
- demo-phoenix.sh
- demo-simple.sh
- demo-ui-flow.sh
- demo-working.sh
- quick-demo.sh

**Replaced by**: Example directories and documentation

### Test Scripts
- e2e-test.sh
- test-coverage.sh
- test-e2e-local.sh
- test-e2e.sh
- test-integration.sh
- test-pushgateway-fix.sh
- test-system.sh

**Replaced by**: Make targets and phoenix-validate.sh

### Validation Scripts
- validate-all.sh
- validate-boundaries.sh
- validate-build.sh
- validate-builds.sh
- validate-integration.sh
- validate-monorepo.sh
- validate-mvp.sh

**Replaced by**: phoenix-validate.sh (comprehensive validation)

### Setup Scripts
- check-prerequisites.sh
- setup-dev-env.sh
- setup-workspace.sh
- standardize-services.sh

**Replaced by**: dev-setup.sh (comprehensive setup)

### Start/Run Scripts
- quick-start.sh
- run-e2e-complete.sh
- run-e2e-demo.sh
- run-local-demo.sh
- run-phoenix.sh
- start-demo-agents.sh
- start-phoenix-simple.sh
- start-phoenix-ui.sh
- start-services.sh

**Replaced by**: dev-start.sh, dev-stop.sh, deploy-*.sh

### Utility Scripts
- fix-composite-store.sh
- generate-proto.sh
- install-protoc.sh
- push-to-remote.sh

**Replaced by**: Make targets or integrated into other scripts

## Migration Guide

| Old Script | New Script/Command |
|------------|-------------------|
| setup-dev-env.sh | dev-setup.sh |
| run-phoenix.sh | dev-start.sh |
| start-phoenix-ui.sh | dev-start.sh |
| validate-all.sh | phoenix-validate.sh |
| test-*.sh | make test-all |
| demo-*.sh | See examples/ directory |

## Note
These scripts are kept for reference only. Please use the new streamlined scripts for all operations.