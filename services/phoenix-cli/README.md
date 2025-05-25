# Phoenix CLI

The Phoenix CLI is a command-line interface for managing Phoenix Platform experiments and pipeline deployments.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/phoenix-platform/phoenix
cd phoenix/phoenix-platform

# Build the CLI
make build-cli

# The binary will be in build/phoenix
./build/phoenix version
```

### Install to PATH

```bash
# Copy to a directory in your PATH
sudo cp build/phoenix /usr/local/bin/

# Verify installation
phoenix version
```

## Configuration

The CLI stores configuration in `~/.phoenix/config.yaml`. You can also use environment variables:

- `PHOENIX_API_ENDPOINT`: API endpoint (default: http://localhost:8080)
- `PHOENIX_OUTPUT_FORMAT`: Default output format (table, json, yaml)

## Authentication

Before using the CLI, you need to authenticate:

```bash
# Login with username and password
phoenix auth login -u admin -p password

# Or login interactively
phoenix auth login

# Check authentication status
phoenix auth status

# Logout
phoenix auth logout
```

## Usage

### Experiments

#### Create an Experiment

```bash
# Basic experiment
phoenix experiment create \
  --name "reduce-cardinality" \
  --baseline process-baseline-v1 \
  --candidate process-topk-v1 \
  --target-selector "app=webserver" \
  --duration 1h

# With critical processes
phoenix experiment create \
  --name "priority-filter-test" \
  --baseline process-baseline-v1 \
  --candidate process-priority-filter-v1 \
  --target-selector "environment=production" \
  --critical-processes "nginx,postgres,redis"

# Check for overlaps
phoenix experiment create \
  --name "test-optimization" \
  --baseline process-baseline-v1 \
  --candidate process-adaptive-v1 \
  --target-selector "tier=frontend" \
  --check-overlap
```

#### List Experiments

```bash
# List all experiments
phoenix experiment list

# Filter by status
phoenix experiment list --status running

# Output as JSON
phoenix experiment list -o json
```

#### Monitor Experiments

```bash
# Get experiment status
phoenix experiment status exp-123

# Follow experiment progress
phoenix experiment status exp-123 --follow

# Get experiment metrics
phoenix experiment metrics exp-123
```

#### Control Experiments

```bash
# Start an experiment
phoenix experiment start exp-123

# Stop an experiment
phoenix experiment stop exp-123

# Promote winning variant
phoenix experiment promote exp-123 --variant candidate
```

### Pipelines (Coming Soon)

```bash
# List available pipelines
phoenix pipeline list

# Deploy a pipeline directly
phoenix pipeline deploy \
  --name process-topk-v1 \
  --selector "environment=production" \
  --param top_k=20

# List deployments
phoenix pipeline list-deployments

# Delete deployment
phoenix pipeline delete-deployment dep-123
```

## Output Formats

The CLI supports multiple output formats:

```bash
# Table format (default)
phoenix experiment list

# JSON format
phoenix experiment list -o json

# YAML format
phoenix experiment list -o yaml
```

## Global Flags

- `--api-endpoint`: Override API endpoint
- `-o, --output`: Output format (table, json, yaml)
- `-v, --verbose`: Enable verbose output
- `--config`: Config file path

## Examples

### Complete Experiment Workflow

```bash
# 1. Login
phoenix auth login

# 2. Create experiment
phoenix experiment create \
  --name "optimize-frontend" \
  --baseline process-baseline-v1 \
  --candidate process-topk-v1 \
  --target-selector "app=frontend,env=staging" \
  --duration 2h

# 3. Monitor progress
phoenix experiment status optimize-frontend --follow

# 4. Check metrics
phoenix experiment metrics optimize-frontend

# 5. Promote if successful
phoenix experiment promote optimize-frontend --variant candidate
```

### Batch Operations

```bash
# List all running experiments as JSON
phoenix experiment list --status running -o json | jq '.[] | .id'

# Stop all pending experiments
for exp in $(phoenix experiment list --status pending -o json | jq -r '.[] | .id'); do
  phoenix experiment stop $exp
done
```

## Troubleshooting

### Authentication Issues

```bash
# Check current auth status
phoenix auth status

# Re-authenticate if token expired
phoenix auth login

# Check API endpoint
phoenix auth status | grep Endpoint
```

### Connection Issues

```bash
# Test API connectivity
curl -s http://localhost:8080/health

# Use different endpoint
phoenix --api-endpoint https://phoenix.company.com experiment list
```

### Debug Mode

```bash
# Enable verbose output
phoenix -v experiment create ...

# Check config file
cat ~/.phoenix/config.yaml
```

## Contributing

See the main Phoenix Platform contributing guide for development setup and guidelines.