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

**Backup**: Ryan Thompson

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

**Backup**: David Park

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

**Backup**: Jessica Zhang

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

**Backup**: Michael Kumar

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

**Backup**: Emma Wilson

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

**Backup**: Sarah Martinez

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

**Mentor**: David Park

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

**Mentor**: Jessica Zhang

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

**Mentor**: Sarah Martinez

---

## 🔄 Rotation Schedule

### Quarterly Rotations
- **Q1**: Focus on current assignments
- **Q2**: Junior devs rotate to new areas
- **Q3**: Cross-training between teams
- **Q4**: Leadership shadowing for mid-level devs

### Knowledge Sharing
- **Weekly**: Team tech talks
- **Bi-weekly**: Code review sessions
- **Monthly**: Architecture reviews
- **Quarterly**: Hack days

---

## 📋 Team Responsibilities Matrix

| Area | Primary | Secondary | Reviewers |
|------|---------|-----------|-----------|
| Architecture | Alex | Michael | Sarah |
| Infrastructure | Sarah | Ryan | Alex |
| Core Services | Michael | David | Alex |
| Frontend | Jessica | Emma | Nathan |
| Analytics | David | Olivia | Michael |
| DevOps | Sarah | Ryan, Sophia | Alex |
| Testing | Michael | David | All |
| Documentation | Emma | Nathan | Jessica |

---

## 🎯 Sprint Assignments

### Current Sprint Focus
| Developer | Sprint Tasks | Story Points |
|-----------|-------------|--------------|
| Alex | Architecture review, pkg refactoring | 13 |
| Sarah | K8s migration, CI/CD updates | 13 |
| Michael | API v2 design, controller updates | 13 |
| Emma | Dashboard features, CLI improvements | 8 |
| David | Analytics pipeline optimization | 8 |
| Jessica | UI component library | 8 |
| Ryan | Operator improvements | 8 |
| Olivia | Benchmark suite expansion | 5 |
| Nathan | Dashboard bug fixes | 5 |
| Sophia | Monitoring setup | 5 |

---

## 🚨 On-Call Rotation

### Primary On-Call (Weekly Rotation)
1. Michael Kumar
2. David Park
3. Emma Wilson
4. Ryan Thompson

### Secondary On-Call
1. Sarah Martinez (Infrastructure)
2. Alex Chen (Architecture)

### Weekend Coverage
- Rotates among all team members
- Junior devs paired with seniors

---

## 📚 Code Review Requirements

### Review Matrix
| Code Area | Required Reviewers | Optional Reviewers |
|-----------|-------------------|-------------------|
| /pkg/* | Alex + 1 Senior | Any |
| /deployments/* | Sarah + 1 | Ryan |
| Core Services | Michael + 1 | David |
| Frontend | Jessica + 1 | Emma |
| Scripts/Tools | Sarah | Sophia |

### Review SLAs
- **Critical**: 2 hours
- **High**: 4 hours
- **Normal**: 24 hours
- **Low**: 48 hours

---

## 🎓 Mentorship Pairs

| Mentor | Mentee | Focus Area |
|--------|--------|------------|
| David Park | Olivia Brown | Backend development, Go best practices |
| Jessica Zhang | Nathan Lee | Frontend architecture, React patterns |
| Sarah Martinez | Sophia Patel | DevOps practices, Kubernetes |
| Michael Kumar | Ryan Thompson | System design, scalability |
| Alex Chen | Emma Wilson | Architecture patterns, leadership |

---

## 📊 Performance Metrics

### Individual KPIs
- **Code Quality**: Maintainability index > 80
- **Test Coverage**: > 80% for owned code
- **PR Turnaround**: < 24 hours
- **Documentation**: Updated with code changes
- **Knowledge Sharing**: 1 presentation/quarter

### Team KPIs
- **Sprint Velocity**: 80 points/sprint
- **Bug Resolution**: < 48 hours
- **Deployment Success**: > 95%
- **System Uptime**: > 99.9%

---

## 🔐 Access Levels

### Production Access
- **Full Access**: Alex, Sarah, Michael
- **Read + Deploy**: Emma, David, Ryan
- **Read Only**: Jessica, Olivia, Nathan, Sophia

### Repository Permissions
- **Admin**: Alex, Sarah
- **Maintain**: Michael, Emma, David
- **Write**: All team members
- **Triage**: External contributors

---

## 📅 Meeting Schedule

### Daily
- **Standup**: 9:30 AM (15 min)
- **Blockers**: 4:30 PM (optional)

### Weekly
- **Architecture Review**: Monday 2 PM (Alex leads)
- **Sprint Planning**: Tuesday 10 AM (All)
- **Tech Talk**: Thursday 3 PM (Rotating)
- **Retrospective**: Friday 2 PM (Bi-weekly)

### Monthly
- **All-Hands**: First Monday (All)
- **1-on-1s**: Throughout month
- **Hack Day**: Last Friday

---

## 📈 Career Development Paths

### Growth Tracks
1. **IC Track**: Junior → Mid → Senior → Staff → Principal
2. **Management Track**: Senior → Tech Lead → Engineering Manager
3. **Architecture Track**: Senior → Solution Architect → Principal Architect

### Skill Development Focus
- **Juniors**: Core skills, code quality, testing
- **Mid-Level**: System design, ownership, mentoring
- **Seniors**: Architecture, strategy, leadership

---

*Last Updated: May 2025*  
*Team Lead: Alex Chen*  
*Department: Platform Engineering*
