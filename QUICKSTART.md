# 🚀 Phoenix Platform - Quick Start Guide

Get Phoenix Platform up and running in 5 minutes!

For a walkthrough of the entire deployment and experiment lifecycle see the [Operations Guide](./docs/operations/OPERATIONS_GUIDE_COMPLETE.md).

## Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for development)
- 8GB RAM minimum
- Ports: 8080, 5432, 6379, 4222, 16686

## 🏃 Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/phoenix/platform.git
cd platform

# Start infrastructure services
make dev-up
```

### 2. Run Phoenix API

```bash
# Option A: Using the demo service
cd projects/hello-phoenix
go run main.go

# Option B: Using make (from root)
make run-hello-phoenix
```

### 3. Verify Installation

```bash
# Check health
curl http://localhost:8080/health

# Get platform info
curl http://localhost:8080/info | jq .
```

## 🎯 Try These Features

### View Active Experiments
```bash
curl http://localhost:8080/api/v1/experiments | jq .
```

Example output:
```json
{
  "experiments": [
    {
      "id": "exp-001",
      "name": "Reduce Prometheus Metrics",
      "status": "running",
      "cost_saving_percent": 45.2
    }
  ]
}
```

### Check Cost Savings
```bash
curl http://localhost:8080/api/v1/metrics | jq .
```

Example output:
```json
{
  "monthly_savings_usd": 45000,
  "cardinality_reduction": "87%",
  "metrics_processed": 1234567
}
```

## 🛠️ Development Workflow

### 1. Create New Service
```bash
# Use the project generator
./scripts/create-project.sh my-service

# Or copy template
cp -r projects/template projects/my-service
```

### 2. Add to Workspace
```bash
# Add to go.work
echo "use ./projects/my-service" >> go.work
go work sync
```

### 3. Build and Test
```bash
cd projects/my-service
make build
make test
make run
```

## 📊 Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| Phoenix API | http://localhost:8080 | - |
| Jaeger UI | http://localhost:16686 | - |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/phoenix |
| PostgreSQL | localhost:5432 | phoenix/phoenix |
| Redis | localhost:6379 | phoenix |

## 🔍 Example: Create an Experiment

```bash
# Create new optimization experiment
curl -X POST http://localhost:8080/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Optimize Datadog Metrics",
    "description": "Reduce Datadog costs by optimizing tag cardinality",
    "baseline_pipeline": "datadog-standard",
    "candidate_pipeline": "datadog-optimized",
    "target_namespaces": ["production", "staging"]
  }'
```

## 🎮 Demo Commands

```bash
# Run the full demo
./scripts/demo-phoenix.sh

# Simulate experiment workflow
./examples/experiment-simulation.sh

# Check system status
./scripts/test-system.sh
```


## 🖥️ Running Collectors on a VM

Generate a static collector config and run it with systemd:

```bash
# Create the configuration
phoenix pipeline vm-config process-topk-v1 \
  --exporter-endpoint otel-phoenix.example.com:4317 \
  --output /etc/otelcol/collector.yaml

# Start the service
sudo systemctl daemon-reload
sudo systemctl enable --now otelcol
```
=======
## 🌐 Full Workflow

1. **Deploy a pipeline**
   ```bash
   curl -X POST http://localhost:8080/api/v1/pipeline-deployments \
     -H "Content-Type: application/json" \
     -d '{"name":"demo","namespace":"default","template":"process-baseline-v1"}'
   ```
2. **Create an experiment**
   ```bash
   curl -X POST http://localhost:8080/api/v1/experiments \
     -H "Content-Type: application/json" \
     -d '{"name":"cost-opt","baseline_pipeline":"process-baseline-v1","candidate_pipeline":"process-intelligent-v1","target_namespaces":["default"]}'
   ```
3. **Analyze results**
   ```bash
   curl http://localhost:8080/api/v1/experiments/<id>/results | jq .
   ```


## 🏗️ Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Web UI    │     │     CLI     │     │   REST API  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                    │                    │
       └────────────────────┴────────────────────┘
                           │
                    ┌──────▼──────┐
                    │ Platform API │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐     ┌──────▼──────┐   ┌─────▼─────┐
    │Postgres │     │    Redis    │   │   NATS    │
    └─────────┘     └─────────────┘   └───────────┘
```

## 🆘 Troubleshooting

### Port Already in Use
```bash
# Check what's using the port
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Docker Issues
```bash
# Reset Docker environment
docker-compose down -v
docker system prune -af
make dev-up
```

### Build Failures
```bash
# Clean and rebuild
make clean
go clean -cache
go mod tidy
make build
```

## 📚 Next Steps

1. **Explore the API**: See [API Documentation](docs/API_REFERENCE.md)
2. **Deploy to Production**: See [Deployment Guide](docs/DEPLOYMENT_GUIDE.md)
3. **Contribute**: See [Contributing Guide](CONTRIBUTING.md)
4. **Architecture Deep Dive**: See [Architecture Overview](docs/ARCHITECTURE_OVERVIEW.md)

## 🤝 Community

- **GitHub**: [github.com/phoenix/platform](https://github.com/phoenix/platform)
- **Discord**: [discord.gg/phoenix](https://discord.gg/phoenix)
- **Twitter**: [@PhoenixPlatform](https://twitter.com/PhoenixPlatform)

---

**Need help?** Join our Discord or open an issue on GitHub!