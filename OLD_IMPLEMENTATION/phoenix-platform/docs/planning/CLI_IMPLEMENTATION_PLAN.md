# Phoenix CLI Implementation Plan

## Overview
The Phoenix CLI is a critical missing component identified in the functional review. This document outlines the implementation plan to deliver a comprehensive CLI that enables scripting, automation, and power user workflows.

## Architecture

### CLI Framework
- **Framework**: Cobra (github.com/spf13/cobra) - Industry standard for Go CLIs
- **Configuration**: Viper (github.com/spf13/viper) - For config file support
- **Output Formats**: Table (default), JSON, YAML
- **Authentication**: JWT token stored in ~/.phoenix/config

### Command Structure
```
phoenix
├── auth
│   ├── login      # Authenticate with Phoenix API
│   └── logout     # Clear stored credentials
├── experiment
│   ├── create     # Create new experiment
│   ├── list       # List experiments
│   ├── get        # Get experiment details
│   ├── start      # Start an experiment
│   ├── stop       # Stop running experiment
│   ├── status     # Get experiment status
│   ├── metrics    # View experiment metrics
│   ├── compare    # Compare baseline vs candidate
│   └── promote    # Promote winning variant
├── pipeline
│   ├── list       # List available pipelines
│   ├── get        # Get pipeline configuration
│   ├── validate   # Validate custom pipeline
│   ├── deploy     # Deploy pipeline (non-experiment)
│   └── delete     # Remove deployed pipeline
├── config
│   ├── set        # Set configuration values
│   ├── get        # Get configuration values
│   └── list       # List all configuration
└── version        # Show CLI version
```

## Implementation Phases

### Phase 1: Core Framework (Week 1)
1. **Setup CLI Structure**
   - Create cmd/phoenix-cli directory
   - Initialize Cobra app with root command
   - Add version command
   - Setup configuration management with Viper

2. **Authentication Module**
   - Implement login command (prompt for credentials)
   - Store JWT token securely
   - Add token refresh logic
   - Implement logout command

3. **API Client Library**
   - Create pkg/client/phoenix package
   - Wrap all REST API calls
   - Handle authentication headers
   - Implement retry logic and error handling

### Phase 2: Experiment Commands (Week 1-2)
1. **Experiment Management**
   ```go
   // Example usage:
   // phoenix experiment create --name "test-optimization" \
   //   --baseline process-baseline-v1 \
   //   --candidate process-topk-v1 \
   //   --duration 1h \
   //   --target-selector "app=webserver"
   ```

2. **Status and Monitoring**
   ```go
   // phoenix experiment status exp-123
   // phoenix experiment metrics exp-123 --follow
   ```

3. **Comparison and Promotion**
   ```go
   // phoenix experiment compare exp-123
   // phoenix experiment promote exp-123 --variant candidate
   ```

### Phase 3: Pipeline Commands (Week 2)
1. **Pipeline Deployment**
   ```go
   // phoenix pipeline deploy --name process-topk-v1 \
   //   --selector "environment=production" \
   //   --namespace phoenix-prod
   ```

2. **Pipeline Management**
   ```go
   // phoenix pipeline list --type optimization
   // phoenix pipeline validate --file custom-pipeline.yaml
   ```

### Phase 4: Polish and Testing (Week 3)
1. **User Experience**
   - Add progress bars for long operations
   - Implement --watch flag for real-time updates
   - Add shell completion scripts
   - Create man pages

2. **Testing**
   - Unit tests for all commands
   - Integration tests with mock API
   - E2E tests against real Phoenix instance

## Code Structure

### Directory Layout
```
phoenix-platform/
├── cmd/
│   └── phoenix-cli/
│       ├── main.go
│       ├── cmd/
│       │   ├── root.go
│       │   ├── auth.go
│       │   ├── experiment.go
│       │   ├── pipeline.go
│       │   └── config.go
│       └── internal/
│           ├── output/       # Formatting helpers
│           ├── prompt/       # Interactive prompts
│           └── validation/   # Input validation
└── pkg/
    └── client/
        └── phoenix/
            ├── client.go     # Main API client
            ├── auth.go       # Auth endpoints
            ├── experiments.go # Experiment endpoints
            └── pipelines.go  # Pipeline endpoints
```

