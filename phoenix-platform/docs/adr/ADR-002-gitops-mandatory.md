# ADR-002: GitOps is Mandatory for All Deployments

## Status
Accepted

## Context
Configuration management and deployment strategies need to be auditable, repeatable, and rollback-capable. Teams often use kubectl directly for quick deployments, which can lead to configuration drift.

## Decision
ALL Phoenix platform deployments MUST use GitOps via ArgoCD. Direct kubectl apply is forbidden except for initial CRD installation.

## Rationale
1. **Auditability**: Every configuration change is tracked in Git
2. **Rollback**: Easy reversion to previous configurations
3. **Review Process**: Changes go through PR review
4. **Consistency**: Ensures all environments match Git state
5. **Security**: No direct cluster access needed for deployments

## Implementation
```yaml
# All configurations stored in Git repository
phoenix-configs/
├── experiments/
│   ├── exp-001/
│   │   ├── pipeline-baseline.yaml
│   │   ├── pipeline-candidate.yaml
│   │   └── kustomization.yaml
│   └── exp-002/
└── argocd/
    └── applications/
```

- ArgoCD syncs from Git repository
- Automated sync on Git push
- Manual sync for production

## Consequences
### Positive
- Complete audit trail
- Prevents configuration drift
- Enables compliance requirements
- Supports multi-environment deployments

### Negative
- Slightly slower deployment process
- Requires Git repository setup
- Learning curve for GitOps

## Enforcement
- API service creates Git commits, never applies directly
- No kubectl permissions for service accounts
- Monitoring alerts on direct modifications

## Alternatives Considered
1. **Direct Deployment**: No audit trail, prone to drift
2. **Helm Only**: Still allows direct applies
3. **Flux**: ArgoCD has better UI for our use case

## References
- GitOps workflow in architecture.md
- Deployment procedures in DEPLOYMENT.md