# Phoenix Platform Architecture - Complete Documentation

## 📐 Architecture Overview

The Phoenix Platform is a cloud-native, microservices-based observability cost optimization system built as a monorepo with completely independent micro-projects. This architecture provides maximum flexibility while maintaining consistency and reducing operational overhead.

## 🎯 Core Architecture Principles

### 1. **Project Independence**
- Each project in `/projects` is completely independent
- No cross-project imports allowed (enforced by tooling)
- Projects can only import from `/pkg` shared packages
- Each project maintains its own lifecycle and can be deployed independently

### 2. **Shared Infrastructure**
- Common code in `/pkg` reduces duplication by ~70%
- Unified build system via shared Makefiles
- Consistent tooling across all projects
- Single CI/CD pipeline with smart detection

### 3. **Boundary Enforcement**
- Automated tools prevent architectural drift
- Pre-commit hooks validate boundaries
- CI/CD enforces rules on every change
- Clear separation of concerns

### 4. **Scalability & Performance**
- Horizontal scaling for all services
- Efficient resource utilization
- Optimized build times
- Smart caching strategies

## 🏗️ System Architecture

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                        Phoenix Platform                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │   Web UI    │  │  Phoenix    │  │   Mobile    │           │
│  │ (Dashboard) │  │    CLI      │  │    App      │           │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘           │
│         │                 │                 │                   │
│         └─────────────────┴─────────────────┘                  │
│                           │                                     │
│                    ┌──────▼──────┐                            │
│                    │ API Gateway │                            │
│                    │(platform-api)│                            │
│                    └──────┬──────┘                            │
│                           │                                     │
│     ┌─────────────────────┼─────────────────────┐             │
│     │                     │                     │              │
│ ┌───▼────┐  ┌────────┐  ┌▼──────────┐  ┌─────▼─────┐       │
│ │Controller│  │Generator│  │ Analytics │  │ Anomaly   │       │
│ │ Service │  │ Service │  │  Engine   │  │ Detector  │       │
│ └────┬────┘  └────┬───┘  └─────┬─────┘  └─────┬─────┘       │
│      │            │             │               │              │
│      └────────────┴─────────────┴───────────────┘             │
│                           │                                     │
│                    ┌──────▼──────┐                            │
│                    │  Data Layer │                            │
│                    │ (PostgreSQL,│                            │
│                    │   Redis)    │                            │
│                    └─────────────┘                            │
│                                                                 │
│  Kubernetes Operators                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  Pipeline    │  │   LoadSim    │  │  Experiment  │       │
│  │  Operator    │  │   Operator   │  │  Controller  │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Component Architecture

#### 1. **API Gateway (platform-api)**
- Central entry point for all external requests
- Authentication and authorization
- Request routing and load balancing
- Rate limiting and throttling
- API versioning support

#### 2. **Controller Service**
- Orchestrates experiment lifecycle
- Manages state transitions
- Coordinates with Kubernetes operators
- Handles scheduling and resource allocation

#### 3. **Generator Service**
- Creates optimized pipeline configurations
- Implements various optimization strategies
- Validates pipeline specifications
- Provides templates and presets

#### 4. **Analytics Engine**
- Real-time metrics analysis
- Cost calculation and projections
- Performance metrics aggregation
- Historical data analysis

#### 5. **Anomaly Detector**
- Identifies unusual patterns
- ML-based anomaly detection
- Alerts on significant deviations
- Predictive analysis

#### 6. **Web Dashboard**
- React-based single-page application
- Real-time WebSocket updates
- Interactive pipeline builder
- Comprehensive metrics visualization

#### 7. **Phoenix CLI**
- Command-line interface for all operations
- Plugin architecture for extensibility
- Scripting support
- CI/CD integration

## 📦 Repository Structure

### Directory Layout
```
phoenix/
├── .github/                     # GitHub Actions workflows
├── build/                       # Shared build infrastructure
│   ├── docker/                  # Base Docker images
│   ├── makefiles/              # Shared Makefiles
│   └── scripts/                # Build scripts
├── configs/                     # Configuration files
│   ├── control/                # Control plane configs
│   ├── monitoring/             # Prometheus/Grafana
│   ├── otel/                   # OpenTelemetry
│   └── production/             # Production configs
├── deployments/                 # Deployment manifests
│   ├── kubernetes/             # K8s manifests
│   ├── helm/                   # Helm charts
│   └── terraform/              # Infrastructure as code
├── docs/                        # Documentation
│   ├── architecture/           # Architecture docs
│   ├── migration/              # Migration history
│   └── operations/             # Operational guides
├── pkg/                         # Shared Go packages
│   ├── auth/                   # Authentication
│   ├── telemetry/              # Observability
│   ├── database/               # DB abstractions
│   ├── http/                   # HTTP utilities
│   ├── grpc/                   # gRPC utilities
│   └── errors/                 # Error handling
├── projects/                    # Independent services
│   ├── platform-api/           # API Gateway
│   ├── controller/             # Controller service
│   ├── generator/              # Generator service
│   ├── analytics/              # Analytics engine
│   ├── anomaly-detector/       # Anomaly detection
│   ├── dashboard/              # Web UI
│   ├── phoenix-cli/            # CLI tool
│   ├── pipeline-operator/      # K8s operator
│   └── loadsim-operator/       # Load testing
├── scripts/                     # Utility scripts
├── tests/                       # Cross-project tests
│   ├── integration/            # Integration tests
│   ├── e2e/                    # End-to-end tests
│   └── performance/            # Performance tests
├── tools/                       # Development tools
│   ├── analyzers/              # Code analyzers
│   └── generators/             # Code generators
├── go.work                      # Go workspace
├── Makefile                     # Root Makefile
└── docker-compose.yml          # Development stack
```

