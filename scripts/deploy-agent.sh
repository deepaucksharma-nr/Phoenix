#!/bin/bash
# Deploy Phoenix agent to edge nodes

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🤖 Phoenix Agent Deployment"
echo "==========================="
echo ""

# Configuration
PHOENIX_API_URL="${PHOENIX_API_URL:-}"
PHOENIX_AGENT_HOST_ID="${PHOENIX_AGENT_HOST_ID:-$(hostname)}"
PHOENIX_CONFIG_DIR="${PHOENIX_CONFIG_DIR:-/etc/phoenix}"
PHOENIX_BIN_DIR="${PHOENIX_BIN_DIR:-/usr/local/bin}"
INSTALL_OTEL="${INSTALL_OTEL:-true}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --api-url)
            PHOENIX_API_URL="$2"
            shift 2
            ;;
        --host-id)
            PHOENIX_AGENT_HOST_ID="$2"
            shift 2
            ;;
        --skip-otel)
            INSTALL_OTEL=false
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--api-url URL] [--host-id ID] [--skip-otel]"
            exit 1
            ;;
    esac
done

# Validate inputs
if [ -z "$PHOENIX_API_URL" ]; then
    echo -e "${RED}Error: Phoenix API URL not specified${NC}"
    echo "Please set PHOENIX_API_URL or use --api-url"
    exit 1
fi

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
    echo -e "${RED}Please run this script without sudo${NC}"
    echo "The script will use sudo when needed"
    exit 1
fi

# Test connectivity to Phoenix API
echo -e "${YELLOW}Testing connectivity to Phoenix API...${NC}"
if curl -s -f "$PHOENIX_API_URL/health" > /dev/null; then
    echo -e "${GREEN}✓ Phoenix API is reachable${NC}"
else
    echo -e "${RED}✗ Cannot reach Phoenix API at $PHOENIX_API_URL${NC}"
    exit 1
fi

# Install Phoenix Agent binary
echo -e "\n${YELLOW}Installing Phoenix Agent...${NC}"

# Check if binary exists
if [ ! -f "$PHOENIX_BIN_DIR/phoenix-agent" ]; then
    echo -e "${RED}Phoenix Agent binary not found at $PHOENIX_BIN_DIR/phoenix-agent${NC}"
    echo "Please copy the binary first:"
    echo "  sudo cp /path/to/phoenix-agent $PHOENIX_BIN_DIR/"
    echo "  sudo chmod +x $PHOENIX_BIN_DIR/phoenix-agent"
    echo "  sudo chown phoenix:phoenix $PHOENIX_BIN_DIR/phoenix-agent"
    exit 1
fi

# Configure agent
echo -e "\n${YELLOW}Configuring Phoenix Agent...${NC}"

# Update environment file
sudo tee "$PHOENIX_CONFIG_DIR/phoenix-agent.env" > /dev/null << EOF
# Phoenix Agent Configuration
PHOENIX_API_URL=$PHOENIX_API_URL
PHOENIX_AGENT_HOST_ID=$PHOENIX_AGENT_HOST_ID
PHOENIX_AGENT_POLL_INTERVAL=15s
PHOENIX_AGENT_VERSION=1.0.0
PHOENIX_AGENT_LABELS=region=$(curl -s http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || echo "unknown")
EOF

# Install OpenTelemetry Collector (optional)
if [ "$INSTALL_OTEL" = true ]; then
    echo -e "\n${YELLOW}Installing OpenTelemetry Collector...${NC}"
    
    # Download otel collector
    OTEL_VERSION="0.88.0"
    OTEL_URL="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${OTEL_VERSION}/otelcol_${OTEL_VERSION}_linux_amd64.tar.gz"
    
    if [ ! -f "$PHOENIX_BIN_DIR/otelcol" ]; then
        echo "Downloading OpenTelemetry Collector..."
        wget -q -O /tmp/otelcol.tar.gz "$OTEL_URL"
        sudo tar -xzf /tmp/otelcol.tar.gz -C "$PHOENIX_BIN_DIR" otelcol
        sudo chown phoenix:phoenix "$PHOENIX_BIN_DIR/otelcol"
        rm /tmp/otelcol.tar.gz
    fi
    
    # Create otel config directory
    sudo mkdir -p "$PHOENIX_CONFIG_DIR/otel"
    
    # Create systemd service for otel collector
    sudo tee /etc/systemd/system/otel-collector.service > /dev/null << EOF
[Unit]
Description=OpenTelemetry Collector
After=network.target

[Service]
Type=simple
User=phoenix
Group=phoenix
ExecStart=$PHOENIX_BIN_DIR/otelcol --config=$PHOENIX_CONFIG_DIR/otel/config.yaml
Restart=always
RestartSec=5
StandardOutput=append:/var/log/phoenix/otel-collector.log
StandardError=append:/var/log/phoenix/otel-collector.log

[Install]
WantedBy=multi-user.target
EOF
fi

# Enable and start services
echo -e "\n${YELLOW}Starting Phoenix Agent...${NC}"
sudo systemctl daemon-reload
sudo systemctl enable phoenix-agent
sudo systemctl restart phoenix-agent

if [ "$INSTALL_OTEL" = true ] && [ -f "$PHOENIX_CONFIG_DIR/otel/config.yaml" ]; then
    sudo systemctl enable otel-collector
    sudo systemctl restart otel-collector
fi

# Wait for agent to register
echo -n "Waiting for agent to register..."
for i in {1..30}; do
    if sudo journalctl -u phoenix-agent -n 10 | grep -q "Successfully registered"; then
        echo -e " ${GREEN}Registered!${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

# Check service status
echo -e "\n${YELLOW}Service Status:${NC}"
sudo systemctl status phoenix-agent --no-pager || true

# Show agent info
echo -e "\n${GREEN}✅ Agent deployment complete!${NC}"
echo ""
echo "Agent Information:"
echo "  - Host ID: $PHOENIX_AGENT_HOST_ID"
echo "  - API URL: $PHOENIX_API_URL"
echo "  - Status: $(sudo systemctl is-active phoenix-agent)"
echo ""
echo "Logs:"
echo "  - Phoenix Agent: /var/log/phoenix/phoenix-agent.log"
echo "  - System logs: journalctl -u phoenix-agent -f"

if [ "$INSTALL_OTEL" = true ]; then
    echo "  - OTel Collector: /var/log/phoenix/otel-collector.log"
fi

echo ""
echo "To verify agent registration:"
echo "  curl $PHOENIX_API_URL/api/v1/agents | jq ."