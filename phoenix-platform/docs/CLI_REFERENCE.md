# Phoenix CLI Reference

## Overview

The Phoenix CLI (`phoenix`) is a command-line tool for managing Phoenix Platform experiments, pipeline deployments, and configurations. It provides a comprehensive interface for all platform operations with support for multiple output formats and shell completion.

## Installation

### Quick Install

```bash
# Download and install the latest release
curl -sSL https://get.phoenix.example.com/cli | bash

# Or install to a specific location
curl -sSL https://get.phoenix.example.com/cli | bash -s -- --install-dir /usr/local/bin
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/phoenix/platform
cd platform/phoenix-platform

# Build the CLI
make build-cli

# Install to system path
sudo mv bin/phoenix /usr/local/bin/

# Verify installation
phoenix version
```

### Package Managers

```bash
# macOS (Homebrew)
brew tap phoenix/tap
brew install phoenix-cli

# Linux (apt)
curl -sSL https://apt.phoenix.example.com/pubkey.gpg | sudo apt-key add -
echo "deb https://apt.phoenix.example.com stable main" | sudo tee /etc/apt/sources.list.d/phoenix.list
sudo apt update && sudo apt install phoenix-cli

# Linux (yum/dnf)
sudo yum-config-manager --add-repo https://rpm.phoenix.example.com/phoenix.repo
sudo yum install phoenix-cli
```

## Configuration

The CLI stores configuration in `~/.phoenix/config.yaml`.

### Initial Setup

```bash
# Set the API endpoint
phoenix config set api_url https://api.phoenix.example.com

# Set default namespace
phoenix config set default_namespace production

# Enable debug output
phoenix config set debug true
```

### Configuration Commands

```bash
# List all configuration
phoenix config list

# Get specific value
phoenix config get api_url

# Set value
phoenix config set key value

# Reset to defaults
phoenix config reset

# Show config file location
phoenix config path
```

### Environment Variables

All configuration can be overridden with environment variables:

```bash
export PHOENIX_API_URL=https://api.phoenix.example.com
export PHOENIX_API_TOKEN=your-token
export PHOENIX_DEFAULT_NAMESPACE=production
export PHOENIX_OUTPUT_FORMAT=json
```

## Authentication

### Login Methods

```bash
# Interactive login (recommended)
phoenix auth login

# Login with credentials
phoenix auth login --username user@example.com --password yourpass

# Login with environment variables
export PHOENIX_USERNAME=user@example.com
export PHOENIX_PASSWORD=yourpass
phoenix auth login

# Login with token directly
phoenix auth login --token your-jwt-token
```

### Managing Authentication

```bash
# Check authentication status
phoenix auth status

# Refresh token
phoenix auth refresh

# Logout
phoenix auth logout

# Show current user info
phoenix auth whoami
```

## Command Reference

### Global Flags

All commands support these global flags:

```bash
--api-url string       Override API URL
--token string         Override auth token
--output string        Output format: table, json, yaml (default "table")
--no-color            Disable colored output
--debug               Enable debug output
--help                Show help for command
```

### Experiment Commands

#### Create Experiment

