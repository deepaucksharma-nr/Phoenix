# Phoenix Platform Documentation Index

## 📁 Documentation Structure

```
docs/
├── architecture/        # System architecture and design
├── guides/             # How-to guides and tutorials  
├── migration/          # Migration documentation and history
├── operations/         # Operational procedures
├── reference/          # API and technical references
└── generated/          # Auto-generated documentation
```

## 🗂️ Documentation by Category

### Getting Started
- **[README.md](./README.md)** - Project overview and quick start
- **[TEAM_ONBOARDING.md](./TEAM_ONBOARDING.md)** - 5-minute onboarding guide
- **[E2E_DEMO_GUIDE.md](./E2E_DEMO_GUIDE.md)** - End-to-end demo walkthrough

### Architecture & Design
- **[docs/architecture/PLATFORM_ARCHITECTURE.md](./docs/architecture/PLATFORM_ARCHITECTURE.md)** - Complete platform architecture
- **[MONOREPO_BOUNDARIES.md](./MONOREPO_BOUNDARIES.md)** - Monorepo structure and rules
- **[docs/INTERFACE_CONTRACTS.md](./docs/INTERFACE_CONTRACTS.md)** - Service interface definitions

### Development
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines
- **[CLAUDE.md](./CLAUDE.md)** - AI assistant usage guidelines
- **[docs/guides/](./docs/guides/)** - Development guides

### Operations
- **[docs/operations/SERVICE_CONSOLIDATION_PLAN.md](./docs/operations/SERVICE_CONSOLIDATION_PLAN.md)** - Service consolidation
- **[docs/generated/SERVICE_INVENTORY.md](./docs/generated/SERVICE_INVENTORY.md)** - Current service inventory
- **[docs/ROLLBACK_PLAN.md](./docs/ROLLBACK_PLAN.md)** - Emergency rollback procedures

### Migration History
- **[docs/migration/MIGRATION_SUMMARY_CONSOLIDATED.md](./docs/migration/MIGRATION_SUMMARY_CONSOLIDATED.md)** - Migration summary
- **[docs/migration/](./docs/migration/)** - Detailed migration documentation

### Project Documentation
Each project has its own README:
- [API Gateway](./projects/api/README.md)
- [Controller](./projects/controller/README.md) 
- [Dashboard](./projects/dashboard/README.md)
- [Phoenix CLI](./projects/phoenix-cli/README.md)
- [And more...](./projects/)

### Configuration Documentation
- [Control Configs](./configs/control/README.md)
- [Monitoring Configs](./configs/monitoring/README.md)
- [OTEL Configs](./configs/otel/README.md)
- [Production Configs](./configs/production/README.md)

## 🔍 Quick Links

### For Developers
1. Start here: [TEAM_ONBOARDING.md](./TEAM_ONBOARDING.md)
2. Understand architecture: [PLATFORM_ARCHITECTURE.md](./docs/architecture/PLATFORM_ARCHITECTURE.md)
3. Contribute: [CONTRIBUTING.md](./CONTRIBUTING.md)

### For Operators
1. Service overview: [SERVICE_INVENTORY.md](./docs/generated/SERVICE_INVENTORY.md)
2. Deployment: See [infrastructure/](./infrastructure/)
3. Monitoring: See [monitoring/](./monitoring/)

### For AI Assistants
1. Read first: [CLAUDE.md](./CLAUDE.md)
2. Understand boundaries: [MONOREPO_BOUNDARIES.md](./MONOREPO_BOUNDARIES.md)
3. Check validation: Run `./tools/analyzers/boundary-check.sh`

## 📋 Documentation Standards

- **Markdown format** for all documentation
- **Clear headings** with emoji indicators
- **Code examples** where applicable
- **Updated regularly** with each major change
- **Reviewed** as part of PR process

---

*Last updated: May 2025 - Post-migration consolidation*