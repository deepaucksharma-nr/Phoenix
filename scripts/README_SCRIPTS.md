# Phoenix Scripts Guide

This directory contains streamlined scripts for Phoenix platform development and deployment.

## Script Organization

### Core Scripts

#### Local Development
- `dev-setup.sh` - One-time development environment setup
- `dev-start.sh` - Start all services for local development
- `dev-stop.sh` - Stop all development services
- `dev-reset.sh` - Reset development environment
- `dev-status.sh` - Check status of development services

#### Multi-VM Deployment
- `deploy-setup.sh` - Setup Phoenix on multiple VMs
- `deploy-control.sh` - Deploy control plane components
- `deploy-agent.sh` - Deploy agents to edge nodes
- `deploy-status.sh` - Check deployment status
- `deploy-upgrade.sh` - Upgrade Phoenix components

#### Common Operations
- `phoenix-validate.sh` - Run all validation checks
- `phoenix-backup.sh` - Backup Phoenix data
- `phoenix-restore.sh` - Restore from backup
- `phoenix-monitor.sh` - Monitor Phoenix health

## Usage

### Local Development Workflow

1. **Initial Setup** (one-time)
   ```bash
   ./scripts/dev-setup.sh
   ```

2. **Start Development**
   ```bash
   ./scripts/dev-start.sh
   ```

3. **Check Status**
   ```bash
   ./scripts/dev-status.sh
   ```

4. **Stop Services**
   ```bash
   ./scripts/dev-stop.sh
   ```

### Multi-VM Deployment Workflow

1. **Setup Control Plane**
   ```bash
   ./scripts/deploy-setup.sh --control-plane
   ./scripts/deploy-control.sh
   ```

2. **Deploy Agents**
   ```bash
   ./scripts/deploy-agent.sh --host edge-node-1
   ./scripts/deploy-agent.sh --host edge-node-2
   ```

3. **Monitor Deployment**
   ```bash
   ./scripts/deploy-status.sh
   ./scripts/phoenix-monitor.sh
   ```

## Script Categories

### Setup & Installation
- Development: `dev-setup.sh`
- Production: `deploy-setup.sh`, `deploy-control.sh`, `deploy-agent.sh`

### Service Management
- Development: `dev-start.sh`, `dev-stop.sh`, `dev-reset.sh`
- Production: `deploy-upgrade.sh`

### Monitoring & Validation
- `dev-status.sh` - Development status
- `deploy-status.sh` - Production status
- `phoenix-validate.sh` - Code validation
- `phoenix-monitor.sh` - Runtime monitoring

### Backup & Recovery
- `phoenix-backup.sh` - Create backups
- `phoenix-restore.sh` - Restore from backup

## Environment Variables

All scripts respect these environment variables:

```bash
# Development
PHOENIX_DEV_DIR=${PHOENIX_DEV_DIR:-$HOME/.phoenix-dev}
PHOENIX_LOG_LEVEL=${PHOENIX_LOG_LEVEL:-debug}

# Production
PHOENIX_CONTROL_HOST=${PHOENIX_CONTROL_HOST:-localhost}
PHOENIX_API_PORT=${PHOENIX_API_PORT:-8080}
PHOENIX_DATA_DIR=${PHOENIX_DATA_DIR:-/var/lib/phoenix}
```

## Best Practices

1. Always run `dev-setup.sh` before first use
2. Use `dev-status.sh` to verify services are running
3. Run `phoenix-validate.sh` before committing code
4. Use `phoenix-backup.sh` before major changes
5. Check logs in `$PHOENIX_DEV_DIR/logs` for troubleshooting