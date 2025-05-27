# NRDOT Troubleshooting Guide

This guide helps diagnose and resolve common issues with NRDOT (New Relic Distribution of OpenTelemetry) integration in Phoenix.

## Common Issues

### 1. NRDOT Binary Not Found

**Symptoms:**
- Error: `exec: "nrdot": executable file not found in $PATH`
- Agent fails to start collector

**Solutions:**
```bash
# Check if NRDOT is installed
which nrdot

# If not found, install manually
cd /tmp
wget https://github.com/newrelic/nrdot-collector-releases/releases/latest/download/nrdot-collector-host_linux_amd64.tar.gz
tar -xzf nrdot-collector-host_linux_amd64.tar.gz
sudo mv nrdot-collector-host /usr/local/bin/nrdot
sudo chmod +x /usr/local/bin/nrdot

# For agents in Docker, rebuild image
docker build -f Dockerfile.phoenix-agent -t phoenix-agent .
```

### 2. License Key Issues

**Symptoms:**
- Error: `NEW_RELIC_LICENSE_KEY is required when using NRDOT collector`
- Error: `401 Unauthorized` from New Relic endpoint

**Solutions:**
```bash
# Verify license key is set
echo $NEW_RELIC_LICENSE_KEY

# Set license key for agent
export NEW_RELIC_LICENSE_KEY=your-actual-license-key

# For Docker Compose
# Add to docker-compose.yml or .env file
NEW_RELIC_LICENSE_KEY=your-actual-license-key

# Verify key format (should be 40 characters)
echo -n "$NEW_RELIC_LICENSE_KEY" | wc -c
```

### 3. OTLP Endpoint Connection Issues

**Symptoms:**
- Error: `failed to export metrics: connection refused`
- Timeout errors connecting to New Relic

**Solutions:**
```bash
# Test connectivity to New Relic endpoint
nc -zv otlp.nr-data.net 4317

# Check for proxy/firewall issues
curl -v telnet://otlp.nr-data.net:4317

# Try alternative endpoints
# US: otlp.nr-data.net:4317
# EU: otlp.eu01.nr-data.net:4317

# Set custom endpoint
export NEW_RELIC_OTLP_ENDPOINT=otlp.eu01.nr-data.net:4317
```

### 4. Cardinality Reduction Not Working

**Symptoms:**
- Metrics cardinality remains high
- No reduction in New Relic UI

**Solutions:**
```yaml
# Check NRDOT configuration has cardinality processor
processors:
  newrelic/cardinality:
    enabled: true
    max_series: 10000
    reduction_target_percentage: 70
    
# Verify processor is in pipeline
service:
  pipelines:
    metrics:
      processors: [newrelic/cardinality, batch]
```

```bash
# Check NRDOT logs
tail -f /etc/phoenix-agent/nrdot-*.log | grep cardinality

# Verify feature gates
nrdot --feature-gates=exporter.newrelic.cardinality_reduction
```

### 5. Agent Not Receiving NRDOT Tasks

**Symptoms:**
- Agent polls but doesn't get NRDOT deployment tasks
- Tasks show collector_type as "otel" instead of "nrdot"

**Solutions:**
```bash
# Check experiment metadata
phoenix-cli experiment get --id exp-123 --format json | jq .metadata

# Verify experiment was created with NRDOT
phoenix-cli experiment create \
  --use-nrdot \
  --nr-license-key "$NEW_RELIC_LICENSE_KEY" \
  ...

# Check task configuration
curl -H "X-Agent-Host-ID: agent-001" \
  http://localhost:8080/api/v1/agent/tasks | jq
```

### 6. Template Rendering Issues

**Symptoms:**
- Error: `template not found: nrdot-cardinality`
- Invalid YAML in rendered configuration

**Solutions:**
```bash
# List available templates
curl http://localhost:8080/api/v1/pipelines/templates | jq

# Test template rendering
curl -X POST http://localhost:8080/api/v1/pipelines/render \
  -H "Content-Type: application/json" \
  -d '{
    "template": "nrdot-cardinality",
    "parameters": {
      "nr_license_key": "test-key",
      "nr_otlp_endpoint": "otlp.nr-data.net:4317"
    }
  }' | jq
```

## Debugging Commands

### Check NRDOT Version
```bash
nrdot --version
```

### Validate NRDOT Configuration
```bash
# Dry run to validate config
nrdot --config /path/to/config.yaml --dry-run

# Check config syntax
nrdot --config /path/to/config.yaml --validate
```

### Monitor NRDOT Process
```bash
# Check if NRDOT is running
ps aux | grep nrdot

# Check NRDOT resource usage
top -p $(pgrep nrdot)

# Check NRDOT logs
journalctl -u phoenix-agent -f | grep nrdot
```

### Test NRDOT Manually
```bash
# Create test config
cat > /tmp/nrdot-test.yaml << EOF
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:
    timeout: 1s

exporters:
  otlp/newrelic:
    endpoint: otlp.nr-data.net:4317
    headers:
      api-key: $NEW_RELIC_LICENSE_KEY

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/newrelic]
EOF

# Run NRDOT manually
nrdot --config /tmp/nrdot-test.yaml
```

## Environment Variable Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| USE_NRDOT | Enable NRDOT collector | false | No |
| NEW_RELIC_LICENSE_KEY | New Relic license key | - | Yes (if USE_NRDOT=true) |
| NEW_RELIC_OTLP_ENDPOINT | OTLP endpoint | otlp.nr-data.net:4317 | No |
| COLLECTOR_TYPE | Collector type (otel/nrdot) | otel | No |
| MAX_CARDINALITY | Max metric cardinality | 10000 | No |
| REDUCTION_PERCENTAGE | Target reduction % | 70 | No |

## Log Locations

- Agent logs: `/var/log/phoenix-agent/agent.log`
- NRDOT logs: `/etc/phoenix-agent/dep-*.log`
- API logs: Check container logs or systemd journal
- Task execution logs: `/var/log/phoenix-agent/tasks/`

## Getting Help

1. Check NRDOT documentation: https://docs.newrelic.com/docs/nrdot
2. Phoenix issues: https://github.com/phoenix/platform/issues
3. New Relic support: https://support.newrelic.com

## Performance Tuning

### Memory Usage
```yaml
# Limit NRDOT memory usage
processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128
```

### Batch Processing
```yaml
# Optimize batch size for New Relic
processors:
  batch:
    timeout: 5s
    send_batch_size: 1000
    send_batch_max_size: 2000
```

### Compression
```yaml
# Enable compression for New Relic
exporters:
  otlp/newrelic:
    compression: gzip
```