```bash
phoenix experiment create [flags]

Flags:
  --name string              Experiment name (required)
  --namespace string         Kubernetes namespace (required)
  --pipeline-a string        Baseline pipeline template (required)
  --pipeline-b string        Candidate pipeline template (required)
  --traffic-split string     Traffic split ratio "A/B" (required)
  --duration string          Experiment duration (e.g., "2h", "30m")
  --selector string          Node selector (e.g., "app=webserver")
  --min-cost-reduction int   Minimum cost reduction percentage
  --max-data-loss float      Maximum acceptable data loss percentage
  --critical-processes list  Comma-separated list of critical processes
  --metadata string          JSON metadata to attach to experiment
  --dry-run                  Validate without creating

Examples:
  # Basic experiment
  phoenix experiment create \
    --name "cost-optimization" \
    --namespace "production" \
    --pipeline-a "process-baseline-v1" \
    --pipeline-b "process-intelligent-v1" \
    --traffic-split "50/50" \
    --duration "2h" \
    --selector "app=webserver"

  # With success criteria
  phoenix experiment create \
    --name "database-optimization" \
    --namespace "staging" \
    --pipeline-a "process-baseline-v1" \
    --pipeline-b "process-topk-v1" \
    --traffic-split "80/20" \
    --duration "4h" \
    --selector "role=database" \
    --min-cost-reduction 30 \
    --max-data-loss 1.5 \
    --critical-processes "postgres,mysql,redis"

  # With metadata
  phoenix experiment create \
    --name "feature-test" \
    --namespace "dev" \
    --pipeline-a "baseline" \
    --pipeline-b "experimental" \
    --traffic-split "10/90" \
    --duration "30m" \
    --metadata '{"team":"platform","ticket":"JIRA-123"}'
```

#### List Experiments

```bash
phoenix experiment list [flags]

Flags:
  --namespace string     Filter by namespace
  --status string        Filter by status: running, completed, failed, stopped
  --limit int           Maximum results to return (default 20)
  --offset int          Pagination offset
  --sort string         Sort by: created, updated, name (default "created")
  --reverse             Reverse sort order

Examples:
  # List all experiments
  phoenix experiment list

  # List running experiments in production
  phoenix experiment list --namespace production --status running

  # List with pagination
  phoenix experiment list --limit 10 --offset 20

  # List in JSON format
  phoenix experiment list --output json
```

#### Get Experiment Status

```bash
phoenix experiment status <experiment-id> [flags]

Flags:
  --follow              Follow status updates in real-time
  --interval duration   Update interval when following (default 5s)

Examples:
  # Get current status
  phoenix experiment status exp-123

  # Follow status updates
  phoenix experiment status exp-123 --follow

  # Custom update interval
  phoenix experiment status exp-123 --follow --interval 10s
```

#### Start Experiment

```bash
phoenix experiment start <experiment-id> [flags]

Flags:
  --force   Force start even if preconditions aren't met

Examples:
  phoenix experiment start exp-123
  phoenix experiment start exp-123 --force
```

#### Stop Experiment

```bash
phoenix experiment stop <experiment-id> [flags]

Flags:
  --reason string   Reason for stopping (required)
  --force          Force stop without cleanup

Examples:
  phoenix experiment stop exp-123 --reason "High error rate detected"
  phoenix experiment stop exp-123 --reason "Emergency stop" --force
```

#### Get Experiment Metrics

```bash
phoenix experiment metrics <experiment-id> [flags]

Flags:
  --interval string   Time interval: 1m, 5m, 1h, 1d (default "5m")
  --start string     Start time (RFC3339 or relative like "-2h")
  --end string       End time (RFC3339 or relative like "now")
  --metric strings   Specific metrics to retrieve

Examples:
  # Get current metrics
  phoenix experiment metrics exp-123

  # Get metrics for last 2 hours
  phoenix experiment metrics exp-123 --start -2h

  # Get specific metrics
  phoenix experiment metrics exp-123 --metric cost_reduction,data_loss

  # Export metrics as JSON
  phoenix experiment metrics exp-123 --output json > metrics.json
```

#### Promote Experiment

```bash
phoenix experiment promote <experiment-id> [flags]

Flags:
  --reason string         Reason for promotion (required)
  --rollout string       Rollout strategy: immediate, gradual (default "immediate")
  --rollout-duration     Duration for gradual rollout (default "1h")

Examples:
  # Immediate promotion
  phoenix experiment promote exp-123 --reason "Met all success criteria"

  # Gradual rollout
  phoenix experiment promote exp-123 \
    --reason "Successful test, gradual rollout" \
    --rollout gradual \
    --rollout-duration 4h
```

