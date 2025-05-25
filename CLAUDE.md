# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL: Documentation Placement Rules

**NEVER create documentation at the repository root level!** 
- Phoenix-specific docs go in: `phoenix-platform/docs/`
- Repository governance docs go in: `docs/`
- Service-specific docs go in: `<service>/docs/`
- ONLY exception: This CLAUDE.md file must remain at root

See `docs/DOCUMENTATION_GOVERNANCE.md` for strict enforcement rules.

## Architectural Integrity Guidelines

### Key Principles for Maintaining Structural Integrity
1. Always preserve the existing folder structure
2. Avoid introducing new top-level directories
3. Keep all code within the `phoenix-platform/` subdirectory
4. Follow mono-repo governance rules strictly
5. Do not create files at the repository root
6. Ensure all updates align with existing architectural patterns
7. Maintain clear separation between services
8. Use GitOps for all configuration changes
9. Validate structural changes against mono-repo governance
10. Prioritize code organization and predictability

### Anti-Drift Measures
- Regularly run `make validate-structure` to catch potential architectural deviations
- Review all changes against `docs/MONO_REPO_GOVERNANCE.md`
- Use existing template files and patterns for new implementations
- Consult architecture documentation before making significant changes

### Update Instructions
- Prefer modifying existing files over creating new ones
- If a new component is required, use the existing service template
- Always update documentation to reflect architectural changes
- Ensure new code follows existing patterns and guidelines

## Phoenix Platform Context (January 2025)

### Project Overview
Phoenix is an observability cost optimization platform that reduces metrics volume by 50-80% through intelligent OpenTelemetry pipeline optimization. It uses A/B testing between baseline and candidate configurations without requiring a service mesh.

Phoenix-vNext (root) is a production-ready 3-Pipeline Cardinality Optimization System for OpenTelemetry metrics collection and processing. The system uses adaptive cardinality management with dynamic switching between optimization profiles (conservative, balanced, aggressive) based on metric volume and system performance through a PID-like control algorithm implemented in Go.

### Key Architectural Decisions (ADRs)
1. **No Service Mesh** (ADR-001): Use dual collectors pattern instead
2. **GitOps Mandatory** (ADR-002): All deployments via ArgoCD, no direct kubectl
3. **Visual Pipeline Builder** (ADR-003): Drag-drop as primary configuration interface
4. **Mono-Repo Boundaries** (ADR-004): Strict service separation enforced
5. **Dual Metrics Export** (ADR-005): Export to both Prometheus and New Relic

### Implementation Status
- **Foundation Phase**: 100% Complete
  - Proto definitions for all services
  - Client libraries with examples
  - Validation scripts enforcing architecture
  - Database migrations framework
  - Pre-commit hooks for automated checks

- **Core Services**: 60% Complete
  - Experiment Controller: 80% (state machine, DB integration)
  - Config Generator: 80% (template engine, manifest generation)
  - Pipeline Operator: 85% (full reconciliation, DaemonSet management)
  - API Service: 30% (proto definitions ready, implementation pending)

### Critical Files & Locations
- **Proto Definitions**: `phoenix-platform/api/proto/v1/`
- **Client Libraries**: `phoenix-platform/pkg/clients/`
- **Validation Scripts**: `phoenix-platform/scripts/validate/`
- **Database Migrations**: `phoenix-platform/migrations/`
- **Service Implementations**: `phoenix-platform/services/`
- **Kubernetes Operators**: `phoenix-platform/operators/`
- **Architecture Docs**: `phoenix-platform/docs/adr/`
- **Implementation Plans**: `phoenix-platform/docs/planning/`

### Development Workflow
1. **Before Making Changes**:
   - Run `make validate` to check structure
   - Review relevant ADRs in `docs/adr/`
   - Check `docs/planning/IMPLEMENTATION_CHECKLIST.md`

2. **Proto Changes**:
   - Edit files in `api/proto/v1/`
   - Run `make generate-proto`
   - Update client libraries if needed

3. **Service Development**:
   - Use existing service structure as template
   - Follow interface definitions in proto files
   - Use client libraries for service communication
   - Add to validation scripts if creating new service

4. **Testing**:
   - Unit tests in `<service>/internal/*/test.go`
   - Integration tests in `<service>/test/integration/`
   - E2E tests in `phoenix-platform/test/e2e/`

