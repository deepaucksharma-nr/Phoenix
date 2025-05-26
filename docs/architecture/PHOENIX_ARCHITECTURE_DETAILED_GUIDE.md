# Phoenix Platform - Ultra Detailed Architecture Guide

## Overview

This document provides an ultra-detailed explanation of the Phoenix Platform architecture as captured in the Mermaid diagram. The platform is designed as a cloud-native, microservices-based observability cost optimization system.

## Architecture Layers

### 1. Client Layer
The entry points for all interactions with the Phoenix Platform:

- **Web Dashboard**: React/TypeScript application providing visual interface
  - Real-time experiment monitoring via WebSocket
  - Pipeline deployment management
  - Metrics visualization
  - Configuration management

- **Phoenix CLI**: Go-based command-line tool
  - Experiment lifecycle management
  - Pipeline deployment operations
  - Plugin system for extensibility
  - Direct gRPC and REST API access

- **WebSocket Clients**: Real-time event subscribers
  - Experiment status updates
  - Metric streams
  - Alert notifications

- **Metrics Producers**: Systems generating telemetry data
  - Application metrics
  - Infrastructure metrics
  - Custom business metrics

### 2. API Gateway Layer

#### Platform API Service (port 8080/5050)
The central API gateway handling all external requests:

**Protocols:**
- REST API (port 8080): HTTP/JSON for web clients
- gRPC Server (port 5050): High-performance RPC for CLI/services
- WebSocket Server (port 8080/ws): Real-time bidirectional communication

**Handlers:**
- **Experiment Handler**: CRUD operations for experiments
- **Pipeline Handler**: Pipeline deployment management
- **Metrics Handler**: Metrics query and aggregation
- **Auth Handler**: JWT-based authentication/authorization

**Services:**
- **Experiment Service**: Business logic for A/B testing experiments
  - State management
  - Validation rules
  - Event publishing
  
- **Pipeline Deployment Service**: Pipeline lifecycle management
  - Deployment orchestration
  - Resource allocation
  - Status tracking

- **WebSocket Hub**: Real-time communication manager
  - Client connection management
  - Message broadcasting
  - Redis PubSub integration

**Store Layer:**
- **Experiment Store**: Persistence for experiment data
- **Pipeline Store**: Pipeline deployment persistence
- **Common PostgresStore**: Shared database abstraction
  - Connection pooling
  - Query optimization
  - Transaction management

### 3. Core Services Layer

#### Controller Service
The brain of the Phoenix Platform, managing experiment lifecycle:

**Components:**
- **Controller Main Loop**: Primary control loop
  - Event processing
  - State synchronization
  - Health monitoring

- **State Machine**: Experiment state transitions
  - Pending → Initializing → Running → Analyzing → Completed
  - Error handling and recovery
  - Rollback capabilities

- **Scheduler**: Time-based operations
  - Experiment duration management
  - Periodic analysis triggers
  - Cleanup operations

- **Experiment Controller**: Core experiment logic
  - Pipeline deployment coordination
  - Traffic splitting management
  - Result collection

- **Analysis Engine**: Statistical analysis
  - A/B test significance calculation
  - Metric comparison
  - Anomaly detection

- **Decision Engine**: Automated decision making
  - Winner selection
  - Rollout recommendations
  - Risk assessment

**Clients:**
- **Kubernetes Client**: K8s API interactions
  - CRD management
  - Pod/Service operations
  - ConfigMap/Secret handling

- **Generator Client**: Pipeline configuration requests
  - Template parameter passing
  - Configuration validation

#### Generator Service
Pipeline configuration generation engine:

**Components:**
- **Generator Server**: gRPC service endpoint
- **Template Engine**: Pipeline template processing
  - Variable substitution
  - Conditional logic
  - Template validation

- **Config Builder**: Final configuration assembly
  - YAML generation
  - Resource specification
  - Label/annotation management

**Templates:**
- **Baseline Template**: Standard pipeline configuration
- **Candidate Template**: Experimental pipeline configuration  
- **Adaptive Template**: Dynamic pipeline with auto-scaling

### 4. Data Processing Layer

#### Analytics Service
Advanced data analysis and visualization:

- **Analytics API**: RESTful interface for analysis requests
- **Correlation Analyzer**: Multi-metric correlation detection
- **Trend Analyzer**: Time-series trend identification
- **Chart Generator**: Visualization data preparation

#### Benchmark Service
Performance and cost benchmarking:

- **Benchmark API**: Benchmarking operations interface
- **Cost Analyzer**: Cloud cost calculation and optimization
- **Drift Detector**: Configuration/performance drift detection
- **Latency Validator**: SLA compliance checking
- **SQLite Store**: Local storage for benchmark data

#### Validator Service
Real-time validation and alerting:

- **Validator API**: Validation rule management
- **Metric Validator**: Metric value validation
- **Threshold Checker**: Threshold breach detection
- **Alert Generator**: Alert creation and routing

### 5. Kubernetes Operators Layer

#### Pipeline Operator
Custom Resource Definition (CRD) management:

- **Pipeline Controller**: Reconciliation loop
- **Pipeline Reconciler**: Desired vs actual state reconciliation
- **CRD Manager**: PhoenixProcessPipeline CRD operations

#### LoadSim Operator
Load simulation for experiments:

- **LoadSim Controller**: Job lifecycle management
- **Job Manager**: Kubernetes Job orchestration
- **Load Generator**: Synthetic load generation

### 6. Infrastructure Services Layer

