# Phoenix Platform Governance Enforcement

This document describes the automated enforcement mechanisms for the Phoenix Platform mono-repo governance rules.

## Overview

The Phoenix Platform uses a multi-layered approach to enforce code quality, architectural boundaries, and development standards:

1. **Pre-commit Hooks** - Catch issues before code is committed
2. **CI/CD Pipeline** - Validate all changes in pull requests
3. **Static Analysis** - Enforce coding standards and detect issues
4. **Structural Validation** - Ensure mono-repo organization is maintained
5. **Code Ownership** - Require appropriate reviews for changes

## Setup

### Initial Setup

Run the governance setup script to install all required tools:

```bash
cd phoenix-platform
./scripts/setup-governance.sh
```

This will install:
- GolangCI-Lint for Go code analysis
- Pre-commit hooks for automatic validation
- Commitlint for commit message standards
- Various linting tools

### Manual Setup

If you prefer manual setup:

```bash
# Install pre-commit
pip install pre-commit

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b $(go env GOPATH)/bin v1.55.0

# Install commitlint
npm install -g @commitlint/cli @commitlint/config-conventional

# Setup hooks
pre-commit install
pre-commit install --hook-type commit-msg
```

## Enforcement Layers

### 1. Pre-commit Hooks

The `.pre-commit-config.yaml` file defines checks that run automatically before each commit:

- **Go Formatting** - Ensures consistent code style
- **Go Imports** - Organizes and validates imports
- **Go Vet** - Catches common Go mistakes
- **GolangCI-Lint** - Comprehensive Go analysis
- **Frontend Linting** - ESLint for JavaScript/TypeScript
- **YAML/Markdown Linting** - Ensures documentation quality
- **Structure Validation** - Checks mono-repo organization
- **Import Validation** - Enforces architectural boundaries
- **Secret Detection** - Prevents committing sensitive data

### 2. Commit Message Standards

All commits must follow the conventional commit format:

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`

Scopes: `api`, `dashboard`, `controller`, `generator`, `operator`, `simulator`, `pkg`, `auth`, `store`, `docs`, `deps`, `ci`, `helm`, `k8s`

Examples:
```bash
feat(api): add experiment validation endpoint
fix(controller): correct state transition logic
docs(api): update API documentation
test(pkg): add unit tests for store package
```

### 3. Code Organization Rules

#### Import Rules

The `validate-imports.go` script enforces:

1. **No cmd imports** - Packages cannot import from `cmd/` directories
2. **Internal package isolation** - Internal packages cannot be imported across service boundaries
3. **Operator isolation** - Operators cannot import from other operators

#### Directory Structure

Required directories:
- `cmd/` - Service entry points
- `pkg/` - Shared packages
- `operators/` - Kubernetes operators
- `dashboard/` - Frontend application
- `docs/` - Documentation
- `scripts/` - Build and utility scripts

### 4. Code Review Requirements

The `CODEOWNERS` file enforces review requirements:

- **Standard changes**: 1 approval from service owner
- **Shared packages**: 2 approvals (service owner + platform lead)
- **Security changes**: Requires security team approval
- **Breaking changes**: 2 approvals including platform lead

### 5. CI/CD Enforcement

GitHub Actions workflow (`.github/workflows/ci.yml`) runs:

1. **Structure Validation** - Ensures mono-repo structure
2. **Go Linting** - Full GolangCI-Lint analysis
3. **Frontend Linting** - ESLint and format checking
4. **Tests** - Unit and integration tests with coverage
5. **Security Scanning** - Trivy and gosec vulnerability scanning
6. **Build Validation** - Ensures all components build

## Available Commands

### Validation Commands

```bash
# Run all validation checks
make validate

# Check mono-repo structure
make validate-structure

# Check Go import rules
make validate-imports

# Run all pre-commit checks
make verify
```

### Code Quality Commands

```bash
# Format all code
make fmt

# Run linters
make lint

# Run tests with coverage
make coverage

# Build all components
make build
```

## Troubleshooting

### Pre-commit Hook Failures

If pre-commit hooks fail:

1. **Format issues**: Run `make fmt` to auto-fix
2. **Lint issues**: Check the specific linter output
3. **Import violations**: Refactor to remove invalid imports
4. **Structure issues**: Ensure required directories exist

### Commit Message Rejection

If your commit message is rejected:

1. Check the format matches: `type(scope): subject`
2. Use a valid type and scope
3. Keep subject line under 72 characters
4. Use present tense ("add" not "added")

### CI Pipeline Failures

If CI fails on your PR:

1. Check the specific job that failed
2. Run the same command locally (e.g., `make lint`)
3. Fix issues and push updates
4. All checks must pass before merge

## Best Practices

1. **Run checks locally** before pushing:
   ```bash
   make verify
   ```

2. **Keep commits focused** - One logical change per commit

3. **Write meaningful commit messages** - Explain why, not just what

4. **Follow the architecture** - Respect service boundaries

5. **Add tests** - Maintain or increase coverage

6. **Update documentation** - Keep docs in sync with code

## Extending Governance

To add new governance rules:

1. **Update validation scripts** in `scripts/`
2. **Add linter rules** to `.golangci.yml`
3. **Update pre-commit hooks** in `.pre-commit-config.yaml`
4. **Add CI checks** to `.github/workflows/ci.yml`
5. **Document changes** in this file

## Support

For questions or issues with governance:

1. Check this documentation
2. Run `make help` for available commands
3. Contact the platform team
4. File an issue in the repository