# Phoenix Platform Static Analysis Rules

## Overview

This document defines the comprehensive static analysis rules that enforce the Phoenix platform's mono-repo structure and coding standards. These rules are automatically enforced through various tools and CI/CD pipelines to prevent architectural drift and maintain code quality.

## Repository Structure Rules

### 1. Root Directory Structure

```yaml
root_structure:
  required_directories:
    - cmd/           # Service entry points
    - pkg/           # Shared packages
    - internal/      # Private packages
    - api/           # API definitions
    - configs/       # Configuration files
    - scripts/       # Build and utility scripts
    - docs/          # Documentation
    - deployments/   # Deployment manifests
    - tests/         # Integration tests
    
  required_files:
    - README.md
    - Makefile
    - go.mod
    - go.sum
    - .gitignore
    - .golangci.yml
    - CLAUDE.md
    
  forbidden_patterns:
    - "*.exe"
    - "*.dll"
    - ".env"
    - "secrets/*"
    - "vendor/*"  # Use go modules
```

### 2. Service Structure Rules

```yaml
service_structure:
  pattern: "cmd/${service_name}/"
  required:
    structure:
      - main.go
      - internal/
      - internal/${service}/
      - configs/
      
    documentation:
      - README.md
      - docs/TECHNICAL_SPEC_${SERVICE_NAME}.md
      
    deployment:
      - deployments/docker/Dockerfile
      - deployments/k8s/${service}.yaml
      
  validation:
    - name: "main_package_check"
      rule: "main.go must be in package main"
      enforced_by: "go vet"
      
    - name: "internal_imports"
      rule: "internal/ packages can only be imported by the same service"
      enforced_by: "go mod verify"
```

### 3. Package Structure Rules

```yaml
package_structure:
  shared_packages:
    location: "pkg/"
    rules:
      - "Must have a clear, single responsibility"
      - "Must include comprehensive tests (>80% coverage)"
      - "Must have package-level documentation"
      - "Cannot import from cmd/ or internal/"
      
  internal_packages:
    location: "internal/"
    rules:
      - "Can only be imported within the same module"
      - "Should follow Domain-Driven Design principles"
      - "Must have unit tests for public functions"
      
  api_packages:
    location: "api/"
    rules:
      - "Must contain protobuf or OpenAPI definitions"
      - "Generated code goes in pkg/${service}/client/"
      - "Must version APIs (v1, v2, etc.)"
```

## Code Quality Rules

### 1. Go Static Analysis

```yaml
# .golangci.yml
run:
  timeout: 5m
  skip-dirs:
    - vendor
    - third_party
    - testdata
    
linters:
  enable:
    # Code quality
    - gofmt
    - goimports
    - golint
    - govet
    - errcheck
    - staticcheck
    - unused
    - structcheck
    - varcheck
    - deadcode
    - ineffassign
    - typecheck
    
    # Best practices
    - goconst
    - gocyclo
    - dupl
    - gosec
    - prealloc
    - exportloopref
    
    # Style
    - godot
    - godox
    - misspell
    - whitespace
    
linters-settings:
  gocyclo:
    min-complexity: 15
    
  dupl:
    threshold: 100
    
  goconst:
    min-len: 3
    min-occurrences: 3
    
  misspell:
    locale: US
    
  godox:
    keywords:
      - NOTE
      - OPTIMIZE
      - HACK
      - BUG
      - FIXME
      - TODO
      
  govet:
    check-shadowing: true
    enable-all: true
    
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl
        - gosec
        
    - path: cmd/
      linters:
        - gochecknoglobals
        
  max-issues-per-linter: 0
  max-same-issues: 0
```

### 2. Import Rules