5. **Validation**:
   - `make validate-structure`: Check mono-repo structure
   - `make validate-imports`: Verify Go import rules
   - `make validate-services`: Check service boundaries
   - `make validate`: Run all checks

## Architecture (Root Phoenix-vNext System)

### Core System Components
- **Main Collector** (`otelcol-main`): Runs 3 parallel pipelines with different cardinality optimization levels using shared processing (40% overhead reduction)
- **Observer Collector** (`otelcol-observer`): Control plane that monitors pipeline metrics and system performance
- **Control Actuator** (`control-actuator-go`): Go-based PID controller with hysteresis and stability management
- **Anomaly Detector** (`anomaly-detector`): Multi-algorithm detection (Z-score, rate of change, pattern matching) with webhook integration
- **Benchmark Controller** (`benchmark`): Performance validation with 4 test scenarios and CI/CD integration
- **Synthetic Generator** (`synthetic-metrics-generator`): Go-based load generator for testing

### Essential Development Commands

### Pipeline Architecture
The system operates 3 distinct pipelines in parallel with shared processing:
1. **Full Fidelity Pipeline** (`pipeline_full_fidelity`) - Complete metrics baseline without optimization
2. **Optimized Pipeline** (`pipeline_optimised`) - Moderate cardinality reduction with configurable aggregation
3. **Experimental TopK Pipeline** (`pipeline_experimental_topk`) - Advanced optimization using TopK sampling techniques

### Adaptive Control System
- Observer monitors `phoenix_observer_kpi_store_phoenix_pipeline_output_cardinality_estimate` metrics
- Control actuator applies discrete profile switching based on time series count thresholds:
  - Conservative: < 15,000 time series
  - Balanced: 15,000 - 25,000 time series  
  - Aggressive: > 25,000 time series
- Control signals written to `configs/control/optimization_mode.yaml` and read by main collector
- PID algorithm: `pidOutput = 0.5*error + 0.1*integral + 0.05*derivative`
- Hysteresis factor (10%) prevents rapid oscillation

### Performance Targets
- Signal preservation: >98%
- Cardinality reduction: 15-40% (mode dependent)
- Control loop latency: <100ms
- Memory usage: <512MB baseline
- P99 processing latency: <50ms

## Development Commands

### Quick Start
```bash
<<<<<<< HEAD
# Initialize environment (creates data dirs, control files, .env from template)
./scripts/initialize-environment.sh

# Start full stack
./run-phoenix.sh

# Or use docker-compose directly
docker-compose up -d

# Stop services
./run-phoenix.sh stop

# Clean everything
./run-phoenix.sh clean
```

### Makefile Commands
```bash
# Main targets
make help                   # Show all available commands
make setup-env             # Initialize environment
make build                 # Build all projects (Turborepo)
make build-docker          # Build all Docker images
make dev                   # Start development mode
make test                  # Run tests
make test-integration      # Run integration tests
make monitor               # Open monitoring dashboards
make clean                 # Clean build artifacts

# Service logs
make collector-logs        # View main collector logs
make observer-logs         # View observer logs
make actuator-logs         # View control actuator logs
make generator-logs        # View generator logs

# Utilities
make validate-config       # Validate YAML configurations
make docs-serve           # Serve documentation locally
```

### Docker Compose Operations
```bash
# Start specific services
docker-compose up -d otelcol-main otelcol-observer prometheus grafana

# Rebuild and restart a specific service
docker-compose build control-actuator-go
docker-compose up -d control-actuator-go

# View logs
docker-compose logs -f otelcol-main
docker-compose logs -f control-actuator-go

# Check service health
curl http://localhost:13133  # Main collector health
curl http://localhost:13134  # Observer health
curl http://localhost:8081/health  # Control actuator health
curl http://localhost:8082/health  # Anomaly detector health
```

### Testing & Validation
```bash
# Run integration tests
./tests/integration/test_core_functionality.sh

# Generate synthetic load
docker-compose up synthetic-metrics-generator

# Run benchmark scenarios
curl http://localhost:8083/benchmark/scenarios  # List scenarios
curl -X POST http://localhost:8083/benchmark/run \
  -H "Content-Type: application/json" \
  -d '{"scenario": "baseline_steady_state"}'

# Monitor control signal changes
watch cat configs/control/optimization_mode.yaml

# Validate configurations
docker-compose config
sha256sum configs/otel/collectors/*.yaml configs/control/*template.yaml > CHECKSUMS.txt
```

