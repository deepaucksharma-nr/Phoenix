#!/usr/bin/env bash
#
# Phoenix Scaling Configuration Validator
# Validates that scaling rules are properly configured and monitoring is working
#

set -euo pipefail

# Configuration
PHOENIX_DIR="${PHOENIX_DIR:-/opt/phoenix}"
SCALING_RULES="$PHOENIX_DIR/config/scaling-rules.yml"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Validation results
PASS=0
WARN=0
FAIL=0

# Helper functions
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASS++))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARN++))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAIL++))
}

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Header
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}           Phoenix Scaling Configuration Validator              ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo

# 1. Check scaling rules file exists
echo "Checking Scaling Configuration..."
echo "────────────────────────────────"

if [[ -f "$SCALING_RULES" ]]; then
    check_pass "Scaling rules file exists"
    
    # Validate YAML syntax
    if command -v yq >/dev/null 2>&1; then
        if yq eval '.' "$SCALING_RULES" >/dev/null 2>&1; then
            check_pass "Scaling rules YAML is valid"
        else
            check_fail "Scaling rules YAML is invalid"
        fi
    else
        check_warn "yq not installed, skipping YAML validation"
    fi
else
    check_fail "Scaling rules file not found at $SCALING_RULES"
fi

echo

# 2. Check if auto-scale monitor is running
echo "Checking Auto-Scale Monitor..."
echo "─────────────────────────────"

if systemctl is-active --quiet phoenix-scale-monitor 2>/dev/null || \
   systemctl is-active --quiet phoenix-autoscale 2>/dev/null; then
    check_pass "Auto-scale monitor service is running"
    
    # Check recent logs
    if journalctl -u phoenix-scale-monitor -n 1 --no-pager >/dev/null 2>&1 || \
       journalctl -u phoenix-autoscale -n 1 --no-pager >/dev/null 2>&1; then
        check_pass "Auto-scale monitor logs are accessible"
    else
        check_warn "Cannot access auto-scale monitor logs"
    fi
else
    check_fail "Auto-scale monitor service is not running"
    info "Start with: systemctl start phoenix-autoscale"
fi

echo

# 3. Check Prometheus queries
echo "Checking Prometheus Metrics..."
echo "─────────────────────────────"

# Test if Prometheus is accessible
if curl -s http://localhost:9090/api/v1/query?query=up >/dev/null 2>&1; then
    check_pass "Prometheus is accessible"
    
    # Test each monitoring query from scaling rules
    queries=(
        "phoenix_agent_last_heartbeat"
        "phoenix_processed_series"
        "phoenix_api_request_duration_seconds_bucket"
    )
    
    for query in "${queries[@]}"; do
        result=$(curl -s "http://localhost:9090/api/v1/query?query=${query}" | \
                 jq -r '.status' 2>/dev/null || echo "error")
        
        if [[ "$result" == "success" ]]; then
            check_pass "Metric '$query' is available"
        else
            check_warn "Metric '$query' not found (may appear when agents connect)"
        fi
    done
else
    check_fail "Prometheus is not accessible"
    info "Check if Prometheus is running: docker-compose ps prometheus"
fi

echo

# 4. Check current resource usage against thresholds
echo "Checking Current Resource Usage..."
echo "─────────────────────────────────"

# CPU Usage
cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1 2>/dev/null || echo "0")
if (( $(echo "$cpu_usage < 70" | bc -l) )); then
    check_pass "CPU usage is healthy: ${cpu_usage}% (threshold: 70%)"
elif (( $(echo "$cpu_usage < 85" | bc -l) )); then
    check_warn "CPU usage is elevated: ${cpu_usage}% (warning: 70%, critical: 85%)"
else
    check_fail "CPU usage is critical: ${cpu_usage}% (critical: 85%)"
fi

# Memory Usage
mem_usage=$(free | grep Mem | awk '{print ($2-$7)/$2 * 100}' 2>/dev/null || echo "0")
mem_usage_int=${mem_usage%.*}
if (( mem_usage_int < 80 )); then
    check_pass "Memory usage is healthy: ${mem_usage_int}% (threshold: 80%)"
