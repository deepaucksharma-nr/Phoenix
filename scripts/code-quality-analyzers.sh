#!/bin/bash
# code-quality-analyzers.sh - Automated code quality checks for Phoenix Platform
# Created by Abhinav as part of code quality analyzers task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=== Phoenix Platform Code Quality Analyzers ===${NC}"
echo ""

# Configuration
DEFAULT_CONFIG_DIR="$REPO_ROOT/.quality"
SKIP_TESTS=${SKIP_TESTS:-false}
VERBOSE=${VERBOSE:-false}
PR_MODE=${PR_MODE:-false}
EXIT_ON_ERROR=${EXIT_ON_ERROR:-true}
ERROR_COUNT=0
WARNING_COUNT=0

# Create quality config directory if it doesn't exist
mkdir -p "$DEFAULT_CONFIG_DIR"

# Create default .golangci.yml if it doesn't exist
if [ ! -f "$REPO_ROOT/.golangci.yml" ]; then
    cat > "$REPO_ROOT/.golangci.yml" << 'EOF'
# Phoenix Platform - GolangCI-Lint Configuration

run:
  timeout: 5m
  issues-exit-code: 1
  tests: true
  modules-download-mode: readonly
  allow-parallel-runners: true
  skip-files:
    - ".*\\.generated\\.go$"
    - ".*\\_test\\.go$"

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
  uniq-by-line: true
  sort-results: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gosec
    - gocyclo
    - gofmt
    - goimports
    - misspell
    - revive
    - bodyclose
    - noctx
    - exportloopref
    - prealloc
    - unconvert
    - unparam
    - gocritic
    - goerr113
    - nolintlint

linters-settings:
  gocyclo:
    min-complexity: 15
  govet:
    check-shadowing: true
  gosec:
    excludes:
      - G104 # Errors unhandled in test files can be excluded
  revive:
    ignore-generated-header: true
    severity: warning
    rules:
      - name: exported
        severity: warning
        disabled: false
  gocritic:
    enabled-tags:
      - diagnostic
      - performance
      - style
      - opinionated
    disabled-checks:
      - commentedOutCode

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - errcheck
  max-issues-per-linter: 50
  max-same-issues: 5
  fix: true
EOF
    echo -e "${GREEN}✓${NC} Created default .golangci.yml"
fi

# Create eslintrc.json if it doesn't exist
if [ ! -f "$REPO_ROOT/.eslintrc.json" ]; then
    cat > "$REPO_ROOT/.eslintrc.json" << 'EOF'
{
  "root": true,
  "env": {
    "browser": true,
    "node": true,
    "es2021": true
  },
  "extends": [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended"
  ],
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": 12,
    "sourceType": "module"
  },
  "plugins": [
    "@typescript-eslint"
  ],
  "rules": {
    "indent": [
      "error",
      2,
      {
        "SwitchCase": 1
      }
    ],
    "linebreak-style": [
      "error",
      "unix"
    ],
    "quotes": [
      "error",
      "single",
      {
        "allowTemplateLiterals": true
      }
    ],
    "semi": [
      "error",
      "always"
    ],
    "no-console": [
      "warn",
      {
        "allow": [
          "warn",
          "error",
          "info"
        ]
      }
    ],
    "@typescript-eslint/no-unused-vars": [
      "warn",
      {
        "argsIgnorePattern": "^_",
        "varsIgnorePattern": "^_"
      }
    ],
    "@typescript-eslint/explicit-module-boundary-types": "off"
  },
  "overrides": [
    {
      "files": [
        "*.test.ts",
        "*.test.tsx",
        "*.spec.ts",
        "*.spec.tsx"
      ],
      "env": {
        "jest": true
      },
      "rules": {
        "@typescript-eslint/no-explicit-any": "off"
      }
    }
  ]
}
EOF
    echo -e "${GREEN}✓${NC} Created default .eslintrc.json"
fi

# Create prettier.config.js if it doesn't exist
if [ ! -f "$REPO_ROOT/prettier.config.js" ]; then
    cat > "$REPO_ROOT/prettier.config.js" << 'EOF'
module.exports = {
  semi: true,
  trailingComma: 'all',
  singleQuote: true,
  printWidth: 100,
  tabWidth: 2,
  endOfLine: 'lf',
  arrowParens: 'avoid',
};
EOF
    echo -e "${GREEN}✓${NC} Created default prettier.config.js"
fi

