#!/bin/bash

# Phoenix Platform Duplicate Analysis Script
# Identifies unique code in duplicate services before elimination

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔍 Phoenix Platform Duplicate Analysis"
echo "===================================="
echo ""

# Create analysis directory
ANALYSIS_DIR="duplicate-analysis-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$ANALYSIS_DIR"

echo "📊 Analysis output directory: $ANALYSIS_DIR"
echo ""

# Function to analyze service differences
analyze_service() {
    local service=$1
    local old_path="services/$service"
    local new_path="projects/$service"
    
    if [ -d "$old_path" ] && [ -d "$new_path" ]; then
        echo -e "${BLUE}Analyzing:${NC} $service"
        
        # Create service analysis directory
        mkdir -p "$ANALYSIS_DIR/$service"
        
        # Count files
        old_files=$(find "$old_path" -name "*.go" -o -name "*.yaml" -o -name "*.proto" | wc -l)
        new_files=$(find "$new_path" -name "*.go" -o -name "*.yaml" -o -name "*.proto" | wc -l)
        
        echo "  Files: services/$service: $old_files, projects/$service: $new_files"
        
        # Find unique files in old location
        echo "  Checking for unique files in services/$service..."
        unique_count=0
        
        find "$old_path" -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.proto" \) | while read old_file; do
            relative_path=${old_file#$old_path/}
            new_file="$new_path/$relative_path"
            
            if [ ! -f "$new_file" ]; then
                echo "    ${YELLOW}UNIQUE:${NC} $relative_path"
                echo "$old_file" >> "$ANALYSIS_DIR/$service/unique-in-services.txt"
                ((unique_count++))
                
                # Copy unique file for review
                mkdir -p "$ANALYSIS_DIR/$service/unique-files/$(dirname "$relative_path")"
                cp "$old_file" "$ANALYSIS_DIR/$service/unique-files/$relative_path"
            fi
        done
        
        # Compare common files for differences
        echo "  Comparing common files..."
        diff_count=0
        
        find "$old_path" -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.proto" \) | while read old_file; do
            relative_path=${old_file#$old_path/}
            new_file="$new_path/$relative_path"
            
            if [ -f "$new_file" ]; then
                if ! diff -q "$old_file" "$new_file" > /dev/null 2>&1; then
                    echo "    ${YELLOW}DIFFERENT:${NC} $relative_path"
                    diff -u "$old_file" "$new_file" > "$ANALYSIS_DIR/$service/diff-$relative_path.diff" 2>/dev/null || true
                    ((diff_count++))
                fi
            fi
        done
        
        # Check for important patterns in old service
        echo "  Checking for critical code patterns..."
        
        # Look for state machines
        if grep -r "StateMachine\|state machine" "$old_path" --include="*.go" > /dev/null 2>&1; then
            echo "    ${RED}CRITICAL:${NC} State machine implementation found"
            grep -r "StateMachine\|state machine" "$old_path" --include="*.go" > "$ANALYSIS_DIR/$service/state-machine-refs.txt"
        fi
        
        # Look for database migrations
        if [ -d "$old_path/migrations" ]; then
            echo "    ${RED}CRITICAL:${NC} Database migrations found"
            cp -r "$old_path/migrations" "$ANALYSIS_DIR/$service/migrations-backup"
        fi
        
        # Look for gRPC implementations
        if grep -r "RegisterServer\|UnimplementedServer" "$old_path" --include="*.go" > /dev/null 2>&1; then
            echo "    ${YELLOW}IMPORTANT:${NC} gRPC server implementation found"
            grep -r "RegisterServer\|UnimplementedServer" "$old_path" --include="*.go" > "$ANALYSIS_DIR/$service/grpc-implementations.txt"
        fi
        
        # Generate summary
        cat > "$ANALYSIS_DIR/$service/summary.txt" << EOF
Service: $service
================
Old Path: $old_path
New Path: $new_path
Files in old: $old_files
Files in new: $new_files
Unique files: $(cat "$ANALYSIS_DIR/$service/unique-in-services.txt" 2>/dev/null | wc -l)
Different files: $(ls "$ANALYSIS_DIR/$service"/diff-*.diff 2>/dev/null | wc -l)

Recommendations:
EOF
        
        # Add recommendations based on findings
        if [ -f "$ANALYSIS_DIR/$service/state-machine-refs.txt" ]; then
            echo "- PRESERVE state machine implementation" >> "$ANALYSIS_DIR/$service/summary.txt"
        fi
        
        if [ -d "$ANALYSIS_DIR/$service/migrations-backup" ]; then
            echo "- MERGE database migrations" >> "$ANALYSIS_DIR/$service/summary.txt"
        fi
        
        if [ -f "$ANALYSIS_DIR/$service/unique-in-services.txt" ]; then
            echo "- REVIEW unique files for important logic" >> "$ANALYSIS_DIR/$service/summary.txt"
        fi
        
        echo -e "  ${GREEN}✓${NC} Analysis complete"
        echo ""
    fi
}

# Analyze each duplicate service
services_to_analyze=(
    "api"
    "controller"
    "generator"
    "dashboard"
    "benchmark"
    "analytics"
    "phoenix-cli"
    "loadsim-operator"
    "pipeline-operator"
)

for service in "${services_to_analyze[@]}"; do
    analyze_service "$service"
done

# Analyze operators
echo -e "${BLUE}Analyzing operators...${NC}"
echo ""

# Pipeline operator
if [ -d "operators/pipeline" ] && [ -d "projects/pipeline-operator" ]; then
    echo "  Comparing pipeline operator implementations..."
    diff -r "operators/pipeline" "projects/pipeline-operator" > "$ANALYSIS_DIR/pipeline-operator-diff.txt" 2>&1 || true
fi

# LoadSim operator
if [ -d "operators/loadsim" ] && [ -d "projects/loadsim-operator" ]; then
    echo "  Comparing loadsim operator implementations..."
    # Check if loadsim-operator is just a stub
    if grep -q "TODO: Implement" "projects/loadsim-operator/cmd/main.go" 2>/dev/null; then
        echo -e "    ${YELLOW}WARNING:${NC} projects/loadsim-operator is a stub implementation"
    fi
fi

# Analyze proto files
echo ""
echo -e "${BLUE}Analyzing proto files...${NC}"
mkdir -p "$ANALYSIS_DIR/protos"

# Find all proto files
find . -name "*.proto" -type f | grep -v node_modules | grep -v ".git" | sort > "$ANALYSIS_DIR/protos/all-protos.txt"

# Group by filename
cat "$ANALYSIS_DIR/protos/all-protos.txt" | while read proto; do
    filename=$(basename "$proto")
    echo "$proto" >> "$ANALYSIS_DIR/protos/by-name-$filename.txt"
done

# Report duplicates
for proto_list in "$ANALYSIS_DIR/protos/by-name-"*.txt; do
    if [ $(wc -l < "$proto_list") -gt 1 ]; then
        filename=$(basename "$proto_list" | sed 's/by-name-//;s/.txt//')
        echo -e "  ${YELLOW}Duplicate:${NC} $filename"
        cat "$proto_list" | while read location; do
            echo "    - $location"
        done
    fi
done

# Generate final report
echo ""
echo "📋 Generating final report..."

cat > "$ANALYSIS_DIR/REPORT.md" << 'EOF'
# Duplicate Analysis Report

Generated: $(date)

## Summary

This report analyzes duplicate implementations in the Phoenix Platform to identify unique code that must be preserved during consolidation.

## Critical Findings

### Services with Unique Implementations

EOF

# Add service summaries
for service in "${services_to_analyze[@]}"; do
    if [ -f "$ANALYSIS_DIR/$service/summary.txt" ]; then
        echo "### $service" >> "$ANALYSIS_DIR/REPORT.md"
        echo '```' >> "$ANALYSIS_DIR/REPORT.md"
        cat "$ANALYSIS_DIR/$service/summary.txt" >> "$ANALYSIS_DIR/REPORT.md"
        echo '```' >> "$ANALYSIS_DIR/REPORT.md"
        echo "" >> "$ANALYSIS_DIR/REPORT.md"
    fi
done

# Add proto analysis
echo "### Proto Files" >> "$ANALYSIS_DIR/REPORT.md"
echo "" >> "$ANALYSIS_DIR/REPORT.md"
echo "Duplicate proto files found:" >> "$ANALYSIS_DIR/REPORT.md"
for proto_list in "$ANALYSIS_DIR/protos/by-name-"*.txt; do
    if [ $(wc -l < "$proto_list") -gt 1 ]; then
        filename=$(basename "$proto_list" | sed 's/by-name-//;s/.txt//')
        echo "- $filename ($(wc -l < "$proto_list") locations)" >> "$ANALYSIS_DIR/REPORT.md"
    fi
done

echo "" >> "$ANALYSIS_DIR/REPORT.md"
echo "## Recommendations" >> "$ANALYSIS_DIR/REPORT.md"
echo "" >> "$ANALYSIS_DIR/REPORT.md"
echo "1. **Controller Service**: Contains critical state machine logic - carefully merge" >> "$ANALYSIS_DIR/REPORT.md"
echo "2. **Database Migrations**: Consolidate all migrations to avoid data loss" >> "$ANALYSIS_DIR/REPORT.md"
echo "3. **Proto Files**: Consolidate to single location, check for version differences" >> "$ANALYSIS_DIR/REPORT.md"
echo "4. **LoadSim Operator**: Keep operators/ version if more complete than stub" >> "$ANALYSIS_DIR/REPORT.md"

echo ""
echo -e "${GREEN}✅ Analysis complete!${NC}"
echo ""
echo "📊 Results saved to: $ANALYSIS_DIR/"
echo "📄 View report: $ANALYSIS_DIR/REPORT.md"
echo ""
echo "Next steps:"
echo "1. Review unique files in each service"
echo "2. Check diff files for important changes"
echo "3. Follow recommendations in REPORT.md"
echo "4. Run safe-eliminate-duplicates.sh after review"