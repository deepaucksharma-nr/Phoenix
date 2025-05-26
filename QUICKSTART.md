# Phoenix Platform Quick Start

Get Phoenix up and running in under 5 minutes.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 8GB RAM minimum
- Ports 3000, 8080, 9090, 9091 available

## 🚀 One-Command Start

```bash
# Clone and start Phoenix
git clone https://github.com/phoenix/platform.git
cd phoenix
./scripts/run-phoenix.sh
```

## 📍 Access Points

After startup (~30 seconds):

- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)

## 🧪 Create Your First Experiment

1. Open the Dashboard at http://localhost:3000
2. Click "Create Experiment"
3. Select target hosts (agents will auto-register)
4. Choose optimization templates:
   - **Baseline**: Standard metrics collection
   - **Optimized**: 90% cardinality reduction
5. Start the experiment and watch real-time results

## 🛠️ Manual Setup

If you prefer manual control:

```bash
# Start infrastructure
docker-compose up -d postgres redis prometheus pushgateway

# Start Phoenix API
docker-compose up -d phoenix-api

# Start agents on target hosts
docker-compose up -d phoenix-agent

# Start dashboard
docker-compose up -d phoenix-dashboard
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file:

```bash
# Phoenix API
PHOENIX_API_URL=http://localhost:8080
DATABASE_URL=postgresql://phoenix:phoenix@localhost:5432/phoenix

# Monitoring
PROMETHEUS_URL=http://localhost:9090
PUSHGATEWAY_URL=http://localhost:9091

# Optional
ENABLE_AUTH=false
LOG_LEVEL=info
```

### Agent Configuration

Agents self-register with the API. Deploy on target hosts:

```bash
# On each target host
docker run -d \
  --name phoenix-agent \
  -e PHOENIX_API_URL=http://phoenix-api:8080 \
  -e HOST_ID=$(hostname) \
  phoenix/agent:latest
```

## 📊 Verify Installation

Check system health:

```bash
# Check services
docker-compose ps

# Verify API
curl http://localhost:8080/health

# Check agent registration
curl http://localhost:8080/api/v1/agents
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