# Helper functions
function log_error() {
    echo -e "${RED}ERROR: $1${NC}"
    ERROR_COUNT=$((ERROR_COUNT + 1))
    if [ "$EXIT_ON_ERROR" = "true" ]; then
        exit 1
    fi
}

function log_warning() {
    echo -e "${YELLOW}WARNING: $1${NC}"
    WARNING_COUNT=$((WARNING_COUNT + 1))
}

function log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

function log_info() {
    echo -e "${BLUE}$1${NC}"
}

function run_command() {
    local cmd="$1"
    local description="$2"
    local error_msg="$3"
    
    echo -e "${BLUE}Running $description...${NC}"
    if [ "$VERBOSE" = "true" ]; then
        eval "$cmd" || log_error "$error_msg"
    else
        eval "$cmd" > /dev/null 2>&1 || log_error "$error_msg"
    fi
    log_success "$description completed"
}

# Function to analyze Go code
function analyze_go_code() {
    log_info "Analyzing Go code..."
    
    # Find Go projects
    local go_modules=$(find "$REPO_ROOT" -name "go.mod" -not -path "*/vendor/*" -not -path "*/node_modules/*")
    
    if [ -z "$go_modules" ]; then
        log_info "No Go modules found"
        return 0
    fi
    
    # Ensure golangci-lint is available
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint not found. Installing..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.55.2
    fi
    
    # Process each module
    for mod in $go_modules; do
        local mod_dir=$(dirname "$mod")
        local rel_path="${mod_dir#$REPO_ROOT/}"
        
        log_info "Analyzing Go module: $rel_path"
        
        # Run golangci-lint
        (cd "$mod_dir" && golangci-lint run --config="$REPO_ROOT/.golangci.yml" ./...) || {
            log_error "golangci-lint failed for $rel_path"
            continue
        }
        
        # Run go vet if golangci-lint succeeds
        (cd "$mod_dir" && go vet ./...) || {
            log_error "go vet failed for $rel_path"
        }
        
        # Check for correct formatting (gofmt)
        local unformatted=$(cd "$mod_dir" && gofmt -l .)
        if [ -n "$unformatted" ]; then
            log_warning "Unformatted Go files in $rel_path:"
            echo "$unformatted"
            log_info "Run 'cd $rel_path && gofmt -w .' to format"
        fi
        
        # Run tests if not skipped
        if [ "$SKIP_TESTS" != "true" ]; then
            (cd "$mod_dir" && go test -short ./...) || {
                log_error "Tests failed for $rel_path"
            }
        fi
        
        log_success "Go analysis completed for $rel_path"
    done
}

# Function to analyze JavaScript/TypeScript code
function analyze_js_ts_code() {
    log_info "Analyzing JavaScript/TypeScript code..."
    
    # Find package.json files (excluding node_modules)
    local package_jsons=$(find "$REPO_ROOT" -name "package.json" -not -path "*/node_modules/*")
    
    if [ -z "$package_jsons" ]; then
        log_info "No JavaScript/TypeScript projects found"
        return 0
    fi
    
    # Process each project
    for pkg in $package_jsons; do
        local pkg_dir=$(dirname "$pkg")
        local rel_path="${pkg_dir#$REPO_ROOT/}"
        
        # Skip root package.json
        if [ "$pkg_dir" = "$REPO_ROOT" ]; then
            continue
        }
        
        log_info "Analyzing JS/TS project: $rel_path"
        
        # Check if project has eslint dependency
        if grep -q '"eslint"' "$pkg" || grep -q '"eslint":' "$pkg"; then
            # Run ESLint if available in project
            if [ -f "$pkg_dir/node_modules/.bin/eslint" ]; then
                (cd "$pkg_dir" && ./node_modules/.bin/eslint . --ext .js,.jsx,.ts,.tsx) || {
                    log_error "ESLint failed for $rel_path"
                }
            elif command -v eslint &> /dev/null; then
                (cd "$pkg_dir" && eslint . --ext .js,.jsx,.ts,.tsx) || {
                    log_error "ESLint failed for $rel_path"
                }
            else
                log_warning "ESLint not found for $rel_path"
            }
        else
            log_info "ESLint not configured in $rel_path"
        }
        
        # Check if project has prettier dependency
        if grep -q '"prettier"' "$pkg" || grep -q '"prettier":' "$pkg"; then
            # Run prettier if available in project
            if [ -f "$pkg_dir/node_modules/.bin/prettier" ]; then
                (cd "$pkg_dir" && ./node_modules/.bin/prettier --check "**/*.{js,jsx,ts,tsx,json,css,scss,md}") || {
                    log_warning "Prettier check failed for $rel_path"
                    log_info "Run 'cd $rel_path && npx prettier --write \"**/*.{js,jsx,ts,tsx,json,css,scss,md}\"' to format"
                }
            elif command -v prettier &> /dev/null; then
                (cd "$pkg_dir" && prettier --check "**/*.{js,jsx,ts,tsx,json,css,scss,md}") || {
                    log_warning "Prettier check failed for $rel_path"
                    log_info "Run 'cd $rel_path && npx prettier --write \"**/*.{js,jsx,ts,tsx,json,css,scss,md}\"' to format"
                }
            else
                log_warning "Prettier not found for $rel_path"
            }
        fi
        
        # Run tests if not skipped and if test script exists
        if [ "$SKIP_TESTS" != "true" ] && grep -q '"test":' "$pkg"; then
            # Determine package manager based on lock files
            if [ -f "$pkg_dir/pnpm-lock.yaml" ]; then
                (cd "$pkg_dir" && pnpm test) || {
                    log_error "Tests failed for $rel_path"
                }
            elif [ -f "$pkg_dir/yarn.lock" ]; then
                (cd "$pkg_dir" && yarn test) || {
                    log_error "Tests failed for $rel_path"
                }
            else
                (cd "$pkg_dir" && npm test) || {
                    log_error "Tests failed for $rel_path"
                }
            fi
        fi
        
        log_success "JS/TS analysis completed for $rel_path"
    done
}