### Cloud Deployment
```bash
# AWS EKS deployment
./deploy-aws.sh

# Azure AKS deployment  
./deploy-azure.sh

# Helm deployment
helm install phoenix ./infrastructure/helm/phoenix \
  --namespace phoenix \
  --values ./infrastructure/helm/phoenix/values.yaml

# Terraform deployment
cd infrastructure/terraform/environments/aws
terraform init && terraform apply
```

## Configuration Architecture

### OpenTelemetry Configurations
- `configs/otel/collectors/main.yaml`: Core collector with 3-pipeline configuration
- `configs/otel/collectors/main-optimized.yaml`: Enhanced version with shared processing
- `configs/otel/collectors/observer.yaml`: Monitoring collector that exposes KPI metrics
- `configs/otel/processors/common_intake_processors.yaml`: Shared processor configurations
- `configs/otel/exporters/newrelic-enhanced.yaml`: New Relic OTLP integration

### Control System
- `configs/control/optimization_mode.yaml`: Dynamic control file modified by actuator
- `configs/control/optimization_mode_template.yaml`: Template defining control file schema
- Version tracking with `config_version` field
- Correlation IDs for tracking changes

### Monitoring Stack
- `configs/monitoring/prometheus/prometheus.yaml`: Prometheus scrape configuration
- `configs/monitoring/prometheus/rules/phoenix_comprehensive_rules.yml`: 25+ recording rules
- `configs/monitoring/grafana/`: Datasource and dashboard provisioning
- `monitoring/grafana/dashboards/`: Phoenix dashboards (adaptive control, ultra overview)

## Key Environment Variables

Critical variables in `.env`:
```bash
# Control thresholds for adaptive switching
TARGET_OPTIMIZED_PIPELINE_TS_COUNT=20000
THRESHOLD_OPTIMIZATION_CONSERVATIVE_MAX_TS=15000
THRESHOLD_OPTIMIZATION_AGGRESSIVE_MIN_TS=25000
HYSTERESIS_FACTOR=0.1

# Resource constraints
OTELCOL_MAIN_MEMORY_LIMIT_MIB="1024"
OTELCOL_MAIN_GOMAXPROCS="2"

# Control loop timing
ADAPTIVE_CONTROLLER_INTERVAL_SECONDS=60
ADAPTIVE_CONTROLLER_STABILITY_SECONDS=120

# Load generation
SYNTHETIC_PROCESS_COUNT_PER_HOST=250
SYNTHETIC_HOST_COUNT=3
SYNTHETIC_METRIC_EMIT_INTERVAL_S=15

# New Relic export
NEW_RELIC_LICENSE_KEY=your_key_here
NEW_RELIC_OTLP_ENDPOINT=https://otlp.nr-data.net:4317
ENABLE_NR_EXPORT_FULL="false"
ENABLE_NR_EXPORT_OPTIMISED="false"
ENABLE_NR_EXPORT_EXPERIMENTAL="false"
```

## Service Endpoints & APIs

### Core Service Endpoints
- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Main Collector Metrics**: http://localhost:8888/metrics
- **Optimized Pipeline**: http://localhost:8889/metrics 
- **Experimental Pipeline**: http://localhost:8890/metrics
- **Observer Metrics**: http://localhost:9888/metrics

### API Endpoints
- **Control Actuator API**: http://localhost:8081
  - `GET /metrics` - Control state and metrics
  - `GET /health` - Health check
  - `POST /anomaly` - Webhook for anomaly events
  - `POST /mode` - Force mode change (testing)
- **Anomaly Detector API**: http://localhost:8082
  - `GET /alerts` - Active anomalies
  - `GET /health` - Health check
  - `GET /metrics` - Prometheus metrics
- **Benchmark Controller**: http://localhost:8083
  - `GET /benchmark/scenarios` - List test scenarios
  - `POST /benchmark/run` - Run benchmark
  - `GET /benchmark/results` - Get results
  - `GET /benchmark/validate` - Check SLO compliance

