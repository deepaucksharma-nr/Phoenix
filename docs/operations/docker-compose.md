# Docker Compose Deployment Guide

This guide covers deploying Phoenix Platform using Docker Compose for production environments.

## Overview

Phoenix Platform uses Docker Compose for container orchestration, providing a simple yet powerful deployment model suitable for most production workloads.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Docker Host                        │
│                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │  Phoenix    │  │  Phoenix    │  │  Dashboard  │ │
│  │    API      │  │  PostgreSQL │  │   (React)   │ │
│  │  Port 8080  │  │  Port 5432  │  │  Port 3000  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
│                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │ Prometheus  │  │   Grafana   │  │ Pushgateway │ │
│  │  Port 9090  │  │  Port 3001  │  │  Port 9091  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────┘
                           │
                    Agent Connections
                           │
        ┌──────────────────┴──────────────────┐
        │                                      │
   ┌────▼────┐                           ┌────▼────┐
   │  Agent  │                           │  Agent  │
   │ (Host 1)│                           │ (Host N)│
   └─────────┘                           └─────────┘
```

## Prerequisites

- Docker Engine 20.10+
- Docker Compose v2.0+
- 4 vCPU, 16GB RAM minimum
- 100GB disk space
- Ubuntu 20.04+ or RHEL 8+

## Installation

### 1. System Preparation

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. Clone Repository

```bash
git clone https://github.com/phoenix/platform.git
cd platform/deployments/single-vm
```

### 3. Configure Environment

```bash
# Copy example configuration
cp .env.example .env

# Edit configuration
vim .env
```

Required configuration:
```bash
# Database
POSTGRES_USER=phoenix
POSTGRES_PASSWORD=<strong-password>
POSTGRES_DB=phoenix

# API
JWT_SECRET=<random-secret>
API_PORT=8080

# Collector Configuration (choose one)
# Option 1: OpenTelemetry Collector
COLLECTOR_TYPE=otel
OTEL_COLLECTOR_ENDPOINT=http://otel-collector:4317

# Option 2: New Relic NRDOT
# COLLECTOR_TYPE=nrdot
# NRDOT_OTLP_ENDPOINT=https://otlp.nr-data.net:4317
# NEW_RELIC_LICENSE_KEY=<your-license-key>

# Monitoring
PROMETHEUS_RETENTION=30d
GRAFANA_ADMIN_PASSWORD=<admin-password>

# TLS (optional)
TLS_ENABLED=true
TLS_CERT_PATH=/etc/phoenix/certs/cert.pem
TLS_KEY_PATH=/etc/phoenix/certs/key.pem
```

### 4. Deploy Services

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

## Service Configuration

### Phoenix API

The API service handles all control plane operations:

```yaml
phoenix-api:
  image: phoenix/api:latest
  ports:
    - "8080:8080"  # REST API + WebSocket
  environment:
    - DATABASE_URL=postgresql://phoenix:${POSTGRES_PASSWORD}@postgres:5432/phoenix
    - JWT_SECRET=${JWT_SECRET}
    - LOG_LEVEL=info
  depends_on:
    - postgres
  restart: unless-stopped
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval: 30s
    timeout: 10s
    retries: 3
```

### PostgreSQL

Primary datastore configuration:

```yaml
postgres:
  image: postgres:15-alpine
  environment:
    - POSTGRES_USER=${POSTGRES_USER}
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    - POSTGRES_DB=${POSTGRES_DB}
  volumes:
    - postgres_data:/var/lib/postgresql/data
  ports:
    - "5432:5432"
  restart: unless-stopped
```

### Monitoring Stack

Prometheus and Grafana for observability:

```yaml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_data:/prometheus
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.retention.time=${PROMETHEUS_RETENTION}'
  ports:
    - "9090:9090"
  restart: unless-stopped

grafana:
  image: grafana/grafana:latest
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
  volumes:
    - grafana_data:/var/lib/grafana
    - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
  ports:
    - "3001:3000"
  restart: unless-stopped
```

## Agent Deployment

Agents run as systemd services on target hosts:

### Install Agent

```bash
# On each target host
curl -sSL http://phoenix-api:8080/install-agent.sh | sudo bash
```

### Manual Installation

```bash
# Download agent binary
wget https://github.com/phoenix/platform/releases/latest/download/phoenix-agent-linux-amd64
chmod +x phoenix-agent-linux-amd64
sudo mv phoenix-agent-linux-amd64 /usr/local/bin/phoenix-agent