#### Export Experiment

```bash
phoenix experiment export <experiment-id> [flags]

Flags:
  --include-metrics   Include historical metrics in export
  --include-logs      Include relevant logs in export

Examples:
  # Export configuration
  phoenix experiment export exp-123 > experiment.yaml

  # Export with metrics
  phoenix experiment export exp-123 --include-metrics > full-export.yaml
```

### Pipeline Commands

#### Deploy Pipeline

```bash
phoenix pipeline deploy [flags]

Flags:
  --name string           Deployment name (required)
  --namespace string      Kubernetes namespace (required)
  --template string       Pipeline template (required)
  --config-override json  Configuration overrides as JSON
  --description string    Deployment description
  --dry-run              Validate without deploying

Examples:
  # Basic deployment
  phoenix pipeline deploy \
    --name "prod-intelligent" \
    --namespace "production" \
    --template "process-intelligent-v1" \
    --description "Production intelligent pipeline"

  # With configuration overrides
  phoenix pipeline deploy \
    --name "staging-optimized" \
    --namespace "staging" \
    --template "process-topk-v1" \
    --config-override '{"topk_count":50,"sampling_rate":0.1}'
```

#### List Pipeline Deployments

```bash
phoenix pipeline deployments list [flags]

Flags:
  --namespace string   Filter by namespace (required)
  --status string      Filter by status: active, suspended, rollback

Examples:
  phoenix pipeline deployments list --namespace production
  phoenix pipeline deployments list --namespace staging --status active
```

#### Get Deployment Status

```bash
phoenix pipeline deployment status <deployment-id> [flags]

Examples:
  phoenix pipeline deployment status dep-789
  phoenix pipeline deployment status dep-789 --output json
```

#### Update Deployment

```bash
phoenix pipeline deployment update <deployment-id> [flags]

Flags:
  --config-override json   New configuration as JSON
  --reason string         Reason for update (required)

Examples:
  phoenix pipeline deployment update dep-789 \
    --config-override '{"sampling_rate":0.05}' \
    --reason "Reduce sampling based on volume"
```

#### Deployment History

```bash
phoenix pipeline deployment history <deployment-id> [flags]

Flags:
  --limit int   Maximum entries to show (default 20)

Examples:
  phoenix pipeline deployment history dep-789
  phoenix pipeline deployment history dep-789 --limit 50
```

#### Rollback Deployment

```bash
phoenix pipeline deployment rollback <deployment-id> [flags]

Flags:
  --history-id string   History entry to rollback to (required)
  --reason string      Reason for rollback (required)

Examples:
  phoenix pipeline deployment rollback dep-789 \
    --history-id hist-1 \
    --reason "Performance regression in v2"
```

#### List Pipeline Templates

```bash
phoenix pipeline templates list [flags]

Examples:
  phoenix pipeline templates list
  phoenix pipeline templates list --output json
```

### Utility Commands

#### Version

```bash
phoenix version [flags]

Flags:
  --short   Show version only

Examples:
  phoenix version
  phoenix version --short
```

#### Completion

```bash
phoenix completion <shell> [flags]

Supported shells: bash, zsh, fish, powershell

Examples:
  # Bash
  phoenix completion bash > /etc/bash_completion.d/phoenix

  # Zsh
  phoenix completion zsh > "${fpath[1]}/_phoenix"

  # Fish
  phoenix completion fish > ~/.config/fish/completions/phoenix.fish

  # PowerShell
  phoenix completion powershell > phoenix.ps1
```

## Output Formats

### Table Format (Default)

```bash
phoenix experiment list
ID        NAME                    STATUS    NAMESPACE    CREATED
exp-123   cost-optimization      running   production   2h ago
exp-124   database-test         completed  staging      1d ago
```

### JSON Format

