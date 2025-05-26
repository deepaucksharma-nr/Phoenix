#!/bin/bash
# Phoenix Agent Installation Script for Linux VMs

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

# Variables
PHOENIX_API_URL="${PHOENIX_API_URL:-http://phoenix-api:8080}"
AGENT_VERSION="${AGENT_VERSION:-latest}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/phoenix-agent"
LOG_DIR="/var/log/phoenix-agent"

echo -e "${GREEN}Installing Phoenix Agent${NC}"

# Create directories
echo "Creating directories..."
mkdir -p "$CONFIG_DIR" "$LOG_DIR"

# Download agent binary
echo "Downloading Phoenix Agent..."
if [[ -f "./phoenix-agent" ]]; then
    # Use local binary if available
    cp ./phoenix-agent "$INSTALL_DIR/phoenix-agent"
else
    # Download from releases (adjust URL as needed)
    curl -L -o "$INSTALL_DIR/phoenix-agent" \
        "https://github.com/phoenix/platform/releases/download/${AGENT_VERSION}/phoenix-agent-linux-amd64"
fi

chmod +x "$INSTALL_DIR/phoenix-agent"

# Download OTel Collector
echo "Downloading OTel Collector..."
if ! command -v otelcol-contrib &> /dev/null; then
    curl -L -o /tmp/otelcol-contrib.tar.gz \
        "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.95.0/otelcol-contrib_0.95.0_linux_amd64.tar.gz"
    tar -xzf /tmp/otelcol-contrib.tar.gz -C /tmp
    mv /tmp/otelcol-contrib "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/otelcol-contrib"
    rm -f /tmp/otelcol-contrib.tar.gz
fi

# Create environment file
echo "Creating configuration..."
cat > "$CONFIG_DIR/environment" <<EOF
# Phoenix Agent Configuration
PHOENIX_API_URL=$PHOENIX_API_URL
POLL_INTERVAL=15s
LOG_LEVEL=info
CONFIG_DIR=$CONFIG_DIR
PUSHGATEWAY_URL=http://prometheus-pushgateway:9091
EOF

# Get hostname for agent ID
HOSTNAME=$(hostname)
echo "PHOENIX_HOST_ID=$HOSTNAME" >> "$CONFIG_DIR/environment"

# Install systemd service
echo "Installing systemd service..."
cp phoenix-agent.service /etc/systemd/system/
systemctl daemon-reload

# Enable and start service
echo "Starting Phoenix Agent..."
systemctl enable phoenix-agent
systemctl start phoenix-agent

# Check status
sleep 2
if systemctl is-active --quiet phoenix-agent; then
    echo -e "${GREEN}✓ Phoenix Agent installed and running successfully${NC}"
    echo ""
    echo "Configuration file: $CONFIG_DIR/environment"
    echo "Logs: journalctl -u phoenix-agent -f"
    echo "Status: systemctl status phoenix-agent"
else
    echo -e "${RED}✗ Phoenix Agent failed to start${NC}"
    echo "Check logs: journalctl -u phoenix-agent -xe"
    exit 1
fi

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Edit $CONFIG_DIR/environment to configure the agent"
echo "2. Restart the service: systemctl restart phoenix-agent"
echo "3. Create experiments via Phoenix API"