### Health Checks
- Main Collector: http://localhost:13133
- Observer Collector: http://localhost:13134
- Docker health checks configured with 20s intervals, 3 retries

### Debug Endpoints
- pprof: http://localhost:1777/debug/pprof
- zpages: http://localhost:55679

## Control Flow & Data Paths

1. **Metrics Ingestion**: Synthetic generator → Main collector OTLP endpoint (4318)
2. **Pipeline Processing**: 3 parallel processing chains with different optimization levels
3. **Metrics Export**: Each pipeline exports to dedicated Prometheus endpoints (8888-8890)
4. **Monitoring**: Observer scrapes main collector metrics and exposes KPIs (9888)
5. **Control Loop**: Actuator queries observer metrics → calculates profile → updates control file
6. **Adaptation**: Main collector reads control file changes → adjusts pipeline behavior
7. **Anomaly Detection**: Detector monitors metrics → sends webhooks to control actuator
8. **Benchmarking**: Controller generates load patterns → validates performance

## Development Patterns

### Adding New Processors
1. Create processor config in `configs/otel/processors/`
2. Include in pipeline via `configs/otel/collectors/main.yaml`
3. Test with synthetic load and monitor cardinality impact

### Modifying Control Logic
1. Update thresholds in `.env` file
2. Modify PID logic in `apps/control-actuator-go/main.go`
3. Test profile transitions using benchmark scenarios
4. Monitor stability score via recording rules

### Performance Tuning
```bash
# Enable debug logging
export OTEL_LOG_LEVEL=debug
docker-compose up -d otelcol-main

# Profile memory usage
curl http://localhost:1777/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Monitor resource usage
docker-compose top

# Check pipeline efficiency
curl -s http://localhost:9090/api/v1/query?query=phoenix:resource_efficiency_score
```

### Local Development
```bash
# Run Go service locally
cd apps/control-actuator-go
go run main.go

# With live reload
air

# Run tests
go test -v -race ./...

# Build binary
go build -o control-actuator
```

## Monorepo Structure

- **`apps/`**: Go-based microservices (control-actuator, anomaly-detector)
- **`services/`**: Service implementations with Dockerfiles
- **`configs/`**: Technology-grouped configurations (otel, monitoring, control)
- **`infrastructure/`**: Cloud deployment (Terraform, Helm, K8s manifests)
- **`packages/`**: Shared packages (managed by npm workspaces)
- **`scripts/`**: Operational utilities and environment setup
- **`tests/`**: Integration and performance tests
- **`tools/`**: Development and migration utilities
- **`data/`**: Persistent storage directories (gitignored)

## Build System

- **Turborepo**: Parallel builds with caching (`turbo.json`)
- **Make**: Developer-friendly commands (see `make help`)
- **npm workspaces**: Package management
- **Docker multi-stage builds**: Optimized images
- **Go modules**: Dependency management for Go services

## Key Metrics & Alerts

### Recording Rules (25+ rules)
- **Efficiency**: `phoenix:signal_preservation_score`, `phoenix:cardinality_efficiency_ratio`
- **Performance**: `phoenix:pipeline_latency_ms_p99`, `phoenix:pipeline_throughput_metrics_per_sec`
- **Control**: `phoenix:control_stability_score`, `phoenix:control_loop_effectiveness`
- **Anomaly**: `phoenix:cardinality_zscore`, `phoenix:cardinality_explosion_risk`
- **Resource**: `phoenix:resource_efficiency_score`, `phoenix:collector_memory_usage_mb`

### Critical Alerts
- `PhoenixCardinalityExplosion`: Exponential growth detected
- `PhoenixResourceExhaustion`: Memory >90%
- `PhoenixControlLoopInstability`: Frequent mode changes
- `PhoenixSLOViolation`: Service objectives not met

## Benchmark Scenarios

1. **baseline_steady_state**: Normal operation validation
2. **cardinality_spike**: Sudden 3x increase testing
3. **gradual_growth**: Linear growth over time
4. **wave_pattern**: Sinusoidal load pattern

## CI/CD Integration

### GitHub Actions Workflows
- `.github/workflows/ci.yml`: Full CI/CD pipeline
- `.github/workflows/security.yml`: Security scanning (Trivy, Gosec, OWASP)

