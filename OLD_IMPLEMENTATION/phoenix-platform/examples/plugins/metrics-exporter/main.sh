#!/bin/bash
# Phoenix CLI Plugin: Metrics Exporter
# Export experiment metrics to various formats and destinations

set -e

PLUGIN_NAME="metrics-exporter"
VERSION="1.0.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
Phoenix Metrics Exporter Plugin v${VERSION}

USAGE:
    phoenix metrics-exporter [OPTIONS] <experiment-id>

DESCRIPTION:
    Export experiment metrics to various formats and destinations.
    Supports CSV, JSON, InfluxDB line protocol, and more.

OPTIONS:
    --format FORMAT         Output format: csv, json, influx, prometheus (default: csv)
    --output FILE          Output file (default: stdout)
    --destination DEST     Export destination: file, influxdb, prometheus, slack
    --influxdb-url URL     InfluxDB URL for direct export
    --influxdb-token TOKEN InfluxDB authentication token
    --influxdb-org ORG     InfluxDB organization
    --influxdb-bucket NAME InfluxDB bucket name
    --slack-webhook URL    Slack webhook URL for notifications
    --interval DURATION    Time interval for metrics (1m, 5m, 1h, 1d)
    --start TIME          Start time (RFC3339 or relative like -2h)
    --end TIME            End time (RFC3339 or relative like now)
    --help                 Show this help message

EXAMPLES:
    # Export to CSV file
    phoenix metrics-exporter --format csv --output experiment-metrics.csv exp-123

    # Export to InfluxDB
    phoenix metrics-exporter --format influx --destination influxdb \\
        --influxdb-url http://localhost:8086 \\
        --influxdb-token mytoken \\
        --influxdb-org myorg \\
        --influxdb-bucket metrics \\
        exp-123

    # Export last 24 hours to JSON
    phoenix metrics-exporter --format json --start -24h --end now exp-123

    # Send summary to Slack
    phoenix metrics-exporter --destination slack \\
        --slack-webhook https://hooks.slack.com/... \\
        exp-123

ENVIRONMENT VARIABLES:
    INFLUXDB_URL          Default InfluxDB URL
    INFLUXDB_TOKEN        Default InfluxDB token
    INFLUXDB_ORG          Default InfluxDB organization
    INFLUXDB_BUCKET       Default InfluxDB bucket
    SLACK_WEBHOOK_URL     Default Slack webhook URL
EOF
}

# Parse command line arguments
FORMAT="csv"
OUTPUT=""
DESTINATION="file"
INFLUXDB_URL="${INFLUXDB_URL:-}"
INFLUXDB_TOKEN="${INFLUXDB_TOKEN:-}"
INFLUXDB_ORG="${INFLUXDB_ORG:-}"
INFLUXDB_BUCKET="${INFLUXDB_BUCKET:-}"
SLACK_WEBHOOK="${SLACK_WEBHOOK_URL:-}"
INTERVAL="5m"
START=""
END=""
EXPERIMENT_ID=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --output)
            OUTPUT="$2"
            shift 2
            ;;
        --destination)
            DESTINATION="$2"
            shift 2
            ;;
        --influxdb-url)
            INFLUXDB_URL="$2"
            shift 2
            ;;
        --influxdb-token)
            INFLUXDB_TOKEN="$2"
            shift 2
            ;;
        --influxdb-org)
            INFLUXDB_ORG="$2"
            shift 2
            ;;
        --influxdb-bucket)
            INFLUXDB_BUCKET="$2"
            shift 2
            ;;
        --slack-webhook)
            SLACK_WEBHOOK="$2"
            shift 2
            ;;
        --interval)
            INTERVAL="$2"
            shift 2
            ;;
        --start)
            START="$2"
            shift 2
            ;;
        --end)
            END="$2"
            shift 2
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        -*)
            log_error "Unknown option: $1"
            exit 1
            ;;
        *)
            if [[ -z "$EXPERIMENT_ID" ]]; then
                EXPERIMENT_ID="$1"
            else
                log_error "Multiple experiment IDs provided"
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate required arguments
if [[ -z "$EXPERIMENT_ID" ]]; then
    log_error "Experiment ID is required"
    echo
    show_help
    exit 1
fi

