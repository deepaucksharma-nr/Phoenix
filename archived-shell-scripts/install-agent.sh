#!/usr/bin/env bash
#
# Phoenix Agent Installation Script
# This script installs the Phoenix agent on edge nodes
#
# Usage: curl -fsSL https://phoenix.my-org.com/install-agent.sh | sudo bash
#

set -euo pipefail

# Configuration
API_URL="${PHOENIX_API_URL:-https://phoenix.my-org.com}"
TOKEN="${PHOENIX_TOKEN:-BOOTSTRAP_TOKEN_CHANGE_ME}"
AGENT_VERSION="${PHOENIX_AGENT_VERSION:-latest}"
INSTALL_DIR="/opt/phoenix-agent"
SERVICE_USER="phoenixagent"
ARCH=$(uname -m)
USE_NRDOT="${USE_NRDOT:-false}"
NRDOT_VERSION="${NRDOT_VERSION:-latest}"
NEW_RELIC_LICENSE_KEY="${NEW_RELIC_LICENSE_KEY:-}"
NEW_RELIC_OTLP_ENDPOINT="${NEW_RELIC_OTLP_ENDPOINT:-otlp.nr-data.net:4317}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root"
        exit 1
    fi
}

# Detect OS and architecture
detect_system() {
    local os=""
    local arch_suffix=""
    
    # Detect OS
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        os=$ID
    elif [[ -f /etc/redhat-release ]]; then
        os="rhel"
    elif [[ -f /etc/debian_version ]]; then
        os="debian"
    else
        error "Unsupported operating system"
        exit 1
    fi
    
    # Map architecture
    case $ARCH in
        x86_64)
            arch_suffix="amd64"
            ;;
        aarch64|arm64)
            arch_suffix="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    BINARY_NAME="phoenix-agent-linux-${arch_suffix}"
    log "Detected system: $os ($ARCH -> $arch_suffix)"
}

# Install dependencies
install_dependencies() {
    log "Installing dependencies..."
    
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update -qq
        apt-get install -y -qq curl wget ca-certificates
    elif command -v yum >/dev/null 2>&1; then
        yum install -y -q curl wget ca-certificates
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y -q curl wget ca-certificates
    else
        warning "Could not install dependencies automatically"
    fi
}

# Create service user
create_user() {
    if ! id "$SERVICE_USER" >/dev/null 2>&1; then
        log "Creating service user: $SERVICE_USER"
        useradd -r -s /bin/false -d /nonexistent -c "Phoenix Agent" "$SERVICE_USER"
    else
        log "Service user already exists: $SERVICE_USER"
    fi
}

# Download and install agent binary
install_agent() {
    log "Installing Phoenix agent..."
    
    # Create directories
    mkdir -p "$INSTALL_DIR"
    mkdir -p /etc/phoenix-agent
    mkdir -p /var/log/phoenix-agent
    mkdir -p /var/lib/phoenix-agent
    
    # Download binary
    local download_url="${API_URL}/downloads/${BINARY_NAME}"
    if [[ "$AGENT_VERSION" != "latest" ]]; then
        download_url="${API_URL}/downloads/${AGENT_VERSION}/${BINARY_NAME}"
    fi
    
    log "Downloading agent from: $download_url"
    if ! curl -fsSL "$download_url" -o "${INSTALL_DIR}/phoenix-agent"; then
        error "Failed to download agent binary"
        exit 1
    fi
    
    chmod +x "${INSTALL_DIR}/phoenix-agent"
    
    # Download OTel Collector (standard or NRDOT)
    if [[ "$USE_NRDOT" == "true" ]]; then
        install_nrdot_collector
    else
        install_otel_collector
    fi
    
    # Set ownership
    chown -R "$SERVICE_USER:$SERVICE_USER" /var/log/phoenix-agent
    chown -R "$SERVICE_USER:$SERVICE_USER" /var/lib/phoenix-agent
}

# Install standard OTel collector
install_otel_collector() {
    log "Installing OpenTelemetry Collector Contrib..."
    
    local otel_version="0.95.0"
    local otel_url="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${otel_version}/otelcol-contrib_${otel_version}_linux_${arch_suffix}.tar.gz"
    
    if ! command -v otelcol-contrib >/dev/null 2>&1; then
        log "Downloading OTel Collector from: $otel_url"
        wget -q "$otel_url" -O /tmp/otelcol-contrib.tar.gz
        tar -xzf /tmp/otelcol-contrib.tar.gz -C /usr/local/bin/ otelcol-contrib
        chmod +x /usr/local/bin/otelcol-contrib
        rm /tmp/otelcol-contrib.tar.gz
        log "OTel Collector installed successfully"
    else
        log "OTel Collector already installed"
    fi
}