### Pipeline Stages
1. Configuration validation
2. Go service testing with coverage
3. Integration testing
4. Docker image building
5. Performance benchmarking
6. Deployment (on main branch)

## Troubleshooting

### Common Issues
1. **High memory usage**: Increase `OTELCOL_MAIN_MEMORY_LIMIT_MIB`
2. **Control instability**: Increase `ADAPTIVE_CONTROLLER_STABILITY_SECONDS`
3. **Poor reduction**: Check mode via :8081/metrics, adjust thresholds
4. **Anomaly noise**: Modify detector threshold in `apps/anomaly-detector/main.go`

### Debug Commands
```bash
# Check control decisions
curl http://localhost:8081/metrics | jq '.current_mode'

# View pipeline cardinality
curl -s http://localhost:9090/api/v1/query?query=phoenix_observer_kpi_store_phoenix_pipeline_output_cardinality_estimate

# Force mode change (testing)
curl -X POST http://localhost:8081/mode \
  -H "Content-Type: application/json" \
  -d '{"mode": "aggressive"}'

# Export metrics for analysis
curl "http://localhost:9090/api/v1/query_range?query=phoenix:cardinality_growth_rate&start=$(date -u -d '1 hour ago' +%s)&end=$(date +%s)&step=60"

# Check for memory leaks
docker stats --no-stream
```
# Setup and dependencies
make deps                    # Install Go and npm dependencies
make setup-hooks            # Setup git pre-commit hooks

# Code generation and validation
make generate               # Generate protobuf code and CRDs
make generate-proto         # Generate only protobuf code
make validate              # Run all validation checks
make validate-structure    # Check mono-repo structure
make validate-imports      # Validate Go import rules

# Building
make build                 # Build all components
make build-api            # Build specific service
make build-dashboard      # Build frontend dashboard
make docker               # Build all Docker images

# Testing
make test                 # Run all tests (unit + integration)
make test-unit           # Unit tests only
make test-integration    # Integration tests only
make test-e2e           # End-to-end tests
make test-dashboard     # Dashboard tests with coverage
make coverage           # Generate test coverage report

# Code quality
make fmt                 # Format Go and frontend code
make lint               # Run linters (Go + frontend)
make verify             # Run all pre-commit checks

# Local development
make dev                # Start local development environment
make dev-down          # Stop local development environment
make dev-logs          # Show development environment logs
make dev-status        # Show development environment status

# Kubernetes development
make cluster-up        # Start local kind cluster
make cluster-down      # Stop local kind cluster
make deploy           # Deploy to Kubernetes
make undeploy         # Remove from Kubernetes
make port-forward     # Forward ports for local access

# Utilities
make clean            # Clean build artifacts
make help             # Show all available targets
```

### Project Structure and Architecture

**Core Services** (all in `phoenix-platform/cmd/`):
- `api/` - Main API service (HTTP/REST endpoints)
- `api-gateway/` - HTTP to gRPC gateway with auth middleware  
- `controller/` - Experiment controller with state machine
- `control-service/` - Control plane gRPC service
- `generator/` - Configuration generator for OTel pipelines
- `simulator/` - Process simulator for testing

**Frontend Dashboard** (`phoenix-platform/dashboard/`):
- React/TypeScript with Vite build system
- Material-UI components and React Flow for pipeline builder
- Vitest for testing, ESLint/Prettier for code quality

**Kubernetes Operators** (`phoenix-platform/operators/`):
- `pipeline/` - Manages OTel collector DaemonSets
- `loadsim/` - Manages load simulation jobs

**Protocol Definitions** (`phoenix-platform/api/proto/`):
- Centralized protobuf definitions for all services
- Generated Go code for gRPC clients/servers

### Technology Stack
- **Backend**: Go 1.21, gRPC, PostgreSQL, Kubernetes client-go
- **Frontend**: React 18, TypeScript, Material-UI, React Flow, Vitest
- **Infrastructure**: Kubernetes, Helm, ArgoCD, Prometheus, Grafana
- **Code Quality**: golangci-lint, pre-commit hooks, ESLint, Prettier

### Important Notes
1. **Never bypass validation scripts** - they enforce architectural integrity
2. **All configuration changes must go through GitOps** - no direct kubectl
3. **Service boundaries are strict** - no cross-service imports allowed
4. **Use proto contracts** - all service communication via defined APIs
5. **Follow existing patterns** - consistency is critical

### Critical Development Practices

**Before Making Any Changes:**
1. Run `make validate` to ensure structural integrity
2. Review the Makefile to understand available commands
3. Check existing tests and follow established patterns
4. Always run `make setup-hooks` after cloning to install git hooks

**Code Quality Requirements:**
- All Go code must pass `golangci-lint` (enforced by pre-commit)
- Frontend code must pass ESLint and Prettier formatting
- All changes must pass mono-repo structure validation
- No commits allowed without passing pre-commit hooks

**Testing Strategy:**
- Unit tests: Individual service/component testing
- Integration tests: Service-to-service communication  
- E2E tests: Full workflow testing in Kubernetes
- Dashboard tests: Component and store testing with Vitest

### References
- Technical Spec: `phoenix-platform/docs/TECHNICAL_SPECIFICATION.md`
- Implementation Roadmap: `phoenix-platform/docs/planning/NEXT_STEPS_ACTION_PLAN.md`
- Project Status: `phoenix-platform/docs/planning/PROJECT_STATUS.md`
- Mono-Repo Rules: `docs/MONO_REPO_GOVERNANCE.md`

## Development Commands (Root Phoenix-vNext)

### Quick Start
```bash
# Initialize environment (creates data dirs, control files, .env from template)
./scripts/initialize-environment.sh

