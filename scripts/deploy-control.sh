#!/bin/bash
# Deploy Phoenix control plane components

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🎛️  Phoenix Control Plane Deployment"
echo "===================================="
echo ""

# Check if running as root or with sudo
if [ "$EUID" -eq 0 ]; then 
    echo -e "${RED}Please run this script without sudo${NC}"
    echo "The script will use sudo when needed"
    exit 1
fi

# Configuration
PHOENIX_VERSION="${PHOENIX_VERSION:-latest}"
PHOENIX_DATA_DIR="${PHOENIX_DATA_DIR:-/var/lib/phoenix}"
PHOENIX_CONFIG_DIR="${PHOENIX_CONFIG_DIR:-/etc/phoenix}"
PHOENIX_BIN_DIR="${PHOENIX_BIN_DIR:-/usr/local/bin}"

# Parse arguments
DRY_RUN=false
SKIP_INFRA=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-infra)
            SKIP_INFRA=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--dry-run] [--skip-infra]"
            exit 1
            ;;
    esac
done

# Function to run command (respects dry-run)
run_cmd() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY-RUN] $@"
    else
        "$@"
    fi
}

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        return 1
    else
        echo -e "${GREEN}✓ $1 is installed${NC}"
        return 0
    fi
}

PREREQ_OK=true
check_command docker || PREREQ_OK=false
check_command docker-compose || PREREQ_OK=false
check_command systemctl || PREREQ_OK=false

if [ "$PREREQ_OK" = false ]; then
    echo -e "${RED}Please install missing prerequisites${NC}"
    exit 1
fi

# Deploy infrastructure services
if [ "$SKIP_INFRA" = false ]; then
    echo -e "\n${YELLOW}Deploying infrastructure services...${NC}"
    
    # Create docker-compose for infrastructure
    sudo mkdir -p "$PHOENIX_CONFIG_DIR/docker"
    sudo tee "$PHOENIX_CONFIG_DIR/docker/docker-compose.yml" > /dev/null << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: phoenix-postgres
    environment:
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-phoenix-prod}
      POSTGRES_DB: phoenix
    ports:
      - "127.0.0.1:5432:5432"
    volumes:
      - phoenix-postgres-data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U phoenix"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: phoenix-redis
    ports:
      - "127.0.0.1:6379:6379"
    volumes:
      - phoenix-redis-data:/data
    restart: unless-stopped
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  prometheus:
    image: prom/prometheus:latest
    container_name: phoenix-prometheus
    ports:
      - "127.0.0.1:9090:9090"
    volumes:
      - /etc/prometheus:/etc/prometheus
      - phoenix-prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    restart: unless-stopped
    user: "65534:65534"

volumes:
  phoenix-postgres-data:
  phoenix-redis-data:
  phoenix-prometheus-data:
EOF

    # Create .env file for docker-compose
    if [ ! -f "$PHOENIX_CONFIG_DIR/docker/.env" ]; then
        echo "POSTGRES_PASSWORD=$(openssl rand -base64 32)" | sudo tee "$PHOENIX_CONFIG_DIR/docker/.env" > /dev/null
        sudo chmod 600 "$PHOENIX_CONFIG_DIR/docker/.env"
    fi
    
    # Start infrastructure
    echo "Starting infrastructure services..."
    cd "$PHOENIX_CONFIG_DIR/docker"
    run_cmd sudo docker-compose up -d
    
    # Wait for services
    echo "Waiting for services to be healthy..."
    sleep 10
    
    # Check service health
    run_cmd sudo docker-compose ps
fi

# Install Phoenix API binary
echo -e "\n${YELLOW}Installing Phoenix API...${NC}"

