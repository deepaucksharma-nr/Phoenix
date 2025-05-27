#!/bin/bash
# Phoenix monitoring script - monitors health of all components

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PHOENIX_API_URL="${PHOENIX_API_URL:-http://localhost:8080}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
REFRESH_INTERVAL="${REFRESH_INTERVAL:-5}"
ALERT_THRESHOLD_CPU="${ALERT_THRESHOLD_CPU:-80}"
ALERT_THRESHOLD_MEM="${ALERT_THRESHOLD_MEM:-80}"

# Parse arguments
CONTINUOUS=false
OUTPUT_FORMAT="terminal"
while [[ $# -gt 0 ]]; do
    case $1 in
        --continuous)
            CONTINUOUS=true
            shift
            ;;
        --json)
            OUTPUT_FORMAT="json"
            shift
            ;;
        --api-url)
            PHOENIX_API_URL="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--continuous] [--json] [--api-url URL]"
            exit 1
            ;;
    esac
done

# Function to get API data
api_get() {
    local endpoint=$1
    curl -s "$PHOENIX_API_URL$endpoint" 2>/dev/null || echo "{}"
}

# Function to get Prometheus data
prom_query() {
    local query=$1
    curl -s "$PROMETHEUS_URL/api/v1/query?query=$query" 2>/dev/null | jq -r '.data.result[0].value[1] // 0'
}

# Function to format bytes
format_bytes() {
    local bytes=$1
    if [ $bytes -lt 1024 ]; then
        echo "${bytes}B"
    elif [ $bytes -lt 1048576 ]; then
        echo "$((bytes/1024))KB"
    elif [ $bytes -lt 1073741824 ]; then
        echo "$((bytes/1048576))MB"
    else
        echo "$((bytes/1073741824))GB"
    fi
}

