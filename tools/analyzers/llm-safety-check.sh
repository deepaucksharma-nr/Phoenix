#!/usr/bin/env bash

set -euo pipefail

# LLM Safety Check - Detects potential issues from AI-generated code
# This script looks for patterns that might indicate problematic AI generations

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
ISSUES=0
SUSPICIOUS=0

# Load AI safety configuration
AI_SAFETY_CONFIG=".ai-safety"

print_error() {
    echo -e "${RED}[SAFETY VIOLATION]${NC} $1"
    ((ISSUES++))
}

print_warning() {
    echo -e "${YELLOW}[SUSPICIOUS]${NC} $1"
    ((SUSPICIOUS++))
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

echo "=== Phoenix Platform LLM Safety Check ==="
echo

# Check if running on specific files or all
if [ $# -gt 0 ]; then
    FILES="$@"
else
    FILES=$(find . -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.py" \) \
        -not -path "./vendor/*" \
        -not -path "./node_modules/*" \
        -not -path "./.git/*" \
        -not -path "./OLD_IMPLEMENTATION/*")
fi

# Pattern checks
echo "Checking for suspicious AI-generated patterns..."

# 1. Check for excessive TODO/FIXME comments (might indicate incomplete generation)
for file in $FILES; do
    if [ -f "$file" ]; then
        todo_count=$(grep -c "TODO\|FIXME\|XXX\|HACK" "$file" 2>/dev/null || echo 0)
        if [ "$todo_count" -gt 5 ]; then
            print_warning "$file has $todo_count TODO/FIXME comments (possible incomplete generation)"
        fi
    fi
done

# 2. Check for placeholder text
placeholder_patterns=(
    "INSERT.*HERE"
    "REPLACE.*WITH"
    "YOUR.*CODE.*HERE"
    "IMPLEMENT.*THIS"
    "<.*PLACEHOLDER.*>"
    "example\.com"
    "test@example"
    "foo.*bar.*baz"
    "lorem.*ipsum"
)

for pattern in "${placeholder_patterns[@]}"; do
    if grep -r -i "$pattern" . \
        --include="*.go" \
        --include="*.js" \
        --include="*.ts" \
        --include="*.yaml" \
        --exclude-dir="vendor" \
        --exclude-dir="node_modules" \
        --exclude-dir=".git" \
        --exclude-dir="OLD_IMPLEMENTATION" \
        --exclude="*_test.go" \
        --exclude="*.test.*" \
        --exclude="*.example.*" | \
        grep -v "^Binary file"; then
        print_error "Placeholder text detected (incomplete AI generation)"
    fi
done

# 3. Check for suspicious import patterns
echo
echo "Checking for suspicious imports..."

# Unsafe or deprecated packages
suspicious_imports=(
    "github.com/dgrijalva/jwt-go"  # Should use github.com/golang-jwt/jwt
    "io/ioutil"                     # Deprecated in Go 1.16+
    "golang.org/x/net/context"      # Should use standard context
    "github.com/pkg/errors"         # Should use standard errors with %w
)

for import in "${suspicious_imports[@]}"; do
    if grep -r "\"$import\"" . \
        --include="*.go" \
        --exclude-dir="vendor" \
        --exclude-dir=".git" \
        --exclude-dir="OLD_IMPLEMENTATION"; then
        print_warning "Deprecated or unsafe import detected: $import"
    fi
done

# 4. Check for potential security issues
echo
echo "Checking for security anti-patterns..."

# Dangerous function calls
dangerous_patterns=(
    'exec\.Command'
    'os\.Exec'
    'eval\('
    'Function\('
    '__import__'
    'subprocess\.'
    'shell=True'
)

for pattern in "${dangerous_patterns[@]}"; do
    if grep -r "$pattern" . \
        --include="*.go" \
        --include="*.js" \
        --include="*.ts" \
        --include="*.py" \
        --exclude-dir="vendor" \
        --exclude-dir="node_modules" \
        --exclude-dir=".git" \
        --exclude-dir="OLD_IMPLEMENTATION" \
        --exclude="*_test.go"; then
        print_error "Potentially dangerous function usage: $pattern"
    fi
done

# 5. Check for inconsistent error handling
echo
echo "Checking error handling patterns..."

# Go specific - unchecked errors
if command -v go >/dev/null 2>&1; then
    for gofile in $(find . -name "*.go" -not -path "./vendor/*" -not -path "./OLD_IMPLEMENTATION/*"); do
        # Simple check for common unchecked errors
        if grep -n "err :=" "$gofile" | grep -v "if err" | grep -v "return.*err" | head -5; then
            print_warning "Potential unchecked errors in $gofile"
        fi
    done
fi

# 6. Check for copy-paste indicators
echo
echo "Checking for copy-paste indicators..."

# Look for duplicate function/class definitions
for file in $FILES; do
    if [ -f "$file" ]; then
        # Check for duplicate function names in same file
        if [ "${file##*.}" = "go" ]; then
            funcs=$(grep -o "^func.*(" "$file" | sort | uniq -d)
            if [ -n "$funcs" ]; then
                print_error "Duplicate function definitions in $file"
            fi
        fi
    fi
done

# 7. Check for metric anomalies
echo
echo "Checking code metrics..."

# Files that are too large (might indicate AI dumping everything in one file)
for file in $FILES; do
    if [ -f "$file" ]; then
        lines=$(wc -l < "$file")
        if [ "$lines" -gt 1000 ]; then
            print_warning "$file has $lines lines (consider splitting)"
        fi
        
        # Check for functions that are too long
        if [ "${file##*.}" = "go" ]; then
            # Simple heuristic: check distance between func declarations
            func_lines=$(grep -n "^func" "$file" | cut -d: -f1)
            prev_line=0
            for line in $func_lines; do
                if [ $prev_line -gt 0 ]; then
                    diff=$((line - prev_line))
                    if [ $diff -gt 200 ]; then
                        print_warning "$file has a function longer than 200 lines at line $prev_line"
                    fi
                fi
                prev_line=$line
            done
        fi
    fi
done

# 8. Check for language mixing (AI confusion)
echo
echo "Checking for language confusion..."

# Python syntax in Go files
if grep -r "def \|import \|from .* import\|print(" . \
    --include="*.go" \
    --exclude-dir="vendor" \
    --exclude-dir=".git" \
    --exclude-dir="OLD_IMPLEMENTATION"; then
    print_error "Python syntax found in Go files"
fi

# Go syntax in JavaScript/TypeScript files
if grep -r "func \|:=\|var .* string\|package " . \
    --include="*.js" \
    --include="*.ts" \
    --exclude-dir="node_modules" \
    --exclude-dir=".git" \
    --exclude-dir="OLD_IMPLEMENTATION" \
    --exclude="*.d.ts"; then
    print_error "Go syntax found in JavaScript/TypeScript files"
fi

# 9. Check against AI safety configuration
if [ -f "$AI_SAFETY_CONFIG" ]; then
    print_info "Validating against .ai-safety configuration..."
    
    # Extract forbidden patterns from .ai-safety file
    # This is a simplified check - in production, parse the YAML properly
    if grep -q "forbidden_patterns:" "$AI_SAFETY_CONFIG"; then
        # Check each forbidden pattern
        while IFS= read -r line; do
            if [[ $line =~ ^[[:space:]]*-[[:space:]]*\"(.*)\" ]]; then
                pattern="${BASH_REMATCH[1]}"
                if grep -r "$pattern" . \
                    --include="*.go" \
                    --include="*.js" \
                    --include="*.ts" \
                    --exclude-dir="vendor" \
                    --exclude-dir="node_modules" \
                    --exclude-dir=".git" \
                    --exclude-dir="OLD_IMPLEMENTATION" \
                    --exclude="$AI_SAFETY_CONFIG"; then
                    print_error "Forbidden pattern found: $pattern"
                fi
            fi
        done < <(sed -n '/forbidden_patterns:/,/^[^ ]/p' "$AI_SAFETY_CONFIG")
    fi
fi

# 10. Check for obvious AI hallucinations
echo
echo "Checking for potential AI hallucinations..."

# Non-existent standard library imports
hallucination_imports=(
    "fmt/printf"
    "strings/join"
    "errors/new"
    "context/context"
    "http/server"
)

for import in "${hallucination_imports[@]}"; do
    if grep -r "\"$import\"" . \
        --include="*.go" \
        --exclude-dir="vendor" \
        --exclude-dir=".git" \
        --exclude-dir="OLD_IMPLEMENTATION"; then
        print_error "Non-existent import detected (AI hallucination): $import"
    fi
done

# Summary
echo
echo "=== LLM Safety Check Summary ==="
if [ $ISSUES -eq 0 ] && [ $SUSPICIOUS -eq 0 ]; then
    echo -e "${GREEN}✓ No AI safety issues detected!${NC}"
    exit 0
else
    echo -e "Safety Violations: ${RED}$ISSUES${NC}"
    echo -e "Suspicious Patterns: ${YELLOW}$SUSPICIOUS${NC}"
    
    if [ $ISSUES -gt 0 ]; then
        echo -e "${RED}✗ LLM safety check failed${NC}"
        echo
        echo "Please review the detected issues. They may indicate:"
        echo "  - Incomplete AI code generation"
        echo "  - Security vulnerabilities"
        echo "  - Poor code quality"
        echo "  - AI hallucinations or confusion"
        exit 1
    else
        echo -e "${YELLOW}⚠ LLM safety check passed with warnings${NC}"
        echo
        echo "Please review the suspicious patterns."
        exit 0
    fi
fi