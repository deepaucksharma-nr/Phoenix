#!/bin/bash
# Setup Phoenix for multi-VM deployment

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🚀 Phoenix Multi-VM Deployment Setup"
echo "===================================="
echo ""

# Parse arguments
ROLE=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --control-plane)
            ROLE="control"
            shift
            ;;
        --agent)
            ROLE="agent"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--control-plane|--agent]"
            exit 1
            ;;
    esac
done

if [ -z "$ROLE" ]; then
    echo "Please specify role: --control-plane or --agent"
    exit 1
fi

# Common setup
echo -e "${YELLOW}Performing common setup...${NC}"

# Create Phoenix user
if ! id "phoenix" &>/dev/null; then
    echo "Creating phoenix user..."
    sudo useradd -r -s /bin/bash -m -d /var/lib/phoenix phoenix
fi

# Create directories
echo "Creating directories..."
sudo mkdir -p /etc/phoenix/{certs,config}
sudo mkdir -p /var/lib/phoenix/{data,logs}
sudo mkdir -p /var/log/phoenix
sudo chown -R phoenix:phoenix /var/lib/phoenix /var/log/phoenix

# Install dependencies
echo -e "\n${YELLOW}Installing dependencies...${NC}"

# Check OS
if [ -f /etc/debian_version ]; then
    # Debian/Ubuntu
    sudo apt-get update
    sudo apt-get install -y curl wget jq netcat-openbsd
elif [ -f /etc/redhat-release ]; then
    # RHEL/CentOS
    sudo yum install -y curl wget jq nc
else
    echo -e "${RED}Unsupported OS${NC}"
    exit 1
fi

# Setup firewall rules
echo -e "\n${YELLOW}Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    # UFW (Ubuntu)
    if [ "$ROLE" = "control" ]; then
        sudo ufw allow 8080/tcp comment "Phoenix API"
        sudo ufw allow 8081/tcp comment "Phoenix WebSocket"
        sudo ufw allow 9090/tcp comment "Prometheus"
        sudo ufw allow 5432/tcp comment "PostgreSQL"
        sudo ufw allow 6379/tcp comment "Redis"
    else
        # Agent only needs outbound
        echo "Agent node - no inbound ports required"
    fi
elif command -v firewall-cmd &> /dev/null; then
    # firewalld (RHEL/CentOS)
    if [ "$ROLE" = "control" ]; then
        sudo firewall-cmd --permanent --add-port=8080/tcp
        sudo firewall-cmd --permanent --add-port=8081/tcp
        sudo firewall-cmd --permanent --add-port=9090/tcp
        sudo firewall-cmd --permanent --add-port=5432/tcp
        sudo firewall-cmd --permanent --add-port=6379/tcp
        sudo firewall-cmd --reload
    fi
fi

# Setup systemd service files
echo -e "\n${YELLOW}Creating systemd service files...${NC}"

if [ "$ROLE" = "control" ]; then
    # Phoenix API service
    sudo tee /etc/systemd/system/phoenix-api.service > /dev/null << 'EOF'
[Unit]
Description=Phoenix API Server
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=phoenix
Group=phoenix
WorkingDirectory=/var/lib/phoenix
ExecStart=/usr/local/bin/phoenix-api
Restart=always
RestartSec=10
StandardOutput=append:/var/log/phoenix/phoenix-api.log
StandardError=append:/var/log/phoenix/phoenix-api.log
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
EnvironmentFile=-/etc/phoenix/phoenix-api.env

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/phoenix /var/log/phoenix

[Install]
WantedBy=multi-user.target
EOF

    # Create environment file template
    sudo tee /etc/phoenix/phoenix-api.env > /dev/null << EOF
# Phoenix API Configuration
PORT=8080
DATABASE_URL=postgresql://phoenix:CHANGE_ME@localhost:5432/phoenix?sslmode=disable
REDIS_URL=redis://localhost:6379
PROMETHEUS_URL=http://localhost:9090
JWT_SECRET=CHANGE_ME_$(openssl rand -hex 32)
ENVIRONMENT=production
LOG_LEVEL=info
SKIP_MIGRATIONS=false
EOF

    echo -e "${GREEN}✓ Control plane systemd services created${NC}"
    
