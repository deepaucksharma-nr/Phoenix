#!/usr/bin/env bash
#
# Phoenix Auto-Scale Monitor
# Monitors system metrics and provides scaling recommendations
#

set -euo pipefail

# Configuration
PHOENIX_DIR="${PHOENIX_DIR:-/opt/phoenix}"
SCALING_RULES="$PHOENIX_DIR/config/scaling-rules.yml"
CHECK_INTERVAL="${CHECK_INTERVAL:-300}"  # 5 minutes
LOG_FILE="/var/log/phoenix-scaling.log"

# State file to track alerts
STATE_FILE="/var/lib/phoenix/scaling-state.json"
mkdir -p "$(dirname "$STATE_FILE")"

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

# Helper functions
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

alert() {
    echo -e "${RED}[ALERT]${NC} $*" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*" | tee -a "$LOG_FILE"
}

# Get current metrics
get_metrics() {
    local metric=$1
    local query=$2
    
    curl -s "http://localhost:9090/api/v1/query?query=$query" | \
        jq -r '.data.result[0].value[1] // "0"' 2>/dev/null || echo "0"
}

# Check CPU usage
check_cpu() {
    local cpu_query='100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)'
    local cpu_usage=$(get_metrics "cpu" "$cpu_query")
    
    # If Prometheus doesn't have node metrics, use docker stats
    if [[ "$cpu_usage" == "0" ]]; then
        cpu_usage=$(docker stats --no-stream --format "{{.CPUPerc}}" | \
            grep -Eo '[0-9]+' | awk '{sum+=$1} END {print sum}')
    fi
    
    echo "$cpu_usage"
}

# Check memory usage
check_memory() {
    local mem_query='(1 - (node_memory_AvailableBytes / node_memory_MemTotal)) * 100'
    local mem_usage=$(get_metrics "memory" "$mem_query")
    
    # Fallback to free command
    if [[ "$mem_usage" == "0" ]]; then
        mem_usage=$(free | grep Mem | awk '{print ($2-$7)/$2 * 100}')
    fi
    
    echo "$mem_usage"
}

# Check API latency
check_api_latency() {
    local latency_query='histogram_quantile(0.95, rate(phoenix_api_request_duration_seconds_bucket[5m]))'
    local latency=$(get_metrics "api_latency" "$latency_query")
    
    # Convert to milliseconds
    echo "$(awk "BEGIN {print $latency * 1000}")"
}

# Check metrics rate
check_metrics_rate() {
    local rate_query='sum(rate(phoenix_processed_series[1m]))'
    local rate=$(get_metrics "metrics_rate" "$rate_query")
    
    echo "$rate"
}

# Check agent count
check_agent_count() {
    local agent_query='count(time() - phoenix_agent_last_heartbeat < 60)'
    local count=$(get_metrics "agents" "$agent_query")
    
    echo "$count"
}

# Check disk usage
check_disk_usage() {
    df -h "$PHOENIX_DIR" | tail -1 | awk '{gsub("%",""); print $5}'
}

# Check database size
check_database_size() {
    docker-compose -f "$PHOENIX_DIR/docker-compose.yml" exec -T db \
        psql -U phoenix -t -c "SELECT pg_database_size('phoenix')/1024/1024/1024;" 2>/dev/null | \
        tr -d ' ' || echo "0"
}

# Save state
save_state() {
    local metric=$1
    local value=$2
    local threshold=$3
    local timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    # Create or update state file
    if [[ ! -f "$STATE_FILE" ]]; then
        echo "{}" > "$STATE_FILE"
    fi
    
    # Update metric state
    jq --arg metric "$metric" \
       --arg value "$value" \
       --arg threshold "$threshold" \
       --arg timestamp "$timestamp" \
       '.[$metric] = {value: $value, threshold: $threshold, timestamp: $timestamp}' \
       "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"
}

# Check if alert was already sent
should_alert() {
    local metric=$1
    local current_value=$2
    local threshold=$3
    
    # Check if we've alerted for this metric recently (within 1 hour)
    if [[ -f "$STATE_FILE" ]]; then
        local last_alert=$(jq -r --arg metric "$metric" '.[$metric].timestamp // ""' "$STATE_FILE")
        if [[ -n "$last_alert" ]]; then
            local last_epoch=$(date -d "$last_alert" +%s 2>/dev/null || echo "0")
            local now_epoch=$(date +%s)
            local diff=$((now_epoch - last_epoch))
            
            # Don't alert if we alerted less than an hour ago
            if [[ $diff -lt 3600 ]]; then
                return 1
            fi
        fi
    fi
    
    return 0
}

