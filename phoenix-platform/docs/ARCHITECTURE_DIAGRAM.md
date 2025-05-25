# Phoenix Platform Architecture

## System Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        UI[Dashboard UI]
        CLI[CLI Tools]
        API_Client[API Clients]
    end

    subgraph "API Gateway Layer"
        GRPC[gRPC API<br/>:50051]
        REST[REST API<br/>:8080]
    end

    subgraph "Core Services"
        EC[Experiment Controller<br/>:50051/:8081]
        CG[Config Generator<br/>:8082]
        SM[State Machine]
        SCH[Scheduler]
    end

    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Database)]
        REDIS[(Redis Cache<br/>Optional)]
    end

    subgraph "Kubernetes Layer"
        PO[Pipeline Operator]
        LSO[LoadSim Operator]
        CRDS[Custom Resources]
    end

    subgraph "Observability"
        PROM[Prometheus<br/>:9090]
        GRAF[Grafana<br/>:3001]
        METRICS[Metrics Endpoint<br/>:8081]
    end

    subgraph "External Systems"
        GIT[Git Repository]
        K8S[Kubernetes API]
        ARGO[ArgoCD]
    end

    %% Client connections
    UI --> REST
    CLI --> GRPC
    API_Client --> GRPC

    %% API Gateway to Services
    GRPC --> EC
    REST --> EC
    REST --> CG

    %% Core Service interactions
    EC --> SM
    EC --> SCH
    SM --> CG
    SM --> PO
    EC --> PG
    CG --> GIT

    %% Kubernetes interactions
    PO --> K8S
    PO --> CRDS
    LSO --> K8S
    ARGO --> GIT

    %% Monitoring
    EC --> METRICS
    METRICS --> PROM
    PROM --> GRAF

    %% Data flow
    EC -.-> REDIS
    SCH --> EC

    classDef service fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef data fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef external fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef client fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px
    classDef k8s fill:#fce4ec,stroke:#880e4f,stroke-width:2px

    class EC,CG,SM,SCH service
    class PG,REDIS data
    class GIT,K8S,ARGO external
    class UI,CLI,API_Client client
    class PO,LSO,CRDS k8s
```

## Component Descriptions

### Client Layer
- **Dashboard UI**: React-based web interface for experiment management
- **CLI Tools**: Command-line tools for automation and scripting
- **API Clients**: External service integrations

### Core Services

#### Experiment Controller (Port 50051/8081)
- Manages experiment lifecycle
- Handles state transitions
- Provides gRPC API for experiment operations
- Exposes Prometheus metrics on port 8081

#### Config Generator (Port 8082)
- Generates OpenTelemetry collector configurations
- Manages pipeline templates
- Provides HTTP API for configuration generation

#### State Machine
- Manages experiment state transitions
- Orchestrates workflow steps
- Ensures valid state changes

#### Scheduler
- Periodically processes experiments
- Triggers state transitions
- Manages timed operations

### Data Layer

#### PostgreSQL Database
- Primary data store for experiments
- Stores experiment metadata and state
- Handles concurrent access

#### Redis Cache (Optional)
- Caches frequently accessed data
- Pub/sub for real-time updates
- Session storage

### Kubernetes Integration

#### Pipeline Operator
- Deploys pipeline configurations
- Manages PhoenixProcessPipeline CRDs
- Monitors pipeline health

#### LoadSim Operator
- Creates load simulation jobs
- Manages LoadSimulationJob CRDs
- Collects performance metrics

### Observability

#### Prometheus
- Collects metrics from all services
- Stores time-series data
- Provides query interface

#### Grafana
- Visualizes metrics
- Custom dashboards for experiments
- Alerting capabilities

## Data Flow

### Experiment Creation Flow

```mermaid
sequenceDiagram
    participant Client
    participant Controller
    participant StateMachine
    participant Generator
    participant Git
    participant K8s

    Client->>Controller: CreateExperiment
    Controller->>Controller: Validate & Store
    Controller->>StateMachine: ProcessExperiment
    StateMachine->>Generator: GenerateConfig
    Generator->>Git: Commit Config
    StateMachine->>K8s: Deploy Pipeline
    K8s-->>Client: Experiment Running
```

### State Transitions

```mermaid
stateDiagram-v2
    [*] --> Pending: Create
    Pending --> Initializing: Start
    Initializing --> Running: Deploy Success
    Initializing --> Failed: Deploy Error
    Running --> Analyzing: Complete
    Running --> Failed: Runtime Error
    Analyzing --> Completed: Analysis Done
    Completed --> [*]
    Failed --> [*]
    
    Pending --> Cancelled: User Cancel
    Initializing --> Cancelled: User Cancel
    Running --> Cancelled: User Cancel
    Cancelled --> [*]
```

## Security Considerations

1. **Authentication**: JWT tokens for API access
2. **Authorization**: Role-based access control
3. **Network Security**: TLS for external communications
4. **Secrets Management**: Kubernetes secrets for sensitive data
5. **Audit Logging**: All operations logged for compliance

## Scalability

1. **Horizontal Scaling**: All services are stateless and can scale
2. **Database Pooling**: Connection pooling for PostgreSQL
3. **Caching Strategy**: Redis for frequently accessed data
4. **Load Balancing**: Kubernetes service load balancing
5. **Resource Limits**: Defined CPU/memory limits for all pods

## High Availability

1. **Service Replicas**: Multiple instances of each service
2. **Database HA**: PostgreSQL with replication
3. **Health Checks**: Liveness and readiness probes
4. **Circuit Breakers**: Prevent cascade failures
5. **Graceful Degradation**: Services continue with reduced functionality