# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Phoenix Observability Platform Overview

Phoenix is an automated, dashboard-driven solution for optimizing process metrics collection in New Relic Infrastructure monitoring. The platform reduces telemetry costs by 50-80% while maintaining 100% visibility for critical processes through intelligent OpenTelemetry pipelines with pre-validated configurations, A/B testing capabilities, and visual pipeline building.

**Current Status**: Early development (25% complete) - Documentation complete, implementation in progress

## Architecture

### Core Components
- **Web Dashboard**: React/TypeScript SPA with visual pipeline builder (drag-and-drop interface)
- **API Gateway**: External interface for dashboard communication
- **Experiment API**: Core business logic for experiment management (Go gRPC/REST)
- **Config Generator**: Creates OTel configs from visual pipelines
- **Experiment Controller**: Manages A/B test lifecycle
- **Git Repository + ArgoCD**: GitOps deployment of experiments
- **OTel Collectors**: Baseline & Candidate collectors for A/B testing
- **Observability Stack**: Prometheus (local metrics), New Relic (production export)
- **Process Simulator**: Generates realistic process workloads for testing

### Key Design Patterns
1. **Visual Pipeline Builder**: React Flow-based drag-drop interface for creating OTel configurations
2. **GitOps Deployment**: All configurations stored in Git with ArgoCD handling deployment
3. **Dual-Collector A/B Testing**: Run baseline and candidate collectors side-by-side
4. **Dual Metrics Export**: Collectors export to both Prometheus and New Relic for comparison
5. **CRD-Based Configuration**: Uses `PhoenixProcessPipeline` CRD for Kubernetes deployments
6. **Catalog-Based Pipelines**: Pre-validated pipeline configurations in `/pipelines/templates/`

## Development Commands

### Prerequisites Setup
```bash
# Navigate to the Phoenix platform directory (project is in subdirectory)
cd phoenix-platform/

# Install development dependencies
make install-deps

# Setup pre-commit hooks (when implemented)
make setup-hooks

# Quick development setup
./scripts/setup-dev.sh  # Note: Creates .env, starts PostgreSQL/Redis
```

### Build Commands
```bash
# From phoenix-platform/ directory
# Build all components
make build

# Build specific components
make build-api
make build-dashboard
make build-operators

# Build Docker images
make docker

# Generate code and CRDs
make generate
```

### Testing
```bash
# NOTE: Tests are not yet implemented (0% coverage)
# These commands will work once test framework is added:

# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration
make test-e2e

# Run linters
make lint

# Format code
make fmt
```

### Local Development
```bash
# Start local Kubernetes cluster (using kind)
make cluster-up

# Deploy Phoenix in development mode
make deploy-dev

# Port forward services for local access
make port-forward

# Run development services (postgres, prometheus, grafana)
docker-compose up -d

# Run API server locally
go run cmd/api/main.go

# Run dashboard dev server (in separate terminal)
cd dashboard && npm run dev
```

### Deployment
```bash
# Deploy to local Kind cluster
make deploy

# Deploy to production
export NEW_RELIC_API_KEY=your-key
./scripts/deploy-phoenix.sh production

# Install CRDs only
make install-crds
```

### Phoenix CLI Commands
```bash
# Pipeline Management
phoenix pipeline list --type process
phoenix pipeline deploy <pipeline_name> --target-host <host>
phoenix pipeline status --target-host <host>
phoenix pipeline validate <config.yaml>

# A/B Experiments
phoenix experiment create --scenario <experiment.yaml>
phoenix experiment run <experiment_name>
phoenix experiment status <experiment_name>
phoenix experiment compare <experiment_name>
phoenix experiment promote <experiment_name> --variant <A|B>

# Load Simulation
phoenix loadsim start --profile <realistic|high-cardinality|high-churn> --target-host <host>
phoenix loadsim stop --target-host <host>
```

## Project Structure & Implementation Status

The codebase recently underwent a restructuring. The main implementation is now in the `phoenix-platform/` subdirectory.

**Implementation Status**: 
- API Service: 30% complete (basic structure, needs gRPC/REST implementation)
- Dashboard: 25% complete (React setup done, needs pipeline builder)
- Controller: 5% complete (stub only)
- Generator: 0% complete (stub only)
- Operators: 10% complete (CRDs defined, controller logic needed)
- Tests: 0% coverage