# Validate format
case "$FORMAT" in
    csv|json|influx|prometheus)
        ;;
    *)
        log_error "Invalid format: $FORMAT"
        log_error "Supported formats: csv, json, influx, prometheus"
        exit 1
        ;;
esac

# Get experiment metrics
log_info "Fetching metrics for experiment $EXPERIMENT_ID..."

METRICS_ARGS="--output json"
if [[ -n "$INTERVAL" ]]; then
    METRICS_ARGS="$METRICS_ARGS --interval $INTERVAL"
fi
if [[ -n "$START" ]]; then
    METRICS_ARGS="$METRICS_ARGS --start $START"
fi
if [[ -n "$END" ]]; then
    METRICS_ARGS="$METRICS_ARGS --end $END"
fi

# Fetch metrics using phoenix CLI
METRICS_JSON=$(phoenix experiment metrics $EXPERIMENT_ID $METRICS_ARGS)
if [[ $? -ne 0 ]]; then
    log_error "Failed to fetch experiment metrics"
    exit 1
fi

# Export based on format
export_to_csv() {
    local json="$1"
    local output="$2"
    
    log_info "Converting metrics to CSV format..."
    
    # Extract summary metrics
    echo "$json" | jq -r '
        ["timestamp", "experiment_id", "cost_reduction_percent", "data_loss_percent", "progress_percent", "estimated_monthly_savings", "pipeline_a_dps", "pipeline_b_dps", "pipeline_a_bytes", "pipeline_b_bytes", "pipeline_a_errors", "pipeline_b_errors"],
        [
            .timestamp,
            .experiment_id,
            .summary.cost_reduction_percent,
            .summary.data_loss_percent,
            .summary.progress_percent,
            .summary.estimated_monthly_savings,
            .pipeline_a.data_points_per_second,
            .pipeline_b.data_points_per_second,
            .pipeline_a.bytes_per_second,
            .pipeline_b.bytes_per_second,
            .pipeline_a.error_rate,
            .pipeline_b.error_rate
        ] | @csv
    ' > "${output:-/dev/stdout}"
}

export_to_influx() {
    local json="$1"
    local output="$2"
    
    log_info "Converting metrics to InfluxDB line protocol..."
    
    echo "$json" | jq -r '
        "experiment_metrics,experiment_id=" + .experiment_id + 
        " cost_reduction_percent=" + (.summary.cost_reduction_percent | tostring) + 
        ",data_loss_percent=" + (.summary.data_loss_percent | tostring) + 
        ",progress_percent=" + (.summary.progress_percent | tostring) + 
        ",estimated_monthly_savings=" + (.summary.estimated_monthly_savings | tostring) + 
        ",pipeline_a_dps=" + (.pipeline_a.data_points_per_second | tostring) + 
        ",pipeline_b_dps=" + (.pipeline_b.data_points_per_second | tostring) + 
        " " + ((.timestamp | fromdateiso8601) * 1000000000 | tostring)
    ' > "${output:-/dev/stdout}"
}

export_to_prometheus() {
    local json="$1"
    local output="$2"
    
    log_info "Converting metrics to Prometheus format..."
    
    EXP_ID=$(echo "$json" | jq -r '.experiment_id')
    
    cat > "${output:-/dev/stdout}" << EOF
# HELP phoenix_experiment_cost_reduction_percent Cost reduction percentage for experiment
# TYPE phoenix_experiment_cost_reduction_percent gauge
phoenix_experiment_cost_reduction_percent{experiment_id="$EXP_ID"} $(echo "$json" | jq -r '.summary.cost_reduction_percent')

# HELP phoenix_experiment_data_loss_percent Data loss percentage for experiment
# TYPE phoenix_experiment_data_loss_percent gauge
phoenix_experiment_data_loss_percent{experiment_id="$EXP_ID"} $(echo "$json" | jq -r '.summary.data_loss_percent')

# HELP phoenix_experiment_progress_percent Progress percentage for experiment
# TYPE phoenix_experiment_progress_percent gauge
phoenix_experiment_progress_percent{experiment_id="$EXP_ID"} $(echo "$json" | jq -r '.summary.progress_percent')

# HELP phoenix_experiment_estimated_monthly_savings Estimated monthly savings in USD
# TYPE phoenix_experiment_estimated_monthly_savings gauge
phoenix_experiment_estimated_monthly_savings{experiment_id="$EXP_ID"} $(echo "$json" | jq -r '.summary.estimated_monthly_savings')

# HELP phoenix_pipeline_data_points_per_second Data points per second by pipeline
# TYPE phoenix_pipeline_data_points_per_second gauge
phoenix_pipeline_data_points_per_second{experiment_id="$EXP_ID",pipeline="a"} $(echo "$json" | jq -r '.pipeline_a.data_points_per_second')
phoenix_pipeline_data_points_per_second{experiment_id="$EXP_ID",pipeline="b"} $(echo "$json" | jq -r '.pipeline_b.data_points_per_second')
EOF
}

