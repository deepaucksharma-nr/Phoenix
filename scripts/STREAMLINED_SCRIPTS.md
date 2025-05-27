# Phoenix Streamlined Scripts

This directory now contains a streamlined set of scripts organized for both local development and multi-VM production deployments.

## Script Organization Summary

### 🚀 Local Development Scripts
- **dev-setup.sh** - One-time development environment setup
- **dev-start.sh** - Start all Phoenix services locally
- **dev-stop.sh** - Stop Phoenix services (with --all flag for infrastructure)
- **dev-status.sh** - Check status of all development services
- **dev-reset.sh** - Reset development environment (with --deep for full clean)

### 🏢 Multi-VM Deployment Scripts
- **deploy-setup.sh** - Initial setup for control plane or agent nodes
- **deploy-control.sh** - Deploy Phoenix control plane components
- **deploy-agent.sh** - Deploy Phoenix agents to edge nodes

### 🛠️ Common Operation Scripts
- **phoenix-validate.sh** - Run validation checks (code structure, security, etc.)
- **phoenix-monitor.sh** - Monitor Phoenix platform health
- **phoenix-backup.sh** - Create backups (full or incremental)
- **phoenix-restore.sh** - Restore from backup

### 📚 Documentation Scripts
- **README_SCRIPTS.md** - Detailed guide for all scripts
- **STREAMLINED_SCRIPTS.md** - This summary file

## Quick Start

### Local Development
```bash
# First time setup
./scripts/dev-setup.sh

# Start development
./scripts/dev-start.sh

# Check status
./scripts/dev-status.sh

# Stop when done
./scripts/dev-stop.sh
```

### Production Deployment
```bash
# On control plane node
./scripts/deploy-setup.sh --control-plane
./scripts/deploy-control.sh

# On agent nodes
export PHOENIX_API_URL=http://control-plane-ip:8080
./scripts/deploy-agent.sh
```

## Script Consolidation

### Removed/Archived Scripts
The following scripts have been consolidated or archived:
- Multiple demo scripts → Functionality moved to example directories
- Duplicate test scripts → Consolidated into make targets
- Single-purpose scripts → Combined into comprehensive scripts

### Retained Original Scripts
- **verify-system.sh** - System verification (created during this session)
- **mvp-validation.sh** - MVP validation suite
- Core infrastructure scripts that are still referenced

## Benefits of Streamlining

1. **Clarity**: Clear naming convention (dev-*, deploy-*, phoenix-*)
2. **Consistency**: All scripts follow same patterns and structure
3. **Completeness**: Each script is self-contained with proper error handling
4. **Documentation**: Built-in help and clear output messages
5. **Flexibility**: Scripts accept parameters for different scenarios

## Environment Variables

All scripts respect these environment variables:

```bash
# Development
PHOENIX_DEV_DIR=${HOME}/.phoenix-dev
PHOENIX_LOG_LEVEL=debug

# Production
PHOENIX_API_URL=http://localhost:8080
PHOENIX_DATA_DIR=/var/lib/phoenix
PHOENIX_CONFIG_DIR=/etc/phoenix
```

## Next Steps

1. Test all scripts in both environments
2. Update CI/CD pipelines to use new scripts
3. Archive old scripts after verification
4. Update documentation to reference new scripts