```
phoenix-platform/
├── cmd/                    # Service entry points
│   ├── api/               # API server (partially implemented)
│   ├── controller/        # Experiment controller (stub only)
│   ├── generator/         # Config generator (stub only)
│   └── simulator/         # Process simulator (basic structure)
├── internal/              # Internal packages (mostly empty)
├── pkg/                   # Public packages
│   ├── api/              # API service logic (basic structure)
│   └── generator/        # Config generation (empty)
├── operators/             # Kubernetes operators
│   ├── pipeline/         # Pipeline controller (CRDs defined)
│   └── loadsim/          # Load simulation controller (basic)
├── dashboard/             # React web UI (basic setup)
├── helm/                  # Helm charts (incomplete)
│   └── phoenix/          # Main Phoenix chart
├── k8s/                   # Kubernetes manifests
│   ├── crds/             # Custom Resource Definitions (complete)
│   ├── base/             # Base configurations
│   └── overlays/         # Environment-specific configs
├── pipelines/            # Pipeline catalog
│   └── templates/        # Pre-validated pipeline configs
├── scripts/              # Deployment and dev scripts
└── docker-compose.yaml   # Local development stack
```

### Pipeline Templates
Pre-validated process metric pipelines in `/pipelines/templates/`:
- `process-baseline-v1.yaml`: No optimization, full fidelity (control group)
- `process-priority-filter-v1.yaml`: Filter by process importance (critical/high/low)
- `process-topk-v1.yaml`: Keep only top CPU/memory consumers
- `process-aggregated-v1.yaml`: Roll up common applications
- `process-adaptive-filter-v1.yaml`: Dynamic filtering based on load

### Critical Files to Understand
- `pkg/api/experiment_service.go`: Core business logic for experiments (basic structure only)
- `operators/pipeline/controllers/pipeline_controller.go`: How collectors are deployed (needs implementation)
- `pkg/generator/service.go`: Config generation from visual pipelines (not yet created)
- `dashboard/src/components/ExperimentBuilder/PipelineCanvas.tsx`: Visual pipeline editor (basic component)
- `cmd/simulator/main.go`: Process simulation logic (stub only)

## Service Communication Flow

1. **Dashboard → API Gateway**: React app communicates via REST/WebSocket
2. **API Gateway → Experiment API**: Internal gRPC with mTLS
3. **Experiment API → Config Generator**: Async job triggering
4. **Config Generator → Git**: Creates PRs with generated configs
5. **ArgoCD → Kubernetes**: GitOps deployment of experiments
6. **Operators → Kubernetes API**: Manage DaemonSets and Jobs

## Experiment Lifecycle

1. User creates experiment through dashboard
2. API validates and stores experiment spec in PostgreSQL
3. Generator creates OTel configs and K8s manifests
4. Git PR created with all artifacts
5. ArgoCD syncs and deploys resources
6. Pipeline operator creates collector DaemonSets
7. LoadSim operator runs process simulator jobs
8. Metrics flow to Prometheus and New Relic
9. Analysis service compares baseline vs candidate
10. User promotes winning variant or stops experiment

## Key Configurations

### PhoenixProcessPipeline CRD
```yaml
apiVersion: phoenix.newrelic.com/v1alpha1
kind: PhoenixProcessPipeline
metadata:
  name: my-host-process-config
spec:
  nodeSelector:
    kubernetes.io/hostname: "node-1"
  pipelineCatalogRef: "process-topk-v1"
  configVariables:
    NEW_RELIC_API_KEY_SECRET_NAME: "nr-api-key-secret"
```

### Environment Variables
Required for development (from `.env` or K8s secrets):
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Authentication secret  
- `NEW_RELIC_API_KEY`: For OTLP export to New Relic
- `NEW_RELIC_OTLP_ENDPOINT`: Default: https://otlp.nr-data.net:4317
- `GIT_TOKEN`: For creating configuration PRs
- `PROMETHEUS_REMOTE_WRITE_ENDPOINT`: For metrics storage

## Testing Strategy (To Be Implemented)

