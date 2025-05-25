#!/bin/bash
# Monitoring Dashboard Example
# This demonstrates how to create a real-time monitoring dashboard using Phoenix CLI

set -e

# Configuration
REFRESH_INTERVAL=${REFRESH_INTERVAL:-5}
NAMESPACE=${NAMESPACE:-"production"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to clear screen and move cursor
clear_screen() {
    clear
    tput cup 0 0
}

# Function to draw a line
draw_line() {
    printf '%*s\n' "${COLUMNS:-80}" '' | tr ' ' '='
}

# Function to format percentage with color
format_percentage() {
    local value=$1
    local threshold=$2
    local reverse=${3:-false}
    
    if [[ "$reverse" == "true" ]]; then
        # For metrics where lower is better (e.g., data loss)
        if (( $(echo "$value > $threshold" | bc -l) )); then
            echo -e "${RED}${value}%${NC}"
        else
            echo -e "${GREEN}${value}%${NC}"
        fi
    else
        # For metrics where higher is better (e.g., cost reduction)
        if (( $(echo "$value < $threshold" | bc -l) )); then
            echo -e "${RED}${value}%${NC}"
        else
            echo -e "${GREEN}${value}%${NC}"
        fi
    fi
}

# Function to draw progress bar
progress_bar() {
    local percent=$1
    local width=30
    local filled=$(echo "scale=0; $percent * $width / 100" | bc)
    local empty=$((width - filled))
    
    printf "["
    printf '%*s' "$filled" '' | tr ' ' '█'
    printf '%*s' "$empty" '' | tr ' ' '░'
    printf "]"
}

# Main monitoring loop
echo "Starting Phoenix Monitoring Dashboard..."
echo "Press Ctrl+C to exit"
sleep 2

while true; do
    clear_screen
    
    # Header
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    Phoenix Monitoring Dashboard                   ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "Namespace: $NAMESPACE"
    echo "Refresh: ${REFRESH_INTERVAL}s"
    echo
    
    # Get active experiments
    draw_line
    echo -e "${YELLOW}Active Experiments${NC}"
    draw_line
    
    EXPERIMENTS=$(phoenix experiment list --namespace "$NAMESPACE" --status running --output json 2>/dev/null || echo '{"experiments":[]}')
    EXPERIMENT_COUNT=$(echo "$EXPERIMENTS" | jq '.experiments | length')
    
    if [ "$EXPERIMENT_COUNT" -eq 0 ]; then
        echo "No active experiments"
    else
        # Table header
        printf "%-20s %-15s %-10s %-15s %-15s %-10s\n" \
            "NAME" "STATUS" "DURATION" "COST REDUCTION" "DATA LOSS" "PROGRESS"
        printf '%*s\n' "${COLUMNS:-80}" '' | tr ' ' '-'
        
        # Process each experiment
        echo "$EXPERIMENTS" | jq -r '.experiments[] | @json' | while read -r exp; do
            EXP_ID=$(echo "$exp" | jq -r '.id')
            NAME=$(echo "$exp" | jq -r '.name' | cut -c1-20)
            STATUS=$(echo "$exp" | jq -r '.status')
            START_TIME=$(echo "$exp" | jq -r '.start_time')
            
            # Get detailed metrics
            METRICS=$(phoenix experiment metrics "$EXP_ID" --output json 2>/dev/null || echo '{}')
            COST_REDUCTION=$(echo "$METRICS" | jq -r '.summary.cost_reduction_percent // 0')
            DATA_LOSS=$(echo "$METRICS" | jq -r '.summary.data_loss_percent // 0')
            PROGRESS=$(echo "$METRICS" | jq -r '.summary.progress_percent // 0')
            
            # Calculate duration
            if [ "$START_TIME" != "null" ]; then
                START_EPOCH=$(date -d "$START_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "$START_TIME" +%s 2>/dev/null || echo 0)
                NOW_EPOCH=$(date +%s)
                DURATION_SECS=$((NOW_EPOCH - START_EPOCH))
                DURATION=$(printf "%02d:%02d:%02d" $((DURATION_SECS/3600)) $((DURATION_SECS%3600/60)) $((DURATION_SECS%60)))
            else
                DURATION="--:--:--"
            fi
            
            # Format status with color
            case "$STATUS" in
                "running") STATUS_COLOR="${GREEN}$STATUS${NC}" ;;
                "failed") STATUS_COLOR="${RED}$STATUS${NC}" ;;
                *) STATUS_COLOR="${YELLOW}$STATUS${NC}" ;;
            esac
            
            # Print row
            printf "%-20s %-25s %-10s %-25s %-25s " \
                "$NAME" \
                "$STATUS_COLOR" \
                "$DURATION" \
                "$(format_percentage "$COST_REDUCTION" 20)" \
                "$(format_percentage "$DATA_LOSS" 2 true)"
            
            progress_bar "$PROGRESS"
            echo
        done
    fi
    
    # Pipeline deployments section
    echo
    draw_line
    echo -e "${YELLOW}Active Pipeline Deployments${NC}"
    draw_line
    
    DEPLOYMENTS=$(phoenix pipeline deployments list --namespace "$NAMESPACE" --output json 2>/dev/null || echo '{"deployments":[]}')
    DEPLOYMENT_COUNT=$(echo "$DEPLOYMENTS" | jq '.deployments | length')
    
    if [ "$DEPLOYMENT_COUNT" -eq 0 ]; then
        echo "No active deployments"
    else
        printf "%-25s %-20s %-15s %-20s\n" \
            "NAME" "TEMPLATE" "STATUS" "LAST UPDATED"
        printf '%*s\n' "${COLUMNS:-80}" '' | tr ' ' '-'
        
        echo "$DEPLOYMENTS" | jq -r '.deployments[] | @json' | while read -r dep; do
            NAME=$(echo "$dep" | jq -r '.name' | cut -c1-25)
            TEMPLATE=$(echo "$dep" | jq -r '.template' | cut -c1-20)
            STATUS=$(echo "$dep" | jq -r '.status')
            UPDATED=$(echo "$dep" | jq -r '.updated_at' | cut -c1-19)
            
            # Format status
            case "$STATUS" in
                "active") STATUS_COLOR="${GREEN}$STATUS${NC}" ;;
                "failed") STATUS_COLOR="${RED}$STATUS${NC}" ;;
                *) STATUS_COLOR="${YELLOW}$STATUS${NC}" ;;
            esac
            
            printf "%-25s %-20s %-25s %-20s\n" \
                "$NAME" "$TEMPLATE" "$STATUS_COLOR" "$UPDATED"
        done
    fi
    
    # System metrics section
    echo
    draw_line
    echo -e "${YELLOW}System Metrics${NC}"
    draw_line
    
    # Get aggregated metrics across all experiments
    TOTAL_COST_SAVED=0
    TOTAL_DATA_PROCESSED=0
    ACTIVE_COLLECTORS=0
    
    if [ "$EXPERIMENT_COUNT" -gt 0 ]; then
        while IFS= read -r exp_id; do
            METRICS=$(phoenix experiment metrics "$exp_id" --output json 2>/dev/null || echo '{}')
            COST_SAVED=$(echo "$METRICS" | jq -r '.summary.estimated_monthly_savings // 0')
            DATA_PROCESSED=$(echo "$METRICS" | jq -r '.summary.data_processed_gb // 0')
            COLLECTORS=$(echo "$METRICS" | jq -r '.summary.active_collectors // 0')
            
            TOTAL_COST_SAVED=$(echo "$TOTAL_COST_SAVED + $COST_SAVED" | bc)
            TOTAL_DATA_PROCESSED=$(echo "$TOTAL_DATA_PROCESSED + $DATA_PROCESSED" | bc)
            ACTIVE_COLLECTORS=$((ACTIVE_COLLECTORS + COLLECTORS))
        done < <(echo "$EXPERIMENTS" | jq -r '.experiments[].id')
    fi
    
    printf "%-30s: $%'.2f/month\n" "Estimated Cost Savings" "$TOTAL_COST_SAVED"
    printf "%-30s: %'.2f GB\n" "Data Processed" "$TOTAL_DATA_PROCESSED"
    printf "%-30s: %d\n" "Active Collectors" "$ACTIVE_COLLECTORS"
    printf "%-30s: %d\n" "Running Experiments" "$EXPERIMENT_COUNT"
    printf "%-30s: %d\n" "Active Deployments" "$DEPLOYMENT_COUNT"
    
    # Recent events
    echo
    draw_line
    echo -e "${YELLOW}Recent Events (Last 5)${NC}"
    draw_line
    
    # In a real implementation, this would query an events API
    # For now, we'll show experiment state changes
    echo "$EXPERIMENTS" | jq -r '.experiments[] | select(.last_event) | .last_event' 2>/dev/null | head -5 || echo "No recent events"
    
    # Footer
    echo
    draw_line
    echo "Last refresh: $(date '+%H:%M:%S') | Next refresh in ${REFRESH_INTERVAL}s | Press Ctrl+C to exit"
    
    # Wait before next refresh
    sleep "$REFRESH_INTERVAL"
done