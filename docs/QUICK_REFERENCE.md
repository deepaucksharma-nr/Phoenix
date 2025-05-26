# Phoenix Platform Quick Reference

## 🚀 Essential Commands

```bash
# Development Setup
make setup                                    # First-time setup
make dev-up                                  # Start dev services
make dev-down                                # Stop dev services

# Building & Testing
make build                                   # Build all projects
make test                                    # Run all tests
make validate                                # Validate structure
make build-<project>                         # Build specific project
make test-<project>                          # Test specific project

# Validation Tools
./tools/analyzers/boundary-check.sh          # Check architecture boundaries
./tools/analyzers/llm-safety-check.sh        # AI safety validation
./scripts/validate-migration.sh              # Migration validation

# Demo & Examples
./scripts/run-e2e-demo.sh                    # Run E2E demo
./scripts/quick-start.sh                     # Quick start script
```

## 📁 Directory Structure

```
phoenix/
├── projects/         # 12 independent services
├── pkg/             # Shared Go packages (strict review)
├── configs/         # All configuration files
├── deployments/     # K8s, Helm, Terraform
├── tools/           # Dev tools and analyzers
├── tests/           # Integration and E2E tests
├── docs/            # Documentation
├── scripts/         # Utility scripts
└── go.work         # Go workspace file
```

## 🏗️ Key Architecture Rules

1. **NO Cross-Project Imports** ❌
   - Projects CANNOT import from other projects
   - Only import from `/pkg/*`

2. **Standard Project Structure** ✅
   ```
   projects/<name>/
   ├── cmd/           # Entry points
   ├── internal/      # Private code
   ├── api/           # API definitions
   ├── build/         # Docker/build configs
   ├── deployments/   # K8s manifests
   └── README.md      # Documentation
   ```

3. **Shared Package Usage** ✅
   ```go
   import "github.com/phoenix/platform/pkg/auth"     // ✅ Correct
   import "github.com/phoenix/platform/projects/api" // ❌ Wrong!
   ```

## 📚 Documentation Map

| Need | Document | Path |
|------|----------|------|
| Get Started | README | `README.md` |
| Architecture | Platform Architecture | `PHOENIX_PLATFORM_ARCHITECTURE.md` |
| AI Help | Claude Guide | `CLAUDE.md` |
| Contributing | Contribution Guide | `CONTRIBUTING.md` |
| Boundaries | Monorepo Rules | `MONOREPO_BOUNDARIES.md` |
| Migration | Complete Guide | `MIGRATION_COMPLETE_GUIDE.md` |
| Demo | E2E Demo | `E2E_DEMO_GUIDE.md` |

## 🔧 Service Overview

### Core Services
- **platform-api** - Main API service (`:8080`)
- **controller** - Experiment orchestration (`:8081`)
- **dashboard** - Web UI (`:3000`)
- **generator** - Pipeline generation (`:8082`)

### Analytics Services
- **analytics** - Data analysis
- **anomaly-detector** - Anomaly detection
- **benchmark** - Performance testing

### Infrastructure
- **phoenix-cli** - Command-line tool
- **pipeline-operator** - K8s operator
- **loadsim-operator** - Load testing operator

## ⚙️ Configuration Locations

```
configs/
├── control/          # Control plane policies
├── monitoring/       # Prometheus/Grafana
├── otel/            # OpenTelemetry
└── production/      # Production configs
```

## 🧪 Testing

```bash
# Unit Tests
make test                    # All tests
make test-<project>         # Specific project

# Integration Tests
cd tests/integration && go test -v

# E2E Tests
./scripts/run-e2e-demo.sh

# Performance Tests
cd tests/performance && make bench
```

## 🚨 Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Cross-project import | Move shared code to `/pkg` |
| Build fails | Run `go work sync` |
| Test fails | Check service dependencies |
| Validation fails | Run boundary check tool |

## 🔐 Security Checklist

- [ ] No hardcoded secrets
- [ ] No direct DB imports (use `/pkg/database`)
- [ ] Run security scan before commit
- [ ] Follow CODEOWNERS for reviews
- [ ] Check with AI safety tool

## 📋 Pre-Commit Checklist

```bash
# Before committing:
1. make fmt              # Format code
2. make lint             # Lint code
3. make test             # Run tests
4. make validate         # Validate structure
5. Update documentation  # If needed
```

## 🌟 Best Practices

1. **One Purpose Per Project** - Keep projects focused
2. **Document Everything** - Update README.md files
3. **Test First** - Write tests before code
4. **Validate Often** - Run validation tools
5. **Use Shared Packages** - Don't duplicate code

## 🔗 Important Links

- Main Docs: [`CONSOLIDATED_DOCUMENTATION.md`](./CONSOLIDATED_DOCUMENTATION.md)
- Doc Index: [`DOCUMENTATION_INDEX.md`](./DOCUMENTATION_INDEX.md)
- Doc Map: [`DOCUMENTATION_MAP.md`](./DOCUMENTATION_MAP.md)
- Summary: [`DOCUMENTATION_SUMMARY.md`](./DOCUMENTATION_SUMMARY.md)

---
*Keep this reference handy for daily Phoenix Platform development!*