# Function to analyze YAML files
function analyze_yaml_files() {
    log_info "Analyzing YAML files..."
    
    # Check if yamllint is available
    if ! command -v yamllint &> /dev/null; then
        log_warning "yamllint not found, skipping YAML analysis"
        return 0
    fi
    
    # Create default yamllint config if it doesn't exist
    if [ ! -f "$DEFAULT_CONFIG_DIR/.yamllint.yml" ]; then
        cat > "$DEFAULT_CONFIG_DIR/.yamllint.yml" << 'EOF'
extends: default

rules:
  line-length:
    max: 120
    level: warning
  document-start:
    present: false
  comments:
    require-starting-space: true
    min-spaces-from-content: 1
  braces:
    min-spaces-inside: 0
    max-spaces-inside: 1
  brackets:
    min-spaces-inside: 0
    max-spaces-inside: 0
  indentation:
    spaces: 2
    indent-sequences: consistent
  truthy:
    allowed-values: ['true', 'false', 'yes', 'no']
EOF
        echo -e "${GREEN}✓${NC} Created default YAML lint config"
    fi
    
    # Find YAML files
    local yaml_files=$(find "$REPO_ROOT" -name "*.yml" -o -name "*.yaml" | grep -v "node_modules\|vendor\|.git")
    
    if [ -z "$yaml_files" ]; then
        log_info "No YAML files found"
        return 0
    fi
    
    # Run yamllint on all YAML files
    yamllint -c "$DEFAULT_CONFIG_DIR/.yamllint.yml" $yaml_files || {
        log_warning "YAML linting found issues"
    }
    
    log_success "YAML analysis completed"
}

# Function to analyze Dockerfiles
function analyze_dockerfiles() {
    log_info "Analyzing Dockerfiles..."
    
    # Check if hadolint is available
    if ! command -v hadolint &> /dev/null; then
        log_warning "hadolint not found, skipping Dockerfile analysis"
        return 0
    fi
    
    # Find Dockerfiles
    local dockerfiles=$(find "$REPO_ROOT" -name "Dockerfile*" | grep -v "node_modules\|vendor\|.git")
    
    if [ -z "$dockerfiles" ]; then
        log_info "No Dockerfiles found"
        return 0
    }
    
    # Create default hadolint config if it doesn't exist
    if [ ! -f "$DEFAULT_CONFIG_DIR/.hadolint.yaml" ]; then
        cat > "$DEFAULT_CONFIG_DIR/.hadolint.yaml" << 'EOF'
ignored:
  - DL3008 # Pin versions in apt-get install
  - DL3009 # Delete the apt-get lists
  - DL3059 # Multiple consecutive RUN instructions
  - SC2046 # Quote this to prevent word splitting

trustedRegistries:
  - docker.io
  - ghcr.io