# Function to display terminal output
display_terminal() {
    clear
    echo "🔍 Phoenix Platform Monitor"
    echo "==========================="
    echo "Time: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
    
    # API Health
    echo -e "${BLUE}API Health:${NC}"
    api_status=$(curl -s -o /dev/null -w "%{http_code}" "$PHOENIX_API_URL/health")
    if [ "$api_status" = "200" ]; then
        echo -e "  Status: ${GREEN}✓ Healthy${NC}"
    else
        echo -e "  Status: ${RED}✗ Unhealthy (HTTP $api_status)${NC}"
    fi
    
    # Get API metrics
    api_metrics=$(api_get "/metrics" | grep -E "^phoenix_" | head -20)
    if [ -n "$api_metrics" ]; then
        requests_total=$(echo "$api_metrics" | grep "phoenix_api_requests_total" | awk '{sum+=$2} END {print sum}')
        requests_rate=$(prom_query "rate(phoenix_api_requests_total[1m])")
        echo "  Total Requests: ${requests_total:-0}"
        echo "  Request Rate: $(printf "%.2f" ${requests_rate:-0}) req/s"
    fi
    echo ""
    
    # Agents Status
    echo -e "${BLUE}Agents:${NC}"
    agents=$(api_get "/api/v1/agents" 2>/dev/null || echo "[]")
    agent_count=$(echo "$agents" | jq 'length // 0' 2>/dev/null || echo "0")
    
    if [ "$agent_count" -gt 0 ]; then
        echo "$agents" | jq -r '.[] | "  \(.host_id): \(.status)"' 2>/dev/null || echo "  No agent data"
    else
        echo "  No agents registered"
    fi
    echo ""
    
    # Experiments
    echo -e "${BLUE}Experiments:${NC}"
    experiments=$(api_get "/api/v1/experiments")
    exp_count=$(echo "$experiments" | jq 'length // 0' 2>/dev/null || echo "0")
    
    if [ "$exp_count" -gt 0 ]; then
        echo "  Total: $exp_count"
        echo "$experiments" | jq -r '.[] | "  - \(.name): \(.status)"' 2>/dev/null | head -5
        [ "$exp_count" -gt 5 ] && echo "  ... and $((exp_count-5)) more"
    else
        echo "  No experiments"
    fi
    echo ""
    
    # System Resources (if Prometheus available)
    if curl -s "$PROMETHEUS_URL/-/healthy" > /dev/null 2>&1; then
        echo -e "${BLUE}System Resources:${NC}"
        
        # CPU usage
        cpu_usage=$(prom_query "100 - (avg(rate(node_cpu_seconds_total{mode=\"idle\"}[1m])) * 100)")
        cpu_formatted=$(printf "%.1f" ${cpu_usage:-0})
        if (( $(echo "$cpu_usage > $ALERT_THRESHOLD_CPU" | bc -l) )); then
            echo -e "  CPU Usage: ${RED}${cpu_formatted}%${NC} ⚠️"
        else
            echo -e "  CPU Usage: ${GREEN}${cpu_formatted}%${NC}"
        fi
        
        # Memory usage
        mem_total=$(prom_query "node_memory_MemTotal_bytes")
        mem_free=$(prom_query "node_memory_MemFree_bytes")
        mem_used=$((mem_total - mem_free))
        mem_percent=$((mem_used * 100 / mem_total))
        
        if [ $mem_percent -gt $ALERT_THRESHOLD_MEM ]; then
            echo -e "  Memory: ${RED}$(format_bytes $mem_used) / $(format_bytes $mem_total) (${mem_percent}%)${NC} ⚠️"
        else
            echo -e "  Memory: ${GREEN}$(format_bytes $mem_used) / $(format_bytes $mem_total) (${mem_percent}%)${NC}"
        fi
        
        # Disk usage
        disk_usage=$(prom_query "100 - (node_filesystem_avail_bytes{mountpoint=\"/\"} / node_filesystem_size_bytes{mountpoint=\"/\"} * 100)")
        disk_formatted=$(printf "%.1f" ${disk_usage:-0})
        echo "  Disk Usage: ${disk_formatted}%"
    fi
    echo ""
    
    # Metrics Summary
    echo -e "${BLUE}Metrics Summary:${NC}"
    metric_count=$(docker exec phoenix-postgres psql -U phoenix -d phoenix -t -c "SELECT COUNT(*) FROM metric_cache;" 2>/dev/null | tr -d ' ' || echo "0")
    echo "  Cached Metrics: $metric_count"
    
    # Calculate cardinality reduction
    if [ -n "$api_metrics" ]; then
        baseline_cardinality=$(echo "$api_metrics" | grep "phoenix_baseline_cardinality" | awk '{print $2}')
        optimized_cardinality=$(echo "$api_metrics" | grep "phoenix_optimized_cardinality" | awk '{print $2}')
        
        if [ -n "$baseline_cardinality" ] && [ -n "$optimized_cardinality" ] && [ "$baseline_cardinality" -gt 0 ]; then
            reduction=$(echo "scale=2; (1 - $optimized_cardinality / $baseline_cardinality) * 100" | bc)
            echo "  Cardinality Reduction: ${reduction}%"
        fi
    fi
    
    echo ""
    echo "Press Ctrl+C to exit"
}

# Function to display JSON output
display_json() {
    # Collect all data
    api_health=$(curl -s -o /dev/null -w "%{http_code}" "$PHOENIX_API_URL/health")
    agents=$(api_get "/api/v1/agents")
    experiments=$(api_get "/api/v1/experiments")
    
    # Build JSON
    cat << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "api": {
    "url": "$PHOENIX_API_URL",
    "health": "$api_health"
  },
  "agents": $agents,
  "experiments": $experiments,
  "metrics": {
    "cached_count": $(docker exec phoenix-postgres psql -U phoenix -d phoenix -t -c "SELECT COUNT(*) FROM metric_cache;" 2>/dev/null | tr -d ' ' || echo "0")
  }
}
EOF
}

# Main monitoring loop
if [ "$OUTPUT_FORMAT" = "json" ]; then
    display_json
else
    if [ "$CONTINUOUS" = true ]; then
        while true; do
            display_terminal
            sleep $REFRESH_INTERVAL
        done
    else
        display_terminal
    fi
fi