```bash
phoenix experiment list --output json
{
  "experiments": [
    {
      "id": "exp-123",
      "name": "cost-optimization",
      "status": "running",
      "namespace": "production",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### YAML Format

```bash
phoenix experiment list --output yaml
experiments:
  - id: exp-123
    name: cost-optimization
    status: running
    namespace: production
    created_at: "2024-01-15T10:00:00Z"
```

### Custom Columns

```bash
# Select specific columns for table output
phoenix experiment list --columns id,name,status

# Available columns vary by resource type
phoenix experiment list --columns help
```

## Advanced Usage

### Scripting and Automation

```bash
#!/bin/bash
# Automated experiment monitoring

EXPERIMENT_ID=$(phoenix experiment create \
  --name "automated-test-$(date +%s)" \
  --namespace "staging" \
  --pipeline-a "baseline" \
  --pipeline-b "candidate" \
  --traffic-split "50/50" \
  --duration "1h" \
  --output json | jq -r '.id')

echo "Created experiment: $EXPERIMENT_ID"

# Start experiment
phoenix experiment start "$EXPERIMENT_ID"

# Monitor until complete
while true; do
  STATUS=$(phoenix experiment status "$EXPERIMENT_ID" --output json | jq -r '.status')
  
  if [[ "$STATUS" == "completed" ]] || [[ "$STATUS" == "failed" ]]; then
    break
  fi
  
  sleep 30
done

# Get final metrics
METRICS=$(phoenix experiment metrics "$EXPERIMENT_ID" --output json)
COST_REDUCTION=$(echo "$METRICS" | jq -r '.summary.cost_reduction_percent')

# Auto-promote if successful
if [[ "$STATUS" == "completed" ]] && (( $(echo "$COST_REDUCTION > 20" | bc -l) )); then
  phoenix experiment promote "$EXPERIMENT_ID" --reason "Automated: Cost reduction > 20%"
fi
```

### CI/CD Integration

```bash
# GitHub Actions
- name: Run Phoenix Experiment
  env:
    PHOENIX_API_TOKEN: ${{ secrets.PHOENIX_TOKEN }}
  run: |
    phoenix experiment create \
      --name "ci-test-${{ github.run_id }}" \
      --namespace "ci" \
      --pipeline-a "baseline" \
      --pipeline-b "pr-${{ github.event.pull_request.number }}" \
      --traffic-split "10/90" \
      --duration "30m"

# Jenkins
sh """
  export PHOENIX_API_TOKEN=${PHOENIX_TOKEN}
  phoenix experiment create \
    --name "jenkins-${BUILD_ID}" \
    --namespace "ci" \
    --pipeline-a "baseline" \
    --pipeline-b "candidate" \
    --traffic-split "20/80" \
    --duration "1h" \
    --metadata '{"build_id":"${BUILD_ID}","job":"${JOB_NAME}"}'
"""

# GitLab CI
phoenix-experiment:
  script:
    - export PHOENIX_API_TOKEN=$PHOENIX_TOKEN
    - |
      phoenix experiment create \
        --name "gitlab-$CI_PIPELINE_ID" \
        --namespace "ci" \
        --pipeline-a "baseline" \
        --pipeline-b "candidate" \
        --traffic-split "30/70" \
        --duration "45m"
```

### Batch Operations

```bash
# Stop all failed experiments
phoenix experiment list --status failed --output json | \
  jq -r '.experiments[].id' | \
  xargs -I {} phoenix experiment stop {} --reason "Batch cleanup of failed experiments"

# Export all completed experiments
phoenix experiment list --status completed --output json | \
  jq -r '.experiments[].id' | \
  while read -r id; do
    phoenix experiment export "$id" > "exports/experiment-$id.yaml"
  done

# Update multiple deployments
for ns in production staging development; do
  phoenix pipeline deployments list --namespace "$ns" --output json | \
    jq -r '.deployments[] | select(.template == "process-baseline-v1") | .id' | \
    xargs -I {} phoenix pipeline deployment update {} \
      --config-override '{"memory_limit_mib":512}' \
      --reason "Standardize memory limits"