# Send alert
send_alert() {
    local level=$1
    local metric=$2
    local value=$3
    local threshold=$4
    local recommendation=$5
    
    # Log the alert
    alert "$level alert: $metric = $value (threshold: $threshold)"
    log "Recommendation: $recommendation"
    
    # Send to webhook if configured
    if [[ -n "${ALERT_WEBHOOK:-}" ]]; then
        curl -X POST "$ALERT_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d @- <<EOF
{
    "level": "$level",
    "metric": "$metric",
    "value": "$value",
    "threshold": "$threshold",
    "recommendation": "$recommendation",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    fi
    
    # Update state
    save_state "$metric" "$value" "$threshold"
}

# Main monitoring loop
monitor() {
    log "Starting Phoenix auto-scale monitor..."
    
    while true; do
        # Get current metrics
        local cpu=$(check_cpu)
        local memory=$(check_memory)
        local api_latency=$(check_api_latency)
        local metrics_rate=$(check_metrics_rate)
        local agent_count=$(check_agent_count)
        local disk_usage=$(check_disk_usage)
        local db_size=$(check_database_size)
        
        # Log current state
        log "Metrics - CPU: ${cpu}%, Memory: ${memory}%, Latency: ${api_latency}ms, " \
            "Metrics: ${metrics_rate}/s, Agents: ${agent_count}, Disk: ${disk_usage}%, DB: ${db_size}GB"
        
        # Check CPU
        if (( $(echo "$cpu > 85" | bc -l) )); then
            if should_alert "cpu_critical" "$cpu" "85"; then
                send_alert "CRITICAL" "CPU" "$cpu%" "85%" \
                    "Immediate vertical scaling needed. Consider t3.xlarge or component separation."
            fi
        elif (( $(echo "$cpu > 70" | bc -l) )); then
            if should_alert "cpu_warning" "$cpu" "70"; then
                send_alert "WARNING" "CPU" "$cpu%" "70%" \
                    "Consider vertical scaling to t3.large. Enable adaptive sampling."
            fi
        fi
        
        # Check Memory
        if (( $(echo "$memory > 90" | bc -l) )); then
            if should_alert "memory_critical" "$memory" "90"; then
                send_alert "CRITICAL" "Memory" "$memory%" "90%" \
                    "Add swap space immediately. Restart API service. Consider scaling."
            fi
        elif (( $(echo "$memory > 80" | bc -l) )); then
            if should_alert "memory_warning" "$memory" "80"; then
                send_alert "WARNING" "Memory" "$memory%" "80%" \
                    "Memory pressure detected. Consider adding swap or scaling up."
            fi
        fi
        
        # Check API Latency
        if (( $(echo "$api_latency > 500" | bc -l) )); then
            if should_alert "latency_critical" "$api_latency" "500"; then
                send_alert "CRITICAL" "API Latency" "${api_latency}ms" "500ms" \
                    "Severe performance degradation. Check database queries and enable caching."
            fi
        elif (( $(echo "$api_latency > 200" | bc -l) )); then
            if should_alert "latency_warning" "$api_latency" "200"; then
                send_alert "WARNING" "API Latency" "${api_latency}ms" "200ms" \
                    "Performance degrading. Consider enabling API response caching."
            fi
        fi
        
        # Check Metrics Rate
        if (( $(echo "$metrics_rate > 1000000" | bc -l) )); then
            if should_alert "metrics_critical" "$metrics_rate" "1000000"; then
                send_alert "CRITICAL" "Metrics Rate" "${metrics_rate}/s" "1M/s" \
                    "Maximum single-VM capacity reached. Horizontal scaling required."
            fi
        elif (( $(echo "$metrics_rate > 800000" | bc -l) )); then
            if should_alert "metrics_warning" "$metrics_rate" "800000"; then
                send_alert "WARNING" "Metrics Rate" "${metrics_rate}/s" "800K/s" \
                    "Approaching metrics limit. Deploy more aggressive filters."
            fi
        fi
        
        # Check Agent Count
        if (( agent_count > 200 )); then
            if should_alert "agents_critical" "$agent_count" "200"; then
                send_alert "CRITICAL" "Agent Count" "$agent_count" "200" \
                    "Maximum recommended agents exceeded. Move to Kubernetes deployment."
            fi
        elif (( agent_count > 150 )); then
            if should_alert "agents_warning" "$agent_count" "150"; then
                send_alert "WARNING" "Agent Count" "$agent_count" "150" \
                    "Approaching agent limit. Plan for horizontal scaling."
            fi
        fi
        
        # Check Disk Usage
        if (( disk_usage > 90 )); then
            if should_alert "disk_critical" "$disk_usage" "90"; then
                send_alert "CRITICAL" "Disk Usage" "$disk_usage%" "90%" \
                    "Disk critically full. Delete old backups and run database VACUUM."
            fi
        elif (( disk_usage > 80 )); then
            if should_alert "disk_warning" "$disk_usage" "80"; then
                send_alert "WARNING" "Disk Usage" "$disk_usage%" "80%" \
                    "Disk filling up. Consider expanding volume or cleanup."
            fi
        fi
        
        # Sleep until next check
        sleep "$CHECK_INTERVAL"
    done
}

# Signal handlers
trap 'log "Shutting down auto-scale monitor..."; exit 0' SIGTERM SIGINT

# Create systemd service if it doesn't exist
create_service() {
    if [[ ! -f /etc/systemd/system/phoenix-scale-monitor.service ]]; then
        cat > /etc/systemd/system/phoenix-scale-monitor.service << EOF
[Unit]
Description=Phoenix Auto-Scale Monitor
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=root
ExecStart=$0
Restart=always
RestartSec=30
Environment="PHOENIX_DIR=$PHOENIX_DIR"

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        log "Created systemd service: phoenix-scale-monitor.service"
    fi
}

# Main
if [[ "${1:-}" == "--install" ]]; then
    create_service
    systemctl enable phoenix-scale-monitor
    systemctl start phoenix-scale-monitor
    log "Phoenix scale monitor installed and started"
else
    monitor
fi