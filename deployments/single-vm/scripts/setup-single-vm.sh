#!/usr/bin/env bash
#
# Phoenix Single-VM Setup Script
# This script sets up a complete Phoenix control plane on a single VM
#
# Usage: ./setup-single-vm.sh
#

set -euo pipefail

# Configuration
PHOENIX_DIR="/opt/phoenix"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    local missing=()
    
    # Check for required commands
    for cmd in docker docker-compose openssl curl; do
        if ! command -v $cmd >/dev/null 2>&1; then
            missing+=($cmd)
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required commands: ${missing[*]}"
        error "Please install them first"
        exit 1
    fi
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    # Check if running as root or with sudo
    if [[ $EUID -ne 0 ]] && ! sudo -n true 2>/dev/null; then
        warning "This script requires sudo access for some operations"
        sudo -v
    fi
    
    log "All prerequisites satisfied"
}

# Create directory structure
create_directories() {
    log "Creating directory structure..."
    
    # Create main directories
    sudo mkdir -p "$PHOENIX_DIR"/{data,config,tls,scripts,backups}
    sudo mkdir -p "$PHOENIX_DIR"/data/{postgres,prometheus,grafana,uploads,nginx/logs}
    sudo mkdir -p "$PHOENIX_DIR"/config/{prometheus,grafana/provisioning,nginx}
    
    # Copy files from repository
    sudo cp -r "$BASE_DIR"/* "$PHOENIX_DIR/"
    
    # Set permissions
    sudo chmod -R 755 "$PHOENIX_DIR"
    
    log "Directory structure created"
}

# Generate secrets and configuration
generate_config() {
    log "Generating configuration..."
    
    # Check if .env already exists
    if [[ -f "$PHOENIX_DIR/.env" ]]; then
        warning ".env file already exists. Backing up to .env.backup"
        sudo cp "$PHOENIX_DIR/.env" "$PHOENIX_DIR/.env.backup.$(date +%s)"
    fi
    
    # Generate secrets
    local jwt_secret=$(openssl rand -base64 32)
    local postgres_password=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-16)
    local grafana_password=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-12)
    local agent_token=$(openssl rand -hex 16)
    
    # Get public URL
    echo
    read -p "Enter the public URL for Phoenix (e.g., https://phoenix.company.com): " public_url
    
    # Get pricing information
    echo
    info "Default pricing is ₹0.00011 per metric series per day (New Relic pricing)"
    read -p "Enter custom price per series (or press Enter for default): " price_per_series
    price_per_series=${price_per_series:-0.00011}
    
    # Create .env file
    cat > "$PHOENIX_DIR/.env" << EOF
# Phoenix Single-VM Configuration
# Generated on $(date)

# Public URL
PHX_PUBLIC_URL=${public_url}

# Security
PHX_JWT_SECRET=${jwt_secret}
POSTGRES_PASSWORD=${postgres_password}
GRAFANA_PASSWORD=${grafana_password}
AGENT_BOOTSTRAP_TOKEN=${agent_token}

# Pricing
PHX_PRICE_PER_SERIES=${price_per_series}

# Features
ENABLE_TLS=true

# Resource Limits
API_MEMORY_LIMIT=2g
DB_MEMORY_LIMIT=1g
PROMETHEUS_MEMORY_LIMIT=2g

# Data Retention
PROMETHEUS_RETENTION_DAYS=30
LOG_RETENTION_DAYS=7
EOF
    
    sudo chmod 600 "$PHOENIX_DIR/.env"
    
    log "Configuration generated successfully"
    info "Agent bootstrap token: ${agent_token}"
}

# Setup TLS certificates
setup_tls() {
    log "Setting up TLS certificates..."
    
    local tls_dir="$PHOENIX_DIR/tls"
    
    # Check if certificates already exist
    if [[ -f "$tls_dir/fullchain.pem" ]] && [[ -f "$tls_dir/privkey.pem" ]]; then
        info "TLS certificates already exist"
        return
    fi
    
    echo
    echo "TLS Certificate Options:"
    echo "1) Use Let's Encrypt (requires domain pointing to this server)"
    echo "2) Use existing certificates"
    echo "3) Generate self-signed certificates (for testing only)"
    echo "4) Skip TLS setup (HTTP only - not recommended)"
    echo
    read -p "Select option [1-4]: " tls_option
    
    case $tls_option in
        1)
            setup_letsencrypt
            ;;
        2)
            setup_existing_certs
            ;;
        3)
            generate_self_signed
            ;;
        4)
            warning "Skipping TLS setup. Phoenix will run in HTTP-only mode."
            warning "This is NOT recommended for production!"
            sudo sed -i 's/ENABLE_TLS=true/ENABLE_TLS=false/' "$PHOENIX_DIR/.env"
            ;;
        *)
            error "Invalid option"
            exit 1
            ;;
    esac
}

# Setup Let's Encrypt
setup_letsencrypt() {
    log "Setting up Let's Encrypt certificates..."
    
    # Install certbot if not present
    if ! command -v certbot >/dev/null 2>&1; then
        info "Installing certbot..."
        if command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update && sudo apt-get install -y certbot
        elif command -v yum >/dev/null 2>&1; then
            sudo yum install -y certbot
        else
            error "Please install certbot manually"
            exit 1
        fi
    fi
    
    # Get domain from public URL
    local domain=$(echo "$public_url" | sed -E 's|https?://||' | sed 's|/.*||')
    
    # Obtain certificate
    sudo certbot certonly --standalone \
        -d "$domain" \
        --non-interactive \
        --agree-tos \
        --email "admin@${domain}" \
        --keep-until-expiring
    
    # Link certificates
    sudo ln -sf "/etc/letsencrypt/live/$domain/fullchain.pem" "$PHOENIX_DIR/tls/fullchain.pem"
    sudo ln -sf "/etc/letsencrypt/live/$domain/privkey.pem" "$PHOENIX_DIR/tls/privkey.pem"
    
    # Setup auto-renewal
    echo "0 2 * * * root certbot renew --quiet && docker-compose -f $PHOENIX_DIR/docker-compose.yml restart api" | \
        sudo tee /etc/cron.d/phoenix-cert-renewal
    
    log "Let's Encrypt setup complete"
}

# Setup existing certificates
setup_existing_certs() {
    echo
    read -p "Enter path to certificate file (fullchain.pem): " cert_path
    read -p "Enter path to private key file (privkey.pem): " key_path
    
    if [[ ! -f "$cert_path" ]] || [[ ! -f "$key_path" ]]; then
        error "Certificate files not found"
        exit 1
    fi
    
    sudo cp "$cert_path" "$PHOENIX_DIR/tls/fullchain.pem"
    sudo cp "$key_path" "$PHOENIX_DIR/tls/privkey.pem"
    sudo chmod 600 "$PHOENIX_DIR/tls/"*.pem
    
    log "Certificates copied successfully"
}

# Generate self-signed certificates
generate_self_signed() {
    log "Generating self-signed certificates..."
    
    local domain=$(echo "$public_url" | sed -E 's|https?://||' | sed 's|/.*||')
    
    sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$PHOENIX_DIR/tls/privkey.pem" \
        -out "$PHOENIX_DIR/tls/fullchain.pem" \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=$domain"
    
    sudo chmod 600 "$PHOENIX_DIR/tls/"*.pem
    
    warning "Self-signed certificate generated. Browsers will show security warnings."
    log "Certificate generation complete"
}

# Initialize database
init_database() {
    log "Initializing database..."
    
    # Start only database service
    cd "$PHOENIX_DIR"
    sudo docker-compose up -d db
    
    # Wait for database to be ready
    info "Waiting for database to be ready..."
    local retries=30
    while ! sudo docker-compose exec -T db pg_isready -U phoenix >/dev/null 2>&1; do
        retries=$((retries - 1))
        if [[ $retries -le 0 ]]; then
            error "Database failed to start"
            exit 1
        fi
        sleep 2
    done
    
    log "Database initialized successfully"
}

# Start Phoenix services
start_services() {
    log "Starting Phoenix services..."
    
    cd "$PHOENIX_DIR"
    
    # Pull latest images
    sudo docker-compose pull
    
    # Start all services
    sudo docker-compose up -d
    
    # Wait for services to be healthy
    info "Waiting for services to be ready..."
    sleep 10
    
    # Check service health
    local services=("api" "prometheus" "pushgateway")
    for service in "${services[@]}"; do
        if sudo docker-compose ps | grep -q "${service}.*Up"; then
            log "$service is running"
        else
            error "$service failed to start"
            sudo docker-compose logs "$service" | tail -20
        fi
    done
    
    log "All services started successfully"
}

# Create agent installer endpoint
create_agent_installer() {
    log "Creating agent installer endpoint..."
    
    # Update install script with actual values
    local install_script="$PHOENIX_DIR/scripts/install-agent.sh"
    
    # Read values from .env
    source "$PHOENIX_DIR/.env"
    
    # Update script
    sudo sed -i "s|PHOENIX_API_URL:-.*}|PHOENIX_API_URL:-${PHX_PUBLIC_URL}}|" "$install_script"
    sudo sed -i "s|PHOENIX_TOKEN:-.*}|PHOENIX_TOKEN:-${AGENT_BOOTSTRAP_TOKEN}}|" "$install_script"
    
    # Make it accessible via API
    sudo mkdir -p "$PHOENIX_DIR/data/uploads/public"
    sudo cp "$install_script" "$PHOENIX_DIR/data/uploads/public/install-agent.sh"
    sudo chmod 644 "$PHOENIX_DIR/data/uploads/public/install-agent.sh"
    
    log "Agent installer created"
}

# Setup backups
setup_backups() {
    log "Setting up automated backups..."
    
    # Create backup script
    cat > "$PHOENIX_DIR/scripts/backup.sh" << 'EOF'
#!/bin/bash
# Phoenix Backup Script

BACKUP_DIR="/opt/phoenix/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup database
docker-compose -f /opt/phoenix/docker-compose.yml exec -T db \
    pg_dump -U phoenix -Fc phoenix > "$BACKUP_DIR/phoenix_db_${TIMESTAMP}.dump"

# Backup Prometheus data
tar -czf "$BACKUP_DIR/prometheus_data_${TIMESTAMP}.tar.gz" \
    -C /opt/phoenix/data prometheus/

# Backup configuration
tar -czf "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" \
    -C /opt/phoenix config/ .env

# Clean old backups
find "$BACKUP_DIR" -name "*.dump" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $TIMESTAMP"
EOF
    
    sudo chmod +x "$PHOENIX_DIR/scripts/backup.sh"
    
    # Setup cron jobs - hourly backups
    cat << EOF | sudo tee /etc/cron.d/phoenix-backup
# Phoenix automated backups
# Hourly incremental backups
0 * * * * root $PHOENIX_DIR/scripts/backup.sh --incremental >> /var/log/phoenix-backup.log 2>&1
# Daily full backup at 2 AM
0 2 * * * root $PHOENIX_DIR/scripts/backup.sh --full >> /var/log/phoenix-backup.log 2>&1
# Weekly backup rotation on Sunday
0 3 * * 0 root $PHOENIX_DIR/scripts/backup.sh --rotate-weekly >> /var/log/phoenix-backup.log 2>&1
EOF
    
    log "Backup automation configured"
}

# Setup monitoring alerts
setup_monitoring() {
    log "Setting up monitoring alerts..."
    
    # Copy scaling rules
    sudo cp "$BASE_DIR/config/scaling-rules.yml" "$PHOENIX_DIR/config/"
    
    # Setup auto-scale monitor service
    sudo cp "$BASE_DIR/scripts/auto-scale-monitor.sh" "$PHOENIX_DIR/scripts/"
    sudo chmod +x "$PHOENIX_DIR/scripts/auto-scale-monitor.sh"
    
    # Create systemd service for auto-scale monitor
    cat > /etc/systemd/system/phoenix-autoscale.service << 'EOF'
[Unit]
Description=Phoenix Auto-Scale Monitor
After=docker.service phoenix-api.service
Requires=docker.service

[Service]
Type=simple
ExecStart=/opt/phoenix/scripts/auto-scale-monitor.sh
Restart=always
RestartSec=30
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl enable phoenix-autoscale.service
    sudo systemctl start phoenix-autoscale.service
    
    # Create Prometheus rules
    cat > "$PHOENIX_DIR/config/prometheus/rules.yml" << 'EOF'
groups:
  - name: phoenix_alerts
    interval: 30s
    rules:
      - alert: PhoenixAPIDown
        expr: up{job="phoenix-api"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Phoenix API is down"
          description: "Phoenix API has been down for more than 2 minutes"
      
      - alert: AgentOffline
        expr: time() - phoenix_agent_last_heartbeat > 300
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Phoenix agent offline"
          description: "Agent {{ $labels.agent_id }} has been offline for more than 5 minutes"
      
      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes{name="phoenix-api"} / container_spec_memory_limit_bytes{name="phoenix-api"} > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Phoenix API memory usage is above 90%"
EOF
    
    log "Monitoring alerts configured"
}

# Setup auto-scale monitor
setup_autoscale_monitor() {
    log "Setting up auto-scale monitor..."
    
    # Copy auto-scale monitor script
    sudo cp "$BASE_DIR/scripts/auto-scale-monitor.sh" "$PHOENIX_DIR/scripts/"
    sudo chmod +x "$PHOENIX_DIR/scripts/auto-scale-monitor.sh"
    
    # Copy scaling rules
    sudo cp "$BASE_DIR/config/scaling-rules.yml" "$PHOENIX_DIR/config/"
    
    # Install as systemd service
    "$PHOENIX_DIR/scripts/auto-scale-monitor.sh" --install
    
    log "Auto-scale monitor installed and started"
}

# Print success message and next steps
print_success() {
    echo
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}     Phoenix Single-VM Installation Completed Successfully!     ${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo
    echo "Access Phoenix:"
    echo "  UI:       ${PHX_PUBLIC_URL}"
    echo "  Grafana:  ${PHX_PUBLIC_URL}:3000 (admin/${GRAFANA_PASSWORD})"
    echo
    echo "Install agents on edge nodes:"
    echo "  curl -fsSL ${PHX_PUBLIC_URL}/install-agent.sh | sudo bash"
    echo
    echo "Useful commands:"
    echo "  View logs:    cd $PHOENIX_DIR && sudo docker-compose logs -f"
    echo "  Stop:         cd $PHOENIX_DIR && sudo docker-compose down"
    echo "  Backup:       sudo $PHOENIX_DIR/scripts/backup.sh"
    echo "  Update:       cd $PHOENIX_DIR && sudo docker-compose pull && sudo docker-compose up -d"
    echo
    echo "Configuration:"
    echo "  Main config:  $PHOENIX_DIR/.env"
    echo "  Prometheus:   $PHOENIX_DIR/config/prometheus.yml"
    echo
    echo "Next steps:"
    echo "  1. Install agents on your edge nodes"
    echo "  2. Create your first experiment in the UI"
    echo "  3. Monitor cost savings in real-time"
    echo
    echo -e "${YELLOW}IMPORTANT: Save the agent bootstrap token:${NC}"
    echo "  ${AGENT_BOOTSTRAP_TOKEN}"
    echo
}

# Main installation flow
main() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}              Phoenix Single-VM Setup Script                    ${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo
    
    check_prerequisites
    create_directories
    generate_config
    setup_tls
    init_database
    start_services
    create_agent_installer
    setup_backups
    setup_monitoring
    setup_autoscale_monitor
    
    # Source .env for final message
    source "$PHOENIX_DIR/.env"
    
    print_success
}

# Run main function
main "$@"