#### Anomaly Detector
Machine learning-based anomaly detection:

- **Anomaly API**: Detection service interface
- **ML Detection Engine**: Statistical anomaly detection
- **Pattern Matcher**: Known pattern identification

#### Control Plane
Closed-loop control system:

**Observer:**
- **Observer Loop**: Continuous monitoring cycle
- **Metric Reader**: Prometheus query interface
- **State Tracker**: System state maintenance

**Actuator:**
- **Actuator Loop**: Action execution cycle
- **Config Writer**: Configuration updates
- **Action Executor**: Remediation actions

### 7. Data Collection Layer

#### OpenTelemetry Collectors
Centralized telemetry collection:

**Collectors:**
- **Main Collector (port 4317)**: Primary metrics ingestion
- **Observer Collector (port 4318)**: Control plane metrics

**Processors:**
- **Batch Processor**: Metric batching for efficiency
- **Filter Processor**: Metric filtering and sampling
- **Transform Processor**: Metric transformation and enrichment

**Exporters:**
- **Prometheus Exporter**: Time-series storage
- **New Relic Exporter**: APM integration
- **OTLP Exporter**: OpenTelemetry protocol export

### 8. Shared Packages Layer

#### go-common Package
Shared business logic and models:

- **Domain Models**: Core business entities
- **Service Interfaces**: Contract definitions
- **Event Bus**: Internal event distribution
- **Auth Package**: Authentication/authorization utilities
- **Metrics Package**: Metrics instrumentation
- **Common Clients**: Shared client implementations
- **Utilities**: Common helper functions

#### pkg Package
Infrastructure and technical packages:

**Database:**
- **Postgres Package**: PostgreSQL client wrapper
- **Redis Package**: Redis client wrapper
- **DB Migrations**: Schema version management

**Telemetry:**
- **Logging Package**: Structured logging (Zap)
- **Tracing Package**: Distributed tracing
- **Metrics Telemetry**: Metrics collection helpers

#### contracts Package
API and protocol definitions:

- **Protocol Buffers**: gRPC service definitions
- **OpenAPI Specs**: REST API specifications
- **K8s API Definitions**: CRD schemas

### 9. External Systems

- **PostgreSQL**: Primary data store for experiments, pipelines
- **Redis**: Cache layer and PubSub message broker
- **Prometheus**: Time-series metrics storage
- **Grafana**: Metrics visualization dashboards
- **Kubernetes API**: Container orchestration platform
- **OpenTelemetry Collector**: Vendor-neutral telemetry collection
- **New Relic**: Application performance monitoring

## Data Flow Patterns

### 1. Experiment Creation Flow
1. Client (Web/CLI) → REST/gRPC → API Gateway
2. API Gateway → Experiment Handler → Experiment Service
3. Experiment Service → Experiment Store → PostgreSQL
4. Experiment Service → Event Bus → Redis PubSub
5. Controller Service (via Event Bus) → State Machine
6. State Machine → Generator Client → Generator Service
7. Generator Service → Template Engine → Configuration
8. Controller → Kubernetes Client → K8s API
9. K8s API → Pipeline Operator → Pipeline Creation

### 2. Metrics Collection Flow
1. Application → OpenTelemetry SDK → OTLP
2. OTLP → OpenTelemetry Collector (Main)
3. Collector → Batch Processor → Filter Processor
4. Filter Processor → Transform Processor
5. Transform Processor → Prometheus Exporter
6. Prometheus Exporter → Prometheus Server
7. Prometheus → Grafana (Visualization)
8. Prometheus → Analytics Service (Analysis)

### 3. Real-time Updates Flow
1. Controller Service → Event Bus
2. Event Bus → Redis PubSub
3. Redis PubSub → WebSocket Hub
4. WebSocket Hub → WebSocket Clients
5. WebSocket Clients → Web Dashboard

### 4. Control Loop Flow
1. Observer → Metric Reader → Prometheus
2. Metric Reader → State Tracker → Redis
3. State Tracker → Analysis Engine
4. Analysis Engine → Decision Engine
5. Decision Engine → Actuator
6. Actuator → Action Executor → K8s API

## Security Architecture

### Authentication & Authorization
- JWT tokens for API authentication
- RBAC for Kubernetes operations
- TLS encryption for all communications
- Secret management via Kubernetes Secrets

### Network Security
- Network policies for pod-to-pod communication
- Service mesh ready architecture
- API rate limiting and throttling

## Scalability Patterns

### Horizontal Scaling
- All services designed as stateless
- Database connection pooling
- Redis for distributed caching
- Kubernetes HPA for auto-scaling

### Performance Optimization
- gRPC for internal service communication
- Batch processing for metrics
- Efficient query patterns with indexes
- Caching at multiple layers

## Observability

### Metrics
- Service-level metrics (RED method)
- Business metrics (experiments, pipelines)
- Infrastructure metrics (CPU, memory)

### Logging
- Structured logging with Zap
- Centralized log aggregation
- Correlation IDs for request tracing

### Tracing
- Distributed tracing support
- Span propagation across services
- Performance bottleneck identification

## Deployment Architecture

### Kubernetes Resources
- Deployments for all services
- StatefulSets for stateful components
- ConfigMaps for configuration
- Secrets for sensitive data
- Services for networking
- Ingress for external access

### GitOps Ready
- Declarative configuration
- Helm charts for packaging
- Kustomize for environment overlays
- ArgoCD compatible

This architecture provides a robust, scalable, and maintainable platform for observability cost optimization through intelligent A/B testing and metric analysis.