# Install NRDOT collector
install_nrdot_collector() {
    log "Installing New Relic NRDOT Collector..."
    
    if [[ -z "$NEW_RELIC_LICENSE_KEY" ]]; then
        error "NEW_RELIC_LICENSE_KEY must be set when using NRDOT"
        exit 1
    fi
    
    local nrdot_url=""
    if [[ "$NRDOT_VERSION" == "latest" ]]; then
        nrdot_url="https://github.com/newrelic/nrdot-collector-releases/releases/latest/download/nrdot-collector-host_linux_${arch_suffix}.tar.gz"
    else
        nrdot_url="https://github.com/newrelic/nrdot-collector-releases/releases/download/v${NRDOT_VERSION}/nrdot-collector-host_linux_${arch_suffix}.tar.gz"
    fi
    
    log "Downloading NRDOT from: $nrdot_url"
    if ! wget -q "$nrdot_url" -O /tmp/nrdot.tar.gz; then
        error "Failed to download NRDOT collector"
        exit 1
    fi
    
    # Extract NRDOT
    tar -xzf /tmp/nrdot.tar.gz -C /tmp/
    
    # Find and move the binary
    if [[ -f /tmp/nrdot-collector-host ]]; then
        mv /tmp/nrdot-collector-host /usr/local/bin/nrdot
    elif [[ -f /tmp/nrdot ]]; then
        mv /tmp/nrdot /usr/local/bin/nrdot
    else
        error "Could not find NRDOT binary in archive"
        exit 1
    fi
    
    chmod +x /usr/local/bin/nrdot
    rm -f /tmp/nrdot.tar.gz
    
    log "NRDOT Collector installed successfully"
}

# Create configuration file
create_config() {
    log "Creating configuration..."
    
    cat > /etc/phoenix-agent/config.yaml << EOF
# Phoenix Agent Configuration
api:
  url: ${API_URL}
  token: ${TOKEN}
  tls_skip_verify: false
  
agent:
  id: $(hostname)-$(date +%s)
  poll_interval: 15s
  heartbeat_interval: 30s
  
metrics:
  pushgateway_url: ${API_URL/https/http}:9091
  collection_interval: 15s
  
logging:
  level: info
  file: /var/log/phoenix-agent/agent.log
  max_size: 100
  max_backups: 3
  max_age: 30
  
storage:
  data_dir: /var/lib/phoenix-agent
  
# Resource limits
resources:
  max_collectors: 2
  max_memory_mb: 512
  max_cpu_percent: 10

# Collector configuration
collector:
  type: ${USE_NRDOT}
  newrelic:
    license_key: ${NEW_RELIC_LICENSE_KEY}
    otlp_endpoint: ${NEW_RELIC_OTLP_ENDPOINT}
EOF
    
    chmod 600 /etc/phoenix-agent/config.yaml
    chown "$SERVICE_USER:$SERVICE_USER" /etc/phoenix-agent/config.yaml
}

# Create systemd service
create_systemd_service() {
    log "Creating systemd service..."
    
    cat > /etc/systemd/system/phoenix-agent.service << EOF
[Unit]
Description=Phoenix Optimization Agent
Documentation=https://phoenix.my-org.com/docs
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
ExecStartPre=/bin/sleep 10
ExecStart=${INSTALL_DIR}/phoenix-agent \\
    --config /etc/phoenix-agent/config.yaml \\
    --api ${API_URL} \\
    --token ${TOKEN} \\
    --pushgateway ${API_URL/https/http}:9091

# Restart configuration
Restart=always
RestartSec=30
StartLimitInterval=600
StartLimitBurst=5

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/phoenix-agent /var/lib/phoenix-agent
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true
RestrictSUIDSGID=true
MemoryLimit=512M
CPUQuota=10%

# Environment
Environment="GOMAXPROCS=2"
Environment="PHOENIX_AGENT_ID=$(hostname)"

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=phoenix-agent

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd
    systemctl daemon-reload
}

# Start and enable service
start_service() {
    log "Starting Phoenix agent service..."
    
    systemctl enable phoenix-agent.service
    systemctl start phoenix-agent.service
    
    # Wait a moment for service to start
    sleep 3
    
    # Check status
    if systemctl is-active --quiet phoenix-agent.service; then
        log "Phoenix agent is running successfully"
    else
        error "Phoenix agent failed to start"
        systemctl status phoenix-agent.service --no-pager
        exit 1
    fi
}

# Create uninstall script
create_uninstall_script() {
    cat > "${INSTALL_DIR}/uninstall.sh" << 'EOF'
#!/bin/bash
# Phoenix Agent Uninstall Script

echo "Stopping Phoenix agent..."
systemctl stop phoenix-agent.service 2>/dev/null
systemctl disable phoenix-agent.service 2>/dev/null

echo "Removing files..."
rm -f /etc/systemd/system/phoenix-agent.service
rm -rf /opt/phoenix-agent
rm -rf /etc/phoenix-agent
rm -rf /var/log/phoenix-agent
rm -rf /var/lib/phoenix-agent

echo "Removing user..."
userdel phoenixagent 2>/dev/null

echo "Reloading systemd..."
systemctl daemon-reload

echo "Phoenix agent has been uninstalled"
EOF
    
    chmod +x "${INSTALL_DIR}/uninstall.sh"
}

# Print success message
print_success() {
    echo
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}       Phoenix Agent Installation Completed Successfully!       ${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo
    echo "Agent Status:"
    echo "  Service: phoenix-agent.service"
    echo "  Config:  /etc/phoenix-agent/config.yaml"
    echo "  Logs:    journalctl -u phoenix-agent -f"
    echo
    echo "Useful Commands:"
    echo "  Check status:  systemctl status phoenix-agent"
    echo "  View logs:     journalctl -u phoenix-agent -f"
    echo "  Restart:       systemctl restart phoenix-agent"
    echo "  Uninstall:     ${INSTALL_DIR}/uninstall.sh"
    echo
    echo "The agent should appear in the Phoenix UI within 30 seconds."
    echo
}

# Main installation flow
main() {
    log "Starting Phoenix agent installation..."
    
    check_root
    detect_system
    install_dependencies
    create_user
    install_agent
    create_config
    create_systemd_service
    start_service
    create_uninstall_script
    print_success
}

# Run main function
main "$@"