# Start full stack
./run-phoenix.sh

# Or use docker-compose directly
docker-compose up -d

# Stop services
./run-phoenix.sh stop

# Clean everything
./run-phoenix.sh clean
```

### Makefile Commands
```bash
# Main targets
make help                   # Show all available commands
make setup-env             # Initialize environment
make build                 # Build all projects (Turborepo)
make build-docker          # Build all Docker images
make dev                   # Start development mode
make test                  # Run tests
make test-integration      # Run integration tests
make monitor               # Open monitoring dashboards
make clean                 # Clean build artifacts

# Service logs
make collector-logs        # View main collector logs
make observer-logs         # View observer logs
make actuator-logs         # View control actuator logs
make generator-logs        # View generator logs

# Utilities
make validate-config       # Validate YAML configurations
make docs-serve           # Serve documentation locally
```

### Docker Compose Operations
```bash
# Start specific services
docker-compose up -d otelcol-main otelcol-observer prometheus grafana

# Rebuild and restart a specific service
docker-compose build control-actuator-go
docker-compose up -d control-actuator-go

# View logs
docker-compose logs -f otelcol-main
docker-compose logs -f control-actuator-go

# Check service health
curl http://localhost:13133  # Main collector health
curl http://localhost:13134  # Observer health
curl http://localhost:8081/health  # Control actuator health
curl http://localhost:8082/health  # Anomaly detector health
```

### Testing & Validation
```bash
# Run integration tests
./tests/integration/test_core_functionality.sh

# Generate synthetic load
docker-compose up synthetic-metrics-generator

# Run benchmark scenarios
curl http://localhost:8083/benchmark/scenarios  # List scenarios
curl -X POST http://localhost:8083/benchmark/run \
  -H "Content-Type: application/json" \
  -d '{"scenario": "baseline_steady_state"}'

# Monitor control signal changes
watch cat configs/control/optimization_mode.yaml

