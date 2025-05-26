#!/bin/bash

# Phoenix Platform Codebase Analysis Script
# Identifies duplicates, dead code, and structural issues

set -e

echo "🔍 Phoenix Platform Codebase Analysis"
echo "===================================="
echo ""

# Check for duplicate services
echo "📦 Checking for duplicate services..."
echo ""

# Compare services/ and projects/ directories
for dir in services/*; do
    if [ -d "$dir" ]; then
        service=$(basename "$dir")
        if [ -d "projects/$service" ]; then
            echo "❌ DUPLICATE: $service exists in both /services/ and /projects/"
            
            # Compare file counts
            service_files=$(find "services/$service" -name "*.go" 2>/dev/null | wc -l)
            project_files=$(find "projects/$service" -name "*.go" 2>/dev/null | wc -l)
            echo "   Files: services/$service: $service_files, projects/$service: $project_files"
            
            # Check if actively being deleted
            if git status --porcelain 2>/dev/null | grep -q "^D.*services/$service"; then
                echo "   ℹ️  Note: services/$service is marked for deletion"
            fi
        fi
    fi
done

echo ""
echo "🔄 Checking for duplicate operators..."
echo ""

# Check operators/ vs projects/*-operator
for op in operators/*; do
    if [ -d "$op" ]; then
        operator=$(basename "$op")
        if [ -d "projects/${operator}-operator" ]; then
            echo "❌ DUPLICATE: $operator operator in both locations"
        fi
    fi
done

echo ""
echo "📄 Checking for duplicate proto files..."
echo ""

# Find all proto files
proto_files=$(find . -name "*.proto" -type f | grep -v node_modules | grep -v ".git" || true)

# Simple duplicate check without associative array
echo "$proto_files" | while read -r file; do
    if [ -n "$file" ]; then
        filename=$(basename "$file")
        # Find other files with same name
        duplicates=$(find . -name "$filename" -type f | grep -v node_modules | grep -v ".git" | grep -v "^$file$" || true)
        if [ -n "$duplicates" ]; then
            echo "❌ DUPLICATE: $filename"
            echo "   - $file"
            echo "$duplicates" | while read -r dup; do
                echo "   - $dup"
            done
        fi
    fi
done | sort -u

echo ""
echo "💀 Checking for dead code and empty directories..."
echo ""

# Find empty directories
empty_dirs=$(find pkg -type d -empty 2>/dev/null)
if [ -n "$empty_dirs" ]; then
    echo "Empty directories in pkg/:"
    echo "$empty_dirs" | while read -r dir; do
        echo "   - $dir"
    done
fi

echo ""
echo "🔍 Checking for TODO/FIXME markers..."
echo ""

# Find TODO/FIXME in Go files
todo_files=$(grep -r "TODO\|FIXME\|HACK" --include="*.go" . 2>/dev/null | grep -v vendor | grep -v node_modules || true)
todo_count=$(echo "$todo_files" | grep -c "TODO\|FIXME\|HACK" || echo "0")

echo "Found $todo_count TODO/FIXME markers"
if [ "$todo_count" -gt 0 ] && [ "$todo_count" -lt 20 ]; then
    echo "$todo_files" | head -10
fi

echo ""
echo "📊 Checking for stub implementations..."
echo ""

# Find potential stub files
stub_files=$(grep -r "not.implemented\|stub.only\|TODO:.Implement" --include="*.go" . 2>/dev/null | grep -v vendor || true)
if [ -n "$stub_files" ]; then
    echo "Potential stub implementations:"
    echo "$stub_files" | head -10
fi

echo ""
echo "📈 Code statistics..."
echo ""

# Count Go files in different directories
echo "Go file distribution:"
echo "  services/: $(find services -name "*.go" 2>/dev/null | wc -l) files"
echo "  projects/: $(find projects -name "*.go" 2>/dev/null | wc -l) files"
echo "  pkg/: $(find pkg -name "*.go" 2>/dev/null | wc -l) files"
echo "  operators/: $(find operators -name "*.go" 2>/dev/null | wc -l) files"

echo ""
echo "🔧 Checking go.work for obsolete entries..."
echo ""

if [ -f "go.work" ]; then
    while IFS= read -r line; do
        if [[ "$line" =~ ^[[:space:]]*\./(.*) ]]; then
            path="${BASH_REMATCH[1]}"
            if [ ! -d "$path" ]; then
                echo "❌ go.work references non-existent path: $path"
            fi
        fi
    done < go.work
fi

echo ""
echo "📦 Checking for unused packages..."
echo ""

# Check for packages with no imports (simplified check)
for pkg_dir in pkg/*/; do
    if [ -d "$pkg_dir" ]; then
        pkg_name=$(basename "$pkg_dir")
        # Check if package is imported anywhere
        import_count=$(grep -r "github.com/phoenix/platform/pkg/$pkg_name" --include="*.go" . 2>/dev/null | grep -v "^$pkg_dir" | wc -l || echo "0")
        if [ "$import_count" -eq 0 ]; then
            go_files=$(find "$pkg_dir" -name "*.go" 2>/dev/null | wc -l)
            if [ "$go_files" -gt 0 ]; then
                echo "⚠️  Package pkg/$pkg_name appears unused (0 imports)"
            fi
        fi
    fi
done

echo ""
echo "✅ Analysis complete!"
echo ""
echo "📋 Summary:"
echo "  - Run './scripts/cleanup-codebase.sh' to clean up duplicates"
echo "  - Review CODEBASE_CLEANUP_PLAN.md for detailed recommendations"
echo "  - Backup before making changes!"