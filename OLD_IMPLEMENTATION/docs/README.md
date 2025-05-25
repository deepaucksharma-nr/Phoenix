# Phoenix Repository Documentation Index

This repository contains the Phoenix Observability Platform and related governance documentation.

## Repository Structure

```
/
├── CLAUDE.md                    # AI assistant guidance (special exception)
├── docs/                        # Repository-level governance
│   ├── DOCUMENTATION_GOVERNANCE.md
│   ├── GOVERNANCE_ENFORCEMENT.md
│   ├── MONO_REPO_GOVERNANCE.md
│   └── STATIC_ANALYSIS_RULES.md
│
└── phoenix-platform/            # Phoenix platform implementation
    ├── README.md               # Platform overview and quick start
    └── docs/                   # Platform documentation
        ├── README.md           # Documentation index
        ├── Getting Started/
        ├── Development/
        ├── Technical Specs/
        ├── Planning/
        └── Reviews/
```

## Documentation Organization

### Repository Governance (`/docs/`)
- **[Documentation Governance](DOCUMENTATION_GOVERNANCE.md)** - Rules for documentation placement
- **[Governance Enforcement](GOVERNANCE_ENFORCEMENT.md)** - How governance is enforced
- **[Mono-repo Governance](MONO_REPO_GOVERNANCE.md)** - Repository structure rules
- **[Static Analysis Rules](STATIC_ANALYSIS_RULES.md)** - Code quality standards

### Phoenix Platform (`/phoenix-platform/docs/`)
All Phoenix-specific documentation is under `phoenix-platform/docs/`. See the [Platform Documentation Index](../phoenix-platform/docs/README.md) for:
- Getting started guides
- API documentation
- Development guides
- Technical specifications
- Implementation status
- Planning documents

## Key Principles

1. **Documentation Placement**: 
   - Phoenix docs → `phoenix-platform/docs/`
   - Governance docs → `/docs/`
   - Service docs → `<service>/docs/`

2. **File Naming**: 
   - Use UPPERCASE_WITH_UNDERSCORES for all .md files
   - Exception: README.md (GitHub standard)

3. **No Root Documentation**: 
   - Only CLAUDE.md allowed at repository root
   - All other docs must be in appropriate subdirectories

## Quick Links

- [Phoenix Platform Overview](../phoenix-platform/README.md)
- [Phoenix Documentation](../phoenix-platform/docs/README.md)
- [Developer Quick Start](../phoenix-platform/docs/DEVELOPER_QUICK_START.md)
- [API Reference](../phoenix-platform/docs/API_REFERENCE.md)
- [Architecture Overview](../phoenix-platform/docs/ARCHITECTURE.md)

## Contributing

When adding new documentation:
1. Review [Documentation Governance](DOCUMENTATION_GOVERNANCE.md)
2. Place files in correct location
3. Follow naming conventions
4. Update relevant index files