elif (( mem_usage_int < 90 )); then
    check_warn "Memory usage is elevated: ${mem_usage_int}% (warning: 80%, critical: 90%)"
else
    check_fail "Memory usage is critical: ${mem_usage_int}% (critical: 90%)"
fi

# Disk Usage
disk_usage=$(df -h "$PHOENIX_DIR" 2>/dev/null | tail -1 | awk '{gsub("%",""); print $5}' || echo "0")
if (( disk_usage < 80 )); then
    check_pass "Disk usage is healthy: ${disk_usage}% (threshold: 80%)"
elif (( disk_usage < 90 )); then
    check_warn "Disk usage is elevated: ${disk_usage}% (warning: 80%, critical: 90%)"
else
    check_fail "Disk usage is critical: ${disk_usage}% (critical: 90%)"
fi

echo

# 5. Check scaling readiness
echo "Checking Scaling Readiness..."
echo "────────────────────────────"

# Check if backups are recent
if [[ -d "$PHOENIX_DIR/backups" ]]; then
    latest_backup=$(find "$PHOENIX_DIR/backups" -name "*.dump" -type f -printf '%T@\n' 2>/dev/null | \
                    sort -n | tail -1 || echo "0")
    if [[ -n "$latest_backup" && "$latest_backup" != "0" ]]; then
        backup_age=$(( ($(date +%s) - ${latest_backup%.*}) / 3600 ))
        if (( backup_age < 24 )); then
            check_pass "Recent backup found (${backup_age} hours old)"
        else
            check_warn "Backup is ${backup_age} hours old (recommend < 24 hours)"
        fi
    else
        check_fail "No backups found"
    fi
else
    check_fail "Backup directory not found"
fi

# Check if scaling scripts exist
scaling_scripts=(
    "$PHOENIX_DIR/scripts/backup.sh"
    "$PHOENIX_DIR/scripts/restore.sh"
    "$PHOENIX_DIR/scripts/auto-scale-monitor.sh"
)

for script in "${scaling_scripts[@]}"; do
    if [[ -x "$script" ]]; then
        check_pass "$(basename "$script") is executable"
    else
        check_fail "$(basename "$script") not found or not executable"
    fi
done

echo

# 6. Scaling recommendations based on current state
echo "Scaling Recommendations..."
echo "─────────────────────────"

# Determine current phase
agent_count=$(curl -s http://localhost:9090/api/v1/query?query='count(phoenix_agent_info)' | \
              jq -r '.data.result[0].value[1] // "0"' 2>/dev/null || echo "0")

info "Current agent count: $agent_count"

if (( agent_count < 100 )); then
    info "Phase: Initial deployment (0-100 agents)"
    info "Current VM size is adequate (t3.medium)"
elif (( agent_count < 150 )); then
    info "Phase: Growth phase (100-150 agents)"
    info "Monitor CPU/Memory, prepare for vertical scaling"
elif (( agent_count < 200 )); then
    info "Phase: Scaling phase (150-200 agents)"
    info "Consider component separation (RDS + dedicated Prometheus)"
else
    info "Phase: Horizontal scaling needed (200+ agents)"
    info "Time to deploy Kubernetes version"
fi

echo

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo "Validation Summary"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "Passed: ${GREEN}$PASS${NC}"
echo -e "Warnings: ${YELLOW}$WARN${NC}"
echo -e "Failed: ${RED}$FAIL${NC}"
echo

if [[ $FAIL -eq 0 ]]; then
    if [[ $WARN -eq 0 ]]; then
        echo -e "${GREEN}✓ All scaling configurations are properly set up!${NC}"
    else
        echo -e "${YELLOW}⚠ Scaling is configured with some warnings.${NC}"
    fi
    exit 0
else
    echo -e "${RED}✗ Scaling configuration has issues that need attention.${NC}"
    exit 1
fi