# Validate configurations
docker-compose config
sha256sum configs/otel/collectors/*.yaml configs/control/*template.yaml > CHECKSUMS.txt
```

### Cloud Deployment
```bash
# AWS ECS deployment
./deploy-aws.sh

# Azure Container Instances deployment  
./deploy-azure.sh

# Docker context deployment
docker context create ecs aws-phoenix --region us-west-2
docker context use aws-phoenix
docker compose up --detach
```

## Configuration Architecture

### OpenTelemetry Configurations
- `configs/otel/collectors/main.yaml`: Core collector with 3-pipeline configuration
- `configs/otel/collectors/main-optimized.yaml`: Enhanced version with shared processing
- `configs/otel/collectors/observer.yaml`: Monitoring collector that exposes KPI metrics
- `configs/otel/processors/common_intake_processors.yaml`: Shared processor configurations
- `configs/otel/exporters/newrelic-enhanced.yaml`: New Relic OTLP integration

### Control System
- `configs/control/optimization_mode.yaml`: Dynamic control file modified by actuator
- `configs/control/optimization_mode_template.yaml`: Template defining control file schema
- Version tracking with `config_version` field
- Correlation IDs for tracking changes

### Monitoring Stack
- `configs/monitoring/prometheus/prometheus.yaml`: Prometheus scrape configuration
- `configs/monitoring/prometheus/rules/phoenix_comprehensive_rules.yml`: 25+ recording rules
- `configs/monitoring/grafana/`: Datasource and dashboard provisioning
- `monitoring/grafana/dashboards/`: Phoenix dashboards (adaptive control, ultra overview)

## Key Environment Variables

Critical variables in `.env`:
```bash
# Control thresholds for adaptive switching
TARGET_OPTIMIZED_PIPELINE_TS_COUNT=20000
THRESHOLD_OPTIMIZATION_CONSERVATIVE_MAX_TS=15000
THRESHOLD_OPTIMIZATION_AGGRESSIVE_MIN_TS=25000
HYSTERESIS_FACTOR=0.1

# Resource constraints
OTELCOL_MAIN_MEMORY_LIMIT_MIB="1024"
OTELCOL_MAIN_GOMAXPROCS="2"

# Control loop timing
ADAPTIVE_CONTROLLER_INTERVAL_SECONDS=60
ADAPTIVE_CONTROLLER_STABILITY_SECONDS=120

# Load generation
SYNTHETIC_PROCESS_COUNT_PER_HOST=250
SYNTHETIC_HOST_COUNT=3
SYNTHETIC_METRIC_EMIT_INTERVAL_S=15

# New Relic export
NEW_RELIC_LICENSE_KEY=your_key_here
NEW_RELIC_OTLP_ENDPOINT=https://otlp.nr-data.net:4317
ENABLE_NR_EXPORT_FULL="false"
ENABLE_NR_EXPORT_OPTIMISED="false"
ENABLE_NR_EXPORT_EXPERIMENTAL="false"
```

## Service Endpoints & APIs

### Core Service Endpoints
- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Main Collector Metrics**: http://localhost:8888/metrics
- **Optimized Pipeline**: http://localhost:8889/metrics 
- **Experimental Pipeline**: http://localhost:8890/metrics
- **Observer Metrics**: http://localhost:9888/metrics

### API Endpoints
- **Control Actuator API**: http://localhost:8081
  - `GET /metrics` - Control state and metrics
  - `GET /health` - Health check
  - `POST /anomaly` - Webhook for anomaly events
  - `POST /mode` - Force mode change (testing)
- **Anomaly Detector API**: http://localhost:8082
  - `GET /alerts` - Active anomalies
  - `GET /health` - Health check
  - `GET /metrics` - Prometheus metrics
- **Benchmark Controller**: http://localhost:8083
  - `GET /benchmark/scenarios` - List test scenarios
  - `POST /benchmark/run` - Run benchmark
  - `GET /benchmark/results` - Get results
  - `GET /benchmark/validate` - Check SLO compliance

### Health Checks
- Main Collector: http://localhost:13133
- Observer Collector: http://localhost:13134
- Docker health checks configured with 20s intervals, 3 retries

### Debug Endpoints
- pprof: http://localhost:1777/debug/pprof
- zpages: http://localhost:55679

## Control Flow & Data Paths

1. **Metrics Ingestion**: Synthetic generator → Main collector OTLP endpoint (4318)
2. **Pipeline Processing**: 3 parallel processing chains with different optimization levels
3. **Metrics Export**: Each pipeline exports to dedicated Prometheus endpoints (8888-8890)
4. **Monitoring**: Observer scrapes main collector metrics and exposes KPIs (9888)
5. **Control Loop**: Actuator queries observer metrics → calculates profile → updates control file
6. **Adaptation**: Main collector reads control file changes → adjusts pipeline behavior
7. **Anomaly Detection**: Detector monitors metrics → sends webhooks to control actuator
8. **Benchmarking**: Controller generates load patterns → validates performance

## Development Patterns

### Adding New Processors
1. Create processor config in `configs/otel/processors/`
2. Include in pipeline via `configs/otel/collectors/main.yaml`
3. Test with synthetic load and monitor cardinality impact

### Modifying Control Logic
1. Update thresholds in `.env` file
2. Modify PID logic in `apps/control-actuator-go/main.go`
3. Test profile transitions using benchmark scenarios
4. Monitor stability score via recording rules

### Performance Tuning
```bash
# Enable debug logging
export OTEL_LOG_LEVEL=debug
docker-compose up -d otelcol-main

# Profile memory usage
curl http://localhost:1777/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Monitor resource usage
docker-compose top

# Check pipeline efficiency
curl -s http://localhost:9090/api/v1/query?query=phoenix:resource_efficiency_score
```

### Local Development
```bash
# Run Go service locally
cd apps/control-actuator-go
go run main.go

# With live reload
air

# Run tests
go test -v -race ./...

# Build binary
go build -o control-actuator
```

## Monorepo Structure

- **`apps/`**: Go-based microservices (control-actuator, anomaly-detector)
- **`services/`**: Service implementations with Dockerfiles
- **`configs/`**: Technology-grouped configurations (otel, monitoring, control)
- **`infrastructure/`**: Cloud deployment configuration (Docker contexts)
- **`packages/`**: Shared packages (managed by npm workspaces)
- **`scripts/`**: Operational utilities and environment setup
- **`tests/`**: Integration and performance tests
- **`tools/`**: Development and migration utilities
- **`data/`**: Persistent storage directories (gitignored)

## Build System

- **Turborepo**: Parallel builds with caching (`turbo.json`)
- **Make**: Developer-friendly commands (see `make help`)
- **npm workspaces**: Package management
- **Docker multi-stage builds**: Optimized images
- **Go modules**: Dependency management for Go services

## Key Metrics & Alerts

### Recording Rules (25+ rules)
- **Efficiency**: `phoenix:signal_preservation_score`, `phoenix:cardinality_efficiency_ratio`
- **Performance**: `phoenix:pipeline_latency_ms_p99`, `phoenix:pipeline_throughput_metrics_per_sec`
- **Control**: `phoenix:control_stability_score`, `phoenix:control_loop_effectiveness`
- **Anomaly**: `phoenix:cardinality_zscore`, `phoenix:cardinality_explosion_risk`
- **Resource**: `phoenix:resource_efficiency_score`, `phoenix:collector_memory_usage_mb`

### Critical Alerts
- `PhoenixCardinalityExplosion`: Exponential growth detected
- `PhoenixResourceExhaustion`: Memory >90%
- `PhoenixControlLoopInstability`: Frequent mode changes
- `PhoenixSLOViolation`: Service objectives not met

## Benchmark Scenarios

1. **baseline_steady_state**: Normal operation validation
2. **cardinality_spike**: Sudden 3x increase testing
3. **gradual_growth**: Linear growth over time
4. **wave_pattern**: Sinusoidal load pattern

## CI/CD Integration

### GitHub Actions Workflows
- `.github/workflows/ci.yml`: Full CI/CD pipeline
- `.github/workflows/security.yml`: Security scanning (Trivy, Gosec, OWASP)

### Pipeline Stages
1. Configuration validation
2. Go service testing with coverage
3. Integration testing
4. Docker image building
5. Performance benchmarking
6. Deployment (on main branch)

## Troubleshooting

### Common Issues
1. **High memory usage**: Increase `OTELCOL_MAIN_MEMORY_LIMIT_MIB`
2. **Control instability**: Increase `ADAPTIVE_CONTROLLER_STABILITY_SECONDS`
3. **Poor reduction**: Check mode via :8081/metrics, adjust thresholds
4. **Anomaly noise**: Modify detector threshold in `apps/anomaly-detector/main.go`

### Debug Commands
```bash
# Check control decisions
curl http://localhost:8081/metrics | jq '.current_mode'

# View pipeline cardinality
curl -s http://localhost:9090/api/v1/query?query=phoenix_observer_kpi_store_phoenix_pipeline_output_cardinality_estimate

# Force mode change (testing)
curl -X POST http://localhost:8081/mode \
  -H "Content-Type: application/json" \
  -d '{"mode": "aggressive"}'

# Export metrics for analysis
curl "http://localhost:9090/api/v1/query_range?query=phoenix:cardinality_growth_rate&start=$(date -u -d '1 hour ago' +%s)&end=$(date +%s)&step=60"

# Check for memory leaks
docker stats --no-stream
```