# Check if binary exists
if [ ! -f "$PHOENIX_BIN_DIR/phoenix-api" ]; then
    echo -e "${RED}Phoenix API binary not found at $PHOENIX_BIN_DIR/phoenix-api${NC}"
    echo "Please copy the binary first:"
    echo "  sudo cp /path/to/phoenix-api $PHOENIX_BIN_DIR/"
    echo "  sudo chmod +x $PHOENIX_BIN_DIR/phoenix-api"
    echo "  sudo chown phoenix:phoenix $PHOENIX_BIN_DIR/phoenix-api"
    exit 1
fi

# Initialize database
echo -e "\n${YELLOW}Initializing database...${NC}"

# Get database password
if [ -f "$PHOENIX_CONFIG_DIR/docker/.env" ]; then
    source "$PHOENIX_CONFIG_DIR/docker/.env"
    DB_PASSWORD="$POSTGRES_PASSWORD"
else
    DB_PASSWORD="phoenix-prod"
fi

# Update phoenix-api.env with correct database password
sudo sed -i "s|postgresql://phoenix:CHANGE_ME@|postgresql://phoenix:$DB_PASSWORD@|" "$PHOENIX_CONFIG_DIR/phoenix-api.env"

# Run migrations
echo "Running database migrations..."
if [ "$DRY_RUN" = false ]; then
    sudo -u phoenix DATABASE_URL="postgresql://phoenix:$DB_PASSWORD@localhost:5432/phoenix?sslmode=disable" \
        "$PHOENIX_BIN_DIR/phoenix-api" migrate up
fi

# Configure Prometheus
echo -e "\n${YELLOW}Configuring Prometheus...${NC}"
sudo tee /etc/prometheus/prometheus.yml > /dev/null << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets: []

rule_files:
  - "phoenix_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

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

# Create TLS certificates (if needed)
if [ ! -f "$PHOENIX_CONFIG_DIR/certs/server.crt" ]; then
    echo -e "\n${YELLOW}Generating TLS certificates...${NC}"
    sudo mkdir -p "$PHOENIX_CONFIG_DIR/certs"
    
    # Generate self-signed cert for development
    run_cmd sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$PHOENIX_CONFIG_DIR/certs/server.key" \
        -out "$PHOENIX_CONFIG_DIR/certs/server.crt" \
        -subj "/C=US/ST=State/L=City/O=Phoenix/OU=IT/CN=phoenix.local"
    
    run_cmd sudo chown -R phoenix:phoenix "$PHOENIX_CONFIG_DIR/certs"
    run_cmd sudo chmod 600 "$PHOENIX_CONFIG_DIR/certs/server.key"
fi

# Enable and start services
echo -e "\n${YELLOW}Starting Phoenix services...${NC}"
run_cmd sudo systemctl daemon-reload
run_cmd sudo systemctl enable phoenix-api
run_cmd sudo systemctl restart phoenix-api

# Wait for API to be ready
echo -n "Waiting for Phoenix API to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e " ${GREEN}Ready!${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

# Check service status
echo -e "\n${YELLOW}Service Status:${NC}"
run_cmd sudo systemctl status phoenix-api --no-pager || true

# Show connection info
echo -e "\n${GREEN}✅ Control plane deployment complete!${NC}"
echo ""
echo "Phoenix API endpoints:"
echo "  - API: http://$(hostname -I | awk '{print $1}'):8080"
echo "  - Health: http://$(hostname -I | awk '{print $1}'):8080/health"
echo "  - Metrics: http://$(hostname -I | awk '{print $1}'):8080/metrics"
echo ""
echo "Infrastructure services:"
echo "  - PostgreSQL: localhost:5432 (user: phoenix)"
echo "  - Redis: localhost:6379"
echo "  - Prometheus: http://localhost:9090"
echo ""
echo "Logs:"
echo "  - Phoenix API: /var/log/phoenix/phoenix-api.log"
echo "  - System logs: journalctl -u phoenix-api -f"
echo ""
echo "To add agents, run on agent nodes:"
echo "  export PHOENIX_API_URL=http://$(hostname -I | awk '{print $1}'):8080"
echo "  ./deploy-agent.sh"