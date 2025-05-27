# Phoenix Platform Quick Start

Get Phoenix up and running in under 5 minutes.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 8GB RAM minimum
- Ports 3000, 8080 (API + WebSocket) available

## 🚀 One-Command Start

```bash
# Clone and start Phoenix
git clone https://github.com/phoenix/platform.git
cd platform

# Run the single-VM setup
./deployments/single-vm/scripts/setup-single-vm.sh

# Start all services
cd deployments/single-vm
docker-compose up -d
```

## 📍 Access Points

After startup (~30 seconds):

- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080 (REST + WebSocket)
- **Prometheus**: http://localhost:9090 (if using Prometheus backend)
- **Grafana**: http://localhost:3001 (admin/admin)

## 🧪 Create Your First Experiment

1. Open the Dashboard at http://localhost:3000
2. Click "Create Experiment"
3. Select target hosts (agents will auto-register)
4. Choose optimization templates:
   - **Baseline**: Standard metrics collection
   - **Candidate**: Choose from Adaptive Filter, TopK, or Hybrid
   - Compare A/B test results with 70% cost reduction potential
5. Start the experiment and watch real-time results

## 🛠️ Manual Setup

If you prefer manual control:

```bash
# Navigate to deployment directory
cd deployments/single-vm

# Start infrastructure
docker-compose up -d postgres prometheus grafana

# Wait for database to be ready
./scripts/wait-for-postgres.sh

# Start Phoenix API (includes WebSocket server)
docker-compose up -d phoenix-api

# Start dashboard
docker-compose up -d phoenix-dashboard

# Install agents on target hosts
./scripts/install-agent.sh
```

## 🔧 Configuration

### Environment Variables

The setup script creates `.env` in `deployments/single-vm/`:

```bash
# Phoenix API
PHOENIX_API_URL=http://localhost:8080
DATABASE_URL=postgresql://phoenix:phoenix@postgres:5432/phoenix

# Security (update these!)
JWT_SECRET=change-me-in-production
POSTGRES_PASSWORD=change-me-in-production

# Agent Authentication
AGENT_HOST_ID=$(hostname)

# Collector Selection (choose one)
# For OpenTelemetry Collector (default)
COLLECTOR_TYPE=otel
OTEL_COLLECTOR_ENDPOINT=http://localhost:4317

# For New Relic NRDOT Collector
# COLLECTOR_TYPE=nrdot
# NRDOT_OTLP_ENDPOINT=https://otlp.nr-data.net:4317
# NEW_RELIC_LICENSE_KEY=your-license-key-here

# Optional
ENABLE_AUTH=true
LOG_LEVEL=info
TASK_POLL_INTERVAL=30s
```

### Using NRDOT Collector

To use New Relic's optimized OpenTelemetry distribution:

```bash
# Set environment variables
export COLLECTOR_TYPE=nrdot
export NEW_RELIC_LICENSE_KEY=your-license-key
export NRDOT_OTLP_ENDPOINT=https://otlp.nr-data.net:4317

# Update agent configuration
sudo systemctl edit phoenix-agent

# Add these environment variables:
[Service]
Environment="COLLECTOR_TYPE=nrdot"
Environment="NEW_RELIC_LICENSE_KEY=your-license-key"
Environment="NRDOT_OTLP_ENDPOINT=https://otlp.nr-data.net:4317"

# Restart agent
sudo systemctl restart phoenix-agent
```

### Agent Installation

Agents run as systemd services on host machines:

```bash
# On each target host
curl -sSL http://your-phoenix-api:8080/install-agent.sh | sudo bash

# Or manually:
sudo ./deployments/single-vm/scripts/install-agent.sh \
  --api-url http://your-phoenix-api:8080 \
  --host-id $(hostname)

# Check agent status
sudo systemctl status phoenix-agent
sudo journalctl -u phoenix-agent -f
```

## 📊 Verify Installation

Check system health:

```bash
# Check services
docker-compose ps

# Verify API
curl http://localhost:8080/api/v2/health

# Check agent registration
curl http://localhost:8080/api/v2/agents

# View active experiments
curl http://localhost:8080/api/v2/experiments
```

## 🚨 Troubleshooting

### Services not starting?
```bash
# Check logs
docker-compose logs phoenix-api
docker-compose logs phoenix-agent

# Restart services
docker-compose restart
```

### Port conflicts?
```bash
# Change ports in docker-compose.yml
# Or stop conflicting services
sudo lsof -i :8080
```

### Database issues?
```bash
# Reset database
docker-compose down -v
docker-compose up -d
```

## 📚 Next Steps

- [Development Guide](DEVELOPMENT_GUIDE.md) - Set up development environment
- [Architecture Overview](docs/architecture/PLATFORM_ARCHITECTURE.md) - Understand the system
- [API Documentation](docs/api/) - Integrate with Phoenix
- [Operations Guide](docs/operations/OPERATIONS_GUIDE_COMPLETE.md) - Production deployment

## 💬 Get Help

- [GitHub Issues](https://github.com/phoenix/platform/issues)
- [Discord Community](https://discord.gg/phoenix)
- [Documentation](docs/)