# Shell Scripts Cleanup Plan

## Scripts to Keep (Essential)

### 1. Tools/Analyzers (Keep)
- `tools/analyzers/boundary-check.sh` - Critical for architecture validation
- `tools/analyzers/llm-safety-check.sh` - Important for AI safety checks

### 2. Configuration (Keep)
- `configs/production/tls/generate_certs.sh` - Needed for TLS certificate generation

### 3. Project-Specific Build/Test (Keep)
- `projects/dashboard/src/test/run-tests.sh` - Dashboard test runner
- `projects/phoenix-cli/cmd/build.sh` - CLI build script

## Scripts to Remove/Replace

### 1. Examples Directory (Remove)
- `examples/experiment-simulation.sh` → Move content to markdown documentation
- `examples/experiment-workflow.sh` → Move content to markdown documentation

### 2. Deployments/single-vm/scripts (Remove - Already replaced)
- `health-check.sh` → Replaced by `phoenix-monitor.sh`
- `restore.sh` → Replaced by `phoenix-restore.sh`
- `backup.sh` → Replaced by `phoenix-backup.sh`
- `backup-incremental.sh` → Replaced by `phoenix-backup.sh --incremental`
- `auto-scale-monitor.sh` → Move to Kubernetes HPA or separate monitoring tool
- `setup-single-vm.sh` → Replaced by `deploy-setup.sh`
- `validate-scaling.sh` → Integrate into `phoenix-validate.sh`
- `install-agent.sh` → Replaced by `deploy-agent.sh`

### 3. Tools/dev-env (Remove)
- `tools/dev-env/setup.sh` → Replaced by `dev-setup.sh`

### 4. Projects/phoenix-agent/deployments/systemd (Update)
- `install.sh` → Keep but update to reference new deploy-agent.sh

### 5. Tests (Keep but Update)
- `tests/e2e/run_e2e_tests.sh` → Keep, essential for E2E testing

## Action Summary

**Keep:** 6 scripts
- tools/analyzers/boundary-check.sh
- tools/analyzers/llm-safety-check.sh  
- configs/production/tls/generate_certs.sh
- projects/dashboard/src/test/run-tests.sh
- projects/phoenix-cli/cmd/build.sh
- tests/e2e/run_e2e_tests.sh

**Remove:** 11 scripts
- All deployments/single-vm/scripts/* (8 scripts)
- tools/dev-env/setup.sh (1 script)
- examples/*.sh (2 scripts)

**Update:** 1 script
- projects/phoenix-agent/deployments/systemd/install.sh