# NRDOT Integration Complete

This document summarizes the complete integration of New Relic Distribution of OpenTelemetry (NRDOT) into the Phoenix platform.

## Overview

The Phoenix platform now supports using NRDOT collector as an alternative to the standard OpenTelemetry collector. This integration enables advanced cardinality reduction features specific to New Relic's distribution.

## Integration Points

### 1. Agent Installation Script
**File**: `/deployments/single-vm/scripts/install-agent.sh`
- Added `install_nrdot_collector()` function
- Downloads NRDOT binary from GitHub releases
- Supports version specification via `NRDOT_VERSION` environment variable
- Validates New Relic license key requirement

### 2. Docker Images
**File**: `/Dockerfile.phoenix-agent`
- Downloads both OTel and NRDOT collectors
- Makes both binaries available at runtime
- Agent decides which to use based on configuration

### 3. Agent Configuration
**File**: `/projects/phoenix-agent/internal/config/config.go`
- Added NRDOT-specific fields:
  - `UseNRDOT`: Enable NRDOT collector
  - `NRLicenseKey`: New Relic license key
  - `NROTLPEndpoint`: New Relic OTLP endpoint
  - `CollectorType`: "otel" or "nrdot"

### 4. Agent Main
**File**: `/projects/phoenix-agent/cmd/phoenix-agent/main.go`
- Added command-line flags for NRDOT configuration
- Environment variable support for NRDOT settings

### 5. Collector Manager
**File**: `/projects/phoenix-agent/internal/supervisor/collector.go`
- Updated to detect and use appropriate collector binary
- Validates New Relic configuration when using NRDOT
- Passes NRDOT-specific environment variables
- Supports dynamic parameter passing via vars map

### 6. Supervisor
**File**: `/projects/phoenix-agent/internal/supervisor/supervisor.go`
- Updated `executeCollectorTask` to handle NRDOT parameters
- Extracts NRDOT configuration from task config
- Passes parameters as environment variables to collector manager
- Handles both "start" and "update" actions with NRDOT support

### 7. Pipeline Templates
**Files**: 
- `/configs/otel-templates/nrdot/baseline.yaml`
- `/configs/otel-templates/nrdot/cardinality-reduction.yaml`
- `/projects/phoenix-api/internal/services/pipeline_template_renderer.go`

Added NRDOT-specific templates:
- `nrdot-baseline`: Basic NRDOT configuration with New Relic exporter
- `nrdot-cardinality`: Advanced configuration with cardinality reduction processor

### 8. CLI Support
**File**: `/projects/phoenix-cli/cmd/experiment_create.go`
- Added NRDOT-specific flags:
  - `--use-nrdot`: Use NRDOT collector
  - `--nr-license-key`: New Relic license key
  - `--nr-otlp-endpoint`: Custom OTLP endpoint
  - `--max-cardinality`: Maximum metric cardinality
  - `--reduction-percentage`: Target reduction percentage

### 9. API Integration
**Files**:
- `/projects/phoenix-api/internal/api/experiments.go`
- `/projects/phoenix-api/internal/controller/experiment_controller.go`

Updated to:
- Accept NRDOT parameters in experiment creation
- Store parameters in experiment metadata
- Propagate parameters to task configuration
- Pass parameters through to agents

### 10. Documentation
**File**: `/docs/operations/nrdot-integration.md`
- Comprehensive guide for NRDOT integration
- Installation instructions
- Configuration examples
- Troubleshooting guide

## Parameter Flow

The NRDOT parameters flow through the system as follows:

1. **CLI** → User specifies NRDOT flags
2. **API** → Parameters stored in experiment metadata
3. **Controller** → Extracts parameters and adds to task config
4. **Task Queue** → Task contains NRDOT configuration
5. **Agent** → Polls task and extracts NRDOT parameters
6. **Supervisor** → Passes parameters to collector manager
7. **Collector Manager** → Starts NRDOT with proper environment variables

## Key Features

1. **Automatic Detection**: System automatically uses NRDOT when:
   - `collector_type` is set to "nrdot"
   - Pipeline variant contains "nrdot"
   - Agent is started with `--use-nrdot` flag

2. **License Key Validation**: Ensures New Relic license key is provided when using NRDOT

3. **Cardinality Reduction**: Supports New Relic's advanced cardinality reduction processor

4. **Flexible Configuration**: Parameters can be set via:
   - Environment variables
   - Command-line flags
   - Experiment parameters
   - Task configuration

## Example Usage

### CLI Example
```bash
phoenix-cli experiment create \
  --name "NRDOT Cardinality Test" \
  --baseline-pipeline "baseline" \
  --candidate-pipeline "nrdot-cardinality" \
  --use-nrdot \
  --nr-license-key "YOUR_LICENSE_KEY" \
  --max-cardinality 5000 \
  --reduction-percentage 80
```

### Docker Compose Example
```yaml
environment:
  - USE_NRDOT=true
  - NEW_RELIC_LICENSE_KEY=your-license-key
  - NEW_RELIC_OTLP_ENDPOINT=otlp.nr-data.net:4317
```

## Testing

To test the NRDOT integration:

1. Set up New Relic license key
2. Create an experiment with NRDOT pipeline
3. Verify NRDOT collector starts on agents
4. Check metrics in New Relic UI
5. Validate cardinality reduction

## Security Considerations

- License keys are passed as environment variables
- Never hardcode license keys in configuration files
- Use secure methods to provide license keys in production

## Future Enhancements

1. Support for additional NRDOT processors
2. Integration with New Relic's native APIs
3. Advanced cardinality analytics
4. Automated NRDOT version updates