# Phoenix CLI

## Overview

Command-line interface for the Phoenix Platform, providing easy access to experiment management, pipeline deployment, and real-time monitoring capabilities.

## Architecture

The CLI communicates with Phoenix API v2 endpoints:

```
┌──────────────────────┐
│    Phoenix CLI       │
│  (Go-based tool)     │
└─────────┬─────────────┘
           │
      REST API v2
           │
    ┌──────▼──────┐
    │ Phoenix API │
    │ (Port 8080) │
    └─────────────┘
```

## Development

### Prerequisites

- Go 1.21+
- Make
- Phoenix API running (default: http://localhost:8080)

### Setup

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build the CLI
make build

# Install to PATH
make install
```

### Running Locally

```bash
# Run directly
go run cmd/phoenix-cli/main.go --help

# Or use the built binary
./bin/phoenix --help

# Set API endpoint
export PHOENIX_API_URL=http://localhost:8080
```

## Configuration

Configuration can be set via:

1. **Environment Variables**:
```bash
export PHOENIX_API_URL=http://localhost:8080
export PHOENIX_TOKEN=your-auth-token
```

2. **Config File** (`~/.phoenix/config.yaml`):
```yaml
api_url: http://localhost:8080
token: your-auth-token
default_output: table

# Collector preferences (optional)
collector_type: nrdot  # or "otel"
nrdot_endpoint: https://otlp.nr-data.net:4317
```

3. **Command Flags**:
```bash
phoenix --api-url http://custom:8080 experiments list
```

## Command Reference

### Experiments

```bash
# List experiments
phoenix experiment list

# Create new experiment (A/B test)
phoenix experiment create \
  --name "Reduce API costs" \
  --hosts "group:prod-api" \
  --baseline standard \
  --candidate adaptive-filter-v1 \
  --duration 24h

# Start experiment
phoenix experiment start exp-123

# Get real-time status
phoenix experiment status exp-123 --watch

# View KPIs (70% cost reduction)
phoenix experiment metrics exp-123

# Promote candidate to production
phoenix experiment promote exp-123
```

### Pipelines

```bash
# List available templates
phoenix pipeline list-templates

# Deploy pipeline
phoenix pipeline deploy \
  --template adaptive-filter-v1 \
  --hosts "group:prod-api" \
  --variant candidate

# Get deployment status
phoenix pipeline status dep-456
```

### Agents

```bash
# List all agents
phoenix agent list

# Get agent details
phoenix agent get agent-001

# View agent tasks
phoenix agent tasks agent-001
```

### Real-time Monitoring

```bash
# Watch experiment progress
phoenix experiment watch exp-123

# Monitor cost flow
phoenix metrics cost-flow --live

# View cardinality breakdown
phoenix metrics cardinality --service api-gateway
```

## Testing

```bash
# Run unit tests
make test

# Run integration tests (requires API)
PHOENIX_API_URL=http://localhost:8080 make test-integration

# Run with coverage
make test-coverage

# Run specific test
go test -v ./internal/client -run TestExperiments
```

## Deployment

```bash
# Build for all platforms
make build-all

# Build Docker image
docker build -t phoenix/cli .

# Run in container
docker run --rm phoenix/cli experiment list

# Release new version
make release VERSION=v1.2.0
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

# Custom columns
phoenix experiment list -o custom -c id,name,status,savings_percent
```

## Shell Completion

```bash
# Bash
phoenix completion bash > /etc/bash_completion.d/phoenix

# Zsh
phoenix completion zsh > "${fpath[1]}/_phoenix"

# Fish
phoenix completion fish > ~/.config/fish/completions/phoenix.fish
```

## Examples

### Quick Experiment Setup
```bash
# Create and start experiment in one command
phoenix experiment quick-start \
  --name "Q1 Cost Reduction" \
  --template adaptive-filter-v1 \
  --hosts "env:production" \
  --auto-promote-threshold 50
```

### Batch Operations
```bash
# Stop all running experiments
phoenix experiment list -o json | \
  jq -r '.[] | select(.status=="running") | .id' | \
  xargs -I {} phoenix experiment stop {}
```

### NRDOT Integration
```bash
# Deploy pipeline with NRDOT collector
phoenix pipeline deploy \
  --name "nrdot-optimized" \
  --template adaptive-filter-v2 \
  --collector-type nrdot \
  --param importance_threshold=0.8

# Check NRDOT-enabled agents
phoenix agent list --filter collector_type=nrdot
```

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md)