EOF
        echo -e "${GREEN}✓${NC} Created default hadolint config"
    fi
    
    # Run hadolint on all Dockerfiles
    for dockerfile in $dockerfiles; do
        local rel_path="${dockerfile#$REPO_ROOT/}"
        hadolint --config "$DEFAULT_CONFIG_DIR/.hadolint.yaml" "$dockerfile" || {
            log_warning "Hadolint found issues in $rel_path"
        }
    done
    
    log_success "Dockerfile analysis completed"
}

# Function to analyze shell scripts
function analyze_shell_scripts() {
    log_info "Analyzing shell scripts..."
    
    # Check if shellcheck is available
    if ! command -v shellcheck &> /dev/null; then
        log_warning "shellcheck not found, skipping shell script analysis"
        return 0
    fi
    
    # Find shell scripts
    local shell_scripts=$(find "$REPO_ROOT" -name "*.sh" | grep -v "node_modules\|vendor\|.git")
    
    if [ -z "$shell_scripts" ]; then
        log_info "No shell scripts found"
        return 0
    }
    
    # Run shellcheck on all shell scripts
    for script in $shell_scripts; do
        local rel_path="${script#$REPO_ROOT/}"
        shellcheck -x "$script" || {
            log_error "Shellcheck found issues in $rel_path"
        }
    done
    
    log_success "Shell script analysis completed"
}

# Function to check for security issues
function analyze_security() {
    log_info "Analyzing for security issues..."
    
    # Check if gosec is available for Go security analysis
    if command -v gosec &> /dev/null; then
        log_info "Running gosec for Go security analysis..."
        
        # Find Go modules
        local go_modules=$(find "$REPO_ROOT" -name "go.mod" -not -path "*/vendor/*" -not -path "*/node_modules/*")
        
        for mod in $go_modules; do
            local mod_dir=$(dirname "$mod")
            local rel_path="${mod_dir#$REPO_ROOT/}"
            
            (cd "$mod_dir" && gosec ./...) || {
                log_warning "Gosec found security issues in $rel_path"
            }
        done
    else
        log_warning "gosec not found, skipping Go security analysis"
    fi
    
    # Check for hardcoded secrets
    log_info "Checking for hardcoded secrets..."
    if command -v detect-secrets &> /dev/null; then
        detect-secrets scan --exclude-files "package-lock.json|yarn.lock|pnpm-lock.yaml" "$REPO_ROOT" || {
            log_warning "Potential secrets found in the codebase"
        }
    elif command -v grep &> /dev/null; then
        # Simple pattern matching for common secrets if detect-secrets isn't available
        local patterns=(
            "password\s*=\s*['\"][^'\"]+['\"]"
            "api[_-]?key\s*=\s*['\"][^'\"]+['\"]"
            "secret\s*=\s*['\"][^'\"]+['\"]"
            "token\s*=\s*['\"][^'\"]+['\"]"
        )
        
        for pattern in "${patterns[@]}"; do
            log_info "Checking for pattern: $pattern"
            grep -r -E -I --include="*.{js,ts,go,py,sh,java,rb}" "$pattern" "$REPO_ROOT" --exclude-dir={node_modules,vendor,.git} || true
        done
    fi
    
    log_success "Security analysis completed"
}

# Function to run PR-mode checks (for CI/CD)
function run_pr_mode() {
    log_info "Running in PR mode (strict checking)..."
    
    EXIT_ON_ERROR="true"
    SKIP_TESTS="false"
    VERBOSE="true"
    
    # Run essential checks
    analyze_go_code
    analyze_js_ts_code
    analyze_yaml_files
    analyze_dockerfiles
    analyze_shell_scripts
    analyze_security
}

# Main execution
if [ "$PR_MODE" = "true" ]; then
    run_pr_mode
else
    # Run all analyzers
    analyze_go_code
    analyze_js_ts_code
    analyze_yaml_files
    analyze_dockerfiles
    analyze_shell_scripts
    analyze_security
fi

# Print summary
echo ""
echo -e "${BLUE}=== Code Quality Analysis Summary ===${NC}"
echo "Errors: $ERROR_COUNT"
echo "Warnings: $WARNING_COUNT"

if [ "$ERROR_COUNT" -gt 0 ]; then
    echo -e "${RED}Analysis completed with errors${NC}"
    exit 1
elif [ "$WARNING_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}Analysis completed with warnings${NC}"
    exit 0
else
    echo -e "${GREEN}Analysis completed successfully${NC}"
    exit 0
fi