### Acceptance Tests
Key acceptance tests for MVP (AT-P01 to AT-P10) - not yet implemented:
- Pipeline deployment verification
- Critical process retention (100%)
- Cardinality reduction (≥50%)
- A/B experiment functionality
- Collector overhead (<5% CPU)

### Load Profiles (Planned)
- **realistic**: 50-200 processes with normal churn
- **high-cardinality**: 1000-2000 mostly idle processes
- **high-churn**: 20-30 processes/sec creation rate

## Performance Targets
- Deployment time: ≤10 min/host
- Cardinality reduction: 50-80%
- Critical process retention: 100%
- Cost savings: 40-70%
- Collector overhead: <5% CPU/core
- Processing latency: <50ms P99
- API response time: <100ms
- Pipeline generation time: <5s
- Concurrent experiments: 100+
- Nodes per experiment: 1000+
- Processes per node: 500+
- Unique time series: 3.5M+

## Current Implementation Gaps

### Critical Missing Components
1. **Experiment Controller**: State machine and business logic not implemented
2. **Config Generator**: No template engine or pipeline optimization
3. **Pipeline Operator**: Reconciliation logic incomplete
4. **Testing**: 0% test coverage, no test framework
5. **CI/CD**: No automated build/test/deploy pipeline

### Key Integration Points Needed
1. API Service → Controller communication
2. Controller → Generator triggering
3. Dashboard → API authentication
4. Service → Database connections
5. Operators → Kubernetes API interactions

## Important Documentation

### For Understanding the Vision
- `docs/PRODUCT_REQUIREMENTS.md`: Complete PRD with acceptance criteria
- `docs/architecture.md`: System design and data flow
- `docs/TECHNICAL_SPEC_*.md`: Detailed specifications for each component

### For Implementation Status
- `PROJECT_STATUS.md`: Real-time tracking of what's built vs planned
- `IMPLEMENTATION_ROADMAP.md`: 12-week plan to complete the platform
- `QUICK_START_GUIDE.md`: Fast onboarding for new developers

### For Code Quality
- `docs/STATIC_ANALYSIS_RULES.md`: Coding standards and enforcement
- `docs/MONO_REPO_GOVERNANCE.md`: Development workflows and practices

## Integration Points
- **New Relic**: OTLP endpoint for process metrics export
- **Prometheus**: Collector metrics and benchmarking data
- **Grafana**: Visualization dashboards
- **Kubernetes**: CRD-based configuration management
- **ArgoCD**: GitOps deployment of experiments
- **Git Repository**: Configuration storage and versioning
- **External Secrets Operator**: Secret management
- **PostgreSQL**: Experiment metadata storage
- **Redis**: Caching and session management
- **MinIO**: Object storage for artifacts

## Helm Installation

```bash
# Add required Helm repositories
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Create namespace
kubectl create namespace phoenix-system

# Install Phoenix
helm install phoenix ./helm/phoenix \
  --namespace phoenix-system \
  --set global.domain=phoenix.example.com \
  --set newrelic.apiKey.secretName=newrelic-secret
```

## Working with Pipeline Templates

Pipeline templates are in `pipelines/templates/`. When modifying pipelines:
1. Processors execute in order listed in service.pipelines
2. Always include memory_limiter first
3. batch processor should be last
4. Use transform processor for classification
5. Use filter processor for dropping metrics
6. Use groupbyattrs for aggregation

Example processor ordering:
```yaml
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [memory_limiter, transform/classify, filter/priority, groupbyattrs/aggregate, batch]
      exporters: [otlphttp/newrelic, prometheus]
```

## Common Issues and Solutions

- **Collector pods not starting**: Check ConfigMap exists and New Relic secret is created
- **High memory usage**: Adjust memory_limiter processor settings
- **Missing metrics**: Verify process include/exclude patterns in receiver
- **Pipeline validation fails**: Check processor ordering and required fields
- **Experiment not deploying**: Verify Git token permissions and ArgoCD sync status
- **Dashboard connection issues**: Check JWT_SECRET matches between API and dashboard

## Security Features

- JWT-based authentication for dashboard
- RBAC authorization in Kubernetes
- Network policies for pod-to-pod communication
- TLS encryption for all external communication
- mTLS for internal service communication
- Secret management via External Secrets Operator