done
```

### Monitoring and Alerting

```bash
# Watch experiment status
watch -n 10 'phoenix experiment list --namespace production --status running'

# Alert on high data loss
while true; do
  phoenix experiment list --status running --output json | \
    jq -r '.experiments[].id' | \
    while read -r id; do
      DATA_LOSS=$(phoenix experiment metrics "$id" --output json | \
        jq -r '.summary.data_loss_percent // 0')
      
      if (( $(echo "$DATA_LOSS > 5" | bc -l) )); then
        # Send alert (example with Slack)
        curl -X POST $SLACK_WEBHOOK -d \
          "{\"text\":\"⚠️ High data loss ($DATA_LOSS%) in experiment $id\"}"
        
        # Auto-stop experiment
        phoenix experiment stop "$id" --reason "Automated: Data loss exceeded 5%"
      fi
    done
  
  sleep 60
done
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   ```bash
   # Check token validity
   phoenix auth status
   
   # Refresh token
   phoenix auth refresh
   
   # Re-login if needed
   phoenix auth login
   ```

2. **Connection Issues**
   ```bash
   # Verify API endpoint
   phoenix config get api_url
   
   # Test connection
   curl -I $(phoenix config get api_url)/health
   
   # Enable debug mode
   phoenix --debug experiment list
   ```

3. **Permission Errors**
   ```bash
   # Check current user permissions
   phoenix auth whoami
   
   # Verify namespace access
   phoenix experiment list --namespace <namespace>
   ```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# One-time debug
phoenix --debug experiment create ...

# Persistent debug mode
phoenix config set debug true

# Debug with API request/response logging
export PHOENIX_DEBUG=true
export PHOENIX_DEBUG_HTTP=true
phoenix experiment list
```

### Log Files

CLI logs are stored in:
- Linux/macOS: `~/.phoenix/logs/`
- Windows: `%APPDATA%\phoenix\logs\`

```bash
# View recent logs
tail -f ~/.phoenix/logs/phoenix.log

# Clean old logs
find ~/.phoenix/logs -name "*.log" -mtime +7 -delete
```

## Best Practices

1. **Use Configuration Files for Complex Commands**
   ```yaml
   # experiment.yaml
   name: production-optimization
   namespace: production
   pipeline_a: process-baseline-v1
   pipeline_b: process-intelligent-v1
   traffic_split: "80/20"
   duration: 4h
   selector: app=webserver
   success_criteria:
     min_cost_reduction: 25
     max_data_loss: 1.0
   ```
   
   ```bash
   phoenix experiment create -f experiment.yaml
   ```

2. **Always Specify Reasons for State Changes**
   ```bash
   # Good
   phoenix experiment stop exp-123 --reason "High error rate detected in candidate pipeline"
   
   # Bad
   phoenix experiment stop exp-123 --reason "stopped"
   ```

3. **Use Meaningful Names**
   ```bash
   # Good
   phoenix experiment create --name "webserver-topk-optimization-v2"
   
   # Bad
   phoenix experiment create --name "test123"
   ```

4. **Monitor Before Promoting**
   ```bash
   # Check metrics thoroughly
   phoenix experiment metrics exp-123
   phoenix experiment metrics exp-123 --interval 1h --start -24h
   
   # Then promote if successful
   phoenix experiment promote exp-123 --reason "24h test successful, 35% cost reduction"
   ```

5. **Use Dry-Run for Validation**
   ```bash
   # Validate before creating
   phoenix experiment create --dry-run ...
   phoenix pipeline deploy --dry-run ...
   ```

## Support

For help and support:

```bash
# Built-in help
phoenix help
phoenix experiment help
phoenix experiment create --help

# Version information
phoenix version

# Report issues
https://github.com/phoenix/platform/issues
```