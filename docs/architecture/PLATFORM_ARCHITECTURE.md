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

Phoenix Platform is an observability cost optimization system that reduces metrics cardinality by up to 70% while maintaining critical visibility. The platform uses an agent-based architecture with centralized control plane, task queue system, and lightweight distributed agents that poll for work.

### Key Achievements
- **70% reduction** in metrics cardinality (demonstrated)
- **70% reduction** in observability costs
- **A/B testing** with baseline/candidate pipelines
- **Real-time monitoring** via WebSocket
- **Agent-based** task distribution with PostgreSQL queue

## Architecture Overview

Phoenix Platform implements an agent-based architecture with task polling:

```
┌─────────────────────────────────────────────────────────────┐
│                     Phoenix Platform                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐         ┌─────────────────┐          │
│  │   Phoenix API   │         │   Dashboard     │          │
│  │  (Port 8080)    │◄────────┤   (React UI)    │          │
│  │  + WebSocket    │         └─────────────────┘          │
│  └────────┬────────┘                                         │
│           │                                                 │
│           │ Task Queue (PostgreSQL)                         │
│           │ Long-polling (30s timeout)                      │
│           │                                                 │
│      ┌────▼─────┐                                         │
│      │ Phoenix   │                                         │
│      │ Agents    │────────► OpenTelemetry ────► Backends │
│      │ (X-Agent-Host-ID)       Collector                   │
│      └──────────┘                                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Phoenix API (Control Plane)
The central control plane running on port 8080 (REST + WebSocket):

- **Experiment Management**: A/B testing with baseline/candidate pipelines
- **Task Queue**: PostgreSQL-based queue with atomic task assignment
- **WebSocket Server**: Real-time updates on same port as REST API
- **REST API v2**: Full API for experiments, agents, and pipelines
- **Metrics Analysis**: Cardinality reduction and cost savings calculations

**Technology Stack**:
- Go 1.21+
- PostgreSQL 15+ (primary datastore)
- WebSocket for real-time monitoring
- Pipeline templates (Adaptive Filter, TopK, Hybrid)

### 2. Phoenix Agent (Data Plane)
Lightweight agents that poll for tasks using X-Agent-Host-ID authentication:

- **Task Polling**: Long-polling with 30-second timeout
- **OTel Management**: Deploys baseline/candidate pipeline configurations
- **Authentication**: Uses X-Agent-Host-ID header for identification
- **Pipeline Templates**: Supports Adaptive Filter, TopK, and Hybrid
- **Status Reporting**: Reports metrics and experiment results

**Key Features**:
- Minimal resource footprint (<50MB RAM)
- Outbound-only connections (security)
- Concurrent A/B test execution
- Automatic task retry on failure

### 3. Dashboard
Modern React 18 + Vite web interface:

- **Real-time Monitoring**: WebSocket connection for live updates
- **Experiment Creation**: A/B testing setup with pipeline selection
- **Live Cost Analytics**: Real-time cost reduction visualization
- **Pipeline Templates**: Pre-configured optimization strategies
- **Agent Fleet View**: Monitor all connected agents

## Data Flow

### 1. Experiment Creation
```
User → Dashboard → Phoenix API → Database
                              ↓
                         Task Queue
```

### 2. Task Distribution
```
Phoenix Agent → Poll /api/v2/tasks/poll → Phoenix API
       │                                         │
       │         X-Agent-Host-ID: agent-123     │
       │                                         │
       └──────── 30s Long Poll ──────────────┘
                              ↓
                    PostgreSQL Task Queue
                              ↓
                    Return Next Task (Atomic)
```

### 3. Metrics Collection & Analysis
```
Baseline Pipeline  ──┐
                     ├──► OTel Collector ──► Metrics Backend
                     │                            │
Candidate Pipeline ──┘                            │
                                                  ↓
                              Phoenix API (Analysis Engine)
                                        │
                  ┌───────────────────────┬───────────────────────┐
                  │  Cardinality: -70%     │  Cost Savings: 70%    │
                  └───────────────────────┴───────────────────────┘
                                    ↓
                              WebSocket ──► Dashboard
```

## Deployment Architecture

### Development Environment
```yaml
# docker-compose.yml
services:
  phoenix-api:
    ports: ["8080:8080"]  # REST API + WebSocket
    environment:
      - DATABASE_URL=postgresql://phoenix:phoenix@postgres:5432/phoenix
  phoenix-agent:
    environment:
      - PHOENIX_API_URL=http://phoenix-api:8080
      - AGENT_HOST_ID=${HOSTNAME}
      - TASK_POLL_INTERVAL=30s
  postgres:
    ports: ["5432:5432"]
    volumes:
      - postgres_data:/var/lib/postgresql/data
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
- JWT-based authentication for users
- X-Agent-Host-ID header for agent authentication
- Role-based access control (RBAC)
- Task queue with row-level security

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
- 30-second long-polling for task distribution
- PostgreSQL-based task queue with atomic operations
- WebSocket support for real-time updates

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
- RESTful API v2 design
- OpenAPI 3.0 specification
- Consistent error responses
- Versioned endpoints (/api/v2)
- WebSocket on same port as REST

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