### Project Structure Standard
Each project follows this structure:
```
projects/<project-name>/
├── cmd/                         # Application entrypoints
├── internal/                    # Private application code
│   ├── api/                    # API handlers
│   ├── domain/                 # Business logic
│   ├── infrastructure/         # External dependencies
│   └── config/                 # Configuration
├── api/                         # API definitions
├── build/                       # Build configurations
├── deployments/                # Deployment configs
├── migrations/                 # Database migrations
├── tests/                      # Project tests
├── Makefile                    # Project Makefile
├── go.mod                      # Go module
└── README.md                   # Documentation
```

## 🔧 Technology Stack

### Core Technologies
- **Language**: Go 1.21+ (backend), TypeScript/React (frontend)
- **Container**: Docker, containerd
- **Orchestration**: Kubernetes 1.28+
- **Database**: PostgreSQL 15+, Redis 7+
- **Messaging**: NATS, Kafka (optional)
- **Observability**: OpenTelemetry, Prometheus, Grafana

### Development Tools
- **Build**: Make, Docker Buildx
- **CI/CD**: GitHub Actions
- **Testing**: Go testing, Jest, Playwright
- **Linting**: golangci-lint, ESLint
- **Security**: Trivy, gosec, OWASP

## 🔐 Security Architecture

### Security Layers
1. **Network Security**
   - mTLS between services
   - Network policies in Kubernetes
   - WAF for external traffic

2. **Application Security**
   - JWT-based authentication
   - RBAC authorization
   - Input validation
   - Output encoding

3. **Data Security**
   - Encryption at rest
   - Encryption in transit
   - Key rotation
   - Secrets management

4. **Supply Chain Security**
   - Dependency scanning
   - Container scanning
   - SBOM generation
   - Signed images

## 📊 Data Architecture

### Data Flow
```
Metrics Source → Collector → Processor → Storage → Analytics
                    ↓           ↓          ↓         ↓
                 Validation  Optimization  Query   Visualization
```

### Storage Strategy
- **PostgreSQL**: Transactional data, configurations
- **Redis**: Caching, session storage, real-time data
- **Object Storage**: Long-term metrics archival
- **Time-series DB**: Optional for high-volume metrics

## 🚀 Deployment Architecture

### Deployment Strategies
1. **Blue-Green Deployment**
   - Zero-downtime deployments
   - Quick rollback capability
   - A/B testing support

2. **Canary Releases**
   - Gradual rollout
   - Automated rollback on errors
   - Metrics-based promotion

3. **Feature Flags**
   - Runtime configuration
   - Gradual feature rollout
   - Quick disable capability

### Environments
- **Development**: Local Docker Compose
- **Staging**: Kubernetes cluster (scaled down)
- **Production**: Multi-region Kubernetes
- **DR**: Hot standby in alternate region

## 📈 Performance Architecture

### Optimization Strategies
1. **Caching**
   - Multi-level caching (Redis, in-memory)
   - Cache invalidation strategies
   - Edge caching for static assets

2. **Async Processing**
   - Message queues for heavy operations
   - Background job processing
   - Event-driven architecture

3. **Resource Optimization**
   - Connection pooling
   - Resource limits and requests
   - Horizontal pod autoscaling

### Performance Targets
- API Response: < 100ms (p95)
- Pipeline Generation: < 5s
- Dashboard Load: < 2s
- Metrics Processing: 1M metrics/second

## 🔄 Integration Architecture

### External Integrations
- **Cloud Providers**: AWS, GCP, Azure
- **Observability**: Datadog, New Relic, Splunk
- **CI/CD**: Jenkins, GitLab, CircleCI
- **Communication**: Slack, PagerDuty, Teams

### Integration Patterns
1. **REST APIs**: Primary integration method
2. **gRPC**: High-performance service communication
3. **WebSockets**: Real-time updates
4. **Webhooks**: Event notifications

## 🎛️ Operational Architecture

### Monitoring Stack
```
Application → OpenTelemetry Collector → Prometheus → Grafana
                    ↓                        ↓          ↓
                  Jaeger                AlertManager  Dashboards
```

### Key Metrics
- **Golden Signals**: Latency, Traffic, Errors, Saturation
- **Business Metrics**: Cost reduction, Data accuracy
- **Infrastructure**: CPU, Memory, Disk, Network

### Alerting Strategy
- **Critical**: Page on-call (< 5 min response)
- **High**: Notify team (< 30 min response)
- **Medium**: Create ticket (< 4 hour response)
- **Low**: Dashboard only

## 🔮 Future Architecture

### Planned Enhancements
1. **Multi-Cloud Support**
   - Provider-agnostic design
   - Cross-cloud data replication
   - Cloud-specific optimizations

2. **AI/ML Integration**
   - Advanced anomaly detection
   - Predictive optimization
   - Automated remediation

3. **Edge Computing**
   - Edge collectors
   - Local processing
   - Reduced latency

4. **GraphQL API**
   - Flexible queries
   - Real-time subscriptions
   - Better mobile support

## 📚 Architecture References

### Design Patterns Used
- **Microservices**: Service independence
- **Event Sourcing**: Audit trail
- **CQRS**: Read/write separation
- **Circuit Breaker**: Fault tolerance
- **Saga**: Distributed transactions

### Architecture Decision Records
- [ADR-001: Monorepo Structure](./decisions/ADR-001-monorepo.md)
- [ADR-002: Service Communication](./decisions/ADR-002-communication.md)
- [ADR-003: Deployment Strategy](./decisions/ADR-003-deployment.md)
- [ADR-004: Security Model](./decisions/ADR-004-security.md)

---

*This document provides a complete view of the Phoenix Platform architecture.*  
*For specific component details, refer to individual project documentation.*  
*Last Updated: May 2025*