```go
// tools/analysis/imports.go
package analysis

import (
    "go/ast"
    "go/parser"
    "go/token"
    "strings"
)

// ImportRules defines allowed import patterns
var ImportRules = []Rule{
    {
        Name: "no_internal_cross_imports",
        Check: func(pkg, imp string) error {
            if strings.Contains(imp, "/internal/") && 
               !strings.HasPrefix(imp, getModulePrefix(pkg)) {
                return fmt.Errorf("cannot import internal package from different module")
            }
            return nil
        },
    },
    {
        Name: "no_cmd_imports",
        Check: func(pkg, imp string) error {
            if strings.Contains(imp, "/cmd/") && !strings.Contains(pkg, "/cmd/") {
                return fmt.Errorf("cannot import from cmd packages")
            }
            return nil
        },
    },
    {
        Name: "api_version_imports",
        Check: func(pkg, imp string) error {
            if strings.Contains(imp, "/api/") && !strings.Contains(imp, "/v") {
                return fmt.Errorf("API imports must be versioned (e.g., /api/v1/)")
            }
            return nil
        },
    },
}
```

### 3. Naming Conventions

```yaml
naming_conventions:
  files:
    - rule: "Go files must use snake_case"
      pattern: "^[a-z][a-z0-9_]*\\.go$"
      exceptions: ["doc.go"]
      
  packages:
    - rule: "Package names must be lowercase single words"
      pattern: "^[a-z]+$"
      enforced_by: "go vet"
      
  types:
    - rule: "Exported types must use PascalCase"
      pattern: "^[A-Z][a-zA-Z0-9]*$"
      
    - rule: "Interfaces should end with 'er'"
      pattern: "^[A-Z][a-zA-Z0-9]*er$"
      exceptions: ["Client", "Server", "Service"]
      
  functions:
    - rule: "Exported functions must use PascalCase"
      pattern: "^[A-Z][a-zA-Z0-9]*$"
      
    - rule: "Test functions must start with Test"
      pattern: "^Test[A-Z][a-zA-Z0-9]*$"
      
  constants:
    - rule: "Exported constants must use PascalCase or SCREAMING_SNAKE_CASE"
      pattern: "^([A-Z][a-zA-Z0-9]*|[A-Z][A-Z0-9_]*)$"
```

## Dependency Rules

### 1. Module Dependencies

```yaml
dependency_rules:
  allowed_dependencies:
    production:
      - "github.com/gin-gonic/gin"          # HTTP framework
      - "google.golang.org/grpc"            # gRPC
      - "github.com/lib/pq"                 # PostgreSQL
      - "github.com/go-redis/redis/v8"      # Redis
      - "k8s.io/client-go"                  # Kubernetes
      - "github.com/prometheus/client_golang" # Metrics
      - "go.uber.org/zap"                   # Logging
      - "github.com/spf13/viper"            # Configuration
      
    test:
      - "github.com/stretchr/testify"       # Assertions
      - "github.com/golang/mock"            # Mocking
      - "github.com/DATA-DOG/go-sqlmock"    # SQL mocking
      
  forbidden_dependencies:
    - "github.com/pkg/errors"  # Use stdlib errors
    - "github.com/sirupsen/logrus"  # Use zap
    - "gopkg.in/*"  # Avoid v1 style imports
    
  version_constraints:
    - rule: "Kubernetes client-go version must match cluster version"
    - rule: "All dependencies must use go modules (no vendor/)"
```

### 2. Dependency Analysis Script

```bash
#!/bin/bash
# scripts/analyze-dependencies.sh

check_forbidden_deps() {
    forbidden=(
        "github.com/pkg/errors"
        "github.com/sirupsen/logrus"
        "gopkg.in/"
    )
    
    for dep in "${forbidden[@]}"; do
        if grep -q "$dep" go.mod; then
            echo "ERROR: Forbidden dependency found: $dep"
            exit 1
        fi
    done
}

check_version_conflicts() {
    # Check for multiple versions of the same dependency
    duplicates=$(go mod graph | cut -d '@' -f 1 | sort | uniq -d)
    if [ -n "$duplicates" ]; then
        echo "WARNING: Multiple versions of dependencies found:"
        echo "$duplicates"
    fi
}

check_indirect_deps() {
    indirect_count=$(grep -c "// indirect" go.mod)
    if [ "$indirect_count" -gt 50 ]; then
        echo "WARNING: Too many indirect dependencies ($indirect_count)"
        echo "Consider upgrading direct dependencies"
    fi
}
```

