# Phoenix Platform Quick Start Guide

## 🚀 Getting Started in 5 Minutes

This guide helps you quickly understand and start working with the Phoenix platform codebase.

## 📋 What is Phoenix?

Phoenix is an **observability cost optimization platform** that:
- Reduces process metrics volume by 50-80%
- Maintains 100% visibility for critical processes
- Uses OpenTelemetry collectors with A/B testing
- Provides visual pipeline configuration

## 🏗️ Project Structure

```
Phoenix/
├── phoenix-platform/          # Main platform code
│   ├── cmd/                  # Service entry points
│   │   ├── api/             # REST/gRPC API server ⭐
│   │   ├── controller/      # Experiment controller 🚧
│   │   ├── generator/       # Config generator 🚧
│   │   └── simulator/       # Process simulator 🚧
│   ├── dashboard/           # React web UI ⭐
│   ├── operators/           # Kubernetes operators 🚧
│   ├── pkg/                 # Shared libraries
│   └── docs/               # Technical documentation
├── docs/                    # Governance & specifications
└── CLAUDE.md               # AI assistant guide
```

**Legend:** ⭐ = Partially implemented, 🚧 = Needs implementation

## 🛠️ Development Setup

### Prerequisites
```bash
# Required tools
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Kubernetes (kind/minikube for local)
- Make
```

### Quick Local Setup
```bash
# 1. Clone the repository
git clone <repo-url>
cd Phoenix/phoenix-platform

# 2. Install dependencies
make install-deps

# 3. Copy environment template
cp .env.example .env  # Note: Need to create .env.example

# 4. Start development services
docker-compose up -d postgres redis

# 5. Run database migrations
make migrate  # Note: Migrations need to be created

# 6. Start services (in separate terminals)
make run-api
make run-dashboard
```

## 🔑 Key Concepts

### 1. **Experiments**
- A/B tests comparing two OTel pipeline configurations
- Run baseline vs optimized collectors side-by-side
- Measure metrics reduction and performance impact

### 2. **Pipelines**
- OpenTelemetry collector configurations
- Visual builder creates YAML configs
- Pre-validated templates available

### 3. **GitOps Deployment**
- All configs stored in Git
- ArgoCD deploys to Kubernetes
- Automatic rollback on failures

## 🎯 Common Development Tasks

### Working on the API Service
```bash
cd phoenix-platform/cmd/api
go run main.go

# Key files:
# - main.go: Service entry point
# - ../../pkg/api/experiment_service.go: Core logic
# - ../../proto/experiment.proto: API definitions
```

### Working on the Dashboard
```bash
cd phoenix-platform/dashboard
npm install
npm run dev

# Key files:
# - src/App.tsx: Main application
# - src/components/ExperimentBuilder/: Pipeline builder
# - src/services/api.service.ts: API client
```

### Creating a New Pipeline Template
```yaml
# phoenix-platform/pipelines/templates/my-pipeline.yaml
receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      process:
        include:
          match_type: regexp
          names: ["critical-.*"]

processors:
  batch:
    timeout: 10s

exporters:
  otlp:
    endpoint: "collector:4317"

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [batch]
      exporters: [otlp]
```

## 📝 Current Implementation Status

### ✅ What's Working
- Basic project structure
- API service skeleton
- Dashboard React setup
- Kubernetes CRDs defined
- Pipeline templates (3 basic ones)

### 🚧 What Needs Implementation
- **Experiment Controller**: Core business logic
- **Config Generator**: Template processing
- **Pipeline Operator**: Kubernetes deployment
- **Process Simulator**: Test data generation
- **Testing**: Unit and integration tests
- **CI/CD**: GitHub Actions pipeline

## 🔍 Where to Start Contributing

### Easy First Tasks
1. **Add Unit Tests**: Start with pkg/api/
2. **Complete API Endpoints**: Implement missing REST endpoints
3. **Dashboard Components**: Build missing UI components
4. **Documentation**: Update component READMEs

### Medium Complexity
1. **Controller State Machine**: Implement experiment lifecycle
2. **Pipeline Validation**: Add config validation logic
3. **Metrics Collection**: Add Prometheus instrumentation

### Complex Tasks
1. **Kubernetes Operator**: Complete reconciliation logic
2. **Visual Pipeline Builder**: React Flow integration
3. **A/B Testing Engine**: Implement comparison logic

## 📚 Essential Documentation

### Must Read First
1. **CLAUDE.md**: Understand the codebase structure
2. **docs/PRODUCT_REQUIREMENTS.md**: Know what we're building
3. **docs/architecture.md**: Understand system design

### For Implementation
1. **docs/TECHNICAL_SPEC_*.md**: Detailed component specs
2. **docs/STATIC_ANALYSIS_RULES.md**: Code standards
3. **docs/MONO_REPO_GOVERNANCE.md**: Development process

## 🧪 Testing

### Run Tests (when implemented)
```bash
# Unit tests
make test

# Integration tests
make test-integration

# E2E tests (requires running cluster)
make test-e2e
```

### Current Test Status
- ⚠️ **No tests implemented yet**
- Test structure needs creation
- Testing framework not selected

## 🚀 Deployment

### Local Kubernetes (kind)
```bash
# Create cluster
kind create cluster --name phoenix-dev

# Deploy CRDs
kubectl apply -f k8s/crds/

# Deploy with Helm (when ready)
helm install phoenix helm/phoenix/
```

### Production
- Uses ArgoCD for GitOps
- Configurations in separate Git repo
- Prometheus & Grafana for monitoring

## ❓ FAQ

### Q: Why is so much not implemented?
A: The project is in early development. Documentation was created first to establish clear specifications.

### Q: Where should I start?
A: Pick a component you're comfortable with (API, Dashboard, or Operators) and check its TECHNICAL_SPEC_*.md file.

### Q: How do experiments work?
A: Two OTel collectors run simultaneously (baseline and optimized), metrics are compared, and the better one is promoted.

### Q: What's the main challenge?
A: Balancing metrics reduction with maintaining visibility for critical processes.

## 🆘 Getting Help

1. **Check Documentation**: Most answers are in docs/
2. **Read CLAUDE.md**: Has common commands and patterns
3. **Review Technical Specs**: Detailed implementation guides
4. **Check PROJECT_STATUS.md**: Current state and blockers

## 🎯 Next Steps

1. **Pick a Component**: Choose what interests you
2. **Read its Spec**: Understand requirements
3. **Check Status**: See what's implemented
4. **Start Small**: Make incremental improvements
5. **Test Thoroughly**: Add tests for new code
6. **Document Changes**: Update relevant docs

---

Remember: The vision is comprehensive, but implementation is incremental. Every contribution moves us closer to the goal of affordable observability!