### Example Implementation

```go
// cmd/phoenix-cli/cmd/experiment.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/phoenix/pkg/client/phoenix"
)

var experimentCmd = &cobra.Command{
    Use:   "experiment",
    Short: "Manage Phoenix experiments",
}

var createExperimentCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a new experiment",
    RunE: func(cmd *cobra.Command, args []string) error {
        client := phoenix.NewClient(getAPIEndpoint(), getAuthToken())
        
        exp := &phoenix.ExperimentRequest{
            Name:             getString(cmd, "name"),
            BaselinePipeline: getString(cmd, "baseline"),
            CandidatePipeline: getString(cmd, "candidate"),
            Duration:         getDuration(cmd, "duration"),
            TargetNodes:      getMap(cmd, "target-selector"),
        }
        
        result, err := client.Experiments.Create(exp)
        if err != nil {
            return err
        }
        
        return outputResult(cmd, result)
    },
}

func init() {
    experimentCmd.AddCommand(createExperimentCmd)
    
    createExperimentCmd.Flags().StringP("name", "n", "", "Experiment name")
    createExperimentCmd.Flags().String("baseline", "", "Baseline pipeline")
    createExperimentCmd.Flags().String("candidate", "", "Candidate pipeline")
    createExperimentCmd.Flags().Duration("duration", 0, "Experiment duration")
    createExperimentCmd.Flags().StringToString("target-selector", nil, "Target node selectors")
    
    createExperimentCmd.MarkFlagRequired("name")
    createExperimentCmd.MarkFlagRequired("baseline")
    createExperimentCmd.MarkFlagRequired("candidate")
}
```

## Integration Points

### API Compatibility
- Use existing REST API endpoints
- Match request/response formats exactly
- Handle all API error codes gracefully
- Support both v1 and future API versions

### Configuration Files
Support YAML config files for complex operations:
```yaml
# ~/.phoenix/experiment.yaml
defaults:
  duration: 1h
  target-selector:
    environment: staging
  critical-processes:
    - nginx
    - postgres
    - redis
```

### Environment Variables
- `PHOENIX_API_URL`: API endpoint
- `PHOENIX_AUTH_TOKEN`: Authentication token
- `PHOENIX_OUTPUT_FORMAT`: Default output format
- `PHOENIX_NAMESPACE`: Default Kubernetes namespace

## Testing Strategy

### Unit Tests
- Mock API client for all commands
- Test command parsing and validation
- Test output formatting
- Test error handling

### Integration Tests
- Test against mock Phoenix API server
- Verify request/response handling
- Test authentication flow
- Test error scenarios

### E2E Tests
```bash
# Test script example
#!/bin/bash
phoenix auth login --username admin --password test
phoenix experiment create --name "cli-test" \
    --baseline process-baseline-v1 \
    --candidate process-topk-v1 \
    --duration 10m \
    --target-selector "app=test"
phoenix experiment status cli-test --follow
phoenix experiment promote cli-test --variant candidate
```

## Documentation

### User Guide
- Installation instructions
- Authentication setup
- Common workflows
- Examples for each command
- Troubleshooting guide

### Command Reference
- Auto-generated from Cobra
- Detailed flag descriptions
- Example usage for each command
- Output format examples

## Success Metrics

1. **Functionality**: All API operations available via CLI
2. **Usability**: Commands complete in <2s (except long operations)
3. **Reliability**: 99.9% success rate for valid commands
4. **Adoption**: 50% of power users using CLI within 1 month

## Timeline

- **Week 1**: Core framework + auth + experiment commands
- **Week 2**: Pipeline commands + advanced features
- **Week 3**: Testing, documentation, and release

## Dependencies

- Existing Phoenix API must be stable
- Authentication endpoints must support CLI flow
- Pipeline deployment API endpoint needs to be added

## Open Questions

1. Should we support config file templates for experiments?
2. Do we need offline mode with cached data?
3. Should CLI support multiple Phoenix instances?
4. Do we need plugin architecture for extensions?

---

This implementation plan addresses the critical gap identified in the functional review and provides a clear path to deliver a comprehensive CLI tool for Phoenix users.