## Testing Rules

### 1. Test Coverage Requirements

```yaml
test_coverage:
  minimum_coverage:
    overall: 80%
    packages:
      - pkg/*: 85%
      - internal/*: 75%
      - cmd/*: 60%
      
  excluded_patterns:
    - "*.pb.go"      # Generated protobuf
    - "*_gen.go"     # Other generated code
    - "testdata/*"   # Test fixtures
    
  enforcement:
    - tool: "go test -coverprofile"
    - ci_check: "coverage must not decrease"
```

### 2. Test Structure Rules

```go
// Test file naming and structure enforcement
package rules

var TestRules = []Rule{
    {
        Name: "test_file_naming",
        Pattern: regexp.MustCompile(`^.*_test\.go$`),
        Message: "Test files must end with _test.go",
    },
    {
        Name: "test_package_naming",
        Check: func(pkg *ast.Package) error {
            if strings.HasSuffix(pkg.Name, "_test") {
                // External test package - OK
                return nil
            }
            // Same package tests - also OK
            return nil
        },
    },
    {
        Name: "table_driven_tests",
        Check: func(fn *ast.FuncDecl) error {
            if !strings.HasPrefix(fn.Name.Name, "Test") {
                return nil
            }
            // Check for table-driven test pattern
            hasTestCases := false
            ast.Inspect(fn, func(n ast.Node) bool {
                if ident, ok := n.(*ast.Ident); ok {
                    if ident.Name == "tests" || ident.Name == "testCases" {
                        hasTestCases = true
                    }
                }
                return true
            })
            if !hasTestCases && countTestCases(fn) > 3 {
                return fmt.Errorf("consider using table-driven tests")
            }
            return nil
        },
    },
}
```

## Security Rules

### 1. Security Scanning

```yaml
security_rules:
  static_analysis:
    - tool: "gosec"
      config:
        severity: "medium"
        confidence: "medium"
        exclude:
          - G104  # Unhandled errors (covered by errcheck)
          
    - tool: "nancy"
      action: "audit dependencies for vulnerabilities"
      
  secret_scanning:
    - tool: "gitleaks"
      config: ".gitleaks.toml"
      
  forbidden_patterns:
    - pattern: "(?i)(api[_-]?key|password|secret|token)\\s*=\\s*[\"'][^\"']+[\"']"
      message: "Hardcoded secrets detected"
      
    - pattern: "http://"
      message: "Use HTTPS for external connections"
      exceptions: ["localhost", "127.0.0.1"]
```

### 2. Security Checklist

```go
// tools/security/checklist.go
package security

var SecurityChecks = []Check{
    {
        Name: "no_sql_string_concat",
        Check: func(node ast.Node) error {
            // Detect SQL string concatenation
            if call, ok := node.(*ast.CallExpr); ok {
                if isStringConcat(call) && containsSQLKeywords(call) {
                    return fmt.Errorf("potential SQL injection via string concatenation")
                }
            }
            return nil
        },
    },
    {
        Name: "no_weak_crypto",
        Check: func(imports []*ast.ImportSpec) error {
            weakCrypto := []string{
                "crypto/md5",
                "crypto/sha1",
                "crypto/des",
            }
            for _, imp := range imports {
                path := strings.Trim(imp.Path.Value, `"`)
                for _, weak := range weakCrypto {
                    if path == weak {
                        return fmt.Errorf("weak cryptography: %s", weak)
                    }
                }
            }
            return nil
        },
    },
}
```

## Documentation Rules

### 1. Code Documentation

```yaml
documentation_rules:
  package_docs:
    - rule: "Every package must have a doc.go file"
    - template: |
        // Package ${package_name} provides ${description}.
        //
        // ${detailed_description}
        package ${package_name}
        
  exported_items:
    - rule: "All exported types, functions, and constants must have godoc comments"
    - format: "Comment must start with the item name"
    
  examples:
    - rule: "Public APIs must have example functions"
    - location: "*_example_test.go"
