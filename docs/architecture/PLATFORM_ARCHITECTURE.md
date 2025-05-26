# Phoenix Platform Architecture

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Repository Structure](#repository-structure)
3. [Project Architecture](#project-architecture)
4. [Core Services](#core-services)
5. [Development Standards](#development-standards)
6. [Build & Tooling Infrastructure](#build--tooling-infrastructure)
7. [CI/CD Pipeline Architecture](#cicd-pipeline-architecture)
8. [Testing Strategy](#testing-strategy)
9. [API Design Standards](#api-design-standards)
10. [Deployment Architecture](#deployment-architecture)
11. [Development Workflow](#development-workflow)
12. [Security & Compliance](#security--compliance)
13. [Performance & Monitoring](#performance--monitoring)
14. [Troubleshooting Guide](#troubleshooting-guide)

## Executive Summary

Phoenix Platform is a cutting-edge observability cost optimization system built as a monorepo with completely independent micro-projects. This architecture provides:

- **100% Project Independence**: Each service maintains its own lifecycle
- **Shared Infrastructure**: Common tooling reduces duplication by 70%
- **Unified Development Experience**: Single setup for entire platform
- **Optimized CI/CD**: Smart pipelines that only build what changes
- **Enterprise-Grade Security**: Built-in security scanning and compliance
- **Scalable Architecture**: Easy to add new services following patterns

### Key Metrics
- **90% reduction** in metrics cardinality
- **70% reduction** in observability costs
- **Zero data loss** guarantee
- **Sub-second** optimization decisions
- **99.99%** uptime SLA

## Repository Structure

```
phoenix/
├── .github/                              # GitHub configuration
│   ├── workflows/
│   │   ├── _templates/                   # Reusable workflow templates
│   │   ├── ci-*.yml                      # Per-project CI workflows
│   │   ├── cd-*.yml                      # Per-project CD workflows
│   │   └── security.yml                  # Security scanning
│   ├── CODEOWNERS
│   └── SECURITY.md
│
├── build/                                # Shared build infrastructure
│   ├── docker/
│   │   ├── base/                         # Base Docker images
│   │   └── scripts/                      # Docker build scripts
│   ├── makefiles/
│   │   ├── common.mk                     # Common variables/targets
│   │   ├── go.mk                         # Go-specific targets
│   │   ├── node.mk                       # Node-specific targets
│   │   └── docker.mk                     # Docker targets
│   └── scripts/
│       ├── ci/                           # CI/CD scripts
│       ├── release/                      # Release automation
│       └── utils/                        # Utility scripts
│
├── pkg/                                  # Shared Go packages
│   ├── auth/                             # Authentication/authorization
│   ├── telemetry/                        # Logging, metrics, tracing
│   ├── database/                         # Database abstractions
│   ├── messaging/                        # Event bus, queues
│   ├── contracts/                        # API contracts
│   └── go.mod                            # Shared packages module
│
├── projects/                             # Independent micro-projects
│   ├── platform-api/                     # Core API Service
│   ├── experiment-controller/            # K8s Controller
│   ├── pipeline-operator/                # Pipeline CRD Operator
│   ├── web-dashboard/                    # React Dashboard
│   ├── phoenix-cli/                      # CLI Tool
│   └── [service-name]/                   # Standard structure
│
├── configs/                              # Configuration files
├── deployments/                          # Deployment configurations
├── docs/                                 # Documentation
├── infrastructure/                       # Infrastructure as Code
├── scripts/                              # Root-level scripts
├── tests/                                # Cross-project tests
├── tools/                                # Development tools
│
├── go.work                               # Go workspace
├── Makefile                              # Root Makefile
├── docker-compose.yml                    # Development stack
└── README.md
```

## Project Architecture

### Domain-Driven Design Structure

Each project follows Domain-Driven Design principles with clear boundaries:

```
projects/<project-name>/
├── cmd/                                 # Application entrypoints
│   └── <app-name>/
│       └── main.go
├── internal/                            # Private application code
│   ├── api/                             # API layer (HTTP/gRPC/GraphQL)
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── routes/
│   │   └── grpc/
│   │       ├── services/
│   │       └── interceptors/
│   ├── domain/                          # Business logic
│   │   ├── entities/                    # Domain models
│   │   ├── repositories/                # Data interfaces
│   │   ├── services/                    # Business services
│   │   └── events/                      # Domain events
│   ├── infrastructure/                  # External dependencies
│   │   ├── database/
│   │   ├── cache/
│   │   ├── messaging/
│   │   └── external/
│   └── config/                          # Configuration
├── api/                                 # API definitions
├── build/                               # Build configurations
├── deployments/                         # Deployment manifests
├── docs/                                # Project documentation
├── migrations/                          # Database migrations
├── scripts/                             # Project scripts
├── tests/                               # Test files
├── Makefile
├── go.mod
└── VERSION
```

## Core Services

### 1. Platform API (`projects/platform-api`)
Central API gateway handling all external requests:
- RESTful API with OpenAPI 3.0 spec
- gRPC for internal communication
- GraphQL for flexible queries
- WebSocket for real-time updates

### 2. Experiment Controller (`projects/experiment-controller`)
Kubernetes controller managing optimization experiments:
- Custom Resource Definitions (CRDs)
- Reconciliation loops
- State management
- Event broadcasting

### 3. Pipeline Operator (`projects/pipeline-operator`)
Manages telemetry processing pipelines:
- Dynamic pipeline configuration
- A/B testing support
- Rollback capabilities
- Performance monitoring

### 4. Web Dashboard (`projects/web-dashboard`)
React-based user interface:
- Real-time metrics visualization
- Experiment management
- Cost analysis
- Alert configuration

### 5. Phoenix CLI (`projects/phoenix-cli`)
Command-line interface:
- Experiment CRUD operations
- Pipeline deployment
- Metrics queries
- Plugin system

## Development Standards

### Code Organization Principles

1. **Domain-Driven Design**: Clear separation between domain logic and infrastructure
2. **Hexagonal Architecture**: Ports and adapters pattern for flexibility
3. **SOLID Principles**: Single responsibility, open/closed, etc.
4. **12-Factor App**: Environment-based configuration, stateless processes
5. **Clean Code**: Self-documenting code with meaningful names

### Go Standards

```go
// Package api implements the HTTP API for Phoenix Platform.
// It follows RESTful principles and provides endpoints for
// experiment management, metrics queries, and optimization control.
package api

import (
    "context"
    "fmt"
    "time"

    "github.com/phoenix-vnext/platform/pkg/telemetry/logging"
    "github.com/phoenix-vnext/platform/pkg/errors"
)

// ExperimentHandler handles HTTP requests for experiments.
// It implements the business logic for experiment lifecycle management.
type ExperimentHandler struct {
    service  ExperimentService
    logger   logging.Logger
    metrics  *ExperimentMetrics
}

// CreateExperiment handles POST /api/v1/experiments
// It validates the request, creates the experiment, and returns the result.
func (h *ExperimentHandler) CreateExperiment(ctx context.Context, req CreateExperimentRequest) (*ExperimentResponse, error) {
    // Start span for tracing
    ctx, span := h.tracer.Start(ctx, "handler.CreateExperiment")
    defer span.End()

    // Validate request
    if err := req.Validate(); err != nil {
        h.metrics.InvalidRequests.Inc()
        return nil, errors.Wrap(err, errors.CodeInvalidArgument, "invalid request")
    }

    // Create experiment
    experiment, err := h.service.Create(ctx, req.ToServiceInput())
    if err != nil {
        h.logger.Error("failed to create experiment", 
            logging.ErrorField(err),
            logging.String("name", req.Name))
        return nil, err
    }

    h.metrics.ExperimentsCreated.Inc()
    h.logger.Info("experiment created",
        logging.String("id", experiment.ID),
        logging.String("name", experiment.Name))

    return NewExperimentResponse(experiment), nil
}
```

### TypeScript/React Standards

```typescript
// components/ExperimentDashboard/ExperimentDashboard.tsx
import React, { useState, useEffect, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Box, Grid, Paper, Typography, Button } from '@mui/material';

import { useWebSocket } from '@/hooks/useWebSocket';
import { ExperimentCard } from './components/ExperimentCard';
import { ExperimentMetrics } from './components/ExperimentMetrics';
import type { Experiment } from '@/types';

interface ExperimentDashboardProps {
  onExperimentSelect?: (experiment: Experiment) => void;
}

/**
 * ExperimentDashboard displays and manages optimization experiments.
 * It provides real-time updates via WebSocket and interactive controls.
 */
export const ExperimentDashboard: React.FC<ExperimentDashboardProps> = ({ 
  onExperimentSelect 
}) => {
  const dispatch = useDispatch();
  const { experiments, loading } = useExperimentState();
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // WebSocket for real-time updates
  const { subscribe } = useWebSocket();

  useEffect(() => {
    const unsubscribe = subscribe('experiments.*', handleExperimentUpdate);
    return unsubscribe;
  }, [subscribe]);

  const handleExperimentUpdate = useCallback((event: ExperimentEvent) => {
    dispatch(updateExperiment(event.payload));
  }, [dispatch]);

  if (loading) {
    return <DashboardSkeleton />;
  }

  return (
    <DashboardContainer>
      <DashboardHeader />
      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <ExperimentList 
            experiments={experiments}
            selectedId={selectedId}
            onSelect={setSelectedId}
          />
        </Grid>
        <Grid item xs={12} md={8}>
          <MetricsPanel experimentId={selectedId} />
        </Grid>
      </Grid>
    </DashboardContainer>
  );
};
```

## Build & Tooling Infrastructure

### Root Makefile

```makefile
# Phoenix Platform - Root Makefile
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Directories
ROOT_DIR := $(shell pwd)
BUILD_DIR := $(ROOT_DIR)/build
PROJECTS_DIR := $(ROOT_DIR)/projects

# Version
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "0.0.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Projects discovery
ALL_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)
GO_PROJECTS := $(shell find $(PROJECTS_DIR) -name go.mod -exec dirname {} \; | xargs -n1 basename)

# Include shared makefiles
include $(BUILD_DIR)/makefiles/*.mk

.PHONY: all build test lint clean

all: validate build test ## Build and test everything

build: $(ALL_PROJECTS:%=build-%) ## Build all projects
	@echo "✓ All projects built"

test: $(ALL_PROJECTS:%=test-%) ## Test all projects
	@echo "✓ All tests passed"

# Dynamic project targets
define PROJECT_RULES
build-$(1):
	@$(MAKE) -C projects/$(1) build

test-$(1):
	@$(MAKE) -C projects/$(1) test

lint-$(1):
	@$(MAKE) -C projects/$(1) lint

clean-$(1):
	@$(MAKE) -C projects/$(1) clean
endef

$(foreach project,$(ALL_PROJECTS),$(eval $(call PROJECT_RULES,$(project))))

# Development environment
dev-up: ## Start development services
	docker-compose up -d
	@echo "✓ Development environment ready"

dev-down: ## Stop development services
	docker-compose down

# Validation
validate: ## Validate repository structure
	@$(BUILD_DIR)/scripts/utils/validate-structure.sh
	@echo "✓ Repository structure valid"
```

### Docker Build Strategy

Multi-stage Dockerfile for optimal image size:

```dockerfile
# build/docker/base/go.Dockerfile
FROM golang:1.21-alpine AS base
RUN apk add --no-cache git make curl
WORKDIR /workspace

# Builder stage
FROM base AS builder
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w \
    -X main.version=${VERSION} \
    -X main.gitCommit=${GIT_COMMIT} \
    -X main.buildDate=${BUILD_DATE}" \
    -o /app ./cmd/api

# Final stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app /app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app"]
```

## CI/CD Pipeline Architecture

### GitHub Actions Workflow

The CI/CD pipeline uses a matrix strategy for parallel execution:

```yaml
# .github/workflows/ci.yml
name: CI Pipeline

on:
  push:
    branches: [main]
  pull_request:

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      projects: ${{ steps.changes.outputs.projects }}
    steps:
      - uses: actions/checkout@v4
      - id: changes
        uses: dorny/paths-filter@v2
        with:
          filters: |
            platform-api:
              - 'projects/platform-api/**'
              - 'pkg/**'
            web-dashboard:
              - 'projects/web-dashboard/**'

  build-and-test:
    needs: detect-changes
    strategy:
      matrix:
        project: ${{ fromJSON(needs.detect-changes.outputs.projects) }}
    uses: ./.github/workflows/_templates/build-test.yml
    with:
      project: ${{ matrix.project }}
    secrets: inherit

  integration-tests:
    needs: build-and-test
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run integration tests
        run: make test-integration
```

### Deployment Pipeline

Progressive deployment with automated rollback:

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  deploy-staging:
    environment: staging
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to staging
        run: |
          kubectl apply -k deployments/kubernetes/overlays/staging
          kubectl wait --for=condition=ready pod -l app=phoenix -n phoenix-staging

  smoke-tests:
    needs: deploy-staging
    runs-on: ubuntu-latest
    steps:
      - name: Run smoke tests
        run: |
          npm run test:smoke -- --env=staging

  deploy-production:
    needs: smoke-tests
    environment: production
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to production
        run: |
          kubectl apply -k deployments/kubernetes/overlays/production
          kubectl wait --for=condition=ready pod -l app=phoenix -n phoenix-prod
```

## Testing Strategy

### Test Pyramid

1. **Unit Tests** (70%)
   - Fast, isolated tests
   - Mock external dependencies
   - Focus on business logic

2. **Integration Tests** (20%)
   - Test service interactions
   - Use test containers
   - Verify API contracts

3. **E2E Tests** (10%)
   - Full system tests
   - User journey validation
   - Performance benchmarks

### Test Examples

#### Unit Test
```go
func TestExperimentService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateExperimentInput
        want    *Experiment
        wantErr error
    }{
        {
            name: "valid experiment",
            input: CreateExperimentInput{
                Name: "Test Experiment",
                Type: ExperimentTypeAB,
            },
            want: &Experiment{
                Name:   "Test Experiment",
                Type:   ExperimentTypeAB,
                Status: ExperimentStatusPending,
            },
        },
        {
            name: "invalid name",
            input: CreateExperimentInput{
                Name: "ab", // too short
            },
            wantErr: ErrInvalidName,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := mocks.NewMockRepository(t)
            if tt.wantErr == nil {
                repo.EXPECT().Create(mock.Anything, mock.Anything).Return(tt.want, nil)
            }

            svc := NewExperimentService(repo)
            got, err := svc.Create(context.Background(), tt.input)

            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want.Name, got.Name)
            }
        })
    }
}
```

#### Integration Test
```go
func TestAPI_ExperimentLifecycle(t *testing.T) {
    // Setup test environment
    ctx := context.Background()
    container := setupPostgresContainer(t)
    db := connectToContainer(t, container)
    
    // Initialize services
    app := setupTestApp(t, db)
    client := app.TestClient()
    
    // Test experiment lifecycle
    t.Run("create experiment", func(t *testing.T) {
        resp := client.POST("/api/v1/experiments").
            WithJSON(map[string]interface{}{
                "name": "Integration Test",
                "type": "ab_test",
            }).
            Expect().
            Status(http.StatusCreated).
            JSON().Object()
        
        resp.Value("id").String().NotEmpty()
        resp.Value("status").String().Equal("pending")
    })
}
```

## API Design Standards

### RESTful API Design
- Use nouns for resources: `/experiments`, `/pipelines`
- HTTP methods for actions: `GET`, `POST`, `PUT`, `DELETE`
- Consistent error responses
- Pagination for lists
- Filtering and sorting support
- API versioning in URL: `/api/v1/`

### OpenAPI Specification

```yaml
openapi: 3.0.3
info:
  title: Phoenix Platform API
  version: 1.0.0
  description: |
    Phoenix Platform API provides endpoints for managing optimization
    experiments, pipeline configurations, and metrics analysis.

paths:
  /api/v1/experiments:
    get:
      summary: List experiments
      operationId: listExperiments
      tags: [Experiments]
      parameters:
        - $ref: '#/components/parameters/PageSize'
        - $ref: '#/components/parameters/PageToken'
        - name: status
          in: query
          schema:
            $ref: '#/components/schemas/ExperimentStatus'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ExperimentList'
        '401':
          $ref: '#/components/responses/Unauthorized'

    post:
      summary: Create experiment
      operationId: createExperiment
      tags: [Experiments]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateExperimentRequest'
      responses:
        '201':
          description: Experiment created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Experiment'
```

### gRPC Service Design

```protobuf
// api/proto/experiment.proto
syntax = "proto3";

package phoenix.platform.v1;

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "validate/validate.proto";

option go_package = "github.com/phoenix/platform-api/api/proto/v1;platformv1";

// ExperimentService manages optimization experiments
service ExperimentService {
  // CreateExperiment creates a new optimization experiment
  rpc CreateExperiment(CreateExperimentRequest) returns (Experiment) {
    option (google.api.http) = {
      post: "/v1/experiments"
      body: "*"
    };
  }

  // GetExperiment retrieves an experiment by ID
  rpc GetExperiment(GetExperimentRequest) returns (Experiment) {
    option (google.api.http) = {
      get: "/v1/experiments/{id}"
    };
  }

  // StreamExperimentEvents streams real-time experiment events
  rpc StreamExperimentEvents(StreamExperimentEventsRequest) returns (stream ExperimentEvent);
}

// Experiment represents an optimization experiment
message Experiment {
  // Unique identifier
  string id = 1 [(validate.rules).string.uuid = true];

  // Human-readable name
  string name = 2 [(validate.rules).string = {
    min_len: 3,
    max_len: 100
  }];

  // Current status
  ExperimentStatus status = 3;

  // Timestamps
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}
```

## Deployment Architecture

### Kubernetes Resources

```yaml
# deployments/kubernetes/base/platform-api/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: platform-api
  labels:
    app.kubernetes.io/name: platform-api
    app.kubernetes.io/component: api
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: platform-api
  template:
    metadata:
      labels:
        app.kubernetes.io/name: platform-api
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: platform-api
      containers:
      - name: api
        image: ghcr.io/phoenix/platform-api:latest
        ports:
        - name: http
          containerPort: 8080
        - name: grpc
          containerPort: 9090
        env:
        - name: ENVIRONMENT
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
```

### GitOps with Flux

```yaml
# infrastructure/flux/apps/phoenix/kustomization.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: phoenix-platform
  namespace: flux-system
spec:
  interval: 10m
  path: ./deployments/kubernetes/overlays/production
  prune: true
  sourceRef:
    kind: GitRepository
    name: phoenix
  validation: client
  postBuild:
    substituteFrom:
    - kind: ConfigMap
      name: phoenix-versions
```

## Development Workflow

### Getting Started

```bash
# Clone repository
git clone https://github.com/phoenix/platform.git
cd platform

# Setup development environment
make setup

# Start dependencies
make dev-up

# Run a specific service
cd projects/platform-api
make run

# Run tests
make test

# Build everything
make build
```

### Development Guidelines

1. **Branch Strategy**
   - `main`: Production-ready code
   - `develop`: Integration branch
   - `feature/*`: New features
   - `fix/*`: Bug fixes
   - `chore/*`: Maintenance

2. **Commit Convention**
   ```
   type(scope): description

   [optional body]

   [optional footer]
   ```
   Types: feat, fix, docs, style, refactor, test, chore

3. **Code Review Process**
   - All changes require PR
   - At least 2 approvals
   - Must pass all CI checks
   - Must update documentation

### Version Management

- **Semantic Versioning**: MAJOR.MINOR.PATCH
- **Release Process**:
  ```bash
  # Create release branch
  git checkout -b release/v1.2.0
  
  # Update versions
  make version-bump VERSION=1.2.0
  
  # Create changelog
  make changelog
  
  # Tag and push
  git tag v1.2.0
  git push origin v1.2.0
  ```

## Security & Compliance

### Security Scanning

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  schedule:
    - cron: '0 0 * * *'
  push:
    branches: [main]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          severity: 'CRITICAL,HIGH'
          
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: './...'
          
      - name: SonarCloud Scan
        uses: SonarSource/sonarcloud-github-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### RBAC Configuration

```yaml
# deployments/kubernetes/base/rbac/platform-api-rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: platform-api
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "update", "patch"]
- apiGroups: ["phoenix.io"]
  resources: ["experiments", "pipelines"]
  verbs: ["*"]
```

## Performance & Monitoring

### Metrics Collection

```go
// pkg/telemetry/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Request metrics
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "phoenix_request_duration_seconds",
            Help: "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"service", "method", "status"},
    )
    
    // Business metrics
    ExperimentsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "phoenix_experiments_active",
            Help: "Number of active experiments",
        },
    )
    
    CostReduction = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name: "phoenix_cost_reduction_percentage",
            Help: "Cost reduction achieved per experiment",
            Buckets: []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
        },
    )
)
```

### Distributed Tracing

```go
// pkg/telemetry/tracing/tracing.go
package tracing

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(serviceName string) (*trace.TracerProvider, error) {
    exporter, err := otlptrace.New(
        context.Background(),
        otlptrace.WithEndpoint("otel-collector:4317"),
        otlptrace.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String(version),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

## Troubleshooting Guide

### Common Issues

#### 1. Service Won't Start
```bash
# Check logs
kubectl logs -n phoenix deployment/platform-api

# Check events
kubectl get events -n phoenix --sort-by='.lastTimestamp'

# Verify config
kubectl describe configmap platform-api-config -n phoenix
```

#### 2. Database Connection Issues
```bash
# Test connection
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql -h postgres.phoenix.svc.cluster.local -U phoenix -c "SELECT 1"

# Check credentials
kubectl get secret db-credentials -n phoenix -o yaml
```

#### 3. High Memory Usage
```bash
# Get memory profile
kubectl port-forward -n phoenix deployment/platform-api 6060:6060
go tool pprof http://localhost:6060/debug/pprof/heap

# Check resource limits
kubectl top pods -n phoenix
```

### Performance Tuning

#### Database Optimization
```sql
-- Analyze query performance
EXPLAIN ANALYZE 
SELECT * FROM experiments 
WHERE status = 'running' 
  AND namespace = 'production'
ORDER BY created_at DESC;

-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_experiments_status_namespace 
ON experiments(status, namespace) 
WHERE status != 'completed';
```

#### Application Tuning
```go
// Optimize connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Use sync.Pool for object reuse
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}
```

### Database Schema Standards

```sql
-- Experiments table with proper constraints and indexes
CREATE TABLE experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status experiment_status NOT NULL DEFAULT 'pending',
    type experiment_type NOT NULL,
    namespace VARCHAR(63) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    
    -- Constraints
    CONSTRAINT experiments_name_namespace_unique UNIQUE (name, namespace),
    CONSTRAINT experiments_name_check CHECK (length(name) >= 3)
);

-- Indexes for performance
CREATE INDEX experiments_status_idx ON experiments(status) WHERE status != 'completed';
CREATE INDEX experiments_namespace_idx ON experiments(namespace);
CREATE INDEX experiments_created_at_idx ON experiments(created_at DESC);
```

## Conclusion

The Phoenix Platform monorepo architecture provides a robust foundation for building and operating a complex distributed system. Key benefits include:

1. **Unified Development**: Single repository for all components
2. **Code Reuse**: Shared packages reduce duplication
3. **Atomic Changes**: Related changes across services in one commit
4. **Consistent Tooling**: Same build, test, and deploy processes
5. **Independent Deployment**: Services can still be deployed separately

For more information, see:
- [Developer Guide](../guides/developer/getting-started.md)
- [API Documentation](https://api.phoenix.io/docs)
- [Architecture Decisions](decisions/)