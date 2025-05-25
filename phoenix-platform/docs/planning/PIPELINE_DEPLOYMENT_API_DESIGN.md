# Pipeline Deployment API Design

## Overview
This document outlines the design for adding direct pipeline deployment capabilities to Phoenix, addressing a key functional gap identified in the platform review. This feature enables users to deploy optimized pipelines without requiring an A/B experiment.

## Use Cases

1. **Production Rollout**: After validating an optimization through experiments, deploy it broadly
2. **Known Good Configurations**: Deploy proven pipelines to new environments
3. **Emergency Rollback**: Quickly deploy a baseline pipeline if issues arise
4. **Template Application**: Apply standard configurations across multiple clusters

## API Design

### New Endpoints

#### 1. Deploy Pipeline
```http
POST /api/v1/pipelines/deployments
Content-Type: application/json
Authorization: Bearer <token>

{
  "pipeline_name": "process-topk-v1",
  "deployment_name": "production-topk-deployment",
  "target_nodes": {
    "environment": "production",
    "tier": "frontend"
  },
  "namespace": "phoenix-prod",
  "parameters": {
    "top_k": 20,
    "critical_processes": ["nginx", "envoy", "app-server"]
  },
  "replicas": 3,
  "resources": {
    "requests": {
      "cpu": "100m",
      "memory": "128Mi"
    },
    "limits": {
      "cpu": "500m",
      "memory": "512Mi"
    }
  }
}

Response:
{
  "deployment_id": "dep-abc123",
  "status": "deploying",
  "pipeline_name": "process-topk-v1",
  "deployment_name": "production-topk-deployment",
  "created_at": "2024-01-15T10:00:00Z",
  "target_nodes": {
    "environment": "production",
    "tier": "frontend"
  }
}
```

#### 2. List Pipeline Deployments
```http
GET /api/v1/pipelines/deployments?namespace=phoenix-prod&status=active

Response:
{
  "deployments": [
    {
      "deployment_id": "dep-abc123",
      "deployment_name": "production-topk-deployment",
      "pipeline_name": "process-topk-v1",
      "status": "active",
      "namespace": "phoenix-prod",
      "target_nodes": {
        "environment": "production",
        "tier": "frontend"
      },
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:05:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}
```

#### 3. Get Deployment Status
```http
GET /api/v1/pipelines/deployments/{deployment_id}

Response:
{
  "deployment_id": "dep-abc123",
  "deployment_name": "production-topk-deployment",
  "pipeline_name": "process-topk-v1",
  "status": "active",
  "phase": "running",
  "namespace": "phoenix-prod",
  "target_nodes": {
    "environment": "production",
    "tier": "frontend"
  },
  "instances": {
    "desired": 15,
    "ready": 15,
    "updated": 15
  },
  "metrics": {
    "cardinality": 12500,
    "throughput": "1.2M/min",
    "error_rate": 0.001
  },
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:05:00Z"
}
```

#### 4. Update Deployment
```http
PATCH /api/v1/pipelines/deployments/{deployment_id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "parameters": {
    "top_k": 25
  },
  "resources": {
    "limits": {
      "memory": "1Gi"
    }
  }
}

Response:
{
  "deployment_id": "dep-abc123",
  "status": "updating",
  "message": "Rolling update initiated"
}
```

#### 5. Delete Deployment
```http
DELETE /api/v1/pipelines/deployments/{deployment_id}

Response:
{
  "deployment_id": "dep-abc123",
  "status": "deleting",
  "message": "Deployment removal initiated"
}
```

## Implementation Details

### Backend Changes

1. **New Service Layer**
   ```go
   // pkg/api/pipeline_deployment_service.go
   type PipelineDeploymentService interface {
       Create(ctx context.Context, req *CreateDeploymentRequest) (*Deployment, error)
       List(ctx context.Context, filters *ListFilters) (*DeploymentList, error)
       Get(ctx context.Context, deploymentID string) (*Deployment, error)
       Update(ctx context.Context, deploymentID string, updates *UpdateRequest) error
       Delete(ctx context.Context, deploymentID string) error
   }
   ```

2. **Database Schema**
   ```sql
   -- migrations/005_create_pipeline_deployments.sql
   CREATE TABLE pipeline_deployments (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       deployment_name VARCHAR(255) NOT NULL,
       pipeline_name VARCHAR(255) NOT NULL,
       namespace VARCHAR(255) NOT NULL,
       target_nodes JSONB NOT NULL,
       parameters JSONB,
       resources JSONB,
       status VARCHAR(50) NOT NULL,
       phase VARCHAR(50),
       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
       deleted_at TIMESTAMP,
       UNIQUE(deployment_name, namespace)
   );
   
   CREATE INDEX idx_deployments_namespace ON pipeline_deployments(namespace);
   CREATE INDEX idx_deployments_status ON pipeline_deployments(status);
   ```

