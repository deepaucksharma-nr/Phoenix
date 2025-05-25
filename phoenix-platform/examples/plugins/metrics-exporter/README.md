# Phoenix Metrics Exporter Plugin

Export experiment metrics to various formats and destinations including CSV, JSON, InfluxDB, Prometheus, and Slack.

## Installation

```bash
# Install the plugin
phoenix plugin install examples/plugins/metrics-exporter

# Verify installation
phoenix plugin list
```

## Usage

### Basic Export to CSV

```bash
# Export experiment metrics to CSV file
phoenix metrics-exporter --format csv --output experiment-123.csv exp-123

# Export to stdout
phoenix metrics-exporter --format csv exp-123
```

### Export to Different Formats

```bash
# JSON format
phoenix metrics-exporter --format json --output metrics.json exp-123

# InfluxDB line protocol
phoenix metrics-exporter --format influx --output metrics.influx exp-123

# Prometheus format
phoenix metrics-exporter --format prometheus --output metrics.prom exp-123
```

### Time Range Filtering

```bash
# Export last 24 hours
phoenix metrics-exporter --start -24h --end now exp-123

# Export specific time range
phoenix metrics-exporter --start 2024-01-15T10:00:00Z --end 2024-01-15T18:00:00Z exp-123

# Export with custom interval
phoenix metrics-exporter --interval 1h --start -7d exp-123
```

### Export to External Systems

#### InfluxDB

```bash
# Direct export to InfluxDB
phoenix metrics-exporter \
  --destination influxdb \
  --format influx \
  --influxdb-url http://localhost:8086 \
  --influxdb-token your-token \
  --influxdb-org your-org \
  --influxdb-bucket metrics \
  exp-123

# Using environment variables
export INFLUXDB_URL=http://localhost:8086
export INFLUXDB_TOKEN=your-token
export INFLUXDB_ORG=your-org
export INFLUXDB_BUCKET=metrics

phoenix metrics-exporter --destination influxdb --format influx exp-123
```

#### Slack Notifications

```bash
# Send metrics summary to Slack
phoenix metrics-exporter \
  --destination slack \
  --slack-webhook https://hooks.slack.com/services/... \
  exp-123

# Using environment variable
export SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
phoenix metrics-exporter --destination slack exp-123
```

## Configuration

### Environment Variables

- `INFLUXDB_URL` - Default InfluxDB URL
- `INFLUXDB_TOKEN` - Default InfluxDB authentication token
- `INFLUXDB_ORG` - Default InfluxDB organization
- `INFLUXDB_BUCKET` - Default InfluxDB bucket name
- `SLACK_WEBHOOK_URL` - Default Slack webhook URL

### Output Formats

#### CSV Format
Contains columns for all key metrics including cost reduction, data loss, progress, and pipeline-specific metrics.

#### JSON Format
Raw JSON response from the Phoenix API with full metric details.

#### InfluxDB Line Protocol
Formatted for direct ingestion into InfluxDB with proper tags and fields.

#### Prometheus Format
Metrics formatted as Prometheus exposition format with proper labels.

## Examples

### Daily Metrics Report

```bash
#!/bin/bash
# Daily metrics export script

DATE=$(date +%Y-%m-%d)
REPORT_DIR="reports/$DATE"
mkdir -p "$REPORT_DIR"

# Get all running experiments
EXPERIMENTS=$(phoenix experiment list --status completed --output json | jq -r '.experiments[].id')

for exp_id in $EXPERIMENTS; do
  echo "Exporting metrics for $exp_id..."
  
  # Export to CSV
  phoenix metrics-exporter \
    --format csv \
    --output "$REPORT_DIR/$exp_id.csv" \
    --start -24h \
    "$exp_id"
  
  # Send to InfluxDB
  phoenix metrics-exporter \
    --destination influxdb \
    --format influx \
    --start -24h \
    "$exp_id"
done

echo "Daily metrics export completed!"
```

### Automated Alerting

```bash
#!/bin/bash
# Check experiments and alert on issues

for exp_id in $(phoenix experiment list --status running --output json | jq -r '.experiments[].id'); do
  # Get current metrics
  METRICS=$(phoenix experiment metrics "$exp_id" --output json)
  DATA_LOSS=$(echo "$METRICS" | jq -r '.summary.data_loss_percent')
  
  # Alert if data loss is high
  if (( $(echo "$DATA_LOSS > 5.0" | bc -l) )); then
    phoenix metrics-exporter \
      --destination slack \
      --slack-webhook "$ALERT_WEBHOOK" \
      "$exp_id"
  fi
done
```

### Batch Processing

```bash
#!/bin/bash
# Process multiple experiments

# Export all completed experiments from last week
phoenix experiment list --status completed --output json | \
  jq -r '.experiments[] | select(.created_at > (now - 7*24*3600 | todate)) | .id' | \
  while read -r exp_id; do
    phoenix metrics-exporter \
      --format json \
      --output "exports/$exp_id.json" \
      --start -7d \
      "$exp_id"
  done
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x ~/.phoenix/plugins/metrics-exporter/main.sh
   ```

2. **Missing Dependencies**
   - Ensure `jq` is installed for JSON processing
   - Ensure `curl` is available for HTTP requests
   - Ensure `bc` is available for floating-point calculations

3. **Authentication Issues**
   ```bash
   # Verify Phoenix CLI authentication
   phoenix auth status
   
   # Re-authenticate if needed
   phoenix auth login
   ```

4. **InfluxDB Connection Issues**
   ```bash
   # Test connection manually
   curl -I "$INFLUXDB_URL/health"
   
   # Verify token permissions
   curl -H "Authorization: Token $INFLUXDB_TOKEN" \
        "$INFLUXDB_URL/api/v2/buckets?org=$INFLUXDB_ORG"
   ```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Run with debug output
bash -x ~/.phoenix/plugins/metrics-exporter/main.sh --help

# Check plugin execution
phoenix --verbose metrics-exporter exp-123
```

## Development

To modify or extend this plugin:

1. Edit `main.sh` for functionality changes
2. Update `plugin.json` for metadata changes
3. Reinstall the plugin: `phoenix plugin install . --force`

### Adding New Export Formats

To add a new export format, add a new function in `main.sh`:

```bash
export_to_newformat() {
    local json="$1"
    local output="$2"
    
    # Process JSON and write to output
    echo "$json" | jq '...' > "${output:-/dev/stdout}"
}
```

Then add the format to the case statement in the conversion section.

### Adding New Destinations

To add a new destination, add a new case in the destination handling section:

```bash
newdestination)
    # Validate required configuration
    # Process and send data
    ;;
```

## License

This plugin is part of the Phoenix Platform and follows the same license terms.