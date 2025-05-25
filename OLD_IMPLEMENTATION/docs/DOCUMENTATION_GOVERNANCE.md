# Phoenix Documentation Governance

## Overview

This document establishes strict rules for documentation placement to maintain mono-repo structure integrity and prevent documentation sprawl.

## Documentation Hierarchy Rules

### 1. Root Level (`/`)

**ONLY these files are allowed at root:**
- `CLAUDE.md` - AI assistant guidance (special case)
- `.gitignore` - Git configuration
- `LICENSE` - Legal requirements
- `.github/` - GitHub specific configs

**FORBIDDEN at root:**
- Project-specific documentation
- Review documents
- Planning documents
- Technical specifications
- Status reports

### 2. Governance Documentation (`/docs/`)

**This directory contains:**
- Cross-repository governance rules
- Static analysis rules
- Documentation standards
- Repository-wide policies

**Examples:**
- `MONO_REPO_GOVERNANCE.md`
- `STATIC_ANALYSIS_RULES.md`
- `DOCUMENTATION_GOVERNANCE.md` (this file)

### 3. Phoenix Platform Documentation (`/phoenix-platform/docs/`)

**All Phoenix-specific documentation MUST go here:**

```
phoenix-platform/docs/
├── README.md                    # Documentation index
├── architecture.md             # System architecture
├── user-guide.md              # End-user documentation
├── api-reference.md           # API documentation
├── troubleshooting.md         # Support documentation
├── QUICK_START_GUIDE.md       # Developer onboarding
│
├── technical-specs/           # Component specifications
│   ├── TECHNICAL_SPEC_MASTER.md
│   ├── TECHNICAL_SPEC_API_SERVICE.md
│   ├── TECHNICAL_SPEC_DASHBOARD.md
│   ├── TECHNICAL_SPEC_EXPERIMENT_CONTROLLER.md
│   ├── TECHNICAL_SPEC_PIPELINE_OPERATOR.md
│   └── TECHNICAL_SPEC_PROCESS_SIMULATOR.md
│
├── planning/                  # Project planning documents
│   ├── PRODUCT_REQUIREMENTS.md
│   ├── IMPLEMENTATION_ROADMAP.md
│   └── PROJECT_STATUS.md
│
├── reviews/                   # Analysis and review documents
│   ├── PHOENIX_DOCUMENTATION_REVIEW.md
│   └── COMPREHENSIVE_REVIEW_SUMMARY.md
│
├── operations/               # Operational documentation
│   ├── DEPLOYMENT.md
│   ├── DEVELOPMENT.md
│   └── runbooks/
│
└── guides/                   # How-to guides
    ├── pipeline-guide.md
    └── experiment-guide.md
```

### 4. Service-Specific Documentation

Each service can have its own README and docs:

```
phoenix-platform/cmd/api/
├── README.md              # Service-specific readme
└── docs/                  # Service-specific docs only
    └── api-internals.md

phoenix-platform/dashboard/
├── README.md              # Frontend-specific readme
└── docs/                  # Frontend-specific docs
    └── component-guide.md
```

## Enforcement Rules

### 1. Pre-commit Checks

Add to `.pre-commit-config.yaml`:
```yaml
- id: doc-location-check
  name: Documentation Location Check
  entry: scripts/check-doc-location.sh
  language: script
  files: '\.md$'
```

### 2. CI/CD Validation

```bash
#!/bin/bash
# scripts/check-doc-location.sh

# Check for MD files at root (except allowed)
root_files=$(find . -maxdepth 1 -name "*.md" | grep -v -E "(CLAUDE\.md|README\.md)")
if [ -n "$root_files" ]; then
    echo "ERROR: Documentation files found at root level:"
    echo "$root_files"
    echo "Move them to appropriate directories under phoenix-platform/docs/"
    exit 1
fi
```

### 3. Documentation Review Checklist

Before creating ANY documentation:

1. **Is it Phoenix-specific?** → Goes in `/phoenix-platform/docs/`
2. **Is it repo-wide governance?** → Goes in `/docs/`
3. **Is it service-specific?** → Goes in service's `docs/` directory
4. **Is it temporary?** → Don't commit it

## Common Mistakes to Avoid

### ❌ DON'T DO THIS:
```
/IMPLEMENTATION_PLAN.md
/PROJECT_STATUS.md
/REVIEW_SUMMARY.md
/MY_ANALYSIS.md
```

### ✅ DO THIS INSTEAD:
```
/phoenix-platform/docs/planning/IMPLEMENTATION_PLAN.md
/phoenix-platform/docs/planning/PROJECT_STATUS.md
/phoenix-platform/docs/reviews/REVIEW_SUMMARY.md
/phoenix-platform/docs/reviews/MY_ANALYSIS.md
```

## Migration Instructions

If you find documentation in the wrong place:

1. **Identify correct location** using the hierarchy above
2. **Move the file**: `git mv <old-path> <new-path>`
3. **Update any references** to the moved file
4. **Commit with message**: `docs: relocate <filename> to follow governance rules`

## Exceptions

The ONLY exceptions to these rules:

1. **CLAUDE.md** - Must be at root for Claude Code to find it
2. **GitHub-required files** - LICENSE, SECURITY.md (if required)
3. **Build tool configs** - Only if they must be at root

## Consequences of Violations

1. **PR will be blocked** by automated checks
2. **Review required** to explain why exception is needed
3. **Documentation debt** tracked if temporary exception granted

## Regular Audits

Monthly documentation audits will check:
- Files in correct locations
- No duplicate documentation
- No orphaned documents
- Proper cross-references

## Summary

**Remember: When in doubt, documentation goes in `/phoenix-platform/docs/`**

This maintains:
- Clean repository root
- Clear navigation structure  
- Proper separation of concerns
- Easy documentation discovery

Following these rules ensures the Phoenix mono-repo remains organized and maintainable as it grows.