```

### 2. API Documentation

```yaml
api_documentation:
  openapi:
    - rule: "All REST APIs must have OpenAPI 3.0 specs"
    - location: "api/openapi/*.yaml"
    - validation: "openapi-generator validate"
    
  grpc:
    - rule: "All gRPC services must have comprehensive proto comments"
    - format: |
        // ServiceName does something.
        //
        // Detailed description...
        service ServiceName {
            // MethodName does something specific.
            rpc MethodName(Request) returns (Response) {
                option (google.api.http) = {
                    post: "/v1/resource"
                    body: "*"
                };
            }
        }
```

## Build Rules

### 1. Makefile Standards

```makefile
# Required Makefile targets
.PHONY: all
all: build

.PHONY: build
build: validate
	@echo "Building all services..."
	@$(MAKE) -C cmd/api build
	@$(MAKE) -C cmd/controller build
	@$(MAKE) -C cmd/simulator build

.PHONY: validate
validate: validate-structure lint test

.PHONY: validate-structure
validate-structure:
	@echo "Validating repository structure..."
	@go run tools/validate/main.go

.PHONY: lint
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	@go test -race -coverprofile=coverage.out ./...

.PHONY: coverage
coverage: test
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/ dist/ coverage.*
```

### 2. Build Constraints

```yaml
build_constraints:
  binary_naming:
    pattern: "phoenix-${service_name}"
    
  build_tags:
    required:
      - "containers"
      - "orchestration"
    optional:
      - "integration"
      - "e2e"
      
  ldflags:
    required:
      - "-X main.version=${VERSION}"
      - "-X main.commit=${GIT_COMMIT}"
      - "-X main.buildDate=${BUILD_DATE}"
```

## CI/CD Enforcement

### 1. GitHub Actions Workflow

```yaml
# .github/workflows/static-analysis.yml
name: Static Analysis

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  structure-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Validate Structure
        run: |
          make validate-structure
          
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: 1.21
          
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
          
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: ./...
          
      - name: Run gitleaks
        uses: gitleaks/gitleaks-action@v2
        
  test-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: 1.21
          
      - name: Run tests
        run: make test
        
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi
```

### 2. Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: Go Format
        entry: gofmt -l -w
        language: system
        files: \.go$
        
      - id: go-imports
        name: Go Imports
        entry: goimports -l -w
        language: system
        files: \.go$
        
      - id: go-vet
        name: Go Vet
        entry: go vet
        language: system
        files: \.go$
        pass_filenames: false
        
      - id: validate-structure
        name: Validate Structure
        entry: make validate-structure
        language: system
        pass_filenames: false
        
      - id: no-secrets
        name: Check Secrets
        entry: gitleaks detect --source . --verbose
        language: system
        pass_filenames: false
```

## Automated Enforcement Tools

### 1. Structure Validator

```go
// tools/validate/main.go
package main

import (
    "fmt"
    "os"
    "path/filepath"
)

var requiredStructure = map[string][]string{
    ".": {
        "cmd/", "pkg/", "internal/", "api/", "configs/",
        "scripts/", "docs/", "deployments/", "tests/",
        "README.md", "Makefile", "go.mod", ".gitignore",
    },
    "cmd/*": {
        "main.go", "internal/", "configs/",
    },
}

func validateStructure(root string) error {
    for pattern, required := range requiredStructure {
        paths, err := filepath.Glob(filepath.Join(root, pattern))
        if err != nil {
            return err
        }
        
        for _, path := range paths {
            for _, req := range required {
                fullPath := filepath.Join(path, req)
                if _, err := os.Stat(fullPath); os.IsNotExist(err) {
                    return fmt.Errorf("missing required: %s", fullPath)
                }
            }
        }
    }
    return nil
}

func main() {
    if err := validateStructure("."); err != nil {
        fmt.Fprintf(os.Stderr, "Structure validation failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Structure validation passed")
}
```