# Convert metrics based on format
TEMP_FILE=$(mktemp)
case "$FORMAT" in
    csv)
        export_to_csv "$METRICS_JSON" "$TEMP_FILE"
        ;;
    json)
        echo "$METRICS_JSON" > "$TEMP_FILE"
        ;;
    influx)
        export_to_influx "$METRICS_JSON" "$TEMP_FILE"
        ;;
    prometheus)
        export_to_prometheus "$METRICS_JSON" "$TEMP_FILE"
        ;;
esac

# Handle destination
case "$DESTINATION" in
    file)
        if [[ -n "$OUTPUT" ]]; then
            cp "$TEMP_FILE" "$OUTPUT"
            log_success "Metrics exported to: $OUTPUT"
        else
            cat "$TEMP_FILE"
        fi
        ;;
    influxdb)
        if [[ -z "$INFLUXDB_URL" || -z "$INFLUXDB_TOKEN" || -z "$INFLUXDB_ORG" || -z "$INFLUXDB_BUCKET" ]]; then
            log_error "InfluxDB configuration incomplete"
            log_error "Required: --influxdb-url, --influxdb-token, --influxdb-org, --influxdb-bucket"
            exit 1
        fi
        
        log_info "Uploading metrics to InfluxDB..."
        if curl -s -X POST "$INFLUXDB_URL/api/v2/write?org=$INFLUXDB_ORG&bucket=$INFLUXDB_BUCKET" \
            -H "Authorization: Token $INFLUXDB_TOKEN" \
            -H "Content-Type: text/plain" \
            --data-binary @"$TEMP_FILE"; then
            log_success "Metrics uploaded to InfluxDB successfully"
        else
            log_error "Failed to upload metrics to InfluxDB"
            exit 1
        fi
        ;;
    slack)
        if [[ -z "$SLACK_WEBHOOK" ]]; then
            log_error "Slack webhook URL is required for Slack destination"
            exit 1
        fi
        
        log_info "Sending metrics summary to Slack..."
        
        # Extract key metrics for Slack message
        COST_REDUCTION=$(echo "$METRICS_JSON" | jq -r '.summary.cost_reduction_percent')
        DATA_LOSS=$(echo "$METRICS_JSON" | jq -r '.summary.data_loss_percent')
        PROGRESS=$(echo "$METRICS_JSON" | jq -r '.summary.progress_percent')
        SAVINGS=$(echo "$METRICS_JSON" | jq -r '.summary.estimated_monthly_savings')
        
        SLACK_MESSAGE=$(cat << EOF
{
  "blocks": [
    {
      "type": "header",
      "text": {
        "type": "plain_text",
        "text": "📊 Experiment Metrics: $EXPERIMENT_ID"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*Cost Reduction:*\n${COST_REDUCTION}%"
        },
        {
          "type": "mrkdwn",
          "text": "*Data Loss:*\n${DATA_LOSS}%"
        },
        {
          "type": "mrkdwn",
          "text": "*Progress:*\n${PROGRESS}%"
        },
        {
          "type": "mrkdwn",
          "text": "*Est. Monthly Savings:*\n\$${SAVINGS}"
        }
      ]
    }
  ]
}
EOF
)
        
        if curl -s -X POST "$SLACK_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "$SLACK_MESSAGE" > /dev/null; then
            log_success "Metrics summary sent to Slack"
        else
            log_error "Failed to send metrics to Slack"
            exit 1
        fi
        ;;
    *)
        log_error "Unknown destination: $DESTINATION"
        exit 1
        ;;
esac

# Cleanup
rm -f "$TEMP_FILE"

log_success "Metrics export completed successfully!"