3. **Integration with Pipeline Operator**
   - Create PhoenixProcessPipeline CR with `variant: "production"`
   - Add new label `phoenix.io/deployment-type: "direct"`
   - Pipeline operator recognizes and handles direct deployments differently

### CLI Integration

```bash
# Deploy a pipeline directly
phoenix pipeline deploy \
  --name process-topk-v1 \
  --deployment-name prod-optimization \
  --namespace phoenix-prod \
  --selector "environment=production,tier=frontend" \
  --param top_k=20 \
  --param critical_processes=nginx,envoy

# List deployments
phoenix pipeline list-deployments --namespace phoenix-prod

# Get deployment status
phoenix pipeline get-deployment dep-abc123

# Update deployment
phoenix pipeline update-deployment dep-abc123 --param top_k=25

# Delete deployment
phoenix pipeline delete-deployment dep-abc123
```

### UI Integration

1. **New Pipeline Deployments Page**
   - List all active deployments
   - Show deployment status and metrics
   - Quick actions: Update, Delete, View Details

2. **Deploy from Pipeline Library**
   - Add "Deploy" button to each pipeline template
   - Modal for deployment configuration
   - Real-time deployment progress

3. **Post-Experiment Workflow**
   - After successful experiment, show "Deploy to Production" option
   - Pre-fill configuration from winning variant
   - Allow scope expansion (e.g., from staging to all environments)

## Migration Path

### From Experiments to Direct Deployments

1. **Promotion Flow Enhancement**
   ```
   Current: Experiment → Promote (limited to experiment scope)
   New:     Experiment → Promote → Deploy Broadly (optional)
   ```

2. **Backwards Compatibility**
   - Existing experiment promotions continue to work
   - New deployments are tracked separately
   - Can query all pipelines (experiment + direct)

### Cleanup and Management

1. **Automatic Cleanup**
   - Deployments marked for deletion after 30 days
   - Orphaned ConfigMaps cleaned up
   - Old Pipeline CRs garbage collected

2. **Deployment History**
   - Track all deployment changes
   - Audit log for compliance
   - Rollback capabilities

## Security Considerations

1. **RBAC for Deployments**
   - Separate permission for direct deployments
   - Namespace-scoped access control
   - Audit logging for all operations

2. **Validation**
   - Pipeline compatibility checks
   - Resource limit enforcement
   - Target node validation

## Monitoring and Observability

1. **Metrics**
   - `phoenix_deployment_total` - Total deployments by status
   - `phoenix_deployment_duration_seconds` - Deployment time
   - `phoenix_deployment_errors_total` - Failed deployments

2. **Events**
   - Kubernetes events for deployment lifecycle
   - Audit events for configuration changes
   - Alert on deployment failures

## Testing Strategy

1. **Unit Tests**
   - Service layer logic
   - Validation rules
   - Database operations

2. **Integration Tests**
   - API endpoint testing
   - Pipeline operator integration
   - End-to-end deployment flow

3. **E2E Tests**
   - Deploy via API
   - Verify collector running
   - Check metrics flow
   - Clean up resources

## Rollout Plan

### Phase 1: API Implementation (Week 1)
- Implement service layer
- Add database migrations
- Create API endpoints
- Unit tests

### Phase 2: Operator Integration (Week 1-2)
- Modify Pipeline CRD handling
- Add deployment tracking
- Integration tests

### Phase 3: UI Integration (Week 2)
- Add deployments page
- Integrate with pipeline library
- Update experiment promotion flow

### Phase 4: CLI Commands (Week 2-3)
- Add pipeline deploy commands
- Testing and documentation
- Release

## Success Metrics

1. **Adoption**: 30% of experiments result in direct deployments
2. **Reliability**: 99.9% successful deployments
3. **Performance**: Deployment completes in <2 minutes
4. **User Satisfaction**: Reduced time from experiment to production by 50%

## Open Questions

1. Should we support canary deployments (gradual rollout)?
2. Do we need deployment templates for common scenarios?
3. Should deployments support auto-rollback on errors?
4. How do we handle multi-cluster deployments?

---

This design addresses the functional gap of deploying pipelines outside of experiments, providing users with flexibility to operationalize their optimizations efficiently.