### 2. Import Analyzer

```go
// tools/analyze/imports/main.go
package main

import (
    "go/ast"
    "go/parser"
    "go/token"
    "strings"
)

func analyzeImports(filename string) error {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
    if err != nil {
        return err
    }
    
    pkg := filepath.Dir(filename)
    
    for _, imp := range node.Imports {
        importPath := strings.Trim(imp.Path.Value, `"`)
        
        // Check internal package imports
        if strings.Contains(importPath, "/internal/") {
            if !strings.HasPrefix(importPath, getModulePrefix(pkg)) {
                return fmt.Errorf(
                    "%s: cannot import internal package from different module: %s",
                    filename, importPath,
                )
            }
        }
        
        // Check cmd package imports
        if strings.Contains(importPath, "/cmd/") {
            return fmt.Errorf(
                "%s: cannot import from cmd packages: %s",
                filename, importPath,
            )
        }
    }
    
    return nil
}
```

## Continuous Monitoring

### 1. Metrics Collection

```yaml
static_analysis_metrics:
  code_quality:
    - metric: "cyclomatic_complexity"
      threshold: 15
      action: "alert"
      
    - metric: "code_duplication"
      threshold: 100
      action: "block_merge"
      
    - metric: "test_coverage"
      threshold: 80
      action: "block_merge"
      
  dependency_health:
    - metric: "outdated_dependencies"
      threshold: 10
      action: "warning"
      
    - metric: "security_vulnerabilities"
      threshold: 0
      action: "block_merge"
      
  structure_compliance:
    - metric: "structure_violations"
      threshold: 0
      action: "block_merge"
```

### 2. Dashboard Configuration

```json
{
  "dashboard": {
    "name": "Phoenix Static Analysis",
    "panels": [
      {
        "title": "Code Quality Trends",
        "metrics": [
          "avg_cyclomatic_complexity",
          "test_coverage_percentage",
          "linter_violations_count"
        ]
      },
      {
        "title": "Dependency Health",
        "metrics": [
          "total_dependencies",
          "outdated_dependencies",
          "security_vulnerabilities"
        ]
      },
      {
        "title": "Structure Compliance",
        "metrics": [
          "structure_violations",
          "import_violations",
          "naming_violations"
        ]
      }
    ]
  }
}
```

## Exception Management

### 1. Exception Rules

```yaml
exceptions:
  structure_exceptions:
    - path: "third_party/*"
      reason: "External code"
      
    - path: "tools/*"
      reason: "Build and analysis tools"
      
  lint_exceptions:
    - rule: "line-length"
      files: ["*.pb.go", "*_gen.go"]
      reason: "Generated code"
      
    - rule: "cyclomatic-complexity"
      files: ["cmd/*/main.go"]
      reason: "Main functions can be complex"
      
  security_exceptions:
    - rule: "G104"
      files: ["*_test.go"]
      reason: "Error handling in tests"
```

### 2. Exception Process

```markdown
## Adding Exceptions

1. Exceptions must be documented in `exceptions.yaml`
2. Include justification for the exception
3. Set expiration date if temporary
4. Get approval from tech lead
5. Add exception to relevant tool configs

## Review Process

1. Monthly review of all exceptions
2. Remove expired exceptions
3. Re-evaluate permanent exceptions quarterly
4. Document decisions in ADRs
```

## Enforcement Summary

All static analysis rules are enforced at multiple levels:

1. **Local Development**: Pre-commit hooks
2. **Pull Requests**: CI/CD checks
3. **Main Branch**: Continuous monitoring
4. **Production**: Runtime validation

Violations result in:
- Build failures
- PR blocks
- Alerts to development team
- Metrics tracking

Regular audits ensure rules remain relevant and effective.