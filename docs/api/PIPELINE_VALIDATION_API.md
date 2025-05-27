# Pipeline Validation API

## Overview

The Phoenix Platform provides comprehensive pipeline configuration validation to ensure that OpenTelemetry Collector configurations are correct before deployment. This prevents runtime errors and ensures smooth pipeline operations.

## Endpoints

### POST /api/v1/pipelines/validate

Validates a pipeline configuration for correctness and completeness.

#### Request

```json
{
  "config": {
    // OpenTelemetry Collector configuration object
  },
  // OR
  "yaml": "string containing YAML configuration"
}
```

You can provide the configuration in either JSON format (`config`) or YAML format (`yaml`).

#### Response

Success Response (200 OK):
```json
{
  "valid": true,
  "message": "Pipeline configuration is valid"
}
```

Validation Failed Response (200 OK):
```json
{
  "valid": false,
  "error": "Detailed error message explaining what's wrong"
}
```

#### Validation Rules

The validation endpoint checks the following:

1. **Required Components**
   - At least one receiver must be defined
   - At least one exporter must be defined
   - At least one service pipeline must be defined

2. **Receiver Validation**
   - OTLP receivers must have protocols configured
   - Protocol endpoints cannot be empty
   - Host metrics collection intervals must be valid durations

3. **Processor Validation**
   - Batch processor timeouts must be valid durations
   - Batch sizes cannot be negative
   - Memory limiter values must be positive
   - Phoenix custom processors have specific validation:
     - Adaptive filter: cardinality limits must be positive
     - TopK: k value must be positive, window size must be valid duration

4. **Exporter Validation**
   - Endpoints cannot be empty
   - TLS configuration must be properly formatted

5. **Service Pipeline Validation**
   - All referenced receivers must be defined
   - All referenced processors must be defined
   - All referenced exporters must be defined
   - Each pipeline must have at least one receiver and exporter

## Examples

### Example 1: Valid Configuration

```bash
curl -X POST http://localhost:8080/api/v1/pipelines/validate \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "receivers": {
        "otlp": {
          "protocols": {
            "grpc": {
              "endpoint": "0.0.0.0:4317"
            },
            "http": {
              "endpoint": "0.0.0.0:4318"
            }
          }
        }
      },
      "processors": {
        "batch": {
          "timeout": "1s",
          "send_batch_size": 1024
        },
        "memory_limiter": {
          "check_interval": "1s",
          "limit_mib": 512
        }
      },
      "exporters": {
        "prometheus": {
          "endpoint": "0.0.0.0:8889"
        }
      },
      "service": {
        "pipelines": {
          "metrics": {
            "receivers": ["otlp"],
            "processors": ["memory_limiter", "batch"],
            "exporters": ["prometheus"]
          }
        }
      }
    }
  }'
```

Response:
```json
{
  "valid": true,
  "message": "Pipeline configuration is valid"
}
```

### Example 2: YAML Configuration

```bash
curl -X POST http://localhost:8080/api/v1/pipelines/validate \
  -H "Content-Type: application/json" \
  -d '{
    "yaml": "receivers:\n  otlp:\n    protocols:\n      grpc:\n        endpoint: 0.0.0.0:4317\n\nprocessors:\n  batch:\n    timeout: 1s\n\nexporters:\n  prometheus:\n    endpoint: 0.0.0.0:8889\n\nservice:\n  pipelines:\n    metrics:\n      receivers: [otlp]\n      processors: [batch]\n      exporters: [prometheus]"
  }'
```

### Example 3: Invalid Configuration

```bash
curl -X POST http://localhost:8080/api/v1/pipelines/validate \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "receivers": {
        "otlp": {
          "protocols": {
            "grpc": {
              "endpoint": ""
            }
          }
        }
      },
      "exporters": {
        "prometheus": {
          "endpoint": "0.0.0.0:8889"
        }
      },
      "service": {
        "pipelines": {
          "metrics": {
            "receivers": ["otlp"],
            "exporters": ["prometheus"]
          }
        }
      }
    }
  }'
```

