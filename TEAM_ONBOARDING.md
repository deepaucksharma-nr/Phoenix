# Phoenix Platform - Team Onboarding Guide

Welcome to the new Phoenix Platform monorepo structure! This guide will help you get up and running quickly.

## 🚀 Quick Start (5 minutes)

1. **Pull the latest changes**
   ```bash
   git pull origin main
   ```

2. **Run the quick start script**
   ```bash
   ./scripts/quick-start.sh
   ```

3. **Start developing**
   ```bash
   make dev     # Start all services locally
   make test    # Run tests
   make build   # Build all services
   ```

## 📁 What Changed?

### Old Structure → New Structure
```
OLD:                          NEW:
phoenix/                      phoenix/
├── apps/                     ├── projects/        # All services here
├── services/                 ├── packages/        # Shared code only
├── pkg/                      ├── deployments/     # K8s, Helm, etc
├── phoenix-platform/         ├── scripts/         # Dev tools
└── (mixed structure)         └── tools/           # Analyzers
```

### Key Changes:
- ✅ All services now in `projects/` directory
- ✅ Shared code in `packages/go-common` and `packages/contracts`
- ✅ No more cross-service imports allowed
- ✅ Each service is independently deployable

## 🛠️ Development Workflow

### 1. Working on a Service
```bash
# Navigate to your service
cd projects/analytics

# Run the service locally
make run

# Run tests
make test

# Build Docker image
make docker
```

### 2. Adding Dependencies
```bash
# Always use go work for dependencies
cd projects/my-service
go get github.com/some/package
go mod tidy
go work sync  # Important!
```

### 3. Importing Shared Code
```go
// ✅ CORRECT - Import from packages
import (
    "github.com/phoenix/platform/packages/go-common/logger"
    "github.com/phoenix/platform/packages/contracts/api"
)

// ❌ WRONG - Never import from other projects
import (
    "github.com/phoenix/platform/projects/api/internal/utils"  // FORBIDDEN!
)
```

## 🔍 Validation & Testing

### Before Committing
The pre-commit hook automatically runs validation. You can also run manually:

```bash
# Check import boundaries
./scripts/validate-boundaries.sh

# Run all validations
make validate

# Format code
make fmt
```

### Testing
```bash
# Test everything
make test

# Test specific project
make test-analytics

# Run integration tests
make test-integration
```

## 📦 Building & Deployment

### Local Development
```bash
# Start all services with docker-compose
make dev-up

# Stop all services
make dev-down

# View logs
make logs
```

### Kubernetes Deployment
```bash
# Deploy to development
./scripts/deploy-dev.sh

# Deploy specific service
./scripts/deploy-dev.sh analytics
```

## 🚨 Common Issues & Solutions

### Issue: Import violation error
**Solution**: Move shared code to `packages/go-common` or duplicate if service-specific

### Issue: go.work out of sync
**Solution**: Run `go work sync`

### Issue: Module not found
**Solution**: Check if module is in go.work, run `go mod tidy`

### Issue: Old imports still present
**Solution**: Run `./scripts/update-imports.sh`

## 📚 Documentation

- **README.md** - Project overview and structure
- **CLAUDE.md** - AI assistance guidelines (for Claude.ai)
- **MIGRATION_SUMMARY.md** - Detailed migration notes
- **E2E_DEMO_GUIDE.md** - End-to-end testing guide

## 🤝 Best Practices

1. **Keep Services Independent**
   - No cross-project imports
   - Communicate via APIs only
   - Share code through packages/

2. **Use the Tools**
   - Run `make validate` before pushing
   - Use `./scripts/quick-start.sh` for setup
   - Check `make help` for all commands

3. **Follow Conventions**
   - Each project has standard structure
   - Use provided Makefiles
   - Follow import rules

## 🆘 Getting Help

1. **Check Documentation**
   - This guide
   - README files in each project
   - Scripts have --help flags

2. **Run Diagnostics**
   ```bash
   ./scripts/verify-migration.sh
   ```

3. **Common Commands**
   ```bash
   make help              # Show all available commands
   go work sync          # Fix workspace issues
   ./scripts/validate-boundaries.sh  # Check imports
   ```

## 🎯 Next Steps

1. Pull latest changes
2. Run quick-start script
3. Explore the new structure
4. Start developing!

Welcome to the new Phoenix Platform! 🚀