# Create systemd service
sudo tee /etc/systemd/system/phoenix-agent.service << EOF
[Unit]
Description=Phoenix Agent
After=network.target

[Service]
Type=simple
User=phoenix
Environment="PHOENIX_API_URL=http://phoenix-api:8080"
Environment="PHOENIX_HOST_ID=%H"
ExecStart=/usr/local/bin/phoenix-agent
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF

# Start agent
sudo systemctl enable phoenix-agent
sudo systemctl start phoenix-agent
```

## Networking

### Port Requirements

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Phoenix API | 8080 | TCP | REST API + WebSocket |
| Dashboard | 3000 | TCP | Web UI |
| PostgreSQL | 5432 | TCP | Database |
| Prometheus | 9090 | TCP | Metrics |
| Grafana | 3001 | TCP | Dashboards |
| Pushgateway | 9091 | TCP | Metric ingestion |

### Docker Networks

```yaml
networks:
  phoenix-net:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## Scaling

### Horizontal Scaling

Scale API instances:
```bash
# Scale to 3 instances
docker-compose up -d --scale phoenix-api=3
```

### Load Balancing

Add nginx for load balancing:
```yaml
nginx:
  image: nginx:alpine
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - ./nginx.conf:/etc/nginx/nginx.conf
  depends_on:
    - phoenix-api
```

## Backup and Restore

### Backup

```bash
# Run backup script
./scripts/backup.sh

# Manual backup
docker exec postgres pg_dump -U phoenix phoenix | gzip > backup-$(date +%Y%m%d).sql.gz
```

### Restore

```bash
# Restore from backup
./scripts/restore.sh backup-20240115.sql.gz

# Manual restore
gunzip -c backup-20240115.sql.gz | docker exec -i postgres psql -U phoenix phoenix
```

## Monitoring

### Health Checks

```bash
# Check all services
./scripts/health-check.sh

# Check specific service
curl http://localhost:8080/health
```

### Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f phoenix-api

# Export logs
docker-compose logs > phoenix-logs-$(date +%Y%m%d).txt
```

## Security

### TLS Configuration

1. Generate certificates:
```bash
./scripts/generate-certs.sh
```

2. Update docker-compose.yml:
```yaml
phoenix-api:
  volumes:
    - ./certs:/etc/phoenix/certs:ro
  environment:
    - TLS_ENABLED=true
    - TLS_CERT=/etc/phoenix/certs/cert.pem
    - TLS_KEY=/etc/phoenix/certs/key.pem
```

### Firewall Rules

```bash
# Allow required ports
sudo ufw allow 8080/tcp  # API
sudo ufw allow 3000/tcp  # Dashboard
sudo ufw allow 22/tcp    # SSH
sudo ufw enable
```

## Troubleshooting

### Common Issues

1. **Container won't start**
   ```bash
   docker-compose logs <service-name>
   docker-compose ps
   ```

2. **Database connection errors**
   ```bash
   docker exec postgres psql -U phoenix -c "SELECT 1;"
   ```

3. **High memory usage**
   ```bash
   docker stats
   docker-compose restart <service-name>
   ```

### Debug Mode

Enable debug logging:
```yaml
environment:
  - LOG_LEVEL=debug
  - DEBUG=true
```

## Maintenance

### Updates

```bash
# Pull latest images
docker-compose pull

# Restart with new images
docker-compose up -d
```

### Cleanup

```bash
# Remove stopped containers
docker-compose down

# Remove all data (WARNING: destructive)
docker-compose down -v

# Prune unused resources
docker system prune -a
```

## Performance Tuning

### Docker Daemon

Edit `/etc/docker/daemon.json`:
```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 64000,
      "Soft": 64000
    }
  }
}
```

### Resource Limits

Set in docker-compose.yml:
```yaml
services:
  phoenix-api:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
```

## Best Practices

1. **Use named volumes** for persistent data
2. **Set resource limits** to prevent resource exhaustion
3. **Enable health checks** for all services
4. **Use secrets** for sensitive data
5. **Regular backups** of PostgreSQL data
6. **Monitor disk usage** and set up alerts
7. **Keep images updated** with security patches
8. **Use .env files** for configuration
9. **Enable TLS** for production deployments
10. **Set up log rotation** to prevent disk fill

## Next Steps

- [Production Checklist](production-checklist.md)
- [Monitoring Setup](monitoring-setup.md)
- [Disaster Recovery](disaster-recovery.md)
- [Security Hardening](security-hardening.md)