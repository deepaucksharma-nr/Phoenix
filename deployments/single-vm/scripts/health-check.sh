#!/usr/bin/env bash
#
# Phoenix Health Check Script
# Verifies all components are running correctly
#

set -euo pipefail

# Configuration
PHOENIX_DIR="${PHOENIX_DIR:-/opt/phoenix}"
PUBLIC_URL="${PHX_PUBLIC_URL:-https://localhost}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Status tracking
CHECKS_PASSED=0
CHECKS_FAILED=0

# Helper functions
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((CHECKS_PASSED++))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((CHECKS_FAILED++))
}

check_warn() {
    echo -e "${YELLOW}!${NC} $1"
}

# Header
echo "═══════════════════════════════════════════════════════════════"
echo "                 Phoenix Health Check Report                    "
echo "═══════════════════════════════════════════════════════════════"
echo

# Check Docker services
echo "Checking Docker Services..."
echo "─────────────────────────"

cd "$PHOENIX_DIR" 2>/dev/null || {
    check_fail "Phoenix directory not found: $PHOENIX_DIR"
    exit 1
}

# Check each service
for service in db pushgateway prometheus api; do
    if docker-compose ps | grep -E "${service}.*Up.*healthy|${service}.*Up.*running" >/dev/null 2>&1; then
        check_pass "Service $service is running"
    else
        check_fail "Service $service is not running"
        docker-compose ps | grep "$service" || echo "  Service not found"
    fi
done

echo

# Check API endpoints
echo "Checking API Endpoints..."
echo "────────────────────────"

# Health endpoint
if curl -fsS "${PUBLIC_URL}/health" -k >/dev/null 2>&1; then
    check_pass "API health endpoint responding"
else
    check_fail "API health endpoint not responding"
fi

# Metrics endpoint
if curl -fsS "http://localhost:8080/metrics" >/dev/null 2>&1; then
    check_pass "API metrics endpoint responding"
else
    check_fail "API metrics endpoint not responding"
fi

# WebSocket endpoint
if curl -fsS "http://localhost:8081" -H "Upgrade: websocket" 2>&1 | grep -q "Bad Request"; then
    check_pass "WebSocket endpoint available"
else
    check_warn "WebSocket endpoint check inconclusive"
fi

echo

# Check Prometheus
echo "Checking Prometheus..."
echo "─────────────────────"

# Prometheus API
if curl -fsS "http://localhost:9090/api/v1/query?query=up" >/dev/null 2>&1; then
    check_pass "Prometheus API responding"
    
    # Check targets
    targets=$(curl -fsS "http://localhost:9090/api/v1/targets" 2>/dev/null | jq -r '.data.activeTargets | length' || echo "0")
    if [[ $targets -gt 0 ]]; then
        check_pass "Prometheus has $targets active targets"
    else
        check_warn "Prometheus has no active targets"
    fi
else
    check_fail "Prometheus API not responding"
fi

# Pushgateway
if curl -fsS "http://localhost:9091/metrics" >/dev/null 2>&1; then
    check_pass "Pushgateway responding"
    
    # Check for Phoenix metrics
    if curl -fsS "http://localhost:9091/metrics" 2>/dev/null | grep -q "phoenix_"; then
        check_pass "Phoenix metrics found in Pushgateway"
    else
        check_warn "No Phoenix metrics in Pushgateway yet"
    fi
else
    check_fail "Pushgateway not responding"
fi

echo

# Check Database
echo "Checking Database..."
echo "───────────────────"

if docker-compose exec -T db pg_isready -U phoenix >/dev/null 2>&1; then
    check_pass "PostgreSQL is ready"
    
    # Check tables
    tables=$(docker-compose exec -T db psql -U phoenix -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | tr -d ' ')
    if [[ $tables -gt 0 ]]; then
        check_pass "Database has $tables tables"
    else
        check_fail "Database has no tables - run migrations"
    fi
    
    # Check connections
    connections=$(docker-compose exec -T db psql -U phoenix -t -c "SELECT COUNT(*) FROM pg_stat_activity WHERE state='active';" 2>/dev/null | tr -d ' ')
    check_pass "Database has $connections active connections"
else
    check_fail "PostgreSQL is not ready"