else
    # Phoenix Agent service
    sudo tee /etc/systemd/system/phoenix-agent.service > /dev/null << 'EOF'
[Unit]
Description=Phoenix Agent
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=phoenix
Group=phoenix
WorkingDirectory=/var/lib/phoenix
ExecStart=/usr/local/bin/phoenix-agent
Restart=always
RestartSec=30
StandardOutput=append:/var/log/phoenix/phoenix-agent.log
StandardError=append:/var/log/phoenix/phoenix-agent.log
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
EnvironmentFile=-/etc/phoenix/phoenix-agent.env

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/phoenix /var/log/phoenix

[Install]
WantedBy=multi-user.target
EOF

    # Create environment file template
    sudo tee /etc/phoenix/phoenix-agent.env > /dev/null << EOF
# Phoenix Agent Configuration
PHOENIX_API_URL=http://CONTROL_PLANE_IP:8080
PHOENIX_AGENT_HOST_ID=$(hostname)
PHOENIX_AGENT_POLL_INTERVAL=15s
PHOENIX_AGENT_VERSION=1.0.0
EOF

    echo -e "${GREEN}✓ Agent systemd service created${NC}"
fi

# Setup log rotation
echo -e "\n${YELLOW}Configuring log rotation...${NC}"
sudo tee /etc/logrotate.d/phoenix > /dev/null << 'EOF'
/var/log/phoenix/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
    create 0644 phoenix phoenix
    sharedscripts
    postrotate
        systemctl reload phoenix-api 2>/dev/null || true
        systemctl reload phoenix-agent 2>/dev/null || true
    endscript
}
EOF

# Setup monitoring
if [ "$ROLE" = "control" ]; then
    echo -e "\n${YELLOW}Setting up monitoring...${NC}"
    
    # Create Prometheus scrape config
    sudo mkdir -p /etc/prometheus
    sudo tee /etc/prometheus/phoenix.yml > /dev/null << 'EOF'
# Phoenix monitoring targets
- job_name: 'phoenix-api'
  static_configs:
    - targets: ['localhost:8080']
  metrics_path: '/metrics'

- job_name: 'phoenix-agents'
  file_sd_configs:
    - files:
      - '/etc/prometheus/phoenix-agents.json'
      refresh_interval: 30s
EOF
    
    # Create empty agents file
    echo '[]' | sudo tee /etc/prometheus/phoenix-agents.json > /dev/null
fi

# Download binaries placeholder
echo -e "\n${YELLOW}Binary installation:${NC}"
echo "Please copy the Phoenix binaries to /usr/local/bin/"
if [ "$ROLE" = "control" ]; then
    echo "  - phoenix-api"
    echo "  - phoenix-cli (optional)"
else
    echo "  - phoenix-agent"
fi

# Summary
echo -e "\n${GREEN}✅ Setup complete for $ROLE node!${NC}"
echo ""
if [ "$ROLE" = "control" ]; then
    echo "Next steps for control plane:"
    echo "1. Install PostgreSQL and Redis"
    echo "2. Copy phoenix-api binary to /usr/local/bin/"
    echo "3. Edit /etc/phoenix/phoenix-api.env"
    echo "4. Initialize database: phoenix-api migrate"
    echo "5. Start service: sudo systemctl start phoenix-api"
    echo "6. Enable on boot: sudo systemctl enable phoenix-api"
else
    echo "Next steps for agent:"
    echo "1. Copy phoenix-agent binary to /usr/local/bin/"
    echo "2. Edit /etc/phoenix/phoenix-agent.env (set CONTROL_PLANE_IP)"
    echo "3. Start service: sudo systemctl start phoenix-agent"
    echo "4. Enable on boot: sudo systemctl enable phoenix-agent"
fi