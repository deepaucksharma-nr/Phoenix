# Phoenix Platform - Team Assignments & Code Ownership

## 👥 Team Structure


- **Palash** - Principal Engineer, Platform Architecture
- **Abhinav** - Senior Engineer, Infrastructure & DevOps
- **Srikanth** - Senior Engineer, Core Services

---

## 📂 Code Ownership Assignments

### 🏗️ Palash

**Primary Ownership**:
```
├── /pkg/                           # Shared packages architecture
│   ├── auth/                       # Authentication framework
│   ├── contracts/                  # Service contracts
│   ├── errors/                     # Error handling patterns
│   └── interfaces/                 # Core interfaces
├── /docs/architecture/             # Architecture documentation
├── MONOREPO_BOUNDARIES.md          # Architectural boundaries
├── PHOENIX_PLATFORM_ARCHITECTURE.md # Platform architecture
└── go.work                         # Go workspace configuration
```

**Responsibilities**:
- Platform architecture decisions
- Code review for architectural changes
- Shared package design and maintenance
- Monorepo structure governance
- Technical mentorship

**Backup**: Michael Kumar

---

### 🔧 Abhinav
**Primary Ownership**:
```
├── /deployments/                   # All deployment configurations
│   ├── kubernetes/                 # K8s manifests
│   ├── helm/                       # Helm charts
│   └── terraform/                  # Infrastructure as code
├── /infrastructure/                # Infrastructure components
├── /.github/workflows/             # CI/CD pipelines
├── /scripts/                       # Automation scripts
├── /tools/                         # Development tools
│   ├── analyzers/                  # Code analyzers
│   └── dev-env/                    # Development environment
└── docker-compose*.yml             # Docker configurations
```

**Responsibilities**:
- CI/CD pipeline management
- Kubernetes deployments
- Infrastructure automation
- Security and monitoring setup
- Production deployments


---

### 💼 Srikanth
**Primary Ownership**:
```
├── /projects/platform-api/         # Main API gateway
├── /projects/controller/           # Experiment controller
├── /pkg/grpc/                      # gRPC infrastructure
├── /pkg/http/                      # HTTP utilities
├── /pkg/database/                  # Database abstractions
└── /tests/integration/             # Integration testing
```

**Responsibilities**:
- Core service architecture
- API design and standards
- Database architecture
- Integration testing strategy
- Performance optimization

---

### 🌐 Praveen

**Primary Ownership**:
```
├── /projects/dashboard/            # React dashboard
│   ├── src/components/             # UI components
│   ├── src/pages/                  # Page components
│   └── src/services/               # API services
├── /projects/phoenix-cli/          # CLI tool
└── /docs/guides/user/              # User documentation
```

**Responsibilities**:
- Dashboard development
- CLI tool maintenance
- User experience improvements
- Frontend-backend integration
- User documentation
---

### ⚙️ Shivani

**Primary Ownership**:
```
├── /projects/analytics/            # Analytics service
├── /projects/anomaly-detector/     # Anomaly detection
├── /pkg/telemetry/                 # Telemetry packages
│   ├── metrics/                    # Metrics collection
│   └── tracing/                    # Distributed tracing
└── /tests/performance/             # Performance tests
```

**Responsibilities**:
- Analytics pipeline development
- Anomaly detection algorithms
- Metrics and monitoring
- Performance testing
- Data processing optimization

---

### 🎨 Jyothi

**Primary Ownership**:
```
├── /projects/dashboard/
│   ├── src/components/            # Shared components
│   ├── src/hooks/                 # React hooks
│   ├── src/store/                 # State management
│   └── src/theme/                 # UI theming
├── /docs/guides/developer/        # Developer guides
└── /tests/e2e/                    # E2E tests
```

**Responsibilities**:
- Frontend architecture
- Component library maintenance
- UI/UX implementation
- Frontend testing
- Developer documentation


---

### 🚀 Anitha

**Primary Ownership**:
```
├── /projects/pipeline-operator/    # K8s operator
├── /projects/loadsim-operator/     # Load testing operator
├── /operators/                     # Operator implementations
├── /configs/                       # Configuration management
│   ├── monitoring/                 # Monitoring configs
│   └── production/                 # Production configs
└── /monitoring/                    # Monitoring setup
```

**Responsibilities**:
- Kubernetes operators
- Platform automation
- Monitoring and alerting
- Configuration management
- Operational tooling


---

### 🔌 Tharun

**Primary Ownership**:
```
├── /projects/benchmark/            # Benchmarking service
├── /projects/validator/            # Validation service
├── /pkg/utils/                     # Utility packages
└── /pkg/testing/                   # Testing utilities
```

**Responsibilities**:
- Benchmark implementation
- Validation logic
- Utility functions
- Test helpers
- Bug fixes


---

### 🖼️ Tanush

**Primary Ownership**:
```
├── /projects/dashboard/
│   ├── src/components/common/      # Common components
│   └── src/utils/                  # Frontend utilities
├── /site/                          # Documentation site
└── /docs/api/                      # API documentation
```

**Responsibilities**:
- Component implementation
- Documentation site maintenance
- UI bug fixes
- Frontend utilities
- API documentation
---

### 🔒 Ramana

**Primary Ownership**:
```
├── /scripts/migration/             # Migration scripts
├── /scripts/validation/            # Validation scripts
├── /configs/otel/                  # OpenTelemetry configs
├── /docs/operations/               # Operations docs
└── /docs/runbooks/                 # Operational runbooks
```

**Responsibilities**:
- Script maintenance
- Configuration updates
- Runbook documentation
- Monitoring setup assistance
- DevOps tooling support

---