fi

echo

# Check Agents
echo "Checking Agents..."
echo "─────────────────"

# Query for registered agents
agent_count=$(curl -fsS "${PUBLIC_URL}/api/v1/agents" -k 2>/dev/null | jq -r '.agents | length' || echo "0")
if [[ $agent_count -gt 0 ]]; then
    check_pass "Found $agent_count registered agents"
    
    # Check agent heartbeats
    recent_heartbeats=$(curl -fsS "http://localhost:9090/api/v1/query?query=phoenix_agent_last_heartbeat" 2>/dev/null | jq -r '.data.result | length' || echo "0")
    if [[ $recent_heartbeats -gt 0 ]]; then
        check_pass "$recent_heartbeats agents with recent heartbeats"
    else
        check_warn "No recent agent heartbeats"
    fi
else
    check_warn "No agents registered yet"
fi

echo

# Check Resources
echo "Checking Resources..."
echo "────────────────────"

# Memory usage
total_memory=$(docker stats --no-stream --format "table {{.Container}}\t{{.MemUsage}}" | tail -n +2)
echo "$total_memory" | while IFS= read -r line; do
    container=$(echo "$line" | awk '{print $1}')
    memory=$(echo "$line" | awk '{print $2}')
    echo "  $container: $memory"
done

# Disk usage
echo
echo "Disk Usage:"
df -h "$PHOENIX_DIR" | tail -n 1 | awk '{print "  Phoenix data: " $3 " / " $2 " (" $5 " used)"}'

echo

# Check TLS
echo "Checking TLS Configuration..."
echo "───────────────────────────"

if [[ -f "$PHOENIX_DIR/tls/fullchain.pem" ]] && [[ -f "$PHOENIX_DIR/tls/privkey.pem" ]]; then
    check_pass "TLS certificates found"
    
    # Check expiry
    expiry=$(openssl x509 -in "$PHOENIX_DIR/tls/fullchain.pem" -noout -enddate 2>/dev/null | cut -d= -f2)
    if [[ -n "$expiry" ]]; then
        expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$expiry" +%s 2>/dev/null || echo "0")
        now_epoch=$(date +%s)
        days_left=$(( (expiry_epoch - now_epoch) / 86400 ))
        
        if [[ $days_left -gt 30 ]]; then
            check_pass "TLS certificate valid for $days_left days"
        elif [[ $days_left -gt 0 ]]; then
            check_warn "TLS certificate expires in $days_left days"
        else
            check_fail "TLS certificate has expired!"
        fi
    fi
else
    check_warn "TLS certificates not found - running in HTTP mode"
fi

echo

# Check Backups
echo "Checking Backups..."
echo "──────────────────"

if [[ -d "$PHOENIX_DIR/backups" ]]; then
    latest_backup=$(ls -t "$PHOENIX_DIR/backups"/*.dump 2>/dev/null | head -1)
    if [[ -n "$latest_backup" ]]; then
        backup_age=$(( ($(date +%s) - $(stat -f%m "$latest_backup" 2>/dev/null || stat -c%Y "$latest_backup")) / 3600 ))
        if [[ $backup_age -lt 48 ]]; then
            check_pass "Latest backup is $backup_age hours old"
        else
            check_warn "Latest backup is $backup_age hours old"
        fi
    else
        check_warn "No backups found"
    fi
else
    check_warn "Backup directory not found"
fi

echo

# Summary
echo "═══════════════════════════════════════════════════════════════"
echo "                           Summary                              "
echo "═══════════════════════════════════════════════════════════════"
echo
echo -e "Checks passed: ${GREEN}$CHECKS_PASSED${NC}"
echo -e "Checks failed: ${RED}$CHECKS_FAILED${NC}"
echo

if [[ $CHECKS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All systems operational!${NC}"
    exit 0
else
    echo -e "${RED}Some checks failed. Please investigate.${NC}"
    echo
    echo "Common fixes:"
    echo "  - Restart services: cd $PHOENIX_DIR && docker-compose restart"
    echo "  - View logs: cd $PHOENIX_DIR && docker-compose logs"
    echo "  - Check configuration: cat $PHOENIX_DIR/.env"
    exit 1
fi