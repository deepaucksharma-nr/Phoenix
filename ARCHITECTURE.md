# Phoenix Platform Architecture

Phoenix is an observability cost optimization platform that reduces metrics cardinality by up to 90% while maintaining critical visibility.

## Overview

Phoenix uses a **lean-core architecture** with three main components:

1. **Phoenix API** - Centralized control plane
2. **Phoenix Agent** - Lightweight data plane agents  
3. **Dashboard** - React-based web UI

```
┌─────────────────┐         ┌─────────────────┐
│   Phoenix API   │◄────────┤   Dashboard     │
│  (Control Plane)│         │   (React UI)    │
└────────┬────────┘         └─────────────────┘
         │ Task Queue (PostgreSQL)
    ┌────▼────┐
    │ Phoenix │────► Pushgateway ────► Prometheus
    │ Agents  │
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
  - Poll API for work assignments (long-polling)
  - Manage OpenTelemetry collectors
  - Push metrics to Prometheus Pushgateway
  - Self-register with zero configuration
- **Technology**: Go, minimal dependencies (<50MB RAM)

### Dashboard
- **Role**: User interface
- **Responsibilities**:
  - Experiment creation and monitoring
  - Real-time metrics visualization
  - Pipeline catalog management
  - Cost analytics display
- **Technology**: React, TypeScript, WebSocket

## Key Design Principles

1. **Simplicity**: Minimal components, clear responsibilities
2. **Scalability**: Handles 10,000+ agents, horizontal scaling
3. **Reliability**: PostgreSQL for state, automatic retries
4. **Security**: JWT auth, no inbound agent connections
5. **Performance**: Sub-second decisions, minimal overhead

## Data Flow

### Experiment Lifecycle
```
User → Dashboard → API → Database → Task Queue
                                         ↓
                              Agent ← Poll Tasks
                                ↓
                          OTel Collector
                                ↓
                    Pushgateway → Prometheus → API (Analysis)
```

### Communication Patterns
- **Dashboard ↔ API**: REST + WebSocket
- **Agent → API**: HTTP long-polling (outbound only)
- **Agent → Pushgateway**: Prometheus metrics push
- **API → Prometheus**: PromQL queries for analysis

## Deployment Architecture

### Development
- Docker Compose for all services
- Hot reload for rapid development
- Single-command startup

### Production
- Kubernetes deployment
- Horizontal scaling for API
- DaemonSet for agents
- Managed PostgreSQL recommended

## Security Model

- **Authentication**: JWT tokens
- **Authorization**: Role-based access control
- **Agent Security**: No incoming connections, API key auth
- **Data Protection**: TLS encryption, secrets management

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
- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+
- **Cache**: Redis (optional)
- **Metrics**: Prometheus + Pushgateway

### Frontend
- **Framework**: React 18+
- **Language**: TypeScript
- **State**: Redux Toolkit
- **Real-time**: WebSocket

### Infrastructure
- **Container**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana

## Extension Points

1. **Custom Processors**: Add OTel processors in collector
2. **Analysis Plugins**: Extend KPI calculations
3. **UI Plugins**: Add dashboard widgets
4. **API Extensions**: RESTful API is versioned

## Related Documentation

- [Platform Architecture Details](docs/architecture/PLATFORM_ARCHITECTURE.md)
- [API Documentation](docs/api/)
- [Operations Guide](docs/operations/OPERATIONS_GUIDE_COMPLETE.md)
- [Development Guide](DEVELOPMENT_GUIDE.md)