Response:
```json
{
  "valid": false,
  "error": "invalid receiver otlp: gRPC endpoint cannot be empty"
}
```

### Example 4: Phoenix Custom Processor Validation

```bash
curl -X POST http://localhost:8080/api/v1/pipelines/validate \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "receivers": {
        "otlp": {
          "protocols": {
            "grpc": {
              "endpoint": "0.0.0.0:4317"
            }
          }
        }
      },
      "processors": {
        "phoenix_adaptive_filter": {
          "adaptive_filter": {
            "enabled": true,
            "thresholds": {
              "cardinality_limit": 50000,
              "cpu_threshold": 80,
              "memory_threshold": 85
            },
            "rules": [
              {
                "name": "drop_debug_metrics",
                "condition": "attributes[\"level\"] == \"debug\"",
                "action": "drop"
              }
            ]
          }
        }
      },
      "exporters": {
        "prometheus": {
          "endpoint": "0.0.0.0:8889"
        }
      },
      "service": {
        "pipelines": {
          "metrics": {
            "receivers": ["otlp"],
            "processors": ["phoenix_adaptive_filter"],
            "exporters": ["prometheus"]
          }
        }
      }
    }
  }'
```

## Error Messages

Common validation errors and their meanings:

| Error Message | Description | Solution |
|--------------|-------------|----------|
| `pipeline must have at least one receiver` | No receivers defined | Add at least one receiver configuration |
| `pipeline must have at least one exporter` | No exporters defined | Add at least one exporter configuration |
| `pipeline must have at least one service pipeline` | No service pipelines defined | Add a service pipeline configuration |
| `OTLP receiver must have protocols configured` | OTLP receiver missing protocol config | Add gRPC or HTTP protocol configuration |
| `gRPC endpoint cannot be empty` | Empty endpoint string | Provide a valid endpoint (e.g., "0.0.0.0:4317") |
| `invalid timeout: time: invalid duration` | Invalid duration format | Use valid Go duration (e.g., "1s", "5m", "1h") |
| `send_batch_size cannot be negative` | Negative batch size | Use a positive integer for batch size |
| `cardinality_limit must be positive` | Non-positive cardinality limit | Set a positive value for cardinality limit |
| `pipeline X references undefined receiver: Y` | Service pipeline references non-existent receiver | Ensure all referenced receivers are defined |
| `pipeline X references undefined processor: Y` | Service pipeline references non-existent processor | Ensure all referenced processors are defined |
| `pipeline X references undefined exporter: Y` | Service pipeline references non-existent exporter | Ensure all referenced exporters are defined |

## Best Practices

1. **Validate Before Deployment**: Always validate pipeline configurations before deploying them to agents
2. **Use YAML for Readability**: When manually creating configs, YAML format is often more readable
3. **Test Incrementally**: Start with minimal configs and add components one at a time
4. **Check References**: Ensure all components referenced in service pipelines are actually defined
5. **Use Appropriate Values**: 
   - Memory limits should be based on available resources
   - Batch sizes should balance latency and throughput
   - Timeouts should account for network conditions

## Integration with CI/CD

You can integrate pipeline validation into your CI/CD pipeline:

```yaml
# Example GitHub Actions step
- name: Validate Pipeline Config
  run: |
    response=$(curl -s -X POST http://phoenix-api:8080/api/v1/pipelines/validate \
      -H "Content-Type: application/json" \
      -d @pipeline-config.json)
    
    if [ "$(echo $response | jq -r .valid)" != "true" ]; then
      echo "Pipeline validation failed: $(echo $response | jq -r .error)"
      exit 1
    fi
```

## Related Endpoints

- `POST /api/v1/pipelines/render` - Render a pipeline template with variables
- `GET /api/v1/pipelines` - List available pipeline templates
- `POST /api/v1/pipelines/deployments` - Deploy a validated pipeline