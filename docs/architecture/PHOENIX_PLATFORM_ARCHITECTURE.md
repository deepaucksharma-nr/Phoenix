# Phoenix Platform - Ultimate Monorepo Architecture & Documentation

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Repository Structure](#repository-structure)
3. [Project Architecture](#project-architecture)
4. [Development Standards](#development-standards)
5. [Build & Tooling Infrastructure](#build--tooling-infrastructure)
6. [CI/CD Pipeline Architecture](#cicd-pipeline-architecture)
7. [Testing Strategy](#testing-strategy)
8. [Documentation Standards](#documentation-standards)
9. [Deployment Architecture](#deployment-architecture)
10. [Development Workflow](#development-workflow)
11. [Version Management](#version-management)
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

## Repository Structure

```
phoenix/
├── .github/                              # GitHub configuration
│   ├── workflows/
│   │   ├── _templates/                   # Reusable workflow templates
│   │   │   ├── go-service.yml
│   │   │   ├── node-service.yml
│   │   │   ├── docker-build.yml
│   │   │   └── security-scan.yml
│   │   ├── ci-*.yml                      # Per-project CI workflows
│   │   ├── cd-*.yml                      # Per-project CD workflows
│   │   ├── integration.yml               # Cross-project integration
│   │   ├── security.yml                  # Security scanning
│   │   ├── release.yml                   # Release automation
│   │   └── maintenance.yml               # Dependency updates
│   ├── dependabot.yml
│   ├── CODEOWNERS
│   ├── SECURITY.md
│   └── pull_request_template.md
│
├── build/                                # Shared build infrastructure
│   ├── docker/
│   │   ├── base/
│   │   │   ├── go.Dockerfile            # Base Go image
│   │   │   ├── node.Dockerfile          # Base Node image
│   │   │   └── security-scanner.Dockerfile
│   │   └── scripts/
│   │       ├── docker-build.sh
│   │       └── docker-push.sh
│   ├── makefiles/
│   │   ├── common.mk                     # Common variables/targets
│   │   ├── go.mk                         # Go-specific targets
│   │   ├── node.mk                       # Node-specific targets
│   │   ├── docker.mk                     # Docker targets
│   │   ├── k8s.mk                        # Kubernetes targets
│   │   └── test.mk                       # Testing targets
│   └── scripts/
│       ├── ci/
│       │   ├── setup-env.sh
│       │   ├── run-tests.sh
│       │   └── publish-artifacts.sh
│       ├── release/
│       │   ├── bump-version.sh
│       │   ├── generate-changelog.sh
│       │   └── create-release.sh
│       └── utils/
│           ├── check-dependencies.sh
│           └── validate-structure.sh
│
├── deployments/                          # Deployment configurations
│   ├── kubernetes/
│   │   ├── base/
│   │   │   ├── namespace.yaml
│   │   │   ├── rbac.yaml
│   │   │   └── network-policies.yaml
│   │   ├── operators/
│   │   │   ├── crds/
│   │   │   └── controllers/
│   │   └── overlays/
│   │       ├── development/
│   │       ├── staging/
│   │       └── production/
│   ├── helm/
│   │   ├── phoenix-platform/             # Umbrella chart
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── values-dev.yaml
│   │   │   ├── values-staging.yaml
│   │   │   ├── values-prod.yaml
│   │   │   └── templates/
│   │   └── charts/                       # Individual charts
│   ├── terraform/
│   │   ├── modules/
│   │   │   ├── networking/
│   │   │   ├── compute/
│   │   │   ├── storage/
│   │   │   └── security/
│   │   └── environments/
│   │       ├── dev/
│   │       ├── staging/
│   │       └── production/
│   └── ansible/
│       ├── playbooks/
│       └── inventories/
│
├── pkg/                                  # Shared Go packages
│   ├── auth/
│   │   ├── jwt/
│   │   ├── oauth/
│   │   └── rbac/
│   ├── telemetry/
│   │   ├── metrics/
│   │   ├── tracing/
│   │   └── logging/
│   ├── database/
│   │   ├── postgres/
│   │   ├── redis/
│   │   └── migrations/
│   ├── messaging/
│   │   ├── kafka/
│   │   ├── nats/
│   │   └── events/
│   ├── k8s/
│   │   ├── client/
│   │   ├── informers/
│   │   └── controllers/
│   ├── http/
│   │   ├── middleware/
│   │   ├── handlers/
│   │   └── client/
│   ├── grpc/
│   │   ├── interceptors/
│   │   ├── health/
│   │   └── reflection/
│   ├── testing/
│   │   ├── fixtures/
│   │   ├── mocks/
│   │   └── integration/
│   ├── utils/
│   │   ├── retry/
│   │   ├── circuit/
│   │   └── pool/
│   └── errors/
│       ├── types/
│       └── handlers/
│
├── tools/                                # Development tools
│   ├── dev-env/
│   │   ├── docker-compose.yml
│   │   ├── docker-compose.override.yml
│   │   ├── kind-config.yaml
│   │   ├── setup.sh
│   │   └── teardown.sh
│   ├── generators/
│   │   ├── service-generator/
│   │   ├── api-generator/
│   │   └── test-generator/
│   ├── linters/
│   │   ├── .golangci.yml
│   │   ├── .eslintrc.js
│   │   └── .prettierrc
│   ├── analyzers/
│   │   ├── dependency-check.sh
│   │   ├── security-scan.sh
│   │   └── performance-test.sh
│   └── migration/
│       ├── db-migrate.sh
│       └── data-transform.sh
│
├── projects/                             # Independent micro-projects
│   ├── platform-api/                     # Core API Service
│   ├── control-plane/                    # Control Plane Service
│   ├── telemetry-collector/             # Custom OTel Collector
│   ├── experiment-controller/           # K8s Experiment Controller
│   ├── pipeline-operator/               # Pipeline CRD Operator
│   ├── config-service/                  # Configuration Management
│   ├── analytics-engine/                # Analytics & ML Service
│   ├── web-dashboard/                   # React Dashboard
│   ├── mobile-app/                      # React Native App
│   ├── cli/                             # Phoenix CLI
│   ├── sdk-go/                          # Go SDK
│   ├── sdk-python/                      # Python SDK
│   ├── sdk-js/                          # JavaScript SDK
│   ├── load-generator/                  # Load Testing Tool
│   └── docs-site/                       # Documentation Website
│
├── tests/                               # Cross-project tests
│   ├── integration/
│   │   ├── scenarios/
│   │   ├── fixtures/
│   │   └── runner/
│   ├── e2e/
│   │   ├── flows/
│   │   ├── pages/
│   │   └── utils/
│   ├── performance/
│   │   ├── benchmarks/
│   │   ├── load/
│   │   └── stress/
│   ├── security/
│   │   ├── penetration/
│   │   └── vulnerability/
│   └── contracts/
│       ├── api/
│       └── events/
│
├── docs/                                # Project-wide documentation
│   ├── architecture/
│   │   ├── decisions/                   # ADRs
│   │   ├── diagrams/
│   │   └── patterns/
│   ├── api/
│   │   ├── rest/
│   │   ├── grpc/
│   │   └── graphql/
│   ├── guides/
│   │   ├── developer/
│   │   ├── operator/
│   │   └── user/
│   ├── runbooks/
│   │   ├── incident-response/
│   │   ├── deployment/
│   │   └── maintenance/
│   └── standards/
│       ├── coding/
│       ├── security/
│       └── testing/
│
├── configs/                             # Configuration files
│   ├── development/
│   ├── staging/
│   └── production/
│
├── scripts/                             # Root-level scripts
│   ├── setup-workspace.sh
│   ├── validate-all.sh
│   └── release-platform.sh
│
├── .gitignore
├── .gitattributes
├── .editorconfig
├── go.work                              # Go workspace
├── go.work.sum
├── Makefile                             # Root Makefile
├── docker-compose.yml                   # Development stack
├── docker-compose.prod.yml              # Production stack
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
├── SECURITY.md
└── README.md
```

## Project Architecture

### Standard Project Structure

Each project under `projects/` follows this structure:

```
projects/<project-name>/
├── cmd/                                 # Application entrypoints
│   └── <app-name>/
│       └── main.go
├── internal/                            # Private application code
│   ├── api/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── routes/
│   │   ├── grpc/
│   │   │   ├── services/
│   │   │   └── interceptors/
│   │   └── graphql/
│   │       ├── resolvers/
│   │       └── schemas/
│   ├── domain/
│   │   ├── entities/
│   │   ├── repositories/
│   │   ├── services/
│   │   └── events/
│   ├── infrastructure/
│   │   ├── database/
│   │   ├── cache/
│   │   ├── messaging/
│   │   └── external/
│   ├── config/
│   │   ├── config.go
│   │   └── validation.go
│   └── utils/
├── pkg/                                 # Public packages (if any)
├── api/                                 # API definitions
│   ├── proto/                           # Protocol buffers
│   ├── openapi/                         # OpenAPI specs
│   └── graphql/                         # GraphQL schemas
├── migrations/                          # Database migrations
│   ├── postgres/
│   └── redis/
├── configs/                             # Configuration files
│   ├── base.yaml
│   ├── development.yaml
│   ├── staging.yaml
│   └── production.yaml
├── build/                               # Build configurations
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   └── docker-compose.yml
├── deployments/                         # Deployment configs
│   ├── k8s/
│   │   ├── base/
│   │   └── overlays/
│   └── helm/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
├── scripts/                             # Project scripts
│   ├── generate.sh
│   ├── migrate.sh
│   └── test.sh
├── tests/                               # Test files
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
├── docs/                                # Project documentation
│   ├── README.md
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── CONTRIBUTING.md
│   └── CHANGELOG.md
├── .gitignore
├── .dockerignore
├── Makefile                             # Project Makefile
├── go.mod                               # Go module
├── go.sum
└── VERSION                              # Semantic version
```

### Example: Platform API Service

```
projects/platform-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   │   ├── experiment_handler.go
│   │   │   │   ├── optimization_handler.go
│   │   │   │   ├── metrics_handler.go
│   │   │   │   └── health_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── cors.go
│   │   │   │   ├── logging.go
│   │   │   │   ├── metrics.go
│   │   │   │   ├── ratelimit.go
│   │   │   │   └── recovery.go
│   │   │   └── routes/
│   │   │       ├── router.go
│   │   │       └── swagger.go
│   │   └── grpc/
│   │       ├── services/
│   │       │   ├── experiment_service.go
│   │       │   └── optimization_service.go
│   │       └── interceptors/
│   │           ├── auth.go
│   │           ├── logging.go
│   │           └── validation.go
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── experiment.go
│   │   │   ├── optimization.go
│   │   │   └── user.go
│   │   ├── repositories/
│   │   │   ├── experiment_repository.go
│   │   │   └── optimization_repository.go
│   │   ├── services/
│   │   │   ├── experiment_service.go
│   │   │   ├── optimization_service.go
│   │   │   └── analytics_service.go
│   │   └── events/
│   │       ├── experiment_events.go
│   │       └── optimization_events.go
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── postgres/
│   │   │   │   ├── connection.go
│   │   │   │   └── repositories/
│   │   │   └── migrations/
│   │   ├── cache/
│   │   │   └── redis/
│   │   ├── messaging/
│   │   │   ├── kafka/
│   │   │   └── nats/
│   │   └── external/
│   │       ├── prometheus/
│   │       └── grafana/
│   └── config/
│       ├── config.go
│       ├── database.go
│       ├── server.go
│       └── validation.go
├── api/
│   ├── proto/
│   │   ├── experiment.proto
│   │   ├── optimization.proto
│   │   └── common.proto
│   └── openapi/
│       └── swagger.yaml
├── migrations/
│   └── postgres/
│       ├── 001_initial_schema.up.sql
│       ├── 001_initial_schema.down.sql
│       ├── 002_add_experiments.up.sql
│       └── 002_add_experiments.down.sql
├── build/
│   ├── Dockerfile
│   └── docker-compose.yml
├── deployments/
│   └── k8s/
│       ├── deployment.yaml
│       ├── service.yaml
│       ├── configmap.yaml
│       └── hpa.yaml
├── Makefile
├── go.mod
├── go.sum
└── VERSION
```

## Development Standards

### Code Organization Principles

1. **Domain-Driven Design**: Clear separation between domain logic and infrastructure
2. **Hexagonal Architecture**: Ports and adapters pattern for flexibility
3. **SOLID Principles**: Single responsibility, open/closed, etc.
4. **12-Factor App**: Environment-based configuration, stateless processes
5. **Clean Code**: Self-documenting code with meaningful names

### Language-Specific Standards

#### Go Standards

```go
// Package comment must be present
// Package api provides HTTP API handlers for the Phoenix platform.
package api

import (
    "context"
    "fmt"
    "time"

    "github.com/phoenix/platform-api/internal/domain"
    "github.com/phoenix/pkg/errors"
    "github.com/phoenix/pkg/logging"
)

// ExperimentHandler handles HTTP requests for experiments.
type ExperimentHandler struct {
    service domain.ExperimentService
    logger  logging.Logger
}

// NewExperimentHandler creates a new experiment handler.
func NewExperimentHandler(service domain.ExperimentService, logger logging.Logger) *ExperimentHandler {
    return &ExperimentHandler{
        service: service,
        logger:  logger,
    }
}

// CreateExperiment handles POST /api/v1/experiments
func (h *ExperimentHandler) CreateExperiment(ctx context.Context, req CreateExperimentRequest) (*ExperimentResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, errors.Wrap(err, "invalid request")
    }

    // Create experiment
    experiment, err := h.service.Create(ctx, domain.CreateExperimentInput{
        Name:        req.Name,
        Description: req.Description,
        Type:        domain.ExperimentType(req.Type),
        Config:      req.Config,
    })
    if err != nil {
        h.logger.Error("failed to create experiment", 
            "error", err,
            "name", req.Name)
        return nil, errors.Wrap(err, "failed to create experiment")
    }

    h.logger.Info("experiment created",
        "id", experiment.ID,
        "name", experiment.Name)

    return toExperimentResponse(experiment), nil
}
```

#### TypeScript/React Standards

```typescript
// src/components/ExperimentDashboard/ExperimentDashboard.tsx

import React, { useState, useEffect, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { 
  Box, 
  Grid, 
  Paper, 
  Typography, 
  Button,
  Skeleton 
} from '@mui/material';

import { RootState } from '@/store';
import { 
  fetchExperiments, 
  selectExperiments, 
  selectExperimentsLoading 
} from '@/store/experiments';
import { ExperimentCard } from './components/ExperimentCard';
import { ExperimentMetrics } from './components/ExperimentMetrics';
import { CreateExperimentDialog } from './components/CreateExperimentDialog';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Experiment } from '@/types';

interface ExperimentDashboardProps {
  onExperimentSelect?: (experiment: Experiment) => void;
}

/**
 * ExperimentDashboard displays and manages optimization experiments.
 */
export const ExperimentDashboard: React.FC<ExperimentDashboardProps> = ({ 
  onExperimentSelect 
}) => {
  const dispatch = useDispatch();
  const experiments = useSelector(selectExperiments);
  const loading = useSelector(selectExperimentsLoading);
  const [selectedExperiment, setSelectedExperiment] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  // WebSocket for real-time updates
  const { subscribe, unsubscribe } = useWebSocket();

  useEffect(() => {
    // Fetch initial experiments
    dispatch(fetchExperiments());

    // Subscribe to real-time updates
    const subscription = subscribe('experiments.*', (event) => {
      if (event.type === 'experiment.updated') {
        dispatch(fetchExperiments());
      }
    });

    return () => {
      unsubscribe(subscription);
    };
  }, [dispatch, subscribe, unsubscribe]);

  const handleExperimentSelect = useCallback((experimentId: string) => {
    setSelectedExperiment(experimentId);
    const experiment = experiments.find(e => e.id === experimentId);
    if (experiment && onExperimentSelect) {
      onExperimentSelect(experiment);
    }
  }, [experiments, onExperimentSelect]);

  const handleCreateExperiment = useCallback(() => {
    setCreateDialogOpen(true);
  }, []);

  if (loading && experiments.length === 0) {
    return <DashboardSkeleton />;
  }

  return (
    <Box sx={{ flexGrow: 1, p: 3 }}>
      <Grid container spacing={3}>
        {/* Header */}
        <Grid item xs={12}>
          <Box display="flex" justifyContent="space-between" alignItems="center">
            <Typography variant="h4" component="h1">
              Optimization Experiments
            </Typography>
            <Button 
              variant="contained" 
              color="primary"
              onClick={handleCreateExperiment}
            >
              Create Experiment
            </Button>
          </Box>
        </Grid>

        {/* Experiments List */}
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 2, height: '70vh', overflow: 'auto' }}>
            <Typography variant="h6" gutterBottom>
              Active Experiments
            </Typography>
            {experiments.map((experiment) => (
              <ExperimentCard
                key={experiment.id}
                experiment={experiment}
                selected={selectedExperiment === experiment.id}
                onSelect={handleExperimentSelect}
              />
            ))}
          </Paper>
        </Grid>

        {/* Metrics Display */}
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 2, height: '70vh' }}>
            {selectedExperiment ? (
              <ExperimentMetrics experimentId={selectedExperiment} />
            ) : (
              <Box 
                display="flex" 
                alignItems="center" 
                justifyContent="center" 
                height="100%"
              >
                <Typography variant="body1" color="text.secondary">
                  Select an experiment to view metrics
                </Typography>
              </Box>
            )}
          </Paper>
        </Grid>
      </Grid>

      {/* Create Dialog */}
      <CreateExperimentDialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        onSuccess={() => {
          setCreateDialogOpen(false);
          dispatch(fetchExperiments());
        }}
      />
    </Box>
  );
};

// Loading skeleton for better UX
const DashboardSkeleton: React.FC = () => (
  <Box sx={{ flexGrow: 1, p: 3 }}>
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Skeleton variant="text" height={40} width={300} />
      </Grid>
      <Grid item xs={12} md={4}>
        <Skeleton variant="rectangular" height="70vh" />
      </Grid>
      <Grid item xs={12} md={8}>
        <Skeleton variant="rectangular" height="70vh" />
      </Grid>
    </Grid>
  </Box>
);
```

### API Design Standards

#### RESTful API

```yaml
# api/openapi/platform-api.yaml
openapi: 3.0.3
info:
  title: Phoenix Platform API
  version: 1.0.0
  description: |
    Phoenix Platform API provides endpoints for managing optimization experiments,
    pipeline configurations, and metrics analysis.
  contact:
    name: Phoenix Team
    email: team@phoenix.io
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0

servers:
  - url: https://api.phoenix.io/v1
    description: Production
  - url: https://staging-api.phoenix.io/v1
    description: Staging
  - url: http://localhost:8080/v1
    description: Development

tags:
  - name: experiments
    description: Experiment management
  - name: optimizations
    description: Optimization configurations
  - name: metrics
    description: Metrics and analytics
  - name: pipelines
    description: Pipeline templates

paths:
  /experiments:
    get:
      tags: [experiments]
      summary: List experiments
      operationId: listExperiments
      parameters:
        - $ref: '#/components/parameters/PageSize'
        - $ref: '#/components/parameters/PageToken'
        - name: status
          in: query
          schema:
            type: string
            enum: [pending, running, completed, failed]
        - name: namespace
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ExperimentList'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalError'

    post:
      tags: [experiments]
      summary: Create experiment
      operationId: createExperiment
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
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '409':
          $ref: '#/components/responses/Conflict'
        '500':
          $ref: '#/components/responses/InternalError'

components:
  schemas:
    Experiment:
      type: object
      required:
        - id
        - name
        - status
        - type
        - createdAt
      properties:
        id:
          type: string
          format: uuid
          example: "123e4567-e89b-12d3-a456-426614174000"
        name:
          type: string
          example: "Production Cost Optimization"
        description:
          type: string
          example: "Reduce metrics cardinality by 50% while maintaining visibility"
        status:
          $ref: '#/components/schemas/ExperimentStatus'
        type:
          $ref: '#/components/schemas/ExperimentType'
        config:
          $ref: '#/components/schemas/ExperimentConfig'
        metrics:
          $ref: '#/components/schemas/ExperimentMetrics'
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
        startedAt:
          type: string
          format: date-time
        completedAt:
          type: string
          format: date-time

    ExperimentStatus:
      type: string
      enum:
        - pending
        - validating
        - deploying
        - running
        - analyzing
        - completed
        - failed
        - cancelled

    ExperimentType:
      type: string
      enum:
        - ab_test
        - canary
        - adaptive
        - scheduled

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

#### gRPC API

```protobuf
// api/proto/experiment.proto
syntax = "proto3";

package phoenix.platform.v1;

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/field_mask.proto";
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

  // ListExperiments lists experiments with filtering
  rpc ListExperiments(ListExperimentsRequest) returns (ListExperimentsResponse) {
    option (google.api.http) = {
      get: "/v1/experiments"
    };
  }

  // UpdateExperiment updates an existing experiment
  rpc UpdateExperiment(UpdateExperimentRequest) returns (Experiment) {
    option (google.api.http) = {
      patch: "/v1/experiments/{experiment.id}"
      body: "experiment"
    };
  }

  // DeleteExperiment deletes an experiment
  rpc DeleteExperiment(DeleteExperimentRequest) returns (DeleteExperimentResponse) {
    option (google.api.http) = {
      delete: "/v1/experiments/{id}"
    };
  }

  // StartExperiment starts an experiment
  rpc StartExperiment(StartExperimentRequest) returns (Experiment) {
    option (google.api.http) = {
      post: "/v1/experiments/{id}:start"
      body: "*"
    };
  }

  // StopExperiment stops a running experiment
  rpc StopExperiment(StopExperimentRequest) returns (Experiment) {
    option (google.api.http) = {
      post: "/v1/experiments/{id}:stop"
      body: "*"
    };
  }

  // GetExperimentMetrics retrieves experiment metrics
  rpc GetExperimentMetrics(GetExperimentMetricsRequest) returns (ExperimentMetrics) {
    option (google.api.http) = {
      get: "/v1/experiments/{id}/metrics"
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

  // Optional description
  string description = 3 [(validate.rules).string.max_len = 500];

  // Current status
  ExperimentStatus status = 4;

  // Experiment type
  ExperimentType type = 5;

  // Namespace where experiment runs
  string namespace = 6 [(validate.rules).string = {
    pattern: "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"
  }];

  // Configuration
  ExperimentConfig config = 7;

  // Current metrics
  ExperimentMetrics metrics = 8;

  // Timestamps
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp updated_at = 10;
  google.protobuf.Timestamp started_at = 11;
  google.protobuf.Timestamp completed_at = 12;

  // User who created the experiment
  string created_by = 13;

  // Labels for filtering
  map<string, string> labels = 14;

  // Annotations for additional metadata
  map<string, string> annotations = 15;
}

// ExperimentStatus represents the current state
enum ExperimentStatus {
  EXPERIMENT_STATUS_UNSPECIFIED = 0;
  EXPERIMENT_STATUS_PENDING = 1;
  EXPERIMENT_STATUS_VALIDATING = 2;
  EXPERIMENT_STATUS_DEPLOYING = 3;
  EXPERIMENT_STATUS_RUNNING = 4;
  EXPERIMENT_STATUS_ANALYZING = 5;
  EXPERIMENT_STATUS_COMPLETED = 6;
  EXPERIMENT_STATUS_FAILED = 7;
  EXPERIMENT_STATUS_CANCELLED = 8;
}

// ExperimentType defines the optimization strategy
enum ExperimentType {
  EXPERIMENT_TYPE_UNSPECIFIED = 0;
  EXPERIMENT_TYPE_AB_TEST = 1;
  EXPERIMENT_TYPE_CANARY = 2;
  EXPERIMENT_TYPE_ADAPTIVE = 3;
  EXPERIMENT_TYPE_SCHEDULED = 4;
}
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
PKG_DIR := $(ROOT_DIR)/pkg
TOOLS_DIR := $(ROOT_DIR)/tools

# Version
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "0.0.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git describe --tags --always --dirty)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Docker
DOCKER_REGISTRY ?= ghcr.io/phoenix
DOCKER_BUILD_ARGS := \
    --build-arg VERSION=$(VERSION) \
    --build-arg GIT_COMMIT=$(GIT_COMMIT) \
    --build-arg BUILD_DATE=$(BUILD_DATE)

# Colors
CYAN := \033[0;36m
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Projects
ALL_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)
GO_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec test -f {}/go.mod \; -print | xargs -n1 basename)
NODE_PROJECTS := $(shell find $(PROJECTS_DIR) -mindepth 1 -maxdepth 1 -type d -exec test -f {}/package.json \; -print | xargs -n1 basename)

# Include shared makefiles
include $(BUILD_DIR)/makefiles/*.mk

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all help clean build test lint fmt security docker release

## General Targets

all: validate build test ## Run validate, build, and test

help: ## Display this help message
	@echo -e "$(CYAN)Phoenix Platform - Monorepo Makefile$(NC)"
	@echo -e "$(CYAN)=====================================$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(CYAN)<target>$(NC)\n\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(CYAN)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo -e "$(GREEN)Project-specific targets:$(NC)"
	@echo -e "  $(CYAN)build-<project>$(NC)  Build specific project"
	@echo -e "  $(CYAN)test-<project>$(NC)   Test specific project"
	@echo -e "  $(CYAN)lint-<project>$(NC)   Lint specific project"
	@echo ""
	@echo -e "$(GREEN)Available projects:$(NC)"
	@for project in $(ALL_PROJECTS); do echo "  - $$project"; done

clean: $(ALL_PROJECTS:%=clean-%) ## Clean all build artifacts
	@echo -e "$(GREEN)✓ All projects cleaned$(NC)"

##@ Development

setup: ## Setup development environment
	@echo -e "$(CYAN)Setting up development environment...$(NC)"
	@$(TOOLS_DIR)/dev-env/setup.sh
	@echo -e "$(GREEN)✓ Development environment ready$(NC)"

dev-up: ## Start development services
	@echo -e "$(CYAN)Starting development services...$(NC)"
	@docker-compose up -d
	@echo -e "$(GREEN)✓ Services started$(NC)"
	@echo -e "$(YELLOW)Services:$(NC)"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000"

dev-down: ## Stop development services
	@echo -e "$(CYAN)Stopping development services...$(NC)"
	@docker-compose down
	@echo -e "$(GREEN)✓ Services stopped$(NC)"

dev-logs: ## Show development service logs
	@docker-compose logs -f

dev-reset: dev-down ## Reset development environment
	@echo -e "$(YELLOW)Removing volumes...$(NC)"
	@docker-compose down -v
	@echo -e "$(GREEN)✓ Development environment reset$(NC)"

##@ Building

build: $(GO_PROJECTS:%=build-%) $(NODE_PROJECTS:%=build-node-%) ## Build all projects
	@echo -e "$(GREEN)✓ All projects built$(NC)"

build-%: ## Build specific project
	@echo -e "$(CYAN)Building $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* build
	@echo -e "$(GREEN)✓ $* built$(NC)"

build-node-%: ## Build Node.js project
	@echo -e "$(CYAN)Building $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* build
	@echo -e "$(GREEN)✓ $* built$(NC)"

build-changed: ## Build only changed projects
	@echo -e "$(CYAN)Building changed projects...$(NC)"
	@$(BUILD_DIR)/scripts/ci/build-changed.sh
	@echo -e "$(GREEN)✓ Changed projects built$(NC)"

##@ Testing

test: $(ALL_PROJECTS:%=test-%) ## Run all tests
	@echo -e "$(GREEN)✓ All tests passed$(NC)"

test-%: ## Test specific project
	@echo -e "$(CYAN)Testing $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* test
	@echo -e "$(GREEN)✓ $* tests passed$(NC)"

test-integration: ## Run integration tests
	@echo -e "$(CYAN)Running integration tests...$(NC)"
	@$(MAKE) -C $(ROOT_DIR)/tests/integration test
	@echo -e "$(GREEN)✓ Integration tests passed$(NC)"

test-e2e: ## Run end-to-end tests
	@echo -e "$(CYAN)Running e2e tests...$(NC)"
	@$(MAKE) -C $(ROOT_DIR)/tests/e2e test
	@echo -e "$(GREEN)✓ E2E tests passed$(NC)"

test-coverage: ## Generate test coverage report
	@echo -e "$(CYAN)Generating coverage report...$(NC)"
	@$(BUILD_DIR)/scripts/ci/coverage.sh
	@echo -e "$(GREEN)✓ Coverage report generated$(NC)"

##@ Code Quality

lint: $(ALL_PROJECTS:%=lint-%) ## Lint all projects
	@echo -e "$(GREEN)✓ All projects linted$(NC)"

lint-%: ## Lint specific project
	@echo -e "$(CYAN)Linting $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* lint
	@echo -e "$(GREEN)✓ $* linted$(NC)"

fmt: $(ALL_PROJECTS:%=fmt-%) ## Format all code
	@echo -e "$(GREEN)✓ All code formatted$(NC)"

fmt-%: ## Format specific project
	@echo -e "$(CYAN)Formatting $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* fmt
	@echo -e "$(GREEN)✓ $* formatted$(NC)"

validate: ## Validate repository structure
	@echo -e "$(CYAN)Validating repository structure...$(NC)"
	@$(BUILD_DIR)/scripts/utils/validate-structure.sh
	@echo -e "$(GREEN)✓ Repository structure valid$(NC)"

##@ Security

security: $(ALL_PROJECTS:%=security-%) ## Run security scans
	@echo -e "$(GREEN)✓ Security scans completed$(NC)"

security-%: ## Security scan specific project
	@echo -e "$(CYAN)Scanning $* for vulnerabilities...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* security
	@echo -e "$(GREEN)✓ $* security scan completed$(NC)"

audit: ## Audit dependencies
	@echo -e "$(CYAN)Auditing dependencies...$(NC)"
	@$(TOOLS_DIR)/analyzers/dependency-check.sh
	@echo -e "$(GREEN)✓ Dependency audit completed$(NC)"

##@ Docker

docker: $(ALL_PROJECTS:%=docker-%) ## Build all Docker images
	@echo -e "$(GREEN)✓ All Docker images built$(NC)"

docker-%: ## Build Docker image for specific project
	@echo -e "$(CYAN)Building Docker image for $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* docker-build
	@echo -e "$(GREEN)✓ $* Docker image built$(NC)"

docker-push: $(ALL_PROJECTS:%=docker-push-%) ## Push all Docker images
	@echo -e "$(GREEN)✓ All Docker images pushed$(NC)"

docker-push-%: ## Push Docker image for specific project
	@echo -e "$(CYAN)Pushing Docker image for $*...$(NC)"
	@$(MAKE) -C $(PROJECTS_DIR)/$* docker-push
	@echo -e "$(GREEN)✓ $* Docker image pushed$(NC)"

##@ Kubernetes

k8s-generate: ## Generate Kubernetes manifests
	@echo -e "$(CYAN)Generating Kubernetes manifests...$(NC)"
	@$(BUILD_DIR)/scripts/k8s/generate-manifests.sh
	@echo -e "$(GREEN)✓ Kubernetes manifests generated$(NC)"

k8s-validate: ## Validate Kubernetes manifests
	@echo -e "$(CYAN)Validating Kubernetes manifests...$(NC)"
	@$(BUILD_DIR)/scripts/k8s/validate-manifests.sh
	@echo -e "$(GREEN)✓ Kubernetes manifests valid$(NC)"

k8s-deploy-dev: ## Deploy to development cluster
	@echo -e "$(CYAN)Deploying to development...$(NC)"
	@$(BUILD_DIR)/scripts/k8s/deploy.sh development
	@echo -e "$(GREEN)✓ Deployed to development$(NC)"

##@ Release

version: ## Display current version
	@echo $(VERSION)

changelog: ## Generate changelog
	@echo -e "$(CYAN)Generating changelog...$(NC)"
	@$(BUILD_DIR)/scripts/release/generate-changelog.sh
	@echo -e "$(GREEN)✓ Changelog generated$(NC)"

release: ## Create a new release
	@echo -e "$(CYAN)Creating release...$(NC)"
	@$(BUILD_DIR)/scripts/release/create-release.sh
	@echo -e "$(GREEN)✓ Release created$(NC)"

release-notes: ## Generate release notes
	@echo -e "$(CYAN)Generating release notes...$(NC)"
	@$(BUILD_DIR)/scripts/release/generate-notes.sh
	@echo -e "$(GREEN)✓ Release notes generated$(NC)"

##@ Utilities

generate: ## Run code generation
	@echo -e "$(CYAN)Running code generation...$(NC)"
	@$(MAKE) -C $(PKG_DIR) generate
	@for project in $(GO_PROJECTS); do \
		$(MAKE) -C $(PROJECTS_DIR)/$$project generate 2>/dev/null || true; \
	done
	@echo -e "$(GREEN)✓ Code generation completed$(NC)"

deps: ## Update dependencies
	@echo -e "$(CYAN)Updating dependencies...$(NC)"
	@go work sync
	@for project in $(GO_PROJECTS); do \
		echo -e "$(CYAN)Updating $$project dependencies...$(NC)"; \
		cd $(PROJECTS_DIR)/$$project && go mod tidy; \
	done
	@echo -e "$(GREEN)✓ Dependencies updated$(NC)"

tools: ## Install development tools
	@echo -e "$(CYAN)Installing development tools...$(NC)"
	@$(TOOLS_DIR)/install-tools.sh
	@echo -e "$(GREEN)✓ Development tools installed$(NC)"

# Project-specific targets
$(foreach project,$(ALL_PROJECTS),$(eval $(call PROJECT_TARGET,$(project))))
```

[Content continues...]