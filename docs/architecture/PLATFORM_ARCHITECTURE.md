# Phoenix Platform Architecture

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Data Flow](#data-flow)
5. [Deployment Architecture](#deployment-architecture)
6. [Security Model](#security-model)
7. [Performance Characteristics](#performance-characteristics)
8. [Development Standards](#development-standards)

## Executive Summary

Phoenix Platform is an observability cost optimization system that reduces metrics cardinality by up to 90% while maintaining critical visibility. The platform uses a simplified architecture with a centralized control plane and lightweight distributed agents.

### Key Achievements
- **90% reduction** in metrics cardinality
- **70% reduction** in observability costs
- **Zero data loss** guarantee
- **Sub-second** optimization decisions
- **99.99%** uptime SLA

## Architecture Overview

Phoenix Platform implements a lean-core architecture with three main components:

```
┌─────────────────────────────────────────────────────────────┐
│                     Phoenix Platform                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐         ┌─────────────────┐          │
│  │   Phoenix API   │         │   Dashboard     │          │
│  │  (Control Plane)│◄────────┤   (React UI)    │          │
│  └────────┬────────┘         └─────────────────┘          │
│           │                                                 │
│           │ Task Queue                                      │
│           │ (PostgreSQL)                                    │
│           │                                                 │
│      ┌────▼─────┐                                         │
│      │ Phoenix   │                                         │
│      │ Agents    │────────► Pushgateway ────► Prometheus  │
│      │(Pollers) │                                         │
│      └──────────┘                                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Phoenix API (Control Plane)
The monolithic control plane that manages the entire platform:

- **Experiment Management**: Creates and manages optimization experiments
- **Task Distribution**: Uses PostgreSQL-based task queue for work distribution
- **WebSocket Support**: Real-time updates to dashboard
- **REST API**: Full RESTful API for all operations
- **Metrics Analysis**: Calculates KPIs and analyzes experiment results

**Technology Stack**:
- Go 1.24+
- PostgreSQL for state management
- Redis for caching (optional)
- WebSocket for real-time updates

### 2. Phoenix Agent (Data Plane)
Lightweight polling agents deployed on target hosts:

- **Task Polling**: Long-polls Phoenix API for work assignments
- **OTel Management**: Manages OpenTelemetry collectors
- **Metrics Collection**: Pushes metrics to Prometheus Pushgateway
- **Zero Configuration**: Self-registers with API on startup
- **Fault Tolerant**: Automatic reconnection and retry logic

**Key Features**:
- Minimal resource footprint (<50MB RAM)
- No incoming connections required
- Supports multiple concurrent OTel collectors
- Built-in health monitoring

### 3. Dashboard
Modern React-based web interface:

- **Real-time Monitoring**: WebSocket-based live updates
- **Experiment Management**: Create and monitor experiments
- **Metrics Visualization**: Integration with Grafana
- **Pipeline Catalog**: Browse and deploy pipeline templates

## Data Flow

### 1. Experiment Creation
```
User → Dashboard → Phoenix API → Database
                              ↓
                         Task Queue
```

### 2. Task Distribution
```
Phoenix Agent → Long Poll → Phoenix API
                              ↓
                         Task Queue
                              ↓
                    Return Next Task
```

### 3. Metrics Collection
```
OTel Collector → Pushgateway → Prometheus
                                    ↓
                              Phoenix API
                                    ↓
                               Analysis
```

## Deployment Architecture

### Development Environment
```yaml
# docker-compose.yml
services:
  phoenix-api:
    ports: ["8080:8080", "8081:8081"]
  phoenix-agent:
    environment:
      - PHOENIX_API_URL=http://phoenix-api:8080
  postgres:
    ports: ["5432:5432"]
  prometheus:
    ports: ["9090:9090"]
  pushgateway:
    ports: ["9091:9091"]
```

### Production Deployment (Kubernetes)
```yaml
# Namespace: phoenix-system
- Phoenix API: StatefulSet with persistent storage
- Phoenix Agents: DaemonSet on target nodes
- PostgreSQL: Managed service or StatefulSet
- Prometheus Stack: Full observability
```

## Security Model

### Authentication & Authorization
- JWT-based authentication
- Role-based access control (RBAC)
- API key support for agents

### Network Security
- TLS encryption for all communications
- Network policies for pod-to-pod communication
- No direct agent-to-agent communication

### Data Security
- Encrypted storage for sensitive data
- Audit logging for all operations
- Secrets management via Kubernetes secrets

## Performance Characteristics

### Phoenix API
- Handles 10,000+ concurrent agents
- Sub-100ms task assignment latency
- Horizontal scaling via replicas

### Phoenix Agents
- <50MB memory footprint
- <1% CPU usage during normal operation
- Supports 100+ OTel collectors per agent

### Database
- PostgreSQL with connection pooling
- Optimized indexes for task queue operations
- Automatic vacuum and maintenance

## Development Standards

### Code Organization
```
phoenix/
├── pkg/                    # Shared packages
├── projects/
│   ├── phoenix-api/       # Control plane
│   ├── phoenix-agent/     # Data plane agent
│   ├── phoenix-cli/       # CLI tool
│   └── dashboard/         # Web UI
├── deployments/           # Deployment configs
└── docs/                  # Documentation
```

### API Standards
- RESTful API design
- OpenAPI 3.0 specification
- Consistent error responses
- Versioned endpoints (/api/v1)

### Testing Requirements
- Unit tests: >80% coverage
- Integration tests for all APIs
- End-to-end tests for critical paths
- Performance benchmarks

### Monitoring & Observability
- Structured logging (JSON)
- Prometheus metrics
- Distributed tracing (optional)
- Health check endpoints

## Migration Path

For organizations migrating from traditional observability solutions:

1. **Deploy Phoenix API** and supporting infrastructure
2. **Install Phoenix Agents** on target hosts
3. **Create experiments** to validate optimization
4. **Monitor results** and adjust configurations
5. **Promote to production** when KPIs are met

The platform's simple architecture ensures minimal operational overhead while delivering significant cost savings through intelligent metrics optimization.