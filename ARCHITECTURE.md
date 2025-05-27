# Phoenix Platform Architecture

Phoenix is an observability cost optimization platform that reduces metrics cardinality by up to 70% while maintaining critical visibility through agent-based task distribution and A/B testing of optimization pipelines.

## Overview

Phoenix uses an **agent-based architecture** with three main components:

1. **Phoenix API** - Centralized control plane
2. **Phoenix Agent** - Lightweight data plane agents  
3. **Dashboard** - React-based web UI

```
┌─────────────────┐         ┌─────────────────┐
│   Phoenix API   │◄────────┤   Dashboard     │
│  (Port 8080)    │         │   (React UI)    │
│  + WebSocket    │         └─────────────────┘
└────────┬────────┘
         │ Task Queue (PostgreSQL)
         │ Long-polling (30s timeout)
    ┌────▼────┐
    │ Phoenix │────► OpenTelemetry ────► Backends
    │ Agents  │      Collector
    └─────────┘
```

## Core Components

### Phoenix API
- **Role**: Central control plane
- **Responsibilities**:
  - Experiment management
  - Task distribution via PostgreSQL queue
  - Metrics analysis and KPI calculation
  - WebSocket support for real-time updates
- **Technology**: Go, PostgreSQL, Redis (optional)

### Phoenix Agent  
- **Role**: Distributed data plane
- **Responsibilities**:
  - Poll API for tasks using X-Agent-Host-ID authentication
  - Manage OpenTelemetry collectors with pipeline templates
  - Execute A/B tests with baseline/candidate configurations
  - Report metrics and status back to API
- **Technology**: Go, minimal dependencies (<50MB RAM)

### Dashboard
- **Role**: User interface
- **Responsibilities**:
  - Experiment creation with A/B testing setup
  - Real-time monitoring via WebSocket
  - Pipeline templates (Adaptive Filter, TopK, Hybrid)
  - Live cost reduction analytics
- **Technology**: React 18, TypeScript, Vite, WebSocket

## Key Design Principles

1. **Simplicity**: Minimal components, clear responsibilities
2. **Scalability**: Handles 10,000+ agents, horizontal scaling
3. **Reliability**: PostgreSQL for state, automatic retries
4. **Security**: JWT auth, no inbound agent connections
5. **Performance**: Sub-second decisions, minimal overhead

## Data Flow

### Experiment Lifecycle
```
User → Dashboard → API → PostgreSQL → Task Queue
                                          ↓
                              Agent ← Poll Tasks (30s)
                                ↓
                          OTel Collector (Baseline/Candidate)
                                ↓
                    Metrics Backend → API (Analysis) → WebSocket → Dashboard
```

### Communication Patterns
- **Dashboard ↔ API**: REST (port 8080) + WebSocket (same port)
- **Agent → API**: HTTP long-polling with X-Agent-Host-ID header
- **Agent → Backends**: OpenTelemetry protocol (OTLP)
- **Task Queue**: PostgreSQL-based with atomic assignment

## Deployment Architecture

### Development
- Docker Compose for all services
- Hot reload for rapid development
- Single-command startup: `make dev-up`
- Integrated monitoring stack

### Production (Single-VM)
- Docker Compose orchestration
- All services containerized except agents
- Agents deployed via systemd on host machines
- Auto-scaling scripts for resource management
- Built-in backup and restore capabilities

### Key Deployment Features
- **No Kubernetes Required**: Simplified operations with Docker Compose
- **Resource Efficient**: Runs on single VM (4 vCPU, 16GB RAM minimum)
- **High Availability**: External PostgreSQL + multiple API replicas
- **Monitoring**: Integrated Prometheus + Grafana stack
- **TLS Support**: Let's Encrypt or self-signed certificates

## Security Model

- **Authentication**: JWT tokens for users, X-Agent-Host-ID for agents
- **Authorization**: Role-based access control (RBAC)
- **Agent Security**: Outbound-only connections, task polling design
- **Network Isolation**: Docker networks with explicit service exposure
- **Secrets Management**: Environment files with proper permissions
- **TLS Everywhere**: HTTPS for API, encrypted database connections

## Migration Path

For users migrating from previous Kubernetes deployments:
- See [MIGRATION_FROM_KUBERNETES.md](MIGRATION_FROM_KUBERNETES.md)
- Data migration scripts provided
- Zero-downtime migration possible
- **Data Protection**: TLS encryption, PostgreSQL row-level security

## Performance Characteristics

| Component | Metric | Target |
|-----------|--------|--------|
| API | Concurrent agents | 10,000+ |
| API | Task latency | <100ms |
| Agent | Memory usage | <50MB |
| Agent | CPU usage | <1% |
| Agent | Collectors per agent | 100+ |

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+ (primary datastore)
- **Task Queue**: PostgreSQL-based with long-polling
- **Metrics**: OpenTelemetry Collector + various backends

### Frontend
- **Framework**: React 18+ with Vite
- **Language**: TypeScript
- **State**: Redux Toolkit + Zustand
- **Real-time**: WebSocket for live updates

### Infrastructure
- **Container**: Docker
- **Orchestration**: Docker Compose + systemd
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana

## Extension Points

1. **Pipeline Templates**: Adaptive Filter, TopK, Hybrid processors
2. **Analysis Metrics**: Cardinality reduction, cost savings calculations
3. **UI Components**: Live monitoring, cost flow visualization
4. **API Extensions**: RESTful API v2 with WebSocket support

## Related Documentation

- [Platform Architecture Details](docs/architecture/PLATFORM_ARCHITECTURE.md)
- [API Documentation](docs/api/)
- [Operations Guide](docs/operations/OPERATIONS_GUIDE_COMPLETE.md)
- [Development Guide](DEVELOPMENT_GUIDE.md)