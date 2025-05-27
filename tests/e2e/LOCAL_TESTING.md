# Running Phoenix Platform End-to-End Locally

This guide walks you through running the complete Phoenix Platform with the revolutionary UI on your local machine.

## Prerequisites Check

Before starting, ensure you have:
- Docker Desktop installed and running
- Docker Compose v2.0+
- Go 1.21+ (optional, for development)
- Node.js 18+ (optional, for dashboard development)
- 8GB RAM available
- Ports available: 3000, 5432, 6379, 8080, 8081, 9090, 9091, 3001

## Step 1: Initial Setup

```bash
# Clone the repository (if not already done)
git clone https://github.com/phoenix/platform.git phoenix
cd phoenix

# Make scripts executable
chmod +x scripts/*.sh

# Check prerequisites
./scripts/check-prerequisites.sh
```

## Step 2: Start Phoenix with UI

The easiest way to start everything:

```bash
./scripts/start-phoenix-ui.sh
```

This will:
1. Start PostgreSQL and Redis
2. Run database migrations
3. Start Phoenix API with WebSocket support
4. Start demo agents
5. Build and start the dashboard
6. Open your browser automatically

## Step 3: Verify Services

Once started, verify all services are running:

```bash
# Check service status
docker-compose ps

# Test API health
curl http://localhost:8080/health

# Test WebSocket
curl http://localhost:8081/api/v1/ws
```

## Step 4: Access the Dashboard

Open http://localhost:3000 in your browser. You should see:
- Fleet Status showing demo agents
- Live Cost Flow Monitor (simulated data initially)
- Quick Actions panel

## Step 5: Create Your First Experiment

### Option 1: Using the Dashboard

1. Click "New Experiment" button
2. Follow the 3-step wizard:
   - Select hosts (choose demo agents)
   - Pick pipeline template (try "Top-K 20")
   - Review and launch

### Option 2: Using the CLI

```bash
# Using the wizard
phoenix ui wizard

# Or directly via API
curl -X POST http://localhost:8080/api/v1/experiments/wizard \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My First Optimization",
    "host_selector": ["group:demo"],
    "pipeline_type": "top-k-20",
    "duration_hours": 1
  }'
```

## Step 6: Explore Key Features

### 1. Live Cost Flow Monitor
- Watch metrics flowing in real-time
- Click any metric to deploy filters instantly
- See cost breakdowns by service/namespace

### 2. Fleet Management
- View all agents grouped by location/service
- Monitor health and active tasks
- Deploy to groups with one click

### 3. Visual Pipeline Builder
- Access from Pipeline Catalog → "Build Custom"
- Drag and drop processors
- See instant impact preview
- Save and deploy pipelines

### 4. Cardinality Explorer
- Navigate to Cost Analytics → Cardinality Explorer
- Interactive sunburst chart
- Drill down into metric hierarchies
- Find high-cardinality culprits

### 5. Instant Rollback
- Go to any running experiment
- Use the time machine slider
- Preview changes before applying
- Rollback in < 10 seconds

## Step 7: Run Demo Flow

To see all features in action:

```bash
./scripts/demo-ui-flow.sh
```

This demonstrates:
- Fleet status visualization
- Cost flow monitoring
- Experiment creation via wizard
- Pipeline impact preview
- Quick deployment
- Real-time analytics

## Step 8: Monitor Real-time Updates

### WebSocket Events
Open developer console in browser and watch WebSocket messages:
- Agent status updates
- Experiment progress
- Metric flow updates
- Task progress

### Prometheus Metrics
Access Prometheus at http://localhost:9090 to query:
```promql
# Cardinality by pipeline
phoenix_pipeline_output_cardinality_estimate

# Agent metrics
phoenix_agent_metrics_per_second

# Cost calculations
phoenix_estimated_cost_per_metric
```

### Grafana Dashboards
Access Grafana at http://localhost:3001 (admin/admin) for:
- Phoenix Overview dashboard
- Agent Performance metrics
- Cost Savings trends

## Step 9: Test Advanced Features

### Quick Deploy
```bash
# Deploy a filter to all demo agents
curl -X POST http://localhost:8080/api/v1/pipelines/quick-deploy \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_template": "priority-sli-slo",
    "target_hosts": ["group:demo"],
    "auto_rollback": true
  }'
```

### Cost Analytics
```bash
# Get weekly cost summary
curl http://localhost:8080/api/v1/cost-analytics?period=7d | jq .
```

### Task Queue Status
```bash
# Check active tasks
curl http://localhost:8080/api/v1/tasks/active | jq .

# Queue status
curl http://localhost:8080/api/v1/tasks/queue | jq .
```

## Step 10: Development Workflow

For active development:

```bash
# Backend development
make ui-dev  # Starts backend services
cd projects/phoenix-api
go run cmd/api/main.go

# Frontend development
cd projects/dashboard
npm install
npm run dev

# Run tests
make ui-test
```

## Troubleshooting

### Common Issues

1. **Port conflicts**
   ```bash
   # Check what's using ports
   lsof -i :3000
   lsof -i :8080
   
   # Stop conflicting services or change ports in docker-compose.yml
   ```

2. **Database connection issues**
   ```bash
   # Reset database
   docker-compose down -v
   docker-compose up -d postgres
   
   # Re-run migrations
   docker-compose exec phoenix-api /app/migrate up
   ```

3. **Dashboard build fails**
   ```bash
   # Clean and rebuild
   cd projects/dashboard
   rm -rf node_modules package-lock.json
   npm install
   npm run build
   ```

4. **Agents not appearing**
   ```bash
   # Check agent logs
   docker-compose logs phoenix-agent
   
   # Restart agents
   docker-compose restart phoenix-agent
   ```

### Logs and Debugging

```bash
# View all logs
docker-compose logs -f

# Specific service logs
docker-compose logs -f phoenix-api
docker-compose logs -f phoenix-agent

# API logs with debug level
PHOENIX_LOG_LEVEL=debug docker-compose up phoenix-api
```

## Cleanup

When done testing:

```bash
# Stop all services
docker-compose down

# Remove all data (full cleanup)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## Next Steps

1. **Explore the codebase**: Check `/projects` for service implementations
2. **Customize pipelines**: Create your own processors in the visual builder
3. **Deploy to production**: Follow the Single VM deployment guide
4. **Contribute**: See CONTRIBUTING.md for guidelines

## Summary

You now have the complete Phoenix Platform running locally with:
- ✅ Phoenix API with WebSocket support
- ✅ Revolutionary Dashboard UI
- ✅ Demo agents generating metrics
- ✅ Real-time cost monitoring
- ✅ Visual pipeline builder
- ✅ Instant experiment creation
- ✅ Complete observability stack

The platform is ready for you to explore cost optimization strategies and experience the revolutionary UI that makes complex operations simple!

For questions or issues, check the troubleshooting section or open an issue on GitHub.